package products

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
	r.GET("/products", h.list)
	r.POST("/products", h.create)
	r.PUT("/products/:id", h.update)
	r.DELETE("/products/:id", h.delete)
	r.GET("/products/:id/movements", h.movements)
	r.POST("/products/:id/stock", h.adjustStock)
}

func (h *Handler) list(c *gin.Context) {
	var q ListQuery
	if err := httpx.BindQuery(c, &q); err != nil {
		httpx.Fail(c, err)
		return
	}

	res, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	httpx.WriteJSON(c, http.StatusOK, res)
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

func (h *Handler) update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var in UpdateRequest
	if err := httpx.BindJSON(c, &in); err != nil {
		httpx.Fail(c, err)
		return
	}

	res, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	httpx.WriteJSON(c, http.StatusOK, res)
}

func (h *Handler) delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httpx.Fail(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) movements(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	res, err := h.svc.ListMovements(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	httpx.WriteJSON(c, http.StatusOK, res)
}

func (h *Handler) adjustStock(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var in StockRequest
	if err := httpx.BindJSON(c, &in); err != nil {
		httpx.Fail(c, err)
		return
	}

	res, err := h.svc.AdjustStock(c.Request.Context(), id, in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	httpx.WriteJSON(c, http.StatusOK, res)
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		httpx.Fail(c, productNotFound())
		return 0, false
	}
	return uint(id), true
}
