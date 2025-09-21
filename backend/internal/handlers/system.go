package handlers

import (
	"net/http"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/internal/services"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetBootstrapStatus(c *gin.Context) {
	required, err := h.authService.IsAdminSetupRequired()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to determine setup status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"admin_setup_required": required})
}

func (h *Handler) SetupAdmin(c *gin.Context) {
	var req models.SetupAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if err := h.authService.CompleteAdminSetup(req.Token, req.Email, req.Password); err != nil {
		switch err {
		case services.ErrInvalidSetupToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid setup token"})
			return
		case services.ErrAdminAlreadyInitialized:
			c.JSON(http.StatusConflict, gin.H{"error": "Admin account already initialized"})
			return
		case services.ErrWeakPassword:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password does not meet requirements"})
			return
		case services.ErrAccountNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Admin account not found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete admin setup"})
			return
		}
	}

	h.response.Success(c, gin.H{"message": "Admin account initialized"})
}
