package main

import (
	"com.grid/gse/socket"
	"fmt"
	"net/http"

	"runtime"
	"time"
)

func serveIndexHtml(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "C:\\Users\\ppincak\\go\\src\\com.grid\\gse\\index.html")
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
	n.Listen("click", func(client * socket.SocketClient, data interface{}) {
		fmt.Println("click received", data)
	})
	server.Run()

	server.Listen("dump", func(client * socket.SocketClient, data interface{}) {
		fmt.Println("dump received", data)
	})

	go func(server *socket.Server) {
		dur, err := time.ParseDuration("1s")
		if err != nil {
			return
		}

		for {
			server.SendEvent("click", []string {"hello"})
			time.Sleep(dur)
		}
	}(server)

	go func(n *socket.Namespace) {
		dur, err := time.ParseDuration("1s")
		if err != nil {
			return
		}

		for {
			n.SendEvent("flick", []string {"some data"})
			time.Sleep(dur)
		}
	}(n)

	http.HandleFunc("/index", serveIndexHtml)
	http.HandleFunc("/socket.io/", server.ServeWebSocket)

	http.ListenAndServe("localhost:8080", nil)

	/*server.AddDisconnectListener(func(client *socket.Client) {
		fmt.Println("disconnected")
	})*/
}