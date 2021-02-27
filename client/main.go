package main

import (
	"crypto/sha256"
	"flag"
	"time"
	"fmt"
	"os"
	"bufio"
	"sync"
	"grpcChatRoom/proto/chatRoom"
	"google.golang.org/grpc"
	"context"
)

var client chatRoom.ChatRoomClient

func connect(client chatRoom.ChatRoomClient, name, timestamp string) {
	var wg sync.WaitGroup
	done := make(chan int)
	id := fmt.Sprintf("%x",sha256.Sum256([]byte(fmt.Sprint(name, timestamp))))[:5]
	user := &chatRoom.User{
		Id: id ,
		Name: name,
	}
	connect := &chatRoom.Connect{
		Active: true,
		User: user,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		stream, err := client.CreateStream(context.Background(), connect)
		if err != nil {
			panic(err)
		}
		for {
			msg, err := stream.Recv()
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s %s : %s\n", msg.Timestamp, msg.Id, msg.Content)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println(connect.User.Name, " says: ")
		input := bufio.NewScanner(os.Stdin)	
		for input.Scan() {
			content := input.Text()
			msg := &chatRoom.Message{Id: id, Content: content, Timestamp: time.Now().String()}
			_, err := client.BroadcastMsg(context.Background(), msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
	
	go func() {
		wg.Wait()
	}()
	<-done
}

func main() {
	
	grpcConn, err := grpc.Dial("192.168.31.87:7160", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	timestamp := time.Now().String()
	
	//Add a flag for user to input name. ex. go run main.go -n Eric
	name := flag.String("n", "", "Username")
	flag.Parse()
	fmt.Println(fmt.Sprint(*name, timestamp))
	client := chatRoom.NewChatRoomClient(grpcConn)
	connect(client, *name, timestamp)
	
}