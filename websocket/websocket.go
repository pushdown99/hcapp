package websocket

import (
  "net/url"
  "log"
  "time"
  "encoding/json"
  "github.com/gorilla/websocket"
)

var Connected int = 0

type WsData struct {
	License   string
	Command   string
	Message   string
	Timestamp int64
}

func Connect(wshost string, license string) *websocket.Conn {
	for {
		u := url.URL{Scheme: "ws", Host: wshost, Path: "/"}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("[w] websocket dial error:", err)
			time.Sleep(1)
			continue
		}
		log.Printf("[w] connected to %s", u.String())
		Connected = 1
		b, _ := json.Marshal(WsData{License: license, Command: "Join", Message: "Hi", Timestamp: time.Now().Unix()})
		c.WriteMessage(websocket.TextMessage, b /*[]byte(b)*/)
		return c
	}
}

