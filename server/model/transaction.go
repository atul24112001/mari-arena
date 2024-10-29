package model

type Transaction struct {
	Id        string `json:"id"`
	Amount    int    `json:"amount"`
	Signature string `json:"signature"`
	UserId    string `json:"userId"`
}
