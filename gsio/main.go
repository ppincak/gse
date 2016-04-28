package main

import (
	"com.grid/chsen/gsio/socket"
	"fmt"
	"net/http"

	"runtime"
)

func serveIndexHtml(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "C:\\Users\\ppincak\\go\\src\\com.grid\\chsen\\gsio\\index.html")
}

func connect(client *socket.Client) {
	fmt.Println("connected")
}

func main() {

	runtime.GOMAXPROCS(4)

 	server := socket.NewServer(nil, nil)
	n, _ := server.AddNamespace("/comments")
	n.AddConnectListener(func(client * socket.SocketClient) {
		fmt.Println("connected")
	})
	n.Listen("click", func(client * socket.SocketClient, data []interface{}) {

	})

	server.Run()

	http.HandleFunc("/index", serveIndexHtml)
	http.HandleFunc("/socket.io/", server.ServeWebSocket)

	http.ListenAndServe("localhost:8080", nil)

	/*server.AddDisconnectListener(func(client *socket.Client) {
		fmt.Println("disconnected")
	})*/
}