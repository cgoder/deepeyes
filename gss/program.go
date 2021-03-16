package gss

import "context"

//ProgramList list all program
func (svr *ServerST) ProgramList() map[string]*ProgramST {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	tmp := make(map[string]*ProgramST)
	for pid, prgm := range svr.Programs {
		tmp[pid] = prgm
	}
	return tmp
}

//ProgramAdd add program
func (svr *ServerST) ProgramNew(name string) (*ProgramST, error) {
	var program ProgramST
	program.UUID = GenerateUUID()
	program.Name = name
	if program.Name == "" {
		program.Name = program.UUID
	}
	program.Channels = make(map[string]*ChannelST)

	return &program, nil
}

//ProgramAdd add program
func (svr *ServerST) ProgramAdd(uuid string, program *ProgramST) error {
	progID := uuid

	svr.mutex.Lock()
	_, ok := svr.Programs[progID]
	svr.mutex.Unlock()
	if ok {
		return ErrorProgramAlreadyExists
	}

	svr.ChannelStop(progID, "")

	svr.mutex.Lock()
	svr.Programs[progID] = program
	svr.mutex.Unlock()

	for chID, ch := range program.Channels {
		if !ch.OnDemand {
			go svr.ChannelRun(context.TODO(), progID, chID)
		}
	}

	svr.Programs[progID] = program
	err := svr.SaveConfig()
	if err != nil {
		return err
	}
	return nil
}

//ProgramDelete delete program
func (svr *ServerST) ProgramDelete(uuid string) error {
	progID := uuid

	svr.mutex.Lock()
	defer svr.mutex.Unlock()
	_, ok := svr.Programs[progID]
	if !ok {
		return ErrorProgramNotFound
	}

	//stop all stream
	svr.ChannelStop(progID, "")
	for chID, _ := range svr.Programs[progID].Channels {
		svr.ChannelDelete(progID, chID)
	}
	svr.Programs[progID].Channels = nil
	delete(svr.Programs, uuid)
	//release memory
	svr.Programs[progID] = nil

	err := svr.SaveConfig()
	if err != nil {
		return err
	}
	return nil
}

//ProgramUpdate update program
func (svr *ServerST) ProgramUpdate(uuid string, program *ProgramST) error {
	progID := uuid

	svr.mutex.Lock()
	_, ok := svr.Programs[progID]
	svr.mutex.Unlock()
	if !ok {
		return ErrorProgramNotFound
	}

	svr.ChannelStop(progID, "")

	svr.mutex.Lock()
	svr.Programs[progID] = program
	svr.mutex.Unlock()

	for _, ch := range program.Channels {
		if !ch.OnDemand {
			go svr.ChannelRun(context.TODO(), progID, ch.UUID)
		}
	}

	err := svr.SaveConfig()
	if err != nil {
		return err
	}
	return nil
}

//ProgramGet get program info
func (svr *ServerST) ProgramGet(uuid string) (*ProgramST, error) {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	if tmp, ok := svr.Programs[uuid]; ok {
		return tmp, nil
	}
	return nil, ErrorProgramNotFound
}

//ProgramStopAll stop all program stream
func (svr *ServerST) ProgramStopAll() {
	svr.ChannelStop("", "")
}

//ProgramReload reload program
func (svr *ServerST) ProgramReload(uuid string) error {
	return nil
}
