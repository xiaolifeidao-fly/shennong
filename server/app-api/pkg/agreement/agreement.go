package agreement

import (
	webAuth "app-api/auth"
	"app-api/pkg/internal/wechat"
	commonRouter "common/middleware/routers"
	"strings"

	agreementService "service/agreement"

	"github.com/gin-gonic/gin"
)

type AgreementHandler struct {
	*commonRouter.BaseHandler
	agreementService *agreementService.AgreementService
}

func NewAgreementHandler() *AgreementHandler {
	service := agreementService.NewAgreementService()
	_ = service.EnsureTable()
	return &AgreementHandler{
		BaseHandler:      &commonRouter.BaseHandler{},
		agreementService: service,
	}
}

func (h *AgreementHandler) RegisterHandler(engine *gin.RouterGroup) {
	webAuth.PublicGET(engine, "/agreement", h.getAgreement)
	webAuth.PublicPOST(engine, "/agreement/status", h.agreementStatus)
	webAuth.PublicPOST(engine, "/agreement/agree", h.agreeAgreement)
}

type codeRequest struct {
	Code string `json:"code"`
}

type agreementResponse struct {
	Version   string                      `json:"version"`
	Documents []agreementService.Document `json:"documents"`
}

type agreementStatusResponse struct {
	Agreed  bool   `json:"agreed"`
	Version string `json:"version"`
}

// getAgreement 返回协议正文与当前版本，供小程序展示。无需登录。
func (h *AgreementHandler) getAgreement(context *gin.Context) {
	docs, err := h.agreementService.Documents()
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	commonRouter.ToJson(context, &agreementResponse{
		Version:   agreementService.Version,
		Documents: docs,
	}, nil)
}

// agreementStatus 根据 wx.login 的 code 换取 openid，返回该微信用户是否已同意当前版本协议。
func (h *AgreementHandler) agreementStatus(context *gin.Context) {
	var req codeRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	session, err := wechat.Code2Session(strings.TrimSpace(req.Code))
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	agreed, err := h.agreementService.IsAgreed(session.OpenID)
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	commonRouter.ToJson(context, &agreementStatusResponse{
		Agreed:  agreed,
		Version: agreementService.Version,
	}, nil)
}

// agreeAgreement 记录该微信用户（openid）已同意当前版本协议。幂等。
func (h *AgreementHandler) agreeAgreement(context *gin.Context) {
	var req codeRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	session, err := wechat.Code2Session(strings.TrimSpace(req.Code))
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	if err := h.agreementService.RecordConsent(session.OpenID, session.UnionID, context.ClientIP()); err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	commonRouter.ToJson(context, &agreementStatusResponse{
		Agreed:  true,
		Version: agreementService.Version,
	}, nil)
}
