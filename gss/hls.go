package gss

import (
	"sort"
	"strconv"
	"time"

	"github.com/cgoder/vdk/av"
)

//StreamHLSAdd add hls seq to buffer
func (srv *ServerST) StreamHLSAdd(uuid string, channelID string, val []*av.Packet, dur time.Duration) {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	if tmp, ok := srv.Programs[uuid]; ok {
		if channelTmp, ok := tmp.Channels[channelID]; ok {
			channelTmp.hlsSegmentNumber++
			channelTmp.hlsSegmentBuffer[channelTmp.hlsSegmentNumber] = &Segment{data: val, dur: dur}
			if len(channelTmp.hlsSegmentBuffer) >= 6 {
				delete(channelTmp.hlsSegmentBuffer, channelTmp.hlsSegmentNumber-6-1)
			}
			tmp.Channels[channelID] = channelTmp
			srv.Programs[uuid] = tmp
		}
	}
}

//StreamHLSm3u8 get hls m3u8 list
func (srv *ServerST) StreamHLSm3u8(uuid string, channelID string) (string, int, error) {
	srv.mutex.RLock()
	defer srv.mutex.RUnlock()
	if tmp, ok := srv.Programs[uuid]; ok {
		if channelTmp, ok := tmp.Channels[channelID]; ok {
			var out string
			//TODO fix  it
			out += "#EXTM3U\r\n#EXT-X-TARGETDURATION:4\r\n#EXT-X-VERSION:4\r\n#EXT-X-MEDIA-SEQUENCE:" + strconv.Itoa(channelTmp.hlsSegmentNumber) + "\r\n"
			var keys []int
			for k := range channelTmp.hlsSegmentBuffer {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			var count int
			for _, i := range keys {
				count++
				out += "#EXTINF:" + strconv.FormatFloat(channelTmp.hlsSegmentBuffer[i].dur.Seconds(), 'f', 1, 64) + ",\r\nsegment/" + strconv.Itoa(i) + "/file.ts\r\n"

			}
			return out, count, nil
		}
	}
	return "", 0, ErrorProgramNotFound
}

//StreamHLSTS send hls segment buffer to clients
func (srv *ServerST) StreamHLSTS(streamID string, channelID string, seq int) ([]*av.Packet, error) {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	ch, ok := srv.Programs[streamID].Channels[channelID]
	if !ok {
		return nil, ErrorChannelNotFound
	}

	if tmp, ok := ch.hlsSegmentBuffer[seq]; ok {
		return tmp.data, nil
	} else {
		return nil, ErrorChannelNotFound
	}
}

//StreamHLSFlush delete hls cache
func (srv *ServerST) StreamHLSFlush(streamID string, channelID string) {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	ch, ok := srv.Programs[streamID].Channels[channelID]
	if ok {
		ch.hlsSegmentBuffer = make(map[int]*Segment)
		ch.hlsSegmentNumber = 0
	}
}
