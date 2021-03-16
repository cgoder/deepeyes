package gss

import "errors"

//Default stream errors
var (
	Success                   = "success"
	ErrorProgramNotFound      = errors.New("stream not found")
	ErrorProgramAlreadyExists = errors.New("stream already exists")
	ErrorChannelAlreadyExists = errors.New("stream channel already exists")
	ErrorChannelNotFound      = errors.New("stream channel not found")
	ErrorChannelNoSource      = errors.New("stream channel no source stream")
	ErrorClientNotFound       = errors.New("stream client not found")
	ErrorStreamCodecNotFound  = errors.New("stream channel codec not ready, possible stream offline")
	ErrorStreamCodecUpdate    = errors.New("stream channel codec update")
	ErrorStreamNotHLSSegments = errors.New("stream hls not ts seq found")
	ErrorStreamNoVideo        = errors.New("stream no video")
	ErrorStreamNoClients      = errors.New("stream no clients")
	ErrorStreamRestart        = errors.New("stream restart")
	ErrorStreamStopCoreSignal = errors.New("stream stop core signal")
	ErrorStreamStopRTSPSignal = errors.New("stream stop rtsp signal")
	ErrorStreamsLen0          = errors.New("streams len zero")
	ErrorStreamAlreadyRunning = errors.New("stream already running")
)
