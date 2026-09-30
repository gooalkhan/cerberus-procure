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
