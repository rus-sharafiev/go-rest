package chat

import (
	"time"
)

type Chat struct {
	ID        *int       `json:"id"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type Message struct {
	ID        *string    `json:"id"`
	ChatID    *int       `json:"chatId"`
	From      *int       `json:"from"`
	To        *int       `json:"to"`
	Message   *string    `json:"message"`
	CreatedAt *time.Time `json:"createdAt"`
	ReadAt    *time.Time `json:"readAt"`
}

type MessageStatus struct {
	Status string     `json:"status"`
	ID     *string    `json:"id"`
	Time   *time.Time `json:"time"`
}
