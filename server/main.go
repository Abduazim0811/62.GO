package main

import (
	"Homework/chatpb"
	"Homework/storage"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type ChatServer struct {
	chatpb.UnimplementedChatServiceServer
	db      *sql.DB
	clients map[string]chatpb.ChatService_ChatServer
	mu      sync.Mutex
}

func NewServer(db *sql.DB) *ChatServer {
	return &ChatServer{
		clients: make(map[string]chatpb.ChatService_ChatServer),
		db:      db,
	}
}

func (s *ChatServer) Chat(stream chatpb.ChatService_ChatServer) error {
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return status.Error(codes.Internal, "could not get peer info")
	}

	clientID := p.Addr.String()
	network := p.LocalAddr.Network()
	fmt.Printf("connected [%s]:%s\n", network, clientID)

	s.mu.Lock()
	s.clients[clientID] = stream
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientID)
		s.mu.Unlock()
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		msg.IpAddress = clientID

		s.broadcastMessage(msg, clientID)
	}
}

func (s *ChatServer) SendMessage(ctx context.Context, in *chatpb.ChatMessage) (*chatpb.Empty, error) {
	_, err := s.db.Exec("INSERT INTO messages (username, message, timestamp) VALUES ($1, $2, $3)", in.Username, in.Message, in.Timestamp)
	if err != nil {
		return nil, err
	}
	return &chatpb.Empty{}, nil
}

func (s *ChatServer) GetMessages(in *chatpb.Empty, stream chatpb.ChatService_GetMessagesServer) error {
	rows, err := s.db.Query("SELECT username, message, timestamp FROM messages ORDER BY timestamp ASC")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var msg chatpb.ChatMessage
		if err := rows.Scan(&msg.Username, &msg.Message, &msg.Timestamp); err != nil {
			return err
		}
		if err := stream.Send(&msg); err != nil {
			return err
		}
	}
	return nil
}

func (s *ChatServer) broadcastMessage(msg *chatpb.ChatMessage, senderID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, client := range s.clients {
		if id == senderID {
			continue
		}

		if err := client.Send(msg); err != nil {
			log.Printf("errror sending message to client %v: %v", id, err)
		}
	}
}

const (
	serverURL   = "localhost:7777"
	postgresURL = "postgres://postgres:Abdu0811@localhost:5432/food_service?sslmode=disable"
	driverName  = "postgres"
)

func main() {
	db, err := storage.OpenSql(driverName, postgresURL)
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	defer db.Close()

	server := NewServer(db)

	listener, err := net.Listen("tcp", serverURL)
	if err != nil {
		log.Fatal("Error starting listener...", err)
	}

	grpcServer := grpc.NewServer()
	chatpb.RegisterChatServiceServer(grpcServer, server)

	log.Println("Server is listening on port:", serverURL)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
