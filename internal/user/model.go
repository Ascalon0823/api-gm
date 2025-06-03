package user

import "time"

type User struct {
	ID          uint      `json:"id" bson:"_id,omitempty" gorm:"primaryKey"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email" gorm:"uniqueIndex"`
	Password    string    `json:"password"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	ResetToken  string    `json:"reset_token,omitempty" gorm:"default:null"`
	ResetExpiry time.Time `json:"reset_expiry,omitempty" gorm:"default:null"`
}
