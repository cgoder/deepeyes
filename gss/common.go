package gss

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/gops/agent"
)

//GenerateUUID function make random uuid for clients and stream
func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

//JsonFormat Json outupt.
func JsonFormat(v interface{}) string {
	// if msg == "" {
	// 	return ""
	// }
	var out bytes.Buffer

	bs, _ := json.Marshal(v)
	json.Indent(&out, bs, "", "\t")

	return out.String()
}

//DebugRuntime debug info
func DebugRuntime() {
	if err := agent.Listen(agent.Options{
		Addr:            "0.0.0.0:8048",
		ShutdownCleanup: true, // automatically closes on os.Interrupt
	}); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Hour)
}
