package tenant

import (
	commonRouter "common/middleware/routers"
	"manager-api/pkg/internal/tenantctx"
	tenantService "service/tenant"
	tenantDTO "service/tenant/dto"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TenantHandler struct {
	*commonRouter.BaseHandler
	service *tenantService.TenantService
}

func NewTenantHandler() *TenantHandler {
	service := tenantService.NewTenantService()
	_ = service.EnsureTable()
	return &TenantHandler{BaseHandler: &commonRouter.BaseHandler{}, service: service}
}

func (h *TenantHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/tenants", h.listTenants)
	engine.GET("/tenants/:id", h.getTenant)
	engine.POST("/tenants", h.createTenant)
	engine.PUT("/tenants/:id", h.updateTenant)
	engine.DELETE("/tenants/:id", h.deleteTenant)
}

func (h *TenantHandler) listTenants(context *gin.Context) {
	var query tenantDTO.TenantQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if tenantID := tenantctx.CurrentTenantID(context); tenantID > 0 {
		query.TenantID = tenantID
	}
	result, err := h.service.ListTenants(query)
	commonRouter.ToJson(context, result, err)
}

func (h *TenantHandler) getTenant(context *gin.Context) {
	id, ok := parseTenantID(context)
	if !ok {
		return
	}
	result, err := h.service.GetTenantByID(id)
	writeTenantResult(context, result, err)
}

func (h *TenantHandler) createTenant(context *gin.Context) {
	if tenantctx.CurrentTenantID(context) > 0 {
		commonRouter.ToError(context, "当前用户无租户创建权限")
		return
	}
	var req tenantDTO.TenantDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.CreateTenant(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *TenantHandler) updateTenant(context *gin.Context) {
	id, ok := parseTenantID(context)
	if !ok {
		return
	}
	if tenantID := tenantctx.CurrentTenantID(context); tenantID > 0 && tenantID != uint64(id) {
		commonRouter.ToError(context, "当前用户无该租户权限")
		return
	}
	var req tenantDTO.TenantDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.UpdateTenant(id, &req)
	writeTenantResult(context, result, err)
}

func (h *TenantHandler) deleteTenant(context *gin.Context) {
	if tenantctx.CurrentTenantID(context) > 0 {
		commonRouter.ToError(context, "当前用户无租户删除权限")
		return
	}
	id, ok := parseTenantID(context)
	if !ok {
		return
	}
	err := h.service.DeleteTenant(id)
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "tenant not found")
		return
	}
	commonRouter.ToJson(context, gin.H{"deleted": true}, err)
}

func writeTenantResult(context *gin.Context, result interface{}, err error) {
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "tenant not found")
		return
	}
	commonRouter.ToJson(context, result, err)
}

func parseTenantID(context *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil || id == 0 {
		commonRouter.ToError(context, "id必须是正整数")
		return 0, false
	}
	return uint(id), true
}
