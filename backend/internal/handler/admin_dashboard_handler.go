package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/zuiyinggg/image2api/backend/internal/repo"
	"github.com/zuiyinggg/image2api/backend/pkg/errcode"
	"github.com/zuiyinggg/image2api/backend/pkg/response"
)

type AdminDashboardHandler struct {
	repo *repo.DashboardRepo
}

func NewAdminDashboardHandler(r *repo.DashboardRepo) *AdminDashboardHandler {
	return &AdminDashboardHandler{repo: r}
}

func (h *AdminDashboardHandler) Overview(c *gin.Context) {
	resp, err := h.repo.Overview(c.Request.Context())
	if err != nil {
		response.Fail(c, errcode.DBError.Wrap(err))
		return
	}
	response.OK(c, resp)
}
