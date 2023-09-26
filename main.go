package main

import (
	"fmt"
	"log"

	"github.com/olifady/proglog/internal/server"
)

func main() {
	svr := server.NewServer(":3000")
	fmt.Println("Starting server at :3000")
	log.Fatal(svr.ListenAndServe())
}
