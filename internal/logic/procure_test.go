package logic_test

import (
	"cerberus-procure/internal/logic"
	"cerberus-procure/internal/models"
	"cerberus-procure/internal/repository/memory"
	"testing"
)

func TestSavePurchaseOrder_ItemReconciliation(t *testing.T) {
	repo := memory.NewMemoryProcurementRepository()
	uc := logic.NewProcurementUseCase(repo)

	// Create a PO with 2 items
	po := &models.PurchaseOrder{
		PONo:     "PO-TEST-001",
		VendorID: 1,
		Items: []models.POItem{
			{ItemID: 1, POQty: 10, UnitPrice: 100, Status: "Not Shipped"},
			{ItemID: 2, POQty: 20, UnitPrice: 50, Status: "Not Shipped"},
		},
	}

	err := uc.SavePurchaseOrder(po)
	if err != nil {
		t.Fatalf("Failed to save PO: %v", err)
	}

	items, err := uc.GetPOItemsByPOID(po.ID)
	if err != nil {
		t.Fatalf("Failed to get PO items: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	// Now remove the second item and save again (Reconciliation test)
	po.Items = []models.POItem{items[0]} // keep only the first item
	err = uc.SavePurchaseOrder(po)
	if err != nil {
		t.Fatalf("Failed to save PO with deleted item: %v", err)
	}

	updatedItems, err := uc.GetPOItemsByPOID(po.ID)
	if err != nil {
		t.Fatalf("Failed to get updated PO items: %v", err)
	}
	if len(updatedItems) != 1 {
		t.Fatalf("Expected 1 item after reconciliation, got %d", len(updatedItems))
	}
	if updatedItems[0].ID != items[0].ID {
		t.Fatalf("Expected item ID %d, got %d", items[0].ID, updatedItems[0].ID)
	}
}

func TestMemory_SaveContainerItem_POItemStatusSync(t *testing.T) {
	repo := memory.NewMemoryProcurementRepository()
	uc := logic.NewProcurementUseCase(repo)

	// 1. Create PO with 1 item of Qty 10
	po := &models.PurchaseOrder{
		PONo:     "PO-MEM-CI-001",
		VendorID: 1,
		Status:   "Open",
		Items: []models.POItem{
			{ItemID: 1, POQty: 10, UnitPrice: 100, Status: "Not Shipped"},
		},
	}
	if err := uc.SavePurchaseOrder(po); err != nil {
		t.Fatalf("Failed to save PO: %v", err)
	}
	poItems, _ := uc.GetPOItemsByPOID(po.ID)
	targetPOItem := poItems[0]

	// 2. Create Container
	container := &models.Container{
		ContainerNo: "CONT-MEM-001",
		Status:      "Loaded",
	}
	if err := uc.SaveContainer(container); err != nil {
		t.Fatalf("Failed to save Container: %v", err)
	}

	// 3. Partially load (LoadQty: 4 / 10)
	ci1 := &models.ContainerItem{
		ContainerID: container.ID,
		POItemID:    targetPOItem.ID,
		LoadQty:     4,
	}
	if err := uc.SaveContainerItem(ci1); err != nil {
		t.Fatalf("Failed to save ContainerItem 1: %v", err)
	}

	// Verify PO_Item status is 'Partially Shipped'
	poItems, _ = uc.GetPOItemsByPOID(po.ID)
	if poItems[0].Status != "Partially Shipped" {
		t.Fatalf("Expected PO Item status 'Partially Shipped', got '%s'", poItems[0].Status)
	}
	reloadedPO, _ := uc.GetPurchaseOrderByID(po.ID)
	if reloadedPO.Status != "Open" {
		t.Fatalf("Expected PO status 'Open', got '%s'", reloadedPO.Status)
	}

	// 4. Fully load with second item (LoadQty: 6 / 10, total 10)
	ci2 := &models.ContainerItem{
		ContainerID: container.ID,
		POItemID:    targetPOItem.ID,
		LoadQty:     6,
	}
	if err := uc.SaveContainerItem(ci2); err != nil {
		t.Fatalf("Failed to save ContainerItem 2: %v", err)
	}

	// Verify PO_Item status is 'Shipped' and PO is 'Closed'
	poItems, _ = uc.GetPOItemsByPOID(po.ID)
	if poItems[0].Status != "Shipped" {
		t.Fatalf("Expected PO Item status 'Shipped', got '%s'", poItems[0].Status)
	}
	reloadedPO, _ = uc.GetPurchaseOrderByID(po.ID)
	if reloadedPO.Status != "Closed" {
		t.Fatalf("Expected PO status 'Closed', got '%s'", reloadedPO.Status)
	}
}
