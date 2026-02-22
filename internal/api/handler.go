package api


import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/willie0x14/ethereum-block-scanner/internal/service"
)


type Handler struct {
	svc *service.ListenerService
}

func NewHandler(svc *service.ListenerService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) Router() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	v1 := api.Group("/v1")
	{
		v1.GET("/health", h.handleHealth)
		v1.GET("/status", h.handleStatus)
		v1.GET("/events", h.handleEvents)
	}

	// v2 := r.Group("/api/v2")
	// {
	// 	v2.GET("/status", h.handleStatusV2)
	// }

	return r
}


// GET /health
func (h *Handler) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// GET /status
func (h *Handler) handleStatus(c *gin.Context) {
	status := h.svc.GetStatus(c.Request.Context())
	c.JSON(http.StatusOK, status)
}

// GET /events?limit=20
func (h *Handler) handleEvents(c *gin.Context) {
	limit := 20

	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	events := h.svc.ListRecentEvents(c.Request.Context(), limit)
	c.JSON(http.StatusOK, events)
}
