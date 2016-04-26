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

type Broker struct {
	conc        		chan *socket.Client
	disc        		chan *socket.Client
	evtc        		chan *socket.Event
}

func main() {

	runtime.GOMAXPROCS(4)

 	server := socket.NewServer(nil, nil)
	go server.Run()

	http.HandleFunc("/index", serveIndexHtml)
	http.HandleFunc("/ws", server.ServeWebSocket)

	broker := &Broker{
		conc: make(chan *socket.Client),
		disc: make(chan *socket.Client),
		evtc: make(chan *socket.Event),
	}

	go func(broker *Broker) {
		for {
			select {
				case client := <- broker.conc:
					fmt.Printf("connected %s", client.GetSessionId())
				case client := <- broker.disc:
					fmt.Printf("disconnected %s", client.GetSessionId())
				case evt := <- broker.evtc:
					fmt.Println(evt.Data)
			}
		}
	}(broker)

	server.AddConnectListener(broker.conc)
	server.AddDisconnectListener(broker.disc)
	server.Listen("click", broker.evtc)
	http.ListenAndServe("localhost:8080", nil)

	/*server.AddDisconnectListener(func(client *socket.Client) {
		fmt.Println("disconnected")
	})*/
}