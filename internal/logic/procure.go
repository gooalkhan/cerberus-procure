package logic

import (
	"cerberus-procure/internal/models"
	"cerberus-procure/internal/repository"
	"fmt"
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
	return uc.repo.SeedData()
}

// Item Master
func (uc *ProcurementUseCase) GetItems() ([]models.ItemMaster, error) {
	return uc.repo.GetItems()
}

func (uc *ProcurementUseCase) GetItemByID(id int) (*models.ItemMaster, error) {
	return uc.repo.GetItemByID(id)
}

func (uc *ProcurementUseCase) SaveItem(item *models.ItemMaster) error {
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
	err := uc.repo.SavePurchaseOrder(po)
	if err != nil {
		return err
	}
	// Reconcile items if present
	if po.Items != nil {
		if po.ID > 0 {
			existingItems, err := uc.repo.GetPOItemsByPOID(po.ID)
			if err == nil {
				keepMap := make(map[int]bool)
				for _, item := range po.Items {
					if item.ID > 0 {
						keepMap[item.ID] = true
					}
				}
				for _, existing := range existingItems {
					if !keepMap[existing.ID] {
						if err := uc.repo.DeletePOItem(existing.ID); err != nil {
							return fmt.Errorf("failed to delete removed PO item %d: %w", existing.ID, err)
						}
					}
				}
			}
		}

		for i := range po.Items {
			po.Items[i].POID = po.ID
			if err := uc.repo.SavePOItem(&po.Items[i]); err != nil {
				return err
			}
		}

		// Recalculate PO total excluding cancelled items
		newTotal := 0.0
		for _, item := range po.Items {
			if item.Status != "Cancelled" {
				newTotal += item.POQty * item.UnitPrice
			}
		}
		if po.TotalAmount != newTotal {
			po.TotalAmount = newTotal
			if err := uc.repo.SavePurchaseOrder(po); err != nil {
				return fmt.Errorf("failed to update PO total amount: %w", err)
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
	return uc.repo.SavePOItem(item)
}

func (uc *ProcurementUseCase) DeletePOItem(id int) error {
	return uc.repo.DeletePOItem(id)
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
	return uc.repo.SaveInventoryLot(lot)
}

// Cost Allocation
func (uc *ProcurementUseCase) GetCostAllocations() ([]models.CostAllocation, error) {
	return uc.repo.GetCostAllocations()
}

func (uc *ProcurementUseCase) GetAvailableCostAllocationLots(late bool) ([]models.CostAllocationLotCandidate, error) {
	if late {
		return uc.repo.GetAllCostAllocationLotCandidates()
	}
	return uc.repo.GetAvailableCostAllocationLots()
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

	// Fetch all candidates to build a base map for already allocated lots
	allCandidates, err := uc.repo.GetAllCostAllocationLotCandidates()
	if err != nil {
		return nil, err
	}
	lotCandidateMap := make(map[int]models.CostAllocationLotCandidate)
	for _, c := range allCandidates {
		lotCandidateMap[c.LotID] = c
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

	// Find APs matching reference UUIDs and not fully allocated
	aps, err := uc.repo.GetAccountPayables()
	if err != nil {
		return nil, err
	}

	var matchedAPs []models.AccountPayable
	for _, ap := range aps {
		if refUUIDs[ap.ReferenceUUID] && ap.AllocationStatus != "Closed" {
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
		case "Unit":
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
		case "Unit":
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

		// Find one selected candidate that matches this AP reference to get the reference ID
		refID := 0
		for _, c := range selectedCandidates {
			if matchesReference(c, ap.ReferenceUUID, ap.ReferenceType) {
				refID = getRefID(c, ap.ReferenceType)
				break
			}
		}
		if refID == 0 {
			continue
		}

		// Compute total base for the referenced entity including unlanded items
		baseUnits, err := uc.repo.GetCostAllocationReferenceBaseUnits(ap.ReferenceType, refID)
		if err != nil {
			continue
		}
		var totalRefBase float64
		for _, u := range baseUnits {
			totalRefBase += computeBaseFromUnit(u, ap.AllocationType)
		}

		// Compute already allocated amount and base for this AP
		alreadyAllocated, allocatedLotIDs, err := uc.repo.GetCostAllocationSummaryByAP(ap.ID)
		if err != nil {
			continue
		}
		var allocatedBase float64
		for _, lotID := range allocatedLotIDs {
			if c, ok := lotCandidateMap[lotID]; ok && matchesReference(c, ap.ReferenceUUID, ap.ReferenceType) {
				allocatedBase += computeBase(c, ap.AllocationType)
			}
		}

		// Remaining AP amount and remaining reference base after excluding already allocated portions
		remainingAP := ap.LocalAmount - alreadyAllocated
		remainingBase := totalRefBase - allocatedBase

		if remainingAP <= 0 || remainingBase <= 0 {
			continue
		}

		// Distribute remaining AP amount to selected lots proportionally based on remaining base
		for _, c := range selectedCandidates {
			if !matchesReference(c, ap.ReferenceUUID, ap.ReferenceType) {
				continue
			}
			base := computeBase(c, ap.AllocationType)
			if base == 0 {
				continue
			}
			amount := remainingAP * base / remainingBase
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
	return uc.repo.SaveContainerItem(item)
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

// DeleteRecord 레코드 삭제
func (uc *ProcurementUseCase) DeleteRecord(tableName string, id int) error {
	return uc.repo.DeleteRecord(tableName, id)
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
	if err := uc.repo.SaveAPTargetGroup(g); err != nil {
		return err
	}
	// Save items if present
	if g.Items != nil {
		for i := range g.Items {
			g.Items[i].GroupID = g.ID
			if err := uc.repo.SaveAPTargetGroupItem(&g.Items[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (uc *ProcurementUseCase) SaveAPTargetGroupItem(item *models.AP_TargetGroupItem) error {
	return uc.repo.SaveAPTargetGroupItem(item)
}

func (uc *ProcurementUseCase) DeleteAPTargetGroupItem(id int) error {
	return uc.repo.DeleteAPTargetGroupItem(id)
}

func (uc *ProcurementUseCase) GetAllReferenceTargets() ([]models.APTargetGroupReference, error) {
	return uc.repo.GetAllReferenceTargets()
}
