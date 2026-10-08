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
	SavePOItem(item *models.POItem) error
	DeletePOItem(id int) error

	// Commercial Invoice
	GetCommercialInvoices() ([]models.CommercialInvoice, error)
	GetCIAggregatedItems(ciID int) ([]models.CIAggregatedItem, error)
	SaveCommercialInvoice(ci *models.CommercialInvoice) error

	// Account Payable
	GetAccountPayables() ([]models.AccountPayable, error)
	SaveAccountPayable(ap *models.AccountPayable) error

	// Container & Logistics
	GetContainers() ([]models.Container, error)
	SaveContainer(c *models.Container) error
	GetBLs() ([]models.BL, error)
	SaveBL(bl *models.BL) error

	// Goods Receipt & Inventory
	GetGoodsReceipts() ([]models.GoodsReceipt, error)
	SaveGoodsReceipt(gr *models.GoodsReceipt) error
	GetInventoryLots() ([]models.InventoryLot, error)
	GetInventoryLotsByGRID(grID int) ([]models.InventoryLot, error)
	SaveInventoryLot(lot *models.InventoryLot) error

	// Container Items
	GetContainerItemsByContainerID(containerID int) ([]models.ContainerItem, error)
	GetContainersByBLID(blID int) ([]models.Container, error)
	SaveContainerItem(item *models.ContainerItem) error

	// Cost Allocation
	GetCostAllocations() ([]models.CostAllocation, error)
	SaveCostAllocation(ca *models.CostAllocation) error

	// Cost Allocation Items
	GetCostAllocationItemsByAllocationID(caID int) ([]models.CostAllocationItem, error)
	SaveCostAllocationItem(item *models.CostAllocationItem) error

	// Cost Allocation Helpers
	GetAvailableCostAllocationLots() ([]models.CostAllocationLotCandidate, error)
	GetCostAllocationLotCandidatesByIDs(ids []int) ([]models.CostAllocationLotCandidate, error)
	GetAllCostAllocationLotCandidates() ([]models.CostAllocationLotCandidate, error)
	GetCostAllocationReferenceBaseUnits(refType string, refID int) ([]models.CostAllocationBaseUnit, error)
	GetAPIDsUsedForLots(lotIDs []int) (map[int]bool, error)
	GetCostAllocationSummaryByAP(apID int) (totalAmount float64, lotIDs []int, err error)
	GetCostAllocationWarningByPO(poID int) (*models.POAllocationWarning, error)

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
}
