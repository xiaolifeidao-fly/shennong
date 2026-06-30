package grain_purchase

import (
	grainPurchaseDTO "service/grain_purchase/dto"
	"testing"
)

func TestEntryExportRowUsesFarmerNameAsReceiptNameForBankPayment(t *testing.T) {
	item := &grainPurchaseDTO.GrainPurchaseEntryDTO{
		FarmerName:       "张三",
		FarmerBankName:   "中国农业银行",
		FarmerBankNumber: "6228480000000000",
		PayType:          "银行卡",
	}

	row := entryExportRow(item, 1, false)

	if len(row) != len(entryExportHeaders(false)) {
		t.Fatalf("row length = %d, want %d", len(row), len(entryExportHeaders(false)))
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

	row := entryExportRow(item, 1, false)

	if len(row) != len(entryExportHeaders(false)) {
		t.Fatalf("row length = %d, want %d", len(row), len(entryExportHeaders(false)))
	}
	if got := row[14]; got != "李四收款人" {
		t.Fatalf("receipt name = %q, want non-bank receipt name", got)
	}
	for _, header := range entryExportHeaders(false) {
		if header == "农户开户行" {
			t.Fatalf("unexpected bank name header: %q", header)
		}
	}
}

func TestEntryExportAdminHeadersAppendIDCardImageColumn(t *testing.T) {
	headers := entryExportHeaders(true)
	row := entryExportRow(&grainPurchaseDTO.GrainPurchaseEntryDTO{}, 1, true)

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
