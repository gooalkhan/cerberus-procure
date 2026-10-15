package repository

import  "cerberus-procure/internal/models"

type TodoRepository interface {
	GetTodos() ([]models.Todo, error)
	AddTodo(title string) (models.Todo, error)
	ToggleTodo(id string) error
	DeleteTodo(id string) error
}

type UserRepository interface {
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	CreateUser(user *models.User) error
}

type ProcurementRepository interface {
	// System
	SeedData() error

	// Item Master
	GetItems() ([]models.ItemMaster, error)
	GetItemByID(id int) (*models.ItemMaster, error)
	SaveItem(item *models.ItemMaster) error

	// Vendor Master
	GetVendors() ([]models.VendorMaster, error)
	GetVendorByID(id int) (*models.VendorMaster, error)
	SaveVendor(vendor *models.VendorMaster) error

	// Purchase Order
	GetPurchaseOrders() ([]models.PurchaseOrder, error)
	GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error)
	SavePurchaseOrder(po *models.PurchaseOrder) error

	// PO Item
	GetPOItemsByPOID(poID int) ([]models.POItem, error)
	GetPOItemByID(id int) (*models.POItem, error)
	GetPOItemByBucket(poID int, itemID int, unitPrice float64) (*models.POItem, error)
	GetPOLoadQty(poItemID int) (float64, error)
	SavePOItem(item *models.POItem) error
	DeletePOItem(id int) error
	SavePOItemAdjustment(a *models.POItemAdjustment) error
	GetPOItemAdjustmentsByPOItemID(poItemID int) ([]models.POItemAdjustment, error)
	GetPOItemAdjustmentsByPOID(poID int) ([]models.POItemAdjustment, error)
	GetPOReferenceStatus(poID int) (hasAP bool, hasContainerItem bool, err error)

	// Commercial Invoice
	GetCommercialInvoices() ([]models.CommercialInvoice, error)
	GetCIAggregatedItems(ciID int) ([]models.CIAggregatedItem, error)
	SaveCommercialInvoice(ci *models.CommercialInvoice) error

	// Account Payable
	GetAccountPayables() ([]models.AccountPayable, error)
	GetAccountPayableByID(id int) (*models.AccountPayable, error)
	GetAPAllocatedAmount(apID int) (hasAny bool, total float64, err error)
	SaveAccountPayable(ap *models.AccountPayable) error

	// Container & Logistics
	GetContainers() ([]models.Container, error)
	GetContainerByID(id int) (*models.Container, error)
	SaveContainer(c *models.Container) error
	GetBLs() ([]models.BL, error)
	SaveBL(bl *models.BL) error

	// Goods Receipt & Inventory
	GetGoodsReceipts() ([]models.GoodsReceipt, error)
	GetGoodsReceiptByID(id int) (*models.GoodsReceipt, error)
	SaveGoodsReceipt(gr *models.GoodsReceipt) error
	GetInventoryLots() ([]models.InventoryLot, error)
	GetInventoryLotByID(id int) (*models.InventoryLot, error)
	GetInventoryLotsByGRID(grID int) ([]models.InventoryLot, error)
	SaveInventoryLot(lot *models.InventoryLot) error

	// Container Items
	GetContainerItemsByContainerID(containerID int) ([]models.ContainerItem, error)
	GetContainerItemByID(id int) (*models.ContainerItem, error)
	GetContainersByBLID(blID int) ([]models.Container, error)
	SaveContainerItem(item *models.ContainerItem) error

	// Cost Allocation
	GetCostAllocations() ([]models.CostAllocation, error)
	SaveCostAllocation(ca *models.CostAllocation) error

	// Cost Allocation Items
	GetCostAllocationItemsByAllocationID(caID int) ([]models.CostAllocationItem, error)
	SaveCostAllocationItem(item *models.CostAllocationItem) error
	DeleteCostAllocationItem(id int) error

	// Cost Allocation Helpers
	GetAvailableCostAllocationLots() ([]models.CostAllocationLotCandidate, error)
	GetCostAllocationLotCandidatesByIDs(ids []int) ([]models.CostAllocationLotCandidate, error)
	GetAllCostAllocationLotCandidates() ([]models.CostAllocationLotCandidate, error)
	GetCostAllocationReferenceBaseUnits(refType string, refID int) ([]models.CostAllocationBaseUnit, error)
	GetAPIDsUsedForLots(lotIDs []int) (map[int]bool, error)
	GetCostAllocationWarningByPO(poID int) (*models.POAllocationWarning, error)
	GetReferenceIDByUUID(refType string, uuid string) (int, error)
	GetReferenceNoByUUID(refType string, uuid string) (string, error)
	GetCostAllocationAmountsByAP(apID int) (map[int]float64, error)

	// Unified Views
	GetBookings() ([]models.BookingView, error)

	// Booking Template
	GetUnbookedPOItems() ([]models.BookingTemplateRow, error)

	// Delete & Reference Check
	CheckReferences(tableName string, id int) (*models.DeleteCheckResult, error)
	DeleteRecord(tableName string, id int) error

	// AP Target Group
	GetAPTargetGroups() ([]models.AP_TargetGroup, error)
	GetAPTargetGroupByID(id int) (*models.AP_TargetGroup, error)
	GetAPTargetGroupItems(groupID int) ([]models.AP_TargetGroupItem, error)
	SaveAPTargetGroup(group *models.AP_TargetGroup) error
	SaveAPTargetGroupItem(item *models.AP_TargetGroupItem) error
	DeleteAPTargetGroupItem(id int) error
	GetAllReferenceTargets() ([]models.APTargetGroupReference, error)

	// Constraint Profile (행정 제약 프로파일)
	GetConstraintProfiles() ([]models.ConstraintProfile, error)
	GetConstraintProfileByID(id int) (*models.ConstraintProfile, error)
	GetConstraintProfileSteps(profileID int) ([]models.ConstraintProfileStep, error)
	SaveConstraintProfile(p *models.ConstraintProfile) error
	SaveConstraintProfileStep(s *models.ConstraintProfileStep) error
	DeleteConstraintProfileStep(id int) error

	// Constraint Master (개별 제약 카탈로그)
	GetConstraints() ([]models.ConstraintMaster, error)
	GetConstraintMasterByID(id int) (*models.ConstraintMaster, error)
	SaveConstraintMaster(c *models.ConstraintMaster) error

	// Constraint Instance (Lot별 행정 제약 인스턴스)
	GetConstraintInstances(lotID int) ([]models.ConstraintInstance, error)
	SaveConstraintInstance(inst *models.ConstraintInstance) error
	DeleteConstraintInstance(id int) error
}
