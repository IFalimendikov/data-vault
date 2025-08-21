package handler

import (
	"data-vault/client/internal/models"
	"net/http"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

// @Summary Get user's URLs
// @Description Retrieves all URLs associated with the authenticated user
// @Tags urls
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {array} models.UserURLResponse "List of user's URLs"
// @Success 204 {string} string "No URLs found!"
// @Failure 400 {string} string "Error finding URLs!"
// @Router /api/user/urls [get]
func (h *Handler) Register(c *gin.Context) {
	var user models.User

	ctx := c.Request.Context()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Cant read body!")
		return
	}

	if len(body) == 0 {
		c.String(http.StatusBadRequest, "Empty body!")
		return
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		c.String(http.StatusBadRequest, "Malformed user data!")
		return
	}

	jwt, err := h.service.Register(ctx, user)
	if err != nil {
		c.String(http.StatusConflict, "URL already exists")
		return
	}

	h.client.SetJWT(jwt)

	c.JSON(http.StatusOK, "ok")
}
