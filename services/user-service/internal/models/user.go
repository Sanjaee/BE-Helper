package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the user model in the database
type User struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string     `json:"email" gorm:"uniqueIndex;not null;size:255" validate:"required,email"`
	Phone         *string    `json:"phone" gorm:"uniqueIndex;size:20"` // Optional phone number
	PasswordHash  string     `json:"-" gorm:"not null"` // Hidden from JSON
	FullName      string     `json:"full_name" gorm:"not null;size:255" validate:"required,min=2,max=255"`
	UserType      string     `json:"user_type" gorm:"not null;size:20;default:'CLIENT'" validate:"required,oneof=CLIENT SERVICE_PROVIDER ADMIN"`
	ProfilePhoto  *string    `json:"profile_photo" gorm:"type:text"` // Profile photo URL
	DateOfBirth   *time.Time `json:"date_of_birth" gorm:"type:date"`
	Gender        *string    `json:"gender" gorm:"size:10" validate:"omitempty,oneof=MALE FEMALE"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	IsVerified    bool       `json:"is_verified" gorm:"default:false"` // Email verified for all, KYC verified for providers
	OTPCode       *string    `json:"-" gorm:"size:6"` // Hidden from JSON
	OTPExpiresAt  *time.Time `json:"-" gorm:"type:timestamp"` // Hidden from JSON
	LastLogin     *time.Time `json:"last_login" gorm:"type:timestamp"`
	LoginAttempts int        `json:"-" gorm:"default:0"` // Hidden from JSON
	LockedUntil   *time.Time `json:"-" gorm:"type:timestamp"` // Hidden from JSON
	GoogleID      *string    `json:"-" gorm:"uniqueIndex;size:255"` // Hidden from JSON - Google OAuth ID
	LoginType     string     `json:"login_type" gorm:"not null;size:20;default:'CREDENTIAL'" validate:"required,oneof=CREDENTIAL GOOGLE"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// UserRegisterRequest represents the request payload for user registration
type UserRegisterRequest struct {
	FullName  string  `json:"full_name" validate:"required,min=2,max=255"`
	Email     string  `json:"email" validate:"required,email"`
	Phone     *string `json:"phone" validate:"omitempty,min=10,max=20"`
	Password  string  `json:"password" validate:"required,min=6"`
	UserType  string  `json:"user_type" validate:"required,oneof=CLIENT SERVICE_PROVIDER ADMIN"`
	Gender    *string `json:"gender" validate:"omitempty,oneof=MALE FEMALE"`
	DateOfBirth *time.Time `json:"date_of_birth" validate:"omitempty"`
}

// UserLoginRequest represents the request payload for user login
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// OTPVerifyRequest represents the request payload for OTP verification
type OTPVerifyRequest struct {
	Email   string `json:"email" validate:"required,email"`
	OTPCode string `json:"otp_code" validate:"required,len=6"`
}

// ResetPasswordRequest represents the request payload for password reset
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// VerifyResetPasswordRequest represents the request payload for reset password verification
type VerifyResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTPCode     string `json:"otp_code" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// UserResponse represents the response payload for user data
type UserResponse struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Phone        *string    `json:"phone"`
	FullName     string     `json:"full_name"`
	UserType     string     `json:"user_type"`
	ProfilePhoto *string    `json:"profile_photo"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	Gender       *string    `json:"gender"`
	IsActive     bool       `json:"is_active"`
	IsVerified   bool       `json:"is_verified"`
	LastLogin    *time.Time `json:"last_login"`
	LoginType    string     `json:"login_type"`
	CreatedAt    time.Time  `json:"created_at"`
}

// AuthResponse represents the response payload for authentication
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

// BeforeCreate hook to set UUID if not provided
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		Phone:        u.Phone,
		FullName:     u.FullName,
		UserType:     u.UserType,
		ProfilePhoto: u.ProfilePhoto,
		DateOfBirth:  u.DateOfBirth,
		Gender:       u.Gender,
		IsActive:     u.IsActive,
		IsVerified:   u.IsVerified,
		LastLogin:    u.LastLogin,
		LoginType:    u.LoginType,
		CreatedAt:    u.CreatedAt,
	}
}
