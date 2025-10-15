package handlers

import (
	"log"
	"net/http"
	"time"

	"user-service/internal/events"
	"user-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	db              *gorm.DB
	passwordService *models.PasswordService
	otpService      *models.OTPService
	JWTService      *JWTService
	validator       *validator.Validate
	eventService    *events.EventService
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *gorm.DB) *UserHandler {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found in user handlers package, using system env")
	}

	// Initialize event service
	eventService, err := events.NewEventService()
	if err != nil {
		log.Printf("⚠️ Failed to initialize event service: %v", err)
		// Continue without event service for now
	}

	return &UserHandler{
		db:              db,
		passwordService: models.NewPasswordService(),
		otpService:      models.NewOTPService(),
		JWTService:      NewJWTService(),
		validator:       validator.New(),
		eventService:    eventService,
	}
}

// Register handles user registration
func (uh *UserHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := uh.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Check if phone is provided and already exists
	if req.Phone != nil && *req.Phone != "" {
		var existingPhoneUser models.User
		if err := uh.db.Where("phone = ?", *req.Phone).First(&existingPhoneUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this phone number already exists"})
			return
		}
	}

	// Hash password
	hashedPassword, err := uh.passwordService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Generate OTP
	otp, err := uh.otpService.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	// Create user
	user := models.User{
		FullName:     req.FullName,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
		UserType:     req.UserType,
		Gender:       req.Gender,
		DateOfBirth:  req.DateOfBirth,
		OTPCode:      &otp,
		LoginType:    "CREDENTIAL",
		IsVerified:   false,
		IsActive:     true,
	}

	// Save user to database
	if err := uh.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Publish user registered event to message broker
	if uh.eventService != nil {
		if err := uh.eventService.PublishUserRegistered(user.ID.String(), user.FullName, user.Email, otp); err != nil {
			log.Printf("⚠️ Failed to publish user registered event: %v", err)
			// Don't fail the registration if event publishing fails
		} else {
			log.Printf("✅ User registered event published for: %s", user.Email)
		}
	} else {
		log.Printf("⚠️ Event service not available, skipping event publishing")
	}

	// Return success response (OTP will be sent via email through message broker)
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. Please check your email for verification code.",
		"user":    user.ToResponse(),
	})
}

// Login handles user login
func (uh *UserHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := uh.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not found",
				"message": "Email tidak terdaftar. Silakan periksa kembali email Anda atau daftar akun baru.",
				"code":    "USER_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user type is credential (not Google OAuth user)
	if user.LoginType != "CREDENTIAL" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Account type mismatch",
			"message": "Akun ini dibuat dengan Google. Silakan gunakan tombol 'Masuk dengan Google' untuk login.",
			"code":    "ACCOUNT_TYPE_MISMATCH",
		})
		return
	}

	// Check if user email is verified
	if !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Email not verified",
			"message": "Email Anda belum terverifikasi. Silakan cek email Anda dan klik link verifikasi atau masukkan kode OTP yang telah dikirim.",
			"code":    "EMAIL_NOT_VERIFIED",
			"user_id": user.ID,
			"email":   user.Email,
		})
		return
	}

	// Verify password
	if err := uh.passwordService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid password",
			"message": "Password yang Anda masukkan salah. Silakan coba lagi.",
			"code":    "INVALID_PASSWORD",
		})
		return
	}

	// Generate tokens
	authResponse, err := uh.JWTService.GenerateTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// VerifyOTP handles OTP verification
func (uh *UserHandler) VerifyOTP(c *gin.Context) {
	var req models.OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate OTP format
	if !uh.otpService.ValidateOTP(req.OTPCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP format"})
		return
	}

	// Find user by email
	var user models.User
	if err := uh.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user is already verified
	// If user is verified but has OTP, it might be for password reset
	if user.IsVerified {
		// Check if this is a password reset OTP by checking if user has OTP but is verified
		if user.OTPCode != nil {
			// This is likely a password reset OTP, redirect to password reset flow
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "This OTP is for password reset. Please use the password reset flow.",
				"code":  "OTP_FOR_PASSWORD_RESET",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already verified"})
		return
	}

	// Verify OTP
	if user.OTPCode == nil || *user.OTPCode != req.OTPCode {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	// Update user as verified and clear OTP
	user.IsVerified = true
	user.OTPCode = nil
	user.UpdatedAt = time.Now()

	if err := uh.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
		return
	}

	// Generate tokens after successful verification
	authResponse, err := uh.JWTService.GenerateTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Publish user login event after successful verification
	if uh.eventService != nil {
		if err := uh.eventService.PublishUserLogin(user.ID.String(), user.FullName, user.Email); err != nil {
			log.Printf("⚠️ Failed to publish user login event: %v", err)
		}
	}

	c.JSON(http.StatusOK, authResponse)
}

// VerifyOTPResetPassword handles OTP verification for password reset
func (uh *UserHandler) VerifyOTPResetPassword(c *gin.Context) {
	var req models.OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate OTP format
	if !uh.otpService.ValidateOTP(req.OTPCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP format"})
		return
	}

	// Find user by email
	var user models.User
	if err := uh.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user is verified (must be verified to reset password)
	if !user.IsVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User account is not verified"})
		return
	}

	// Verify OTP
	if user.OTPCode == nil || *user.OTPCode != req.OTPCode {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	// For password reset, we don't clear OTP yet - it will be cleared when password is changed
	// Just return success to allow user to proceed to change password
	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully. You can now change your password.",
		"user":    user.ToResponse(),
	})
}

// CheckUserStatus checks user status and OTP type
func (uh *UserHandler) CheckUserStatus(c *gin.Context) {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := uh.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"status":    "not_found",
				"needs_otp": false,
				"otp_type":  nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Determine status based on user verification and OTP
	status := "verified"
	needsOtp := false
	var otpType *string

	if !user.IsVerified {
		status = "unverified"
		needsOtp = true
		otpType = stringPtr("registration")
	} else if user.OTPCode != nil {
		// User is verified but has OTP (likely for password reset)
		status = "verified_with_otp"
		needsOtp = true
		otpType = stringPtr("password_reset")
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"needs_otp": needsOtp,
		"otp_type":  otpType,
		"user":      user.ToResponse(),
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// ResendOTP handles OTP resending
func (uh *UserHandler) ResendOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" validate:"required,email"`
		Type  string `json:"type"` // "registration" or "password_reset"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := uh.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Generate new OTP
	otp, err := uh.otpService.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	// Update user with new OTP
	user.OTPCode = &otp
	user.UpdatedAt = time.Now()

	if err := uh.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update OTP"})
		return
	}

	// Determine which event to publish based on type
	if req.Type == "password_reset" {
		// For password reset, user must be verified
		if !user.IsVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Account not verified. Please verify your email first."})
			return
		}

		// Publish password reset event
		if uh.eventService != nil {
			if err := uh.eventService.PublishPasswordReset(user.ID.String(), user.FullName, user.Email, otp); err != nil {
				log.Printf("⚠️ Failed to publish password reset event: %v", err)
				// Don't fail the resend if event publishing fails
			} else {
				log.Printf("✅ Password reset OTP resend event published for: %s", user.Email)
			}
		}
	} else if req.Type == "registration" {
		// For registration, user must not be verified
		if user.IsVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User is already verified"})
			return
		}

		// Publish user registered event again to resend OTP
		if uh.eventService != nil {
			if err := uh.eventService.PublishUserRegistered(user.ID.String(), user.FullName, user.Email, otp); err != nil {
				log.Printf("⚠️ Failed to publish resend OTP event: %v", err)
				// Don't fail the resend if event publishing fails
			} else {
				log.Printf("✅ Resend OTP event published for: %s", user.Email)
			}
		}
	} else {
		// Default to registration flow for backward compatibility
		if user.IsVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User is already verified"})
			return
		}

		// Publish user registered event again to resend OTP
		if uh.eventService != nil {
			if err := uh.eventService.PublishUserRegistered(user.ID.String(), user.FullName, user.Email, otp); err != nil {
				log.Printf("⚠️ Failed to publish resend OTP event: %v", err)
				// Don't fail the resend if event publishing fails
			} else {
				log.Printf("✅ Resend OTP event published for: %s", user.Email)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully. Please check your email.",
	})
}

// GetProfile handles getting user profile
func (uh *UserHandler) GetProfile(c *gin.Context) {
	userID, _, _, _, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := uh.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
}

// GetUserByID handles getting user by ID (for other services)
func (uh *UserHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")

	// Parse UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID format",
		})
		return
	}

	var user models.User
	if err := uh.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Return user data in the format expected by payment service
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":        user.ID.String(),
			"full_name": user.FullName,
			"email":     user.Email,
		},
	})
}

// UpdateProfile handles updating user profile
func (uh *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _, _, _, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		FullName    string     `json:"full_name" validate:"omitempty,min=2,max=255"`
		Phone       *string    `json:"phone" validate:"omitempty,min=10,max=20"`
		Gender      *string    `json:"gender" validate:"omitempty,oneof=MALE FEMALE"`
		DateOfBirth *time.Time `json:"date_of_birth" validate:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := uh.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if phone is already taken by another user
	if req.Phone != nil && *req.Phone != "" && (user.Phone == nil || *req.Phone != *user.Phone) {
		var existingUser models.User
		if err := uh.db.Where("phone = ? AND id != ?", *req.Phone, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Phone number already taken"})
			return
		}
		user.Phone = req.Phone
	}

	// Update other fields
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Gender != nil {
		user.Gender = req.Gender
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = req.DateOfBirth
	}

	user.UpdatedAt = time.Now()

	if err := uh.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user.ToResponse(),
	})
}

// RefreshToken handles token refresh
func (uh *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate refresh token
	claims, err := uh.JWTService.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Find user
	var user models.User
	if err := uh.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Generate new tokens
	authResponse, err := uh.JWTService.GenerateTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// RequestResetPassword handles password reset request
func (uh *UserHandler) RequestResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := uh.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return error for non-existent users to prevent unnecessary processing
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Email tidak terdaftar dalam sistem",
				"message": "Email yang Anda masukkan tidak terdaftar. Silakan periksa kembali atau daftar akun baru.",
				"code":    "EMAIL_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user is verified
	if !user.IsVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account not verified. Please verify your email first."})
		return
	}

	// Generate OTP for password reset
	otp, err := uh.otpService.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset code"})
		return
	}

	// Update user with reset OTP
	user.OTPCode = &otp
	user.UpdatedAt = time.Now()

	if err := uh.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset code"})
		return
	}

	// Publish password reset event to message broker
	if uh.eventService != nil {
		if err := uh.eventService.PublishPasswordReset(user.ID.String(), user.FullName, user.Email, otp); err != nil {
			log.Printf("⚠️ Failed to publish password reset event: %v", err)
			// Don't fail the request if event publishing fails
		} else {
			log.Printf("✅ Password reset event published for: %s", user.Email)
		}
	} else {
		log.Printf("⚠️ Event service not available, skipping password reset event publishing")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a reset code has been sent.",
	})
}

// VerifyResetPassword handles password reset verification
func (uh *UserHandler) VerifyResetPassword(c *gin.Context) {
	var req models.VerifyResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate OTP format
	if !uh.otpService.ValidateOTP(req.OTPCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset code format"})
		return
	}

	// Find user by email
	var user models.User
	if err := uh.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user is verified
	if !user.IsVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account not verified. Please verify your email first."})
		return
	}

	// Verify OTP
	if user.OTPCode == nil || *user.OTPCode != req.OTPCode {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset code"})
		return
	}

	// Hash new password
	hashedPassword, err := uh.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process new password"})
		return
	}

	// Update user password and clear OTP
	user.PasswordHash = hashedPassword
	user.OTPCode = nil
	user.UpdatedAt = time.Now()

	if err := uh.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Generate new tokens after successful password reset
	authResponse, err := uh.JWTService.GenerateTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Password reset successfully",
		"user":          user.ToResponse(),
		"access_token":  authResponse.AccessToken,
		"refresh_token": authResponse.RefreshToken,
		"expires_in":    authResponse.ExpiresIn,
	})
}

// GoogleOAuth handles Google OAuth user creation/update
func (uh *UserHandler) GoogleOAuth(c *gin.Context) {
	var req struct {
		Email        string `json:"email" validate:"required,email"`
		FullName     string `json:"full_name" validate:"required,min=2,max=255"`
		ProfilePhoto string `json:"profile_photo"`
		GoogleID     string `json:"google_id" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := uh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists by email
	var user models.User
	err := uh.db.Where("email = ?", req.Email).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			FullName:     req.FullName,
			Email:        req.Email,
			ProfilePhoto: &req.ProfilePhoto,
			GoogleID:     &req.GoogleID,
			LoginType:    "GOOGLE",
			UserType:     "CLIENT", // Default to CLIENT for Google OAuth users
			IsVerified:   true,     // Google users are automatically verified
			IsActive:     true,
		}

		if err := uh.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		// Check if existing user is credential type
		if user.LoginType == "CREDENTIAL" {
			c.JSON(http.StatusConflict, gin.H{"error": "This email is already registered with credentials. Please use email/password login instead."})
			return
		}

		// Update existing Google user with new info
		user.ProfilePhoto = &req.ProfilePhoto
		user.GoogleID = &req.GoogleID
		user.IsVerified = true // Ensure Google users are verified
		user.UpdatedAt = time.Now()

		if err := uh.db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	}

	// Generate tokens
	authResponse, err := uh.JWTService.GenerateTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}
