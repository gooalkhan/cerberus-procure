package sqlite

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// TestRecomputePOItemStatuses 마이그레이션 재계산 SQL이 상관 서브쿼리(PO_Item.PO_Item_ID)를
// 올바로 해석하고 Over Shipped 를 포함한 4개 상태를 정확히 산출하는지 검증한다.
func TestRecomputePOItemStatuses(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	stmts := []string{
		`CREATE TABLE PO_Item (PO_Item_ID INTEGER PRIMARY KEY, PO_Qty REAL, Status TEXT)`,
		`CREATE TABLE Container_Item (PO_Item_ID INTEGER, Load_Qty REAL)`,
		`INSERT INTO PO_Item (PO_Item_ID, PO_Qty) VALUES (1, 10), (2, 5), (3, 0), (4, 7), (5, 0)`,
		`INSERT INTO Container_Item (PO_Item_ID, Load_Qty) VALUES (1, 12), (2, 5), (3, 3)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("setup %q: %v", s, err)
		}
	}

	if err := recomputePOItemStatuses(db); err != nil {
		t.Fatalf("recompute failed: %v", err)
	}

	want := map[int]string{
		1: "Over Shipped", // load 12 > qty 10
		2: "Shipped",      // load 5 == qty 5
		3: "Over Shipped", // qty 0, load 3
		4: "Not Shipped",  // qty 7, no load
		5: "Not Shipped",  // qty 0, no load
	}
	for id, w := range want {
		var got string
		if err := db.QueryRow("SELECT Status FROM PO_Item WHERE PO_Item_ID=?", id).Scan(&got); err != nil {
			t.Fatalf("query item %d: %v", id, err)
		}
		if got != w {
			t.Errorf("item %d: want %q, got %q", id, w, got)
		}
	}
}
