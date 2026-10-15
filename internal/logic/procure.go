package logic

import (
	"cerberus-procure/internal/models"
	"cerberus-procure/internal/repository"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

type ProcurementUseCase struct {
	repo repository.ProcurementRepository
}

func NewProcurementUseCase(repo repository.ProcurementRepository) *ProcurementUseCase {
	return &ProcurementUseCase{repo: repo}
}

func (uc *ProcurementUseCase) SeedData() error {
	if err := uc.repo.SeedData(); err != nil {
		return err
	}
	// 시드 데이터의 파생값(상태/집계) 재계산 (트리거 이관 이후 일관성 보장)
	if pos, err := uc.repo.GetPurchaseOrders(); err == nil {
		for _, po := range pos {
			if items, err := uc.repo.GetPOItemsByPOID(po.ID); err == nil {
				for _, it := range items {
					_ = uc.recalcPOItemStatus(it.ID)
				}
			}
		}
	}
	if cs, err := uc.repo.GetContainers(); err == nil {
		for _, c := range cs {
			_ = uc.recalcContainerTotals(c.ID)
		}
	}
	return nil
}

// Item Master
func (uc *ProcurementUseCase) GetItems() ([]models.ItemMaster, error) {
	return uc.repo.GetItems()
}

func (uc *ProcurementUseCase) GetItemByID(id int) (*models.ItemMaster, error) {
	return uc.repo.GetItemByID(id)
}

func (uc *ProcurementUseCase) SaveItem(item *models.ItemMaster) error {
	// 박스 입수(Box Qty)는 필수
	if item.BoxQty <= 0 {
		return errors.New("박스 입수(Box Qty)는 필수입니다.")
	}
	// 팔렛당 박스 적재량 또는 박스 규격(가로/세로/높이) 중 하나는 필수
	hasPallet := item.BoxesPerPallet > 0
	hasDims := item.BoxWidth > 0 && item.BoxLength > 0 && item.BoxHeight > 0
	if !hasPallet && !hasDims {
		return errors.New("팔렛당 박스 적재량 또는 박스 규격(가로/세로/높이) 중 하나는 필수입니다.")
	}
	// 개당 CBM: 직접 입력이 없으면 유추 (1팔렛 = 1.536 CBM)
	if item.CBM <= 0 {
		if hasPallet {
			// 팔렛 CBM / 팔렛당 박스 수 / 박스당 입수
			item.CBM = (1.536 / float64(item.BoxesPerPallet)) / float64(item.BoxQty)
		} else {
			// 박스 규격(cm) → CBM, / 박스당 입수
			boxCBM := (item.BoxWidth * item.BoxLength * item.BoxHeight) / 1000000.0
			item.CBM = boxCBM / float64(item.BoxQty)
		}
	}
	return uc.repo.SaveItem(item)
}

// Vendor Master
func (uc *ProcurementUseCase) GetVendors() ([]models.VendorMaster, error) {
	return uc.repo.GetVendors()
}

func (uc *ProcurementUseCase) GetVendorByID(id int) (*models.VendorMaster, error) {
	return uc.repo.GetVendorByID(id)
}

func (uc *ProcurementUseCase) SaveVendor(vendor *models.VendorMaster) error {
	return uc.repo.SaveVendor(vendor)
}

// Purchase Order
func (uc *ProcurementUseCase) GetPurchaseOrders() ([]models.PurchaseOrder, error) {
	return uc.repo.GetPurchaseOrders()
}

func (uc *ProcurementUseCase) GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error) {
	return uc.repo.GetPurchaseOrderByID(id)
}

func (uc *ProcurementUseCase) SavePurchaseOrder(po *models.PurchaseOrder) error {
	if po.UUID == "" {
		po.UUID = uuid.New().String()
	}
	locked := false
	existingMap := map[int]models.POItem{}
	if po.ID > 0 {
		hasAP, hasCI, err := uc.repo.GetPOReferenceStatus(po.ID)
		if err != nil {
			return err
		}
		locked = hasAP || hasCI
		if po.Items != nil {
			existingItems, err := uc.repo.GetPOItemsByPOID(po.ID)
			if err != nil {
				return err
			}
			for _, ex := range existingItems {
				existingMap[ex.ID] = ex
			}
		}
	}
	// 참조(AP / Container Item)가 있는 PO는 수량/단가 직접수정 및 품목 추가·삭제 금지 (증분/이벤트 조정 사용)
	if locked && po.Items != nil {
		keepMap := make(map[int]bool)
		for _, it := range po.Items {
			if it.ID > 0 {
				keepMap[it.ID] = true
			}
		}
		for _, ex := range existingMap {
			if !keepMap[ex.ID] {
				return fmt.Errorf("참조(AP/Container Item)가 있는 PO는 품목을 삭제할 수 없습니다. (PO Item %d)", ex.ID)
			}
		}
		for _, it := range po.Items {
			// 신규 품목 추가는 허용(증분). 기존 품목의 수량/단가 변경은 차단.
			if it.ID == 0 {
				continue
			}
			if ex, ok := existingMap[it.ID]; ok {
				if it.POQty != ex.POQty {
					return errors.New("참조(AP/Container Item)가 있는 PO는 기존 품목의 수량을 직접 수정할 수 없습니다. 수량 조정(증분)을 사용하세요.")
				}
				if it.UnitPrice != ex.UnitPrice {
					return errors.New("참조(AP/Container Item)가 있는 PO는 기존 품목의 단가를 직접 수정할 수 없습니다. 단가 조정(이벤트)을 사용하세요.")
				}
			}
		}
	}
	err := uc.repo.SavePurchaseOrder(po)
	if err != nil {
		return err
	}
	// Reconcile items if present
	if po.Items != nil {
		if po.ID > 0 {
			for _, ex := range existingMap {
				keep := false
				for _, it := range po.Items {
					if it.ID == ex.ID {
						keep = true
						break
					}
				}
				if !keep {
					if err := uc.repo.DeletePOItem(ex.ID); err != nil {
						return fmt.Errorf("failed to delete removed PO item %d: %w", ex.ID, err)
					}
				}
			}
		}

		// 1) 기존 품목 저장 (단가는 불변, Base_Qty 유지/설정)
		for i := range po.Items {
			po.Items[i].POID = po.ID
			if po.Items[i].ID == 0 {
				continue
			}
			if locked {
				if ex, ok := existingMap[po.Items[i].ID]; ok {
					po.Items[i].BaseQty = ex.BaseQty
				}
			} else {
				po.Items[i].BaseQty = po.Items[i].POQty
			}
			if err := uc.repo.SavePOItem(&po.Items[i]); err != nil {
				return err
			}
		}

		// 2) 신규 품목(A1): 동일 (PO, SKU, 단가) 버킷이 있으면 수량 Delta로 병합, 없으면 새 버킷 생성
		for i := range po.Items {
			po.Items[i].POID = po.ID
			if po.Items[i].ID != 0 {
				continue
			}
			bucket, err := uc.repo.GetPOItemByBucket(po.ID, po.Items[i].ItemID, po.Items[i].UnitPrice)
			if err != nil {
				return err
			}
			if bucket != nil {
				delta := po.Items[i].POQty
				if delta != 0 {
					bucket.POQty += delta
					if loadQty, e := uc.repo.GetPOLoadQty(bucket.ID); e == nil {
						bucket.Status = poItemStatusFor(bucket.POQty, loadQty)
					}
					if err := uc.repo.SavePOItem(bucket); err != nil {
						return err
					}
					adj := &models.POItemAdjustment{
						POItemID:  bucket.ID,
						Kind:      "QTY",
						QtyDelta:  delta,
						OldValue:  bucket.POQty - delta,
						NewValue:  bucket.POQty,
						Reason:    po.Items[i].Remark,
						CreatedBy: po.Items[i].CreatedBy,
					}
					if err := uc.repo.SavePOItemAdjustment(adj); err != nil {
						return err
					}
				}
			} else {
				po.Items[i].BaseQty = po.Items[i].POQty
				if err := uc.repo.SavePOItem(&po.Items[i]); err != nil {
					return err
				}
			}
		}

		// 3) PO 총액은 DB 기준으로 재계산 (병합 반영)
		items, err := uc.repo.GetPOItemsByPOID(po.ID)
		if err != nil {
			return err
		}
		newTotal := 0.0
		for _, item := range items {
			newTotal += item.POQty * item.UnitPrice
		}
		if po.TotalAmount != newTotal {
			po.TotalAmount = newTotal
			if err := uc.repo.SavePurchaseOrder(po); err != nil {
				return fmt.Errorf("failed to update PO total amount: %w", err)
			}
		}

		// 4) 파생 상태 재계산 (트리거 이관): 품목 상태 → PO 상태
		for _, it := range items {
			if err := uc.recalcPOItemStatus(it.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// PO Item
func (uc *ProcurementUseCase) GetPOItemsByPOID(poID int) ([]models.POItem, error) {
	return uc.repo.GetPOItemsByPOID(poID)
}

func (uc *ProcurementUseCase) SavePOItem(item *models.POItem) error {
	if err := uc.repo.SavePOItem(item); err != nil {
		return err
	}
	return uc.recalcPOItemStatus(item.ID)
}

func (uc *ProcurementUseCase) DeletePOItem(id int) error {
	item, err := uc.repo.GetPOItemByID(id)
	if err != nil {
		return err
	}
	if err := uc.repo.DeletePOItem(id); err != nil {
		return err
	}
	if item != nil {
		return uc.recalcPOStatus(item.POID)
	}
	return nil
}

// GetPOReferenceStatus PO 편집 잠금 여부 (AP 참조 / Container Item 존재)
func (uc *ProcurementUseCase) GetPOReferenceStatus(poID int) (bool, bool, error) {
	return uc.repo.GetPOReferenceStatus(poID)
}

// GetPOItemAdjustmentsByPOItemID PO 품목의 조정 이력
func (uc *ProcurementUseCase) GetPOItemAdjustmentsByPOItemID(poItemID int) ([]models.POItemAdjustment, error) {
	return uc.repo.GetPOItemAdjustmentsByPOItemID(poItemID)
}

// GetPOItemAdjustmentsByPOID PO 단위 일괄 조회
func (uc *ProcurementUseCase) GetPOItemAdjustmentsByPOID(poID int) ([]models.POItemAdjustment, error) {
	return uc.repo.GetPOItemAdjustmentsByPOID(poID)
}

// AdjustPOItemQty PO 품목 수량을 증분(+/-)으로 조정하고 이력을 기록한다. (0까지 감소 가능 = 취소 표현)
func (uc *ProcurementUseCase) AdjustPOItemQty(adj *models.POItemAdjustment) error {
	if adj.POItemID <= 0 {
		return errors.New("PO Item ID가 필요합니다.")
	}
	if adj.QtyDelta == 0 {
		return errors.New("조정 수량(증분)이 필요합니다.")
	}
	item, err := uc.repo.GetPOItemByID(adj.POItemID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("PO 품목을 찾을 수 없습니다.")
	}
	newQty := item.POQty + adj.QtyDelta
	if newQty < 0 {
		return errors.New("조정 후 수량은 0 이상이어야 합니다.")
	}
	adj.Kind = "QTY"
	adj.OldValue = item.POQty
	adj.NewValue = newQty
	item.POQty = newQty
	// 수량 변경 시 상태 재계산 (선적량 대비)
	loadQty, err := uc.repo.GetPOLoadQty(item.ID)
	if err == nil {
		item.Status = poItemStatusFor(newQty, loadQty)
	}
	if err := uc.repo.SavePOItem(item); err != nil {
		return err
	}
	if err := uc.repo.SavePOItemAdjustment(adj); err != nil {
		return err
	}
	if err := uc.recalcPOTotal(item.POID); err != nil {
		return err
	}
	return uc.recalcPOStatus(item.POID)
}

// AdjustPOItemPrice A1에서는 단가가 불변입니다. 단가 변경은 새 버킷(새 품목) 추가로 처리합니다.
func (uc *ProcurementUseCase) AdjustPOItemPrice(adj *models.POItemAdjustment) error {
	return errors.New("A1 정책: 단가는 불변입니다. 단가 변경은 동일 SKU의 새 품목(새 단가 버킷)을 추가하고, 기존 버킷 수량을 조정하세요.")
}

// poItemStatusFor 선적량 대비 PO 품목 상태 계산
// (SQLite의 상태 트리거 CASE와 동일한 규칙을 유지해야 한다)
func poItemStatusFor(qty, loadQty float64) string {
	if loadQty > qty {
		return "Over Shipped"
	}
	if qty > 0 && loadQty >= qty {
		return "Shipped"
	}
	if loadQty > 0 {
		return "Partially Shipped"
	}
	return "Not Shipped"
}

func (uc *ProcurementUseCase) recalcPOTotal(poID int) error {
	items, err := uc.repo.GetPOItemsByPOID(poID)
	if err != nil {
		return err
	}
	po, err := uc.repo.GetPurchaseOrderByID(poID)
	if err != nil {
		return err
	}
	if po != nil {
		total := 0.0
		for _, it := range items {
			total += it.POQty * it.UnitPrice
		}
		po.TotalAmount = total
		if err := uc.repo.SavePurchaseOrder(po); err != nil {
			return err
		}
	}
	return nil
}

func (uc *ProcurementUseCase) GetCostAllocationWarningByPO(poID int) (*models.POAllocationWarning, error) {
	return uc.repo.GetCostAllocationWarningByPO(poID)
}

// Commercial Invoice
func (uc *ProcurementUseCase) GetCommercialInvoices() ([]models.CommercialInvoice, error) {
	return uc.repo.GetCommercialInvoices()
}

func (uc *ProcurementUseCase) GetCIAggregatedItems(ciID int) ([]models.CIAggregatedItem, error) {
	return uc.repo.GetCIAggregatedItems(ciID)
}

func (uc *ProcurementUseCase) SaveCommercialInvoice(ci *models.CommercialInvoice) error {
	if ci.UUID == "" {
		ci.UUID = uuid.New().String()
	}
	return uc.repo.SaveCommercialInvoice(ci)
}

// Account Payable
func (uc *ProcurementUseCase) GetAccountPayables() ([]models.AccountPayable, error) {
	return uc.repo.GetAccountPayables()
}

func (uc *ProcurementUseCase) SaveAccountPayable(ap *models.AccountPayable) error {
	// AP가 이미 비용 배분에 사용된 경우 금액(Amount/Local_Amount) 수정을 금지한다.
	if ap.ID > 0 {
		hasAllocation, _, err := uc.repo.GetAPAllocatedAmount(ap.ID)
		if err != nil {
			return err
		}
		if hasAllocation {
			existing, err := uc.repo.GetAccountPayableByID(ap.ID)
			if err != nil {
				return err
			}
			if existing != nil &&
				(math.Abs(existing.Amount-ap.Amount) > 0.005 || math.Abs(existing.LocalAmount-ap.LocalAmount) > 0.005) {
				return errors.New("이미 비용 배분이 완료된 AP는 금액(Amount/Local Amount)을 수정할 수 없습니다.")
			}
		}
	}
	if ap.UUID == "" {
		ap.UUID = uuid.New().String()
	}
	// Payment Date가 입력되면 Pay Status를 paid로 자동 변경
	if !ap.DateOfPayment.IsZero() {
		ap.Status = "paid"
	}
	return uc.repo.SaveAccountPayable(ap)
}

// Container & Logistics
func (uc *ProcurementUseCase) GetContainers() ([]models.Container, error) {
	return uc.repo.GetContainers()
}

func (uc *ProcurementUseCase) SaveContainer(c *models.Container) error {
	if c.UUID == "" {
		c.UUID = uuid.New().String()
	}
	return uc.repo.SaveContainer(c)
}

func (uc *ProcurementUseCase) GetBLs() ([]models.BL, error) {
	return uc.repo.GetBLs()
}

func (uc *ProcurementUseCase) SaveBL(bl *models.BL) error {
	if bl.UUID == "" {
		bl.UUID = uuid.New().String()
	}
	return uc.repo.SaveBL(bl)
}

// Goods Receipt & Inventory
func (uc *ProcurementUseCase) GetGoodsReceipts() ([]models.GoodsReceipt, error) {
	return uc.repo.GetGoodsReceipts()
}

func (uc *ProcurementUseCase) SaveGoodsReceipt(gr *models.GoodsReceipt) error {
	if gr.UUID == "" {
		gr.UUID = uuid.New().String()
	}
	return uc.repo.SaveGoodsReceipt(gr)
}

func (uc *ProcurementUseCase) GetInventoryLots() ([]models.InventoryLot, error) {
	return uc.repo.GetInventoryLots()
}

func (uc *ProcurementUseCase) GetInventoryLotsByGRID(grID int) ([]models.InventoryLot, error) {
	return uc.repo.GetInventoryLotsByGRID(grID)
}

func (uc *ProcurementUseCase) SaveInventoryLot(lot *models.InventoryLot) error {
	if lot.UUID == "" {
		lot.UUID = uuid.New().String()
	}
	if err := uc.repo.SaveInventoryLot(lot); err != nil {
		return err
	}
	// GR/Lot 생성 시, Container_Item의 제약 프로파일을 Lot 인스턴스로 클론(아직 없으면)
	if err := uc.cloneConstraintInstancesForLot(lot); err != nil {
		return err
	}
	return nil
}

// Cost Allocation
func (uc *ProcurementUseCase) GetCostAllocations() ([]models.CostAllocation, error) {
	return uc.repo.GetCostAllocations()
}

func (uc *ProcurementUseCase) GetAvailableCostAllocationLots(late bool) ([]models.CostAllocationLotCandidate, error) {
	var list []models.CostAllocationLotCandidate
	var err error
	if late {
		list, err = uc.repo.GetAllCostAllocationLotCandidates()
	} else {
		list, err = uc.repo.GetAvailableCostAllocationLots()
	}
	if err != nil {
		return nil, err
	}
	return uc.filterReadyCostAllocationLots(list)
}

// isLotConstraintReady 로트의 행정 제약 체인이 완료(가능일 도달)되었는지 판정한다.
// (제약 없음 / 전 단계 실제완료 / 가능일 ≤ 오늘)
func (uc *ProcurementUseCase) isLotConstraintReady(lotID int) (bool, error) {
	sched, err := uc.GetLotConstraintSchedule(lotID)
	if err != nil {
		return false, err
	}
	if len(sched.Instances) == 0 {
		return true, nil
	}
	if sched.ReadyIsActual {
		return true, nil
	}
	if !sched.ReadyDate.IsZero() && !sched.ReadyDate.After(time.Now()) {
		return true, nil
	}
	return false, nil
}

// filterReadyCostAllocationLots Cost Allocation 가능일 미도달(행정 절차 미완료) 로트를 제외한다. (Hard gate)
func (uc *ProcurementUseCase) filterReadyCostAllocationLots(list []models.CostAllocationLotCandidate) ([]models.CostAllocationLotCandidate, error) {
	out := make([]models.CostAllocationLotCandidate, 0, len(list))
	for _, c := range list {
		ready, err := uc.isLotConstraintReady(c.LotID)
		if err != nil {
			return nil, err
		}
		if ready {
			out = append(out, c)
		}
	}
	return out, nil
}

func (uc *ProcurementUseCase) GetCostAllocationLotCandidatesByIDs(ids []int) ([]models.CostAllocationLotCandidate, error) {
	return uc.repo.GetCostAllocationLotCandidatesByIDs(ids)
}

func (uc *ProcurementUseCase) CalculateCostAllocation(lotIDs []int) ([]models.CostAllocationProposal, error) {
	selectedCandidates, err := uc.repo.GetCostAllocationLotCandidatesByIDs(lotIDs)
	if err != nil {
		return nil, err
	}
	if len(selectedCandidates) == 0 {
		return []models.CostAllocationProposal{}, nil
	}

	// Hard gate: 행정 제약 체인 미완료(가능일 미도달) 로트는 배분 불가
	for _, c := range selectedCandidates {
		ready, err := uc.isLotConstraintReady(c.LotID)
		if err != nil {
			return nil, err
		}
		if !ready {
			return nil, fmt.Errorf("행정 절차(제약 체인)가 완료되지 않은 로트입니다. Cost Allocation 가능일 이후에 다시 시도하세요. (Lot %d)", c.LotID)
		}
	}

	// Collect reference UUIDs from selected lots
	refUUIDs := make(map[string]bool)
	for _, c := range selectedCandidates {
		if c.POUUID != "" {
			refUUIDs[c.POUUID] = true
		}
		if c.CIUUID != "" {
			refUUIDs[c.CIUUID] = true
		}
		if c.BLUUID != "" {
			refUUIDs[c.BLUUID] = true
		}
		if c.ContainerUUID != "" {
			refUUIDs[c.ContainerUUID] = true
		}
		if c.ContainerItemUUID != "" {
			refUUIDs[c.ContainerItemUUID] = true
		}
		if c.GRUUID != "" {
			refUUIDs[c.GRUUID] = true
		}
		if c.LotUUID != "" {
			refUUIDs[c.LotUUID] = true
		}
	}

	if len(refUUIDs) == 0 {
		return []models.CostAllocationProposal{}, nil
	}

	// Load AP Target Groups so that group-based APs can be matched to their member references
	groups, err := uc.repo.GetAPTargetGroups()
	if err != nil {
		return nil, err
	}
	groupByUUID := make(map[string]models.AP_TargetGroup)
	groupMembers := make(map[string]map[string]bool)
	for _, g := range groups {
		if g.UUID == "" || g.ReferenceType == "" {
			continue
		}
		groupByUUID[g.UUID] = g
		set := make(map[string]bool)
		items, err := uc.repo.GetAPTargetGroupItems(g.ID)
		if err != nil {
			continue
		}
		for _, it := range items {
			if it.ReferenceUUID != "" {
				set[it.ReferenceUUID] = true
			}
		}
		groupMembers[g.UUID] = set
	}

	// Helper to get the UUID of a candidate for a given reference type
	candidateUUID := func(c models.CostAllocationLotCandidate, refType string) string {
		switch refType {
		case "PO":
			return c.POUUID
		case "CI":
			return c.CIUUID
		case "BL":
			return c.BLUUID
		case "Container":
			return c.ContainerUUID
		case "ContainerItem", "Container Item":
			return c.ContainerItemUUID
		case "GR":
			return c.GRUUID
		case "Lot":
			return c.LotUUID
		default:
			return ""
		}
	}

	// Helper to check whether a candidate belongs to a group AP's member references
	matchesGroup := func(c models.CostAllocationLotCandidate, groupUUID string) bool {
		g, ok := groupByUUID[groupUUID]
		if !ok {
			return false
		}
		return groupMembers[groupUUID][candidateUUID(c, g.ReferenceType)]
	}

	// Find APs matching reference UUIDs (or group members) and not fully allocated
	aps, err := uc.repo.GetAccountPayables()
	if err != nil {
		return nil, err
	}

	var matchedAPs []models.AccountPayable
	for _, ap := range aps {
		if ap.AllocationStatus == "Closed" {
			continue
		}
		if ap.ReferenceType == "AP_Target_Group" {
			for memberUUID := range groupMembers[ap.ReferenceUUID] {
				if refUUIDs[memberUUID] {
					matchedAPs = append(matchedAPs, ap)
					break
				}
			}
			continue
		}
		if refUUIDs[ap.ReferenceUUID] {
			matchedAPs = append(matchedAPs, ap)
		}
	}

	// Helper to compute base value for a candidate based on allocation type
	computeBase := func(c models.CostAllocationLotCandidate, allocationType string) float64 {
		switch allocationType {
		case "Weight":
			if c.GrossWeight > 0 {
				return c.GrossWeight
			}
			return c.Qty
		case "Volume":
			if c.CBM > 0 {
				return c.CBM
			}
			return c.Qty
		case "Quantity":
			return c.Qty
		case "Value":
			return c.Qty * c.UnitPrice
		case "Row":
			return 1
		default:
			return c.Qty
		}
	}

	// Helper to compute base value from a base unit based on allocation type
	computeBaseFromUnit := func(u models.CostAllocationBaseUnit, allocationType string) float64 {
		switch allocationType {
		case "Weight":
			if u.GrossWeight > 0 {
				return u.GrossWeight
			}
			return u.Qty
		case "Volume":
			if u.CBM > 0 {
				return u.CBM
			}
			return u.Qty
		case "Quantity":
			return u.Qty
		case "Value":
			return u.Qty * u.UnitPrice
		case "Row":
			return 1
		default:
			return u.Qty
		}
	}

	// Helper to check if a candidate belongs to an AP reference
	matchesReference := func(c models.CostAllocationLotCandidate, refUUID string, refType string) bool {
		switch refType {
		case "PO":
			return c.POUUID == refUUID
		case "CI":
			return c.CIUUID == refUUID
		case "BL":
			return c.BLUUID == refUUID
		case "Container":
			return c.ContainerUUID == refUUID
		case "ContainerItem", "Container Item":
			return c.ContainerItemUUID == refUUID
		case "GR":
			return c.GRUUID == refUUID
		case "Lot":
			return c.LotUUID == refUUID
		default:
			return false
		}
	}

	// Helper to get reference ID from a selected candidate based on reference type
	getRefID := func(c models.CostAllocationLotCandidate, refType string) int {
		switch refType {
		case "PO":
			return c.POID
		case "CI":
			return c.CIID
		case "BL":
			return c.BLID
		case "Container":
			return c.ContainerID
		case "ContainerItem", "Container Item":
			return c.ContainerItemID
		case "GR":
			return c.GRID
		case "Lot":
			return c.LotID
		default:
			return 0
		}
	}

	// Calculate distribution for each AP
	proposals := make([]models.CostAllocationProposal, 0, len(matchedAPs))
	for _, ap := range matchedAPs {
		proposal := models.CostAllocationProposal{
			APID:           ap.ID,
			APNo:           ap.APNo,
			Amount:         ap.Amount,
			LocalAmount:    ap.LocalAmount,
			Currency:       ap.Currency,
			AllocationType: ap.AllocationType,
			Proposals:      []models.CostAllocationItemProposal{},
		}

		isGroup := ap.ReferenceType == "AP_Target_Group"
		groupRefType := ""
		if isGroup {
			if g, ok := groupByUUID[ap.ReferenceUUID]; ok {
				groupRefType = g.ReferenceType
			}
		}

		// matchesCandidate reports whether a candidate belongs to this AP's reference scope
		matchesCandidate := func(c models.CostAllocationLotCandidate) bool {
			if isGroup {
				return matchesGroup(c, ap.ReferenceUUID)
			}
			return matchesReference(c, ap.ReferenceUUID, ap.ReferenceType)
		}

		// Skip this AP if none of the selected lots fall within its scope
		hasSelected := false
		for _, c := range selectedCandidates {
			if matchesCandidate(c) {
				hasSelected = true
				break
			}
		}
		if !hasSelected {
			continue
		}

		// Compute total base for the referenced entity (or all group members), including unlanded items
		var totalRefBase float64
		if isGroup {
			for memberUUID := range groupMembers[ap.ReferenceUUID] {
				memberID, err := uc.repo.GetReferenceIDByUUID(groupRefType, memberUUID)
				if err != nil || memberID == 0 {
					continue
				}
				baseUnits, err := uc.repo.GetCostAllocationReferenceBaseUnits(groupRefType, memberID)
				if err != nil {
					continue
				}
				for _, u := range baseUnits {
					totalRefBase += computeBaseFromUnit(u, ap.AllocationType)
				}
			}
		} else {
			// Find one selected candidate that matches this AP reference to get the reference ID
			refID := 0
			for _, c := range selectedCandidates {
				if matchesCandidate(c) {
					refID = getRefID(c, ap.ReferenceType)
					break
				}
			}
			if refID == 0 {
				continue
			}
			baseUnits, err := uc.repo.GetCostAllocationReferenceBaseUnits(ap.ReferenceType, refID)
			if err != nil {
				continue
			}
			for _, u := range baseUnits {
				totalRefBase += computeBaseFromUnit(u, ap.AllocationType)
			}
		}

		if totalRefBase <= 0 {
			continue
		}

		// Amount already allocated per lot for this AP (manual/partial allocations included)
		allocatedByLot, err := uc.repo.GetCostAllocationAmountsByAP(ap.ID)
		if err != nil {
			continue
		}
		var alreadyAllocated float64
		for _, amt := range allocatedByLot {
			alreadyAllocated += amt
		}
		remainingAP := ap.LocalAmount - alreadyAllocated
		isCredit := ap.LocalAmount < 0
		// Skip APs with nothing left to distribute: positive APs fully (or over) allocated,
		// and credit notes (negative APs) that have been fully (or over) applied.
		if !isCredit && remainingAP <= 0 {
			continue
		}
		if isCredit && remainingAP >= 0 {
			continue
		}

		// For each selected lot, compute its fair share of the full AP amount based on the
		// current base ratio, then subtract what was already allocated (fixed amount).
		// This keeps the AP's distribution ratio intact even when lots are partially allocated.
		// For credit notes the fair share is negative, so the need is clamped toward zero
		// in the direction opposite to the AP sign.
		remainingNeed := make(map[int]float64)
		var needTotal float64
		for _, c := range selectedCandidates {
			if !matchesCandidate(c) {
				continue
			}
			base := computeBase(c, ap.AllocationType)
			if base == 0 {
				continue
			}
			fair := ap.LocalAmount * base / totalRefBase
			need := fair - allocatedByLot[c.LotID]
			if !isCredit && need < 0 {
				need = 0
			}
			if isCredit && need > 0 {
				need = 0
			}
			remainingNeed[c.LotID] = need
			needTotal += need
		}
		if needTotal == 0 {
			continue
		}
		if (!isCredit && needTotal <= 0) || (isCredit && needTotal >= 0) {
			continue
		}

		// Distribute the remaining AP amount. If the selected lots' combined remaining need
		// exceeds the remaining amount, scale down proportionally; otherwise allocate up to
		// their fair share and leave the excess (e.g. unselected lots' shares) unallocated.
		scale := 1.0
		if (!isCredit && needTotal > remainingAP) || (isCredit && needTotal < remainingAP) {
			scale = remainingAP / needTotal
		}
		for _, c := range selectedCandidates {
			need, ok := remainingNeed[c.LotID]
			if !ok || need == 0 {
				continue
			}
			amount := need * scale
			proposal.Proposals = append(proposal.Proposals, models.CostAllocationItemProposal{
				LotID:           c.LotID,
				AllocatedAmount: amount,
			})
		}

		proposals = append(proposals, proposal)
	}

	return proposals, nil
}

func (uc *ProcurementUseCase) SaveCostAllocation(ca *models.CostAllocation) error {
	if err := uc.repo.SaveCostAllocation(ca); err != nil {
		return err
	}
	if ca.Items != nil {
		// 수정 시 화면에서 제거된 기존 항목을 삭제 (reconciliation)
		if ca.ID > 0 {
			existing, err := uc.repo.GetCostAllocationItemsByAllocationID(ca.ID)
			if err == nil {
				keep := make(map[int]bool)
				for _, it := range ca.Items {
					if it.ID > 0 {
						keep[it.ID] = true
					}
				}
				for _, ex := range existing {
					if !keep[ex.ID] {
						if err := uc.repo.DeleteCostAllocationItem(ex.ID); err != nil {
							return err
						}
					}
				}
			}
		}
		for i := range ca.Items {
			ca.Items[i].CostAllocationID = ca.ID
			if err := uc.repo.SaveCostAllocationItem(&ca.Items[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

// Container Items
func (uc *ProcurementUseCase) GetContainerItemsByContainerID(containerID int) ([]models.ContainerItem, error) {
	return uc.repo.GetContainerItemsByContainerID(containerID)
}

func (uc *ProcurementUseCase) SaveContainerItem(item *models.ContainerItem) error {
	if item.UUID == "" {
		item.UUID = uuid.New().String()
	}
	// 업데이트 시 이전 위치 확보 (컨테이너/PO 품목 이동 대응)
	var old *models.ContainerItem
	if item.ID > 0 {
		old, _ = uc.repo.GetContainerItemByID(item.ID)
	}
	// Item_Master 기반 파생값(중량/부피) 계산 (기존 trg_container_item_calc_* 이관)
	if err := uc.applyContainerItemDerived(item); err != nil {
		return err
	}
	if err := uc.repo.SaveContainerItem(item); err != nil {
		return err
	}
	// 컨테이너 집계(총 CBM/중량) 재계산
	if err := uc.recalcContainerTotals(item.ContainerID); err != nil {
		return err
	}
	if old != nil && old.ContainerID != 0 && old.ContainerID != item.ContainerID {
		if err := uc.recalcContainerTotals(old.ContainerID); err != nil {
			return err
		}
	}
	// PO 품목 상태 재계산 (선적량 기준) + 상위 PO 상태 동기화
	if err := uc.recalcPOItemStatus(item.POItemID); err != nil {
		return err
	}
	if old != nil && old.POItemID != 0 && old.POItemID != item.POItemID {
		if err := uc.recalcPOItemStatus(old.POItemID); err != nil {
			return err
		}
	}
	return nil
}

// GetContainerItemByID 단건 조회
func (uc *ProcurementUseCase) GetContainerItemByID(id int) (*models.ContainerItem, error) {
	return uc.repo.GetContainerItemByID(id)
}

// applyContainerItemDerived Item_Master 기준으로 Gross/Net/CBM을 계산한다.
// CBM은 앱 제어 스냅샷: 신규(ID=0)이고 값이 0일 때만 유도한다.
func (uc *ProcurementUseCase) applyContainerItemDerived(item *models.ContainerItem) error {
	if item.POItemID <= 0 {
		return nil
	}
	poItem, err := uc.repo.GetPOItemByID(item.POItemID)
	if err != nil || poItem == nil {
		return err
	}
	master, err := uc.repo.GetItemByID(poItem.ItemID)
	if err != nil || master == nil {
		return err
	}
	item.GrossWeight = master.GrossWeight * item.LoadQty
	item.NetWeight = master.NetWeight * item.LoadQty
	if item.ID == 0 && item.CBM == 0 {
		item.CBM = master.CBM * item.LoadQty
	}
	return nil
}

// recalcContainerTotals 컨테이너의 총 CBM/Net/Gross를 자식 Container_Item 합계로 재계산한다.
func (uc *ProcurementUseCase) recalcContainerTotals(containerID int) error {
	if containerID <= 0 {
		return nil
	}
	items, err := uc.repo.GetContainerItemsByContainerID(containerID)
	if err != nil {
		return err
	}
	var cbm, net, gross float64
	for _, it := range items {
		cbm += it.CBM
		net += it.NetWeight
		gross += it.GrossWeight
	}
	c, err := uc.repo.GetContainerByID(containerID)
	if err != nil || c == nil {
		return err
	}
	c.TotalCBM = cbm
	c.TotalNetWgt = net
	c.TotalGrossWgt = gross
	return uc.repo.SaveContainer(c)
}

// recalcPOItemStatus PO 품목 상태를 선적량 기준으로 재계산하고 상위 PO 상태까지 동기화한다.
// (기존 trg_container_item_po_item_status_* / trg_sync_po_status 트리거 이관)
func (uc *ProcurementUseCase) recalcPOItemStatus(poItemID int) error {
	if poItemID <= 0 {
		return nil
	}
	item, err := uc.repo.GetPOItemByID(poItemID)
	if err != nil || item == nil {
		return err
	}
	load, err := uc.repo.GetPOLoadQty(poItemID)
	if err != nil {
		return err
	}
	st := poItemStatusFor(item.POQty, load)
	if item.Status != st {
		item.Status = st
		if err := uc.repo.SavePOItem(item); err != nil {
			return err
		}
	}
	return uc.recalcPOStatus(item.POID)
}

// recalcPOStatus PO 상태(Open/Closed)를 품목 상태 기준으로 재계산한다.
// 모든 품목이 Shipped/Over Shipped이고 1건 이상이면 Closed, 아니면 Open.
func (uc *ProcurementUseCase) recalcPOStatus(poID int) error {
	if poID <= 0 {
		return nil
	}
	items, err := uc.repo.GetPOItemsByPOID(poID)
	if err != nil {
		return err
	}
	po, err := uc.repo.GetPurchaseOrderByID(poID)
	if err != nil || po == nil {
		return err
	}
	newStatus := "Open"
	if len(items) > 0 {
		allTerminal := true
		for _, it := range items {
			if it.Status != "Shipped" && it.Status != "Over Shipped" {
				allTerminal = false
				break
			}
		}
		if allTerminal {
			newStatus = "Closed"
		}
	}
	if po.Status != newStatus {
		po.Status = newStatus
		return uc.repo.SavePurchaseOrder(po)
	}
	return nil
}

// Cost Allocation Items
func (uc *ProcurementUseCase) GetCostAllocationItemsByAllocationID(caID int) ([]models.CostAllocationItem, error) {
	return uc.repo.GetCostAllocationItemsByAllocationID(caID)
}

func (uc *ProcurementUseCase) SaveCostAllocationItem(item *models.CostAllocationItem) error {
	return uc.repo.SaveCostAllocationItem(item)
}
func (uc *ProcurementUseCase) GetContainersByBLID(blID int) ([]models.Container, error) {
	return uc.repo.GetContainersByBLID(blID)
}

func (uc *ProcurementUseCase) GetBookings() ([]models.BookingView, error) {
	return uc.repo.GetBookings()
}

// GetBookingTemplateData 엑셀 템플릿용 미예약 PO 항목 데이터 조회
func (uc *ProcurementUseCase) GetBookingTemplateData() ([]models.BookingTemplateRow, error) {
	return uc.repo.GetUnbookedPOItems()
}

// BulkCreateContainerItems 대량으로 Container_Item 레코드 생성
func (uc *ProcurementUseCase) BulkCreateContainerItems(rows []models.BulkImportRow) (created int, err error) {
	for _, row := range rows {
		item := &models.ContainerItem{
			POItemID:    row.POItemID,
			ContainerID: row.ContainerID,
			BLID:        row.BLID,
			LoadQty:     row.LoadQty,
			UUID:        "",
			Remark:      row.Remark,
		}

		if row.TemporaryETA != "" {
			t, parseErr := time.Parse("2006-01-02", row.TemporaryETA)
			if parseErr == nil {
				item.TemporaryETA = t
			}
		}

		if err := uc.SaveContainerItem(item); err != nil {
			return created, fmt.Errorf("failed to create container item for PO Item %d: %w", row.POItemID, err)
		}
		created++
	}
	return created, nil
}

// CheckReferences 삭제 가능 여부 확인
func (uc *ProcurementUseCase) CheckReferences(tableName string, id int) (*models.DeleteCheckResult, error) {
	return uc.repo.CheckReferences(tableName, id)
}

// DeleteRecord 레코드 삭제 (파생값 재계산 포함)
func (uc *ProcurementUseCase) DeleteRecord(tableName string, id int) error {
	// 삭제 전 파생값 재계산을 위해 연관 정보를 확보
	var ci *models.ContainerItem
	var poItem *models.POItem
	switch tableName {
	case "Container_Item":
		ci, _ = uc.repo.GetContainerItemByID(id)
	case "PO_Item":
		poItem, _ = uc.repo.GetPOItemByID(id)
	}
	if err := uc.repo.DeleteRecord(tableName, id); err != nil {
		return err
	}
	if ci != nil {
		if err := uc.recalcContainerTotals(ci.ContainerID); err != nil {
			return err
		}
		if err := uc.recalcPOItemStatus(ci.POItemID); err != nil {
			return err
		}
	}
	if poItem != nil {
		if err := uc.recalcPOStatus(poItem.POID); err != nil {
			return err
		}
	}
	return nil
}

// AP Target Group
func (uc *ProcurementUseCase) GetAPTargetGroups() ([]models.AP_TargetGroup, error) {
	return uc.repo.GetAPTargetGroups()
}

func (uc *ProcurementUseCase) GetAPTargetGroupByID(id int) (*models.AP_TargetGroup, error) {
	return uc.repo.GetAPTargetGroupByID(id)
}

func (uc *ProcurementUseCase) SaveAPTargetGroup(g *models.AP_TargetGroup) error {
	if g.UUID == "" {
		g.UUID = uuid.New().String()
	}
	if g.ReferenceType == "" {
		return errors.New("AP Target Group의 Reference Type은 필수입니다.")
	}
	// 참조 항목이 있는 그룹의 Reference Type 변경 금지 (단일 타입 제약 유지)
	var existingItems []models.AP_TargetGroupItem
	if g.ID > 0 {
		existing, err := uc.repo.GetAPTargetGroupByID(g.ID)
		if err != nil {
			return err
		}
		if existing != nil {
			existingItems = existing.Items
			if existing.ReferenceType != "" && existing.ReferenceType != g.ReferenceType && len(existing.Items) > 0 {
				return fmt.Errorf("이미 참조 항목이 있는 그룹의 Reference Type은 변경할 수 없습니다. (기존: %s)", existing.ReferenceType)
			}
		}
	}
	if err := uc.repo.SaveAPTargetGroup(g); err != nil {
		return err
	}
	// Save items if present (그룹 타입과 참조 문서 타입 일치 검증 + 제거된 항목 정리)
	if g.Items != nil {
		keep := make(map[int]bool)
		for i := range g.Items {
			g.Items[i].GroupID = g.ID
			if err := uc.validateAPTargetGroupItem(g, &g.Items[i]); err != nil {
				return err
			}
			if err := uc.repo.SaveAPTargetGroupItem(&g.Items[i]); err != nil {
				return err
			}
			if g.Items[i].ID > 0 {
				keep[g.Items[i].ID] = true
			}
		}
		for _, ex := range existingItems {
			if !keep[ex.ID] {
				if err := uc.repo.DeleteAPTargetGroupItem(ex.ID); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (uc *ProcurementUseCase) SaveAPTargetGroupItem(item *models.AP_TargetGroupItem) error {
	g, err := uc.repo.GetAPTargetGroupByID(item.GroupID)
	if err != nil {
		return err
	}
	if g == nil {
		return errors.New("대상 AP Target Group을 찾을 수 없습니다.")
	}
	if err := uc.validateAPTargetGroupItem(g, item); err != nil {
		return err
	}
	return uc.repo.SaveAPTargetGroupItem(item)
}

// Constraint Profile (행정 제약 프로파일)
func (uc *ProcurementUseCase) GetConstraintProfiles() ([]models.ConstraintProfile, error) {
	return uc.repo.GetConstraintProfiles()
}

func (uc *ProcurementUseCase) GetConstraintProfileByID(id int) (*models.ConstraintProfile, error) {
	return uc.repo.GetConstraintProfileByID(id)
}

// SaveConstraintProfile 프로파일과 단계 목록을 저장한다(제거된 단계 정리 포함).
func (uc *ProcurementUseCase) SaveConstraintProfile(p *models.ConstraintProfile) error {
	if p.UUID == "" {
		p.UUID = uuid.New().String()
	}
	if p.ProfileNo == "" {
		return errors.New("Profile No는 필수입니다.")
	}
	if p.Name == "" {
		return errors.New("프로파일 이름은 필수입니다.")
	}
	var existingSteps []models.ConstraintProfileStep
	if p.ID > 0 {
		if ex, err := uc.repo.GetConstraintProfileByID(p.ID); err == nil && ex != nil {
			existingSteps = ex.Steps
		}
	}
	if err := uc.repo.SaveConstraintProfile(p); err != nil {
		return err
	}
	if p.Steps != nil {
		keep := make(map[int]bool)
		for i := range p.Steps {
			p.Steps[i].ProfileID = p.ID
			if p.Steps[i].ConstraintType == "" && p.Steps[i].ConstraintID > 0 {
				if cm, _ := uc.repo.GetConstraintMasterByID(p.Steps[i].ConstraintID); cm != nil {
					p.Steps[i].ConstraintType = cm.Name
					if p.Steps[i].LeadTimeDays == 0 {
						p.Steps[i].LeadTimeDays = cm.LeadTimeDays
					}
				}
			}
			if p.Steps[i].ConstraintType == "" {
				return errors.New("단계 유형(Constraint Type)은 필수입니다.")
			}
			if p.Steps[i].Sequence <= 0 {
				p.Steps[i].Sequence = i + 1
			}
			if err := uc.repo.SaveConstraintProfileStep(&p.Steps[i]); err != nil {
				return err
			}
			if p.Steps[i].ID > 0 {
				keep[p.Steps[i].ID] = true
			}
		}
		for _, ex := range existingSteps {
			if !keep[ex.ID] {
				if err := uc.repo.DeleteConstraintProfileStep(ex.ID); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Constraint Master (개별 제약 카탈로그)
func (uc *ProcurementUseCase) GetConstraints() ([]models.ConstraintMaster, error) {
	return uc.repo.GetConstraints()
}

func (uc *ProcurementUseCase) GetConstraintMasterByID(id int) (*models.ConstraintMaster, error) {
	return uc.repo.GetConstraintMasterByID(id)
}

func (uc *ProcurementUseCase) SaveConstraintMaster(c *models.ConstraintMaster) error {
	if c.Name == "" {
		return errors.New("제약 이름은 필수입니다.")
	}
	return uc.repo.SaveConstraintMaster(c)
}

// ArrivalToGRLeadDays 도착→입고 표준 리드(일). 추후 룰/설정으로 승격 가능.
const ArrivalToGRLeadDays = 1

// ComputeConstraintAvailability Container_Item의 예상 Cost Allocation 가능일을 계산한다.
// 앵커 = Temporary_ETA, 그 뒤(도착→입고 + 프로파일 단계 리드)를 정방향 합산한다.
func (uc *ProcurementUseCase) ComputeConstraintAvailability(containerItemID int) (*models.ConstraintAvailability, error) {
	ci, err := uc.repo.GetContainerItemByID(containerItemID)
	if err != nil {
		return nil, err
	}
	if ci == nil {
		return nil, errors.New("Container Item을 찾을 수 없습니다.")
	}
	// 프로파일: Container_Item 선택 → 없으면 SKU(Item_Master) 기본값
	profileID := ci.ConstraintProfileID
	if profileID <= 0 {
		if pi, _ := uc.repo.GetPOItemByID(ci.POItemID); pi != nil {
			if it, _ := uc.repo.GetItemByID(pi.ItemID); it != nil {
				profileID = it.DefaultConstraintProfileID
			}
		}
	}
	res := &models.ConstraintAvailability{
		ContainerItemID: containerItemID,
		ProfileID:       profileID,
		AnchorDate:      ci.TemporaryETA,
		ArrivalToGRDays: ArrivalToGRLeadDays,
	}
	if profileID > 0 {
		if p, _ := uc.repo.GetConstraintProfileByID(profileID); p != nil {
			res.ProfileName = p.Name
			res.Steps = p.Steps
			for _, s := range p.Steps {
				res.TotalLeadDays += s.LeadTimeDays
			}
		}
	}
	res.TotalLeadDays += ArrivalToGRLeadDays
	if !res.AnchorDate.IsZero() {
		res.ExpectedReadyDate = res.AnchorDate.AddDate(0, 0, res.TotalLeadDays)
	}
	return res, nil
}

// Constraint Instance (Lot별 행정 제약 인스턴스)
func (uc *ProcurementUseCase) GetConstraintInstances(lotID int) ([]models.ConstraintInstance, error) {
	return uc.repo.GetConstraintInstances(lotID)
}

func (uc *ProcurementUseCase) SaveConstraintInstance(inst *models.ConstraintInstance) error {
	if inst.LotID <= 0 {
		return errors.New("Lot ID가 필요합니다.")
	}
	if inst.ConstraintType == "" {
		return errors.New("단계 유형(Constraint Type)이 필요합니다.")
	}
	if inst.Status == "" {
		inst.Status = "Planned"
	}
	return uc.repo.SaveConstraintInstance(inst)
}

func (uc *ProcurementUseCase) DeleteConstraintInstance(id int) error {
	return uc.repo.DeleteConstraintInstance(id)
}

// cloneConstraintInstancesForLot Lot에 제약 인스턴스가 없으면 Container_Item의 프로파일에서 클론한다.
// (선택(optional) 단계는 제외 — 필요 시 enable)
func (uc *ProcurementUseCase) cloneConstraintInstancesForLot(lot *models.InventoryLot) error {
	if lot == nil || lot.ID <= 0 || lot.ContainerItemID <= 0 {
		return nil
	}
	existing, err := uc.repo.GetConstraintInstances(lot.ID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}
	profileID := 0
	if ci, _ := uc.repo.GetContainerItemByID(lot.ContainerItemID); ci != nil {
		profileID = ci.ConstraintProfileID
		if profileID <= 0 {
			if pi, _ := uc.repo.GetPOItemByID(ci.POItemID); pi != nil {
				if it, _ := uc.repo.GetItemByID(pi.ItemID); it != nil {
					profileID = it.DefaultConstraintProfileID
				}
			}
		}
	}
	if profileID <= 0 {
		return nil
	}
	profile, _ := uc.repo.GetConstraintProfileByID(profileID)
	if profile == nil {
		return nil
	}
	var prevID int
	seq := 0
	for _, st := range profile.Steps {
		if st.IsOptional {
			continue
		}
		seq++
		inst := &models.ConstraintInstance{
			LotID:                 lot.ID,
			SourceContainerItemID: lot.ContainerItemID,
			ProfileID:             profileID,
			PredecessorID:         prevID,
			Sequence:              seq,
			ConstraintType:        st.ConstraintType,
			LeadTimeDays:          st.LeadTimeDays,
			Status:                "Planned",
			CreatedBy:             lot.CreatedBy,
		}
		if err := uc.repo.SaveConstraintInstance(inst); err != nil {
			return err
		}
		prevID = inst.ID
	}
	return nil
}

// GetLotConstraintSchedule Lot의 제약 체인 일정을 cascade 계산하여 반환한다.
// 앵커 = GR.Receive_Date. start = Actual_Start ?? (선행이면 선행 end / 시작이면 앵커), end = Actual_End ?? start + lead.
func (uc *ProcurementUseCase) GetLotConstraintSchedule(lotID int) (*models.LotConstraintSchedule, error) {
	insts, err := uc.repo.GetConstraintInstances(lotID)
	if err != nil {
		return nil, err
	}
	sched := &models.LotConstraintSchedule{LotID: lotID, Instances: insts}
	if lot, _ := uc.repo.GetInventoryLotByID(lotID); lot != nil {
		if gr, _ := uc.repo.GetGoodsReceiptByID(lot.GRID); gr != nil {
			sched.AnchorDate = gr.ReceiveDate
		}
	}
	endByID := map[int]time.Time{}
	allActual := len(insts) > 0
	var ready time.Time
	for i := range sched.Instances {
		s := &sched.Instances[i]
		var start time.Time
		if !s.ActualStart.IsZero() {
			start = s.ActualStart
		} else if s.PredecessorID != 0 {
			if e, ok := endByID[s.PredecessorID]; ok {
				start = e
			}
		} else {
			start = sched.AnchorDate
		}
		var end time.Time
		if !s.ActualEnd.IsZero() {
			end = s.ActualEnd
		} else if !start.IsZero() {
			end = start.AddDate(0, 0, s.LeadTimeDays)
		}
		s.ComputedStart = start
		s.ComputedEnd = end
		if s.ID != 0 {
			endByID[s.ID] = end
		}
		if !end.IsZero() && (ready.IsZero() || end.After(ready)) {
			ready = end
		}
		if s.ActualEnd.IsZero() {
			allActual = false
		}
	}
	sched.ReadyDate = ready
	sched.ReadyIsActual = allActual
	return sched, nil
}

// BulkAppendConstraintToShipment 같은 선적(Container_Item)의 모든 Lot 체인 끝에 단계를 추가한다.
func (uc *ProcurementUseCase) BulkAppendConstraintToShipment(containerItemID int, constraintType string, leadTimeDays int) error {
	if containerItemID <= 0 || constraintType == "" {
		return errors.New("선적 ID와 단계 유형이 필요합니다.")
	}
	lots, err := uc.repo.GetInventoryLots()
	if err != nil {
		return err
	}
	for _, lot := range lots {
		if lot.ContainerItemID != containerItemID {
			continue
		}
		insts, _ := uc.repo.GetConstraintInstances(lot.ID)
		prevID := 0
		maxSeq := 0
		if len(insts) > 0 {
			prevID = insts[len(insts)-1].ID
			maxSeq = insts[len(insts)-1].Sequence
		}
		inst := &models.ConstraintInstance{
			LotID:                 lot.ID,
			SourceContainerItemID: lot.ContainerItemID,
			PredecessorID:         prevID,
			Sequence:              maxSeq + 1,
			ConstraintType:        constraintType,
			LeadTimeDays:          leadTimeDays,
			Status:                "Planned",
		}
		if err := uc.repo.SaveConstraintInstance(inst); err != nil {
			return err
		}
	}
	return nil
}

// validateAPTargetGroupItem 그룹의 Reference_Type과 참조 문서의 타입이 일치하는지 검증한다.
// (AP Target Group은 단일 타입 문서만 그룹핑할 수 있다.)
func (uc *ProcurementUseCase) validateAPTargetGroupItem(g *models.AP_TargetGroup, item *models.AP_TargetGroupItem) error {
	if item.ReferenceUUID == "" {
		return errors.New("참조 UUID는 필수입니다.")
	}
	refID, err := uc.repo.GetReferenceIDByUUID(g.ReferenceType, item.ReferenceUUID)
	if err != nil {
		return err
	}
	if refID == 0 {
		return fmt.Errorf("이 그룹은 %s 타입 문서만 포함할 수 있습니다. (참조 UUID가 %s 문서가 아님)", g.ReferenceType, g.ReferenceType)
	}
	return nil
}

func (uc *ProcurementUseCase) DeleteAPTargetGroupItem(id int) error {
	return uc.repo.DeleteAPTargetGroupItem(id)
}

func (uc *ProcurementUseCase) GetAllReferenceTargets() ([]models.APTargetGroupReference, error) {
	return uc.repo.GetAllReferenceTargets()
}

// GetReferenceNoByUUID 참조 문서의 표시 번호(GR No, CI No 등) 조회
func (uc *ProcurementUseCase) GetReferenceNoByUUID(refType string, uuid string) (string, error) {
	return uc.repo.GetReferenceNoByUUID(refType, uuid)
}
