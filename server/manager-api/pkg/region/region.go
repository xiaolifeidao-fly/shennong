package region

import (
	commonRouter "common/middleware/routers"
	regionService "service/region"

	"github.com/gin-gonic/gin"
)

type RegionHandler struct {
	*commonRouter.BaseHandler
	service *regionService.RegionService
}

func NewRegionHandler() *RegionHandler {
	service := regionService.NewRegionService()
	_ = service.EnsureTable()
	return &RegionHandler{BaseHandler: &commonRouter.BaseHandler{}, service: service}
}

func (h *RegionHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/regions/tree", h.listTree)
}

func (h *RegionHandler) listTree(context *gin.Context) {
	result, err := h.service.ListTree()
	commonRouter.ToJson(context, result, err)
}
