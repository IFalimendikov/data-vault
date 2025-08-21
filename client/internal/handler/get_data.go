package handler

import (
	"data-vault/client/internal/models"
	"net/http"

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
func (h *Handler) GetData(c *gin.Context) {
	var res []models.Data

	ctx := c.Request.Context()
	jwt := h.client.JWT
	if jwt == "" {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	data, err := h.service.GetData(ctx, jwt)
	if err != nil {
		c.String(http.StatusBadRequest, "Error finding data!")
		return
	}

	if len(res) == 0 {
		c.String(http.StatusNoContent, "No data found!")
		return
	}

	for _, d := range data {
		res = append(res, models.Data{
			ID:         d.ID,
			Data:       d.Data,
			UploadedAt: d.UploadedAt,
			Status:     d.Status,
		})
	}

	c.JSON(http.StatusOK, res)
}
