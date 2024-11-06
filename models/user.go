package models

type User struct {
	ID           int    `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	PasswordHash string `json:"-" db:"password_hash"`
	Role         string `json:"-" db:"role"`
}

func NewUser() *User {
	return &User{}
}
