package models

// User represents a user with login credentials
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Data represents a data entry in the vault
type Data struct {
	ID         string `json:"id"`
	User       string `json:"user"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	Data       []byte `json:"data"`
	UploadedAt string `json:"uploaded_at"`
}
