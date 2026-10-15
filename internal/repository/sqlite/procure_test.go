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

func TestSQLite_SaveContainerItem_POItemStatusSync(t *testing.T) {
	dbPath := "test_procure_ci.db"
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

	// 1. Create PO with 1 item of Qty 10
	po := &models.PurchaseOrder{
		PONo:     "PO-CI-TEST-001",
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
		ContainerNo: "CONT-TEST-001",
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

func TestSQLite_POItem_OverShipped(t *testing.T) {
	dbPath := "test_procure_over.db"
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
		PONo:     "PO-OVER-001",
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
	target := poItems[0]

	container := &models.Container{ContainerNo: "CONT-OVER-001", Status: "Loaded"}
	if err := uc.SaveContainer(container); err != nil {
		t.Fatalf("Failed to save Container: %v", err)
	}

	// 과다 선적: LoadQty(12) > POQty(10)
	ci := &models.ContainerItem{ContainerID: container.ID, POItemID: target.ID, LoadQty: 12}
	if err := uc.SaveContainerItem(ci); err != nil {
		t.Fatalf("Failed to save ContainerItem: %v", err)
	}

	poItems, _ = uc.GetPOItemsByPOID(po.ID)
	if poItems[0].Status != "Over Shipped" {
		t.Fatalf("Expected PO Item status 'Over Shipped', got '%s'", poItems[0].Status)
	}
	// 과다 선적도 완납으로 간주하여 PO는 Closed
	reloadedPO, _ := uc.GetPurchaseOrderByID(po.ID)
	if reloadedPO.Status != "Closed" {
		t.Fatalf("Expected PO status 'Closed' on over shipment, got '%s'", reloadedPO.Status)
	}

	// 취소(음수 증분)로 PO_Qty를 0으로: 이미 선적된 버킷은 여전히 Over Shipped
	adj := &models.POItemAdjustment{POItemID: target.ID, QtyDelta: -10, Reason: "전량 취소"}
	if err := uc.AdjustPOItemQty(adj); err != nil {
		t.Fatalf("Failed to adjust qty: %v", err)
	}
	poItems, _ = uc.GetPOItemsByPOID(po.ID)
	if poItems[0].POQty != 0 {
		t.Fatalf("Expected PO_Qty 0, got %v", poItems[0].POQty)
	}
	if poItems[0].Status != "Over Shipped" {
		t.Fatalf("Expected 'Over Shipped' after cancel with existing shipment, got '%s'", poItems[0].Status)
	}
}

func TestSQLite_ContainerTotalsRecompute(t *testing.T) {
	dbPath := "test_procure_totals.db"
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

	if err := repo.SaveItem(&models.ItemMaster{SKUCode: "SKU-TOT", Name: "Tot", VendorID: 1, CBM: 2, NetWeight: 10, GrossWeight: 11, BoxQty: 1, BoxesPerPallet: 96}); err != nil {
		t.Fatalf("Failed to save item: %v", err)
	}
	po := &models.PurchaseOrder{
		PONo:     "PO-TOT-001",
		VendorID: 1,
		Status:   "Open",
		Items:    []models.POItem{{ItemID: 1, POQty: 5, UnitPrice: 1, Status: "Not Shipped"}},
	}
	if err := uc.SavePurchaseOrder(po); err != nil {
		t.Fatalf("Failed to save PO: %v", err)
	}
	poItems, _ := uc.GetPOItemsByPOID(po.ID)
	target := poItems[0]

	container := &models.Container{ContainerNo: "C-TOT-001", Status: "Loaded"}
	if err := uc.SaveContainer(container); err != nil {
		t.Fatalf("Failed to save container: %v", err)
	}

	ci := &models.ContainerItem{ContainerID: container.ID, POItemID: target.ID, LoadQty: 3}
	if err := uc.SaveContainerItem(ci); err != nil {
		t.Fatalf("Failed to save container item: %v", err)
	}

	// 파생값: 중량/부피 = Item_Master × Load_Qty (SQL 트리거 없이 앱 레이어 계산)
	if ci.GrossWeight != 33 || ci.NetWeight != 30 || ci.CBM != 6 {
		t.Fatalf("derived item expected 33/30/6, got %v/%v/%v", ci.GrossWeight, ci.NetWeight, ci.CBM)
	}
	c, _ := repo.GetContainerByID(container.ID)
	if c.TotalCBM != 6 || c.TotalNetWgt != 30 || c.TotalGrossWgt != 33 {
		t.Fatalf("container totals expected 6/30/33, got %v/%v/%v", c.TotalCBM, c.TotalNetWgt, c.TotalGrossWgt)
	}
	// 상태: load 3 < qty 5 → Partially Shipped
	re, _ := uc.GetPOItemsByPOID(po.ID)
	if re[0].Status != "Partially Shipped" {
		t.Fatalf("expected Partially Shipped, got %q", re[0].Status)
	}

	// 삭제 시 집계/상태 재계산
	cis, _ := repo.GetContainerItemsByContainerID(container.ID)
	if len(cis) != 1 {
		t.Fatalf("expected 1 container item, got %d", len(cis))
	}
	if err := uc.DeleteRecord("Container_Item", cis[0].ID); err != nil {
		t.Fatalf("Failed to delete container item: %v", err)
	}
	c2, _ := repo.GetContainerByID(container.ID)
	if c2.TotalCBM != 0 || c2.TotalNetWgt != 0 || c2.TotalGrossWgt != 0 {
		t.Fatalf("after delete expected 0/0/0, got %v/%v/%v", c2.TotalCBM, c2.TotalNetWgt, c2.TotalGrossWgt)
	}
	re2, _ := uc.GetPOItemsByPOID(po.ID)
	if re2[0].Status != "Not Shipped" {
		t.Fatalf("after delete expected Not Shipped, got %q", re2[0].Status)
	}
}

func TestSQLite_POReferenceStatus_ViaAPTargetGroup(t *testing.T) {
	dbPath := "test_procure_group_lock.db"
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
		PONo:     "PO-GRP-001",
		VendorID: 1,
		Status:   "Open",
		Items:    []models.POItem{{ItemID: 1, POQty: 10, UnitPrice: 5, Status: "Not Shipped"}},
	}
	if err := uc.SavePurchaseOrder(po); err != nil {
		t.Fatalf("Failed to save PO: %v", err)
	}

	// 참조 전: 미잠금
	if hasAP, hasCI, _ := uc.GetPOReferenceStatus(po.ID); hasAP || hasCI {
		t.Fatalf("expected unlocked before refs, got hasAP=%v hasCI=%v", hasAP, hasCI)
	}

	// AP Target Group 생성 (PO를 멤버로 포함)
	g := &models.AP_TargetGroup{
		GroupNo:       "GRP-LOCK-001",
		GroupName:     "lock-group",
		ReferenceType: "PO",
		Status:        "Open",
		UUID:          "grp-uuid-lock",
		Items:         []models.AP_TargetGroupItem{{ReferenceUUID: po.UUID}},
	}
	if err := uc.SaveAPTargetGroup(g); err != nil {
		t.Fatalf("Failed to save group: %v", err)
	}

	// 그룹만으로는 아직 AP 참조가 아니므로 미잠금
	if hasAP, _, _ := uc.GetPOReferenceStatus(po.ID); hasAP {
		t.Fatalf("group without referencing AP should not lock the PO")
	}

	// AP가 그룹을 참조 → 간접 AP 참조 → 잠금
	ap := &models.AccountPayable{
		APNo: "AP-GRP-001", VendorID: 1, Amount: 100, Currency: "USD", LocalAmount: 100,
		AllocationType: "Value", ReferenceType: "AP_Target_Group", ReferenceUUID: "grp-uuid-lock",
		Status: "unpaid", AllocationStatus: "Open", UUID: "ap-uuid-lock",
	}
	if err := uc.SaveAccountPayable(ap); err != nil {
		t.Fatalf("Failed to save AP: %v", err)
	}

	hasAP, hasCI, err := uc.GetPOReferenceStatus(po.ID)
	if err != nil {
		t.Fatalf("Failed to get ref status: %v", err)
	}
	if !hasAP {
		t.Fatalf("expected hasAP=true via AP Target Group indirect reference")
	}
	if hasCI {
		t.Fatalf("expected hasCI=false")
	}

	// 잠금 후 기존 품목 삭제 시도 → 차단
	po.Items = []models.POItem{}
	if err := uc.SavePurchaseOrder(po); err == nil {
		t.Fatalf("expected item deletion to be blocked after group-based AP reference")
	}
}
