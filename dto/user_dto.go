package dto

type UserDto struct {
	Email   string `json:"email"`
	Pasword string `json:"password"`
}

type UserEmail struct {
	Email string `json:"email"`
}
