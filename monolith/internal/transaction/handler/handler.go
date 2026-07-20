package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kanchan755/wallet_app/monolith/internal/errors"
	"github.com/kanchan755/wallet_app/monolith/internal/transaction/model"
	"github.com/kanchan755/wallet_app/monolith/internal/transaction/service"
)

type TransactionHandler struct {
	svc service.TransactionService
}

func NewTransactionHandler(svc service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

// transfer godoc
// @Summary     Transfer money between wallets
// @Description Perform a money transfer from one user's wallet to another user's wallet.
// @Tags        Transactions
// @Accept      json
// @Produce     json
// @Param       request body model.TransferRequest true "Transfer details"
// @Security    Bearer
// @Success     200 {object} model.Transaction "Transfer successful"
// @Failure     400 {object} errors.AppError "Invalid request"
// @Failure     401 {object} errors.AppError "Unauthorized"
// @Failure     404 {object} errors.AppError "Wallet or user not found"
// @Failure     409 {object} errors.AppError "Concurrency conflict"
// @Failure     500 {object} errors.AppError "Internal server error"
// @Router      /api/v1/transfer [post]
func (h *TransactionHandler) Transfer(c *gin.Context) {
	//get senderUserID from auth middleware
	senderUserID, exist := c.Get("user_id")
	if !exist {
		c.Error(errors.NewAppError(http.StatusUnauthorized, "Unauthorized", "Unauthorized"))
		return
	}
	var req model.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	//
	tx, err := h.svc.Transfer(c.Request.Context(), senderUserID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction successful",
		"data":    tx,
	})

}

func (h *TransactionHandler) GetHistory(c *gin.Context) {
	UserID, exist := c.Get("user_id")
	if !exist {
		c.Error(errors.NewAppError(http.StatusUnauthorized, "Unauthorized", "Unauthorized"))
		return
	}
	var params model.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(errors.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}
	tx, meta, err := h.svc.GetHistory(c.Request.Context(), UserID.(string), &params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Success: true,
		Data:    tx,
		Meta:    *meta,
	})

}
