package api

import (
	"github.com/cgoder/deepeyes/gss"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

//HTTPAPIServerStreams function return stream list
func HTTPAPIServerStreams(c *gin.Context) {
	c.IndentedJSON(200, Message{Status: 1, Payload: service.ProgramList()})
}

//HTTPAPIServerStreamsMultiControlAdd function add new stream's
func HTTPAPIServerStreamsMultiControlAdd(c *gin.Context) {
	var payload gss.ServerST
	err := c.BindJSON(&payload)
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"func":   "HTTPAPIServerStreamsMultiControlAdd",
			"call":   "BindJSON",
		}).Errorln(err.Error())
		return
	}
	if payload.Programs == nil || len(payload.Programs) < 1 {
		c.IndentedJSON(400, Message{Status: 0, Payload: gss.ErrorStreamsLen0.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"func":   "HTTPAPIServerStreamsMultiControlAdd",
			"call":   "len(payload)",
		}).Errorln(gss.ErrorStreamsLen0.Error())
		return
	}
	var resp = make(map[string]Message)
	var FoundError bool
	for k, v := range service.Programs {
		err = service.ProgramAdd(k, v)
		if err != nil {
			log.WithFields(log.Fields{
				"module": "http_stream",
				"stream": k,
				"func":   "HTTPAPIServerStreamsMultiControlAdd",
				"call":   "StreamAdd",
			}).Errorln(err.Error())
			resp[k] = Message{Status: 0, Payload: err.Error()}
			FoundError = true
		} else {
			resp[k] = Message{Status: 1, Payload: gss.Success}
		}
	}
	if FoundError {
		c.IndentedJSON(200, Message{Status: 0, Payload: resp})
	} else {
		c.IndentedJSON(200, Message{Status: 1, Payload: resp})
	}
}

//HTTPAPIServerStreamsMultiControlDelete function delete stream's
func HTTPAPIServerStreamsMultiControlDelete(c *gin.Context) {
	var payload []string
	err := c.BindJSON(&payload)
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"func":   "HTTPAPIServerStreamsMultiControlDelete",
			"call":   "BindJSON",
		}).Errorln(err.Error())
		return
	}
	if len(payload) < 1 {
		c.IndentedJSON(400, Message{Status: 0, Payload: gss.ErrorStreamsLen0.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"func":   "HTTPAPIServerStreamsMultiControlDelete",
			"call":   "len(payload)",
		}).Errorln(gss.ErrorStreamsLen0.Error())
		return
	}
	var resp = make(map[string]Message)
	var FoundError bool
	for _, key := range payload {
		err := service.ProgramDelete(key)
		if err != nil {
			log.WithFields(log.Fields{
				"module": "http_stream",
				"stream": key,
				"func":   "HTTPAPIServerStreamsMultiControlDelete",
				"call":   "StreamDelete",
			}).Errorln(err.Error())
			resp[key] = Message{Status: 0, Payload: err.Error()}
			FoundError = true
		} else {
			resp[key] = Message{Status: 1, Payload: gss.Success}
		}
	}
	if FoundError {
		c.IndentedJSON(200, Message{Status: 0, Payload: resp})
	} else {
		c.IndentedJSON(200, Message{Status: 1, Payload: resp})
	}
}

//HTTPAPIServerStreamAdd function add new stream
func HTTPAPIServerStreamAdd(c *gin.Context) {
	var payload gss.ProgramST
	err := c.BindJSON(&payload)
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"stream": c.Param("uuid"),
			"func":   "HTTPAPIServerStreamAdd",
			"call":   "BindJSON",
		}).Errorln(err.Error())
		return
	}
	err = service.ProgramAdd(c.Param("uuid"), &payload)
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"stream": c.Param("uuid"),
			"func":   "HTTPAPIServerStreamAdd",
			"call":   "StreamAdd",
		}).Errorln(err.Error())
		return
	}
	c.IndentedJSON(200, Message{Status: 1, Payload: gss.Success})
}

//HTTPAPIServerStreamEdit function edit stream
func HTTPAPIServerStreamEdit(c *gin.Context) {
	var payload gss.ProgramST
	err := c.BindJSON(&payload)
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"stream": c.Param("uuid"),
			"func":   "HTTPAPIServerStreamEdit",
			"call":   "BindJSON",
		}).Errorln(err.Error())
		return
	}
	err = service.ProgramUpdate(c.Param("uuid"), &payload)
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"stream": c.Param("uuid"),
			"func":   "HTTPAPIServerStreamEdit",
			"call":   "StreamEdit",
		}).Errorln(err.Error())
		return
	}
	c.IndentedJSON(200, Message{Status: 1, Payload: gss.Success})
}

//HTTPAPIServerStreamDelete function delete stream
func HTTPAPIServerStreamDelete(c *gin.Context) {
	err := service.ProgramDelete(c.Param("uuid"))
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"stream": c.Param("uuid"),
			"func":   "HTTPAPIServerStreamDelete",
			"call":   "StreamDelete",
		}).Errorln(err.Error())
		return
	}
	c.IndentedJSON(200, Message{Status: 1, Payload: gss.Success})
}

//HTTPAPIServerStreamDelete function reload stream
func HTTPAPIServerStreamReload(c *gin.Context) {
	err := service.ProgramReload(c.Param("uuid"))
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"stream": c.Param("uuid"),
			"func":   "HTTPAPIServerStreamReload",
			"call":   "StreamReload",
		}).Errorln(err.Error())
		return
	}
	c.IndentedJSON(200, Message{Status: 1, Payload: gss.Success})
}

//HTTPAPIServerStreamInfo function return stream info struct
func HTTPAPIServerStreamInfo(c *gin.Context) {
	info, err := service.ProgramGet(c.Param("uuid"))
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.WithFields(log.Fields{
			"module": "http_stream",
			"stream": c.Param("uuid"),
			"func":   "HTTPAPIServerStreamInfo",
			"call":   "StreamInfo",
		}).Errorln(err.Error())
		return
	}
	c.IndentedJSON(200, Message{Status: 1, Payload: info})
}
