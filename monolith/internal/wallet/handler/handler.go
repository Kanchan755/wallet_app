package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kanchan755/wallet_app/monolith/internal/wallet/service"
)

type WalletHandler struct {
	svc service.WalletService
}

func NewWalletHandler(svc service.WalletService) WalletHandler {
	return &WalletHandler{
		svc: svc,
	}
}

func (h *WalletHandler) GetWalletByUserID(c *gin.Context) {
	// user_id from jwt context
	userID, _ := c.Get("user_id")

	wallet, err := h.svc.GetWalletByUserID(c, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    wallet,
	})
}
