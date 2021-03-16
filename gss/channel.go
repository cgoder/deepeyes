package gss

import (
	"context"
	"sync"

	"github.com/cgoder/vdk/av/pubsub"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

//ChannelNew new channel svr.
func ChannelNew(ch *ChannelST) *ChannelST {
	var tmpCh ChannelST
	if ch != nil {
		tmpCh = *ch
	}

	//gen uuid
	if tmpCh.UUID == "" {
		tmpCh.UUID = GenerateUUID()
	}
	if tmpCh.Name == "" {
		tmpCh.Name = tmpCh.UUID
	}

	//init source stream
	var source AvStream
	source.avQue = pubsub.NewQueue()
	tmpCh.source = source
	//init client's
	tmpCh.clients = make(map[string]*ClientST)

	// make chan for av.Codec update
	tmpCh.cond = sync.NewCond(&sync.Mutex{})
	//make signals buffer chain
	tmpCh.signals = make(chan int, 10)

	//make hls buffer
	tmpCh.hlsSegmentBuffer = make(map[int]*Segment)
	tmpCh.hlsSegmentNumber = 0

	//init debug
	tmpCh.OnDemand = true
	tmpCh.Debug = false

	return &tmpCh

}

//ChannelRelease renew channel for GC.
func ChannelRelease(ch *ChannelST) {
	var tmpCh ChannelST

	//release
	if ch.source.avQue != nil {
		ch.source.avQue.Close()
		ch.source.avQue = nil
	}
	var source AvStream
	ch.source = source
	ch.clients = make(map[string]*ClientST)
	ch.signals = make(chan int, 10)
	ch.hlsSegmentBuffer = make(map[int]*Segment)
	ch.hlsSegmentNumber = 0
	ch.OnDemand = true
	ch.Debug = false

	ch = &tmpCh

}

//ChannelRunAll run all stream go
func (svr *ServerST) ChannelRunAll(ctx context.Context) {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	for prgmID, program := range svr.Programs {
		for chID, ch := range program.Channels {
			if !ch.OnDemand {
				go svr.ChannelRun(ctx, prgmID, chID)
			}
		}
	}
}

//ChannelRun one stream and lock
func (svr *ServerST) ChannelRun(ctx context.Context, programID string, channelID string) error {
	// log.WithFields(logrus.Fields{
	// 	"module":  "Channel",
	// 	"stream":  programID,
	// 	"channel": channelID,
	// 	"func":    "ChannelRun",
	// 	"call":    "ChannelRun",
	// }).Debugln("ChannelRun ->>>>>>>>>>>>>>>>>")

	// get stream channel
	svr.mutex.Lock()
	ch, ok := svr.Programs[programID].Channels[channelID]
	svr.mutex.Unlock()
	if !ok {
		log.WithFields(logrus.Fields{
			"module":  "Channel",
			"stream":  programID,
			"channel": channelID,
			"func":    "ChannelRun",
			"call":    "ChannelGet",
		}).Errorln(ErrorChannelNotFound)
		return ErrorChannelNotFound
	}

	if ch.source.status != STREAM_ONLINE {
		go svr.StreamRun(context.Background(), programID, channelID)
	} else {
		log.WithFields(logrus.Fields{
			"module":  "Channel",
			"stream":  programID,
			"channel": channelID,
			"func":    "ChannelRun",
			"call":    "StreamRun",
		}).Debugln("stream is running... clients:", len(ch.clients))
	}

	return nil
}

//ChannelStop stop stream
func (svr *ServerST) ChannelStop(programID string, channelID string) error {
	// get stream channel
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	if programID == "" && channelID == "" {
		for _, program := range svr.Programs {
			for _, ch := range program.Channels {
				//TODO: stop stream.

				//set stream status
				ch.source.status = STREAM_OFFLINE
			}
		}
		return nil
	} else {
		if channelID != "" {
			ch, ok := svr.Programs[programID].Channels[channelID]
			if !ok {
				return ErrorChannelNotFound
			}
			//TODO: stop stream.

			//set stream status
			ch.source.status = STREAM_OFFLINE
		} else {
			program, ok := svr.Programs[programID]
			if !ok {
				return ErrorProgramNotFound
			}
			for _, ch := range program.Channels {
				//TODO: stop stream.

				//set stream status
				ch.source.status = STREAM_OFFLINE
			}
		}
	}
	return nil
}

// //ChannelCast broadcast stream av.Pkt
// func (svr *ServerST) ChannelCast(programID string, channelID string, avPkt *av.Packet) error {
// 	svr.mutex.Lock()
// 	ch, ok := svr.Programs[programID].Channels[channelID]
// 	svr.mutex.Unlock()
// 	if !ok || ch == nil {
// 		log.WithFields(logrus.Fields{
// 			"module":  "Channel",
// 			"stream":  programID,
// 			"channel": channelID,
// 			"func":    "ChannelCast",
// 			"call":    "ChannelGet",
// 		}).Errorf("Get channel fail!")
// 		return ErrorChannelNotFound
// 	}

// 	if len(ch.clients) > 0 {
// 		for _, client := range ch.clients {
// 			if client.protocol == PLAY_RTSP {
// 				continue
// 			}
// 			if len(client.outgoingAVPacket) < lenAvPacketQueue {
// 				client.outgoingAVPacket <- avPkt
// 			}
// 	}
// }

//ChannelSDP codec storage or wait
// func (svr *ServerST) ChannelSDP(programID string, channelID string) ([]byte, error) {
// 	// why for 100 times?
// 	for i := 0; i < 100; i++ {
// 		svr.mutex.RLock()
// 		tmp, ok := svr.Programs[programID]
// 		svr.mutex.RUnlock()
// 		if !ok {
// 			return nil, ErrorProgramNotFound
// 		}
// 		channelTmp, ok := tmp.Channels[channelID]
// 		if !ok {
// 			return nil, ErrorChannelNotFound
// 		}

// 		if len(channelTmp.source.sdp) > 0 {
// 			return channelTmp.source.sdp, nil
// 		}
// 		// why sleep?
// 		time.Sleep(50 * time.Millisecond)
// 	}
// 	return nil, ErrorProgramNotFound
// }

//ChannelAdd add stream
func (svr *ServerST) ChannelAdd(ctx context.Context, programID string, channelID string, ch *ChannelST) error {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	if prgm, ok := svr.Programs[programID]; ok {
		if _, ok := prgm.Channels[channelID]; ok {
			return ErrorChannelAlreadyExists
		}
	}

	svr.Programs[programID].Channels[channelID] = ch

	if !ch.OnDemand {
		go svr.ChannelRun(ctx, programID, channelID)
	}

	return nil
}

//ChannelDelete stream
func (svr *ServerST) ChannelDelete(programID string, channelID string) error {
	svr.ChannelStop(programID, channelID)

	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	if prgm, ok := svr.Programs[programID]; ok {
		if ch, ok := prgm.Channels[channelID]; ok {
			for cid, _ := range ch.clients {
				svr.ClientDelete(programID, channelID, cid)
			}

			svr.Programs[programID].Channels[channelID].clients = nil
			svr.Programs[programID].Channels[channelID].hlsSegmentBuffer = nil
			delete(svr.Programs[programID].Channels, channelID)
			return nil
		}
	}
	return ErrorChannelNotFound
}

//ChannelUpdate edit stream
func (svr *ServerST) ChannelUpdate(ctx context.Context, programID string, channelID string, ch *ChannelST) error {
	svr.ChannelStop(programID, channelID)

	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	if prgm, ok := svr.Programs[programID]; ok {
		if _, ok := prgm.Channels[channelID]; ok {
			svr.Programs[programID].Channels[channelID] = ch
			if !ch.OnDemand {
				go svr.ChannelRun(ctx, programID, channelID)
			}
			return nil
		}
	}
	return ErrorChannelNotFound
}

//ChannelGet get stream
func (svr *ServerST) ChannelGet(programID string, channelID string) (*ChannelST, error) {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	if prgm, ok := svr.Programs[programID]; ok {
		if ch, ok := prgm.Channels[channelID]; ok {
			return ch, nil
		}
	}
	return nil, ErrorChannelNotFound
}

//ChannelExist check stream exist
func (svr *ServerST) ChannelExist(programID string, channelID string) bool {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	if prgm, ok := svr.Programs[programID]; ok {
		if _, ok := prgm.Channels[channelID]; ok {
			return true
		}
	}
	return false
}

//ChannelGet get stream
func (svr *ServerST) ChannelFind(url string) (string, string, error) {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	for prgmID, st := range svr.Programs {
		for chID, ch := range st.Channels {
			if ch.URL == url {
				return prgmID, chID, nil
			}
		}
	}
	return "", "", ErrorChannelNotFound
}

//ChannelReload reload stream
func (svr *ServerST) ChannelReload(uuid string, channelID string) error {
	return nil
}

//ChannelCount count online stream channel.
func (svr *ServerST) ChannelCount() int {
	var cnt int
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	for _, st := range svr.Programs {
		cnt = cnt + len(st.Channels)
	}

	return cnt
}

//ChannelCountRunning count online stream channel.
func (svr *ServerST) ChannelCountRunning() int {
	var cnt int
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	for _, st := range svr.Programs {
		for _, ch := range st.Channels {
			if ch.source.status == STREAM_ONLINE {
				cnt++
			}
		}
	}

	return cnt
}
