package utils

type User struct {
	ID       int
	Username string
	Password string // Sensitive, should not expose
	Email    string
}

// UserDTO is a subset of User for API responses
type UserDTO struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ToDTO converts User to UserDTO
func (u *User) ToDTO() UserDTO {
	return UserDTO{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
	}
}
