package api

import (
	"context"
	"time"

	"github.com/cgoder/deepeyes/gss"
	webrtc "github.com/cgoder/vdk/format/webrtcv3"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

//HTTPAPIServerStreamWebRTC stream video over WebRTC
func HTTPAPIServerStreamWebRTC(c *gin.Context) {
	streamID := c.Param("uuid")
	channelID := c.Param("channel")

	// ctx, cancel := context.WithCancel(context.Background())
	ctx := context.Background()
	// defer func() {
	// 	// _ = ws.Close()
	// 	// c.IndentedJSON(500, Message{Status: 0, Payload: ErrorProgramNotFound.Error()})
	// 	log.WithFields(log.Fields{
	// 		"module":  "http_webrtc",
	// 		"stream":  streamID,
	// 		"channel": channelID,
	// 		"func":    "HTTPAPIServerStreamWebRTC",
	// 		"call":    "WebRTC",
	// 	}).Debugln("WebRTC Exit")

	// 	cancel()
	// 	log.WithFields(log.Fields{
	// 		"module":  "http_webrtc",
	// 		"stream":  streamID,
	// 		"channel": channelID,
	// 		"func":    "HTTPAPIServerStreamWebRTC",
	// 		"call":    "recv avpkt",
	// 	}).Infoln("Cancell av send goroutine.")
	// }()

	log.Println("webrtc++++++")

	ch, err := service.ChannelGet(streamID, channelID)
	if err != nil {
		log.WithFields(log.Fields{
			"module":  "http_mse",
			"stream":  streamID,
			"channel": channelID,
			"func":    "HTTPAPIServerStreamWebRTC",
			"call":    "ChannelGet",
		}).Errorln(err.Error())
		return
	}

	// get stream av.Codec
	// codecs, err := service.PlayPre(ctx, streamID, channelID)
	cid, codecs, avChanR, err := service.PlayPre(ctx, ch.URL, gss.PLAY_WEBRTC)
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module":  "http_mse",
			"stream":  streamID,
			"channel": channelID,
			"func":    "HTTPAPIServerStreamWebRTC",
			"call":    "PlayPre",
		}).Errorln(err.Error())
		return
	}

	log.Println("webrtc++++++ ok")

	// make writeable chan for read av.pkt
	// eofSignal := make(chan interface{}, 1)
	// defer func() {
	// 	eofSignal <- "wsEOF"
	// }()

	//webrtc init. will be close by VDK webrtc mod if timeout.
	muxerWebRTC := webrtc.NewMuxer(webrtc.Options{})
	answer, err := muxerWebRTC.WriteHeader(codecs, c.PostForm("data"))
	if err != nil {
		// c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module":  "http_webrtc",
			"stream":  streamID,
			"channel": channelID,
			"func":    "HTTPAPIServerStreamWebRTC",
			"call":    "WriteHeader",
		}).Errorln(err.Error())
		return
	}
	_, err = c.Writer.Write([]byte(answer))
	if err != nil {
		// c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module":  "http_webrtc",
			"stream":  streamID,
			"channel": channelID,
			"func":    "HTTPAPIServerStreamWebRTC",
			"call":    "Write",
		}).Errorln(err.Error())
		return
	}

	go func() {
		// delete rtc
		defer muxerWebRTC.Close()
		// delete client/player
		defer service.PlayStop(cid)

		var videoStart bool

		for {
			select {
			case <-ctx.Done():
				// c.IndentedJSON(500, Message{Status: 0, Payload: gss.ErrorStreamNoVideo.Error()})
				log.WithFields(log.Fields{
					"module":  "http_webrtc",
					"stream":  streamID,
					"channel": channelID,
					"func":    "HTTPAPIServerStreamWebRTC",
					"call":    "context.Done",
				}).Debugln(ctx.Err())
				return
			// case <-eofSignal:
			// 	log.WithFields(log.Fields{
			// 		"module":  "http_webrtc",
			// 		"stream":  streamID,
			// 		"channel": channelID,
			// 		"func":    "HTTPAPIServerStreamWebRTC",
			// 		"call":    "got eof signal.",
			// 	}).Debugln("got eof signal.")
			// 	return
			case avPkt := <-avChanR:
				if avPkt.IsKeyFrame {
					videoStart = true
				}
				// noVideo.Reset(time.Duration(timeout_novideo) * time.Second)

				if !videoStart {
					continue
				}

				// if avPkt.IsKeyFrame {
				// 	log.Println("got avPkt. ", avPkt.IsKeyFrame, len(avPkt.Data))
				// }
				err = muxerWebRTC.WritePacket(*avPkt)
				if err != nil {
					log.WithFields(log.Fields{
						"module":  "http_webrtc",
						"stream":  streamID,
						"channel": channelID,
						"func":    "HTTPAPIServerStreamWebRTC",
						"call":    "WritePacket",
					}).Errorln(err.Error())
					return
				}

				// if ch.Debug && avPkt.IsKeyFrame {
				// 	log.WithFields(log.Fields{
				// 		"module":  "http_webrtc",
				// 		"stream":  streamID,
				// 		"channel": channelID,
				// 		"func":    "HTTPAPIServerStreamWebRTC",
				// 		"call":    "recv avpkt",
				// 	}).Debugf("Send frame, key:%v, len:%v, DTS:%v, Dur:%v", avPkt.IsKeyFrame, len(avPkt.Data), avPkt.Time, avPkt.Duration)
				// }

			}
		}
	}()

	log.Println("webrtc++++++ end")
}

func HTTPAPIServerStreamWebRTC_orignal(c *gin.Context) {
	// log.Println("enter webrtc ....................")

	if !service.ChannelExist(c.Param("uuid"), c.Param("channel")) {
		c.IndentedJSON(500, Message{Status: 0, Payload: gss.ErrorProgramNotFound.Error()})
		log.WithFields(log.Fields{
			"module":  "http_webrtc",
			"stream":  c.Param("uuid"),
			"channel": c.Param("channel"),
			"func":    "HTTPAPIServerStreamWebRTC",
			"call":    "StreamChannelExist",
		}).Errorln(gss.ErrorProgramNotFound.Error())
		return
	}
	service.ChannelRun(context.Background(), c.Param("uuid"), c.Param("channel"))
	codecs, err := service.StreamCodecGet(c.Param("uuid"), c.Param("channel"))
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module":  "http_webrtc",
			"stream":  c.Param("uuid"),
			"channel": c.Param("channel"),
			"func":    "HTTPAPIServerStreamWebRTC",
			"call":    "StreamCodecs",
		}).Errorln(err.Error())
		return
	}
	muxerWebRTC := webrtc.NewMuxer(webrtc.Options{})
	// log.Println("webrtc headr data ->>>> ",c.PostForm("data"))
	answer, err := muxerWebRTC.WriteHeader(codecs, c.PostForm("data"))
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module":  "http_webrtc",
			"stream":  c.Param("uuid"),
			"channel": c.Param("channel"),
			"func":    "HTTPAPIServerStreamWebRTC",
			"call":    "WriteHeader",
		}).Errorln(err.Error())
		return
	}
	_, err = c.Writer.Write([]byte(answer))
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module":  "http_webrtc",
			"stream":  c.Param("uuid"),
			"channel": c.Param("channel"),
			"func":    "HTTPAPIServerStreamWebRTC",
			"call":    "Write",
		}).Errorln(err.Error())
		return
	}
	go func() {
		cid, ch, err := service.ClientAdd(c.Param("uuid"), c.Param("channel"), gss.PLAY_WEBRTC)
		if err != nil {
			c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
			log.WithFields(log.Fields{
				"module":  "http_webrtc",
				"stream":  c.Param("uuid"),
				"channel": c.Param("channel"),
				"func":    "HTTPAPIServerStreamWebRTC",
				"call":    "ClientAdd",
			}).Errorln(err.Error())
			return
		}
		defer service.ClientDelete(c.Param("uuid"), cid, c.Param("channel"))
		var videoStart bool
		noVideo := time.NewTimer(10 * time.Second)
		defer noVideo.Stop()
		for {
			select {
			case <-noVideo.C:
				c.IndentedJSON(500, Message{Status: 0, Payload: gss.ErrorStreamNoVideo.Error()})
				log.WithFields(log.Fields{
					"module":  "http_webrtc",
					"stream":  c.Param("uuid"),
					"channel": c.Param("channel"),
					"func":    "HTTPAPIServerStreamWebRTC",
					"call":    "ErrorStreamNoVideo",
				}).Errorln(gss.ErrorStreamNoVideo.Error())
				return
			case pck := <-ch:
				if pck.IsKeyFrame {
					noVideo.Reset(10 * time.Second)
					videoStart = true
				}
				if !videoStart {
					continue
				}
				err = muxerWebRTC.WritePacket(*pck)
				if err != nil {
					log.WithFields(log.Fields{
						"module":  "http_webrtc",
						"stream":  c.Param("uuid"),
						"channel": c.Param("channel"),
						"func":    "HTTPAPIServerStreamWebRTC",
						"call":    "WritePacket",
					}).Errorln(err.Error())
					return
				}
			}
		}
	}()

	// log.Println("exit webrtc ....................")
}
