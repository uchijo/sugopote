package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {return true},
}

var addr = "localhost:1323"

func relay(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err !=nil {
		log.Print("upgrade: ", err)
	}
	defer c.Close()

	connection := Connection{}
	aggregator.SetConnection(&connection)

	log.Println("upgrade done.")

	loop:
	for {
		// Read
		var msg = []any{}
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Printf("failed to read initial message: %v", err)
			break
		}

		// 変なのをはじく
		if len(msg) < 1 {
			c.WriteMessage(websocket.TextMessage, []byte("invalid message!"))
			break
		}

		switch msg[0] {
		// aggregatorに流す
		case "EVENT":
			fmt.Printf("[event]: %v\n", msg[1])
			event, ok := msg[1].(map[string]any)
			if !ok {
				c.WriteMessage(websocket.TextMessage, []byte("[\"NOTICE\", \"error\"]"))
			}
			id, ok := event["id"]
			if !ok {
				c.WriteMessage(websocket.TextMessage, []byte("[\"NOTICE\", \"error\"]"))
			}
			bytes, _ := json.Marshal(msg[1])
			aggregator.SendEvent(string(bytes))
			resp := fmt.Appendf(nil, `["OK","%v",true,""]`, id)
			fmt.Printf("resp: %v\n", string(resp))
			err := c.WriteMessage(websocket.TextMessage, resp)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}

			err = c.Close()
			if err != nil {
				fmt.Printf("failed to close ws connection: %v\n", err)
			} else {
				fmt.Println("successfully closed ws connection.")
			}
			break loop
		case "REQ":
			fmt.Printf("[req]: %v\n", msg[1])
			parsed, ok := (msg[1].(string))
			if !ok {
				fmt.Printf("failed to parse subscriptionId!")
				break
			}
			s := &Sink{
				id:     parsed,
				socket: c,
			}
			s.SendEOSE()
			connection.Subscribe(s)
		}
	}
}

func main() {
	http.HandleFunc("/", relay)
	log.Fatal(http.ListenAndServe(addr, nil))
}
