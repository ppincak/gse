package main

import (
	"log"
	"google.golang.org/grpc"

	"com.grid/chsen/gsio/queue"
	"golang.org/x/net/context"
	"fmt"
	"runtime"
)

const (
	address     = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := queue.NewQueueServiceClient(conn)

	runtime.GOMAXPROCS(4)

	go func() {
		for i := 0; i < 100000; i++ {
			_, err := client.AddRoom(context.Background(), &queue.AddRoomReq{
				RoomName: "Test",
				RoomCapacity: 10,
			})

			if err != nil {
				fmt.Println("err")
				continue
			}

			//fmt.Println("req", i)
		}

		fmt.Println("done")
	}();

	go func() {
		for i := 0; i < 100000; i++ {
			_, err := client.AddRoom(context.Background(), &queue.AddRoomReq{
				RoomName: "Test",
				RoomCapacity: 10,
			})

			if err != nil {
				fmt.Println("err")
				continue
			}

			//fmt.Println("req", i)
		}

		fmt.Println("done")
	}();


		for i := 0; i < 10000; i++ {
			_, err := client.AddRoom(context.Background(), &queue.AddRoomReq{
				RoomName: "Test",
				RoomCapacity: 10,
			})

			if err != nil {
				fmt.Println("err")
				continue
			}
		}

		fmt.Println("done")



}

