package login

import (
	webAuth "app-api/auth"
	"bytes"
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
	webAuth.PublicPOST(engine, "/wechat-phone-login", h.wechatPhoneLogin)
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
	user, err := service.FindActiveWechatUser(session.OpenID)
	if err != nil {
		commonRouter.ToError(context, "微信账号未开通业务员账号")
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

func (h *LoginHandler) wechatPhoneLogin(context *gin.Context) {
	var req appUserDTO.WechatPhoneDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	phone, err := getWechatPhoneNumber(strings.TrimSpace(req.Code))
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}

	service := appUserService.NewAppUserService()
	user, err := service.FindActiveUserByPhone(phone)
	if err != nil {
		commonRouter.ToError(context, "手机号未开通业务员账号")
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

type wechatAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type wechatPhoneResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
	} `json:"phone_info"`
}

func getWechatPhoneNumber(code string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("手机号授权 code 不能为空")
	}
	accessToken, err := getWechatAccessToken()
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(gin.H{"code": code})
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 8 * time.Second}
	apiURL := "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=" + url.QueryEscape(accessToken)
	response, err := client.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var result wechatPhoneResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("获取微信手机号失败: %s", result.ErrMsg)
	}
	phone := strings.TrimSpace(result.PhoneInfo.PurePhoneNumber)
	if phone == "" {
		phone = strings.TrimSpace(result.PhoneInfo.PhoneNumber)
	}
	if phone == "" {
		return "", fmt.Errorf("微信未返回手机号")
	}
	return phone, nil
}

func getWechatAccessToken() (string, error) {
	appID := strings.TrimSpace(vipper.GetString("wechat.appid"))
	secret := strings.TrimSpace(vipper.GetString("wechat.secret"))
	if appID == "" || secret == "" {
		return "", fmt.Errorf("未配置 wechat.appid 或 wechat.secret")
	}

	values := url.Values{}
	values.Set("grant_type", "client_credential")
	values.Set("appid", appID)
	values.Set("secret", secret)

	client := &http.Client{Timeout: 8 * time.Second}
	response, err := client.Get("https://api.weixin.qq.com/cgi-bin/token?" + values.Encode())
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var result wechatAccessTokenResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("获取微信 access_token 失败: %s", result.ErrMsg)
	}
	if strings.TrimSpace(result.AccessToken) == "" {
		return "", fmt.Errorf("微信未返回 access_token")
	}
	return result.AccessToken, nil
}
