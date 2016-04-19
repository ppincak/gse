package main

import (

	//"net/http"

	//"com.grid/chsen/chsen/socket"
	"runtime"

	"com.grid/chsen/chsen/socket"

)

type T struct {
	Event	string 	`json:"event"`
	Data    interface{}  `json:"data"`
}

func main() {

	runtime.GOMAXPROCS(4)
   	server := socket.NewServer(nil)



 	server.Run()


	//http.ListenAndServe("localhost:8080", nil)
}
