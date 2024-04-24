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

type Server struct {
	listenAddr string
	db         storage.Storage
	serverMux  *http.ServeMux
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func NewApiServer(listenAddr string, storage storage.Storage) *Server {
	serverMux := http.NewServeMux()
	return &Server{
		listenAddr: listenAddr,
		serverMux:  serverMux,
		db:         storage,
	}
}

func (o *Server) HandleEndpoints() {
	o.serverMux.HandleFunc("/getProducts", interceptError(interceptLogger(o.getProducts)))
	o.serverMux.HandleFunc("/getProduct/{id}", interceptError(interceptLogger(o.getProduct)))
	o.serverMux.HandleFunc("/createProduct", interceptError(interceptLogger(o.createProduct)))
}

func (o *Server) Run() error {
	if err := http.ListenAndServe(":8080", o.serverMux); err != nil {
		slog.Error(err.Error())
	}

	return nil
}

func interceptLogger(f apiFunc) apiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		slog.Info("intercept Logger")
		return f(w, r)
	}
}

type WebError struct {
	Error string
}

func interceptError(f apiFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("interceptError")
		if err := f(w, r); err != nil {
			printStackTrace(err)
			//slog.Error(err.Error(), "data", err)
			if err := writeJSON(w, http.StatusBadRequest, WebError{Error: err.Error()}); err != nil {
				slog.Error("couldn't write")
				return
			}
		}
		slog.Info("interceptError after")
	}
}

func printStackTrace(err error) {
	stackTrace := string(debug.Stack())
	fmt.Printf("%v\n%s\n", err.Error(), stackTrace)
}

type getProductResponse struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
}

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

type CreateProductRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type CreateProductResponse struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
}

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

type GetProductsResponse struct {
	Products []*storage.Product `json:"products"`
}

func (o *Server) getProducts(w http.ResponseWriter, _ *http.Request) error {
	products, err := o.db.GetProducts()
	if err != nil {
		return err
	}

	getProductsResponse := new(GetProductsResponse)
	getProductsResponse.Products = products

	return writeJSON(w, http.StatusOK, getProductsResponse)
}

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

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
