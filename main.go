package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var ready bool = false
var ch = make(chan int)
var resumeGatewayUrl string
var sessionId string
var seq int

func heartbeat(heartbeat_interval int) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(heartbeat_interval))
	for {
		tm := <-ticker.C
		fmt.Println("Heartbeat", tm)
		ch <- 1
	}
}
func responseWriter(ws *websocket.Conn, token string) {
	for {
		opCode := <-ch
		fmt.Println(opCode)
		switch opCode {
		case 1:
			ws.WriteJSON(map[string]interface{}{
				"op": 1, "d": nil,
			})
		case 6:
			ws.WriteJSON((map[string]interface{}{
				"op": 6,
				"d": map[string]interface{}{
					"token":     token,
					"sessionId": sessionId,
					"seq":       seq,
				},
			}))
		case 7:
			ws.WriteJSON(map[string]interface{}{
				"op": 7,
				"d":  nil,
			})
		case 10:
			ws.WriteJSON(map[string]interface{}{
				"op": 2,
				"d": map[string]interface{}{
					"token":   token,
					"intents": 512,
					"properties": map[string]interface{}{
						"os":      "macos",
						"browser": "chrome",
						"device":  "macbook air",
					},
				},
			})
		}
	}

}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Load", err)
	}

	tkn := os.Getenv("TOKEN")
	u := url.URL{Scheme: "wss", Host: "gateway.discord.gg", Path: "/"}
	fmt.Println(u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	if err != nil {
		fmt.Println("Socket", err)
	}
	go responseWriter(conn, tkn)
	for {
		_, content, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read fail ", err)
			return
		}

		data := map[string]interface{}{}
		_ = json.Unmarshal(content, &data)
		var opCode float64
		if data["op"] != nil {
			opCode = data["op"].(float64)
		}
		seq = int(opCode)
		switch opCode {
		case 10:
			if !ready {
				d := data["d"].(map[string]interface{})
				heartbeat_interval := int(d["heartbeat_interval"].(float64))
				ch <- 10
				go heartbeat(heartbeat_interval)
			}
			ready = true
		case 6:
			ch <- 6
			u := url.URL{Scheme: "wss", Host: resumeGatewayUrl, Path: "/"}
			fmt.Println(u.String())
			conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				fmt.Println(err)
			}
		case 7:
			ch <- 7

		}
		status := data["t"]
		switch status {

		case "READY":
			fmt.Println("Gateway ready")
			d := data["d"].(map[string]interface{})
			sessionId = d["session_id"].(string)
			resumeGatewayUrl = d["resume_gateway_url"].(string)

		case "MESSAGE_CREATE":
			d := data["d"].(map[string]interface{})
			content := d["content"].(string)
			author := d["author"].(map[string]interface{})
			username := author["username"].(string)
			fmt.Println(username, " said ", content)
		default:
			fmt.Println(status)
		}

	}
}
