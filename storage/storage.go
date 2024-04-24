package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"math/rand/v2"
	"time"
)

type Product struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewProduct(name, code string) *Product {
	return &Product{
		Id:        rand.Int64(),
		Name:      name,
		Code:      code,
		CreatedAt: time.Now().UTC(),
	}
}

type Storage interface {
	CreateProduct(*Product) (*Product, error)
	GetProducts() ([]*Product, error)
	GetProductById(int64) (*Product, error)
}

type PgStorage struct {
	db *sql.DB
}

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

// docker run --name apigo -e POSTGRES_USER=apigo -e POSTGRES_PASSWORD=apigo -e POSTGRES_DB=apigo -p 5439:5432 -d postgres
func (o *PgStorage) Init() error {
	_, err := o.db.Exec(`
		create table if not exists product
		(
			id 		  serial primary key,
			name      varchar(50),
			code      varchar(50),
			createdAt timestamp
		)
    `)

	return err
}

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

func (o *PgStorage) GetProducts() ([]*Product, error) {
	rows, err := o.db.Query("select * from product")
	if err != nil {
		return nil, err
	}

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

func (o *PgStorage) GetProductById(id int64) (*Product, error) {
	rows, err := o.db.Query("select * from product where id=$1", id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("product with ID %d not found", id)
	}

	product := new(Product)

	if err = rows.Scan(&product.Id, &product.Name, &product.Code, &product.CreatedAt); err != nil {
		return nil, err
	}

	return product, nil
}
