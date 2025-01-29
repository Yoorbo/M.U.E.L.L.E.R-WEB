package main

import (
	"MUELLER/MUELLER"
	"fmt"
)

func main() {
	server := MUELLER.NewServer(2000)

	server.AddRoute("GET", "/students", func() string {
		return "GET request received: students"
	})

	server.AddRoute("POST", "/students", func() string {
		return "POST request received: students"
	})

	server.AddRoute("PUT", "/students", func() string {
		return "PUT request received: students"
	})

	server.AddRoute("DELETE", "/students", func() string {
		return "DELETE request received: students"
	})

	server.AddRoute("GET", "/schools", func() string {
		return "GET request received: schools"
	})

	fmt.Println("Starting server...")
	server.Start()
}
