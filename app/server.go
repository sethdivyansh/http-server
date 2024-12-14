package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go response(conn)
	}

}

func response(conn net.Conn) {
	buff := make([]byte, 1024)
	n, _ := conn.Read(buff)
	request := string(buff[:n])

	parts := strings.SplitN(request, "\r\n", -1)

	var reqData []string
	for _, part := range parts {
		reqData = append(reqData, strings.Split(part, " ")...)
	}

	requestType := reqData[0]
	route := reqData[1]
	userAgent := reqData[6]

	var res string

	if requestType == "GET" {
		if route == "/" {
			res = "HTTP/1.1 200 OK\r\n\r\n"

		} else if strings.Split(route, "/")[1] == "echo" {
			message := strings.Split(route, "/")[2]
			//res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)
			if len(reqData) > 9 {
				if reqData[10] == "gzip" {
					res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\n\r\n")
				} else {
					res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\n")
				}
			} else {
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)
			}
		} else if strings.Split(route, "/")[1] == "user-agent" {
			//re := regexp.MustCompile(`User-Agent:\s([^\s]+)`)
			//userAgent := re.FindStringSubmatch(request)[1]
			res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)

		} else if strings.Split(route, "/")[1] == "files" {
			fileName := strings.Split(route, "/")[2]
			dir := os.Args[2]
			data, err := os.ReadFile(dir + fileName)

			if err != nil {
				res = "HTTP/1.1 404 Not Found\r\n\r\n"
			} else {
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
			}

		} else {
			res = "HTTP/1.1 404 Not Found\r\n\r\n"
		}
	} else if requestType == "POST" {
		reqBody := parts[len(parts)-1]
		if strings.Split(route, "/")[1] == "files" {
			fileName := strings.Split(route, "/")[2]
			dir := os.Args[2]
			fileData := []byte(strings.Trim(reqBody, "\x00"))

			if err := os.WriteFile(dir+fileName, fileData, 0644); err == nil {
				fmt.Println(string(fileData))
				res = "HTTP/1.1 201 Created\r\n\r\n"
			} else {
				res = "HTTP/1.1 404 Not found\r\n\r\n"
			}

		}
	} else {
		fmt.Println("Unknown request type: ", requestType)
	}
	conn.Write([]byte(res))
}
