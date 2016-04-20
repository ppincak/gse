package main

import (

	//"net/http"

	//"com.grid/chsen/chsen/socket"
	"runtime"

	"com.grid/chsen/gsio/socket"


)

type T struct {
	Event	string 	`json:"event"`
	Data    interface{}  `json:"data"`
}

func main() {

	runtime.GOMAXPROCS(4)
 	socket.NewServer(nil)



	//http.ListenAndServe("localhost:8080", nil)
}
