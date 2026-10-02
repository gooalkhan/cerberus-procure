package models

import "time"

// AP_Target_Group AP 청구 대상 그룹
type AP_TargetGroup struct {
	ID            int       `json:"ap_target_group_id"`
	GroupNo       string    `json:"group_no"`
	GroupName     string    `json:"group_name"`
	ReferenceType string    `json:"reference_type"`
	Status        string    `json:"status"`
	Remark        string    `json:"remark"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedBy     string    `json:"updated_by"`
	UpdatedAt     time.Time `json:"updated_at"`
	UUID          string    `json:"uuid"`
	Items         []AP_TargetGroupItem `json:"items,omitempty"`
}

// AP_Target_Group_Item 그룹에 포함된 개별 참조
type AP_TargetGroupItem struct {
	ID              int     `json:"ap_target_group_item_id"`
	GroupID         int     `json:"ap_target_group_id"`
	ReferenceUUID   string  `json:"reference_uuid"`
	AllocatedAmount float64 `json:"allocated_amount"`
	Remark          string  `json:"remark"`
}

// APTargetGroupReference 그룹에 추가 가능한 참조 타겟 정보
type APTargetGroupReference struct {
	ReferenceUUID string  `json:"reference_uuid"`
	ReferenceType string  `json:"reference_type"`
	ReferenceNo   string  `json:"reference_no"`
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}