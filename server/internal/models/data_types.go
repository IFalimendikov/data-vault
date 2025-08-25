package models

// Data type constants for the vault
const (
	DataTypeText     = "text"
	DataTypePassword = "password"
	DataTypeBinary   = "binary"
	DataTypeCard     = "card"
)

// DataType represents the type of data stored in the vault
type DataType string

// LoginPasswordData represents login/password pair data
type LoginPasswordData struct {
	Website  string `json:"website"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Notes    string `json:"notes"`
}

// BankCardData represents banking card data
type BankCardData struct {
	Bank     string `json:"bank"`
	Number   string `json:"number"`
	Holder   string `json:"holder"`
	CVV      string `json:"cvv"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
	Notes    string `json:"notes"`
}

// TextData represents arbitrary text data
type TextData struct {
	Content string `json:"content"`
	Notes   string `json:"notes"`
}

// BinaryData represents arbitrary binary data
type BinaryData struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	Notes    string `json:"notes"`
}
