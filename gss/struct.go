package gss

import (
	"sync"
	"time"

	"github.com/cgoder/vdk/av"
	"github.com/cgoder/vdk/av/pubsub"
)

//player type
const (
	PLAY_MSE = iota
	PLAY_WEBRTC
	PLAY_HLS
	PLAY_HTTPFLV
	PLAY_RTSP
	PLAY_RTMP
)

//stream status type
const (
	STREAM_OFFLINE = iota
	STREAM_ONLINE
	STREAM_PAUSE
)

//signal type
const (
	SIGNAL_STREAM_UNKNOWN = iota
	SIGNAL_STREAM_STOP
	SIGNAL_STREAM_REFRESH
	SIGNAL_STREAM_AVCODEC_UPDATE
)

//StorageST main storage struct
type ServerST struct {
	mutex    sync.RWMutex
	Conf     ConfigST              `mapstructure:"conf"`
	Programs map[string]*ProgramST `mapstructure:"program"`
}

//ConfigST server storage section
type ConfigST struct {
	Debug        bool   `mapstructure:"debug"`
	LogLevel     string `mapstructure:"log_level"`
	HTTPPort     string `mapstructure:"http_port"`
	HTTPDir      string `mapstructure:"http_dir"`
	HTTPLogin    string `mapstructure:"http_login"`
	HTTPPassword string `mapstructure:"http_password"`
	HTTPDebug    bool   `mapstructure:"http_debug"`
	HTTPDemo     bool   `mapstructure:"http_demo"`
	RTSPPort     string `mapstructure:"rtsp_port"`
}

//ProgramST stream storage section
type ProgramST struct {
	UUID     string                `mapstructure:"uuid,omitempty"`
	Name     string                `mapstructure:"name,omitempty"`
	Channels map[string]*ChannelST `mapstructure:"channels,omitempty"`
}

type ChannelST struct {
	// channel uid
	UUID string `mapstructure:"name,omitempty"`
	Name string `mapstructure:"name,omitempty"`
	// channel source stream url
	URL string `mapstructure:"url,omitempty"`
	// channel source stream
	source AvStream
	// channel stream clients
	clients map[string]*ClientST
	// channel update. or codec update.
	cond *sync.Cond

	// opreation signal.
	signals chan int
	///////////////////////////////////////////////////

	// auto streaming flag. FALSE==auto. default:false.
	OnDemand bool `mapstructure:"on_demand,omitempty"`
	Debug    bool `mapstructure:"debug,omitempty"`
	///////////////////////////////////////////////////

	// HLS
	hlsSegmentBuffer map[int]*Segment
	hlsSegmentNumber int
}

//AvStream read data from source stream.
type AvStream struct {
	protocol int
	status   int
	avCodecs []av.CodecData
	sdp      []byte
	//ring buffer?
	avQue *pubsub.Queue
}

//ClientST client read avPkt from queue, write to chan.
type ClientST struct {
	UUID             string
	protocol         int
	signals          chan int
	outgoingAVPacket chan *av.Packet
}

//Segment HLS cache section
type Segment struct {
	dur  time.Duration
	data []*av.Packet
}
