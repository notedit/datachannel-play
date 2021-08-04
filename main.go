package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v3"
)

func test(c *gin.Context) {
	c.String(200, "Hello World")
}

func index(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{})
}

func publish(c *gin.Context) {

	var data struct {
		Sdp string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	var config = webrtc.Configuration{
		ICEServers:   []webrtc.ICEServer{},
		BundlePolicy: webrtc.BundlePolicyMaxBundle,
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnOpen(func() {
			log.Printf("OnOpen: %s-%d. Start receiving data", dc.Label(), dc.ID())

		})

		// Register the OnMessage to handle incoming messages
		dc.OnMessage(func(dcMsg webrtc.DataChannelMessage) {
			n := len(dcMsg.Data)
			fmt.Println("OnMessage ", n)
			fmt.Println(dcMsg.Data)
		})
	})

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  data.Sdp,
	}

	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	peerConnection.SetLocalDescription(answer)

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	<-gatherComplete

	c.JSON(200, gin.H{
		"sdp": peerConnection.LocalDescription().SDP,
	})
}

func main() {

	router := gin.Default()
	corsc := cors.DefaultConfig()
	corsc.AllowAllOrigins = true
	corsc.AllowCredentials = true
	router.Use(cors.New(corsc))

	router.LoadHTMLFiles("./index.html")

	router.GET("/test", test)

	router.GET("/", index)

	router.POST("/rtc/v1/publish", publish)

	router.Run(":8080")

}
