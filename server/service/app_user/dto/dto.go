package dto

import (
	baseDTO "common/base/dto"
	"time"
)

type AppUserDTO struct {
	baseDTO.BaseDTO
	Name            string     `json:"name"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	Department      string     `json:"department"`
	Password        string     `json:"password"`
	OriginPassword  string     `json:"originPassword"`
	Status          string     `json:"status"`
	LastLoginTime   *time.Time `json:"lastLoginTime"`
	SecretKey       string     `json:"secretKey"`
	Remark          string     `json:"remark"`
	BanCount        uint32     `json:"banCount"`
	OpenUID         string     `json:"openUid"`
	UnionID         string     `json:"unionId"`
	WxSessionKey    string     `json:"wxSessionKey"`
	WxNickname      string     `json:"wxNickname"`
	WxAvatar        string     `json:"wxAvatar"`
	WxGender        int        `json:"wxGender"`
	WxCountry       string     `json:"wxCountry"`
	WxProvince      string     `json:"wxProvince"`
	WxCity          string     `json:"wxCity"`
	WxLanguage      string     `json:"wxLanguage"`
	WxLastLoginTime *time.Time `json:"wxLastLoginTime"`
	StationID       uint64     `json:"stationId"`
	StationName     string     `json:"stationName"`
}

type CreateAppUserDTO struct {
	Name            string     `json:"name"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	Department      string     `json:"department"`
	Password        string     `json:"password"`
	OriginPassword  string     `json:"originPassword"`
	Status          string     `json:"status"`
	LastLoginTime   *time.Time `json:"lastLoginTime"`
	SecretKey       string     `json:"secretKey"`
	Remark          string     `json:"remark"`
	BanCount        uint32     `json:"banCount"`
	OpenUID         string     `json:"openUid"`
	UnionID         string     `json:"unionId"`
	WxSessionKey    string     `json:"wxSessionKey"`
	WxNickname      string     `json:"wxNickname"`
	WxAvatar        string     `json:"wxAvatar"`
	WxGender        int        `json:"wxGender"`
	WxCountry       string     `json:"wxCountry"`
	WxProvince      string     `json:"wxProvince"`
	WxCity          string     `json:"wxCity"`
	WxLanguage      string     `json:"wxLanguage"`
	WxLastLoginTime *time.Time `json:"wxLastLoginTime"`
	StationID       uint64     `json:"stationId"`
}

type RegisterAppUserDTO struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateAppUserDTO struct {
	Name            *string    `json:"name,omitempty"`
	Username        *string    `json:"username,omitempty"`
	Email           *string    `json:"email,omitempty"`
	Phone           *string    `json:"phone,omitempty"`
	Department      *string    `json:"department,omitempty"`
	Password        *string    `json:"password,omitempty"`
	OriginPassword  *string    `json:"originPassword,omitempty"`
	Status          *string    `json:"status,omitempty"`
	LastLoginTime   *time.Time `json:"lastLoginTime,omitempty"`
	SecretKey       *string    `json:"secretKey,omitempty"`
	Remark          *string    `json:"remark,omitempty"`
	BanCount        *uint32    `json:"banCount,omitempty"`
	OpenUID         *string    `json:"openUid,omitempty"`
	UnionID         *string    `json:"unionId,omitempty"`
	WxSessionKey    *string    `json:"wxSessionKey,omitempty"`
	WxNickname      *string    `json:"wxNickname,omitempty"`
	WxAvatar        *string    `json:"wxAvatar,omitempty"`
	WxGender        *int       `json:"wxGender,omitempty"`
	WxCountry       *string    `json:"wxCountry,omitempty"`
	WxProvince      *string    `json:"wxProvince,omitempty"`
	WxCity          *string    `json:"wxCity,omitempty"`
	WxLanguage      *string    `json:"wxLanguage,omitempty"`
	WxLastLoginTime *time.Time `json:"wxLastLoginTime,omitempty"`
	StationID       *uint64    `json:"stationId,omitempty"`
}

type AppUserQueryDTO struct {
	Page       int    `form:"page"`
	PageIndex  int    `form:"pageIndex"`
	PageSize   int    `form:"pageSize"`
	Search     string `form:"search"`
	Name       string `form:"name"`
	Username   string `form:"username"`
	Email      string `form:"email"`
	Phone      string `form:"phone"`
	Department string `form:"department"`
	Status     string `form:"status"`
	SecretKey  string `form:"secretKey"`
	OpenUID    string `form:"openUid"`
	UnionID    string `form:"unionId"`
}

type AppUserStatsDTO struct {
	VisibleUsers     int `json:"visibleUsers"`
	RecentLoginUsers int `json:"recentLoginUsers"`
	ActiveUsers      int `json:"activeUsers"`
}

type CurrentAppUserProfileDTO struct {
	Id                int        `json:"id"`
	Name              string     `json:"name"`
	Username          string     `json:"username"`
	Email             string     `json:"email"`
	Phone             string     `json:"phone"`
	Department        string     `json:"department"`
	Remark            string     `json:"remark"`
	LastLoginTime     *time.Time `json:"lastLoginTime"`
	OpenUID           string     `json:"openUid"`
	UnionID           string     `json:"unionId"`
	WxNickname        string     `json:"wxNickname"`
	WxAvatar          string     `json:"wxAvatar"`
	WxLastLoginTime   *time.Time `json:"wxLastLoginTime"`
	HasOriginPassword bool       `json:"hasOriginPassword"`
	StationID         uint64     `json:"stationId"`
	StationName       string     `json:"stationName"`
}

type WechatLoginDTO struct {
	Code      string            `json:"code"`
	RawData   string            `json:"rawData"`
	Signature string            `json:"signature"`
	UserInfo  WechatUserInfoDTO `json:"userInfo"`
}

type WechatUserInfoDTO struct {
	NickName  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
	Gender    int    `json:"gender"`
	Country   string `json:"country"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Language  string `json:"language"`
}

type UpdateCurrentAppUserProfileDTO struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Department string `json:"department"`
	Remark     string `json:"remark"`
	WxNickname string `json:"wxNickname"`
	WxAvatar   string `json:"wxAvatar"`
}

type WechatPhoneDTO struct {
	Code string `json:"code"`
}

type WechatPhoneResponseDTO struct {
	Phone string `json:"phone"`
}

type ChangeCurrentAppUserPasswordDTO struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type AppUserLoginRecordDTO struct {
	baseDTO.BaseDTO
	IP        string `json:"ip"`
	AppUserID uint64 `json:"appUserId"`
}

type CreateAppUserLoginRecordDTO struct {
	IP        string `json:"ip"`
	AppUserID uint64 `json:"appUserId"`
}

type UpdateAppUserLoginRecordDTO struct {
	IP        *string `json:"ip,omitempty"`
	AppUserID *uint64 `json:"appUserId,omitempty"`
}

type AppUserLoginRecordQueryDTO struct {
	Page      int    `form:"page"`
	PageIndex int    `form:"pageIndex"`
	PageSize  int    `form:"pageSize"`
	AppUserID uint64 `form:"appUserId"`
	IP        string `form:"ip"`
}
