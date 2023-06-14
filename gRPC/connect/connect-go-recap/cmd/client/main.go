package main

import (
	"context"
	greetv1 "example/pkg/api/greet/v1"
	"example/pkg/api/greet/v1/v1connect"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
)

func main() {
	client := v1connect.NewGreetServiceClient(http.DefaultClient, "http://localhost:8080")
	res, err := client.Greet(context.Background(), connect.NewRequest(&greetv1.GreetRequest{Name: "ginoh"}))
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(res.Msg.Greeting)
}
