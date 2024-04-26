// Package api implements the HTTP server for handling API endpoints.

package api

import (
	"apiGo/storage"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// Server represents the API server configuration.
type Server struct {
	listenAddr string          // Address the server listens on.
	db         storage.Storage // Database instance.
	serverMux  *http.ServeMux  // HTTP request multiplexer.
}

// NewApiServer creates a new instance of the API server.
func NewApiServer(listenAddr string, storage storage.Storage) *Server {
	serverMux := http.NewServeMux()
	return &Server{
		listenAddr: listenAddr,
		serverMux:  serverMux,
		db:         storage,
	}
}

// HandleEndpoints sets up the API endpoints and their corresponding handlers.
func (o *Server) HandleEndpoints() {
	o.serverMux.HandleFunc("/getProducts", interceptError(interceptLogger(o.getProducts)))
	o.serverMux.HandleFunc("/getProduct/{id}", interceptError(interceptLogger(o.getProduct)))
	o.serverMux.HandleFunc("/createProduct", interceptError(interceptLogger(o.createProduct)))
	o.serverMux.HandleFunc("/updateProduct/{id}", interceptError(interceptLogger(o.updateProduct)))
}

// Run starts the API server.
func (o *Server) Run() error {
	if err := http.ListenAndServe(":8080", o.serverMux); err != nil {
		slog.Error(err.Error())
	}

	return nil
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

// interceptLogger is a middleware that logs information about API requests.
func interceptLogger(f apiFunc) apiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		serviceName := getServiceName(r.URL.Path)
		slog.Info("service call", "serviceName", serviceName)
		return f(w, r)
	}
}

// WebError represents an error response sent to clients.
type WebError struct {
	Error string `json:"error"`
}

// interceptError is a middleware that intercepts errors and sends appropriate responses to clients.
func interceptError(f apiFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("interceptError")
		if err := f(w, r); err != nil {
			printStackTrace(err)
			if err := writeJSON(w, http.StatusBadRequest, WebError{Error: err.Error()}); err != nil {
				slog.Error("couldn't write")
				return
			}
		}
		slog.Info("interceptError after")
	}
}

// printStackTrace prints the stack trace for the given error.
func printStackTrace(err error) {
	stackTrace := string(debug.Stack())
	fmt.Printf("%v\n%s\n", err.Error(), stackTrace)
}

// getProductResponse represents the response structure for getProduct API.
type getProductResponse struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
}

// getProduct retrieves a product by its ID.
func (o *Server) getProduct(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}

	p, err := o.db.GetProductById(id)
	if err != nil {
		return err
	}

	response := getProductResponse{
		Id:        p.Id,
		Name:      p.Name,
		Code:      p.Code,
		CreatedAt: p.CreatedAt,
	}

	return writeJSON(w, http.StatusOK, response)
}

// CreateProductRequest represents the request structure for createProduct API.
type CreateProductRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// CreateProductResponse represents the response structure for createProduct API.
type CreateProductResponse struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
}

// createProduct creates a new product.
func (o *Server) createProduct(w http.ResponseWriter, r *http.Request) error {
	request := new(CreateProductRequest)
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return err
	}

	p := storage.NewProduct(request.Name, request.Code)

	product, err := o.db.CreateProduct(p)
	if err != nil {
		return err
	}

	response := CreateProductResponse{
		Id:        product.Id,
		Name:      product.Name,
		Code:      product.Code,
		CreatedAt: product.CreatedAt,
	}

	return writeJSON(w, http.StatusOK, response)
}

// UpdateProductRequest represents the request structure for updateProduct API.
type UpdateProductRequest struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// updateProduct updates an existing product.
func (o *Server) updateProduct(w http.ResponseWriter, r *http.Request) error {
	request := new(UpdateProductRequest)
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return err
	}

	p := &storage.Product{
		Id:   request.Id,
		Name: request.Name,
		Code: request.Code,
	}

	updatedProduct, err := o.db.UpdateProduct(p)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, updatedProduct)
}

// GetProductsResponse represents the response structure for getProducts API.
type GetProductsResponse struct {
	Products []*storage.Product `json:"products"`
}

// getProducts retrieves all products.
func (o *Server) getProducts(w http.ResponseWriter, _ *http.Request) error {
	products, err := o.db.GetProducts()
	if err != nil {
		return err
	}

	getProductsResponse := new(GetProductsResponse)
	getProductsResponse.Products = products

	return writeJSON(w, http.StatusOK, getProductsResponse)
}

// getId extracts the ID from the request URL.
func getId(r *http.Request) (int64, error) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return 0, errors.New("the id argument is not present")
	}
	id := parts[2]
	n, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("numeric id is expected. Given: %s", id)
	}
	return int64(n), nil
}

// getServiceName extracts the service name from the request URL.
func getServiceName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "Unknown"
	}
	return parts[1]
}

// writeJSON writes JSON response to the client.
func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
