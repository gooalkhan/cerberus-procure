package models

// BookingTemplateRow 엑셀 템플릿의 각 행을 나타내는 모델
type BookingTemplateRow struct {
	POItemID    int     `json:"po_item_id"`
	PONo        string  `json:"po_no"`
	SKUCode     string  `json:"sku_code"`
	ItemName    string  `json:"item_name"`
	VendorName  string  `json:"vendor_name"`
	OrderedQty  float64 `json:"ordered_qty"`
	RemainingQty float64 `json:"remaining_qty"`
	LoadQty     float64 `json:"load_qty"`
	UnitPrice   float64 `json:"unit_price"`
	Currency    string  `json:"currency"`
	ContainerNo string  `json:"container_no"`
	ContainerID int     `json:"container_id"`
	BLNo        string  `json:"bl_no"`
	BLID        int     `json:"bl_id"`
	TemporaryETA string `json:"temporary_eta"`
	Remark      string  `json:"remark"`
}

// BulkImportRow 클라이언트에서 업로드된 엑셀 데이터의 각 행
type BulkImportRow struct {
	POItemID     int     `json:"po_item_id"`
	LoadQty      float64 `json:"load_qty"`
	UnitPrice    float64 `json:"unit_price"`
	Currency     string  `json:"currency"`
	ContainerNo  string  `json:"container_no"`
	ContainerID  int     `json:"container_id"`
	BLNo         string  `json:"bl_no"`
	BLID         int     `json:"bl_id"`
	TemporaryETA string  `json:"temporary_eta"`
	Remark       string  `json:"remark"`
}

// BulkImportPreview 대량 가져오기 미리보기 결과
type BulkImportPreview struct {
	ValidRows   []BulkImportPreviewRow `json:"valid_rows"`
	ErrorRows   []BulkImportPreviewRow `json:"error_rows"`
	WarnRows    []BulkImportPreviewRow `json:"warn_rows"`
	TotalCount  int                    `json:"total_count"`
	ValidCount  int                    `json:"valid_count"`
	ErrorCount  int                    `json:"error_count"`
	WarnCount   int                    `json:"warn_count"`
}

// BulkImportPreviewRow 미리보기에서 각 행의 상태
type BulkImportPreviewRow struct {
	RowNumber    int     `json:"row_number"`
	POItemID     int     `json:"po_item_id"`
	PONo         string  `json:"po_no"`
	ItemName     string  `json:"item_name"`
	LoadQty      float64 `json:"load_qty"`
	RemainingQty float64 `json:"remaining_qty"`
	UnitPrice    float64 `json:"unit_price"`
	Currency     string  `json:"currency"`
	ContainerNo  string  `json:"container_no"`
	ContainerID  int     `json:"container_id"`
	BLNo         string  `json:"bl_no"`
	BLID         int     `json:"bl_id"`
	TemporaryETA string  `json:"temporary_eta"`
	Remark       string  `json:"remark"`
	Status       string  `json:"status"` // "valid", "error", "warn"
	Message      string  `json:"message"`
}