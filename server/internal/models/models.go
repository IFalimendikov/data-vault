package models

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Data struct {
	ID         string `json:"id"`
	User       string `json:"user"`
	Status     string `json:"status"`
	Data       string `json:"data"`
	UploadedAt string `json:"uploaded_at"`
}
