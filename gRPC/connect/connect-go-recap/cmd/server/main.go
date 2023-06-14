package main

import (
	"context"
	greetv1 "example/pkg/api/greet/v1"
	greetv1connect "example/pkg/api/greet/v1/v1connect" // greet + v1 + v1connect で重複削除したものを識別子として利用
	"fmt"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
	grpchealth "github.com/bufbuild/connect-grpchealth-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type GreetServer struct{}

func (s *GreetServer) Greet(ctx context.Context, req *connect.Request[greetv1.GreetRequest]) (*connect.Response[greetv1.GreetResponse], error) {
	log.Println("Request headers: ", req.Header())
	res := connect.NewResponse(
		&greetv1.GreetResponse{
			Greeting: fmt.Sprintf("Hello, %s!", req.Msg.Name),
		},
	)
	res.Header().Set("Greet-Version", "v1")
	return res, nil
}

func main() {
	greeter := &GreetServer{}
	mux := http.NewServeMux()
	path, handler := greetv1connect.NewGreetServiceHandler(greeter)
	mux.Handle(path, handler)
	reflector := grpcreflect.NewStaticReflector(
		"greet.v1.GreetService",
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	// Many tools still expect the older version of the server reflection API, so
	// most servers should mount both handlers.
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	checker := grpchealth.NewStaticChecker(
		"greet.v1.GreetService",
	)
	mux.Handle(grpchealth.NewHandler(checker))
	http.ListenAndServe(
		"localhost:8080",
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
}
