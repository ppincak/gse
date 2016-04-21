package main

import (
	"com.grid/chsen/gsio/socket"
	"fmt"
	"net/http"

)

func serveIndexHtml(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "C:\\Users\\ppincak\\go\\src\\com.grid\\chsen\\gsio\\index.html")
}

func connect(client *socket.Client) {
	fmt.Println("connected")
}

func main() {

 	server := socket.NewServer(nil, nil)
	go server.Run()

	http.HandleFunc("/index", serveIndexHtml)
	http.HandleFunc("/ws", server.ServeWebSocket)
	server.AddConnectListener(connect)
	server.AddDisconnectListener(func(client *socket.Client) {
		fmt.Println("disconnected")
	})
	server.Listen("click", func(client *socket.Client, data interface{}) {
		fmt.Println("event callback", data)
	})
	http.ListenAndServe("localhost:8080", nil)

	/*server.AddDisconnectListener(func(client *socket.Client) {
		fmt.Println("disconnected")
	})*/
}