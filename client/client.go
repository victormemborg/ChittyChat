package main

import (
	"bufio"
	"context"
	"fmt"
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
	for {
		r, err := c.GetUpdates(context.Background(), info)
		if err != nil {
			log.Printf("could not get update from server")
		}
		client.syncTime(r.Time)
		log.Printf(r.Sender + " : " + r.Text)
	}
}

// function to sync client with server time - maybe not needed
func (c *Client) syncTime(serverTime int32) {
	if serverTime > c.time {
		c.time = serverTime
	}
	c.time++
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

	// Make scanner
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

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
