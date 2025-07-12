package main

import (
	"flag"
	"fmt"
	"hot-coffee/internal/handler"
	"hot-coffee/internal/repository"
	"hot-coffee/internal/service"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "4000", "HTTP network address")
	dir := flag.String("dir", "data", "Path to the directory")
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

	// Data Access Layer
	// orderRepo := repository.NewJSONOrderRepo(*dir)
	// menuRepo := repository.NewJSONMenuRepo(*dir)
	invRepo := repository.NewJSONInventoryRepo(*dir)

	// // Service layer
	// orderSvc := service.NewOrderService(orderRepo)
	// menuSvc := service.NewMenuService(menuRepo)
	invSvc := service.NewInventoryService(invRepo)

	// // Handler layer
	// orderHandler := handler.NewOrderHandler(orderSvc)
	// menuHandler := handler.NewMenuHandler(menuSvc)
	invHandler := handler.NewInventoryHandler(invSvc)

	// mux.HandleFunc("/orders", orderHandler.Orders)     // GET/POST /orders
	// mux.HandleFunc("/orders/", orderHandler.OrderByID) // GET/PUT/DELETE /orders/{id}

	// mux.HandleFunc("/menu_items", menuHandler.MenuItems)
	// mux.HandleFunc("/menu_items/", menuHandler.MenuItemByID)

	mux.HandleFunc("/inventory", invHandler.Inventory)
	// mux.HandleFunc("/inventory/", invHandler.InventoryByID)

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
