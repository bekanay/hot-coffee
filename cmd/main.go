package main

import (
	"flag"
	"fmt"
	"hot-coffee/internal/handler"
	"hot-coffee/internal/repository"
	"hot-coffee/internal/service"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(jsonHandler))

	port := flag.String("port", ":4000", "HTTP network address")
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

	slog.Info("Starting Hot-Coffee", "port", *port, "dataDir", *dir)

	// Data Access Layer
	orderRepo := repository.NewJSONOrderRepo(*dir)
	menuRepo := repository.NewJSONMenuRepo(*dir)
	invRepo := repository.NewJSONInventoryRepo(*dir)

	// // Service layer
	orderSvc := service.NewOrderService(orderRepo, menuRepo)
	menuSvc := service.NewMenuService(menuRepo, invRepo)
	invSvc := service.NewInventoryService(invRepo)

	// // Handler layer
	orderHandler := handler.NewOrderHandler(orderSvc)
	menuHandler := handler.NewMenuHandler(menuSvc)
	invHandler := handler.NewInventoryHandler(invSvc)

	mux := http.NewServeMux()

	mux.HandleFunc("/orders", orderHandler.Orders)     // GET/POST /orders
	mux.HandleFunc("/orders/", orderHandler.OrderByID) // GET/PUT/DELETE /orders/{id}

	mux.HandleFunc("/menu", menuHandler.Menu)
	mux.HandleFunc("/menu/", menuHandler.MenuByID)

	mux.HandleFunc("/inventory", invHandler.Inventory)
	mux.HandleFunc("/inventory/", invHandler.InventoryByID)

	mux.HandleFunc("/reports/total-sales", orderHandler.GetTotalSales)
	mux.HandleFunc("/reports/popular-items", orderHandler.GetPopularMenuItems)

	slog.Info("Listening", "address", *port)
	err := http.ListenAndServe(*port, mux)
	if err != nil {
		slog.Error("Server failed", err)
		os.Exit(1)
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
