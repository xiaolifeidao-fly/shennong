package repository

import (
	"common/middleware/db"
	"time"
)

type GrainPurchaseEntry struct {
	db.BaseEntity
	StationID       uint64     `gorm:"column:station_id;type:bigint unsigned;index:idx_station_id" description:"粮站ID"`
	AppUserID       uint64     `gorm:"column:app_user_id;type:bigint unsigned;index:idx_app_user_id" description:"录入业务员ID"`
	FarmerID        uint64     `gorm:"column:farmer_id;type:bigint unsigned;index:idx_farmer_id" description:"农户ID"`
	PurchaseTypeID  uint64     `gorm:"column:purchase_type_id;type:bigint unsigned;index:idx_purchase_type_id" description:"收购类型ID"`
	Crop            string     `gorm:"column:crop;type:varchar(100);index:idx_crop" description:"收购类型名称冗余"`
	Quantity        float64    `gorm:"column:quantity;type:decimal(12,3);default:0" description:"重量（公斤）"`
	Unit            string     `gorm:"column:unit;type:varchar(20)" description:"单位（内部固定公斤）"`
	DisplayUnit     string     `gorm:"column:display_unit;type:varchar(20)" description:"展示单位"`
	Amount          float64    `gorm:"column:amount;type:decimal(12,2);default:0" description:"金额"`
	UnitPrice       float64    `gorm:"column:unit_price;type:decimal(12,4);default:0" description:"单价"`
	BuyTime         *time.Time `gorm:"column:buy_time;type:datetime;index:idx_buy_time" description:"收购时间"`
	PayTime         *time.Time `gorm:"column:pay_time;type:datetime;index:idx_pay_time" description:"支付时间（可选，后期补填）"`
	PlaceID         uint64     `gorm:"column:place_id;type:bigint unsigned;index:idx_place_id" description:"收购地点ID"`
	Place           string     `gorm:"column:place;type:varchar(100);index:idx_place" description:"收购地点名称冗余"`
	LocationName    string     `gorm:"column:location_name;type:varchar(100)" description:"定位名称"`
	LocationAddress string     `gorm:"column:location_address;type:varchar(255)" description:"定位地址"`
	Longitude       string     `gorm:"column:longitude;type:varchar(50)" description:"经度"`
	Latitude        string     `gorm:"column:latitude;type:varchar(50)" description:"纬度"`
	Province        string     `gorm:"column:province;type:varchar(50)" description:"省"`
	City            string     `gorm:"column:city;type:varchar(50)" description:"市"`
	District        string     `gorm:"column:district;type:varchar(50)" description:"区县"`
	PaymentMethodID uint64     `gorm:"column:payment_method_id;type:bigint unsigned;index:idx_payment_method_id" description:"付款方式ID"`
	PayType         string     `gorm:"column:pay_type;type:varchar(100)" description:"付款方式名称冗余"`
	Status          string     `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
	Version         int        `gorm:"column:version;type:int;default:1" description:"当前版本号"`
	Remark          string     `gorm:"column:remark;type:varchar(255)" description:"备注"`
}

func (g *GrainPurchaseEntry) TableName() string {
	return "grain_purchase_entry"
}

type GrainPurchaseEntrySnapshot struct {
	db.BaseEntity
	EntryID           uint64     `gorm:"column:entry_id;type:bigint unsigned;index:idx_entry_id" description:"收粮录入ID"`
	SnapshotVersion   int        `gorm:"column:snapshot_version;type:int;index:idx_snapshot_version" description:"快照版本"`
	SnapshotAction    string     `gorm:"column:snapshot_action;type:varchar(50);index:idx_snapshot_action" description:"快照动作"`
	SnapshotTime      *time.Time `gorm:"column:snapshot_time;type:datetime;index:idx_snapshot_time" description:"快照时间"`
	OperatorAppUserID uint64     `gorm:"column:operator_app_user_id;type:bigint unsigned;index:idx_operator_app_user_id" description:"操作人ID"`
	OperatorName      string     `gorm:"column:operator_name;type:varchar(100)" description:"操作人名称"`

	StationID        uint64 `gorm:"column:station_id;type:bigint unsigned;index:idx_station_id" description:"粮站ID"`
	AppUserID        uint64 `gorm:"column:app_user_id;type:bigint unsigned;index:idx_app_user_id" description:"录入业务员ID"`
	FarmerID         uint64 `gorm:"column:farmer_id;type:bigint unsigned;index:idx_farmer_id" description:"农户ID"`
	FarmerName       string `gorm:"column:farmer_name;type:varchar(255)" description:"农户姓名"`
	FarmerIDNumber   string `gorm:"column:farmer_id_number;type:varchar(128)" description:"农户身份证号"`
	FarmerPhone      string `gorm:"column:farmer_phone;type:varchar(32)" description:"农户电话"`
	FarmerAddress    string `gorm:"column:farmer_address;type:varchar(255)" description:"农户地址"`
	FarmerBankNumber string `gorm:"column:farmer_bank_number;type:varchar(160)" description:"农户银行卡号"`
	FarmerBankName   string `gorm:"column:farmer_bank_name;type:varchar(255)" description:"农户开户行"`

	PurchaseTypeID  uint64     `gorm:"column:purchase_type_id;type:bigint unsigned" description:"收购类型ID"`
	Crop            string     `gorm:"column:crop;type:varchar(100)" description:"收购类型名称"`
	Quantity        float64    `gorm:"column:quantity;type:decimal(12,3);default:0" description:"重量"`
	Unit            string     `gorm:"column:unit;type:varchar(20)" description:"单位"`
	Amount          float64    `gorm:"column:amount;type:decimal(12,2);default:0" description:"金额"`
	UnitPrice       float64    `gorm:"column:unit_price;type:decimal(12,4);default:0" description:"单价"`
	BuyTime         *time.Time `gorm:"column:buy_time;type:datetime" description:"收购时间"`
	PayTime         *time.Time `gorm:"column:pay_time;type:datetime" description:"支付时间"`
	PlaceID         uint64     `gorm:"column:place_id;type:bigint unsigned" description:"收购地点ID"`
	Place           string     `gorm:"column:place;type:varchar(100)" description:"收购地点"`
	LocationName    string     `gorm:"column:location_name;type:varchar(100)" description:"定位名称"`
	LocationAddress string     `gorm:"column:location_address;type:varchar(255)" description:"定位地址"`
	Longitude       string     `gorm:"column:longitude;type:varchar(50)" description:"经度"`
	Latitude        string     `gorm:"column:latitude;type:varchar(50)" description:"纬度"`
	Province        string     `gorm:"column:province;type:varchar(50)" description:"省"`
	City            string     `gorm:"column:city;type:varchar(50)" description:"市"`
	District        string     `gorm:"column:district;type:varchar(50)" description:"区县"`
	PaymentMethodID uint64     `gorm:"column:payment_method_id;type:bigint unsigned" description:"付款方式ID"`
	PayType         string     `gorm:"column:pay_type;type:varchar(100)" description:"付款方式"`
	EntryStatus     string     `gorm:"column:entry_status;type:varchar(50)" description:"录入状态"`
	EntryRemark     string     `gorm:"column:entry_remark;type:varchar(255)" description:"录入备注"`
}

func (g *GrainPurchaseEntrySnapshot) TableName() string {
	return "grain_purchase_entry_snapshot"
}

type GrainFarmerPurchaseSummary struct {
	db.BaseEntity
	StationID       uint64     `gorm:"column:station_id;type:bigint unsigned;uniqueIndex:idx_grain_farmer_purchase_summary_dim" description:"粮站ID"`
	AppUserID       uint64     `gorm:"column:app_user_id;type:bigint unsigned;index:idx_app_user_id" description:"录入业务员ID"`
	PurchaseTypeID  uint64     `gorm:"column:purchase_type_id;type:bigint unsigned;uniqueIndex:idx_grain_farmer_purchase_summary_dim" description:"收购类型ID"`
	Crop            string     `gorm:"column:crop;type:varchar(100);index:idx_crop" description:"收购类型名称"`
	SummaryDate     *time.Time `gorm:"column:summary_date;type:date;uniqueIndex:idx_grain_farmer_purchase_summary_dim" description:"汇总日期"`
	FarmerID        uint64     `gorm:"column:farmer_id;type:bigint unsigned;uniqueIndex:idx_grain_farmer_purchase_summary_dim" description:"农户ID"`
	PaymentMethodID uint64     `gorm:"column:payment_method_id;type:bigint unsigned;uniqueIndex:idx_grain_farmer_purchase_summary_dim" description:"付款方式ID"`
	PayType         string     `gorm:"column:pay_type;type:varchar(100)" description:"付款方式名称"`
	EntryCount      int        `gorm:"column:entry_count;type:int;default:0" description:"录入笔数"`
	TotalAmount     float64    `gorm:"column:total_amount;type:decimal(14,2);default:0" description:"总金额"`
	TotalQuantity   float64    `gorm:"column:total_quantity;type:decimal(14,3);default:0" description:"总公斤"`
}

func (g *GrainFarmerPurchaseSummary) TableName() string {
	return "grain_farmer_purchase_summary"
}

type GrainStationPurchaseSummary struct {
	db.BaseEntity
	StationID      uint64     `gorm:"column:station_id;type:bigint unsigned;uniqueIndex:idx_grain_station_purchase_summary_dim" description:"粮仓ID"`
	AppUserID      uint64     `gorm:"column:app_user_id;type:bigint unsigned;uniqueIndex:idx_grain_station_purchase_summary_dim" description:"业务员ID"`
	PurchaseTypeID uint64     `gorm:"column:purchase_type_id;type:bigint unsigned;uniqueIndex:idx_grain_station_purchase_summary_dim" description:"收粮类型ID"`
	Crop           string     `gorm:"column:crop;type:varchar(100);index:idx_crop" description:"收粮类型名称"`
	SummaryDate    *time.Time `gorm:"column:summary_date;type:date;uniqueIndex:idx_grain_station_purchase_summary_dim" description:"汇总日期"`
	EntryCount     int        `gorm:"column:entry_count;type:int;default:0" description:"录入笔数"`
	TotalAmount    float64    `gorm:"column:total_amount;type:decimal(14,2);default:0" description:"总金额"`
	TotalQuantity  float64    `gorm:"column:total_quantity;type:decimal(14,3);default:0" description:"总公斤"`
}

func (g *GrainStationPurchaseSummary) TableName() string {
	return "grain_station_purchase_summary"
}

type GrainEntryMaterial struct {
	db.BaseEntity
	StationID       uint64 `gorm:"column:station_id;type:bigint unsigned;index:idx_station_id" description:"粮站ID"`
	EntryID         uint64 `gorm:"column:entry_id;type:bigint unsigned;index:idx_entry_id;uniqueIndex:uniq_entry_image_hash" description:"收粮录入ID"`
	FarmerID        uint64 `gorm:"column:farmer_id;type:bigint unsigned;index:idx_farmer_id" description:"农户ID"`
	AppUserID       uint64 `gorm:"column:app_user_id;type:bigint unsigned;index:idx_app_user_id" description:"业务员ID"`
	MaterialBizType string `gorm:"column:material_biz_type;type:varchar(50);index:idx_material_biz_type" description:"材料业务类型"`
	MaterialType    string `gorm:"column:material_type;type:varchar(50);index:idx_material_type" description:"材料类型"`
	OssBucket       string `gorm:"column:oss_bucket;type:varchar(100)" description:"OSS Bucket"`
	OssObjectKey    string `gorm:"column:oss_object_key;type:varchar(500);index:idx_oss_object_key" description:"OSS Object Key"`
	OssURL          string `gorm:"column:oss_url;type:varchar(1000)" description:"OSS URL"`
	FileName        string `gorm:"column:file_name;type:varchar(255)" description:"文件名"`
	ImageHash       string `gorm:"column:image_hash;type:char(64);not null;uniqueIndex:uniq_entry_image_hash" description:"文件名SHA-256，与EntryID组成唯一键"`
	FileSize        int64  `gorm:"column:file_size;type:bigint;default:0" description:"文件大小"`
	MimeType        string `gorm:"column:mime_type;type:varchar(100)" description:"文件类型"`
	ETag            string `gorm:"column:etag;type:varchar(100)" description:"ETag"`
	SortOrder       int    `gorm:"column:sort_order;type:int;default:0" description:"排序"`
}

func (g *GrainEntryMaterial) TableName() string {
	return "grain_entry_material"
}
