package gss

import (
	"github.com/cgoder/vdk/av"
)

//ClientAdd Add New Client to Translations
func (svr *ServerST) ClientAdd(streamID string, channelID string, mode int) (string, chan *av.Packet, error) {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	ch, ok := svr.Programs[streamID].Channels[channelID]
	if !ok {
		return "", nil, ErrorChannelNotFound
	}

	//Generate UUID client
	cid := GenerateUUID()
	chAV := make(chan *av.Packet, lenAvPacketQueue)
	chSignal := make(chan int, lenClientSignalQueue)
	client := &ClientST{UUID: cid, protocol: mode, outgoingAVPacket: chAV, signals: chSignal}

	ch.clients[client.UUID] = client

	// log.WithFields(logrus.Fields{
	// 	"module":  "storageClient",
	// 	"stream":  streamID,
	// 	"channel": channelID,
	// 	"func":    "ClientAdd",
	// 	"call":    "ClientAdd",
	// }).Debugln("client Add ---> ")

	// log.Println(svr)

	return cid, chAV, nil

}

//ClientDelete Delete Client
func (svr *ServerST) ClientDelete(streamID string, channelID string, cid string) {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	if _, ok := svr.Programs[streamID]; ok {
		if _, ok := svr.Programs[streamID].Channels[channelID].clients[cid]; ok {
			delete(svr.Programs[streamID].Channels[channelID].clients, cid)
		}
	}
}

//ClientHas check is client ext
func (svr *ServerST) ClientHas(streamID string, channelID string) bool {
	defer svr.mutex.Unlock()
	ch, ok := svr.Programs[streamID].Channels[channelID]
	if !ok {
		return false
	}

	if len(ch.clients) > 0 {
		return true
	}
	return false
}

//ClientFind get client
func (svr *ServerST) ClientFind(clientID string) (string, string, error) {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	for prgmID, st := range svr.Programs {
		for chID, ch := range st.Channels {
			for cID, _ := range ch.clients {
				if cID == clientID {
					return prgmID, chID, nil
				}
			}
		}
	}
	return "", "", ErrorClientNotFound
}

//ClientHas check is client ext
func (svr *ServerST) ClientCount(streamID string, channelID string) int {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	ch, ok := svr.Programs[streamID].Channels[channelID]
	if !ok {
		return 0
	}

	return len(ch.clients)
}

//ClientCountAll count all clients
func (svr *ServerST) ClientCountAll() int {
	var cnt int
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	for _, st := range svr.Programs {
		for _, ch := range st.Channels {
			cnt = cnt + len(ch.clients)
		}
	}

	return cnt
}
