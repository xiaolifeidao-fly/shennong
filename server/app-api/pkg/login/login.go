package login

import (
	webAuth "app-api/auth"
	commonRouter "common/middleware/routers"
	"common/middleware/vipper"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	appUserService "service/app_user"
	appUserDTO "service/app_user/dto"
	authService "service/auth"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type WechatLoginResponse struct {
	Token string                               `json:"token"`
	User  *appUserDTO.CurrentAppUserProfileDTO `json:"user"`
}

type AuthStateResponse struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username"`
	DisplayName   string `json:"displayName"`
}

type LoginHandler struct {
	*commonRouter.BaseHandler
	authService *authService.AuthService
}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{
		BaseHandler: &commonRouter.BaseHandler{},
		authService: authService.NewAuthService(),
	}
}

func (h *LoginHandler) RegisterHandler(engine *gin.RouterGroup) {
	webAuth.PublicPOST(engine, "/login", h.login)
	webAuth.PublicPOST(engine, "/wechat-login", h.wechatLogin)
	webAuth.PublicPOST(engine, "/register", h.register)
	engine.GET("/auth-state", h.authState)
	engine.POST("/logout", h.logout)
}

func (h *LoginHandler) login(context *gin.Context) {
	var req LoginRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}

	maxLoginErrorNum := vipper.GetInt64("user.max.login.error.num")
	if maxLoginErrorNum <= 0 {
		maxLoginErrorNum = 20
	}

	token, _, err := h.authService.Login(req.Username, req.Password, context.ClientIP(), maxLoginErrorNum)
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	commonRouter.ToJson(context, &LoginResponse{Token: token}, nil)
}

func (h *LoginHandler) wechatLogin(context *gin.Context) {
	var req appUserDTO.WechatLoginDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}

	session, err := code2Session(strings.TrimSpace(req.Code))
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}

	service := appUserService.NewAppUserService()
	user, err := service.UpsertWechatUser(session.OpenID, session.UnionID, session.SessionKey, req.UserInfo)
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}

	token, _, err := h.authService.LoginAppUser(user, context.ClientIP())
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}

	profile, err := service.GetCurrentUserProfile(uint(user.Id))
	commonRouter.ToJson(context, &WechatLoginResponse{Token: token, User: profile}, err)
}

func (h *LoginHandler) logout(context *gin.Context) {
	token := webAuth.ExtractToken(context)
	if value, ok := context.Get(webAuth.ContextTokenKey); ok {
		if contextToken, typeOK := value.(string); typeOK && contextToken != "" {
			token = contextToken
		}
	}
	if err := h.authService.Logout(token); err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	commonRouter.ToJson(context, gin.H{"loggedOut": true}, nil)
}

func (h *LoginHandler) authState(context *gin.Context) {
	value, ok := context.Get(webAuth.ContextUserKey)
	if !ok {
		commonRouter.ToError(context, authService.ErrNotLogin.Error())
		return
	}

	user, typeOK := value.(*authService.LoginUser)
	if !typeOK || user == nil {
		commonRouter.ToError(context, authService.ErrNotLogin.Error())
		return
	}

	commonRouter.ToJson(context, &AuthStateResponse{
		Authenticated: true,
		Username:      user.Username,
		DisplayName:   user.Name,
	}, nil)
}

func (h *LoginHandler) register(context *gin.Context) {
	var req RegisterRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}

	service := appUserService.NewAppUserService()
	result, err := service.RegisterUser(&appUserDTO.RegisterAppUserDTO{
		Name:     req.Name,
		Username: req.Username,
		Password: req.Password,
	})
	commonRouter.ToJson(context, result, err)
}

type wechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func code2Session(code string) (*wechatSessionResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("微信登录 code 不能为空")
	}
	appID := strings.TrimSpace(vipper.GetString("wechat.appid"))
	secret := strings.TrimSpace(vipper.GetString("wechat.secret"))
	if appID == "" || secret == "" {
		return nil, fmt.Errorf("未配置 wechat.appid 或 wechat.secret")
	}

	values := url.Values{}
	values.Set("appid", appID)
	values.Set("secret", secret)
	values.Set("js_code", code)
	values.Set("grant_type", "authorization_code")

	client := &http.Client{Timeout: 8 * time.Second}
	response, err := client.Get("https://api.weixin.qq.com/sns/jscode2session?" + values.Encode())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var session wechatSessionResponse
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		return nil, err
	}
	if session.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败: %s", session.ErrMsg)
	}
	if strings.TrimSpace(session.OpenID) == "" {
		return nil, fmt.Errorf("微信未返回 openid")
	}
	return &session, nil
}
