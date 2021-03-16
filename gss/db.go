package gss

func (svr *ServerST) LoadDB() {
	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	for prgmID, programs := range svr.Programs {
		for chID, ch := range programs.Channels {
			tmpCh := ChannelNew(ch)
			svr.Programs[prgmID].Channels[chID] = tmpCh
		}
	}
}

func (svr *ServerST) SaveDB() {
	svr.SaveConfig()
}
