package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
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
	stream, err := c.GetUpdates(context.Background(), info)
	if err != nil {
		log.Fatalf("Unable to monitor server: %v", err)
	}

	// We will keep listening for updates from the server
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Server stream closed at %d", client.time)
			break
		}
		if err != nil {
			log.Fatalf("Error recieving an update: %v", err)
		}

		// sync client time with server time
		client.syncTime(message.Time)

		// print message
		log.Printf("%s: %s (time: %d)", message.Sender, message.Text, message.Time)
	}
}

// function to sync client with server time - maybe not needed
func (c *Client) syncTime(serverTime int32) {
	c.time = max(c.time, serverTime)
}

func main() {
	clientName := os.Args[1]
	info := &pb.ClientInfo{Name: clientName}

	// Establish connection
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pb.NewChittyChatClient(conn)

	// Join chat
	_, err = c.JoinChat(context.Background(), info)
	if err != nil {
		log.Fatalf("couldnt join chat: %v", err)
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

		if text == ".exit" {
			_, err = c.LeaveChat(context.Background(), info)
			if err != nil {
				log.Fatalf("couldnt exit chat: %v", err)
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

func setLog() {
	f, err := os.OpenFile("clients.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
}
