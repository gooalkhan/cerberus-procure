package models

import (
	"math"
	"time"
)

// ComputeContainerItemStatus Container_Item의 표시 상태(파생)를 계산한다.
//
//  1. BL 미연결                       → "Expected"
//  2. now < ETD                       → "BL Released"
//  3. ETD <= now < ETA                → "Shipping"
//  4. now >= ETA && landed(GR) == 0   → "Arrived"
//  5. 0 < landed < Load_Qty           → "Partially Landed"
//  6. landed >= Load_Qty              → "Landed"
func ComputeContainerItemStatus(hasBL bool, etd, eta time.Time, loadQty, landedQty float64, now time.Time) string {
	if !hasBL {
		return "Expected"
	}
	if now.Before(etd) {
		return "BL Released"
	}
	if now.Before(eta) {
		return "Shipping"
	}
	if landedQty <= 0 {
		return "Arrived"
	}
	if loadQty > 0 && landedQty < loadQty {
		return "Partially Landed"
	}
	return "Landed"
}

// ComputeClearanceStatus GR(입고)에 포함된 Lot들의 행정 제약 클리어 상태를 집계한다.
// totalLots가 0이면 "Not Cleared".
func ComputeClearanceStatus(clearedCount, totalLots int) string {
	if totalLots <= 0 || clearedCount <= 0 {
		return "Not Cleared"
	}
	if clearedCount >= totalLots {
		return "All Cleared"
	}
	return "Partially Cleared"
}

// ComputeAPAllocationStatus AP의 배분 상태(파생)를 계산한다.
//
//  1. Local_Amount 전액이 배분됨   → "Closed"
//  2. 일부만 배분됨                → "Open"
//  3. 배분 이력 없음               → 기존 상태 유지 (단, 이전 Closed는 "Open"으로 복원)
//
// 음수(크레딧) AP는 배분 합계가 음수 방향으로 Local_Amount에 도달하면 Closed로 본다.
func ComputeAPAllocationStatus(localAmount, allocated float64, hasAllocation bool, current string) string {
	const eps = 0.01
	if hasAllocation {
		fully := false
		switch {
		case localAmount > eps:
			fully = allocated >= localAmount-eps
		case localAmount < -eps:
			fully = allocated <= localAmount+eps
		default:
			fully = math.Abs(allocated) <= eps
		}
		if fully {
			return "Closed"
		}
		return "Open"
	}
	if current == "Closed" {
		return "Open"
	}
	return current
}
