package gss

import (
	"context"
	"time"

	"github.com/cgoder/vdk/av"
	log "github.com/sirupsen/logrus"
)

var (
	timeoutClientCheck int = 60
	timeoutAvReadCheck int = 20

	lenAvPacketQueue     int = 100
	lenClientSignalQueue int = 1
)

//StreamCodecGet get stream codec storage or wait
func (svr *ServerST) StreamCodecGet(streamID string, channelID string) ([]av.CodecData, error) {
	svr.mutex.Lock()
	ch, ok := svr.Programs[streamID].Channels[channelID]
	svr.mutex.Unlock()
	if !ok {
		return nil, ErrorChannelNotFound
	}

	if ch.source.status == STREAM_ONLINE && ch.source.avCodecs != nil {
		// log.WithFields(log.Fields{
		// 	"module":  "Stream",
		// 	"stream":  streamID,
		// 	"channel": channelID,
		// 	"func":    "StreamCodecGet",
		// 	"call":    "chan.updated",
		// }).Debugln("Got old codec!")
		return ch.source.avCodecs, nil
	} else {
		log.WithFields(log.Fields{
			"module":  "stream",
			"stream":  streamID,
			"channel": channelID,
			"func":    "StreamCodecGet",
			"call":    "Get Stream codec ...",
		}).Debugf("Get Stream codec ...")

		t1 := time.Now().UTC()

		//wait for stream status update, recv cond.Broadcast().
		ch.cond.L.Lock()
		defer ch.cond.L.Unlock()
		ch.cond.Wait()

		svr.mutex.Lock()
		ch, ok := svr.Programs[streamID].Channels[channelID]
		svr.mutex.Unlock()
		if !ok {
			return nil, ErrorChannelNotFound
		}

		t2 := time.Now().UTC().Sub(t1)
		log.WithFields(log.Fields{
			"module":  "stream",
			"stream":  streamID,
			"channel": channelID,
			"func":    "StreamCodecGet",
			"call":    "chan.updated",
		}).Debugf("Got Stream codec update! cost:%v", t2.String())

		return ch.source.avCodecs, nil
	}

}

//StreamStatusUpdate change stream status
func (svr *ServerST) StreamStatusUpdate(streamID string, channelID string, status int) error {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	ch, ok := svr.Programs[streamID].Channels[channelID]
	if !ok {
		return ErrorChannelNotFound
	}

	ch.source.status = status

	return nil
}

//StreamCodecUpdate update stream codec storage
func (svr *ServerST) StreamCodecUpdate(streamID string, channelID string, avcodec []av.CodecData, sdp []byte) error {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	ch, ok := svr.Programs[streamID].Channels[channelID]
	if !ok {
		return ErrorChannelNotFound
	}

	ch.source.sdp = sdp
	ch.source.avCodecs = avcodec

	return nil
}

//StreamRunAll run all stream/channel.
func (svr *ServerST) StreamRunAll() {
	ctx := context.Background()

	for prgmID, program := range svr.Programs {
		for chID, ch := range program.Channels {
			if !ch.OnDemand {
				go svr.StreamRun(ctx, prgmID, chID)
			}
		}
	}

}

//StreamRun run stream/channel.
func (svr *ServerST) StreamRun(ctx context.Context, streamID string, channelID string) error {
	// get stream channel
	// ch, err := svr.ChannelGet(streamID, channelID)
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	ch, ok := svr.Programs[streamID].Channels[channelID]
	if !ok {
		log.WithFields(log.Fields{
			"module":  "streaming",
			"stream":  streamID,
			"channel": channelID,
			"func":    "ChannelRun",
			"call":    "ChannelGet",
		}).Errorln(ErrorChannelNotFound)
		return ErrorChannelNotFound
	}

	// channle is streaming?
	if ch.source.status == STREAM_ONLINE {
		log.WithFields(log.Fields{
			"module":  "streaming",
			"stream":  streamID,
			"channel": channelID,
			"func":    "ChannelRun",
			"call":    "channel.Status",
		}).Infoln("channel is streaming...")

		return nil
	}

	ch.source.status = STREAM_ONLINE

	log.WithFields(log.Fields{
		"module":  "core",
		"stream":  streamID,
		"channel": channelID,
		"func":    "ChannelRun",
		"call":    "ChannelRun",
	}).Infoln("Run stream-> ", streamID, channelID)

	// real stream run
	go svr.readPktByProtocol(ctx, streamID, channelID, ch)

	return nil

}
