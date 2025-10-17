# Order Management System — Go + gRPC

A simple demonstration of a **Go microservice architecture** using **gRPC** communication between three independent services:

* **Product Service** — manages product catalog
* **Payment Service** — simulates payments
* **Order Service** — orchestrates orders, validating products and charging payments

This project showcases inter-service communication, Protobuf contract sharing, and service orchestration patterns in Go.

---

## 📦 Architecture Overview

```
+----------------+        +----------------+        +----------------+
| ProductService | <----> | OrderService   | <----> | PaymentService |
+----------------+        +----------------+        +----------------+
| gRPC :50051    |        | gRPC :50055    |        | gRPC :50052    |
+----------------+        +----------------+        +----------------+
```

Each service runs independently and communicates over gRPC using a shared Protobuf definition (`proto/order.proto`).

---

## 🚀 Getting Started

---

### Run all services

Open **three terminals** and start each service:

```bash
make run-product   # runs on :50051
make run-payment   # runs on :50053
make run-order     # runs on :50055
```

You should see logs like:

```
  product service listening :50051
  payment service listening :50053
  order service listening :50055
```

---

### 3. Test using gRPCurl or Evans

#### Create an order

```bash
grprcurl -plaintext -d '{"user_id":"u1","product_ids":["p1","p2"]}' localhost:50055 orderproto.OrderService/CreateOrder
```

#### Get order details

```bash
grprcurl -plaintext -d '{"order_id":"ord_123456789"}' localhost:50055 orderproto.OrderService/GetOrder
```

---

## 🧠 Implementation Details

* Each service is an independent Go binary with its own `main.go`.
* The **OrderService** connects to ProductService and PaymentService using `grpc.Dial()`.
* Data is stored in-memory for simplicity.
* PaymentService always returns a successful mock transaction.

---

## ⚙️ Makefile Commands

```makefile
make run-product  # Run product service
make run-payment  # Run payment service
make run-order    # Run order service
```

---

## 🧩 Next Steps

You can extend this system by:

* Introducing authentication (JWT interceptor).
* Implementing retries and circuit breakers.
* Deploying all services with Docker Compose.
* Adding tracing via OpenTelemetry (otelgrpc middleware).

---

## 🧰 Tech Stack

* **Language:** Go 1.21+
* **RPC Framework:** gRPC
* **IDL:** Protocol Buffers (`.proto`)
* **Build Tool:** Makefile

---

## 📝 License

MIT License. Free for educational and commercial use.
