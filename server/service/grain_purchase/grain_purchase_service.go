package grain_purchase

import (
	"archive/zip"
	"bytes"
	baseDTO "common/base/dto"
	"common/middleware/db"
	"common/middleware/storage/image_source"
	"common/middleware/storage/oss"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	farmerImageRepository "service/farmer_image/repository"
	grainFarmerService "service/grain_farmer"
	grainFarmerRepository "service/grain_farmer/repository"
	grainPurchaseDTO "service/grain_purchase/dto"
	grainPurchaseRepository "service/grain_purchase/repository"
	permissionRepository "service/manager_permission/repository"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type GrainPurchaseService struct {
	farmerRepository         *grainFarmerRepository.GrainFarmerRepository
	idcardImageRepository    *farmerImageRepository.FarmerIDCardImageRepository
	entryRepository          *grainPurchaseRepository.GrainPurchaseEntryRepository
	snapshotRepository       *grainPurchaseRepository.GrainPurchaseEntrySnapshotRepository
	summaryRepository        *grainPurchaseRepository.GrainFarmerPurchaseSummaryRepository
	stationSummaryRepository *grainPurchaseRepository.GrainStationPurchaseSummaryRepository
	materialRepository       *grainPurchaseRepository.GrainEntryMaterialRepository
	exportBatchRepository    *grainPurchaseRepository.GrainPurchaseEntryExportBatchRepository
	roleRepository           *permissionRepository.RoleRepository
}

type GrainEntryMaterialContent struct {
	Data      []byte
	MimeType  string
	FileName  string
	StationID uint64
	Base64    string
}

type GrainEntryMaterialURL struct {
	StationID uint64
	ImageURL  string
}

func NewGrainPurchaseService() *GrainPurchaseService {
	return &GrainPurchaseService{
		farmerRepository:         db.GetRepository[grainFarmerRepository.GrainFarmerRepository](),
		idcardImageRepository:    db.GetRepository[farmerImageRepository.FarmerIDCardImageRepository](),
		entryRepository:          db.GetRepository[grainPurchaseRepository.GrainPurchaseEntryRepository](),
		snapshotRepository:       db.GetRepository[grainPurchaseRepository.GrainPurchaseEntrySnapshotRepository](),
		summaryRepository:        db.GetRepository[grainPurchaseRepository.GrainFarmerPurchaseSummaryRepository](),
		stationSummaryRepository: db.GetRepository[grainPurchaseRepository.GrainStationPurchaseSummaryRepository](),
		materialRepository:       db.GetRepository[grainPurchaseRepository.GrainEntryMaterialRepository](),
		exportBatchRepository:    db.GetRepository[grainPurchaseRepository.GrainPurchaseEntryExportBatchRepository](),
		roleRepository:           db.GetRepository[permissionRepository.RoleRepository](),
	}
}

func (s *GrainPurchaseService) EnsureTable() error {
	steps := []func() error{
		s.entryRepository.EnsureTable,
		s.snapshotRepository.EnsureTable,
		s.summaryRepository.EnsureTable,
		s.stationSummaryRepository.EnsureTable,
		s.materialRepository.EnsureTable,
		s.exportBatchRepository.EnsureTable,
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (s *GrainPurchaseService) ListEntries(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainPurchaseEntryDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	prepareEntryFarmerSearchIndexes(&query)
	total, err := s.entryRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	dtos, err := s.entryRepository.ListDTOByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	if err := s.enrichEntryFarmerProfiles(dtos); err != nil {
		return nil, err
	}
	dtos = filterEntryDTOs(dtos, query.Search)
	return baseDTO.BuildPage(int(total), dtos), nil
}

func (s *GrainPurchaseService) CountEntriesForExport(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO) (*grainPurchaseDTO.GrainPurchaseEntryExportCountDTO, error) {
	prepareEntryFarmerSearchIndexes(&query)
	total, err := s.entryRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	return &grainPurchaseDTO.GrainPurchaseEntryExportCountDTO{TotalCount: int(total)}, nil
}

func (s *GrainPurchaseService) ListEntryExportBatches(userID uint64, query grainPurchaseDTO.GrainPurchaseEntryExportQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainPurchaseEntryExportBatchDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.exportBatchRepository.CountByUser(userID, query)
	if err != nil {
		return nil, err
	}
	entities, err := s.exportBatchRepository.ListByUser(userID, query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainPurchaseDTO.GrainPurchaseEntryExportBatchDTO](entities)), nil
}

func (s *GrainPurchaseService) CreateEntryExportBatch(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO, userID uint64, username string, roleIDs []uint64) (*grainPurchaseDTO.GrainPurchaseEntryExportCreateDTO, error) {
	if running, err := s.exportBatchRepository.FindLatestRunningByUser(userID); err == nil && running != nil {
		return nil, fmt.Errorf("当前用户已有进行中的导出批次：%s", running.BatchNo)
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	includeIDCardImages := s.hasAdminRole(roleIDs)
	prepareEntryFarmerSearchIndexes(&query)
	total, err := s.entryRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	filterBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	batch := &grainPurchaseRepository.GrainPurchaseEntryExportBatch{
		BatchNo:    fmt.Sprintf("GPE%s%d", now.Format("20060102150405"), now.UnixNano()%1000000),
		UserID:     userID,
		Username:   strings.TrimSpace(username),
		Status:     "pending",
		TotalCount: int(total),
		FilterJSON: string(filterBytes),
		StartedAt:  &now,
	}
	created, err := s.exportBatchRepository.Create(batch)
	if err != nil {
		return nil, err
	}
	go NewGrainPurchaseService().runEntryExport(created.Id, query, includeIDCardImages)
	return &grainPurchaseDTO.GrainPurchaseEntryExportCreateDTO{
		TotalCount: int(total),
		Batch:      db.ToDTO[grainPurchaseDTO.GrainPurchaseEntryExportBatchDTO](created),
	}, nil
}

func (s *GrainPurchaseService) hasAdminRole(roleIDs []uint64) bool {
	if len(roleIDs) == 0 || s.roleRepository == nil || s.roleRepository.Db == nil {
		return false
	}
	var count int64
	if err := s.roleRepository.Db.
		Model(&permissionRepository.Role{}).
		Where("active = ? AND id IN ? AND code IN ?", 1, roleIDs, []string{"admin", "super_admin"}).
		Count(&count).Error; err != nil {
		log.Printf("[grain-entry-export] check admin role failed roleIDs=%v err=%v", roleIDs, err)
		return false
	}
	return count > 0
}

type GrainEntryExportFileContent struct {
	Data     []byte
	FileName string
	MimeType string
}

func (s *GrainPurchaseService) GetEntryExportFileContent(batchNo string, userID uint64) (*GrainEntryExportFileContent, error) {
	batch, err := s.exportBatchRepository.FindByBatchNoForUser(strings.TrimSpace(batchNo), userID)
	if err != nil {
		return nil, err
	}
	if (batch.Status != "success" && batch.Status != "partial_success") || strings.TrimSpace(batch.FilePath) == "" {
		return nil, fmt.Errorf("导出批次尚未完成")
	}
	data, err := oss.GetByKey(strings.TrimSpace(batch.FilePath))
	if err != nil {
		return nil, err
	}
	return &GrainEntryExportFileContent{
		Data:     data,
		FileName: batch.FileName,
		MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}, nil
}

func (s *GrainPurchaseService) runEntryExport(batchID int, query grainPurchaseDTO.GrainPurchaseEntryQueryDTO, includeIDCardImages bool) {
	fileName := ""
	objectKey := ""
	successCount := 0
	failCount := 0
	finishedAt := func() *time.Time {
		now := time.Now()
		return &now
	}
	if err := s.exportBatchRepository.UpdateProgress(batchID, "running", 0, 0, "", "", "", nil); err != nil {
		log.Printf("[grain-entry-export] mark running failed batchID=%d err=%v", batchID, err)
		return
	}
	fileName = fmt.Sprintf("grain_purchase_entries_%d.xlsx", batchID)
	objectPath := fmt.Sprintf("grain-purchase-entry-exports/%s/%s", time.Now().Format("20060102"), fileName)
	data, successCount, failCount, err := s.buildEntryExportWorkbook(query, batchID, includeIDCardImages, func(success, fail int) {
		_ = s.exportBatchRepository.UpdateProgress(batchID, "running", success, fail, fileName, objectKey, "", nil)
	})
	if err != nil {
		_ = s.exportBatchRepository.UpdateProgress(batchID, "failed", successCount, failCount, fileName, objectKey, err.Error(), finishedAt())
		return
	}
	if err := oss.Put(objectPath, data); err != nil {
		_ = s.exportBatchRepository.UpdateProgress(batchID, "failed", successCount, failCount, fileName, objectKey, err.Error(), finishedAt())
		return
	}
	if oss.Oss != nil {
		objectKey = oss.Oss.BuildKey(objectPath)
	} else {
		objectKey = objectPath
	}
	status := "success"
	if failCount > 0 {
		status = "partial_success"
	}
	_ = s.exportBatchRepository.UpdateProgress(batchID, status, successCount, failCount, fileName, objectKey, "", finishedAt())
}

func (s *GrainPurchaseService) buildEntryExportWorkbook(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO, batchID int, includeIDCardImages bool, onProgress func(success, fail int)) ([]byte, int, int, error) {
	materialImageColumnCount := 0
	if includeIDCardImages {
		var err error
		materialImageColumnCount, err = s.entryExportMaterialImageColumnCount(query)
		if err != nil {
			return nil, 0, 0, err
		}
	}
	imageColumnCount := entryExportImageColumnCount(includeIDCardImages, materialImageColumnCount)
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	if err := writeXLSXStaticParts(zipWriter, imageColumnCount > 0); err != nil {
		return nil, 0, 0, err
	}
	sheet, err := zipWriter.Create("xl/worksheets/sheet1.xml")
	if err != nil {
		return nil, 0, 0, err
	}
	if _, err := sheet.Write([]byte(entryExportWorksheetHeader(imageColumnCount))); err != nil {
		return nil, 0, 0, err
	}
	writeEntryExportXLSXRow(sheet, 1, entryExportHeaders(includeIDCardImages, materialImageColumnCount), 1, imageColumnCount > 0)
	pageIndex := 1
	successCount := 0
	failCount := 0
	images := make([]entryExportXLSXImage, 0)
	const exportPageSize = 500
	for {
		pageQuery := query
		pageQuery.PageIndex = pageIndex
		pageQuery.PageSize = exportPageSize
		page, err := s.ListEntries(pageQuery)
		if err != nil {
			return nil, successCount, failCount, err
		}
		materialsByEntryID := map[uint64][]*grainPurchaseRepository.GrainEntryMaterial{}
		if includeIDCardImages && materialImageColumnCount > 0 {
			materialsByEntryID, err = s.entryExportMaterialsByEntries(page.Data)
			if err != nil {
				return nil, successCount, failCount, err
			}
		}
		for _, item := range page.Data {
			if item == nil {
				failCount++
				continue
			}
			rowIndex := successCount + failCount + 2
			if err := writeEntryExportXLSXRow(sheet, rowIndex, entryExportRow(item, successCount+failCount+1, includeIDCardImages, materialImageColumnCount), 2, imageColumnCount > 0); err != nil {
				failCount++
				continue
			}
			if includeIDCardImages {
				if image, err := s.entryExportIDCardImage(item, rowIndex); err == nil && image != nil {
					images = append(images, *image)
				} else if err != nil {
					log.Printf("[grain-entry-export] skip idcard image batchID=%d entryID=%d farmerID=%d err=%v", batchID, item.Id, item.FarmerID, err)
				}
				materialImages := s.entryExportMaterialImages(materialsByEntryID[uint64(item.Id)], rowIndex, entryExportMaterialImageStartColumn(includeIDCardImages))
				for _, image := range materialImages {
					if image.Err != nil {
						log.Printf("[grain-entry-export] skip material image batchID=%d entryID=%d materialID=%d err=%v", batchID, item.Id, image.MaterialID, image.Err)
						continue
					}
					images = append(images, image.Image)
				}
			}
			successCount++
		}
		if onProgress != nil {
			onProgress(successCount, failCount)
		}
		if pageIndex*exportPageSize >= page.Total {
			break
		}
		pageIndex++
	}
	if _, err := sheet.Write([]byte(entryExportWorksheetFooter(imageColumnCount > 0))); err != nil {
		return nil, successCount, failCount, err
	}
	if imageColumnCount > 0 {
		if err := writeEntryExportXLSXImages(zipWriter, images); err != nil {
			return nil, successCount, failCount, err
		}
	}
	if err := zipWriter.Close(); err != nil {
		return nil, successCount, failCount, err
	}
	log.Printf("[grain-entry-export] xlsx generated batchID=%d bytes=%d success=%d fail=%d", batchID, buf.Len(), successCount, failCount)
	return buf.Bytes(), successCount, failCount, nil
}

func entryExportHeaders(includeIDCardImages bool, materialImageColumnCount int) []string {
	headers := []string{
		"序号",
		"买方信息\n姓名",
		"农户姓名",
		"农户住址",
		"农户电话",
		"农户身份证\n号码",
		"购粮品种",
		"购粮数量\n(公斤)",
		"计量单价",
		"购粮金额",
		"收购时间",
		"收购地点",
		"付款方式",
		"农户银行卡账号",
		"农户收款人姓名",
		"付款账号\n（粮站公户账号）",
		"付款日期",
		"备注",
	}
	if includeIDCardImages {
		headers = append(headers, "身份证图片")
	}
	for i := 1; i <= materialImageColumnCount; i++ {
		headers = append(headers, fmt.Sprintf("材料图片%d", i))
	}
	return headers
}

func entryExportRow(item *grainPurchaseDTO.GrainPurchaseEntryDTO, index int, includeIDCardImages bool, materialImageColumnCount int) []string {
	row := []string{
		strconv.Itoa(index),
		item.StationName,
		item.FarmerName,
		item.FarmerAddress,
		item.FarmerPhone,
		item.FarmerIDNumber,
		item.Crop,
		fmt.Sprintf("%.3f", item.Quantity),
		fmt.Sprintf("%.4f", item.UnitPrice),
		fmt.Sprintf("%.2f", item.Amount),
		formatExportTime(item.BuyTime),
		firstNonEmpty(item.Place, item.LocationAddress),
		item.PayType,
		item.FarmerBankNumber,
		entryExportReceiptName(item),
		item.StationBankAccountNumber,
		formatExportTime(item.PayTime),
		item.Remark,
	}
	if includeIDCardImages {
		row = append(row, "")
	}
	for i := 0; i < materialImageColumnCount; i++ {
		row = append(row, "")
	}
	return row
}

func entryExportReceiptName(item *grainPurchaseDTO.GrainPurchaseEntryDTO) string {
	if item == nil {
		return ""
	}
	if isEntryBankPayment(item) {
		return item.FarmerName
	}
	return item.FarmerBankName
}

func isEntryBankPayment(item *grainPurchaseDTO.GrainPurchaseEntryDTO) bool {
	if item == nil {
		return false
	}
	text := strings.ToLower(strings.TrimSpace(item.PayType))
	return strings.Contains(text, "bank") || strings.Contains(text, "银行") || strings.Contains(text, "银行卡")
}

func formatExportTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func entryExportBaseColumnCount() int {
	return 18
}

func entryExportIDCardImageColumn() int {
	return entryExportBaseColumnCount() + 1
}

func entryExportMaterialImageStartColumn(includeIDCardImages bool) int {
	if includeIDCardImages {
		return entryExportIDCardImageColumn() + 1
	}
	return entryExportBaseColumnCount() + 1
}

func entryExportImageColumnCount(includeIDCardImages bool, materialImageColumnCount int) int {
	if materialImageColumnCount < 0 {
		materialImageColumnCount = 0
	}
	if includeIDCardImages {
		return materialImageColumnCount + 1
	}
	return materialImageColumnCount
}

func writeEntryExportXLSXRow(writer interface{ Write([]byte) (int, error) }, rowIndex int, values []string, styleID int, includeImages bool) error {
	if _, err := writer.Write([]byte(fmt.Sprintf(`<row r="%d" ht="%s" customHeight="1">`, rowIndex, rowHeight(rowIndex, includeImages)))); err != nil {
		return err
	}
	for colIndex, value := range values {
		cell := fmt.Sprintf(
			`<c r="%s%d" s="%d" t="inlineStr"><is><t>%s</t></is></c>`,
			columnName(colIndex+1),
			rowIndex,
			styleID,
			xlsxEscape(value),
		)
		if _, err := writer.Write([]byte(cell)); err != nil {
			return err
		}
	}
	_, err := writer.Write([]byte(`</row>`))
	return err
}

func rowHeight(rowIndex int, includeImages bool) string {
	if rowIndex == 1 {
		return "34"
	}
	if includeImages {
		return "82"
	}
	return "22"
}

func columnName(index int) string {
	name := ""
	for index > 0 {
		index--
		name = string(rune('A'+index%26)) + name
		index /= 26
	}
	return name
}

func xlsxEscape(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(value)
}

func entryExportWorksheetHeader(imageColumnCount int) string {
	imageCols := ""
	if imageColumnCount > 0 {
		imageStartCol := entryExportBaseColumnCount() + 1
		imageEndCol := entryExportBaseColumnCount() + imageColumnCount
		imageCols = fmt.Sprintf(`<col min="%d" max="%d" width="22" customWidth="1"/>`+"\n", imageStartCol, imageEndCol)
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>
<cols>
<col min="1" max="1" width="8" customWidth="1"/>
<col min="2" max="2" width="16" customWidth="1"/>
<col min="3" max="3" width="14" customWidth="1"/>
<col min="4" max="4" width="24" customWidth="1"/>
<col min="5" max="5" width="15" customWidth="1"/>
<col min="6" max="6" width="22" customWidth="1"/>
<col min="7" max="7" width="14" customWidth="1"/>
<col min="8" max="8" width="14" customWidth="1"/>
<col min="9" max="10" width="12" customWidth="1"/>
<col min="11" max="12" width="18" customWidth="1"/>
<col min="13" max="13" width="14" customWidth="1"/>
<col min="14" max="14" width="22" customWidth="1"/>
<col min="15" max="15" width="18" customWidth="1"/>
<col min="16" max="16" width="18" customWidth="1"/>
<col min="17" max="17" width="18" customWidth="1"/>
<col min="18" max="18" width="18" customWidth="1"/>
` + imageCols + `</cols><sheetData>`
}

func entryExportWorksheetFooter(includeImages bool) string {
	drawing := ""
	if includeImages {
		drawing = `<drawing r:id="rId1"/>`
	}
	return `</sheetData><pageMargins left="0.7" right="0.7" top="0.75" bottom="0.75" header="0.3" footer="0.3"/>` + drawing + `</worksheet>`
}

func writeXLSXStaticParts(zipWriter *zip.Writer, includeImages bool) error {
	contentTypesExtra := ""
	if includeImages {
		contentTypesExtra = `
<Default Extension="png" ContentType="image/png"/>
<Default Extension="jpg" ContentType="image/jpeg"/>
<Default Extension="jpeg" ContentType="image/jpeg"/>
<Override PartName="/xl/drawings/drawing1.xml" ContentType="application/vnd.openxmlformats-officedocument.drawing+xml"/>`
	}
	parts := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>
` + contentTypesExtra + `
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`,
		"xl/workbook.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets><sheet name="收粮明细" sheetId="1" r:id="rId1"/></sheets>
</workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`,
		"xl/styles.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font></fonts>
<fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FFD9EAD3"/><bgColor indexed="64"/></patternFill></fill></fills>
<borders count="2"><border><left/><right/><top/><bottom/><diagonal/></border><border><left style="thin"><color indexed="64"/></left><right style="thin"><color indexed="64"/></right><top style="thin"><color indexed="64"/></top><bottom style="thin"><color indexed="64"/></bottom><diagonal/></border></borders>
<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
<cellXfs count="3">
<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
<xf numFmtId="0" fontId="1" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
</cellXfs>
<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`,
		"docProps/core.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:creator>shennong</dc:creator><cp:lastModifiedBy>shennong</cp:lastModifiedBy></cp:coreProperties>`,
		"docProps/app.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"><Application>Shennong</Application></Properties>`,
	}
	if includeImages {
		parts["xl/worksheets/_rels/sheet1.xml.rels"] = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/drawing" Target="../drawings/drawing1.xml"/>
</Relationships>`
	}
	for name, content := range parts {
		part, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		if _, err := part.Write([]byte(content)); err != nil {
			return err
		}
	}
	return nil
}

type entryExportXLSXImage struct {
	RowIndex  int
	ColIndex  int
	FileIndex int
	Ext       string
	Data      []byte
}

type entryExportMaterialImageResult struct {
	Image      entryExportXLSXImage
	MaterialID int
	Err        error
}

func (s *GrainPurchaseService) entryExportIDCardImage(item *grainPurchaseDTO.GrainPurchaseEntryDTO, rowIndex int) (*entryExportXLSXImage, error) {
	if item == nil || item.FarmerID == 0 || s.idcardImageRepository == nil || s.idcardImageRepository.Db == nil {
		return nil, nil
	}
	record, err := s.idcardImageRepository.FindLatestBySide(item.FarmerID, item.AppUserID, "front")
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	data, err := getOssObject(record.OssObjectKey, record.OssURL)
	if err != nil {
		return nil, err
	}
	ext := entryExportImageExt(data, record.ImageName)
	if ext == "" {
		return nil, fmt.Errorf("unsupported idcard image type: %s", detectImageMimeType(data, record.ImageName))
	}
	return &entryExportXLSXImage{
		RowIndex: rowIndex,
		ColIndex: entryExportIDCardImageColumn(),
		Ext:      ext,
		Data:     data,
	}, nil
}

func (s *GrainPurchaseService) entryExportMaterialImageColumnCount(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO) (int, error) {
	if s.materialRepository == nil || s.materialRepository.Db == nil {
		return 0, nil
	}
	prepareEntryFarmerSearchIndexes(&query)
	return s.materialRepository.MaxActiveCountByEntryQuery(query)
}

func (s *GrainPurchaseService) entryExportMaterialsByEntries(entries []*grainPurchaseDTO.GrainPurchaseEntryDTO) (map[uint64][]*grainPurchaseRepository.GrainEntryMaterial, error) {
	result := make(map[uint64][]*grainPurchaseRepository.GrainEntryMaterial)
	if len(entries) == 0 || s.materialRepository == nil || s.materialRepository.Db == nil {
		return result, nil
	}
	entryIDs := make([]uint64, 0, len(entries))
	for _, entry := range entries {
		if entry == nil || entry.Id <= 0 {
			continue
		}
		entryIDs = append(entryIDs, uint64(entry.Id))
	}
	materials, err := s.materialRepository.ListActiveByEntryIDs(entryIDs)
	if err != nil {
		return nil, err
	}
	for _, material := range materials {
		if material == nil {
			continue
		}
		result[material.EntryID] = append(result[material.EntryID], material)
	}
	return result, nil
}

func (s *GrainPurchaseService) entryExportMaterialImages(materials []*grainPurchaseRepository.GrainEntryMaterial, rowIndex, startColumn int) []entryExportMaterialImageResult {
	results := make([]entryExportMaterialImageResult, 0, len(materials))
	for index, material := range materials {
		if material == nil {
			continue
		}
		result := entryExportMaterialImageResult{MaterialID: material.Id}
		data, err := getOssObject(material.OssObjectKey, material.OssURL)
		if err != nil {
			result.Err = err
			results = append(results, result)
			continue
		}
		ext := entryExportImageExt(data, material.FileName)
		if ext == "" {
			result.Err = fmt.Errorf("unsupported material image type: %s", detectImageMimeType(data, material.FileName))
			results = append(results, result)
			continue
		}
		result.Image = entryExportXLSXImage{
			RowIndex: rowIndex,
			ColIndex: startColumn + index,
			Ext:      ext,
			Data:     data,
		}
		results = append(results, result)
	}
	return results
}

func entryExportImageExt(data []byte, fileName string) string {
	mimeType := strings.ToLower(strings.TrimSpace(detectImageMimeType(data, fileName)))
	switch mimeType {
	case "image/png":
		return "png"
	case "image/jpeg", "image/jpg":
		return "jpg"
	default:
		switch strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), ".")) {
		case "png":
			return "png"
		case "jpg", "jpeg":
			return "jpg"
		}
	}
	return ""
}

func writeEntryExportXLSXImages(zipWriter *zip.Writer, images []entryExportXLSXImage) error {
	for i := range images {
		images[i].FileIndex = i + 1
		part, err := zipWriter.Create(fmt.Sprintf("xl/media/image%d.%s", images[i].FileIndex, images[i].Ext))
		if err != nil {
			return err
		}
		if _, err := part.Write(images[i].Data); err != nil {
			return err
		}
	}
	if err := writeEntryExportDrawing(zipWriter, images); err != nil {
		return err
	}
	return writeEntryExportDrawingRels(zipWriter, images)
}

func writeEntryExportDrawing(zipWriter *zip.Writer, images []entryExportXLSXImage) error {
	part, err := zipWriter.Create("xl/drawings/drawing1.xml")
	if err != nil {
		return err
	}
	if _, err := part.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<xdr:wsDr xmlns:xdr="http://schemas.openxmlformats.org/drawingml/2006/spreadsheetDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">`)); err != nil {
		return err
	}
	for _, image := range images {
		if _, err := part.Write([]byte(entryExportDrawingAnchor(image))); err != nil {
			return err
		}
	}
	_, err = part.Write([]byte(`</xdr:wsDr>`))
	return err
}

func entryExportDrawingAnchor(image entryExportXLSXImage) string {
	imageColZeroBased := image.ColIndex - 1
	if imageColZeroBased < 0 {
		imageColZeroBased = entryExportIDCardImageColumn() - 1
	}
	rowZeroBased := image.RowIndex - 1
	return fmt.Sprintf(`<xdr:twoCellAnchor editAs="oneCell">
<xdr:from><xdr:col>%d</xdr:col><xdr:colOff>95250</xdr:colOff><xdr:row>%d</xdr:row><xdr:rowOff>95250</xdr:rowOff></xdr:from>
<xdr:to><xdr:col>%d</xdr:col><xdr:colOff>1047750</xdr:colOff><xdr:row>%d</xdr:row><xdr:rowOff>857250</xdr:rowOff></xdr:to>
<xdr:pic>
<xdr:nvPicPr><xdr:cNvPr id="%d" name="EntryExportImage%d"/><xdr:cNvPicPr><a:picLocks noChangeAspect="1"/></xdr:cNvPicPr></xdr:nvPicPr>
<xdr:blipFill><a:blip xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" r:embed="rId%d"/><a:stretch><a:fillRect/></a:stretch></xdr:blipFill>
<xdr:spPr><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></xdr:spPr>
</xdr:pic>
<xdr:clientData/>
</xdr:twoCellAnchor>`, imageColZeroBased, rowZeroBased, imageColZeroBased, rowZeroBased, image.FileIndex, image.FileIndex, image.FileIndex)
}

func writeEntryExportDrawingRels(zipWriter *zip.Writer, images []entryExportXLSXImage) error {
	part, err := zipWriter.Create("xl/drawings/_rels/drawing1.xml.rels")
	if err != nil {
		return err
	}
	if _, err := part.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)); err != nil {
		return err
	}
	for _, image := range images {
		rel := fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image%d.%s"/>`, image.FileIndex, image.FileIndex, image.Ext)
		if _, err := part.Write([]byte(rel)); err != nil {
			return err
		}
	}
	_, err = part.Write([]byte(`</Relationships>`))
	return err
}

func (s *GrainPurchaseService) CreateEntry(req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	entity := db.ToPO[grainPurchaseRepository.GrainPurchaseEntry](req)
	normalizeEntry(entity)
	entity.Version = 1
	var result *grainPurchaseRepository.GrainPurchaseEntry
	err := s.withTransaction(func(txService *GrainPurchaseService) error {
		var err error
		result, err = txService.entryRepository.Create(entity)
		if err != nil {
			return err
		}
		if err := txService.createEntrySnapshot(result, "create", operatorAppUserID, operatorName); err != nil {
			return err
		}
		return txService.applyEntryToSummary(result, 1)
	})
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainPurchaseDTO.GrainPurchaseEntryDTO](result), nil
}

func (s *GrainPurchaseService) UpdateEntry(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	return s.updateEntry(id, req, operatorAppUserID, operatorName, 0)
}

func (s *GrainPurchaseService) UpdateEntryInStation(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string, stationID uint64) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	req.StationID = stationID
	return s.updateEntry(id, req, operatorAppUserID, operatorName, stationID)
}

func (s *GrainPurchaseService) UpdateEntryInStationForAppUser(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string, stationID uint64) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	req.StationID = stationID
	req.AppUserID = operatorAppUserID
	return s.updateEntryForAppUser(id, req, operatorAppUserID, operatorName, stationID, operatorAppUserID)
}

func (s *GrainPurchaseService) updateEntry(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string, stationID uint64) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	return s.updateEntryForAppUser(id, req, operatorAppUserID, operatorName, stationID, 0)
}

func (s *GrainPurchaseService) updateEntryForAppUser(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string, stationID, ownerAppUserID uint64) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	var result *grainPurchaseRepository.GrainPurchaseEntry
	err := s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.entryRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) || (ownerAppUserID > 0 && entity.AppUserID != ownerAppUserID) {
			return gorm.ErrRecordNotFound
		}
		previous := *entity
		previousBase := entity.BaseEntity
		copier.Copy(entity, req)
		preserveEntryBaseEntityFields(&entity.BaseEntity, previousBase)
		entity.Id = int(id)
		normalizeEntry(entity)
		if entryBusinessEqual(&previous, entity) {
			result = entity
			return nil
		}
		entity.Version++
		result, err = txService.entryRepository.SaveOrUpdate(entity)
		if err != nil {
			return err
		}
		if err := txService.createEntrySnapshot(result, "update", operatorAppUserID, operatorName); err != nil {
			return err
		}
		if err := txService.applyEntryToSummary(&previous, -1); err != nil {
			return err
		}
		return txService.applyEntryToSummary(result, 1)
	})
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainPurchaseDTO.GrainPurchaseEntryDTO](result), nil
}

func (s *GrainPurchaseService) VoidEntry(id uint, operatorAppUserID uint64, operatorName string) error {
	return s.voidEntry(id, operatorAppUserID, operatorName, 0)
}

func (s *GrainPurchaseService) VoidEntryInStation(id uint, operatorAppUserID uint64, operatorName string, stationID uint64) error {
	return s.voidEntry(id, operatorAppUserID, operatorName, stationID)
}

func (s *GrainPurchaseService) VoidEntryInStationForAppUser(id uint, operatorAppUserID uint64, operatorName string, stationID uint64) error {
	return s.voidEntryForAppUser(id, operatorAppUserID, operatorName, stationID, operatorAppUserID)
}

func (s *GrainPurchaseService) voidEntry(id uint, operatorAppUserID uint64, operatorName string, stationID uint64) error {
	return s.voidEntryForAppUser(id, operatorAppUserID, operatorName, stationID, 0)
}

func (s *GrainPurchaseService) voidEntryForAppUser(id uint, operatorAppUserID uint64, operatorName string, stationID, ownerAppUserID uint64) error {
	return s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.entryRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) || (ownerAppUserID > 0 && entity.AppUserID != ownerAppUserID) {
			return gorm.ErrRecordNotFound
		}
		previous := *entity
		entity.Status = "voided"
		entity.Version++
		if _, err := txService.entryRepository.SaveOrUpdate(entity); err != nil {
			return err
		}
		if err := txService.applyEntryToSummary(&previous, -1); err != nil {
			return err
		}
		return txService.createEntrySnapshot(entity, "void", operatorAppUserID, operatorName)
	})
}

func (s *GrainPurchaseService) DeleteEntry(id uint, operatorAppUserID uint64, operatorName string) error {
	return s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.entryRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 {
			return gorm.ErrRecordNotFound
		}
		previous := *entity
		entity.Active = 0
		entity.Version++
		if _, err := txService.entryRepository.SaveOrUpdate(entity); err != nil {
			return err
		}
		if err := txService.applyEntryToSummary(&previous, -1); err != nil {
			return err
		}
		return txService.createEntrySnapshot(entity, "delete", operatorAppUserID, operatorName)
	})
}

func (s *GrainPurchaseService) ListEntrySnapshots(query grainPurchaseDTO.GrainEntrySnapshotQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainPurchaseEntrySnapshotDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.snapshotRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.snapshotRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	if err := decryptSnapshotFarmerFields(entities); err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainPurchaseDTO.GrainPurchaseEntrySnapshotDTO](entities)), nil
}

func (s *GrainPurchaseService) ListFarmerPurchaseSummaries(query grainPurchaseDTO.GrainFarmerPurchaseSummaryQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainFarmerPurchaseSummaryDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.summaryRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.summaryRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainPurchaseDTO.GrainFarmerPurchaseSummaryDTO](entities)), nil
}

func (s *GrainPurchaseService) ListDailyFarmerSummaries(query grainPurchaseDTO.GrainFarmerDailySummaryQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainFarmerDailySummaryDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	applyTodayDefault(&query.StartDate, &query.EndDate)
	prepareDailyFarmerSearchIndexes(&query)
	total, err := s.summaryRepository.CountDailyFarmerSummaries(query)
	if err != nil {
		return nil, err
	}
	summaries, err := s.summaryRepository.ListDailyFarmerSummaries(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	if err := decryptDailySummaryFarmerFields(summaries); err != nil {
		return nil, err
	}
	summaries = filterDailySummaryDTOs(summaries, query.Search)
	return baseDTO.BuildPage(int(total), summaries), nil
}

func prepareEntryFarmerSearchIndexes(query *grainPurchaseDTO.GrainPurchaseEntryQueryDTO) {
	if query == nil {
		return
	}
	index := grainFarmerService.BuildFarmerSearchIndex(query.Search)
	query.SearchIDNumberDigest = index.IDNumberDigest
	query.SearchIDNumberLast4Digest = index.IDNumberLast4Digest
	query.SearchNameDigest = index.NameDigest
	query.SearchNamePrefixCode = index.NamePrefixCode
}

func prepareDailyFarmerSearchIndexes(query *grainPurchaseDTO.GrainFarmerDailySummaryQueryDTO) {
	if query == nil {
		return
	}
	index := grainFarmerService.BuildFarmerSearchIndex(query.Search)
	query.SearchIDNumberDigest = index.IDNumberDigest
	query.SearchIDNumberLast4Digest = index.IDNumberLast4Digest
	query.SearchNameDigest = index.NameDigest
	query.SearchNamePrefixCode = index.NamePrefixCode
}

func filterEntryDTOs(entries []*grainPurchaseDTO.GrainPurchaseEntryDTO, search string) []*grainPurchaseDTO.GrainPurchaseEntryDTO {
	search = strings.TrimSpace(search)
	if search == "" {
		return entries
	}
	filtered := make([]*grainPurchaseDTO.GrainPurchaseEntryDTO, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		if grainFarmerService.FarmerProfileMatchesKeyword(entry.FarmerName, entry.FarmerIDNumber, entry.FarmerPhone, search) ||
			strings.Contains(entry.Crop, search) ||
			strings.Contains(entry.Place, search) ||
			strings.Contains(entry.PayType, search) ||
			strings.Contains(entry.LocationAddress, search) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func filterDailySummaryDTOs(summaries []*grainPurchaseDTO.GrainFarmerDailySummaryDTO, search string) []*grainPurchaseDTO.GrainFarmerDailySummaryDTO {
	search = strings.TrimSpace(search)
	if search == "" {
		return summaries
	}
	filtered := make([]*grainPurchaseDTO.GrainFarmerDailySummaryDTO, 0, len(summaries))
	for _, summary := range summaries {
		if summary == nil {
			continue
		}
		if grainFarmerService.FarmerProfileMatchesKeyword(summary.FarmerName, summary.FarmerIDNumber, summary.FarmerPhone, search) ||
			strings.Contains(summary.FarmerAddress, search) ||
			strings.Contains(summary.MainCrop, search) {
			filtered = append(filtered, summary)
		}
	}
	return filtered
}

func (s *GrainPurchaseService) GetDashboard(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) (*grainPurchaseDTO.GrainPurchaseDashboardDTO, error) {
	applyTodayDefault(&query.StartDate, &query.EndDate)
	if s.stationSummaryRepository == nil || s.stationSummaryRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	overview, err := s.summaryRepository.DashboardOverview(query)
	if err != nil {
		return nil, err
	}
	byStation, err := s.summaryRepository.DashboardByStation(query)
	if err != nil {
		return nil, err
	}
	byCrop, err := s.summaryRepository.DashboardByCrop(query)
	if err != nil {
		return nil, err
	}
	enrichDashboardRows(byStation, overview, func(row *grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO) int { return row.FarmerCount })
	enrichDashboardRows(byCrop, overview, func(row *grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO) int { return row.FarmerCount })
	now := time.Now()
	return &grainPurchaseDTO.GrainPurchaseDashboardDTO{
		StartDate: formatDashboardDate(query.StartDate),
		EndDate:   formatDashboardDate(query.EndDate),
		StationID: query.StationID,
		Overview:  *overview,
		ByStation: byStation,
		ByCrop:    byCrop,
		Generated: &now,
	}, nil
}

func (s *GrainPurchaseService) ListMaterials(query grainPurchaseDTO.GrainEntryMaterialQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainEntryMaterialDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.materialRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.materialRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	dtos := db.ToDTOs[grainPurchaseDTO.GrainEntryMaterialDTO](entities)
	for _, dto := range dtos {
		if dto != nil && dto.Id > 0 {
			dto.ImageURL = materialDisplayURL(dto.OssObjectKey, dto.OssURL)
		}
	}
	return baseDTO.BuildPage(int(total), dtos), nil
}

func (s *GrainPurchaseService) CreateMaterial(req *grainPurchaseDTO.GrainEntryMaterialDTO) (*grainPurchaseDTO.GrainEntryMaterialDTO, error) {
	entity := db.ToPO[grainPurchaseRepository.GrainEntryMaterial](req)
	if strings.TrimSpace(entity.OssObjectKey) == "" && strings.TrimSpace(entity.OssURL) == "" {
		return nil, fmt.Errorf("oss object key or oss url is required")
	}
	if entity.EntryID == 0 {
		return nil, fmt.Errorf("entry_id is required")
	}
	if strings.TrimSpace(entity.MaterialBizType) == "" {
		entity.MaterialBizType = "entry"
	}
	entity.LastSource = normalizeImageSource(entity.LastSource)
	if strings.TrimSpace(entity.ImageHash) == "" {
		entity.ImageHash = hashImageName(entity.FileName)
	}
	var result *grainPurchaseRepository.GrainEntryMaterial
	err := s.withTransaction(func(txService *GrainPurchaseService) error {
		var created bool
		var err error
		result, created, err = txService.materialRepository.FindOrCreate(entity)
		if err != nil || !created {
			if err == nil && result != nil && entity.WXCloudURL != "" && (result.WXCloudURL != entity.WXCloudURL || result.LastSource != entity.LastSource) {
				if updateErr := txService.materialRepository.UpdateCloudSource(result.Id, entity.WXCloudURL, entity.LastSource); updateErr != nil {
					return updateErr
				}
				result.WXCloudURL = entity.WXCloudURL
				result.LastSource = entity.LastSource
			}
			return err
		}
		return txService.createMaterialChangeSnapshot(result.EntryID, "material_create")
	})
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainPurchaseDTO.GrainEntryMaterialDTO](result), nil
}

func hashImageName(name string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(name)))
}

func (s *GrainPurchaseService) enrichEntryFarmerProfiles(entries []*grainPurchaseDTO.GrainPurchaseEntryDTO) error {
	for _, entry := range entries {
		if entry == nil || entry.FarmerID == 0 {
			continue
		}
		farmer, err := s.farmerRepository.FindById(uint(entry.FarmerID))
		if err == gorm.ErrRecordNotFound {
			continue
		}
		if err != nil {
			return err
		}
		if farmer.Active == 0 {
			continue
		}
		name, idNumber, phone, address, bankNumber, bankName, err := grainFarmerService.DecryptFarmerProfileValues(
			farmer.Name,
			farmer.IDNumber,
			farmer.Phone,
			farmer.Address,
			farmer.BankNumber,
			farmer.BankName,
		)
		if err != nil {
			return err
		}
		entry.FarmerName = name
		entry.FarmerIDNumber = idNumber
		entry.FarmerPhone = phone
		entry.FarmerAddress = address
		entry.FarmerBankNumber = bankNumber
		entry.FarmerBankName = bankName
	}
	return nil
}

func (s *GrainPurchaseService) DeleteMaterial(id uint) error {
	return s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.materialRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 {
			return gorm.ErrRecordNotFound
		}
		entryID := entity.EntryID
		if err = txService.materialRepository.DeletePhysicalByID(id); err != nil {
			return err
		}
		return txService.createMaterialChangeSnapshot(entryID, "material_delete")
	})
}

func (s *GrainPurchaseService) GetMaterialImageContent(id uint) (*GrainEntryMaterialContent, error) {
	log.Printf("[grain-material-image] get material image content start id=%d", id)
	entity, err := s.materialRepository.FindById(id)
	if err != nil {
		log.Printf("[grain-material-image] find material failed id=%d err=%v", id, err)
		return nil, err
	}
	if entity.Active == 0 {
		log.Printf("[grain-material-image] material inactive id=%d entryID=%d stationID=%d", id, entity.EntryID, entity.StationID)
		return nil, gorm.ErrRecordNotFound
	}
	log.Printf("[grain-material-image] material record found id=%d entryID=%d stationID=%d appUserID=%d materialBizType=%s materialType=%s fileName=%s mimeType=%s ossBucket=%s ossObjectKey=%s fallbackURL=%s", id, entity.EntryID, entity.StationID, entity.AppUserID, entity.MaterialBizType, entity.MaterialType, entity.FileName, entity.MimeType, entity.OssBucket, entity.OssObjectKey, safeURLForLog(entity.OssURL))
	data, err := getOssObject(entity.OssObjectKey, entity.OssURL)
	if err != nil {
		log.Printf("[grain-material-image] get oss object failed id=%d entryID=%d stationID=%d fileName=%s ossObjectKey=%s fallbackURL=%s err=%v", id, entity.EntryID, entity.StationID, entity.FileName, entity.OssObjectKey, safeURLForLog(entity.OssURL), err)
		return nil, err
	}
	mimeType := strings.TrimSpace(entity.MimeType)
	if !strings.HasPrefix(mimeType, "image/") {
		mimeType = detectImageMimeType(data, entity.FileName)
	}
	base64Content := ""
	if image_source.IsWXCloud() && normalizeImageSource(entity.LastSource) == image_source.OSS {
		base64Content = base64.StdEncoding.EncodeToString(data)
	}
	log.Printf("[grain-material-image] get material image content success id=%d entryID=%d stationID=%d fileName=%s mimeType=%s bytes=%d", id, entity.EntryID, entity.StationID, entity.FileName, mimeType, len(data))
	return &GrainEntryMaterialContent{
		Data:      data,
		MimeType:  mimeType,
		FileName:  entity.FileName,
		StationID: entity.StationID,
		Base64:    base64Content,
	}, nil
}

func (s *GrainPurchaseService) GetMaterialImageURL(id uint) (*GrainEntryMaterialURL, error) {
	entity, err := s.materialRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &GrainEntryMaterialURL{
		StationID: entity.StationID,
		ImageURL:  materialDisplayURL(entity.OssObjectKey, entity.OssURL),
	}, nil
}

func materialDisplayURL(ossObjectKey, fallbackURL string) string {
	key := strings.TrimSpace(ossObjectKey)
	if key != "" {
		expiry := 30 * time.Minute
		if oss.Oss != nil {
			if url, err := oss.Oss.GetUrlByKey(key, &expiry); err == nil {
				return url
			}
		}
		if url, err := oss.GetUrl(key, &expiry); err == nil {
			return url
		}
	}
	return strings.TrimSpace(fallbackURL)
}

func normalizeImageSource(source string) string {
	if strings.EqualFold(strings.TrimSpace(source), image_source.WXCloud) {
		return image_source.WXCloud
	}
	return image_source.OSS
}

func getOssObject(ossObjectKey, fallbackURL string) ([]byte, error) {
	key := strings.TrimSpace(ossObjectKey)
	if key == "" {
		key = objectKeyFromURL(fallbackURL)
		log.Printf("[grain-material-image] oss object key empty, parsed key from fallbackURL parsedKey=%s fallbackURL=%s", key, safeURLForLog(fallbackURL))
	}
	if key == "" {
		return nil, fmt.Errorf("oss object key is empty")
	}
	if data, err := oss.GetByKey(key); err == nil {
		log.Printf("[grain-material-image] oss get by key success key=%s bytes=%d", key, len(data))
		return data, nil
	} else {
		log.Printf("[grain-material-image] oss get by key failed key=%s err=%v", key, err)
	}
	data, err := oss.Get(key)
	if err != nil {
		log.Printf("[grain-material-image] oss get by path failed path=%s err=%v", key, err)
		return nil, err
	}
	log.Printf("[grain-material-image] oss get by path success path=%s bytes=%d", key, len(data))
	return data, nil
}

func objectKeyFromURL(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Path == "" {
		return ""
	}
	return strings.TrimLeft(parsed.Path, "/")
}

func safeURLForLog(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "<invalid-url>"
	}
	hasQuery := parsed.RawQuery != ""
	if parsed.Scheme == "" && parsed.Host == "" {
		return fmt.Sprintf("path=%s hasQuery=%t", parsed.Path, hasQuery)
	}
	return fmt.Sprintf("scheme=%s host=%s path=%s hasQuery=%t", parsed.Scheme, parsed.Host, parsed.Path, hasQuery)
}

func detectImageMimeType(data []byte, fileName string) string {
	if ext := strings.ToLower(filepath.Ext(fileName)); ext != "" {
		if mimeType := mime.TypeByExtension(ext); strings.HasPrefix(mimeType, "image/") {
			return mimeType
		}
	}
	if len(data) > 0 {
		return http.DetectContentType(data)
	}
	return "application/octet-stream"
}

func (s *GrainPurchaseService) withTransaction(fn func(*GrainPurchaseService) error) error {
	if s.entryRepository == nil || s.entryRepository.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return s.entryRepository.Db.Transaction(func(tx *gorm.DB) error {
		return fn(s.withDB(tx))
	})
}

func (s *GrainPurchaseService) withDB(tx *gorm.DB) *GrainPurchaseService {
	txService := *s

	farmerRepository := *s.farmerRepository
	farmerRepository.SetDb(tx)
	txService.farmerRepository = &farmerRepository

	entryRepository := *s.entryRepository
	entryRepository.SetDb(tx)
	txService.entryRepository = &entryRepository

	snapshotRepository := *s.snapshotRepository
	snapshotRepository.SetDb(tx)
	txService.snapshotRepository = &snapshotRepository

	summaryRepository := *s.summaryRepository
	summaryRepository.SetDb(tx)
	txService.summaryRepository = &summaryRepository

	stationSummaryRepository := *s.stationSummaryRepository
	stationSummaryRepository.SetDb(tx)
	txService.stationSummaryRepository = &stationSummaryRepository

	materialRepository := *s.materialRepository
	materialRepository.SetDb(tx)
	txService.materialRepository = &materialRepository

	return &txService
}

func (s *GrainPurchaseService) createEntrySnapshot(entry *grainPurchaseRepository.GrainPurchaseEntry, action string, operatorAppUserID uint64, operatorName string) error {
	now := time.Now()
	materialCount, materialDigest, materialSummary, err := s.entryMaterialSnapshot(entry.Id)
	if err != nil {
		return err
	}
	snapshot := &grainPurchaseRepository.GrainPurchaseEntrySnapshot{
		EntryID:           uint64(entry.Id),
		SnapshotVersion:   entry.Version,
		SnapshotAction:    action,
		SnapshotTime:      &now,
		OperatorAppUserID: operatorAppUserID,
		OperatorName:      strings.TrimSpace(operatorName),
		StationID:         entry.StationID,
		AppUserID:         entry.AppUserID,
		FarmerID:          entry.FarmerID,
		PurchaseTypeID:    entry.PurchaseTypeID,
		Crop:              entry.Crop,
		Quantity:          entry.Quantity,
		Unit:              entry.Unit,
		Amount:            entry.Amount,
		UnitPrice:         entry.UnitPrice,
		BuyTime:           entry.BuyTime,
		PayTime:           entry.PayTime,
		PlaceID:           entry.PlaceID,
		Place:             entry.Place,
		LocationName:      entry.LocationName,
		LocationAddress:   entry.LocationAddress,
		Longitude:         entry.Longitude,
		Latitude:          entry.Latitude,
		Province:          entry.Province,
		City:              entry.City,
		District:          entry.District,
		PaymentMethodID:   entry.PaymentMethodID,
		PayType:           entry.PayType,
		EntryStatus:       entry.Status,
		EntryRemark:       entry.Remark,
		MaterialCount:     materialCount,
		MaterialDigest:    materialDigest,
		MaterialSummary:   materialSummary,
	}
	if farmer, err := s.farmerRepository.FindById(uint(entry.FarmerID)); err == nil && farmer.Active == 1 {
		snapshot.FarmerName = farmer.Name
		snapshot.FarmerIDNumber = farmer.IDNumber
		snapshot.FarmerPhone = farmer.Phone
		snapshot.FarmerAddress = farmer.Address
		snapshot.FarmerBankNumber = farmer.BankNumber
		snapshot.FarmerBankName = farmer.BankName
	}
	_, err = s.snapshotRepository.Create(snapshot)
	return err
}

func (s *GrainPurchaseService) createMaterialChangeSnapshot(entryID uint64, action string) error {
	entry, err := s.entryRepository.FindById(uint(entryID))
	if err != nil {
		return err
	}
	if entry.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entry.Version++
	if _, err = s.entryRepository.SaveOrUpdate(entry); err != nil {
		return err
	}
	return s.createEntrySnapshot(entry, action, entry.AppUserID, "")
}

type entryMaterialSnapshotItem struct {
	ID              int    `json:"id"`
	MaterialBizType string `json:"materialBizType"`
	MaterialType    string `json:"materialType"`
	FileName        string `json:"fileName"`
	ImageHash       string `json:"imageHash"`
	SortOrder       int    `json:"sortOrder"`
}

type entryMaterialSnapshotSummary struct {
	Count int                         `json:"count"`
	Items []entryMaterialSnapshotItem `json:"items"`
}

func (s *GrainPurchaseService) entryMaterialSnapshot(entryID int) (int, string, string, error) {
	materials, err := s.materialRepository.ListActiveByEntryID(uint64(entryID))
	if err != nil {
		return 0, "", "", err
	}
	summary := entryMaterialSnapshotSummary{Count: len(materials), Items: make([]entryMaterialSnapshotItem, 0, len(materials))}
	hash := sha256.New()
	for _, material := range materials {
		if material == nil {
			continue
		}
		item := entryMaterialSnapshotItem{
			ID:              material.Id,
			MaterialBizType: strings.TrimSpace(material.MaterialBizType),
			MaterialType:    strings.TrimSpace(material.MaterialType),
			FileName:        strings.TrimSpace(material.FileName),
			ImageHash:       strings.TrimSpace(material.ImageHash),
			SortOrder:       material.SortOrder,
		}
		summary.Items = append(summary.Items, item)
		fmt.Fprintf(hash, "%d|%s|%s|%s|%s|%d\n", item.ID, item.MaterialBizType, item.MaterialType, item.FileName, item.ImageHash, item.SortOrder)
	}
	data, err := json.Marshal(summary)
	if err != nil {
		return 0, "", "", err
	}
	return summary.Count, fmt.Sprintf("%x", hash.Sum(nil)), string(data), nil
}

func (s *GrainPurchaseService) applyEntryToSummary(entry *grainPurchaseRepository.GrainPurchaseEntry, sign int) error {
	if entry == nil || entry.Active == 0 || entry.Status == "voided" || entry.Status == "deleted" {
		return nil
	}
	if sign == 0 {
		return nil
	}
	if err := s.applyEntryToFarmerSummary(entry, sign); err != nil {
		return err
	}
	return s.applyEntryToStationSummary(entry, sign)
}

func (s *GrainPurchaseService) applyEntryToFarmerSummary(entry *grainPurchaseRepository.GrainPurchaseEntry, sign int) error {
	summaryDate := entryCreatedSummaryDay(entry)
	deltaCount := sign
	deltaAmount := float64(sign) * entry.Amount
	deltaQuantity := float64(sign) * entry.Quantity
	dimension := &grainPurchaseRepository.GrainFarmerPurchaseSummary{
		StationID:       entry.StationID,
		AppUserID:       entry.AppUserID,
		PurchaseTypeID:  entry.PurchaseTypeID,
		Crop:            entry.Crop,
		SummaryDate:     &summaryDate,
		FarmerID:        entry.FarmerID,
		PaymentMethodID: entry.PaymentMethodID,
		PayType:         entry.PayType,
		EntryCount:      deltaCount,
		TotalAmount:     deltaAmount,
		TotalQuantity:   deltaQuantity,
	}
	existing, err := s.summaryRepository.FindByDimension(dimension)
	if err == gorm.ErrRecordNotFound {
		if sign < 0 {
			return nil
		}
		_, err = s.summaryRepository.Create(dimension)
		return err
	}
	if err != nil {
		return err
	}
	existing.AppUserID = entry.AppUserID
	existing.Crop = entry.Crop
	existing.PayType = entry.PayType
	existing.EntryCount += deltaCount
	existing.TotalAmount += deltaAmount
	existing.TotalQuantity += deltaQuantity
	if existing.EntryCount < 0 {
		existing.EntryCount = 0
	}
	if existing.TotalAmount < 0 {
		existing.TotalAmount = 0
	}
	if existing.TotalQuantity < 0 {
		existing.TotalQuantity = 0
	}
	_, err = s.summaryRepository.SaveOrUpdate(existing)
	return err
}

func (s *GrainPurchaseService) applyEntryToStationSummary(entry *grainPurchaseRepository.GrainPurchaseEntry, sign int) error {
	summaryDate := entryCreatedSummaryDay(entry)
	deltaCount := sign
	deltaAmount := float64(sign) * entry.Amount
	deltaQuantity := float64(sign) * entry.Quantity
	dimension := &grainPurchaseRepository.GrainStationPurchaseSummary{
		StationID:      entry.StationID,
		AppUserID:      entry.AppUserID,
		PurchaseTypeID: entry.PurchaseTypeID,
		Crop:           entry.Crop,
		SummaryDate:    &summaryDate,
		EntryCount:     deltaCount,
		TotalAmount:    deltaAmount,
		TotalQuantity:  deltaQuantity,
	}
	existing, err := s.stationSummaryRepository.FindByDimension(dimension)
	if err == gorm.ErrRecordNotFound {
		if sign < 0 {
			return nil
		}
		_, err = s.stationSummaryRepository.Create(dimension)
		return err
	}
	if err != nil {
		return err
	}
	existing.Crop = entry.Crop
	existing.EntryCount += deltaCount
	existing.TotalAmount += deltaAmount
	existing.TotalQuantity += deltaQuantity
	if existing.EntryCount < 0 {
		existing.EntryCount = 0
	}
	if existing.TotalAmount < 0 {
		existing.TotalAmount = 0
	}
	if existing.TotalQuantity < 0 {
		existing.TotalQuantity = 0
	}
	_, err = s.stationSummaryRepository.SaveOrUpdate(existing)
	return err
}

func summaryDay(value *time.Time) time.Time {
	source := time.Now()
	if value != nil {
		source = *value
	}
	location := grainBusinessLocation()
	source = source.In(location)
	return time.Date(source.Year(), source.Month(), source.Day(), 0, 0, 0, 0, location)
}

func entryCreatedSummaryDay(entry *grainPurchaseRepository.GrainPurchaseEntry) time.Time {
	if entry == nil || entry.CreatedTime.IsZero() {
		return summaryDay(nil)
	}
	return summaryDay(&entry.CreatedTime)
}

func applyTodayDefault(startDate, endDate **time.Time) {
	if *startDate != nil || *endDate != nil {
		return
	}
	today := summaryDay(nil)
	*startDate = &today
	*endDate = &today
}

func enrichDashboardRows(rows []*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO, overview *grainPurchaseDTO.GrainPurchaseDashboardMetricDTO, farmerCount func(*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO) int) {
	for _, row := range rows {
		if row == nil {
			continue
		}
		if strings.TrimSpace(row.Key) == "" {
			row.Key = row.Name
		}
		row.FarmerCount = farmerCount(row)
		if row.TotalQuantity > 0 {
			row.AverageUnitPrice = row.TotalAmount / row.TotalQuantity
		}
		if overview.TotalAmount > 0 {
			row.AmountShare = row.TotalAmount / overview.TotalAmount
		}
		if overview.TotalQuantity > 0 {
			row.QuantityShare = row.TotalQuantity / overview.TotalQuantity
		}
	}
}

func formatDashboardDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.In(grainBusinessLocation()).Format("2006-01-02")
}

func grainBusinessLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Local
	}
	return location
}

func decryptDailySummaryFarmerFields(summaries []*grainPurchaseDTO.GrainFarmerDailySummaryDTO) error {
	for _, summary := range summaries {
		name, idNumber, phone, address, bankNumber, bankName, err := grainFarmerService.DecryptFarmerProfileValues(
			summary.FarmerName,
			summary.FarmerIDNumber,
			summary.FarmerPhone,
			summary.FarmerAddress,
			summary.BankNumber,
			summary.BankName,
		)
		if err != nil {
			return err
		}
		summary.FarmerName = name
		summary.FarmerIDNumber = idNumber
		summary.FarmerPhone = phone
		summary.FarmerAddress = address
		summary.BankNumber = bankNumber
		summary.BankName = bankName
	}
	return nil
}

func decryptSnapshotFarmerFields(entities []*grainPurchaseRepository.GrainPurchaseEntrySnapshot) error {
	for _, entity := range entities {
		name, idNumber, phone, address, bankNumber, bankName, err := grainFarmerService.DecryptFarmerProfileValues(
			entity.FarmerName,
			entity.FarmerIDNumber,
			entity.FarmerPhone,
			entity.FarmerAddress,
			entity.FarmerBankNumber,
			entity.FarmerBankName,
		)
		if err != nil {
			return err
		}
		entity.FarmerName = name
		entity.FarmerIDNumber = idNumber
		entity.FarmerPhone = phone
		entity.FarmerAddress = address
		entity.FarmerBankNumber = bankNumber
		entity.FarmerBankName = bankName
	}
	return nil
}

func normalizePage(page, pageIndex, pageSize int) (int, int) {
	if pageIndex <= 0 {
		pageIndex = page
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return pageIndex, pageSize
}

func normalizeEntry(entity *grainPurchaseRepository.GrainPurchaseEntry) {
	if strings.TrimSpace(entity.Unit) == "" {
		entity.Unit = "公斤"
	}
	if strings.TrimSpace(entity.DisplayUnit) == "" {
		entity.DisplayUnit = "公斤"
	}
	if entity.Quantity > 0 {
		entity.UnitPrice = entity.Amount / entity.Quantity
	}
	if strings.TrimSpace(entity.Status) == "" {
		entity.Status = "submitted"
	}
	if entity.Version <= 0 {
		entity.Version = 1
	}
}

func entryBusinessEqual(left, right *grainPurchaseRepository.GrainPurchaseEntry) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.StationID == right.StationID &&
		left.AppUserID == right.AppUserID &&
		left.FarmerID == right.FarmerID &&
		left.PurchaseTypeID == right.PurchaseTypeID &&
		left.Crop == right.Crop &&
		left.Quantity == right.Quantity &&
		left.Unit == right.Unit &&
		left.DisplayUnit == right.DisplayUnit &&
		left.Amount == right.Amount &&
		left.UnitPrice == right.UnitPrice &&
		timesEqual(left.BuyTime, right.BuyTime) &&
		timesEqual(left.PayTime, right.PayTime) &&
		left.PlaceID == right.PlaceID &&
		left.Place == right.Place &&
		left.LocationName == right.LocationName &&
		left.LocationAddress == right.LocationAddress &&
		left.Longitude == right.Longitude &&
		left.Latitude == right.Latitude &&
		left.Province == right.Province &&
		left.City == right.City &&
		left.District == right.District &&
		left.PaymentMethodID == right.PaymentMethodID &&
		left.PayType == right.PayType &&
		left.Status == right.Status &&
		left.Remark == right.Remark
}

func timesEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Equal(*right)
}

func preserveEntryBaseEntityFields(base *db.BaseEntity, previous db.BaseEntity) {
	base.Active = previous.Active
	base.CreatedTime = previous.CreatedTime
	base.CreatedBy = previous.CreatedBy
}
