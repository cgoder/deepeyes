package gss

import (
	"context"

	"github.com/cgoder/vdk/av"
	log "github.com/sirupsen/logrus"
)

//PlayPre stream url.
//return clientid, avcode, read-only avpacket chan, and err.
func (svr *ServerST) PlayPre(ctx context.Context, url string, play_mod int) (string, []av.CodecData, <-chan *av.Packet, error) {
	// check stream status
	programID, channelID, err := svr.ChannelFind(url)
	if err != nil {
		log.WithFields(log.Fields{
			"module": "api",
			"func":   "PlayPre",
			"call":   "ChannelFind",
		}).Errorln(err.Error())
		return "", nil, nil, err
	}

	// streaming
	err = svr.ChannelRun(ctx, programID, channelID)
	if err != nil {
		log.WithFields(log.Fields{
			"module":  "api",
			"stream":  programID,
			"channel": channelID,
			"func":    "HTTPAPIServerStreamMSE",
			"call":    "StreamChannelRun",
		}).Errorln(err)
		return "", nil, nil, ErrorProgramNotFound
	}

	log.WithFields(log.Fields{
		"module":  "api",
		"stream":  programID,
		"channel": channelID,
		"func":    "HTTPAPIServerStreamMSE",
		"call":    "StreamChannelRun",
	}).Debugln("play stream ---> ", programID, channelID)

	// get stream av.Codec
	codecs, err := svr.StreamCodecGet(programID, channelID)
	if err != nil {
		log.WithFields(log.Fields{
			"module":  "http_mse",
			"stream":  programID,
			"channel": channelID,
			"func":    "HTTPAPIServerStreamMSE",
			"call":    "StreamCodecs",
		}).Errorln(err.Error())
		return "", nil, nil, ErrorStreamCodecNotFound
	}

	// add client/player
	cid, avCh, err := svr.ClientAdd(programID, channelID, play_mod)
	if err != nil {
		log.WithFields(log.Fields{
			"module":  "http_mse",
			"stream":  programID,
			"channel": channelID,
			"func":    "HTTPAPIServerStreamMSE",
			"call":    "ClientAdd",
		}).Errorln(err.Error())
		return "", nil, nil, ErrorStreamNoClients
	}

	return cid, codecs, avCh, nil
}

//PlayStop stop client pull stream.
func (svr *ServerST) PlayStop(clientID string) error {
	programID, channelID, err := svr.ClientFind(clientID)
	if err != nil {
		log.WithFields(log.Fields{
			"module": "api",
			"func":   "PlayStop",
			"call":   "ClientFind",
		}).Errorln(err.Error())
		return err
	}

	svr.ClientDelete(programID, channelID, clientID)

	return nil
}
