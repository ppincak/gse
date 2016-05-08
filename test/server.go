package main

import (
	"com.grid/gse/socket"
	"net/http"
	"github.com/gorilla/websocket"
)

const (
	serverUrl 	 = "localhost:9999"
	websocketUrl = "localhost:9999/ws"
)

func main() {
	server := socket.NewServer(nil, nil)
	chat, err := server.AddNamespace("/chat")

	if err != nil {

	}

	chat.AddConnectListener(func(client *socket.Client) {

	})

	chat.AddDisconnectListener(func(client *socket.Client) {

	})

	chat.Listen("message", func() {

	})

	chat.Listen("message:user", func() {

	})

	chat.Listen("quit", func() {

	});

	server.Run()

	http.HandleFunc("/ws", server.ServeWebSocket)
	go http.ListenAndServe(serverUrl, nil)

	con, resp, err := websocket.DefaultDialer.Dial(websocketUrl, http.Header{})
	if err != nil {

	}
	if resp != nil && con != nil {

	}

}

func addError() {

}