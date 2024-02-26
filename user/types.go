package user

import "time"

type Data struct {
	ID        *int       `json:"id"`
	Email     *string    `json:"email"`
	FirstName *string    `json:"firstName"`
	LastName  *string    `json:"lastName"`
	Phone     *string    `json:"phone"`
	Avatar    *string    `json:"avatar"`
	CreatedAt *time.Time `json:"createdAt"`
	Access    *string    `json:"access"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type WithPassword struct {
	Data
	Password *string `json:"password"`
}

type CreateDto struct {
	Email        string
	FirstName    string
	PasswordHash string
}

type UpdateDto struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Phone     *string `json:"phone"`
	Avatar    *string `json:"avatar"`
}
