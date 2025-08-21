package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Delete URLs
// @Description Delete multiple URLs for a specific user
// @Tags urls
// @Accept json
// @Produce plain
// @Param Authorization header string true "Bearer JWT token"
// @Param request body []string true "Array of URLs to delete"
// @Success 202 {string} string "Accepted"
// @Failure 400 {string} string "Error reading body!/Error unmarshalling body!/Empty or malformed body sent!"
// @Router /api/urls [delete]
func (h *Handler) DeleteData(c *gin.Context) {
	var id string

	ctx := c.Request.Context()
	jwt := h.client.JWT
	if jwt == "" {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}	

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Error reading body!")
		return
	}

	err = json.Unmarshal(body, &id)
	if err != nil {
		c.String(http.StatusBadRequest, "Error unmarshalling body!")
		return
	}

	if len(id) == 0 {
		c.String(http.StatusBadRequest, "Empty or malformed body sent!")
		return
	}

	err = h.service.DeleteData(ctx, jwt, id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error deleting data: %v", err)
		return
	}

	c.Status(http.StatusAccepted)
}
