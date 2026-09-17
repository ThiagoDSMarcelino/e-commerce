package main

type Products struct {
	Users []Product `json:"products"`
}

type Product struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}
