package main

import "errors"

func validateCreateOrder(req CreateOrderRequest) error {
	if req.Item == "" {
		return errors.New("item is required")
	}
	if req.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	if req.Price <= 0 {
		return errors.New("price must be greater than zero")
	}
	return nil
}
