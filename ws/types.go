package ws

import (
	"time"
)

type Message struct {
	ID        *string    `json:"id"`
	From      *string    `json:"from"`
	To        *string    `json:"to"`
	Message   *string    `json:"message"`
	CreatedAt *time.Time `json:"createdAt"`
	ReadAt    *time.Time `json:"readAt"`
}

type MessageStatus struct {
	Status string  `json:"status"`
	ID     *string `json:"id"`
}
