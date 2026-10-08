package models

import (
	"time"
)

// ItemMaster 아이템 기본 속성
type ItemMaster struct {
	ID          int       `json:"item_id"`
	SKUCode     string    `json:"sku_code"`
	Name        string    `json:"name"`
	VendorID    int       `json:"vendor_id"`
	CBM         float64   `json:"cbm"`
	NetWeight   float64   `json:"net_weight"`
	GrossWeight float64   `json:"gross_weight"`
	Remark      string    `json:"remark"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedBy   string    `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VendorMaster 협력사 정보
type VendorMaster struct {
	ID            int       `json:"vendor_id"`
	Name          string    `json:"name"`
	Category      string    `json:"category"` // Supplier, Forwarder, Customs_Broker
	BusinessRegNo string    `json:"business_reg_no"`
	BankAccount   string    `json:"bank_account"`
	Remark        string    `json:"remark"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedBy     string    `json:"updated_by"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PurchaseOrder 수입 발주 기본 정보
type PurchaseOrder struct {
	ID          int       `json:"po_id"`
	PODate      time.Time `json:"po_date"`
	PONo        string    `json:"po_no"`
	VendorID    int       `json:"vendor_id"`
	Currency    string    `json:"currency"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	Remark      string    `json:"remark"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedBy   string    `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	UUID        string    `json:"uuid"`
	Items       []POItem  `json:"items,omitempty"`
}

// POItem 발주 품목 정보
type POItem struct {
	ID        int       `json:"po_item_id"`
	POID      int       `json:"po_id"`
	ItemID    int       `json:"item_id"`
	POQty     float64   `json:"po_qty"`
	UnitPrice float64   `json:"unit_price"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedBy string    `json:"updated_by"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CommercialInvoice 상업 송장 기본 정보
type CommercialInvoice struct {
	ID          int       `json:"ci_id"`
	CINo        string    `json:"ci_no"`
	InvoiceDate time.Time `json:"invoice_date"`
	VendorID    int       `json:"vendor_id"`
	Currency    string    `json:"currency"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	Remark      string    `json:"remark"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedBy   string    `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	UUID        string    `json:"uuid"`
}

// CIAggregatedItem 상업 송장 품목 집계 정보
type CIAggregatedItem struct {
	ItemID   int     `json:"item_id"`
	ItemName string  `json:"item_name"`
	TotalQty float64 `json:"total_qty"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// AccountPayable 통합 매입채무
type AccountPayable struct {
	ID               int       `json:"ap_id"`
	VendorID         int       `json:"vendor_id"`
	APNo             string    `json:"ap_no"`
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	LocalAmount      float64   `json:"local_amount"`
	AllocationType   string    `json:"allocation_type"` // Weight, Volume, Quantity, Value, Unit
	ReferenceUUID    string    `json:"reference_uuid"`
	ReferenceType    string    `json:"reference_type"` // BL, Container, PO, CI, GR, Lot, ContainerItem
	DueDate          time.Time `json:"due_date"`
	DateOfPayment    time.Time `json:"date_of_payment"`
	Status           string    `json:"status"` // paid, unpaid
	AllocationStatus string    `json:"allocation_status"`
	Remark           string    `json:"remark"`
	CreatedBy        string    `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedBy        string    `json:"updated_by"`
	UpdatedAt        time.Time `json:"updated_at"`
	UUID             string    `json:"uuid"`
}

// Container 운송 단위 정보
type Container struct {
	ID            int     `json:"container_id"`
	ContainerNo   string  `json:"container_no"`
	Remark        string  `json:"remark"`
	TotalCBM      float64 `json:"total_cbm"`
	TotalNetWgt   float64 `json:"total_net_wgt"`
	TotalGrossWgt float64 `json:"total_gross_wgt"`
	Status        string  `json:"status"`
	UUID          string  `json:"uuid"`
}

// ContainerItem 컨테이너 적재 품목 매핑
type ContainerItem struct {
	ID          int     `json:"container_item_id"`
	POItemID    int     `json:"po_item_id"`
	ContainerID int     `json:"container_id"`
	CIID        int     `json:"ci_id"`
	BLID        int     `json:"bl_id"`
	LoadQty     float64 `json:"load_qty"`
	GrossWeight float64 `json:"gross_weight"`
	NetWeight   float64 `json:"net_weight"`
	CBM          float64   `json:"cbm"`
	TemporaryETA time.Time `json:"temporary_eta"`
	UUID         string    `json:"uuid"`
	Remark       string    `json:"remark"`
}

// BL 선하증권 정보
type BL struct {
	ID         int       `json:"bl_id"`
	BLNo       string    `json:"bl_no"`
	ETD        time.Time `json:"etd"`
	ETA        time.Time `json:"eta"`
	POL        string    `json:"pol"`
	POD        string    `json:"pod"`
	Carrier    string    `json:"carrier"`
	VesselName string    `json:"vessel_name"`
	Status     string    `json:"status"`
	Remark     string    `json:"remark"`
	UUID       string    `json:"uuid"`
}

// GoodsReceipt 입고 기록
type GoodsReceipt struct {
	ID          int       `json:"gr_id"`
	ReceiveDate time.Time `json:"receive_date"`
	Remark      string    `json:"remark"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedBy   string    `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	UUID        string    `json:"uuid"`
}

// InventoryLot 재고 최소 단위 (로트)
type InventoryLot struct {
	ID                int       `json:"lot_id"`
	GRID              int       `json:"gr_id"`
	ContainerItemID   int       `json:"container_item_id"`
	LotNo             string    `json:"lot_no"`
	ExpiryDate        time.Time `json:"expiry_date"`
	Qty               float64   `json:"qty"`
	LandedCostPerUnit float64   `json:"landed_cost_per_unit"`
	QuarantineStatus  string    `json:"quarantine_status"`
	QuarantineRemark  string    `json:"quarantine_remark"`
	CreatedBy         string    `json:"created_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedBy         string    `json:"updated_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	Remark            string    `json:"remark"`
	UUID              string    `json:"uuid"`
}

// CostAllocation 랜딩 코스트 배분 헤더
type CostAllocation struct {
	ID                   int                  `json:"cost_allocation_id"`
	AllocationDate       time.Time            `json:"allocation_date"`
	TotalAllocatedAmount float64              `json:"total_allocated_amount"`
	IsLateCostAllocation bool                 `json:"is_late_cost_allocation"`
	Remark               string               `json:"remark"`
	CreatedBy            string               `json:"created_by"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedBy            string               `json:"updated_by"`
	UpdatedAt            time.Time            `json:"updated_at"`
	Items                []CostAllocationItem `json:"items,omitempty"`
}

// CostAllocationItem 랜딩 코스트 배분 상세
type CostAllocationItem struct {
	ID               int     `json:"cost_allocation_item_id"`
	CostAllocationID int     `json:"cost_allocation_id"`
	LotID            int     `json:"lot_id"`
	AllocatedAmount  float64 `json:"allocated_amount"`
	APID             int     `json:"ap_id"`
}

// CostAllocationBaseUnit 배분 기준량 단위
type CostAllocationBaseUnit struct {
	Qty         float64
	UnitPrice   float64
	GrossWeight float64
	CBM         float64
}

// CostAllocationLotCandidate 배분 대상 로트 후보
type CostAllocationLotCandidate struct {
	LotID           int     `json:"lot_id"`
	LotNo           string  `json:"lot_no"`
	LotUUID         string  `json:"lot_uuid"`
	GRID            int     `json:"gr_id"`
	GRUUID          string  `json:"gr_uuid"`
	ContainerItemID int     `json:"container_item_id"`
	ContainerItemUUID string `json:"container_item_uuid"`
	CIID            int     `json:"ci_id"`
	CIUUID          string  `json:"ci_uuid"`
	ContainerID     int     `json:"container_id"`
	ContainerUUID   string  `json:"container_uuid"`
	ContainerNo     string  `json:"container_no"`
	BLID            int     `json:"bl_id"`
	BLUUID          string  `json:"bl_uuid"`
	BLNo            string  `json:"bl_no"`
	POID            int     `json:"po_id"`
	POUUID          string  `json:"po_uuid"`
	PONo            string  `json:"po_no"`
	ItemID          int     `json:"item_id"`
	ItemName        string  `json:"item_name"`
	Qty             float64 `json:"qty"`
	GrossWeight     float64 `json:"gross_weight"`
	NetWeight       float64 `json:"net_weight"`
	CBM             float64 `json:"cbm"`
	UnitPrice       float64 `json:"unit_price"`
	Currency        string  `json:"currency"`
}

// CostAllocationProposal AP 배분 제안
type CostAllocationProposal struct {
	APID            int     `json:"ap_id"`
	APNo            string  `json:"ap_no"`
	Amount          float64 `json:"amount"`
	LocalAmount     float64 `json:"local_amount"`
	Currency        string  `json:"currency"`
	AllocationType  string  `json:"allocation_type"`
	Proposals       []CostAllocationItemProposal `json:"proposals"`
}

// CostAllocationItemProposal 개별 로트별 배분 제안
type CostAllocationItemProposal struct {
	LotID           int     `json:"lot_id"`
	AllocatedAmount float64 `json:"allocated_amount"`
}

// BookingView 물류 선적 조회를 위한 통합 뷰 모델
type BookingView struct {
	ContainerItemID int       `json:"container_item_id"`
	ContainerID     int       `json:"container_id"`
	ContainerNo     string    `json:"container_no"`
	Status          string    `json:"status"`
	TotalCBM        float64   `json:"total_cbm"`
	TotalNetWgt     float64   `json:"total_net_wgt"`
	TotalGrossWgt   float64   `json:"total_gross_wgt"`
	BLID            int       `json:"bl_id"`
	BLNo            string    `json:"bl_no"`
	BLStatus        string    `json:"bl_status"`
	ETD             time.Time `json:"etd"`
	ETA             time.Time `json:"eta"`
	POL             string    `json:"pol"`
	POD             string    `json:"pod"`
	Carrier         string    `json:"carrier"`
	VesselName      string    `json:"vessel_name"`
	POItemID        int       `json:"po_item_id"`
	POID            int       `json:"po_id"`
	PONo            string    `json:"po_no"`
	ItemID          int       `json:"item_id"`
	ItemName        string    `json:"item_name"`
	CIID            int       `json:"ci_id"`
	CINo            string    `json:"ci_no"`
	LoadQty         float64   `json:"load_qty"`
	GrossWeight     float64   `json:"gross_weight"`
	NetWeight       float64   `json:"net_weight"`
	CBM             float64   `json:"cbm"`
	TemporaryETA    time.Time `json:"temporary_eta"`
	Remark          string    `json:"remark"`
}
