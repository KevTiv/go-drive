package models

type User struct {
	ID        string  `json:"id" gorm:"default:uuid_generate_v4()"`
	FirstName string  `json:"first_name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Country   string  `json:"country" gorm:"default:'The netherlands'"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
}
