package main

import (
	"net"
	"fmt"
	"sync"
	"google.golang.org/grpc"
	"grpcChatRoom/proto/chatRoom"
	"context"
)

type connection struct {
	stream chatRoom.ChatRoom_CreateStreamServer
	id string
	active bool
	err chan error
}

type server struct {
	connections []connection
}

func (s *server) CreateStream(connect *chatRoom.Connect, stream chatRoom.ChatRoom_CreateStreamServer) error {
	conn := connection {
		stream: stream,
		id: connect.User.Id,
		active: connect.Active,
		err: make(chan error),
	}
	s.connections = append(s.connections, conn)
	return <-conn.err
}

func (s *server) BroadcastMsg(context context.Context, message *chatRoom.Message) (*chatRoom.Close, error) {
	var wg sync.WaitGroup
	for _, conn := range s.connections {
		wg.Add(1)
		err := conn.stream.Send(message)
		if err != nil {
			fmt.Println("Connection send error: ", conn)
			conn.err <- err
			wg.Done()
		}
	}
	go func(){
		wg.Wait()
	}()
	return &chatRoom.Close{}, nil
}

func main() {
	
	lis, err := net.Listen("tcp", ":7160")
	if err != nil{
		panic(err)
	}

	chatServer := server{[]connection{}}
	svr := grpc.NewServer()
	chatRoom.RegisterChatRoomServer(svr, &chatServer)
	
	fmt.Println("Start Listining on port 7160")
	err = svr.Serve(lis)
	if err != nil {
		panic(err)
	}
}