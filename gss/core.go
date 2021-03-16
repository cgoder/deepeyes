package gss

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cgoder/vdk/av"
	"github.com/cgoder/vdk/format/rtmp"
	"github.com/cgoder/vdk/format/rtspv2"
	log "github.com/sirupsen/logrus"
)

//readPktByProtocol read av.Pkt from av.Source.
func (svr *ServerST) readPktByProtocol(ctx context.Context, streamID string, channelID string, channel *ChannelST) error {
	if strings.HasPrefix(channel.URL, "rtmp://") {
		_, err := svr.streamRtmp(ctx, streamID, channelID, channel)
		if err != nil {
			log.WithFields(log.Fields{
				"module":  "core",
				"stream":  streamID,
				"channel": channelID,
				"func":    "streamByProtocol",
				"call":    "streamRtmp",
			}).Errorln(err)
		}
		return err
	} else if strings.HasPrefix(channel.URL, "rtsp://") {
		_, err := svr.streamRtsp(ctx, streamID, channelID, channel)
		if err != nil {
			log.WithFields(log.Fields{
				"module":  "core",
				"stream":  streamID,
				"channel": channelID,
				"func":    "streamByProtocol",
				"call":    "streamRtsp",
			}).Errorln(err)
		}
		return err
	} else {
		log.WithFields(log.Fields{
			"module":  "core",
			"stream":  streamID,
			"channel": channelID,
			"func":    "streamByProtocol",
			"call":    "protocal select",
		}).Errorln("Unsupport protocol. ", channel.URL)
		return errors.New("Unsupport protocol")
	}
}

//streamRtmp read av.Pkt from rtmp stream to av.avQue.
func (svr *ServerST) streamRtmp(ctx context.Context, streamID string, channelID string, channel *ChannelST) (int, error) {
	var RTMPConn *rtmp.Conn
	var err error

	// rtmp connect.
	if err := func() error {
		// rtmp client dial
		RTMPConn, err = rtmp.Dial(channel.URL)

		if err != nil {
			log.WithFields(log.Fields{
				"module":  "core",
				"stream":  streamID,
				"channel": channelID,
				"func":    "streamRtmp",
				"call":    "rtmp.Dial",
			}).Errorln("RTMP Dial ---> ", JsonFormat(channel.URL), err)
			return err
		}
		log.WithFields(log.Fields{
			"module":  "core",
			"stream":  streamID,
			"channel": channelID,
			"func":    "streamRtmp",
			"call":    "Start",
		}).Debugln("RTMP Conn---> ", JsonFormat(channel.URL))

		// get av.Codec
		t1 := time.Now().Local().UTC()
		streams, err := RTMPConn.Streams()
		if err != nil {
			log.WithFields(log.Fields{
				"module":  "core",
				"stream":  streamID,
				"channel": channelID,
				"func":    "streamRtmp",
				"call":    "RTMPConn.Streams",
			}).Errorln("RTMP get stream codec err. ", err)
			return err
		}

		// update av.Codec
		if len(streams) > 0 {
			svr.StreamCodecUpdate(streamID, channelID, streams, nil)
			channel.cond.Broadcast()

			log.WithFields(log.Fields{
				"module":  "core",
				"stream":  streamID,
				"channel": channelID,
				"func":    "streamRtmp",
				"call":    "RTMPConn.Streams",
			}).Debugln("rtmp get stream codec DONE! time: ", time.Now().Local().UTC().Sub(t1).String())
		} else {
			log.WithFields(log.Fields{
				"module":  "core",
				"stream":  streamID,
				"channel": channelID,
				"func":    "streamRtmp",
				"call":    "RTMPConn.Streams",
			}).Errorln("rtmp get stream codec fail! time: ", time.Now().Local().UTC().Sub(t1).String())
			return ErrorStreamCodecNotFound
		}

		log.WithFields(log.Fields{
			"module":  "core",
			"stream":  streamID,
			"channel": channelID,
			"func":    "streamRtmp",
			"call":    "Start",
		}).Debugln("Success connection RTMP")

		return nil
	}(); err != nil {
		log.WithFields(log.Fields{
			"module":  "core",
			"stream":  streamID,
			"channel": channelID,
			"func":    "streamRtmp",
			"call":    "rtmp connect",
		}).Errorln("RTMP connect fail. ---> ", JsonFormat(channel.URL), err)
		return 0, err
	}

	ctx4Write, cancel4Write := context.WithCancel(context.Background())
	go readPktToQueue(ctx4Write, streamID, channelID, channel)

	// release hls cache
	defer func() {
		//cancel writePktToQueue goroutine.
		cancel4Write()
		RTMPConn.Close()
		svr.StreamCodecUpdate(streamID, channelID, nil, nil)
		svr.StreamStatusUpdate(streamID, channelID, STREAM_OFFLINE)
		// svr.StreamHLSFlush(streamID, channelID)
	}()

	checkClients := time.NewTimer(time.Duration(timeoutClientCheck) * time.Second)
	defer checkClients.Stop()
	// var preKeyTS = time.Duration(0)
	// var Seq []*av.Packet
	// var pktCnt int
	for {
		select {
		// case <-ctx.Done():
		// 	log.WithFields(log.Fields{
		// 		"module":  "core",
		// 		"stream":  streamID,
		// 		"channel": channelID,
		// 		"func":    "streamRtmp",
		// 		"call":    "ctx.Done()",
		// 	}).Debugln("Stream close by cancel. ")
		// 	return 0, nil
		//Check stream have clients
		case <-checkClients.C:
			cCnt := svr.ClientCount(streamID, channelID)
			if cCnt == 0 {
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "streamRtmp",
					"call":    "ClientCount",
				}).Debugln("Stream close has no client. ")

				return 0, ErrorStreamNoClients
			}

			// log.Println("clients: ", cCnt)
			if len(checkClients.C) > 0 {
				<-checkClients.C
			}
			if b := checkClients.Reset(time.Duration(timeoutClientCheck) * time.Second); !b {
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "streamRtmp",
					"call":    "timer.Reset",
				}).Errorln("checkClients timer reset err", cCnt)
			}
		//Read core signals
		case signals := <-channel.signals:
			switch signals {
			case SIGNAL_STREAM_STOP:
				return 0, ErrorStreamStopCoreSignal
			case SIGNAL_STREAM_AVCODEC_UPDATE:
				//TODO:
				// return 0, ErrorStreamCodecUpdate
			}
		//Read av.Pkt,and proxy for all clients.
		// TODO: av.Pkt be save file here.
		default:
			avPkt, err := RTMPConn.ReadPacket()
			if err != nil {
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "streamRtmp",
					"call":    "ReadPacket",
				}).Errorln("ReadPacket error ", err)
				continue
			}

			// pktCnt++
			// if avPkt.IsKeyFrame {
			// 	log.Println("Write keyframe to queue. ", avPkt.Time, len(avPkt.Data))
			// }

			// write av.Pkt to avQue
			if channel.source.avQue != nil {
				channel.source.avQue.WritePacket(avPkt)
			} else {
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "streamRtmp",
					"call":    "ReadPacket",
				}).Errorln("channel source avQue nil. ", err)
				panic("channel source avQue nil. ")
			}

			// if avPkt.IsKeyFrame {
			// 	if preKeyTS > 0 {
			// 		svr.StreamHLSAdd(streamID, channelID, Seq, avPkt.Time-preKeyTS)
			// 		Seq = []*av.Packet{}
			// 	}
			// 	preKeyTS = avPkt.Time
			// }
			// Seq = append(Seq, &avPkt)

		}
	}
}

//streamRtsp read av.Pkt from rtmp stream to av.avQue.
func (svr *ServerST) streamRtsp(ctx context.Context, streamID string, channelID string, channel *ChannelST) (int, error) {
	t1 := time.Now().Local().UTC()
	// rtsp client dial
	RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{URL: channel.URL, DisableAudio: false, DialTimeout: 3 * time.Second, ReadWriteTimeout: 5 * time.Second, Debug: false, OutgoingProxy: false})
	if err != nil {
		log.WithFields(log.Fields{
			"module":  "core",
			"stream":  streamID,
			"channel": channelID,
			"func":    "streamRtsp",
			"call":    "rtspv2.Dial",
		}).Errorln("RTSPClient.Dial fail. ", err)
		return 0, err
	}
	log.WithFields(log.Fields{
		"module":  "core",
		"stream":  streamID,
		"channel": channelID,
		"func":    "streamRtsp",
		"call":    "rtspv2.Dial",
	}).Debugln("RTSPClient.SDPRaw---> ", JsonFormat(RTSPClient.SDPRaw))

	if len(RTSPClient.CodecData) > 0 {
		svr.StreamCodecUpdate(streamID, channelID, RTSPClient.CodecData, RTSPClient.SDPRaw)

		// channel.updated <- true
		channel.cond.Broadcast()

		log.WithFields(log.Fields{
			"module":  "core",
			"stream":  streamID,
			"channel": channelID,
			"func":    "streamRtsp",
			"call":    "rtspv2.Dial",
		}).Debugln("rtsp get stream codec DONE! time: ", time.Now().Local().UTC().Sub(t1).String())
	} else {
		log.WithFields(log.Fields{
			"module":  "core",
			"stream":  streamID,
			"channel": channelID,
			"func":    "streamRtsp",
			"call":    "rtspv2.Dial",
		}).Errorln("rtsp get stream codec fail! time: ", time.Now().Local().UTC().Sub(t1).String())
		return 0, ErrorStreamCodecNotFound
	}

	log.WithFields(log.Fields{
		"module":  "core",
		"stream":  streamID,
		"channel": channelID,
		"func":    "streamRtsp",
		"call":    "ChannelStatusUpdate",
	}).Debugln("Success connection RTSP")

	ctx4Write, cancel4Write := context.WithCancel(context.Background())
	go readPktToQueue(ctx4Write, streamID, channelID, channel)

	// release hls cache
	defer func() {
		cancel4Write()
		RTSPClient.Close()
		svr.StreamStatusUpdate(streamID, channelID, STREAM_OFFLINE)
		svr.StreamCodecUpdate(streamID, channelID, nil, nil)
		// svr.StreamHLSFlush(streamID, channelID)
	}()

	checkClients := time.NewTimer(time.Duration(timeoutClientCheck) * time.Second)
	defer checkClients.Stop()
	// var preKeyTS = time.Duration(0)
	// var Seq []*av.Packet

	for {
		select {
		// case <-ctx.Done():
		// 	log.WithFields(log.Fields{
		// 		"module":  "core",
		// 		"stream":  streamID,
		// 		"channel": channelID,
		// 		"func":    "streamRtsp",
		// 		"call":    "ctx.Done()",
		// 	}).Debugln("Stream close by cancel. ")
		// 	return 0, nil
		case <-checkClients.C:
			cCnt := svr.ClientCount(streamID, channelID)
			if cCnt == 0 {
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "streamRtsp",
					"call":    "ClientCount",
				}).Debugln("Stream close has no client. ")

				return 0, ErrorStreamNoClients
			}

			// log.Println("clients: ", cCnt)
			if len(checkClients.C) > 0 {
				<-checkClients.C
			}
			if b := checkClients.Reset(time.Duration(timeoutClientCheck) * time.Second); !b {
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "streamRtsp",
					"call":    "timer.Reset",
				}).Errorln("checkClients timer reset err", cCnt)
			}
		//Read core signals
		case signals := <-channel.signals:
			switch signals {
			case SIGNAL_STREAM_STOP:
				return 0, ErrorStreamStopCoreSignal
			case SIGNAL_STREAM_AVCODEC_UPDATE:
				//TODO:
				// return 0, ErrorStreamCodecUpdate
			}
		//Read rtsp signals
		case signals := <-RTSPClient.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				svr.StreamCodecUpdate(streamID, channelID, RTSPClient.CodecData, RTSPClient.SDPRaw)
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "streamRtsp",
					"call":    "RTSPClient.Signals",
				}).Debugln("Stream codec update. ")
			case rtspv2.SignalStreamRTPStop:
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "streamRtsp",
					"call":    "RTSPClient.Signals",
				}).Errorln(ErrorStreamStopRTSPSignal)
				return 0, ErrorStreamStopRTSPSignal
			}
		// read rtp.Pkt for cast all clients. MUST read from OutgoingProxyQueue.
		// case <-RTSPClient.OutgoingProxyQueue:
		// Svr.ChannelCastProxy(streamID, channelID, packetRTP)
		// read av.Pkt for cast all clients.
		case avPkt := <-RTSPClient.OutgoingPacketQueue:
			// Svr.ChannelCast(streamID, channelID, avPkt)

			if channel.source.avQue != nil {
				channel.source.avQue.WritePacket(*avPkt)
			} else {
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "streamRtsp",
					"call":    "ReadPacket",
				}).Errorln("channel source avQue nil. ", err)
				panic("channel source avQue nil. ")
			}

			// if avPkt.IsKeyFrame {
			// 	if preKeyTS > 0 {
			// 		svr.StreamHLSAdd(streamID, channelID, Seq, avPkt.Time-preKeyTS)
			// 		Seq = []*av.Packet{}
			// 	}
			// 	preKeyTS = avPkt.Time
			// }
			// Seq = append(Seq, avPkt)
		}
	}
}

//writePktToClients cast av.Pkt to all clients.
func writePktToClients(clients map[string]*ClientST, avPkt *av.Packet) {

	if len(clients) > 0 {
		for _, client := range clients {
			if client.protocol == PLAY_RTSP {
				continue
			} else {
				if len(client.outgoingAVPacket) < lenAvPacketQueue {
					// log.Println("w2c ", avPkt.Idx, avPkt.IsKeyFrame)
					client.outgoingAVPacket <- avPkt
				} else {
					// log.WithFields(log.Fields{
					// 	"module": "core",
					// 	// "stream":  streamID,
					// 	// "channel": channelID,
					// 	"func": "writePktToAllClient",
					// 	"call": "client.outgoingAVPacket",
					// }).Errorln("client av chan full. ", len(client.outgoingAVPacket))
				}
			}
		}
	}
}

//readPktToQueue write av.Pkt for all clients. which from channel.av.queue.
func readPktToQueue(ctx context.Context, streamID string, channelID string, channel *ChannelST) error {

	// get av.Pkt from queue.
	cursor := channel.source.avQue.Latest()

	var videoStart bool
	clients := channel.clients

	// checkClients := time.NewTimer(time.Duration(timeoutClientCheck) * time.Second)
	// defer checkClients.Stop()
	for {
		select {
		case <-ctx.Done():
			log.WithFields(log.Fields{
				"module":  "core",
				"stream":  streamID,
				"channel": channelID,
				"func":    "writePktToQueue",
				"call":    "ctx.Done()",
			}).Debugln("End write avPkt by cancel. ")
			return nil
		// //Check stream have clients
		// case <-checkClients.C:
		// 	cCnt := Svr.ClientCount(streamID, channelID)
		// 	if cCnt == 0 {
		// 		log.WithFields(log.Fields{
		// 			"module":  "core",
		// 			"stream":  streamID,
		// 			"channel": channelID,
		// 			"func":    "writePktToQueue",
		// 			"call":    "ClientCount",
		// 		}).Debugln("Stream close has no client. ")
		// 		return ErrorStreamNoClients
		// 	}
		// 	log.Println("clients: ", cCnt)
		// 	checkClients.Reset(time.Duration(timeoutClientCheck) * time.Second)
		//Read core signals
		// case signals := <-channel.signals:
		// 	switch signals {
		// 	case SIGNAL_STREAM_STOP:
		// 		return ErrorStreamStopCoreSignal
		// 	case SignalStreamRestart:
		// 		//TODO:
		// 		// return 0, ErrorStreamRestart
		// 	case SIGNAL_STREAM_AVCODEC_UPDATE:
		// 		//TODO:
		// 		// return 0, ErrorStreamCodecUpdate
		// 	}
		//Read av.Pkt,and proxy for all clients.
		// TODO: av.Pkt be save file here.
		default:
			packet, err := cursor.ReadPacket()
			if err != nil {
				log.WithFields(log.Fields{
					"module":  "core",
					"stream":  streamID,
					"channel": channelID,
					"func":    "writePktToQueue",
					"call":    "ReadPacket",
				}).Errorln("Queue ReadPacket error ", err)
				continue
			}

			// checkAvRead.Reset(time.Duration(timeoutAvReadCheck) * time.Second)

			if packet.IsKeyFrame {
				// log.Println("Queue Write keyframe to clients. ", packet.Time, len(packet.Data))
				videoStart = true
			}

			if !videoStart {
				continue
			}

			writePktToClients(clients, &packet)

		}
	}

}
