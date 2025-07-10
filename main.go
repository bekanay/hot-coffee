package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "4000", "HTTP network address")
	dir := flag.String("dir", "", "Path to the directory")
	help := flag.Bool("help", false, "Print usage information")
	flag.Parse()

	if *help {
		printUsage()
		return
	}

	if *dir == "" {
		log.Fatal("You must specify a directory with -dir")
	}

	mux := http.NewServeMux()

	log.Printf("Listening on %s", *port)
	err := http.ListenAndServe(*port, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Printf(`$ ./hot-coffee --help
Coffee Shop Management System

Usage:
  hot-coffee [--port <N>] [--dir <S>] 
  hot-coffee --help

Options:
  --help       Show this screen.
  --port N     Port number.
  --dir S      Path to the data directory.
`)
}
