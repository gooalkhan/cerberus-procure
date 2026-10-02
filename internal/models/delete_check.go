package models

// ReferenceInfo 삭제 대상 레코드를 참조하는 다른 테이블의 정보
type ReferenceInfo struct {
	TableName  string        `json:"table_name"`
	Count      int           `json:"count"`
	Records    []map[string]interface{} `json:"records"`
}

// DeleteCheckResult 삭제 가능 여부와 참조 정보
type DeleteCheckResult struct {
	CanDelete    bool            `json:"can_delete"`
	References   []ReferenceInfo `json:"references"`
	TotalRefs    int             `json:"total_refs"`
}