package domain

type UserRole int

const (
	UserRoleClient UserRole = iota
	UserRoleAdmin
)

type User struct {
	ID       string   `json:"id" bson:"_id"`
	Name     string   `json:"name" bson:"name"`
	Role     UserRole `json:"role" bson:"role"`
	Email    string   `json:"email" bson:"email"`
	Password string   `json:"password" bson:"password"`
	Verified bool     `json:"verified" bson:"verified"`
}
