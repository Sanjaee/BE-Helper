package handlers

import (
	"net/http"

	"user-service/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CheckVerificationStatus checks if user email is verified
func (uh *UserHandler) CheckVerificationStatus(c *gin.Context) {
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
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "Email tidak terdaftar.",
				"code":    "USER_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user type is credential (not Google OAuth user)
	if user.LoginType != "CREDENTIAL" {
		c.JSON(http.StatusOK, gin.H{
			"message": "User is verified (Google OAuth)",
			"data": gin.H{
				"user_id":     user.ID,
				"email":       user.Email,
				"is_verified": true,
				"login_type":  user.LoginType,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification status retrieved",
		"data": gin.H{
			"user_id":         user.ID,
			"email":           user.Email,
			"is_verified":     user.IsVerified,
			"login_type":      user.LoginType,
			"has_pending_otp": user.OTPCode != nil,
		},
	})
}
