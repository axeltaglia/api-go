package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log/slog"
	"math/rand/v2"
	"time"
)

// Product represents a product entity.
type Product struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
}

// NewProduct creates a new Product instance with the provided name and code.
func NewProduct(name, code string) *Product {
	return &Product{
		Id:        rand.Int64(),
		Name:      name,
		Code:      code,
		CreatedAt: time.Now().UTC(),
	}
}

// Storage is an interface for interacting with product data.
type Storage interface {
	CreateProduct(*Product) (*Product, error)
	GetProducts() ([]*Product, error)
	GetProductById(int64) (*Product, error)
	UpdateProduct(*Product) (*Product, error)
}

// PgStorage represents PostgreSQL storage implementation.
type PgStorage struct {
	db *sql.DB
}

// NewPgStorage creates a new instance of PgStorage.
func NewPgStorage() (*PgStorage, error) {
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5439, "apigo", "apigo", "apigo")

	db, err := sql.Open("postgres", postgresqlDbInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PgStorage{db: db}, nil
}

// Init initializes the database schema.
func (o *PgStorage) Init() error {
	_, err := o.db.Exec(`
		create table if not exists product
		(
			id        serial primary key,
			name      varchar(50),
			code      varchar(50),
			createdAt timestamp
		)
    `)

	return err
}

// CreateProduct inserts a new product into the database.
func (o *PgStorage) CreateProduct(p *Product) (*Product, error) {
	_, err := o.db.Exec("insert into product (name, code, createdAt) values($1, $2, $3)", p.Name, p.Code, p.CreatedAt)
	if err != nil {
		return nil, err
	}

	var lastInsertId int64
	err = o.db.QueryRow("SELECT lastval()").Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}

	p.Id = lastInsertId

	return p, nil
}

// GetProducts retrieves all products from the database.
func (o *PgStorage) GetProducts() ([]*Product, error) {
	rows, err := o.db.Query("select * from product")
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(rows)

	products := make([]*Product, 0)

	for rows.Next() {
		product := &Product{}
		if err := rows.Scan(&product.Id, &product.Name, &product.Code, &product.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

// GetProductById retrieves a product from the database by its ID.
func (o *PgStorage) GetProductById(id int64) (*Product, error) {
	rows, err := o.db.Query("select * from p where id=$1", id)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(rows)

	if !rows.Next() {
		return nil, fmt.Errorf("p with ID %d not found", id)
	}

	p := new(Product)

	if err = rows.Scan(&p.Id, &p.Name, &p.Code, &p.CreatedAt); err != nil {
		return nil, err
	}

	return p, nil
}

// UpdateProduct updates an existing product in the database.
func (o *PgStorage) UpdateProduct(p *Product) (*Product, error) {
	_, err := o.db.Exec("update product set name=$1, code=$2, createdAt=$3 where id=$4", p.Name, p.Code, p.CreatedAt, p.Id)
	if err != nil {
		return nil, err
	}

	return p, nil
}
