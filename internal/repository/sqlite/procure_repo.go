package sqlite

import (
	"cerberus-procure/internal/models"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func nullIfZero(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}

type SQLiteProcurementRepository struct {
	db *sql.DB
}

func NewSQLiteProcurementRepository(db *sql.DB) (*SQLiteProcurementRepository, error) {
	if err := migrateProcurement(db); err != nil {
		return nil, err
	}
	return &SQLiteProcurementRepository{db: db}, nil
}

func migrateProcurement(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS Item_Master (
			Item_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SKU_Code TEXT UNIQUE NOT NULL,
			Name TEXT NOT NULL,
			Vendor_ID INTEGER NOT NULL,
			CBM REAL CHECK (CBM >= 0),
			Box_Qty INTEGER NOT NULL DEFAULT 1,
			Boxes_Per_Pallet INTEGER,
			Box_Width REAL,
			Box_Length REAL,
			Box_Height REAL,
			Net_Weight REAL CHECK (Net_Weight >= 0),
			Gross_Weight REAL CHECK (Gross_Weight >= 0),
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (Vendor_ID) REFERENCES Vendor_Master(Vendor_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS Vendor_Master (
			Vendor_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			Category TEXT,
			Business_Reg_No TEXT,
			Bank_Account TEXT,
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS Purchase_Order (
			PO_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			PO_Date DATETIME NOT NULL,
			PO_No TEXT UNIQUE NOT NULL,
			Vendor_ID INTEGER NOT NULL,
			Currency TEXT NOT NULL,
			Total_Amount REAL NOT NULL DEFAULT 0,
			Status TEXT CHECK (Status IN ('Open', 'Closed')) DEFAULT 'Open',
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			UUID TEXT UNIQUE NOT NULL,
			FOREIGN KEY (Vendor_ID) REFERENCES Vendor_Master(Vendor_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS PO_Item (
			PO_Item_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			PO_ID INTEGER NOT NULL,
			Item_ID INTEGER NOT NULL,
			Base_Qty REAL,
			PO_Qty REAL CHECK (PO_Qty >= 0),
			Unit_Price REAL NOT NULL,
			Status TEXT CHECK (Status IN ('Shipped', 'Partially Shipped', 'Over Shipped', 'Not Shipped')) DEFAULT 'Not Shipped',
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (PO_ID) REFERENCES Purchase_Order(PO_ID),
			FOREIGN KEY (Item_ID) REFERENCES Item_Master(Item_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS PO_Item_Adjustment (
			Adjustment_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			PO_Item_ID INTEGER NOT NULL,
			Kind TEXT NOT NULL DEFAULT 'QTY',
			Qty_Delta REAL,
			Old_Value REAL,
			New_Value REAL,
			Reason TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (PO_Item_ID) REFERENCES PO_Item(PO_Item_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS Commercial_Invoice (
			CI_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			CI_No TEXT UNIQUE NOT NULL,
			Invoice_Date DATETIME NOT NULL,
			Vendor_ID INTEGER NOT NULL,
			Currency TEXT NOT NULL,
			Total_Amount REAL NOT NULL,
			Status TEXT CHECK (Status IN ('Draft', 'Open', 'Closed')),
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			UUID TEXT UNIQUE NOT NULL,
			FOREIGN KEY (Vendor_ID) REFERENCES Vendor_Master(Vendor_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS Account_Payable (
			AP_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Vendor_ID INTEGER NOT NULL,
			AP_No TEXT UNIQUE NOT NULL,
			Amount REAL NOT NULL,
			Currency TEXT NOT NULL,
			Local_Amount REAL,
			Allocation_Type TEXT CHECK (Allocation_Type IN ('Weight', 'Volume', 'Quantity', 'Value', 'Row')),
			Reference_UUID TEXT,
			Reference_Type TEXT,
			Due_Date DATETIME,
			Date_of_Payment DATETIME,
			Status TEXT CHECK (Status IN ('paid', 'unpaid')) DEFAULT 'unpaid',
			Allocation_Status TEXT CHECK (Allocation_Status IN ('Draft', 'Open', 'Closed')),
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			UUID TEXT UNIQUE NOT NULL,
			FOREIGN KEY (Vendor_ID) REFERENCES Vendor_Master(Vendor_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS Container (
			Container_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Container_No TEXT NOT NULL,
			Remark TEXT,
			Total_CBM REAL,
			Total_Net_Wgt REAL,
			Total_Gross_Wgt REAL,
			Status TEXT CHECK (Status IN ('Loaded', 'Shipping', 'Arrived')),
			UUID TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS Container_Item (
			Container_Item_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			PO_Item_ID INTEGER NOT NULL,
			Container_ID INTEGER,
			CI_ID INTEGER,
			BL_ID INTEGER,
			Load_Qty REAL,
			Gross_Weight REAL,
			Net_Weight REAL,
			Cbm REAL,
			Temporary_ETA DATETIME,
			UUID TEXT UNIQUE NOT NULL,
			Remark TEXT,
			FOREIGN KEY (PO_Item_ID) REFERENCES PO_Item(PO_Item_ID),
			FOREIGN KEY (Container_ID) REFERENCES Container(Container_ID),
			FOREIGN KEY (CI_ID) REFERENCES Commercial_Invoice(CI_ID),
			FOREIGN KEY (BL_ID) REFERENCES BL(BL_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS BL (
			BL_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			BL_No TEXT UNIQUE NOT NULL,
			ETD DATETIME NOT NULL,
			ETA DATETIME NOT NULL,
			POL TEXT,
			POD TEXT,
			Carrier TEXT,
			Vessel_Name TEXT,
			Remark TEXT,
			UUID TEXT UNIQUE NOT NULL,
			CHECK (ETA >= ETD)
		)`,
		`CREATE TABLE IF NOT EXISTS Goods_Receipt (
			GR_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			GR_No TEXT,
			Receive_Date DATETIME NOT NULL,
			Remark TEXT,
			Created_By TEXT,
			UUID TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS Inventory_Lot (
			Lot_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			GR_ID INTEGER NOT NULL,
			Container_Item_ID INTEGER NOT NULL,
			Lot_No TEXT NOT NULL,
			Expiry_Date DATETIME,
			Qty REAL NOT NULL,
			Landed_Cost_Per_Unit REAL,
			Quarantine_Status TEXT,
			Quarantine_Remark TEXT,
			Remark TEXT,
			UUID TEXT UNIQUE NOT NULL,
			FOREIGN KEY (GR_ID) REFERENCES Goods_Receipt(GR_ID),
			FOREIGN KEY (Container_Item_ID) REFERENCES Container_Item(Container_Item_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS Cost_Allocation (
			Cost_Allocation_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Allocation_date DATETIME NOT NULL,
			Total_Allocated_Amount REAL NOT NULL,
			Is_Late_Cost_Allocation INTEGER DEFAULT 0,
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS Cost_Allocation_Item (
			Cost_Allocation_Item_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Cost_Allocation_ID INTEGER NOT NULL,
			Lot_ID INTEGER NOT NULL,
			Allocated_Amount REAL NOT NULL,
			AP_ID INTEGER,
			FOREIGN KEY (Cost_Allocation_ID) REFERENCES Cost_Allocation(Cost_Allocation_ID),
			FOREIGN KEY (Lot_ID) REFERENCES Inventory_Lot(Lot_ID),
			FOREIGN KEY (AP_ID) REFERENCES Account_Payable(AP_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS AP_Target_Group (
			Group_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Group_No TEXT UNIQUE NOT NULL,
			Group_Name TEXT,
			Reference_Type TEXT NOT NULL,
			Status TEXT CHECK (Status IN ('Draft', 'Open', 'Closed')) DEFAULT 'Draft',
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			UUID TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS AP_Target_Group_Item (
			Group_Item_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Group_ID INTEGER NOT NULL,
			Reference_UUID TEXT NOT NULL,
			Remark TEXT,
			FOREIGN KEY (Group_ID) REFERENCES AP_Target_Group(Group_ID) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS Constraint_Profile (
			Profile_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Profile_No TEXT UNIQUE NOT NULL,
			Name TEXT NOT NULL,
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			UUID TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS Constraint_Master (
			Constraint_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT UNIQUE NOT NULL,
			Lead_Time_Days INTEGER DEFAULT 0,
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS Constraint_Profile_Step (
			Step_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Profile_ID INTEGER NOT NULL,
			Sequence INTEGER NOT NULL,
			Constraint_Type TEXT NOT NULL,
			Lead_Time_Days INTEGER DEFAULT 0,
			Is_Optional INTEGER DEFAULT 0,
			Remark TEXT,
			FOREIGN KEY (Profile_ID) REFERENCES Constraint_Profile(Profile_ID) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS Constraint_Instance (
			Instance_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Lot_ID INTEGER NOT NULL,
			Source_Container_Item_ID INTEGER,
			Profile_ID INTEGER,
			Predecessor_ID INTEGER,
			Sequence INTEGER,
			Constraint_Type TEXT NOT NULL,
			Lead_Time_Days INTEGER DEFAULT 0,
			Is_Optional INTEGER DEFAULT 0,
			Planned_Start DATETIME,
			Actual_Start DATETIME,
			Planned_End DATETIME,
			Actual_End DATETIME,
			Status TEXT DEFAULT 'Planned',
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (Lot_ID) REFERENCES Inventory_Lot(Lot_ID)
		)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}

	// Schema migrations for AP_Target_Group (tolerate errors if already applied)
	_ = migrateLegacySchemas(db)

	return nil
}

func migrateLegacySchemas(db *sql.DB) error {
	// Add Reference_Type to AP_Target_Group if missing
	var hasRefType int
	err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('AP_Target_Group') WHERE name = 'Reference_Type'`).Scan(&hasRefType)
	if err == nil && hasRefType == 0 {
		_, _ = db.Exec(`ALTER TABLE AP_Target_Group ADD COLUMN Reference_Type TEXT NOT NULL DEFAULT 'PO'`)
	}

	// Drop legacy columns from AP_Target_Group if they still exist (Vendor_ID, Currency, Total_Amount)
	for _, col := range []string{"Vendor_ID", "Currency", "Total_Amount"} {
		var has int
		err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('AP_Target_Group') WHERE name = ?`, col).Scan(&has)
		if err == nil && has == 1 {
			_, _ = db.Exec(fmt.Sprintf(`ALTER TABLE AP_Target_Group DROP COLUMN %s`, col))
		}
	}

	// Drop legacy columns from AP_Target_Group_Item if they still exist (Allocated_Amount, Reference_Type)
	for _, col := range []string{"Allocated_Amount", "Reference_Type"} {
		var has int
		err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('AP_Target_Group_Item') WHERE name = ?`, col).Scan(&has)
		if err == nil && has == 1 {
			_, _ = db.Exec(fmt.Sprintf(`ALTER TABLE AP_Target_Group_Item DROP COLUMN %s`, col))
		}
	}

	// Drop legacy columns from Container_Item if they still exist (Unit_Price, Currency)
	for _, col := range []string{"Unit_Price", "Currency"} {
		var has int
		err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('Container_Item') WHERE name = ?`, col).Scan(&has)
		if err == nil && has == 1 {
			_, _ = db.Exec(fmt.Sprintf(`ALTER TABLE Container_Item DROP COLUMN %s`, col))
		}
	}

	// Remove CHECK constraint from Vendor_Master.Category by recreating the table
	_ = migrateVendorMasterCategoryConstraint(db)

	// Make Cost_Allocation_Item.AP_ID nullable
	_ = migrateCostAllocationItemAPIDNullable(db)

	// Remove redundant audit columns from Cost_Allocation_Item (parent Cost_Allocation already audits)
	_ = migrateCostAllocationItemDropAuditColumns(db)

	// Add Is_Late_Cost_Allocation to Cost_Allocation if missing
	var hasLateCol int
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('Cost_Allocation') WHERE name = 'Is_Late_Cost_Allocation'`).Scan(&hasLateCol)
	if err == nil && hasLateCol == 0 {
		_, _ = db.Exec(`ALTER TABLE Cost_Allocation ADD COLUMN Is_Late_Cost_Allocation INTEGER DEFAULT 0`)
	}

	// Item_Master packaging fields (박스 입수 / 팔렛 적재량 / 박스 규격)
	_ = migrateItemMasterPackagingColumns(db)

	// Goods_Receipt GR_No 컬럼
	_ = migrateGoodsReceiptGRNo(db)

	// PO 품목 변경 이력(Kind/Old/New) 컬럼 + PO_Item 스키마(Base_Qty, PO_Qty>=0, Cancelled 제거)
	_ = migratePOItemAdjustmentColumns(db)
	_ = migratePOItemSchema(db)

	// A1: (PO, SKU, 단가) 단가 버킷 — 기존 중복 병합 후 유니크 인덱스 생성
	_ = migratePOItemBucketDedup(db)

	// 레거시 DB에 남아있을 수 있는 파생 트리거 제거 (파생 로직은 앱 레이어로 이관됨)
	_ = dropLegacyProcurementTriggers(db)

	// 배분 기준 'Unit' → 'Row' 로 이관 (CHECK 제약 갱신)
	_ = migrateAllocationTypeRow(db)

	// PO_Item 상태(파생 캐시)를 현재 선적량 기준으로 재계산하여 일관성 보장
	_ = recomputePOItemStatuses(db)

	// 행정 제약 프로파일 참조 컬럼 (Item_Master / Container_Item)
	_ = migrateConstraintColumns(db)

	// BL.Status 컬럼 제거 (Status는 Container_Item에서 계산)
	_ = migrateDropBLStatus(db)

	return nil
}

// migrateConstraintColumns Item_Master/Container_Item에 제약 프로파일 참조 컬럼을 추가한다.
func migrateConstraintColumns(db *sql.DB) error {
	cols := []struct {
		table string
		name  string
		sql   string
	}{
		{"Item_Master", "Default_Constraint_Profile_ID", "INTEGER"},
		{"Container_Item", "Constraint_Profile_ID", "INTEGER"},
		{"Constraint_Profile_Step", "Constraint_ID", "INTEGER"},
	}
	for _, c := range cols {
		var has int
		if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, c.table, c.name).Scan(&has); err != nil {
			continue
		}
		if has == 0 {
			if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", c.table, c.name, c.sql)); err != nil {
				return err
			}
		}
	}
	return nil
}

// migrateAllocationTypeRow Allocation_Type의 'Unit'을 'Row'로 변경하고 CHECK 제약을 갱신한다.
func migrateAllocationTypeRow(db *sql.DB) error {
	var sqlText string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='Account_Payable'`).Scan(&sqlText); err != nil {
		return nil
	}
	if strings.Contains(sqlText, "'Row'") {
		return nil
	}
	queries := []string{
		`CREATE TABLE Account_Payable_new (
			AP_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Vendor_ID INTEGER NOT NULL,
			AP_No TEXT UNIQUE NOT NULL,
			Amount REAL NOT NULL,
			Currency TEXT NOT NULL,
			Local_Amount REAL,
			Allocation_Type TEXT CHECK (Allocation_Type IN ('Weight', 'Volume', 'Quantity', 'Value', 'Row')),
			Reference_UUID TEXT,
			Reference_Type TEXT,
			Due_Date DATETIME,
			Date_of_Payment DATETIME,
			Status TEXT CHECK (Status IN ('paid', 'unpaid')) DEFAULT 'unpaid',
			Allocation_Status TEXT CHECK (Allocation_Status IN ('Draft', 'Open', 'Closed')),
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			UUID TEXT UNIQUE NOT NULL,
			FOREIGN KEY (Vendor_ID) REFERENCES Vendor_Master(Vendor_ID)
		)`,
		`INSERT INTO Account_Payable_new (AP_ID, Vendor_ID, AP_No, Amount, Currency, Local_Amount, Allocation_Type, Reference_UUID, Reference_Type, Due_Date, Date_of_Payment, Status, Allocation_Status, Remark, Created_By, Created_At, Updated_By, Updated_At, UUID)
		 SELECT AP_ID, Vendor_ID, AP_No, Amount, Currency, Local_Amount,
		        CASE WHEN Allocation_Type = 'Unit' THEN 'Row' ELSE Allocation_Type END,
		        Reference_UUID, Reference_Type, Due_Date, Date_of_Payment, Status, Allocation_Status, Remark, Created_By, Created_At, Updated_By, Updated_At, UUID
		 FROM Account_Payable`,
		`DROP TABLE Account_Payable`,
		`ALTER TABLE Account_Payable_new RENAME TO Account_Payable`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// migratePOItemBucketDedup (PO, SKU, 단가) 중복 버킷을 하나로 병합하고 유니크 인덱스를 생성한다.
func migratePOItemBucketDedup(db *sql.DB) error {
	rows, err := db.Query(`SELECT MIN(PO_Item_ID), PO_ID, Item_ID, Unit_Price
		FROM PO_Item GROUP BY PO_ID, Item_ID, Unit_Price HAVING COUNT(*) > 1`)
	if err != nil {
		return err
	}
	type grp struct {
		keepID int
		poID   int
		itemID int
		price  float64
	}
	groups := []grp{}
	for rows.Next() {
		var g grp
		if err := rows.Scan(&g.keepID, &g.poID, &g.itemID, &g.price); err != nil {
			continue
		}
		groups = append(groups, g)
	}
	rows.Close()

	for _, g := range groups {
		// 그룹 내 다른 항목 ID 수집
		idRows, err := db.Query("SELECT PO_Item_ID FROM PO_Item WHERE PO_ID = ? AND Item_ID = ? AND Unit_Price = ?", g.poID, g.itemID, g.price)
		if err != nil {
			continue
		}
		dropIDs := []int{}
		for idRows.Next() {
			var id int
			idRows.Scan(&id)
			if id != g.keepID {
				dropIDs = append(dropIDs, id)
			}
		}
		idRows.Close()

		for _, drop := range dropIDs {
			// 자식 참조 재지정
			_, _ = db.Exec("UPDATE Container_Item SET PO_Item_ID = ? WHERE PO_Item_ID = ?", g.keepID, drop)
			_, _ = db.Exec("UPDATE PO_Item_Adjustment SET PO_Item_ID = ? WHERE PO_Item_ID = ?", g.keepID, drop)
			// 수량 합산
			var baseQty, poQty float64
			_ = db.QueryRow("SELECT IFNULL(Base_Qty,0), IFNULL(PO_Qty,0) FROM PO_Item WHERE PO_Item_ID = ?", drop).Scan(&baseQty, &poQty)
			_, _ = db.Exec("UPDATE PO_Item SET Base_Qty = IFNULL(Base_Qty,0) + ?, PO_Qty = IFNULL(PO_Qty,0) + ? WHERE PO_Item_ID = ?", baseQty, poQty, g.keepID)
			// 중복 항목 삭제
			_, _ = db.Exec("DELETE FROM PO_Item WHERE PO_Item_ID = ?", drop)
		}
	}

	_, _ = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_po_item_bucket ON PO_Item(PO_ID, Item_ID, Unit_Price)`)
	return nil
}

func migratePOItemAdjustmentColumns(db *sql.DB) error {
	cols := []struct {
		name string
		sql  string
	}{
		{"Kind", "TEXT NOT NULL DEFAULT 'QTY'"},
		{"Old_Value", "REAL"},
		{"New_Value", "REAL"},
	}
	for _, c := range cols {
		var has int
		if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('PO_Item_Adjustment') WHERE name = ?`, c.name).Scan(&has); err != nil {
			continue
		}
		if has == 0 {
			if _, err := db.Exec(fmt.Sprintf("ALTER TABLE PO_Item_Adjustment ADD COLUMN %s %s", c.name, c.sql)); err != nil {
				return err
			}
		}
	}
	return nil
}

func migratePOItemSchema(db *sql.DB) error {
	var sqlText string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='PO_Item'`).Scan(&sqlText); err != nil {
		return nil
	}
	needRebuild := !strings.Contains(sqlText, "Base_Qty") || strings.Contains(sqlText, "Cancelled") || !strings.Contains(sqlText, "Over Shipped")
	if !needRebuild {
		return nil
	}
	// 기존 취소(Cancelled) 항목을 음수 증분 이벤트로 이관
	_, _ = db.Exec(`INSERT INTO PO_Item_Adjustment (PO_Item_ID, Kind, Qty_Delta, Old_Value, New_Value, Reason, Created_By)
		SELECT PO_Item_ID, 'QTY', -PO_Qty, PO_Qty, 0, '기존 취소 이관', 'system' FROM PO_Item WHERE Status = 'Cancelled'`)

	queries := []string{
		`CREATE TABLE PO_Item_new (
			PO_Item_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			PO_ID INTEGER NOT NULL,
			Item_ID INTEGER NOT NULL,
			Base_Qty REAL,
			PO_Qty REAL CHECK (PO_Qty >= 0),
			Unit_Price REAL NOT NULL,
			Status TEXT CHECK (Status IN ('Shipped', 'Partially Shipped', 'Over Shipped', 'Not Shipped')) DEFAULT 'Not Shipped',
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (PO_ID) REFERENCES Purchase_Order(PO_ID),
			FOREIGN KEY (Item_ID) REFERENCES Item_Master(Item_ID)
		)`,
		`INSERT INTO PO_Item_new (PO_Item_ID, PO_ID, Item_ID, Base_Qty, PO_Qty, Unit_Price, Status, Remark, Created_By, Created_At, Updated_By, Updated_At)
		 SELECT PO_Item_ID, PO_ID, Item_ID, PO_Qty, CASE WHEN Status='Cancelled' THEN 0 ELSE PO_Qty END, Unit_Price,
		        CASE WHEN Status='Cancelled' THEN 'Not Shipped' ELSE Status END, Remark, Created_By, Created_At, Updated_By, Updated_At
		 FROM PO_Item`,
		`DROP TABLE PO_Item`,
		`ALTER TABLE PO_Item_new RENAME TO PO_Item`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	// 기존 데이터의 상태를 새 규칙(Over Shipped 포함)으로 재계산
	return recomputePOItemStatuses(db)
}

// recomputePOItemStatuses 기존 PO_Item의 상태를 현재 선적량(SUM Load_Qty) 기준으로 재계산한다.
func recomputePOItemStatuses(db *sql.DB) error {
	_, err := db.Exec(`UPDATE PO_Item SET Status = CASE
		WHEN (SELECT IFNULL(SUM(Load_Qty), 0) FROM Container_Item WHERE PO_Item_ID = PO_Item.PO_Item_ID) > PO_Qty THEN 'Over Shipped'
		WHEN PO_Qty > 0 AND (SELECT IFNULL(SUM(Load_Qty), 0) FROM Container_Item WHERE PO_Item_ID = PO_Item.PO_Item_ID) >= PO_Qty THEN 'Shipped'
		WHEN (SELECT IFNULL(SUM(Load_Qty), 0) FROM Container_Item WHERE PO_Item_ID = PO_Item.PO_Item_ID) > 0 THEN 'Partially Shipped'
		ELSE 'Not Shipped'
	END`)
	return err
}

// migrateDropBLStatus 기존 BL 테이블에서 Status 컬럼을 제거한다.
// (상태는 이제 Container_Item에서 계산해 표시)
func migrateDropBLStatus(db *sql.DB) error {
	var sqlText string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='BL'`).Scan(&sqlText); err != nil {
		return nil
	}
	if !strings.Contains(sqlText, "Status") {
		return nil
	}
	queries := []string{
		`CREATE TABLE BL_new (
			BL_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			BL_No TEXT UNIQUE NOT NULL,
			ETD DATETIME NOT NULL,
			ETA DATETIME NOT NULL,
			POL TEXT,
			POD TEXT,
			Carrier TEXT,
			Vessel_Name TEXT,
			Remark TEXT,
			UUID TEXT UNIQUE NOT NULL,
			CHECK (ETA >= ETD)
		)`,
		`INSERT INTO BL_new (BL_ID, BL_No, ETD, ETA, POL, POD, Carrier, Vessel_Name, Remark, UUID)
		 SELECT BL_ID, BL_No, ETD, ETA, POL, POD, Carrier, Vessel_Name, Remark, UUID FROM BL`,
		`DROP TABLE BL`,
		`ALTER TABLE BL_new RENAME TO BL`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// dropLegacyProcurementTriggers 과거 스키마에 존재하던 파생 로직 트리거를 제거한다.
// 파생값(상태/집계)은 이제 애플리케이션(유스케이스) 레이어에서 계산한다.
func dropLegacyProcurementTriggers(db *sql.DB) error {
	names := []string{
		"trg_sync_po_status", "trg_sync_po_status_delete",
		"trg_container_item_calc_insert", "trg_container_item_calc_update", "trg_container_item_calc_delete",
		"trg_container_item_po_item_status_insert", "trg_container_item_po_item_status_update", "trg_container_item_po_item_status_delete",
	}
	for _, n := range names {
		if _, err := db.Exec("DROP TRIGGER IF EXISTS " + n); err != nil {
			return err
		}
	}
	return nil
}

func migrateGoodsReceiptGRNo(db *sql.DB) error {
	var has int
	err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('Goods_Receipt') WHERE name = 'GR_No'`).Scan(&has)
	if err == nil && has == 0 {
		if _, err := db.Exec(`ALTER TABLE Goods_Receipt ADD COLUMN GR_No TEXT`); err != nil {
			return err
		}
	}
	return nil
}

func migrateItemMasterPackagingColumns(db *sql.DB) error {
	cols := []struct {
		name string
		sql  string
	}{
		{"Box_Qty", "INTEGER NOT NULL DEFAULT 1"},
		{"Boxes_Per_Pallet", "INTEGER"},
		{"Box_Width", "REAL"},
		{"Box_Length", "REAL"},
		{"Box_Height", "REAL"},
	}
	for _, c := range cols {
		var has int
		err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('Item_Master') WHERE name = ?`, c.name).Scan(&has)
		if err == nil && has == 0 {
			if _, err := db.Exec(fmt.Sprintf("ALTER TABLE Item_Master ADD COLUMN %s %s", c.name, c.sql)); err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateCostAllocationItemDropAuditColumns(db *sql.DB) error {
	// Cost_Allocation_Item audit fields are redundant with the parent Cost_Allocation.
	// Drop them if present (best effort; SQLite 3.35+ supports DROP COLUMN).
	for _, col := range []string{"Created_By", "Created_At", "Updated_By", "Updated_At"} {
		var has int
		if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('Cost_Allocation_Item') WHERE name = ?`, col).Scan(&has); err != nil {
			continue
		}
		if has == 0 {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE Cost_Allocation_Item DROP COLUMN %s", col)); err != nil {
			return err
		}
	}
	return nil
}


func migrateVendorMasterCategoryConstraint(db *sql.DB) error {
	var constraintCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='Vendor_Master' AND sql LIKE '%CHECK%Category%'`).Scan(&constraintCount)
	if err != nil || constraintCount == 0 {
		return nil
	}

	queries := []string{
		`ALTER TABLE Vendor_Master RENAME TO Vendor_Master_old`,
		`CREATE TABLE Vendor_Master (
			Vendor_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			Category TEXT,
			Business_Reg_No TEXT,
			Bank_Account TEXT,
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO Vendor_Master (Vendor_ID, Name, Category, Business_Reg_No, Bank_Account, Remark, Created_By, Created_At, Updated_By, Updated_At)
		 SELECT Vendor_ID, Name, Category, Business_Reg_No, Bank_Account, Remark, Created_By, Created_At, Updated_By, Updated_At FROM Vendor_Master_old`,
		`DROP TABLE Vendor_Master_old`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func migrateCostAllocationItemAPIDNullable(db *sql.DB) error {
	var notNull int
	err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('Cost_Allocation_Item') WHERE name = 'AP_ID' AND "notnull" = 1`).Scan(&notNull)
	if err != nil || notNull == 0 {
		return nil
	}

	queries := []string{
		`ALTER TABLE Cost_Allocation_Item RENAME TO Cost_Allocation_Item_old`,
		`CREATE TABLE Cost_Allocation_Item (
			Cost_Allocation_Item_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Cost_Allocation_ID INTEGER NOT NULL,
			Lot_ID INTEGER NOT NULL,
			Allocated_Amount REAL NOT NULL,
			AP_ID INTEGER,
			FOREIGN KEY (Cost_Allocation_ID) REFERENCES Cost_Allocation(Cost_Allocation_ID),
			FOREIGN KEY (Lot_ID) REFERENCES Inventory_Lot(Lot_ID),
			FOREIGN KEY (AP_ID) REFERENCES Account_Payable(AP_ID)
		)`,
		`INSERT INTO Cost_Allocation_Item (Cost_Allocation_Item_ID, Cost_Allocation_ID, Lot_ID, Allocated_Amount, AP_ID)
		 SELECT Cost_Allocation_Item_ID, Cost_Allocation_ID, Lot_ID, Allocated_Amount, CASE WHEN AP_ID = 0 THEN NULL ELSE AP_ID END FROM Cost_Allocation_Item_old`,
		`DROP TABLE Cost_Allocation_Item_old`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (r *SQLiteProcurementRepository) SeedData() error {
	// SQLite normally doesn't need re-seeding as it's persistent.
	// We could implement initial data check here if needed.
	return nil
}

// Item Master
func (r *SQLiteProcurementRepository) GetItems() ([]models.ItemMaster, error) {
	rows, err := r.db.Query("SELECT Item_ID, SKU_Code, Name, Vendor_ID, IFNULL(Default_Constraint_Profile_ID, 0), IFNULL(CBM, 0), IFNULL(Box_Qty, 1), IFNULL(Boxes_Per_Pallet, 0), IFNULL(Box_Width, 0), IFNULL(Box_Length, 0), IFNULL(Box_Height, 0), IFNULL(Net_Weight, 0), IFNULL(Gross_Weight, 0), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Item_Master")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.ItemMaster{}
	for rows.Next() {
		var i models.ItemMaster
		err := rows.Scan(&i.ID, &i.SKUCode, &i.Name, &i.VendorID, &i.DefaultConstraintProfileID, &i.CBM, &i.BoxQty, &i.BoxesPerPallet, &i.BoxWidth, &i.BoxLength, &i.BoxHeight, &i.NetWeight, &i.GrossWeight, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetItemByID(id int) (*models.ItemMaster, error) {
	var i models.ItemMaster
	err := r.db.QueryRow("SELECT Item_ID, SKU_Code, Name, Vendor_ID, IFNULL(Default_Constraint_Profile_ID, 0), IFNULL(CBM, 0), IFNULL(Box_Qty, 1), IFNULL(Boxes_Per_Pallet, 0), IFNULL(Box_Width, 0), IFNULL(Box_Length, 0), IFNULL(Box_Height, 0), IFNULL(Net_Weight, 0), IFNULL(Gross_Weight, 0), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Item_Master WHERE Item_ID = ?", id).
		Scan(&i.ID, &i.SKUCode, &i.Name, &i.VendorID, &i.DefaultConstraintProfileID, &i.CBM, &i.BoxQty, &i.BoxesPerPallet, &i.BoxWidth, &i.BoxLength, &i.BoxHeight, &i.NetWeight, &i.GrossWeight, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &i, err
}

func (r *SQLiteProcurementRepository) SaveItem(i *models.ItemMaster) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Item_Master (SKU_Code, Name, Vendor_ID, Default_Constraint_Profile_ID, CBM, Box_Qty, Boxes_Per_Pallet, Box_Width, Box_Length, Box_Height, Net_Weight, Gross_Weight, Remark, Created_By) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			i.SKUCode, i.Name, i.VendorID, i.DefaultConstraintProfileID, i.CBM, i.BoxQty, i.BoxesPerPallet, i.BoxWidth, i.BoxLength, i.BoxHeight, i.NetWeight, i.GrossWeight, i.Remark, i.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Item_Master SET SKU_Code=?, Name=?, Vendor_ID=?, Default_Constraint_Profile_ID=?, CBM=?, Box_Qty=?, Boxes_Per_Pallet=?, Box_Width=?, Box_Length=?, Box_Height=?, Net_Weight=?, Gross_Weight=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE Item_ID=?",
		i.SKUCode, i.Name, i.VendorID, i.DefaultConstraintProfileID, i.CBM, i.BoxQty, i.BoxesPerPallet, i.BoxWidth, i.BoxLength, i.BoxHeight, i.NetWeight, i.GrossWeight, i.Remark, i.UpdatedBy, i.ID)
	return err
}

// Vendor Master
func (r *SQLiteProcurementRepository) GetVendors() ([]models.VendorMaster, error) {
	rows, err := r.db.Query("SELECT Vendor_ID, Name, IFNULL(Category, ''), IFNULL(Business_Reg_No, ''), IFNULL(Bank_Account, ''), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Vendor_Master")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.VendorMaster{}
	for rows.Next() {
		var v models.VendorMaster
		err := rows.Scan(&v.ID, &v.Name, &v.Category, &v.BusinessRegNo, &v.BankAccount, &v.Remark, &v.CreatedBy, &v.CreatedAt, &v.UpdatedBy, &v.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetVendorByID(id int) (*models.VendorMaster, error) {
	var v models.VendorMaster
	err := r.db.QueryRow("SELECT Vendor_ID, Name, IFNULL(Category, ''), IFNULL(Business_Reg_No, ''), IFNULL(Bank_Account, ''), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Vendor_Master WHERE Vendor_ID = ?", id).
		Scan(&v.ID, &v.Name, &v.Category, &v.BusinessRegNo, &v.BankAccount, &v.Remark, &v.CreatedBy, &v.CreatedAt, &v.UpdatedBy, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &v, err
}

func (r *SQLiteProcurementRepository) SaveVendor(v *models.VendorMaster) error {
	if v.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Vendor_Master (Name, Category, Business_Reg_No, Bank_Account, Remark, Created_By) VALUES (?, ?, ?, ?, ?, ?)",
			v.Name, v.Category, v.BusinessRegNo, v.BankAccount, v.Remark, v.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		v.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Vendor_Master SET Name=?, Category=?, Business_Reg_No=?, Bank_Account=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE Vendor_ID=?",
		v.Name, v.Category, v.BusinessRegNo, v.BankAccount, v.Remark, v.UpdatedBy, v.ID)
	return err
}

// Purchase Order
func (r *SQLiteProcurementRepository) GetPurchaseOrders() ([]models.PurchaseOrder, error) {
	rows, err := r.db.Query("SELECT PO_ID, PO_Date, PO_No, Vendor_ID, IFNULL(Currency, ''), IFNULL(Total_Amount, 0), Status, IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID FROM Purchase_Order")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.PurchaseOrder{}
	for rows.Next() {
		var p models.PurchaseOrder
		err := rows.Scan(&p.ID, &p.PODate, &p.PONo, &p.VendorID, &p.Currency, &p.TotalAmount, &p.Status, &p.Remark, &p.CreatedBy, &p.CreatedAt, &p.UpdatedBy, &p.UpdatedAt, &p.UUID)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error) {
	var p models.PurchaseOrder
	err := r.db.QueryRow("SELECT PO_ID, PO_Date, PO_No, Vendor_ID, IFNULL(Currency, ''), IFNULL(Total_Amount, 0), Status, IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID FROM Purchase_Order WHERE PO_ID = ?", id).
		Scan(&p.ID, &p.PODate, &p.PONo, &p.VendorID, &p.Currency, &p.TotalAmount, &p.Status, &p.Remark, &p.CreatedBy, &p.CreatedAt, &p.UpdatedBy, &p.UpdatedAt, &p.UUID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *SQLiteProcurementRepository) SavePurchaseOrder(p *models.PurchaseOrder) error {
	if p.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Purchase_Order (PO_Date, PO_No, Vendor_ID, Currency, Total_Amount, Status, Remark, Created_By, UUID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			p.PODate, p.PONo, p.VendorID, p.Currency, p.TotalAmount, p.Status, p.Remark, p.CreatedBy, p.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		p.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Purchase_Order SET PO_Date=?, PO_No=?, Vendor_ID=?, Currency=?, Total_Amount=?, Status=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE PO_ID=?",
		p.PODate, p.PONo, p.VendorID, p.Currency, p.TotalAmount, p.Status, p.Remark, p.UpdatedBy, p.ID)
	return err
}

// PO Item
func (r *SQLiteProcurementRepository) GetPOItemsByPOID(poID int) ([]models.POItem, error) {
	rows, err := r.db.Query("SELECT PO_Item_ID, PO_ID, Item_ID, IFNULL(Base_Qty, IFNULL(PO_Qty,0)), IFNULL(PO_Qty, 0), IFNULL(Unit_Price, 0), Status, IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM PO_Item WHERE PO_ID = ?", poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.POItem{}
	for rows.Next() {
		var i models.POItem
		err := rows.Scan(&i.ID, &i.POID, &i.ItemID, &i.BaseQty, &i.POQty, &i.UnitPrice, &i.Status, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SavePOItem(i *models.POItem) error {
	baseQty := i.BaseQty
	if baseQty <= 0 {
		baseQty = i.POQty
	}
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO PO_Item (PO_ID, Item_ID, Base_Qty, PO_Qty, Unit_Price, Status, Remark, Created_By) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			i.POID, i.ItemID, baseQty, i.POQty, i.UnitPrice, i.Status, i.Remark, i.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	// A1: Unit_Price는 불변 (단가 변경은 새 PO_Item 버킷 생성으로 처리)
	_, err := r.db.Exec("UPDATE PO_Item SET Item_ID=?, Base_Qty=?, PO_Qty=?, Status=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE PO_Item_ID=?",
		i.ItemID, baseQty, i.POQty, i.Status, i.Remark, i.UpdatedBy, i.ID)
	return err
}

// GetPOItemByBucket (PO_ID, Item_ID, Unit_Price)로 단가 버킷 조회
func (r *SQLiteProcurementRepository) GetPOItemByBucket(poID int, itemID int, unitPrice float64) (*models.POItem, error) {
	var i models.POItem
	err := r.db.QueryRow("SELECT PO_Item_ID, PO_ID, Item_ID, IFNULL(Base_Qty, IFNULL(PO_Qty,0)), IFNULL(PO_Qty, 0), IFNULL(Unit_Price, 0), Status, IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM PO_Item WHERE PO_ID = ? AND Item_ID = ? AND Unit_Price = ?", poID, itemID, unitPrice).
		Scan(&i.ID, &i.POID, &i.ItemID, &i.BaseQty, &i.POQty, &i.UnitPrice, &i.Status, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

// GetPOLoadQty PO 품목에 누적된 선적량(SUM Load_Qty)
func (r *SQLiteProcurementRepository) GetPOLoadQty(poItemID int) (float64, error) {
	var qty float64
	err := r.db.QueryRow("SELECT IFNULL(SUM(Load_Qty), 0) FROM Container_Item WHERE PO_Item_ID = ?", poItemID).Scan(&qty)
	return qty, err
}

func (r *SQLiteProcurementRepository) DeletePOItem(id int) error {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM Container_Item WHERE PO_Item_ID = ?", id).Scan(&count)
	if err == nil && count > 0 {
		return fmt.Errorf("cannot delete PO item %d: referenced by %d container item(s)", id, count)
	}
	_, err = r.db.Exec("DELETE FROM PO_Item WHERE PO_Item_ID = ?", id)
	return err
}

func (r *SQLiteProcurementRepository) GetPOItemByID(id int) (*models.POItem, error) {
	var i models.POItem
	err := r.db.QueryRow("SELECT PO_Item_ID, PO_ID, Item_ID, IFNULL(Base_Qty, IFNULL(PO_Qty,0)), IFNULL(PO_Qty, 0), IFNULL(Unit_Price, 0), Status, IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM PO_Item WHERE PO_Item_ID = ?", id).
		Scan(&i.ID, &i.POID, &i.ItemID, &i.BaseQty, &i.POQty, &i.UnitPrice, &i.Status, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *SQLiteProcurementRepository) SavePOItemAdjustment(a *models.POItemAdjustment) error {
	kind := a.Kind
	if kind == "" {
		kind = "QTY"
	}
	res, err := r.db.Exec("INSERT INTO PO_Item_Adjustment (PO_Item_ID, Kind, Qty_Delta, Old_Value, New_Value, Reason, Created_By) VALUES (?, ?, ?, ?, ?, ?, ?)",
		a.POItemID, kind, a.QtyDelta, a.OldValue, a.NewValue, a.Reason, a.CreatedBy)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	a.ID = int(id)
	return nil
}

func (r *SQLiteProcurementRepository) GetPOItemAdjustmentsByPOItemID(poItemID int) ([]models.POItemAdjustment, error) {
	rows, err := r.db.Query("SELECT Adjustment_ID, PO_Item_ID, IFNULL(Kind, 'QTY'), IFNULL(Qty_Delta, 0), IFNULL(Old_Value, 0), IFNULL(New_Value, 0), IFNULL(Reason, ''), IFNULL(Created_By, ''), Created_At FROM PO_Item_Adjustment WHERE PO_Item_ID = ? ORDER BY Adjustment_ID ASC", poItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.POItemAdjustment{}
	for rows.Next() {
		var a models.POItemAdjustment
		if err := rows.Scan(&a.ID, &a.POItemID, &a.Kind, &a.QtyDelta, &a.OldValue, &a.NewValue, &a.Reason, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

// GetPOItemAdjustmentsByPOID PO 단위 일괄 조회 (품목별 N+1 방지)
func (r *SQLiteProcurementRepository) GetPOItemAdjustmentsByPOID(poID int) ([]models.POItemAdjustment, error) {
	rows, err := r.db.Query(`
		SELECT a.Adjustment_ID, a.PO_Item_ID, IFNULL(a.Kind, 'QTY'), IFNULL(a.Qty_Delta, 0), IFNULL(a.Old_Value, 0), IFNULL(a.New_Value, 0), IFNULL(a.Reason, ''), IFNULL(a.Created_By, ''), a.Created_At
		FROM PO_Item_Adjustment a
		JOIN PO_Item pi ON a.PO_Item_ID = pi.PO_Item_ID
		WHERE pi.PO_ID = ?
		ORDER BY a.PO_Item_ID ASC, a.Adjustment_ID ASC`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.POItemAdjustment{}
	for rows.Next() {
		var a models.POItemAdjustment
		if err := rows.Scan(&a.ID, &a.POItemID, &a.Kind, &a.QtyDelta, &a.OldValue, &a.NewValue, &a.Reason, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

// GetPOReferenceStatus PO의 편집 잠금 여부 판단용: AP 참조 여부 / Container Item 존재 여부
func (r *SQLiteProcurementRepository) GetPOReferenceStatus(poID int) (bool, bool, error) {
	// AP 직접 참조 또는 AP Target Group(멤버에 이 PO 포함)을 통한 간접 참조
	var hasAP int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM Purchase_Order p
		WHERE p.PO_ID = ?
		  AND (
			EXISTS (SELECT 1 FROM Account_Payable ap WHERE ap.Reference_Type = 'PO' AND ap.Reference_UUID = p.UUID)
			OR EXISTS (
				SELECT 1 FROM Account_Payable ap
				JOIN AP_Target_Group g ON g.UUID = ap.Reference_UUID AND ap.Reference_Type = 'AP_Target_Group'
				JOIN AP_Target_Group_Item gi ON gi.Group_ID = g.Group_ID
				WHERE gi.Reference_UUID = p.UUID
			)
		  )`, poID).Scan(&hasAP)
	if err != nil {
		return false, false, err
	}
	var hasCI int
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM Container_Item ci
		JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
		WHERE pi.PO_ID = ?`, poID).Scan(&hasCI)
	if err != nil {
		return false, false, err
	}
	return hasAP > 0, hasCI > 0, nil
}

// Commercial Invoice
func (r *SQLiteProcurementRepository) GetCommercialInvoices() ([]models.CommercialInvoice, error) {
	rows, err := r.db.Query("SELECT CI_ID, CI_No, Invoice_Date, Vendor_ID, IFNULL(Currency, ''), IFNULL(Total_Amount, 0), Status, IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID FROM Commercial_Invoice")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.CommercialInvoice{}
	for rows.Next() {
		var i models.CommercialInvoice
		rows.Scan(&i.ID, &i.CINo, &i.InvoiceDate, &i.VendorID, &i.Currency, &i.TotalAmount, &i.Status, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt, &i.UUID)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetCIAggregatedItems(ciID int) ([]models.CIAggregatedItem, error) {
	query := `SELECT pi.Item_ID, im.Name, SUM(ci.Load_Qty), SUM(ci.Load_Qty * pi.Unit_Price), po.Currency
	          FROM Container_Item ci
	          JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
	          JOIN Purchase_Order po ON pi.PO_ID = po.PO_ID
	          LEFT JOIN Item_Master im ON pi.Item_ID = im.Item_ID
	          WHERE ci.CI_ID = ?
	          GROUP BY pi.Item_ID, im.Name, po.Currency`
	rows, err := r.db.Query(query, ciID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.CIAggregatedItem{}
	for rows.Next() {
		var i models.CIAggregatedItem
		if err := rows.Scan(&i.ItemID, &i.ItemName, &i.TotalQty, &i.Amount, &i.Currency); err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveCommercialInvoice(i *models.CommercialInvoice) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Commercial_Invoice (CI_No, Invoice_Date, Vendor_ID, Currency, Total_Amount, Status, Remark, Created_By, UUID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			i.CINo, i.InvoiceDate, i.VendorID, i.Currency, i.TotalAmount, i.Status, i.Remark, i.CreatedBy, i.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Commercial_Invoice SET CI_No=?, Invoice_Date=?, Vendor_ID=?, Currency=?, Total_Amount=?, Status=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE CI_ID=?",
		i.CINo, i.InvoiceDate, i.VendorID, i.Currency, i.TotalAmount, i.Status, i.Remark, i.UpdatedBy, i.ID)
	return err
}

// Account Payable
func (r *SQLiteProcurementRepository) GetAccountPayables() ([]models.AccountPayable, error) {
	rows, err := r.db.Query(`SELECT AP_ID, Vendor_ID, AP_No, IFNULL(Amount, 0), IFNULL(Currency, ''), IFNULL(Local_Amount, 0), IFNULL(Allocation_Type, ''), IFNULL(Reference_UUID, ''), IFNULL(Reference_Type, ''), Due_Date, Date_of_Payment, Status, IFNULL(Allocation_Status, ''), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID,
		EXISTS(SELECT 1 FROM Cost_Allocation_Item WHERE AP_ID = Account_Payable.AP_ID) AS Has_Allocation
		FROM Account_Payable`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.AccountPayable{}
	for rows.Next() {
		var i models.AccountPayable
		var dueDate, dateOfPayment *time.Time
		var hasAlloc bool
		err := rows.Scan(&i.ID, &i.VendorID, &i.APNo, &i.Amount, &i.Currency, &i.LocalAmount, &i.AllocationType, &i.ReferenceUUID, &i.ReferenceType, &dueDate, &dateOfPayment, &i.Status, &i.AllocationStatus, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt, &i.UUID, &hasAlloc)
		if err != nil {
			return nil, err
		}
		i.HasAllocation = hasAlloc
		if dueDate != nil {
			i.DueDate = *dueDate
		}
		if dateOfPayment != nil {
			i.DateOfPayment = *dateOfPayment
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetAccountPayableByID(id int) (*models.AccountPayable, error) {
	var i models.AccountPayable
	var dueDate, dateOfPayment *time.Time
	err := r.db.QueryRow(`SELECT AP_ID, Vendor_ID, AP_No, IFNULL(Amount, 0), IFNULL(Currency, ''), IFNULL(Local_Amount, 0), IFNULL(Allocation_Type, ''), IFNULL(Reference_UUID, ''), IFNULL(Reference_Type, ''), Due_Date, Date_of_Payment, Status, IFNULL(Allocation_Status, ''), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID FROM Account_Payable WHERE AP_ID = ?`, id).
		Scan(&i.ID, &i.VendorID, &i.APNo, &i.Amount, &i.Currency, &i.LocalAmount, &i.AllocationType, &i.ReferenceUUID, &i.ReferenceType, &dueDate, &dateOfPayment, &i.Status, &i.AllocationStatus, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt, &i.UUID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if dueDate != nil {
		i.DueDate = *dueDate
	}
	if dateOfPayment != nil {
		i.DateOfPayment = *dateOfPayment
	}
	return &i, nil
}

func (r *SQLiteProcurementRepository) GetAPAllocatedAmount(apID int) (bool, float64, error) {
	var hasAny bool
	var total float64
	err := r.db.QueryRow("SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END, IFNULL(SUM(Allocated_Amount), 0) FROM Cost_Allocation_Item WHERE AP_ID = ?", apID).Scan(&hasAny, &total)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	return hasAny, total, err
}

func (r *SQLiteProcurementRepository) SaveAccountPayable(i *models.AccountPayable) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Account_Payable (Vendor_ID, AP_No, Amount, Currency, Local_Amount, Allocation_Type, Reference_UUID, Reference_Type, Due_Date, Date_of_Payment, Status, Allocation_Status, Remark, Created_By, UUID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			i.VendorID, i.APNo, i.Amount, i.Currency, i.LocalAmount, i.AllocationType, i.ReferenceUUID, i.ReferenceType, nullIfZero(i.DueDate), nullIfZero(i.DateOfPayment), i.Status, i.AllocationStatus, i.Remark, i.CreatedBy, i.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Account_Payable SET Vendor_ID=?, AP_No=?, Amount=?, Currency=?, Local_Amount=?, Allocation_Type=?, Reference_UUID=?, Reference_Type=?, Due_Date=?, Date_of_Payment=?, Status=?, Allocation_Status=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE AP_ID=?",
		i.VendorID, i.APNo, i.Amount, i.Currency, i.LocalAmount, i.AllocationType, i.ReferenceUUID, i.ReferenceType, nullIfZero(i.DueDate), nullIfZero(i.DateOfPayment), i.Status, i.AllocationStatus, i.Remark, i.UpdatedBy, i.ID)
	return err
}

// Container & Logistics
func (r *SQLiteProcurementRepository) GetContainers() ([]models.Container, error) {
	rows, err := r.db.Query("SELECT Container_ID, Container_No, IFNULL(Remark, ''), IFNULL(Total_CBM, 0), IFNULL(Total_Net_Wgt, ''), IFNULL(Total_Gross_Wgt, ''), Status, UUID FROM Container")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.Container{}
	for rows.Next() {
		var i models.Container
		rows.Scan(&i.ID, &i.ContainerNo, &i.Remark, &i.TotalCBM, &i.TotalNetWgt, &i.TotalGrossWgt, &i.Status, &i.UUID)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetContainerByID(id int) (*models.Container, error) {
	var i models.Container
	err := r.db.QueryRow("SELECT Container_ID, Container_No, IFNULL(Remark, ''), IFNULL(Total_CBM, 0), IFNULL(Total_Net_Wgt, 0), IFNULL(Total_Gross_Wgt, 0), Status, UUID FROM Container WHERE Container_ID = ?", id).
		Scan(&i.ID, &i.ContainerNo, &i.Remark, &i.TotalCBM, &i.TotalNetWgt, &i.TotalGrossWgt, &i.Status, &i.UUID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *SQLiteProcurementRepository) SaveContainer(i *models.Container) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Container (Container_No, Remark, Total_CBM, Total_Net_Wgt, Total_Gross_Wgt, Status, UUID) VALUES (?, ?, ?, ?, ?, ?, ?)",
			i.ContainerNo, i.Remark, i.TotalCBM, i.TotalNetWgt, i.TotalGrossWgt, i.Status, i.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Container SET Container_No=?, Remark=?, Total_CBM=?, Total_Net_Wgt=?, Total_Gross_Wgt=?, Status=? WHERE Container_ID=?",
		i.ContainerNo, i.Remark, i.TotalCBM, i.TotalNetWgt, i.TotalGrossWgt, i.Status, i.ID)
	return err
}

func (r *SQLiteProcurementRepository) GetBLs() ([]models.BL, error) {
	rows, err := r.db.Query("SELECT BL_ID, BL_No, ETD, ETA, IFNULL(POL, ''), IFNULL(POD, ''), IFNULL(Carrier, ''), IFNULL(Vessel_Name, ''), IFNULL(Remark, ''), UUID FROM BL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.BL{}
	for rows.Next() {
		var i models.BL
		rows.Scan(&i.ID, &i.BLNo, &i.ETD, &i.ETA, &i.POL, &i.POD, &i.Carrier, &i.VesselName, &i.Remark, &i.UUID)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveBL(i *models.BL) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO BL (BL_No, ETD, ETA, POL, POD, Carrier, Vessel_Name, Remark, UUID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			i.BLNo, i.ETD, i.ETA, i.POL, i.POD, i.Carrier, i.VesselName, i.Remark, i.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE BL SET BL_No=?, ETD=?, ETA=?, POL=?, POD=?, Carrier=?, Vessel_Name=?, Remark=? WHERE BL_ID=?",
		i.BLNo, i.ETD, i.ETA, i.POL, i.POD, i.Carrier, i.VesselName, i.Remark, i.ID)
	return err
}

// Goods Receipt & Inventory
func (r *SQLiteProcurementRepository) GetGoodsReceipts() ([]models.GoodsReceipt, error) {
	rows, err := r.db.Query("SELECT GR_ID, IFNULL(GR_No, ''), Receive_Date, IFNULL(Remark, ''), IFNULL(Created_By, ''), UUID FROM Goods_Receipt")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.GoodsReceipt{}
	for rows.Next() {
		var i models.GoodsReceipt
		rows.Scan(&i.ID, &i.GRNo, &i.ReceiveDate, &i.Remark, &i.CreatedBy, &i.UUID)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetGoodsReceiptByID(id int) (*models.GoodsReceipt, error) {
	var i models.GoodsReceipt
	err := r.db.QueryRow("SELECT GR_ID, IFNULL(GR_No, ''), Receive_Date, IFNULL(Remark, ''), IFNULL(Created_By, ''), UUID FROM Goods_Receipt WHERE GR_ID = ?", id).
		Scan(&i.ID, &i.GRNo, &i.ReceiveDate, &i.Remark, &i.CreatedBy, &i.UUID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *SQLiteProcurementRepository) SaveGoodsReceipt(i *models.GoodsReceipt) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Goods_Receipt (GR_No, Receive_Date, Remark, Created_By, UUID) VALUES (?, ?, ?, ?, ?)",
			i.GRNo, i.ReceiveDate, i.Remark, i.CreatedBy, i.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Goods_Receipt SET GR_No=?, Receive_Date=?, Remark=? WHERE GR_ID=?", i.GRNo, i.ReceiveDate, i.Remark, i.ID)
	return err
}

func (r *SQLiteProcurementRepository) GetInventoryLots() ([]models.InventoryLot, error) {
	rows, err := r.db.Query("SELECT Lot_ID, GR_ID, Container_Item_ID, Lot_No, Expiry_Date, IFNULL(Qty, 0), IFNULL(Landed_Cost_Per_Unit, 0), IFNULL(Quarantine_Status, ''), IFNULL(Quarantine_Remark, ''), IFNULL(Remark, ''), UUID FROM Inventory_Lot")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.InventoryLot{}
	for rows.Next() {
		var i models.InventoryLot
		var expiryDate *time.Time
		rows.Scan(&i.ID, &i.GRID, &i.ContainerItemID, &i.LotNo, &expiryDate, &i.Qty, &i.LandedCostPerUnit, &i.QuarantineStatus, &i.QuarantineRemark, &i.Remark, &i.UUID)
		if expiryDate != nil {
			i.ExpiryDate = *expiryDate
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetInventoryLotByID(id int) (*models.InventoryLot, error) {
	var i models.InventoryLot
	var expiryDate *time.Time
	err := r.db.QueryRow("SELECT Lot_ID, GR_ID, Container_Item_ID, Lot_No, Expiry_Date, IFNULL(Qty, 0), IFNULL(Landed_Cost_Per_Unit, 0), IFNULL(Quarantine_Status, ''), IFNULL(Quarantine_Remark, ''), IFNULL(Remark, ''), UUID FROM Inventory_Lot WHERE Lot_ID = ?", id).
		Scan(&i.ID, &i.GRID, &i.ContainerItemID, &i.LotNo, &expiryDate, &i.Qty, &i.LandedCostPerUnit, &i.QuarantineStatus, &i.QuarantineRemark, &i.Remark, &i.UUID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if expiryDate != nil {
		i.ExpiryDate = *expiryDate
	}
	return &i, nil
}

func (r *SQLiteProcurementRepository) GetInventoryLotsByGRID(grID int) ([]models.InventoryLot, error) {
	rows, err := r.db.Query("SELECT Lot_ID, GR_ID, Container_Item_ID, Lot_No, Expiry_Date, IFNULL(Qty, 0), IFNULL(Landed_Cost_Per_Unit, 0), IFNULL(Quarantine_Status, ''), IFNULL(Quarantine_Remark, ''), IFNULL(Remark, ''), UUID FROM Inventory_Lot WHERE GR_ID = ?", grID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.InventoryLot{}
	for rows.Next() {
		var i models.InventoryLot
		var expiryDate *time.Time
		rows.Scan(&i.ID, &i.GRID, &i.ContainerItemID, &i.LotNo, &expiryDate, &i.Qty, &i.LandedCostPerUnit, &i.QuarantineStatus, &i.QuarantineRemark, &i.Remark, &i.UUID)
		if expiryDate != nil {
			i.ExpiryDate = *expiryDate
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveInventoryLot(i *models.InventoryLot) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Inventory_Lot (GR_ID, Container_Item_ID, Lot_No, Expiry_Date, Qty, Landed_Cost_Per_Unit, Quarantine_Status, Quarantine_Remark, Remark, UUID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			i.GRID, i.ContainerItemID, i.LotNo, nullIfZero(i.ExpiryDate), i.Qty, i.LandedCostPerUnit, i.QuarantineStatus, i.QuarantineRemark, i.Remark, i.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Inventory_Lot SET GR_ID=?, Container_Item_ID=?, Lot_No=?, Expiry_Date=?, Qty=?, Landed_Cost_Per_Unit=?, Quarantine_Status=?, Quarantine_Remark=?, Remark=? WHERE Lot_ID=?",
		i.GRID, i.ContainerItemID, i.LotNo, nullIfZero(i.ExpiryDate), i.Qty, i.LandedCostPerUnit, i.QuarantineStatus, i.QuarantineRemark, i.Remark, i.ID)
	return err
}

// Cost Allocation
func (r *SQLiteProcurementRepository) GetCostAllocations() ([]models.CostAllocation, error) {
	rows, err := r.db.Query("SELECT Cost_Allocation_ID, Allocation_date, IFNULL(Total_Allocated_Amount, 0), IFNULL(Is_Late_Cost_Allocation, 0), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Cost_Allocation")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.CostAllocation{}
	for rows.Next() {
		var i models.CostAllocation
		var late int
		rows.Scan(&i.ID, &i.AllocationDate, &i.TotalAllocatedAmount, &late, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
		i.IsLateCostAllocation = late == 1
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveCostAllocation(ca *models.CostAllocation) error {
	late := 0
	if ca.IsLateCostAllocation {
		late = 1
	}
	if ca.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Cost_Allocation (Allocation_date, Total_Allocated_Amount, Is_Late_Cost_Allocation, Remark, Created_By) VALUES (?, ?, ?, ?, ?)",
			ca.AllocationDate, ca.TotalAllocatedAmount, late, ca.Remark, ca.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		ca.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Cost_Allocation SET Allocation_date=?, Total_Allocated_Amount=?, Is_Late_Cost_Allocation=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE Cost_Allocation_ID=?",
		ca.AllocationDate, ca.TotalAllocatedAmount, late, ca.Remark, ca.UpdatedBy, ca.ID)
	return err
}

// Container Items
func (r *SQLiteProcurementRepository) GetContainerItemsByContainerID(containerID int) ([]models.ContainerItem, error) {
	rows, err := r.db.Query(`
		SELECT
			ci.Container_Item_ID, ci.PO_Item_ID, IFNULL(ci.Container_ID, 0), IFNULL(ci.CI_ID, 0), IFNULL(ci.BL_ID, 0), IFNULL(ci.Constraint_Profile_ID, 0),
			IFNULL(ci.Load_Qty, 0),
			IFNULL(ci.Gross_Weight, 0), IFNULL(ci.Net_Weight, 0), IFNULL(ci.Cbm, 0), ci.Temporary_ETA, IFNULL(ci.UUID, ''), IFNULL(ci.Remark, ''),
			b.ETD, b.ETA,
			IFNULL((SELECT SUM(Qty) FROM Inventory_Lot l WHERE l.Container_Item_ID = ci.Container_Item_ID), 0) as Landed_Qty
		FROM Container_Item ci
		LEFT JOIN BL b ON ci.BL_ID = b.BL_ID
		WHERE ci.Container_ID = ?`, containerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.ContainerItem{}
	now := time.Now()
	for rows.Next() {
		var i models.ContainerItem
		var tempEta, etd, eta *time.Time
		var landedQty float64
		err := rows.Scan(&i.ID, &i.POItemID, &i.ContainerID, &i.CIID, &i.BLID, &i.ConstraintProfileID, &i.LoadQty, &i.GrossWeight, &i.NetWeight, &i.CBM, &tempEta, &i.UUID, &i.Remark, &etd, &eta, &landedQty)
		if err != nil {
			return nil, err
		}
		if tempEta != nil {
			i.TemporaryETA = *tempEta
		}
		var etdT, etaT time.Time
		if etd != nil {
			etdT = *etd
		}
		if eta != nil {
			etaT = *eta
		}
		i.Status = models.ComputeContainerItemStatus(i.BLID > 0, etdT, etaT, i.LoadQty, landedQty, now)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetContainerItemByID(id int) (*models.ContainerItem, error) {
	var i models.ContainerItem
	var tempEta, etd, eta *time.Time
	var landedQty float64
	err := r.db.QueryRow(`
		SELECT
			ci.Container_Item_ID, ci.PO_Item_ID, IFNULL(ci.Container_ID, 0), IFNULL(ci.CI_ID, 0), IFNULL(ci.BL_ID, 0), IFNULL(ci.Constraint_Profile_ID, 0),
			IFNULL(ci.Load_Qty, 0),
			IFNULL(ci.Gross_Weight, 0), IFNULL(ci.Net_Weight, 0), IFNULL(ci.Cbm, 0), ci.Temporary_ETA, IFNULL(ci.UUID, ''), IFNULL(ci.Remark, ''),
			b.ETD, b.ETA,
			IFNULL((SELECT SUM(Qty) FROM Inventory_Lot l WHERE l.Container_Item_ID = ci.Container_Item_ID), 0) as Landed_Qty
		FROM Container_Item ci
		LEFT JOIN BL b ON ci.BL_ID = b.BL_ID
		WHERE ci.Container_Item_ID = ?`, id).
		Scan(&i.ID, &i.POItemID, &i.ContainerID, &i.CIID, &i.BLID, &i.ConstraintProfileID, &i.LoadQty, &i.GrossWeight, &i.NetWeight, &i.CBM, &tempEta, &i.UUID, &i.Remark, &etd, &eta, &landedQty)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if tempEta != nil {
		i.TemporaryETA = *tempEta
	}
	var etdT, etaT time.Time
	if etd != nil {
		etdT = *etd
	}
	if eta != nil {
		etaT = *eta
	}
	i.Status = models.ComputeContainerItemStatus(i.BLID > 0, etdT, etaT, i.LoadQty, landedQty, time.Now())
	return &i, nil
}

func (r *SQLiteProcurementRepository) SaveContainerItem(i *models.ContainerItem) error {
	var err error
	if i.ID == 0 {
		_, err = r.db.Exec(`INSERT INTO Container_Item (PO_Item_ID, Container_ID, CI_ID, BL_ID, Constraint_Profile_ID, Load_Qty, Gross_Weight, Net_Weight, Cbm, Temporary_ETA, UUID, Remark) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			i.POItemID, i.ContainerID, i.CIID, i.BLID, i.ConstraintProfileID, i.LoadQty, i.GrossWeight, i.NetWeight, i.CBM, nullIfZero(i.TemporaryETA), i.UUID, i.Remark)
	} else {
		_, err = r.db.Exec(`UPDATE Container_Item SET PO_Item_ID=?, Container_ID=?, CI_ID=?, BL_ID=?, Constraint_Profile_ID=?, Load_Qty=?, Gross_Weight=?, Net_Weight=?, Cbm=?, Temporary_ETA=?, UUID=?, Remark=? WHERE Container_Item_ID=?`,
			i.POItemID, i.ContainerID, i.CIID, i.BLID, i.ConstraintProfileID, i.LoadQty, i.GrossWeight, i.NetWeight, i.CBM, nullIfZero(i.TemporaryETA), i.UUID, i.Remark, i.ID)
	}
	return err
}

// Cost Allocation Items
func (r *SQLiteProcurementRepository) GetCostAllocationItemsByAllocationID(caID int) ([]models.CostAllocationItem, error) {
	rows, err := r.db.Query("SELECT Cost_Allocation_Item_ID, Cost_Allocation_ID, Lot_ID, IFNULL(Allocated_Amount, 0), AP_ID FROM Cost_Allocation_Item WHERE Cost_Allocation_ID = ?", caID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.CostAllocationItem{}
	for rows.Next() {
		var i models.CostAllocationItem
		rows.Scan(&i.ID, &i.CostAllocationID, &i.LotID, &i.AllocatedAmount, &i.APID)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveCostAllocationItem(i *models.CostAllocationItem) error {
	var apID interface{}
	if i.APID > 0 {
		apID = i.APID
	} else {
		apID = nil
	}
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Cost_Allocation_Item (Cost_Allocation_ID, Lot_ID, Allocated_Amount, AP_ID) VALUES (?, ?, ?, ?)",
			i.CostAllocationID, i.LotID, i.AllocatedAmount, apID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Cost_Allocation_Item SET Cost_Allocation_ID=?, Lot_ID=?, Allocated_Amount=?, AP_ID=? WHERE Cost_Allocation_Item_ID=?",
		i.CostAllocationID, i.LotID, i.AllocatedAmount, apID, i.ID)
	return err
}

func (r *SQLiteProcurementRepository) DeleteCostAllocationItem(id int) error {
	_, err := r.db.Exec("DELETE FROM Cost_Allocation_Item WHERE Cost_Allocation_Item_ID = ?", id)
	return err
}

func (r *SQLiteProcurementRepository) GetAvailableCostAllocationLots() ([]models.CostAllocationLotCandidate, error) {
	query := `
		SELECT 
			l.Lot_ID,
			IFNULL(l.Lot_No, ''),
			IFNULL(l.UUID, ''),
			IFNULL(l.GR_ID, 0),
			IFNULL(gr.UUID, ''),
			IFNULL(l.Container_Item_ID, 0),
			IFNULL(ci.UUID, ''),
			IFNULL(ci.CI_ID, 0),
			IFNULL(ci_tbl.UUID, ''),
			IFNULL(ci.Container_ID, 0),
			IFNULL(c.UUID, ''),
			IFNULL(c.Container_No, ''),
			IFNULL(ci.BL_ID, 0),
			IFNULL(b.UUID, ''),
			IFNULL(b.BL_No, ''),
			IFNULL(pi.PO_ID, 0),
			IFNULL(p.UUID, ''),
			IFNULL(p.PO_No, ''),
			IFNULL(pi.Item_ID, 0),
			IFNULL(im.Name, ''),
			IFNULL(l.Qty, 0),
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Gross_Weight, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Net_Weight, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Cbm, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
			IFNULL(pi.Unit_Price, 0),
			IFNULL(p.Currency, '')
		FROM Inventory_Lot l
		JOIN Container_Item ci ON l.Container_Item_ID = ci.Container_Item_ID
		JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
		JOIN Purchase_Order p ON pi.PO_ID = p.PO_ID
		LEFT JOIN Goods_Receipt gr ON l.GR_ID = gr.GR_ID
		LEFT JOIN Commercial_Invoice ci_tbl ON ci.CI_ID = ci_tbl.CI_ID
		LEFT JOIN Container c ON ci.Container_ID = c.Container_ID
		LEFT JOIN BL b ON ci.BL_ID = b.BL_ID
		LEFT JOIN Item_Master im ON pi.Item_ID = im.Item_ID
		LEFT JOIN Cost_Allocation_Item cai ON l.Lot_ID = cai.Lot_ID
		GROUP BY l.Lot_ID
		HAVING COUNT(cai.Cost_Allocation_Item_ID) = 0
		ORDER BY l.Lot_ID DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.CostAllocationLotCandidate{}
	for rows.Next() {
		var c models.CostAllocationLotCandidate
		err := rows.Scan(&c.LotID, &c.LotNo, &c.LotUUID, &c.GRID, &c.GRUUID, &c.ContainerItemID, &c.ContainerItemUUID,
			&c.CIID, &c.CIUUID, &c.ContainerID, &c.ContainerUUID, &c.ContainerNo,
			&c.BLID, &c.BLUUID, &c.BLNo, &c.POID, &c.POUUID, &c.PONo,
			&c.ItemID, &c.ItemName, &c.Qty,
			&c.GrossWeight, &c.NetWeight, &c.CBM, &c.UnitPrice, &c.Currency)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetCostAllocationLotCandidatesByIDs(ids []int) ([]models.CostAllocationLotCandidate, error) {
	if len(ids) == 0 {
		return []models.CostAllocationLotCandidate{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf(`
		SELECT 
			l.Lot_ID,
			IFNULL(l.Lot_No, ''),
			IFNULL(l.UUID, ''),
			IFNULL(l.GR_ID, 0),
			IFNULL(gr.UUID, ''),
			IFNULL(l.Container_Item_ID, 0),
			IFNULL(ci.UUID, ''),
			IFNULL(ci.CI_ID, 0),
			IFNULL(ci_tbl.UUID, ''),
			IFNULL(ci.Container_ID, 0),
			IFNULL(c.UUID, ''),
			IFNULL(c.Container_No, ''),
			IFNULL(ci.BL_ID, 0),
			IFNULL(b.UUID, ''),
			IFNULL(b.BL_No, ''),
			IFNULL(pi.PO_ID, 0),
			IFNULL(p.UUID, ''),
			IFNULL(p.PO_No, ''),
			IFNULL(pi.Item_ID, 0),
			IFNULL(im.Name, ''),
			IFNULL(l.Qty, 0),
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Gross_Weight, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Net_Weight, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Cbm, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
			IFNULL(pi.Unit_Price, 0),
			IFNULL(p.Currency, '')
		FROM Inventory_Lot l
		JOIN Container_Item ci ON l.Container_Item_ID = ci.Container_Item_ID
		JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
		JOIN Purchase_Order p ON pi.PO_ID = p.PO_ID
		LEFT JOIN Goods_Receipt gr ON l.GR_ID = gr.GR_ID
		LEFT JOIN Commercial_Invoice ci_tbl ON ci.CI_ID = ci_tbl.CI_ID
		LEFT JOIN Container c ON ci.Container_ID = c.Container_ID
		LEFT JOIN BL b ON ci.BL_ID = b.BL_ID
		LEFT JOIN Item_Master im ON pi.Item_ID = im.Item_ID
		WHERE l.Lot_ID IN (%s)
	`, strings.Join(placeholders, ","))
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.CostAllocationLotCandidate{}
	for rows.Next() {
		var c models.CostAllocationLotCandidate
		err := rows.Scan(&c.LotID, &c.LotNo, &c.LotUUID, &c.GRID, &c.GRUUID, &c.ContainerItemID, &c.ContainerItemUUID,
			&c.CIID, &c.CIUUID, &c.ContainerID, &c.ContainerUUID, &c.ContainerNo,
			&c.BLID, &c.BLUUID, &c.BLNo, &c.POID, &c.POUUID, &c.PONo,
			&c.ItemID, &c.ItemName, &c.Qty,
			&c.GrossWeight, &c.NetWeight, &c.CBM, &c.UnitPrice, &c.Currency)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetAllCostAllocationLotCandidates() ([]models.CostAllocationLotCandidate, error) {
	query := `
		SELECT 
			l.Lot_ID,
			IFNULL(l.Lot_No, ''),
			IFNULL(l.UUID, ''),
			IFNULL(l.GR_ID, 0),
			IFNULL(gr.UUID, ''),
			IFNULL(l.Container_Item_ID, 0),
			IFNULL(ci.UUID, ''),
			IFNULL(ci.CI_ID, 0),
			IFNULL(ci_tbl.UUID, ''),
			IFNULL(ci.Container_ID, 0),
			IFNULL(c.UUID, ''),
			IFNULL(c.Container_No, ''),
			IFNULL(ci.BL_ID, 0),
			IFNULL(b.UUID, ''),
			IFNULL(b.BL_No, ''),
			IFNULL(pi.PO_ID, 0),
			IFNULL(p.UUID, ''),
			IFNULL(p.PO_No, ''),
			IFNULL(pi.Item_ID, 0),
			IFNULL(im.Name, ''),
			IFNULL(l.Qty, 0),
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Gross_Weight, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Net_Weight, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Cbm, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
			IFNULL(pi.Unit_Price, 0),
			IFNULL(p.Currency, '')
		FROM Inventory_Lot l
		JOIN Container_Item ci ON l.Container_Item_ID = ci.Container_Item_ID
		JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
		JOIN Purchase_Order p ON pi.PO_ID = p.PO_ID
		LEFT JOIN Goods_Receipt gr ON l.GR_ID = gr.GR_ID
		LEFT JOIN Commercial_Invoice ci_tbl ON ci.CI_ID = ci_tbl.CI_ID
		LEFT JOIN Container c ON ci.Container_ID = c.Container_ID
		LEFT JOIN BL b ON ci.BL_ID = b.BL_ID
		LEFT JOIN Item_Master im ON pi.Item_ID = im.Item_ID
		ORDER BY l.Lot_ID DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.CostAllocationLotCandidate{}
	for rows.Next() {
		var c models.CostAllocationLotCandidate
		err := rows.Scan(&c.LotID, &c.LotNo, &c.LotUUID, &c.GRID, &c.GRUUID, &c.ContainerItemID, &c.ContainerItemUUID,
			&c.CIID, &c.CIUUID, &c.ContainerID, &c.ContainerUUID, &c.ContainerNo,
			&c.BLID, &c.BLUUID, &c.BLNo, &c.POID, &c.POUUID, &c.PONo,
			&c.ItemID, &c.ItemName, &c.Qty,
			&c.GrossWeight, &c.NetWeight, &c.CBM, &c.UnitPrice, &c.Currency)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetAPIDsUsedForLots(lotIDs []int) (map[int]bool, error) {
	if len(lotIDs) == 0 {
		return make(map[int]bool), nil
	}
	placeholders := make([]string, len(lotIDs))
	args := make([]interface{}, len(lotIDs))
	for i, id := range lotIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf("SELECT DISTINCT AP_ID FROM Cost_Allocation_Item WHERE Lot_ID IN (%s) AND AP_ID IS NOT NULL AND AP_ID > 0", strings.Join(placeholders, ","))
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, nil
}

func (r *SQLiteProcurementRepository) GetCostAllocationAmountsByAP(apID int) (map[int]float64, error) {
	rows, err := r.db.Query("SELECT Lot_ID, IFNULL(Allocated_Amount, 0) FROM Cost_Allocation_Item WHERE AP_ID = ?", apID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int]float64)
	for rows.Next() {
		var lotID int
		var amount float64
		if err := rows.Scan(&lotID, &amount); err != nil {
			return nil, err
		}
		m[lotID] += amount
	}
	return m, nil
}

func (r *SQLiteProcurementRepository) GetCostAllocationWarningByPO(poID int) (*models.POAllocationWarning, error) {
	query := `
		SELECT COUNT(DISTINCT cai.Lot_ID), IFNULL(SUM(cai.Allocated_Amount), 0)
		FROM Cost_Allocation_Item cai
		JOIN Inventory_Lot l ON cai.Lot_ID = l.Lot_ID
		JOIN Container_Item ci ON l.Container_Item_ID = ci.Container_Item_ID
		JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
		WHERE pi.PO_ID = ?
	`
	var lotCount int
	var totalAmount float64
	if err := r.db.QueryRow(query, poID).Scan(&lotCount, &totalAmount); err != nil {
		return nil, err
	}
	return &models.POAllocationWarning{
		HasAllocation:        lotCount > 0,
		AllocatedLotCount:    lotCount,
		TotalAllocatedAmount: totalAmount,
	}, nil
}

func (r *SQLiteProcurementRepository) GetReferenceIDByUUID(refType string, uuid string) (int, error) {
	var table, col string
	switch refType {
	case "PO":
		table, col = "Purchase_Order", "PO_ID"
	case "CI":
		table, col = "Commercial_Invoice", "CI_ID"
	case "BL":
		table, col = "BL", "BL_ID"
	case "Container":
		table, col = "Container", "Container_ID"
	case "ContainerItem", "Container Item":
		table, col = "Container_Item", "Container_Item_ID"
	case "GR":
		table, col = "Goods_Receipt", "GR_ID"
	case "Lot":
		table, col = "Inventory_Lot", "Lot_ID"
	default:
		return 0, nil
	}
	var id int
	err := r.db.QueryRow(fmt.Sprintf("SELECT %s FROM %s WHERE UUID = ?", col, table), uuid).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetReferenceNoByUUID 참조 문서의 표시 번호(GR No, CI No 등)를 UUID로 조회
func (r *SQLiteProcurementRepository) GetReferenceNoByUUID(refType string, uuid string) (string, error) {
	var query string
	switch refType {
	case "PO":
		query = "SELECT IFNULL(PO_No, '') FROM Purchase_Order WHERE UUID = ?"
	case "CI":
		query = "SELECT IFNULL(CI_No, '') FROM Commercial_Invoice WHERE UUID = ?"
	case "BL":
		query = "SELECT IFNULL(BL_No, '') FROM BL WHERE UUID = ?"
	case "Container":
		query = "SELECT IFNULL(Container_No, '') FROM Container WHERE UUID = ?"
	case "ContainerItem", "Container Item":
		// Container_Item에는 사람이 읽는 번호가 없음
		return "", nil
	case "GR":
		query = "SELECT IFNULL(GR_No, '') FROM Goods_Receipt WHERE UUID = ?"
	case "Lot":
		query = "SELECT IFNULL(Lot_No, '') FROM Inventory_Lot WHERE UUID = ?"
	case "AP_Target_Group":
		query = "SELECT IFNULL(Group_No, '') FROM AP_Target_Group WHERE UUID = ?"
	default:
		return "", nil
	}
	var no string
	err := r.db.QueryRow(query, uuid).Scan(&no)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return no, nil
}

func (r *SQLiteProcurementRepository) GetCostAllocationReferenceBaseUnits(refType string, refID int) ([]models.CostAllocationBaseUnit, error) {
	var query string
	switch refType {
	case "PO":
		query = `
			SELECT IFNULL(PO_Qty, 0), IFNULL(Unit_Price, 0), 0, 0
			FROM PO_Item
			WHERE PO_ID = ? AND IFNULL(PO_Qty, 0) > 0
		`
	case "CI":
		query = `
			SELECT IFNULL(ci.Load_Qty, 0), IFNULL(pi.Unit_Price, 0), IFNULL(ci.Gross_Weight, 0), IFNULL(ci.Cbm, 0)
			FROM Container_Item ci
			JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
			WHERE ci.CI_ID = ? AND IFNULL(ci.Load_Qty, 0) > 0
		`
	case "BL":
		query = `
			SELECT IFNULL(ci.Load_Qty, 0), IFNULL(pi.Unit_Price, 0), IFNULL(ci.Gross_Weight, 0), IFNULL(ci.Cbm, 0)
			FROM Container_Item ci
			JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
			WHERE ci.BL_ID = ? AND IFNULL(ci.Load_Qty, 0) > 0
		`
	case "Container":
		query = `
			SELECT IFNULL(ci.Load_Qty, 0), IFNULL(pi.Unit_Price, 0), IFNULL(ci.Gross_Weight, 0), IFNULL(ci.Cbm, 0)
			FROM Container_Item ci
			JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
			WHERE ci.Container_ID = ? AND IFNULL(ci.Load_Qty, 0) > 0
		`
	case "ContainerItem", "Container Item":
		query = `
			SELECT IFNULL(ci.Load_Qty, 0), IFNULL(pi.Unit_Price, 0), IFNULL(ci.Gross_Weight, 0), IFNULL(ci.Cbm, 0)
			FROM Container_Item ci
			JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
			WHERE ci.Container_Item_ID = ? AND IFNULL(ci.Load_Qty, 0) > 0
		`
	case "GR":
		query = `
			SELECT IFNULL(l.Qty, 0), IFNULL(pi.Unit_Price, 0),
				CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Gross_Weight, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
				CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Cbm, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END
			FROM Inventory_Lot l
			JOIN Container_Item ci ON l.Container_Item_ID = ci.Container_Item_ID
			JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
			WHERE l.GR_ID = ? AND IFNULL(l.Qty, 0) > 0
		`
	case "Lot":
		query = `
			SELECT IFNULL(l.Qty, 0), IFNULL(pi.Unit_Price, 0),
				CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Gross_Weight, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END,
				CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Cbm, 0) * IFNULL(l.Qty, 0) / ci.Load_Qty ELSE 0 END
			FROM Inventory_Lot l
			JOIN Container_Item ci ON l.Container_Item_ID = ci.Container_Item_ID
			JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
			WHERE l.Lot_ID = ? AND IFNULL(l.Qty, 0) > 0
		`
	default:
		return []models.CostAllocationBaseUnit{}, nil
	}

	rows, err := r.db.Query(query, refID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.CostAllocationBaseUnit{}
	for rows.Next() {
		var u models.CostAllocationBaseUnit
		if err := rows.Scan(&u.Qty, &u.UnitPrice, &u.GrossWeight, &u.CBM); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetContainersByBLID(blID int) ([]models.Container, error) {
	query := `SELECT DISTINCT c.Container_ID, c.Container_No, c.Remark, c.Total_CBM, c.Total_Net_Wgt, c.Total_Gross_Wgt, c.Status, c.UUID
	          FROM Container c
	          JOIN Container_Item ci ON c.Container_ID = ci.Container_ID
	          WHERE ci.BL_ID = ?`
	rows, err := r.db.Query(query, blID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.Container{}
	for rows.Next() {
		var c models.Container
		if err := rows.Scan(&c.ID, &c.ContainerNo, &c.Remark, &c.TotalCBM, &c.TotalNetWgt, &c.TotalGrossWgt, &c.Status, &c.UUID); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

// GetUnbookedPOItems 미예약 PO 항목 조회 (Open 상태 PO의 남은 수량)
func (r *SQLiteProcurementRepository) GetUnbookedPOItems() ([]models.BookingTemplateRow, error) {
	query := `
		SELECT 
			pi.PO_Item_ID,
			p.PO_No,
			im.SKU_Code,
			im.Name,
			vm.Name,
			pi.PO_Qty,
			pi.PO_Qty - IFNULL(SUM(ci.Load_Qty), 0) as Remaining_Qty
		FROM PO_Item pi
		JOIN Purchase_Order p ON pi.PO_ID = p.PO_ID
		JOIN Item_Master im ON pi.Item_ID = im.Item_ID
		JOIN Vendor_Master vm ON p.Vendor_ID = vm.Vendor_ID
		LEFT JOIN Container_Item ci ON pi.PO_Item_ID = ci.PO_Item_ID
		WHERE p.Status = 'Open'
		GROUP BY pi.PO_Item_ID, p.PO_No, im.SKU_Code, im.Name, vm.Name, pi.PO_Qty
		HAVING Remaining_Qty > 0
		ORDER BY p.PO_No, pi.PO_Item_ID
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.BookingTemplateRow{}
	for rows.Next() {
		var row models.BookingTemplateRow
		err := rows.Scan(&row.POItemID, &row.PONo, &row.SKUCode, &row.ItemName, &row.VendorName,
			&row.OrderedQty, &row.RemainingQty)
		if err != nil {
			return nil, err
		}
		// 기본값: Load Qty = Remaining Qty
		row.LoadQty = row.RemainingQty
		list = append(list, row)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetBookings() ([]models.BookingView, error) {
	query := `
		SELECT 
			ci.Container_Item_ID,
			IFNULL(ci.UUID, ''),
			IFNULL(ci.Container_ID, 0),
			IFNULL(c.Container_No, ''),
			IFNULL(c.Total_CBM, 0),
			IFNULL(c.Total_Net_Wgt, 0),
			IFNULL(c.Total_Gross_Wgt, 0),
			IFNULL(ci.BL_ID, 0),
			IFNULL(b.BL_No, ''),
			b.ETD,
			b.ETA,
			IFNULL(b.POL, ''),
			IFNULL(b.POD, ''),
			IFNULL(b.Carrier, ''),
			IFNULL(b.Vessel_Name, ''),
			IFNULL(ci.PO_Item_ID, 0),
			IFNULL(pi.PO_ID, 0),
			IFNULL(p.PO_No, ''),
			IFNULL(pi.Item_ID, 0),
			IFNULL(im.Name, '') as Item_Name,
			IFNULL(ci.CI_ID, 0),
			IFNULL(ci_tbl.CI_No, ''),
			IFNULL(ci.Load_Qty, 0),
			IFNULL(ci.Gross_Weight, 0),
			IFNULL(ci.Net_Weight, 0),
			IFNULL(ci.Cbm, 0),
			CASE WHEN IFNULL(ci.Load_Qty, 0) > 0 THEN IFNULL(ci.Cbm, 0) / ci.Load_Qty ELSE 0 END as Unit_CBM,
			IFNULL(im.CBM, 0) as Master_CBM,
			IFNULL(ci.Constraint_Profile_ID, 0),
			ci.Temporary_ETA,
			IFNULL(ci.Remark, ''),
			IFNULL((SELECT SUM(Qty) FROM Inventory_Lot l WHERE l.Container_Item_ID = ci.Container_Item_ID), 0) as Landed_Qty
		FROM Container_Item ci
		LEFT JOIN Container c ON ci.Container_ID = c.Container_ID
		LEFT JOIN BL b ON ci.BL_ID = b.BL_ID
		JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
		JOIN Purchase_Order p ON pi.PO_ID = p.PO_ID
		LEFT JOIN Commercial_Invoice ci_tbl ON ci.CI_ID = ci_tbl.CI_ID
		LEFT JOIN Item_Master im ON pi.Item_ID = im.Item_ID
		ORDER BY IFNULL(b.ETA, ci.Temporary_ETA) ASC, c.Container_No ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.BookingView{}
	for rows.Next() {
		var b models.BookingView
		var etd, eta *time.Time
		var tempEta *time.Time
		var landedQty float64
		err := rows.Scan(
			&b.ContainerItemID, &b.UUID, &b.ContainerID, &b.ContainerNo,
			&b.TotalCBM, &b.TotalNetWgt, &b.TotalGrossWgt,
			&b.BLID, &b.BLNo, &etd, &eta,
			&b.POL, &b.POD, &b.Carrier, &b.VesselName,
			&b.POItemID, &b.POID, &b.PONo, &b.ItemID, &b.ItemName, &b.CIID, &b.CINo, &b.LoadQty,
			&b.GrossWeight, &b.NetWeight, &b.CBM, &b.UnitCBM, &b.MasterCBM, &b.ConstraintProfileID, &tempEta, &b.Remark, &landedQty,
		)
		if err != nil {
			return nil, err
		}
		if tempEta != nil {
			b.TemporaryETA = *tempEta
		}
		if etd != nil {
			b.ETD = *etd
		}
		if eta != nil && !eta.IsZero() {
			b.ETA = *eta
		} else if tempEta != nil {
			b.ETA = *tempEta
		}
		b.Status = models.ComputeContainerItemStatus(b.BLID > 0, b.ETD, b.ETA, b.LoadQty, landedQty, time.Now())
		list = append(list, b)
	}
	return list, nil
}

// AP Target Group
func (r *SQLiteProcurementRepository) GetAPTargetGroups() ([]models.AP_TargetGroup, error) {
	rows, err := r.db.Query(`
		SELECT Group_ID, Group_No, IFNULL(Group_Name, ''), IFNULL(Reference_Type, ''), IFNULL(Status, 'Draft'), IFNULL(Remark, ''),
		       IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID
		FROM AP_Target_Group
		ORDER BY Created_At DESC`)
	if err != nil {
		return nil, err
	}

	list := []models.AP_TargetGroup{}
	for rows.Next() {
		var g models.AP_TargetGroup
		err := rows.Scan(&g.ID, &g.GroupNo, &g.GroupName, &g.ReferenceType, &g.Status, &g.Remark,
			&g.CreatedBy, &g.CreatedAt, &g.UpdatedBy, &g.UpdatedAt, &g.UUID)
		if err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, g)
	}
	rows.Close()

	// 각 그룹의 참조 항목(items)을 함께 로드
	for i := range list {
		items, _ := r.GetAPTargetGroupItems(list[i].ID)
		list[i].Items = items
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetAPTargetGroupByID(id int) (*models.AP_TargetGroup, error) {
	var g models.AP_TargetGroup
	err := r.db.QueryRow(`
		SELECT Group_ID, Group_No, IFNULL(Group_Name, ''), IFNULL(Reference_Type, ''), IFNULL(Status, 'Draft'), IFNULL(Remark, ''),
		       IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID
		FROM AP_Target_Group WHERE Group_ID = ?`, id).
		Scan(&g.ID, &g.GroupNo, &g.GroupName, &g.ReferenceType, &g.Status, &g.Remark,
			&g.CreatedBy, &g.CreatedAt, &g.UpdatedBy, &g.UpdatedAt, &g.UUID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	items, _ := r.GetAPTargetGroupItems(id)
	g.Items = items
	return &g, nil
}

func (r *SQLiteProcurementRepository) GetAPTargetGroupItems(groupID int) ([]models.AP_TargetGroupItem, error) {
	rows, err := r.db.Query(`
		SELECT Group_Item_ID, Group_ID, IFNULL(Reference_UUID, ''), IFNULL(Remark, '')
		FROM AP_Target_Group_Item WHERE Group_ID = ?`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.AP_TargetGroupItem{}
	for rows.Next() {
		var i models.AP_TargetGroupItem
		err := rows.Scan(&i.ID, &i.GroupID, &i.ReferenceUUID, &i.Remark)
		if err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveAPTargetGroup(g *models.AP_TargetGroup) error {
	if g.ID == 0 {
		res, err := r.db.Exec(`
			INSERT INTO AP_Target_Group (Group_No, Group_Name, Reference_Type, Status, Remark, Created_By, UUID)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			g.GroupNo, g.GroupName, g.ReferenceType, g.Status, g.Remark, g.CreatedBy, g.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		g.ID = int(id)
		return nil
	}
	_, err := r.db.Exec(`
		UPDATE AP_Target_Group SET Group_No=?, Group_Name=?, Reference_Type=?, Status=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP
		WHERE Group_ID=?`,
		g.GroupNo, g.GroupName, g.ReferenceType, g.Status, g.Remark, g.UpdatedBy, g.ID)
	return err
}

func (r *SQLiteProcurementRepository) SaveAPTargetGroupItem(item *models.AP_TargetGroupItem) error {
	if item.ID == 0 {
		res, err := r.db.Exec(`
			INSERT INTO AP_Target_Group_Item (Group_ID, Reference_UUID, Remark)
			VALUES (?, ?, ?)`,
			item.GroupID, item.ReferenceUUID, item.Remark)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		item.ID = int(id)
		return nil
	}
	_, err := r.db.Exec(`
		UPDATE AP_Target_Group_Item SET Group_ID=?, Reference_UUID=?, Remark=?
		WHERE Group_Item_ID=?`,
		item.GroupID, item.ReferenceUUID, item.Remark, item.ID)
	return err
}

func (r *SQLiteProcurementRepository) DeleteAPTargetGroupItem(id int) error {
	_, err := r.db.Exec("DELETE FROM AP_Target_Group_Item WHERE Group_Item_ID = ?", id)
	return err
}

// Constraint Profile (행정 제약 프로파일)
func (r *SQLiteProcurementRepository) GetConstraintProfiles() ([]models.ConstraintProfile, error) {
	rows, err := r.db.Query(`
		SELECT Profile_ID, Profile_No, IFNULL(Name, ''), IFNULL(Remark, ''),
		       IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID
		FROM Constraint_Profile ORDER BY Created_At DESC`)
	if err != nil {
		return nil, err
	}
	list := []models.ConstraintProfile{}
	for rows.Next() {
		var p models.ConstraintProfile
		if err := rows.Scan(&p.ID, &p.ProfileNo, &p.Name, &p.Remark,
			&p.CreatedBy, &p.CreatedAt, &p.UpdatedBy, &p.UpdatedAt, &p.UUID); err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, p)
	}
	rows.Close()
	for i := range list {
		steps, _ := r.GetConstraintProfileSteps(list[i].ID)
		list[i].Steps = steps
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetConstraintProfileByID(id int) (*models.ConstraintProfile, error) {
	var p models.ConstraintProfile
	err := r.db.QueryRow(`
		SELECT Profile_ID, Profile_No, IFNULL(Name, ''), IFNULL(Remark, ''),
		       IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID
		FROM Constraint_Profile WHERE Profile_ID = ?`, id).
		Scan(&p.ID, &p.ProfileNo, &p.Name, &p.Remark,
			&p.CreatedBy, &p.CreatedAt, &p.UpdatedBy, &p.UpdatedAt, &p.UUID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	steps, _ := r.GetConstraintProfileSteps(id)
	p.Steps = steps
	return &p, nil
}

func (r *SQLiteProcurementRepository) GetConstraintProfileSteps(profileID int) ([]models.ConstraintProfileStep, error) {
	rows, err := r.db.Query(`
		SELECT Step_ID, Profile_ID, IFNULL(Constraint_ID, 0), IFNULL(Sequence, 0), IFNULL(Constraint_Type, ''),
		       IFNULL(Lead_Time_Days, 0), IFNULL(Is_Optional, 0), IFNULL(Remark, '')
		FROM Constraint_Profile_Step WHERE Profile_ID = ? ORDER BY Sequence, Step_ID`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.ConstraintProfileStep{}
	for rows.Next() {
		var s models.ConstraintProfileStep
		var opt int
		if err := rows.Scan(&s.ID, &s.ProfileID, &s.ConstraintID, &s.Sequence, &s.ConstraintType, &s.LeadTimeDays, &opt, &s.Remark); err != nil {
			return nil, err
		}
		s.IsOptional = opt != 0
		list = append(list, s)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveConstraintProfile(p *models.ConstraintProfile) error {
	if p.ID == 0 {
		res, err := r.db.Exec(`
			INSERT INTO Constraint_Profile (Profile_No, Name, Remark, Created_By, UUID)
			VALUES (?, ?, ?, ?, ?)`,
			p.ProfileNo, p.Name, p.Remark, p.CreatedBy, p.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		p.ID = int(id)
		return nil
	}
	_, err := r.db.Exec(`
		UPDATE Constraint_Profile SET Profile_No=?, Name=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP
		WHERE Profile_ID=?`,
		p.ProfileNo, p.Name, p.Remark, p.UpdatedBy, p.ID)
	return err
}

func (r *SQLiteProcurementRepository) SaveConstraintProfileStep(s *models.ConstraintProfileStep) error {
	opt := 0
	if s.IsOptional {
		opt = 1
	}
	if s.ID == 0 {
		res, err := r.db.Exec(`
			INSERT INTO Constraint_Profile_Step (Profile_ID, Constraint_ID, Sequence, Constraint_Type, Lead_Time_Days, Is_Optional, Remark)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			s.ProfileID, s.ConstraintID, s.Sequence, s.ConstraintType, s.LeadTimeDays, opt, s.Remark)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		s.ID = int(id)
		return nil
	}
	_, err := r.db.Exec(`
		UPDATE Constraint_Profile_Step SET Profile_ID=?, Constraint_ID=?, Sequence=?, Constraint_Type=?, Lead_Time_Days=?, Is_Optional=?, Remark=?
		WHERE Step_ID=?`,
		s.ProfileID, s.ConstraintID, s.Sequence, s.ConstraintType, s.LeadTimeDays, opt, s.Remark, s.ID)
	return err
}

func (r *SQLiteProcurementRepository) DeleteConstraintProfileStep(id int) error {
	_, err := r.db.Exec("DELETE FROM Constraint_Profile_Step WHERE Step_ID = ?", id)
	return err
}

// Constraint Master (개별 제약 카탈로그)
func (r *SQLiteProcurementRepository) GetConstraints() ([]models.ConstraintMaster, error) {
	rows, err := r.db.Query(`SELECT Constraint_ID, IFNULL(Name, ''), IFNULL(Lead_Time_Days, 0), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Constraint_Master ORDER BY Name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.ConstraintMaster{}
	for rows.Next() {
		var c models.ConstraintMaster
		if err := rows.Scan(&c.ID, &c.Name, &c.LeadTimeDays, &c.Remark, &c.CreatedBy, &c.CreatedAt, &c.UpdatedBy, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetConstraintMasterByID(id int) (*models.ConstraintMaster, error) {
	var c models.ConstraintMaster
	err := r.db.QueryRow(`SELECT Constraint_ID, IFNULL(Name, ''), IFNULL(Lead_Time_Days, 0), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Constraint_Master WHERE Constraint_ID = ?`, id).
		Scan(&c.ID, &c.Name, &c.LeadTimeDays, &c.Remark, &c.CreatedBy, &c.CreatedAt, &c.UpdatedBy, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *SQLiteProcurementRepository) SaveConstraintMaster(c *models.ConstraintMaster) error {
	if c.ID == 0 {
		res, err := r.db.Exec(`INSERT INTO Constraint_Master (Name, Lead_Time_Days, Remark, Created_By) VALUES (?, ?, ?, ?)`,
			c.Name, c.LeadTimeDays, c.Remark, c.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		c.ID = int(id)
		return nil
	}
	_, err := r.db.Exec(`UPDATE Constraint_Master SET Name=?, Lead_Time_Days=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE Constraint_ID=?`,
		c.Name, c.LeadTimeDays, c.Remark, c.UpdatedBy, c.ID)
	return err
}

// Constraint Instance (Lot별 행정 제약 인스턴스)
func (r *SQLiteProcurementRepository) GetConstraintInstances(lotID int) ([]models.ConstraintInstance, error) {
	rows, err := r.db.Query(`
		SELECT Instance_ID, Lot_ID, IFNULL(Source_Container_Item_ID, 0), IFNULL(Profile_ID, 0), IFNULL(Predecessor_ID, 0), IFNULL(Sequence, 0),
		       IFNULL(Constraint_Type, ''), IFNULL(Lead_Time_Days, 0), IFNULL(Is_Optional, 0),
		       Planned_Start, Actual_Start, Planned_End, Actual_End, IFNULL(Status, 'Planned'), IFNULL(Remark, ''),
		       IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At
		FROM Constraint_Instance WHERE Lot_ID = ? ORDER BY Sequence, Instance_ID`, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.ConstraintInstance{}
	for rows.Next() {
		var s models.ConstraintInstance
		var opt int
		var ps, as, pe, ae sql.NullTime
		if err := rows.Scan(&s.ID, &s.LotID, &s.SourceContainerItemID, &s.ProfileID, &s.PredecessorID, &s.Sequence,
			&s.ConstraintType, &s.LeadTimeDays, &opt, &ps, &as, &pe, &ae, &s.Status, &s.Remark,
			&s.CreatedBy, &s.CreatedAt, &s.UpdatedBy, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.IsOptional = opt != 0
		if ps.Valid {
			s.PlannedStart = ps.Time
		}
		if as.Valid {
			s.ActualStart = as.Time
		}
		if pe.Valid {
			s.PlannedEnd = pe.Time
		}
		if ae.Valid {
			s.ActualEnd = ae.Time
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveConstraintInstance(s *models.ConstraintInstance) error {
	opt := 0
	if s.IsOptional {
		opt = 1
	}
	if s.Status == "" {
		s.Status = "Planned"
	}
	if s.ID == 0 {
		res, err := r.db.Exec(`
			INSERT INTO Constraint_Instance
			(Lot_ID, Source_Container_Item_ID, Profile_ID, Predecessor_ID, Sequence, Constraint_Type, Lead_Time_Days, Is_Optional,
			 Planned_Start, Actual_Start, Planned_End, Actual_End, Status, Remark, Created_By)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.LotID, s.SourceContainerItemID, s.ProfileID, s.PredecessorID, s.Sequence, s.ConstraintType, s.LeadTimeDays, opt,
			nullIfZero(s.PlannedStart), nullIfZero(s.ActualStart), nullIfZero(s.PlannedEnd), nullIfZero(s.ActualEnd), s.Status, s.Remark, s.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		s.ID = int(id)
		return nil
	}
	_, err := r.db.Exec(`
		UPDATE Constraint_Instance SET Lot_ID=?, Source_Container_Item_ID=?, Profile_ID=?, Predecessor_ID=?, Sequence=?, Constraint_Type=?, Lead_Time_Days=?, Is_Optional=?,
			Planned_Start=?, Actual_Start=?, Planned_End=?, Actual_End=?, Status=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP
		WHERE Instance_ID=?`,
		s.LotID, s.SourceContainerItemID, s.ProfileID, s.PredecessorID, s.Sequence, s.ConstraintType, s.LeadTimeDays, opt,
		nullIfZero(s.PlannedStart), nullIfZero(s.ActualStart), nullIfZero(s.PlannedEnd), nullIfZero(s.ActualEnd), s.Status, s.Remark, s.UpdatedBy, s.ID)
	return err
}

func (r *SQLiteProcurementRepository) DeleteConstraintInstance(id int) error {
	_, err := r.db.Exec("DELETE FROM Constraint_Instance WHERE Instance_ID = ?", id)
	return err
}

func (r *SQLiteProcurementRepository) GetAllReferenceTargets() ([]models.APTargetGroupReference, error) {
	var list []models.APTargetGroupReference

	// PO
	rows, err := r.db.Query("SELECT PO_ID, PO_No, Vendor_ID, Total_Amount, Currency, UUID FROM Purchase_Order")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var no, currency, uuid string
			var vendorID int
			var amount float64
			if err := rows.Scan(&id, &no, &vendorID, &amount, &currency, &uuid); err == nil {
				list = append(list, models.APTargetGroupReference{
					ReferenceUUID: uuid,
					ReferenceType: "PO",
					ReferenceNo:   no,
					Description:   fmt.Sprintf("Purchase Order #%d", id),
					Amount:        amount,
					Currency:      currency,
				})
			}
		}
	}

	// CI
	rows, err = r.db.Query("SELECT CI_ID, CI_No, Vendor_ID, Total_Amount, Currency, UUID FROM Commercial_Invoice")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var no, currency, uuid string
			var vendorID int
			var amount float64
			if err := rows.Scan(&id, &no, &vendorID, &amount, &currency, &uuid); err == nil {
				list = append(list, models.APTargetGroupReference{
					ReferenceUUID: uuid,
					ReferenceType: "CI",
					ReferenceNo:   no,
					Description:   fmt.Sprintf("Commercial Invoice #%d", id),
					Amount:        amount,
					Currency:      currency,
				})
			}
		}
	}

	// BL
	rows, err = r.db.Query("SELECT BL_ID, BL_No, UUID FROM BL")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var no, uuid string
			if err := rows.Scan(&id, &no, &uuid); err == nil {
				list = append(list, models.APTargetGroupReference{
					ReferenceUUID: uuid,
					ReferenceType: "BL",
					ReferenceNo:   no,
					Description:   fmt.Sprintf("B/L #%d", id),
				})
			}
		}
	}

	// Container
	rows, err = r.db.Query("SELECT Container_ID, Container_No, UUID FROM Container")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var no, uuid string
			if err := rows.Scan(&id, &no, &uuid); err == nil {
				list = append(list, models.APTargetGroupReference{
					ReferenceUUID: uuid,
					ReferenceType: "Container",
					ReferenceNo:   no,
					Description:   fmt.Sprintf("Container #%d", id),
				})
			}
		}
	}

	// Container_Item
	rows, err = r.db.Query(`
		SELECT ci.Container_Item_ID, ci.Load_Qty, IFNULL(pi.Unit_Price, 0), IFNULL(po.Currency, ''), IFNULL(ci.UUID, '')
		FROM Container_Item ci
		JOIN PO_Item pi ON ci.PO_Item_ID = pi.PO_Item_ID
		JOIN Purchase_Order po ON pi.PO_ID = po.PO_ID`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var loadQty, unitPrice float64
			var currency, uuid string
			if err := rows.Scan(&id, &loadQty, &unitPrice, &currency, &uuid); err == nil {
				list = append(list, models.APTargetGroupReference{
					ReferenceUUID: uuid,
					ReferenceType: "Container Item",
					ReferenceNo:   fmt.Sprintf("CI-%d", id),
					Description:   fmt.Sprintf("Container Item #%d", id),
					Amount:        loadQty * unitPrice,
					Currency:      currency,
				})
			}
		}
	}

	// GR
	rows, err = r.db.Query("SELECT GR_ID, IFNULL(GR_No, ''), Receive_Date, Remark, UUID FROM Goods_Receipt")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var grNo, date, remark, uuid string
			if err := rows.Scan(&id, &grNo, &date, &remark, &uuid); err == nil {
				refNo := grNo
				if refNo == "" {
					refNo = fmt.Sprintf("GR-%d", id)
				}
				list = append(list, models.APTargetGroupReference{
					ReferenceUUID: uuid,
					ReferenceType: "GR",
					ReferenceNo:   refNo,
					Description:   fmt.Sprintf("GR #%d (%s)", id, date),
				})
			}
		}
	}

	// Lot
	rows, err = r.db.Query("SELECT Lot_ID, Lot_No, Qty, UUID FROM Inventory_Lot")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var no, uuid string
			var qty float64
			if err := rows.Scan(&id, &no, &qty, &uuid); err == nil {
				list = append(list, models.APTargetGroupReference{
					ReferenceUUID: uuid,
					ReferenceType: "Lot",
					ReferenceNo:   no,
					Description:   fmt.Sprintf("Lot #%d", id),
				})
			}
		}
	}

	return list, nil
}

// CheckReferences 삭제 대상 레코드를 참조하는 다른 테이블 확인
func (r *SQLiteProcurementRepository) CheckReferences(tableName string, id int) (*models.DeleteCheckResult, error) {
	result := &models.DeleteCheckResult{CanDelete: true, References: []models.ReferenceInfo{}}

	type refCheck struct {
		table  string
		column string
		label  string
		query  string
	}

	var checks []refCheck
	var refUUID string

	switch tableName {
	case "Item_Master":
		checks = []refCheck{
			{table: "PO_Item", column: "Item_ID", label: "PO Items",
				query: "SELECT PO_Item_ID, PO_ID, PO_Qty FROM PO_Item WHERE Item_ID = ?"},
		}
	case "Vendor_Master":
		checks = []refCheck{
			{table: "Purchase_Order", column: "Vendor_ID", label: "Purchase Orders",
				query: "SELECT PO_ID, PO_No, Status FROM Purchase_Order WHERE Vendor_ID = ?"},
			{table: "Commercial_Invoice", column: "Vendor_ID", label: "Commercial Invoices",
				query: "SELECT CI_ID, CI_No, Status FROM Commercial_Invoice WHERE Vendor_ID = ?"},
			{table: "Account_Payable", column: "Vendor_ID", label: "Account Payables",
				query: "SELECT AP_ID, AP_No, Status FROM Account_Payable WHERE Vendor_ID = ?"},
		}
	case "Purchase_Order":
		r.db.QueryRow("SELECT UUID FROM Purchase_Order WHERE PO_ID = ?", id).Scan(&refUUID)
		checks = []refCheck{
			{table: "PO_Item", column: "PO_ID", label: "PO Items",
				query: "SELECT PO_Item_ID, Item_ID, PO_Qty FROM PO_Item WHERE PO_ID = ?"},
			{table: "Account_Payable", column: "Reference_UUID", label: "Linked Account Payables",
				query: "SELECT AP_ID, AP_No, Status FROM Account_Payable WHERE Reference_UUID = ?"},
		}
	case "PO_Item":
		checks = []refCheck{
			{table: "Container_Item", column: "PO_Item_ID", label: "Container Items (Bookings)",
				query: "SELECT Container_Item_ID, Container_ID, Load_Qty FROM Container_Item WHERE PO_Item_ID = ?"},
		}
	case "Commercial_Invoice":
		r.db.QueryRow("SELECT UUID FROM Commercial_Invoice WHERE CI_ID = ?", id).Scan(&refUUID)
		checks = []refCheck{
			{table: "Container_Item", column: "CI_ID", label: "Container Items (Bookings)",
				query: "SELECT Container_Item_ID, PO_Item_ID, Load_Qty FROM Container_Item WHERE CI_ID = ?"},
			{table: "Account_Payable", column: "Reference_UUID", label: "Linked Account Payables",
				query: "SELECT AP_ID, AP_No, Status FROM Account_Payable WHERE Reference_UUID = ?"},
		}
	case "Container":
		r.db.QueryRow("SELECT UUID FROM Container WHERE Container_ID = ?", id).Scan(&refUUID)
		checks = []refCheck{
			{table: "Container_Item", column: "Container_ID", label: "Container Items (Bookings)",
				query: "SELECT Container_Item_ID, PO_Item_ID, Load_Qty FROM Container_Item WHERE Container_ID = ?"},
			{table: "Account_Payable", column: "Reference_UUID", label: "Linked Account Payables",
				query: "SELECT AP_ID, AP_No, Status FROM Account_Payable WHERE Reference_UUID = ?"},
		}
	case "BL":
		r.db.QueryRow("SELECT UUID FROM BL WHERE BL_ID = ?", id).Scan(&refUUID)
		checks = []refCheck{
			{table: "Container_Item", column: "BL_ID", label: "Container Items (Bookings)",
				query: "SELECT Container_Item_ID, PO_Item_ID, Load_Qty FROM Container_Item WHERE BL_ID = ?"},
			{table: "Account_Payable", column: "Reference_UUID", label: "Linked Account Payables",
				query: "SELECT AP_ID, AP_No, Status FROM Account_Payable WHERE Reference_UUID = ?"},
		}
	case "Container_Item":
		r.db.QueryRow("SELECT UUID FROM Container_Item WHERE Container_Item_ID = ?", id).Scan(&refUUID)
		checks = []refCheck{
			{table: "Inventory_Lot", column: "Container_Item_ID", label: "Inventory Lots",
				query: "SELECT Lot_ID, Lot_No, Qty FROM Inventory_Lot WHERE Container_Item_ID = ?"},
			{table: "Account_Payable", column: "Reference_UUID", label: "Linked Account Payables",
				query: "SELECT AP_ID, AP_No, Status FROM Account_Payable WHERE Reference_UUID = ?"},
		}
	case "Goods_Receipt":
		r.db.QueryRow("SELECT UUID FROM Goods_Receipt WHERE GR_ID = ?", id).Scan(&refUUID)
		checks = []refCheck{
			{table: "Inventory_Lot", column: "GR_ID", label: "Inventory Lots",
				query: "SELECT Lot_ID, Lot_No, Qty FROM Inventory_Lot WHERE GR_ID = ?"},
			{table: "Account_Payable", column: "Reference_UUID", label: "Linked Account Payables",
				query: "SELECT AP_ID, AP_No, Status FROM Account_Payable WHERE Reference_UUID = ?"},
		}
	case "Inventory_Lot":
		r.db.QueryRow("SELECT UUID FROM Inventory_Lot WHERE Lot_ID = ?", id).Scan(&refUUID)
		checks = []refCheck{
			{table: "Cost_Allocation_Item", column: "Lot_ID", label: "Cost Allocation Items",
				query: "SELECT Cost_Allocation_Item_ID, Cost_Allocation_ID, Allocated_Amount FROM Cost_Allocation_Item WHERE Lot_ID = ?"},
			{table: "Account_Payable", column: "Reference_UUID", label: "Linked Account Payables",
				query: "SELECT AP_ID, AP_No, Status FROM Account_Payable WHERE Reference_UUID = ?"},
		}
	case "Account_Payable":
		checks = []refCheck{
			{table: "Cost_Allocation_Item", column: "AP_ID", label: "Cost Allocation Items",
				query: "SELECT Cost_Allocation_Item_ID, Cost_Allocation_ID, Allocated_Amount FROM Cost_Allocation_Item WHERE AP_ID = ?"},
		}
	case "Cost_Allocation":
		checks = []refCheck{
			{table: "Cost_Allocation_Item", column: "Cost_Allocation_ID", label: "Cost Allocation Items",
				query: "SELECT Cost_Allocation_Item_ID, Lot_ID, Allocated_Amount FROM Cost_Allocation_Item WHERE Cost_Allocation_ID = ?"},
		}
	case "AP_Target_Group":
		if err := r.db.QueryRow("SELECT UUID FROM AP_Target_Group WHERE Group_ID = ?", id).Scan(&refUUID); err != nil {
			return result, nil
		}
		checks = []refCheck{
			{table: "Account_Payable", column: "Reference_UUID", label: "Account Payables",
				query: "SELECT AP_ID, AP_No, Status FROM Account_Payable WHERE Reference_UUID = ? AND Reference_Type = 'AP_Target_Group'"},
		}
	default:
		return result, nil
	}

	for _, check := range checks {
		param := interface{}(id)
		if refUUID != "" && check.column == "Reference_UUID" {
			param = refUUID
		}
		rows, err := r.db.Query(check.query, param)
		if err != nil {
			continue
		}

		cols, _ := rows.Columns()
		records := []map[string]interface{}{}

		for rows.Next() {
			values := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range values {
				ptrs[i] = &values[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				continue
			}
			record := make(map[string]interface{})
			for i, col := range cols {
				val := values[i]
				if b, ok := val.([]byte); ok {
					record[col] = string(b)
				} else {
					record[col] = val
				}
			}
			records = append(records, record)
		}
		rows.Close()

		if len(records) > 0 {
			result.CanDelete = false
			result.References = append(result.References, models.ReferenceInfo{
				TableName: check.label,
				Count:     len(records),
				Records:   records,
			})
			result.TotalRefs += len(records)
		}
	}

	return result, nil
}

// DeleteRecord 지정된 테이블의 레코드 삭제
func (r *SQLiteProcurementRepository) DeleteRecord(tableName string, id int) error {
	var query string
	switch tableName {
	case "Item_Master":
		query = "DELETE FROM Item_Master WHERE Item_ID = ?"
	case "Vendor_Master":
		query = "DELETE FROM Vendor_Master WHERE Vendor_ID = ?"
	case "Purchase_Order":
		query = "DELETE FROM Purchase_Order WHERE PO_ID = ?"
	case "PO_Item":
		query = "DELETE FROM PO_Item WHERE PO_Item_ID = ?"
	case "Commercial_Invoice":
		query = "DELETE FROM Commercial_Invoice WHERE CI_ID = ?"
	case "Container":
		query = "DELETE FROM Container WHERE Container_ID = ?"
	case "BL":
		query = "DELETE FROM BL WHERE BL_ID = ?"
	case "Container_Item":
		query = "DELETE FROM Container_Item WHERE Container_Item_ID = ?"
	case "Goods_Receipt":
		query = "DELETE FROM Goods_Receipt WHERE GR_ID = ?"
	case "Inventory_Lot":
		_, _ = r.db.Exec("DELETE FROM Constraint_Instance WHERE Lot_ID = ?", id)
		query = "DELETE FROM Inventory_Lot WHERE Lot_ID = ?"
	case "Account_Payable":
		query = "DELETE FROM Account_Payable WHERE AP_ID = ?"
	case "Cost_Allocation":
		_, _ = r.db.Exec("DELETE FROM Cost_Allocation_Item WHERE Cost_Allocation_ID = ?", id)
		query = "DELETE FROM Cost_Allocation WHERE Cost_Allocation_ID = ?"
	case "Cost_Allocation_Item":
		query = "DELETE FROM Cost_Allocation_Item WHERE Cost_Allocation_Item_ID = ?"
	case "AP_Target_Group":
		query = "DELETE FROM AP_Target_Group WHERE Group_ID = ?"
	case "Constraint_Profile":
		_, _ = r.db.Exec("DELETE FROM Constraint_Profile_Step WHERE Profile_ID = ?", id)
		query = "DELETE FROM Constraint_Profile WHERE Profile_ID = ?"
	case "Constraint_Master":
		query = "DELETE FROM Constraint_Master WHERE Constraint_ID = ?"
	default:
		return fmt.Errorf("unsupported table: %s", tableName)
	}

	_, err := r.db.Exec(query, id)
	return err
}
