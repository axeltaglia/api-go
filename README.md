# API Server with Product Management

This project is a simple API server built in Go for managing products. It provides endpoints for creating, updating, and retrieving products from a database.

## Features

- **Create Product**: Allows users to add new products to the database.
- **Update Product**: Enables users to update existing product information.
- **Get Product**: Retrieves product details by ID.
- **Get Products**: Retrieves a list of all products in the database.

## Setup

### Prerequisites

- Go installed on your machine.
- PostgreSQL database.

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/axeltaglia/api-go.git

2. Navigate to the project directory:
   
   ```bash
   cd product-api

3. Set up the PostgreSQL database and update the connection details in storage.go.
4. Build the project:

   ```bash
   go build
   
5. Run the server:
   ```bash
   ./product-api


### Usage

Once the server is running, you can interact with the API using HTTP requests. Here are some sample requests:

- Create product
```bash
POST /createProduct
Content-Type: application/json

{
  "name": "Product Name",
  "code": "ABC123"
}
```

- Update product
```bash
PUT /updateProduct/{id}
Content-Type: application/json

{
  "id": 1,
  "name": "Updated Product Name",
  "code": "XYZ456"
}
```


- Get product
```bash
GET /getProduct/{id}
```

- Get products
```bash
GET /getProducts
```

