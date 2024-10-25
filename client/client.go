package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	pb "github.com/victormemborg/ChittyChat/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	name string
	time int32
}

func monitorServer(c pb.ChittyChatClient, info *pb.ClientInfo, client *Client) {
	stream, err := c.Listen(context.Background(), info)
	if err != nil {
		fmt.Printf("Unable to monitor server: %v", err)
	}

	// We will keep listening for updates from the server
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("Server stream closed at %d", client.time)
			break
		}
		if err != nil {
			fmt.Printf("Error recieving an update: %v", err)
		}

		// sync client time with other client time
		client.time++
		client.syncTime(message.Time)

		// print message
		fmt.Printf("%s: %s (%s's time: %d)\n", message.Sender, message.Text, client.name, client.time)
	}
}

// function to sync client with server time - maybe not needed
func (c *Client) syncTime(otherClientTime int32) {
	c.time = max(c.time, otherClientTime)
}

func main() {
	clientName := os.Args[1]
	info := &pb.ClientInfo{Name: clientName}

	// Establish connection
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("did not connect: %v", err)
	}
	c := pb.NewChittyChatClient(conn)

	// Join chat
	_, err = c.JoinChat(context.Background(), info)
	if err != nil {
		fmt.Printf("couldnt join chat: %v", err)
	}

	// Initialize client
	client := &Client{name: clientName, time: 0}
	go monitorServer(c, info, client)

	// Make scanner
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) > 128 {
			fmt.Println("Error: Message is too long")
			continue
		}

		if text == ".debug" {
			fmt.Println(client.time)
			continue
		}

		if text == ".exit" {
			client.time++
			info.ClientTime = client.time
			_, err = c.LeaveChat(context.Background(), info)
			if err != nil {
				fmt.Printf("couldnt exit chat: %v", err)
			}
			break
		}
		// increment client time before sending message
		client.time++

		message := &pb.Message{
			Sender: client.name,
			Text:   text,
			Time:   client.time,
		}

		_, err = c.PublishMessage(context.Background(), message)
		if err != nil {
			fmt.Println("An error occurred while sending message")
		}
	}
}
