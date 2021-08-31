package serializers

import (
	"time"

	"github.com/google/uuid"
)

type LoginSerializer struct {
	Token string             `json:"token"`
	User  ShowUserSerializer `json:"user"`
}

type ShowUserSerializer struct {
	ID        uuid.UUID   `json:"id"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Tenants   interface{} `json:"tenants"`
}
