package dto

import (
	baseDTO "common/base/dto"
	"time"
)

type GrainPurchaseEntryDTO struct {
	baseDTO.BaseDTO
	StationID        uint64     `json:"stationId"`
	StationName      string     `json:"stationName"`
	AppUserID        uint64     `json:"appUserId"`
	FarmerID         uint64     `json:"farmerId"`
	FarmerName       string     `json:"farmerName"`
	FarmerIDNumber   string     `json:"farmerIdNumber"`
	FarmerPhone      string     `json:"farmerPhone"`
	FarmerAddress    string     `json:"farmerAddress"`
	FarmerBankNumber string     `json:"farmerBankNumber"`
	FarmerBankName   string     `json:"farmerBankName"`
	PurchaseTypeID   uint64     `json:"purchaseTypeId"`
	Crop             string     `json:"crop"`
	Quantity         float64    `json:"quantity"`
	Unit             string     `json:"unit"`
	DisplayUnit      string     `json:"displayUnit"`
	Amount           float64    `json:"amount"`
	UnitPrice        float64    `json:"unitPrice"`
	BuyTime          *time.Time `json:"buyTime"`
	PlaceID          uint64     `json:"placeId"`
	Place            string     `json:"place"`
	LocationName     string     `json:"locationName"`
	LocationAddress  string     `json:"locationAddress"`
	Longitude        string     `json:"longitude"`
	Latitude         string     `json:"latitude"`
	Province         string     `json:"province"`
	City             string     `json:"city"`
	District         string     `json:"district"`
	PaymentMethodID  uint64     `json:"paymentMethodId"`
	PayType          string     `json:"payType"`
	Status           string     `json:"status"`
	Version          int        `json:"version"`
	Remark           string     `json:"remark"`
}

type GrainPurchaseEntrySnapshotDTO struct {
	baseDTO.BaseDTO
	EntryID           uint64     `json:"entryId"`
	SnapshotVersion   int        `json:"snapshotVersion"`
	SnapshotAction    string     `json:"snapshotAction"`
	SnapshotTime      *time.Time `json:"snapshotTime"`
	OperatorAppUserID uint64     `json:"operatorAppUserId"`
	OperatorName      string     `json:"operatorName"`
	StationID         uint64     `json:"stationId"`
	AppUserID         uint64     `json:"appUserId"`
	FarmerID          uint64     `json:"farmerId"`
	FarmerName        string     `json:"farmerName"`
	FarmerIDNumber    string     `json:"farmerIdNumber"`
	FarmerPhone       string     `json:"farmerPhone"`
	FarmerAddress     string     `json:"farmerAddress"`
	FarmerBankNumber  string     `json:"farmerBankNumber"`
	FarmerBankName    string     `json:"farmerBankName"`
	PurchaseTypeID    uint64     `json:"purchaseTypeId"`
	Crop              string     `json:"crop"`
	Quantity          float64    `json:"quantity"`
	Unit              string     `json:"unit"`
	Amount            float64    `json:"amount"`
	UnitPrice         float64    `json:"unitPrice"`
	BuyTime           *time.Time `json:"buyTime"`
	PlaceID           uint64     `json:"placeId"`
	Place             string     `json:"place"`
	LocationName      string     `json:"locationName"`
	LocationAddress   string     `json:"locationAddress"`
	Longitude         string     `json:"longitude"`
	Latitude          string     `json:"latitude"`
	Province          string     `json:"province"`
	City              string     `json:"city"`
	District          string     `json:"district"`
	PaymentMethodID   uint64     `json:"paymentMethodId"`
	PayType           string     `json:"payType"`
	EntryStatus       string     `json:"entryStatus"`
	EntryRemark       string     `json:"entryRemark"`
}

type GrainFarmerPurchaseSummaryDTO struct {
	baseDTO.BaseDTO
	StationID       uint64     `json:"stationId"`
	AppUserID       uint64     `json:"appUserId"`
	PurchaseTypeID  uint64     `json:"purchaseTypeId"`
	Crop            string     `json:"crop"`
	SummaryDate     *time.Time `json:"summaryDate"`
	FarmerID        uint64     `json:"farmerId"`
	PaymentMethodID uint64     `json:"paymentMethodId"`
	PayType         string     `json:"payType"`
	EntryCount      int        `json:"entryCount"`
	TotalAmount     float64    `json:"totalAmount"`
	TotalQuantity   float64    `json:"totalQuantity"`
}

type GrainFarmerDailySummaryDTO struct {
	baseDTO.BaseDTO
	StationID      uint64     `json:"stationId"`
	AppUserID      uint64     `json:"appUserId"`
	FarmerID       uint64     `json:"farmerId"`
	FarmerName     string     `json:"farmerName"`
	FarmerIDNumber string     `json:"farmerIdNumber"`
	FarmerPhone    string     `json:"farmerPhone"`
	FarmerAddress  string     `json:"farmerAddress"`
	BankNumber     string     `json:"bankNumber"`
	BankName       string     `json:"bankName"`
	Status         string     `json:"status"`
	StatusText     string     `json:"statusText"`
	SummaryDate    *time.Time `json:"summaryDate"`
	MainCrop       string     `json:"mainCrop"`
	EntryCount     int        `json:"entryCount"`
	TotalAmount    float64    `json:"totalAmount"`
	TotalQuantity  float64    `json:"totalQuantity"`
	LatestTime     *time.Time `json:"latestTime"`
}

type GrainPurchaseDashboardMetricDTO struct {
	NewFarmerCount    int     `json:"newFarmerCount"`
	ActiveUserCount   int     `json:"activeUserCount"`
	TotalQuantity     float64 `json:"totalQuantity"`
	TotalAmount       float64 `json:"totalAmount"`
	EntryCount        int     `json:"entryCount"`
	AverageUnitPrice  float64 `json:"averageUnitPrice"`
	AverageFarmerDeal float64 `json:"averageFarmerDeal"`
}

type GrainPurchaseDashboardDimensionDTO struct {
	Key              string  `json:"key"`
	Name             string  `json:"name"`
	StationID        uint64  `json:"stationId"`
	PurchaseTypeID   uint64  `json:"purchaseTypeId"`
	EntryCount       int     `json:"entryCount"`
	FarmerCount      int     `json:"farmerCount"`
	ActiveUserCount  int     `json:"activeUserCount"`
	TotalQuantity    float64 `json:"totalQuantity"`
	TotalAmount      float64 `json:"totalAmount"`
	AverageUnitPrice float64 `json:"averageUnitPrice"`
	AmountShare      float64 `json:"amountShare"`
	QuantityShare    float64 `json:"quantityShare"`
}

type GrainPurchaseDashboardDTO struct {
	StartDate string                                `json:"startDate"`
	EndDate   string                                `json:"endDate"`
	StationID uint64                                `json:"stationId"`
	Overview  GrainPurchaseDashboardMetricDTO       `json:"overview"`
	ByStation []*GrainPurchaseDashboardDimensionDTO `json:"byStation"`
	ByCrop    []*GrainPurchaseDashboardDimensionDTO `json:"byCrop"`
	Generated *time.Time                            `json:"generated"`
}

type GrainEntryMaterialDTO struct {
	baseDTO.BaseDTO
	StationID       uint64 `json:"stationId"`
	EntryID         uint64 `json:"entryId"`
	FarmerID        uint64 `json:"farmerId"`
	AppUserID       uint64 `json:"appUserId"`
	MaterialBizType string `json:"materialBizType"`
	MaterialType    string `json:"materialType"`
	OssBucket       string `json:"ossBucket"`
	OssObjectKey    string `json:"ossObjectKey"`
	OssURL          string `json:"ossUrl"`
	ImageURL        string `json:"imageUrl"`
	FileName        string `json:"fileName"`
	ImageHash       string `json:"imageHash"`
	FileSize        int64  `json:"fileSize"`
	MimeType        string `json:"mimeType"`
	ETag            string `json:"etag"`
	SortOrder       int    `json:"sortOrder"`
}

type GrainPurchaseEntryQueryDTO struct {
	Page            int        `form:"page"`
	PageIndex       int        `form:"pageIndex"`
	PageSize        int        `form:"pageSize"`
	StationID       uint64     `form:"stationId"`
	StationIDs      []uint64   `form:"-"`
	AppUserID       uint64     `form:"appUserId"`
	AppUserIDs      []uint64   `form:"-"`
	FarmerID        uint64     `form:"farmerId"`
	FarmerIDs       []uint64   `form:"-"`
	PurchaseTypeID  uint64     `form:"purchaseTypeId"`
	PurchaseTypeIDs []uint64   `form:"-"`
	Search          string     `form:"search"`
	Status          string     `form:"status"`
	StartDate       string     `form:"startDate"`
	EndDate         string     `form:"endDate"`
	StartTime       *time.Time `form:"startTime" time_format:"2006-01-02 15:04:05"`
	EndTime         *time.Time `form:"endTime" time_format:"2006-01-02 15:04:05"`
	MinAmount       *float64   `form:"minAmount"`
	MaxAmount       *float64   `form:"maxAmount"`
}

type GrainEntrySnapshotQueryDTO struct {
	Page       int      `form:"page"`
	PageIndex  int      `form:"pageIndex"`
	PageSize   int      `form:"pageSize"`
	EntryID    uint64   `form:"entryId"`
	StationID  uint64   `form:"stationId"`
	StationIDs []uint64 `form:"-"`
	AppUserID  uint64   `form:"appUserId"`
}

type GrainFarmerPurchaseSummaryQueryDTO struct {
	Page       int        `form:"page"`
	PageIndex  int        `form:"pageIndex"`
	PageSize   int        `form:"pageSize"`
	StationID  uint64     `form:"stationId"`
	StationIDs []uint64   `form:"-"`
	AppUserID  uint64     `form:"appUserId"`
	FarmerID   uint64     `form:"farmerId"`
	StartDate  *time.Time `form:"startDate" time_format:"2006-01-02"`
	EndDate    *time.Time `form:"endDate" time_format:"2006-01-02"`
	Search     string     `form:"search"`
}

type GrainFarmerDailySummaryQueryDTO struct {
	Page       int        `form:"page"`
	PageIndex  int        `form:"pageIndex"`
	PageSize   int        `form:"pageSize"`
	StationID  uint64     `form:"stationId"`
	StationIDs []uint64   `form:"-"`
	AppUserID  uint64     `form:"appUserId"`
	FarmerID   uint64     `form:"farmerId"`
	StartDate  *time.Time `form:"startDate" time_format:"2006-01-02"`
	EndDate    *time.Time `form:"endDate" time_format:"2006-01-02"`
	Search     string     `form:"search"`
}

type GrainPurchaseDashboardQueryDTO struct {
	StationID  uint64     `form:"stationId"`
	StationIDs []uint64   `form:"-"`
	AppUserID  uint64     `form:"appUserId"`
	StartDate  *time.Time `form:"startDate" time_format:"2006-01-02"`
	EndDate    *time.Time `form:"endDate" time_format:"2006-01-02"`
}

type GrainEntryMaterialQueryDTO struct {
	Page            int      `form:"page"`
	PageIndex       int      `form:"pageIndex"`
	PageSize        int      `form:"pageSize"`
	StationID       uint64   `form:"stationId"`
	StationIDs      []uint64 `form:"-"`
	EntryID         uint64   `form:"entryId"`
	FarmerID        uint64   `form:"farmerId"`
	AppUserID       uint64   `form:"appUserId"`
	MaterialBizType string   `form:"materialBizType"`
	MaterialType    string   `form:"materialType"`
}
