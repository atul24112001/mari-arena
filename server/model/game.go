package model

type GameType struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Entry     int    `json:"entry"`
	Winner    int    `json:"winner"`
	Currency  string `json:"currency"`
	MaxPlayer int    `json:"maxPlayer"`
}
