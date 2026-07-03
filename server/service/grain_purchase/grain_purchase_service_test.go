package grain_purchase

import (
	grainPurchaseDTO "service/grain_purchase/dto"
	"strings"
	"testing"
)

func TestEntryExportRowUsesFarmerNameAsReceiptNameForBankPayment(t *testing.T) {
	item := &grainPurchaseDTO.GrainPurchaseEntryDTO{
		FarmerName:       "张三",
		FarmerBankName:   "中国农业银行",
		FarmerBankNumber: "6228480000000000",
		PayType:          "银行卡",
	}

	row := entryExportRow(item, 1, false, 0)

	if len(row) != len(entryExportHeaders(false, 0)) {
		t.Fatalf("row length = %d, want %d", len(row), len(entryExportHeaders(false, 0)))
	}
	if got := row[14]; got != "张三" {
		t.Fatalf("receipt name = %q, want farmer name", got)
	}
}

func TestEntryExportRowUsesStoredReceiptNameForNonBankPayment(t *testing.T) {
	item := &grainPurchaseDTO.GrainPurchaseEntryDTO{
		FarmerName:       "李四",
		FarmerBankName:   "李四收款人",
		FarmerBankNumber: "wx-001",
		PayType:          "微信",
	}

	row := entryExportRow(item, 1, false, 0)

	if len(row) != len(entryExportHeaders(false, 0)) {
		t.Fatalf("row length = %d, want %d", len(row), len(entryExportHeaders(false, 0)))
	}
	if got := row[14]; got != "李四收款人" {
		t.Fatalf("receipt name = %q, want non-bank receipt name", got)
	}
	for _, header := range entryExportHeaders(false, 0) {
		if header == "农户开户行" {
			t.Fatalf("unexpected bank name header: %q", header)
		}
	}
}

func TestEntryExportAdminHeadersAppendIDCardImageColumn(t *testing.T) {
	headers := entryExportHeaders(true, 0)
	row := entryExportRow(&grainPurchaseDTO.GrainPurchaseEntryDTO{}, 1, true, 0)

	if len(headers) != 19 {
		t.Fatalf("headers length = %d, want 19", len(headers))
	}
	if len(row) != len(headers) {
		t.Fatalf("row length = %d, want %d", len(row), len(headers))
	}
	if got := headers[len(headers)-1]; got != "身份证图片" {
		t.Fatalf("last header = %q, want id card image", got)
	}
	if got := row[len(row)-1]; got != "" {
		t.Fatalf("image placeholder = %q, want empty", got)
	}
}

func TestEntryExportAdminHeadersAppendMaterialImageColumnsAfterIDCard(t *testing.T) {
	headers := entryExportHeaders(true, 3)
	row := entryExportRow(&grainPurchaseDTO.GrainPurchaseEntryDTO{}, 1, true, 3)

	if len(headers) != 22 {
		t.Fatalf("headers length = %d, want 22", len(headers))
	}
	if len(row) != len(headers) {
		t.Fatalf("row length = %d, want %d", len(row), len(headers))
	}
	if got := headers[18]; got != "身份证图片" {
		t.Fatalf("id card header = %q, want 身份证图片", got)
	}
	for i, want := range []string{"材料图片1", "材料图片2", "材料图片3"} {
		if got := headers[19+i]; got != want {
			t.Fatalf("material header %d = %q, want %q", i+1, got, want)
		}
	}
}

func TestEntryExportDrawingAnchorUsesImageColumn(t *testing.T) {
	anchor := entryExportDrawingAnchor(entryExportXLSXImage{RowIndex: 2, ColIndex: 20, FileIndex: 1})

	if !strings.Contains(anchor, "<xdr:col>19</xdr:col>") {
		t.Fatalf("anchor should use zero-based column 19 for excel column 20: %s", anchor)
	}
}
