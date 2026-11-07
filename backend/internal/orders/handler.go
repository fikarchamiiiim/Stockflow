package orders

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"stockflow/backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/orders", h.create)
	r.GET("/orders/:id", h.get)
	r.POST("/orders/:id/cancel", h.cancel)
}

func (h *Handler) create(c *gin.Context) {
	var in CreateRequest
	if err := httpx.BindJSON(c, &in); err != nil {
		httpx.Fail(c, err)
		return
	}

	res, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	httpx.WriteJSON(c, http.StatusCreated, res)
}

func (h *Handler) get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	res, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	httpx.WriteJSON(c, http.StatusOK, res)
}

func (h *Handler) cancel(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	res, err := h.svc.Cancel(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	httpx.WriteJSON(c, http.StatusOK, res)
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		httpx.Fail(c, orderNotFound())
		return 0, false
	}
	return uint(id), true
}
