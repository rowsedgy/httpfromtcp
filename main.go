package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		ch := getLinesChannel(conn)
		for line := range ch {
			fmt.Println(line)
		}

		fmt.Println("=======Connection terminated=======")

	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer f.Close()
		defer close(ch)

		currentLine := ""

		for {
			buf := make([]byte, 8)
			n, err := f.Read(buf)
			if err == io.EOF {
				break
			}

			data := string(buf[:n])

			parts := strings.Split(data, "\n")

			for i := 0; i < len(parts)-1; i++ {
				currentLine += parts[i]
				ch <- currentLine
				currentLine = ""
			}

			currentLine += parts[len(parts)-1]
		}

		if currentLine != "" {
			ch <- currentLine
		}

	}()
	return ch
}
