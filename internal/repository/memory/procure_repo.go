package memory

import (
	"cerberus-procure/internal/models"
	"sort"
	"sync"
	"time"
)

type MemoryProcurementRepository struct {
	mu      sync.RWMutex
	items   map[int]models.ItemMaster
	vendors map[int]models.VendorMaster
	pos     map[int]models.PurchaseOrder
	poItems map[int][]models.POItem
	cis     map[int]models.CommercialInvoice
	aps     map[int]models.AccountPayable
	containers map[int]models.Container
	bls     map[int]models.BL
	grs     map[int]models.GoodsReceipt
	lots    map[int]models.InventoryLot
	cas     map[int]models.CostAllocation
	containerItems map[int][]models.ContainerItem
	caItems map[int][]models.CostAllocationItem
	nextID int
}

func NewMemoryProcurementRepository() *MemoryProcurementRepository {
	r := &MemoryProcurementRepository{
		items:      make(map[int]models.ItemMaster),
		vendors:    make(map[int]models.VendorMaster),
		pos:        make(map[int]models.PurchaseOrder),
		poItems:    make(map[int][]models.POItem),
		cis:        make(map[int]models.CommercialInvoice),
		aps:        make(map[int]models.AccountPayable),
		containers: make(map[int]models.Container),
		bls:        make(map[int]models.BL),
		grs:        make(map[int]models.GoodsReceipt),
		lots:       make(map[int]models.InventoryLot),
		cas:        make(map[int]models.CostAllocation),
		containerItems: make(map[int][]models.ContainerItem),
		caItems:    make(map[int][]models.CostAllocationItem),
		nextID:     10000,
	}
	r.seed()
	return r
}

func (r *MemoryProcurementRepository) seed() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Reset all maps
	r.items = make(map[int]models.ItemMaster)
	r.vendors = make(map[int]models.VendorMaster)
	r.pos = make(map[int]models.PurchaseOrder)
	r.poItems = make(map[int][]models.POItem)
	r.cis = make(map[int]models.CommercialInvoice)
	r.aps = make(map[int]models.AccountPayable)
	r.containers = make(map[int]models.Container)
	r.bls = make(map[int]models.BL)
	r.grs = make(map[int]models.GoodsReceipt)
	r.lots = make(map[int]models.InventoryLot)
	r.cas = make(map[int]models.CostAllocation)
	r.containerItems = make(map[int][]models.ContainerItem)
	r.caItems = make(map[int][]models.CostAllocationItem)

	// Mock Items
	item1 := models.ItemMaster{ID: 1, SKUCode: "ITEM-001", Name: "Premium Coffee Beans", VendorID: 1, CBM: 0.05, NetWeight: 10, GrossWeight: 10.5, CreatedAt: time.Now()}
	item2 := models.ItemMaster{ID: 2, SKUCode: "ITEM-002", Name: "Organic Tea Leaves", VendorID: 1, CBM: 0.03, NetWeight: 5, GrossWeight: 5.2, CreatedAt: time.Now()}
	r.items[item1.ID] = item1
	r.items[item2.ID] = item2

	// Mock Vendors
	v1 := models.VendorMaster{ID: 1, Name: "Global Trading Co.", Category: "Supplier", CreatedAt: time.Now()}
	v2 := models.VendorMaster{ID: 2, Name: "Fast Logistics", Category: "Forwarder", CreatedAt: time.Now()}
	r.vendors[v1.ID] = v1
	r.vendors[v2.ID] = v2

	// Mock POs
	po1 := models.PurchaseOrder{ID: 1, PONo: "PO-2024-001", VendorID: 1, Currency: "USD", TotalAmount: 3150.0, Status: "Open", PODate: time.Now().AddDate(0, 0, -5), UUID: "po-uuid-1"}
	po2 := models.PurchaseOrder{ID: 2, PONo: "PO-2024-002", VendorID: 2, Currency: "EUR", TotalAmount: 800.0, Status: "Open", PODate: time.Now(), UUID: "po-uuid-2"}
	r.pos[po1.ID] = po1
	r.pos[po2.ID] = po2

	// Mock PO Items
	r.poItems[po1.ID] = []models.POItem{
		{ID: 1, POID: 1, ItemID: 1, POQty: 100, UnitPrice: 15.5, Status: "Not Shipped"},
		{ID: 2, POID: 1, ItemID: 2, POQty: 200, UnitPrice: 8.0, Status: "Not Shipped"},
	}
	r.poItems[po2.ID] = []models.POItem{
		{ID: 3, POID: 2, ItemID: 1, POQty: 50, UnitPrice: 16.0, Status: "Not Shipped"},
	}

	// Mock Containers
	c1 := models.Container{ID: 1, ContainerNo: "MSCU1234567", Status: "Shipping", UUID: "cont-uuid-1"}
	c2 := models.Container{ID: 2, ContainerNo: "HMMU7654321", Status: "Loaded", UUID: "cont-uuid-2"}
	r.containers[c1.ID] = c1
	r.containers[c2.ID] = c2

	// Mock BLs
	b1 := models.BL{ID: 1, BLNo: "BL-999-001", ETD: time.Now().AddDate(0, 0, -10), ETA: time.Now().AddDate(0, 0, 5), VesselName: "Cerberus Star", Status: "Shipping", UUID: "bl-uuid-1"}
	r.bls[b1.ID] = b1

	// Mock CIs
	invoice1 := models.CommercialInvoice{ID: 1, CINo: "CI-INV-2024-001", InvoiceDate: time.Now().AddDate(0, 0, -7), VendorID: 1, Currency: "USD", TotalAmount: 1950.0, Status: "Open", UUID: "ci-uuid-1", CreatedAt: time.Now()}
	invoice2 := models.CommercialInvoice{ID: 2, CINo: "CI-INV-2024-002", InvoiceDate: time.Now().AddDate(0, 0, -2), VendorID: 1, Currency: "USD", TotalAmount: 400.0, Status: "Draft", UUID: "ci-uuid-2", CreatedAt: time.Now()}
	r.cis[invoice1.ID] = invoice1
	r.cis[invoice2.ID] = invoice2

	// Mock Container Items
	ci1 := models.ContainerItem{
		ID: 1, ContainerID: 1, BLID: 1, CIID: 1, POItemID: 1, ItemID: 1, 
		LoadQty: 100, UnitPrice: 15.5, Currency: "USD",
		GrossWeight: 1050, NetWeight: 1000, CBM: 5.0,
	}
	ci2 := models.ContainerItem{
		ID: 2, ContainerID: 2, BLID: 1, CIID: 1, POItemID: 2, ItemID: 2, 
		LoadQty: 50, UnitPrice: 8.0, Currency: "USD",
		GrossWeight: 260, NetWeight: 250, CBM: 1.5,
	}
	r.containerItems[c1.ID] = append(r.containerItems[c1.ID], ci1)
	r.containerItems[c2.ID] = append(r.containerItems[c2.ID], ci2)

	// Mock APs
	ap1 := models.AccountPayable{
		ID: 1, APNo: "AP-2024-001", VendorID: 1, Amount: 1950.0, Currency: "USD", 
		DueDate: time.Now().AddDate(0, 1, 0), DateOfPayment: time.Time{}, 
		Status: "unpaid", AllocationStatus: "Open", UUID: "ap-uuid-1", CreatedAt: time.Now(),
	}
	r.aps[ap1.ID] = ap1

	// Mock Goods Receipts
	gr1 := models.GoodsReceipt{ID: 1, ContainerID: 1, BLID: 1, ReceiveDate: time.Now().AddDate(0, 0, -1), Remark: "First Shipment Arrived", UUID: "gr-uuid-1", CreatedAt: time.Now()}
	r.grs[gr1.ID] = gr1

	// Mock Inventory Lots
	lot1 := models.InventoryLot{ID: 1, GRID: 1, ContainerItemID: 1, LotNo: "LOT-COFFEE-001", Qty: 100, LandedCostPerUnit: 16.2, QuarantineStatus: "Passed", UUID: "lot-uuid-1", CreatedAt: time.Now()}
	r.lots[lot1.ID] = lot1

	// Update aggregations
	r.updateContainerAggregation(c1.ID)
	r.updateContainerAggregation(c2.ID)
}

func (r *MemoryProcurementRepository) SeedData() error {
	r.seed()
	return nil
}

// Item Master
func (r *MemoryProcurementRepository) GetItems() ([]models.ItemMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]models.ItemMaster, 0, len(r.items))
	for _, it := range r.items {
		items = append(items, it)
	}
	return items, nil
}

func (r *MemoryProcurementRepository) GetItemByID(id int) (*models.ItemMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if it, ok := r.items[id]; ok {
		return &it, nil
	}
	return nil, nil
}

func (r *MemoryProcurementRepository) SaveItem(it *models.ItemMaster) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if it.ID == 0 {
		it.ID = len(r.items) + 1
		it.CreatedAt = time.Now()
	}
	it.UpdatedAt = time.Now()
	r.items[it.ID] = *it
	return nil
}

// Vendor Master
func (r *MemoryProcurementRepository) GetVendors() ([]models.VendorMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	vendors := make([]models.VendorMaster, 0, len(r.vendors))
	for _, v := range r.vendors {
		vendors = append(vendors, v)
	}
	return vendors, nil
}

func (r *MemoryProcurementRepository) GetVendorByID(id int) (*models.VendorMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if v, ok := r.vendors[id]; ok {
		return &v, nil
	}
	return nil, nil
}

func (r *MemoryProcurementRepository) SaveVendor(v *models.VendorMaster) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v.ID == 0 {
		v.ID = len(r.vendors) + 1
		v.CreatedAt = time.Now()
	}
	v.UpdatedAt = time.Now()
	r.vendors[v.ID] = *v
	return nil
}

// Purchase Order
func (r *MemoryProcurementRepository) GetPurchaseOrders() ([]models.PurchaseOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pos := make([]models.PurchaseOrder, 0, len(r.pos))
	for _, p := range r.pos {
		pos = append(pos, p)
	}
	return pos, nil
}

func (r *MemoryProcurementRepository) GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.pos[id]; ok {
		return &p, nil
	}
	return nil, nil
}

func (r *MemoryProcurementRepository) SavePurchaseOrder(p *models.PurchaseOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p.ID == 0 {
		p.ID = len(r.pos) + 1
		p.CreatedAt = time.Now()
	}
	p.UpdatedAt = time.Now()
	r.pos[p.ID] = *p
	return nil
}

// PO Item
func (r *MemoryProcurementRepository) GetPOItemsByPOID(poID int) ([]models.POItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.poItems[poID]
	if list == nil {
		return []models.POItem{}, nil
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SavePOItem(i *models.POItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.ID == 0 {
		i.ID = 1000 + len(r.poItems[i.POID]) + 1
		i.CreatedAt = time.Now()
		r.poItems[i.POID] = append(r.poItems[i.POID], *i)
	} else {
		for idx, item := range r.poItems[i.POID] {
			if item.ID == i.ID {
				i.UpdatedAt = time.Now()
				r.poItems[i.POID][idx] = *i
				break
			}
		}
	}

	// Trigger logic: Close PO if all items are terminal
	allTerminal := true
	for _, item := range r.poItems[i.POID] {
		if item.Status != "Shipped" && item.Status != "Cancelled" {
			allTerminal = false
			break
		}
	}
	if allTerminal && len(r.poItems[i.POID]) > 0 {
		po := r.pos[i.POID]
		po.Status = "Closed"
		r.pos[i.POID] = po
	} else if !allTerminal {
		po := r.pos[i.POID]
		po.Status = "Open"
		r.pos[i.POID] = po
	}

	return nil
}

// Commercial Invoice
func (r *MemoryProcurementRepository) GetCommercialInvoices() ([]models.CommercialInvoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.CommercialInvoice, 0, len(r.cis))
	for _, i := range r.cis {
		list = append(list, i)
	}
	return list, nil
}

func (r *MemoryProcurementRepository) GetCIAggregatedItems(ciID int) ([]models.CIAggregatedItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	aggregates := make(map[int]*models.CIAggregatedItem)
	for _, items := range r.containerItems {
		for _, ci := range items {
			if ci.CIID == ciID {
				itemName := ""
				if m, ok := r.items[ci.ItemID]; ok {
					itemName = m.Name
				}
				if agg, ok := aggregates[ci.ItemID]; ok {
					agg.TotalQty += ci.LoadQty
					agg.Amount += ci.LoadQty * ci.UnitPrice
				} else {
					aggregates[ci.ItemID] = &models.CIAggregatedItem{
						ItemID:   ci.ItemID,
						ItemName: itemName,
						TotalQty: ci.LoadQty,
						Amount:   ci.LoadQty * ci.UnitPrice,
						Currency: ci.Currency,
					}
				}
			}
		}
	}

	var list []models.CIAggregatedItem
	for _, agg := range aggregates {
		list = append(list, *agg)
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SaveCommercialInvoice(i *models.CommercialInvoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.ID == 0 {
		i.ID = len(r.cis) + 1
		i.CreatedAt = time.Now()
	}
	i.UpdatedAt = time.Now()
	r.cis[i.ID] = *i
	return nil
}

// Account Payable
func (r *MemoryProcurementRepository) GetAccountPayables() ([]models.AccountPayable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.AccountPayable, 0, len(r.aps))
	for _, i := range r.aps {
		list = append(list, i)
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SaveAccountPayable(i *models.AccountPayable) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.ID == 0 {
		i.ID = len(r.aps) + 1
		i.CreatedAt = time.Now()
	}
	i.UpdatedAt = time.Now()
	r.aps[i.ID] = *i
	return nil
}

// Container & Logistics
func (r *MemoryProcurementRepository) GetContainers() ([]models.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.Container, 0, len(r.containers))
	for _, i := range r.containers {
		list = append(list, i)
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SaveContainer(i *models.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.ID == 0 {
		i.ID = len(r.containers) + 1
	}
	r.containers[i.ID] = *i
	return nil
}

func (r *MemoryProcurementRepository) GetBLs() ([]models.BL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.BL, 0, len(r.bls))
	for _, i := range r.bls {
		list = append(list, i)
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SaveBL(i *models.BL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.ID == 0 {
		i.ID = len(r.bls) + 1
	}
	r.bls[i.ID] = *i
	return nil
}

// Goods Receipt & Inventory
func (r *MemoryProcurementRepository) GetGoodsReceipts() ([]models.GoodsReceipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.GoodsReceipt, 0, len(r.grs))
	for _, i := range r.grs {
		list = append(list, i)
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SaveGoodsReceipt(i *models.GoodsReceipt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.ID == 0 {
		i.ID = len(r.grs) + 1
	}
	r.grs[i.ID] = *i
	return nil
}

func (r *MemoryProcurementRepository) GetInventoryLots() ([]models.InventoryLot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.InventoryLot, 0, len(r.lots))
	for _, i := range r.lots {
		list = append(list, i)
	}
	return list, nil
}

func (r *MemoryProcurementRepository) GetInventoryLotsByGRID(grID int) ([]models.InventoryLot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := []models.InventoryLot{}
	for _, lot := range r.lots {
		if lot.GRID == grID {
			list = append(list, lot)
		}
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SaveInventoryLot(i *models.InventoryLot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.ID == 0 {
		i.ID = len(r.lots) + 1
	}
	r.lots[i.ID] = *i
	return nil
}

// Cost Allocation
func (r *MemoryProcurementRepository) GetCostAllocations() ([]models.CostAllocation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.CostAllocation, 0, len(r.cas))
	for _, i := range r.cas {
		list = append(list, i)
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SaveCostAllocation(ca *models.CostAllocation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ca.ID == 0 {
		ca.ID = len(r.cas) + 1
		ca.CreatedAt = time.Now()
	}
	ca.UpdatedAt = time.Now()
	r.cas[ca.ID] = *ca
	return nil
}

// Container Items
func (r *MemoryProcurementRepository) GetContainerItemsByContainerID(containerID int) ([]models.ContainerItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.containerItems[containerID]
	if list == nil {
		return []models.ContainerItem{}, nil
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SaveContainerItem(i *models.ContainerItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Fetch Item details for calculation
	var masterItem *models.ItemMaster
	for _, items := range r.poItems {
		for _, pi := range items {
			if pi.ID == i.POItemID {
				i.ItemID = pi.ItemID
				if m, ok := r.items[pi.ItemID]; ok {
					masterItem = &m
				}
				break
			}
		}
		if masterItem != nil {
			break
		}
	}

	if masterItem != nil {
		i.GrossWeight = masterItem.GrossWeight * i.LoadQty
		i.NetWeight = masterItem.NetWeight * i.LoadQty
		i.CBM = masterItem.CBM * i.LoadQty
	}

	if i.ID == 0 {
		i.ID = 2000 + (r.nextID + 1)
		r.nextID++
		r.containerItems[i.ContainerID] = append(r.containerItems[i.ContainerID], *i)
		r.updateContainerAggregation(i.ContainerID)
	} else {
		// Find and remove from old container (if exists)
		oldContainerID := -1
		for cid, items := range r.containerItems {
			for idx, item := range items {
				if item.ID == i.ID {
					oldContainerID = cid
					// Remove from old
					r.containerItems[cid] = append(items[:idx], items[idx+1:]...)
					break
				}
			}
			if oldContainerID != -1 {
				break
			}
		}
		// Add to new container
		r.containerItems[i.ContainerID] = append(r.containerItems[i.ContainerID], *i)
		
		// Update aggregations for both
		r.updateContainerAggregation(i.ContainerID)
		if oldContainerID != -1 && oldContainerID != i.ContainerID {
			r.updateContainerAggregation(oldContainerID)
		}
	}
	return nil
}

func (r *MemoryProcurementRepository) updateContainerAggregation(containerID int) {
	items := r.containerItems[containerID]
	var totalCBM, totalNet, totalGross float64
	for _, item := range items {
		totalCBM += item.CBM
		totalNet += item.NetWeight
		totalGross += item.GrossWeight
	}
	if c, ok := r.containers[containerID]; ok {
		c.TotalCBM = totalCBM
		c.TotalNetWgt = totalNet
		c.TotalGrossWgt = totalGross
		r.containers[containerID] = c
	}
}

// Cost Allocation Items
func (r *MemoryProcurementRepository) GetCostAllocationItemsByAllocationID(caID int) ([]models.CostAllocationItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.caItems[caID]
	if list == nil {
		return []models.CostAllocationItem{}, nil
	}
	return list, nil
}

func (r *MemoryProcurementRepository) SaveCostAllocationItem(i *models.CostAllocationItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.ID == 0 {
		i.ID = 3000 + len(r.caItems[i.CostAllocationID]) + 1
	}
	r.caItems[i.CostAllocationID] = append(r.caItems[i.CostAllocationID], *i)
	return nil
}

func (r *MemoryProcurementRepository) GetContainersByBLID(blID int) ([]models.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	containerIDs := make(map[int]bool)
	for _, items := range r.containerItems {
		for _, ci := range items {
			if ci.BLID == blID {
				containerIDs[ci.ContainerID] = true
			}
		}
	}

	list := []models.Container{}
	for cID := range containerIDs {
		if c, ok := r.containers[cID]; ok {
			list = append(list, c)
		}
	}
	return list, nil
}

func (r *MemoryProcurementRepository) GetBookings() ([]models.BookingView, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := []models.BookingView{}
	for cID, items := range r.containerItems {
		container := r.containers[cID]
		for _, ci := range items {
			bl := r.bls[ci.BLID]
			itemName := ""
			if m, ok := r.items[ci.ItemID]; ok {
				itemName = m.Name
			}
			list = append(list, models.BookingView{
				ContainerItemID: ci.ID,
				ContainerID:     ci.ContainerID,
				ContainerNo:     container.ContainerNo,
				Status:          container.Status,
				TotalCBM:        container.TotalCBM,
				TotalNetWgt:     container.TotalNetWgt,
				TotalGrossWgt:   container.TotalGrossWgt,
				BLID:            ci.BLID,
				BLNo:            bl.BLNo,
				BLStatus:        bl.Status,
				ETD:             bl.ETD,
				ETA:             bl.ETA,
				POL:             bl.POL,
				POD:             bl.POD,
				Carrier:         bl.Carrier,
				VesselName:      bl.VesselName,
				POItemID:        ci.POItemID,
				ItemID:          ci.ItemID,
				ItemName:        itemName,
				CIID:            ci.CIID,
				LoadQty:         ci.LoadQty,
				UnitPrice:       ci.UnitPrice,
				Currency:        ci.Currency,
				GrossWeight:     ci.GrossWeight,
				NetWeight:       ci.NetWeight,
				CBM:             ci.CBM,
				Remark:          ci.Remark,
			})
		}
	}

	// Sort: Status (Loaded, Shipping, Arrived), ETA, ContainerNo
	statusOrder := map[string]int{"Loaded": 1, "Shipping": 2, "Arrived": 3}
	sort.Slice(list, func(i, j int) bool {
		si := statusOrder[list[i].Status]
		sj := statusOrder[list[j].Status]
		if si != sj {
			return si < sj
		}
		if !list[i].ETA.Equal(list[j].ETA) {
			return list[i].ETA.Before(list[j].ETA)
		}
		return list[i].ContainerNo < list[j].ContainerNo
	})

	return list, nil
}
