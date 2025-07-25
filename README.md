# Hot-Coffee

A simple coffee shop management system written in Go, using JSON files for data persistence. It provides RESTful HTTP APIs to manage orders, menu items, and inventory.

## Table of Contents

* [Features](#features)
* [Prerequisites](#prerequisites)
* [Installation](#installation)
* [Project Structure](#project-structure)
* [Configuration](#configuration)
* [Data Files](#data-files)
* [Usage](#usage)

    * [Running the Server](#running-the-server)
    * [API Endpoints](#api-endpoints)
* [Logging](#logging)
* [Testing](#testing)

## Features

* **Orders**: Create, retrieve, update, delete, and close orders.
* **Menu Items**: Manage menu items and their ingredients.
* **Inventory**: Track ingredient stock levels and enforce business rules when placing orders.

## Project Structure

```
hot-coffee/
├── cmd/                   # Main application entrypoint
│   └── main.go
├── internal/
│   ├── handler/           # HTTP handlers
│   ├── repository/        # JSON-based repositories (DAL)
│   └── service/           # Business logic layer (services)
├── models/                # Domain models and shared types
├── data/                  # JSON data files: orders.json, menu_items.json, inventory.json
├── tests/                 # Unit tests for repositories and services
└── go.mod                 # Go module definition
```

## Configuration

The server accepts the following flags:

* `--port` (default `:4000`): HTTP network address to listen on.
* `--dir` (default `data`): Path to the directory containing JSON data files.

```bash
./hot-coffee --port :4000 --dir ./data
```

## Data Files

All data is persisted as JSON arrays in `data/`:

* `inventory.json`: Array of `InventoryItem` objects.
* `menu_items.json`: Array of `MenuItem` objects, each listing ingredients.
* `orders.json`: Array of `Order` objects, each listing order items.

Refer to the `models/` folder for the exact struct definitions and JSON field names.

## Usage

### Running the Server

```bash
./hot-coffee --port :4000 --dir ./data
```

### API Endpoints

#### Orders

| Method | URI                  | Description                     |
| ------ | -------------------- | ------------------------------- |
| GET    | `/orders`            | List all orders                 |
| POST   | `/orders`            | Create a new order              |
| GET    | `/orders/{id}`       | Get order by ID                 |
| PUT    | `/orders/{id}`       | Update existing order           |
| DELETE | `/orders/{id}`       | Delete an order                 |
| POST   | `/orders/{id}/close` | Close an order (mark as closed) |

#### Menu Items

| Method | URI                | Description            |
| ------ | ------------------ | ---------------------- |
| GET    | `/menu_items`      | List all menu items    |
| POST   | `/menu_items`      | Create a new menu item |
| GET    | `/menu_items/{id}` | Get menu item by ID    |
| PUT    | `/menu_items/{id}` | Update a menu item     |
| DELETE | `/menu_items/{id}` | Delete a menu item     |

#### Inventory

| Method | URI               | Description              |
| ------ | ----------------- | ------------------------ |
| GET    | `/inventory`      | List all inventory items |
| POST   | `/inventory`      | Add a new inventory item |
| GET    | `/inventory/{id}` | Get inventory item by ID |
| PUT    | `/inventory/{id}` | Update an inventory item |
| DELETE | `/inventory/{id}` | Delete an inventory item |

## Logging

Uses Go's `log/slog` package to emit structured logs at different levels:

* **Info**: High-level events (service start, entity created).
* **Debug**: Detailed diagnostic information (file I/O operations).
* **Warn**: Recoverable issues (duplicate keys, validation failures).
* **Error**: Unrecoverable errors (I/O failures, missing files).

Logs include contextual key/value pairs like `orderID`, `productID`, and file paths.

## Testing

Unit tests live in the `tests/` directory. Run:

```bash
go test ./internal/repository ./internal/service -v
```

