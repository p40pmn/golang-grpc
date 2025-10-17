# Run Product Service
run-product:
	cd product-service go run cmd/*.go

# Run Payment Service
run-payment:
	cd payment-service && go run cmd/*.go

# Run Order Service
run-order:
	cd order-service && go run cmd/*.go