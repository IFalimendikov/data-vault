package handler

import (
	"encoding/json"
	"io"
	"net/http"
	
	"github.com/gin-gonic/gin"
)

// @Summary Create shortened URL
// @Description Creates a shortened version of a provided URL
// @Tags urls
// @Accept plain
// @Produce plain
// @Security Bearer
// @Param Authorization header string true "Bearer JWT token"
// @Param url body string true "Original URL to shorten"
// @Success 201 {string} string "Shortened URL"
// @Failure 400 {string} string "Can't read body!/Empty body!/Malformed URI!/Couldn't encode URL!"
// @Failure 409 {string} string "URL already exists"
// @Router /api/url [post]
func (h *Handler) PostData(c *gin.Context) {
	var data string

	ctx := c.Request.Context()
	jwt := h.client.JWT
	if jwt == "" {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Cant read body!")
		return
	}

	if len(body) == 0 {
		c.String(http.StatusBadRequest, "Empty body!")
		return
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		c.String(http.StatusBadRequest, "Malformed URI!")
		return
	}

	err = h.service.PostData(ctx, jwt, data)
	if err != nil {
		c.String(http.StatusConflict, "URL already exists")
		return
	}

	c.String(http.StatusCreated, "ok")
}
