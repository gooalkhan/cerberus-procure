package models

import "time"

// ConstraintProfile 행정 제약 프로파일(템플릿) — SKU에 기본값으로 연결되는 제약 조합
type ConstraintProfile struct {
	ID        int                     `json:"constraint_profile_id"`
	ProfileNo string                  `json:"profile_no"`
	Name      string                  `json:"name"`
	Remark    string                  `json:"remark"`
	CreatedBy string                  `json:"created_by"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedBy string                  `json:"updated_by"`
	UpdatedAt time.Time               `json:"updated_at"`
	UUID      string                  `json:"uuid"`
	Steps     []ConstraintProfileStep `json:"steps,omitempty"`
}

// ConstraintMaster 개별 제약(step) 카탈로그 — 이름 + 기본 소요기간(일)
type ConstraintMaster struct {
	ID           int       `json:"constraint_id"`
	Name         string    `json:"name"`
	LeadTimeDays int       `json:"lead_time_days"`
	Remark       string    `json:"remark"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedBy    string    `json:"updated_by"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ConstraintProfileStep 프로파일을 구성하는 행정 절차 단계 (카탈로그 참조 + 오버라이드)
type ConstraintProfileStep struct {
	ID             int    `json:"constraint_profile_step_id"`
	ProfileID      int    `json:"constraint_profile_id"`
	ConstraintID   int    `json:"constraint_id"`   // 카탈로그 참조(Constraint_Master)
	Sequence       int    `json:"sequence"`
	ConstraintType string `json:"constraint_type"` // 이름 스냅샷(카탈로그에서 기본값, 변경 가능)
	LeadTimeDays   int    `json:"lead_time_days"`  // 유효 소요기간(기본값: 카탈로그, 오버라이드 가능)
	IsOptional     bool   `json:"is_optional"`
	Remark         string `json:"remark"`
}

// ConstraintAvailability 예상 Cost Allocation 가능일(투영 결과)
type ConstraintAvailability struct {
	ContainerItemID   int                     `json:"container_item_id"`
	ProfileID         int                     `json:"constraint_profile_id"`
	ProfileName       string                  `json:"profile_name"`
	AnchorDate        time.Time               `json:"anchor_date"` // 앵커(BL ETA / Temporary_ETA 등)
	AnchorSource      string                  `json:"anchor_source"` // 앵커 출처: "BL ETA" | "Temporary ETA" | ""
	ArrivalToGRDays   int                     `json:"arrival_to_gr_days"`
	TotalLeadDays     int                     `json:"total_lead_days"`
	ExpectedReadyDate time.Time               `json:"expected_ready_date"`
	Steps             []ConstraintProfileStep `json:"steps,omitempty"`
}

// ConstraintInstance Lot별 행정 제약 인스턴스 (GR/Lot 생성 시 프로파일에서 클론)
type ConstraintInstance struct {
	ID                    int       `json:"constraint_instance_id"`
	LotID                 int       `json:"lot_id"`
	SourceContainerItemID int       `json:"source_container_item_id"` // 클론 출처(선적)
	ProfileID             int       `json:"constraint_profile_id"`    // 클론 출처 프로파일
	PredecessorID         int       `json:"predecessor_id"`           // 선행 제약(체인 연결). 0이면 체인 시작
	Sequence              int       `json:"sequence"`                 // 표시/정렬용
	ConstraintType        string    `json:"constraint_type"`          // 요건승인 / 검역 / 통관 / 불시검문검사
	LeadTimeDays          int       `json:"lead_time_days"`
	IsOptional            bool      `json:"is_optional"`
	PlannedStart          time.Time `json:"planned_start"`
	ActualStart           time.Time `json:"actual_start"`
	PlannedEnd            time.Time `json:"planned_end"`
	ActualEnd             time.Time `json:"actual_end"`
	Status                string    `json:"status"` // Planned / InProgress / Done / Waived
	Remark                string    `json:"remark"`
	CreatedBy             string    `json:"created_by"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedBy             string    `json:"updated_by"`
	UpdatedAt             time.Time `json:"updated_at"`
	// 조회 시 계산되어 채워지는 값(cascade 결과)
	ComputedStart time.Time `json:"computed_start"`
	ComputedEnd   time.Time `json:"computed_end"`
}

// LotConstraintSchedule Lot 제약 체인 일정(cascade 계산 결과)
type LotConstraintSchedule struct {
	LotID         int                  `json:"lot_id"`
	AnchorDate    time.Time            `json:"anchor_date"`
	Instances     []ConstraintInstance `json:"instances"`
	ReadyDate     time.Time            `json:"ready_date"`
	ReadyIsActual bool                 `json:"ready_is_actual"`
}

