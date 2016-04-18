package main

import (

	"net/http"

	"com.grid/chsen/chsen/socket"
	"runtime"
)

func main() {

	runtime.GOMAXPROCS(4)
	server := socket.NewServer(nil)
 	go server.Run()
	http.ListenAndServe("localhost:8080", nil)
}
