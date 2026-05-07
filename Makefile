.PHONY: orders fulfillment

orders:
	cd services/orders && go run .

fulfillment:
	cd services/fulfillment && go run .
