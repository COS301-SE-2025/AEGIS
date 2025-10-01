// handlers/encryption.go
package handlers

import (
	"context"
	"net/http"
	"time"

	"aegis-api/pkg/encryption"

	"github.com/gin-gonic/gin"
)

type EncryptionHandler struct {
	Service encryption.Service
}

type EncryptRequest struct {
	Plaintext string `json:"plaintext" binding:"required"`
}

type EncryptResponse struct {
	Ciphertext string `json:"ciphertext"`
	Version    int    `json:"version"`
}

type DecryptRequest struct {
	Ciphertext string `json:"ciphertext" binding:"required"`
}

type DecryptResponse struct {
	Plaintext string `json:"plaintext"`
}

type BatchEncryptRequest struct {
	Items []EncryptRequest `json:"items" binding:"required,min=1"`
}

type BatchEncryptResponse struct {
	Items []EncryptResponse `json:"items"`
}

type BatchDecryptRequest struct {
	Items []DecryptRequest `json:"items" binding:"required,min=1"`
}

type BatchDecryptResponse struct {
	Items []DecryptResponse `json:"items"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *EncryptionHandler) Encrypt(c *gin.Context) {
	var req EncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	encrypted, err := h.Service.Encrypt(ctx, []byte(req.Plaintext))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, EncryptResponse{
		Ciphertext: encrypted.CipherText,
		Version:    encrypted.KeyVersion,
	})
}

func (h *EncryptionHandler) Decrypt(c *gin.Context) {
	var req DecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	plaintext, err := h.Service.Decrypt(ctx, req.Ciphertext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, DecryptResponse{
		Plaintext: string(plaintext),
	})
}

func (h *EncryptionHandler) BatchEncrypt(c *gin.Context) {
	var req BatchEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	var responses []EncryptResponse
	for _, item := range req.Items {
		encrypted, err := h.Service.Encrypt(ctx, []byte(item.Plaintext))
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		responses = append(responses, EncryptResponse{
			Ciphertext: encrypted.CipherText,
			Version:    encrypted.KeyVersion,
		})
	}

	c.JSON(http.StatusOK, BatchEncryptResponse{Items: responses})
}

func (h *EncryptionHandler) BatchDecrypt(c *gin.Context) {
	var req BatchDecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	var responses []DecryptResponse
	for _, item := range req.Items {
		plaintext, err := h.Service.Decrypt(ctx, item.Ciphertext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		responses = append(responses, DecryptResponse{
			Plaintext: string(plaintext),
		})
	}

	c.JSON(http.StatusOK, BatchDecryptResponse{Items: responses})
}
