package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/victormemborg/ChittyChat/grpc"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedChittyChatServer
	time int32
}

type MessageQueue []*pb.Message

func (self *MessageQueue) Push(m *pb.Message) {
	*self = append(*self, m)
}

func (self *MessageQueue) Pop() *pb.Message {
	if 
	h := *self
	var m *pb.Message
	l := len(h)
	m, *self = h[0], h[1:l]
	return m
}

func NewQueue() *MessageQueue {
	return &MessageQueue{}
}

var chatMembers = make(map[string]bool)
var buffers = make(map[string]*MessageQueue)

func main() {
	setLog()
	s := &Server{}
	s.startServer()
}

func (s *Server) updateTime(clientTime int32) {
	if clientTime > s.time {
		s.time = clientTime
	}
	s.time++
}

func (s *Server) startServer() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Unable to start connection to server..")
	}

	pb.RegisterChittyChatServer(grpcServer, s)
	log.Printf("server listening at %v", listener.Addr())
	if grpcServer.Serve(listener) != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *Server) PublishMessage(_ context.Context, in *pb.Message) (*pb.Empty, error) {
	//log.Printf(in.Timestamp + " : " + in.Sender + " : " + in.Text)
	if !chatMembers[in.Sender] {
		return nil, fmt.Errorf("User %s is not in chat", in.Sender)
	}

	// update server time to client time
	s.updateTime(in.Time)

	for k, v := range buffers {
		if k == in.Sender {
			continue
		}
		v.Push(in)
	}

	return &pb.Empty{}, nil
}

func (s *Server) JoinChat(_ context.Context, in *pb.ClientInfo) (*pb.Empty, error) {
	if chatMembers[in.Name] {
		return nil, fmt.Errorf("User %s already in chat", in.Name)
	}
	buffers[in.Name] = NewQueue()

	log.Println(in.Name + " joined the chat")
	chatMembers[in.Name] = true
	return &pb.Empty{}, nil
}

func (s *Server) LeaveChat(_ context.Context, in *pb.ClientInfo) (*pb.Empty, error) {
	if !chatMembers[in.Name] {
		return nil, fmt.Errorf("User %s is not in chat", in.Name)
	}

	log.Println(in.Name + " left the chat")
	chatMembers[in.Name] = false
	return &pb.Empty{}, nil
}

func (s *Server) GetUpdates(_ context.Context, in *pb.ClientInfo) (*pb.Message, error) {
	m := buffers[in.Name].Pop()
	return m, nil
}
func setLog() {
	f, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
}
