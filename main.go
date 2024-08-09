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

func heartbeat(heartbeat_interval int, c *websocket.Conn) {

	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(heartbeat_interval))
			if err := c.WriteJSON(map[string]interface{}{"op": 1,
				"d": nil}); err != nil {
				fmt.Println("Heartbeat ", err)
			}
		}
	}()
}
func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Load", err)
	}

	token := os.Getenv("TOKEN")
	fmt.Println(token)
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

	for {
		_, content, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read fail ", err)
			return
		}
		fmt.Println(string(content))

		data := map[string]interface{}{}
		_ = json.Unmarshal(content, &data)
		var opCode float64
		if data["op"] != nil {
			opCode = data["op"].(float64)
		}
		fmt.Println(opCode)
		switch opCode {
		case 0:
			fmt.Println("Success?? ", string(content))
		case 10:
			if !ready {
				d := data["d"].(map[string]interface{})
				heartbeat_interval := int(d["heartbeat_interval"].(float64))
				heartbeat(heartbeat_interval, conn)
			}
			ready = true
			err := conn.WriteJSON(map[string]interface{}{
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
			if err != nil {
				fmt.Println("Intents err ", err)
			}

		}

	}
}
