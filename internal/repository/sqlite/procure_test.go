package sqlite_test

import (
	"cerberus-procure/internal/logic"
	"cerberus-procure/internal/models"
	"cerberus-procure/internal/repository/sqlite"
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSQLite_SavePurchaseOrder_Reconciliation(t *testing.T) {
	dbPath := "test_procure.db"
	defer os.Remove(dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	repo, err := sqlite.NewSQLiteProcurementRepository(db)
	if err != nil {
		t.Fatalf("Failed to create sqlite repo: %v", err)
	}

	uc := logic.NewProcurementUseCase(repo)

	po := &models.PurchaseOrder{
		PONo:     "PO-SQLITE-001",
		VendorID: 1,
		Status:   "Open",
		Items: []models.POItem{
			{ItemID: 1, POQty: 10, UnitPrice: 100, Status: "Not Shipped"},
			{ItemID: 2, POQty: 20, UnitPrice: 50, Status: "Not Shipped"},
		},
	}

	if err := uc.SavePurchaseOrder(po); err != nil {
		t.Fatalf("Failed to save PO: %v", err)
	}

	items, err := uc.GetPOItemsByPOID(po.ID)
	if err != nil {
		t.Fatalf("Failed to get PO items: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	// Remove second item
	po.Items = []models.POItem{items[0]}
	if err := uc.SavePurchaseOrder(po); err != nil {
		t.Fatalf("Failed to reconcile PO: %v", err)
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
