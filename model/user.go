package model

import "time"

type User struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Name      string    `json:"name" bson:"name" validate:"required"`
	Email     string    `json:"email" bson:"email" validate:"required,email"`
	Password  string    `json:"-" bson:"password"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
