package logic

import (
	"cerberus-procure/internal/models"
	"cerberus-procure/internal/repository"
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
	// Save items if present
	for i := range po.Items {
		po.Items[i].POID = po.ID
		uc.repo.SavePOItem(&po.Items[i])
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

func (uc *ProcurementUseCase) SaveCostAllocation(ca *models.CostAllocation) error {
	return uc.repo.SaveCostAllocation(ca)
}

// Container Items
func (uc *ProcurementUseCase) GetContainerItemsByContainerID(containerID int) ([]models.ContainerItem, error) {
	return uc.repo.GetContainerItemsByContainerID(containerID)
}

func (uc *ProcurementUseCase) SaveContainerItem(item *models.ContainerItem) error {
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
