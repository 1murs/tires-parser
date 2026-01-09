package models

type Category struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type TireData struct {
	Name     string
	Quantity int
	Year     int
	Country  string
	Price    float64
}
