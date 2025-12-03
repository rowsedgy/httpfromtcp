package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("========Connection accepted========")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)

		fmt.Println("=======Connection terminated=======")

	}

}
