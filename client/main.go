package main

import (
	"Homework/chatpb"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Error connecting to the server: ", err)
	}
	defer conn.Close()

	client := chatpb.NewChatServiceClient(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Fatal("Error creating stream: ", err)
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatalf("failed to receive a message : %v", err)
			}
			log.Printf("Received message: %s from %s at %s", in.Message, in.Username, in.Timestamp)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Start chatting:")
	for scanner.Scan() {
		message := scanner.Text()
		err := stream.Send(&chatpb.ChatMessage{
			Username:  "Abduazim_Yusufov",
			Message:   message,
			Timestamp: time.Now().UTC().String(),
		})
		if err != nil {
			log.Fatal("error sending message: ", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading input: ", err)
	}
}
