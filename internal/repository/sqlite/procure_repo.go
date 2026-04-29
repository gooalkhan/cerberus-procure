package sqlite

import (
	"cerberus-procure/internal/models"
	"database/sql"
	"fmt"
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
			Category TEXT CHECK (Category IN ('Supplier', 'Forwarder', 'Customs_Broker', 'Etc')),
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
			PO_Qty REAL CHECK (PO_Qty > 0),
			Unit_Price REAL NOT NULL,
			Status TEXT CHECK (Status IN ('Shipped', 'Partially Shipped', 'Not Shipped', 'Cancelled')) DEFAULT 'Not Shipped',
			Remark TEXT,
			Created_By TEXT,
			Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			Updated_By TEXT,
			Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (PO_ID) REFERENCES Purchase_Order(PO_ID),
			FOREIGN KEY (Item_ID) REFERENCES Item_Master(Item_ID)
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
			Allocation_Type TEXT CHECK (Allocation_Type IN ('Weight', 'Volume', 'Quantity', 'Value', 'Unit')),
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
			Item_ID INTEGER NOT NULL,
			Unit_Price REAL,
			Currency TEXT,
			Load_Qty REAL,
			Gross_Weight REAL,
			Net_Weight REAL,
			Cbm REAL,
			Remark TEXT,
			FOREIGN KEY (PO_Item_ID) REFERENCES PO_Item(PO_Item_ID),
			FOREIGN KEY (Container_ID) REFERENCES Container(Container_ID),
			FOREIGN KEY (CI_ID) REFERENCES Commercial_Invoice(CI_ID),
			FOREIGN KEY (BL_ID) REFERENCES BL(BL_ID),
			FOREIGN KEY (Item_ID) REFERENCES Item_Master(Item_ID)
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
			Status TEXT CHECK (Status IN ('Released', 'Partially Shipping', 'Shipping', 'Partially Arrived', 'Arrived')),
			Remark TEXT,
			UUID TEXT UNIQUE NOT NULL,
			CHECK (ETA >= ETD)
		)`,
		`CREATE TABLE IF NOT EXISTS Goods_Receipt (
			GR_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Container_ID INTEGER NOT NULL,
			BL_ID INTEGER NOT NULL,
			Receive_Date DATETIME NOT NULL,
			Remark TEXT,
			Created_By TEXT,
			UUID TEXT UNIQUE NOT NULL,
			FOREIGN KEY (Container_ID) REFERENCES Container(Container_ID),
			FOREIGN KEY (BL_ID) REFERENCES BL(BL_ID)
		)`,
		`CREATE TABLE IF NOT EXISTS Inventory_Lot (
			Lot_ID INTEGER PRIMARY KEY AUTOINCREMENT,
			GR_ID INTEGER NOT NULL,
			Container_Item_ID INTEGER NOT NULL,
			Lot_No TEXT NOT NULL,
			Expiry_Date DATETIME NOT NULL,
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
			AP_ID INTEGER NOT NULL,
			FOREIGN KEY (Cost_Allocation_ID) REFERENCES Cost_Allocation(Cost_Allocation_ID),
			FOREIGN KEY (Lot_ID) REFERENCES Inventory_Lot(Lot_ID),
			FOREIGN KEY (AP_ID) REFERENCES Account_Payable(AP_ID)
		)`,
		`CREATE TRIGGER IF NOT EXISTS trg_sync_po_status AFTER UPDATE OF Status ON PO_Item
		BEGIN
			-- Close PO if all items are Shipped or Cancelled
			UPDATE Purchase_Order
			SET Status = 'Closed'
			WHERE PO_ID = NEW.PO_ID
			  AND NOT EXISTS (
				  SELECT 1 FROM PO_Item
				  WHERE PO_ID = NEW.PO_ID
					AND Status NOT IN ('Shipped', 'Cancelled')
			  );
			
			-- Reopen PO if any item is Not Shipped or Partially Shipped
			UPDATE Purchase_Order
			SET Status = 'Open'
			WHERE PO_ID = NEW.PO_ID
			  AND EXISTS (
				  SELECT 1 FROM PO_Item
				  WHERE PO_ID = NEW.PO_ID
					AND Status NOT IN ('Shipped', 'Cancelled')
			  );
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_container_item_calc_insert
		AFTER INSERT ON Container_Item
		BEGIN
			-- 1. Calculate item fields
			UPDATE Container_Item
			SET Item_ID = (SELECT Item_ID FROM PO_Item WHERE PO_Item_ID = NEW.PO_Item_ID),
				Gross_Weight = (SELECT Gross_Weight FROM Item_Master im JOIN PO_Item pi ON im.Item_ID = pi.Item_ID WHERE pi.PO_Item_ID = NEW.PO_Item_ID) * NEW.Load_Qty,
				Net_Weight = (SELECT Net_Weight FROM Item_Master im JOIN PO_Item pi ON im.Item_ID = pi.Item_ID WHERE pi.PO_Item_ID = NEW.PO_Item_ID) * NEW.Load_Qty,
				Cbm = (SELECT CBM FROM Item_Master im JOIN PO_Item pi ON im.Item_ID = pi.Item_ID WHERE pi.PO_Item_ID = NEW.PO_Item_ID) * NEW.Load_Qty
			WHERE Container_Item_ID = NEW.Container_Item_ID;

			-- 2. Aggregate to Container
			UPDATE Container
			SET Total_CBM = (SELECT SUM(Cbm) FROM Container_Item WHERE Container_ID = NEW.Container_ID),
				Total_Net_Wgt = (SELECT SUM(Net_Weight) FROM Container_Item WHERE Container_ID = NEW.Container_ID),
				Total_Gross_Wgt = (SELECT SUM(Gross_Weight) FROM Container_Item WHERE Container_ID = NEW.Container_ID)
			WHERE Container_ID = NEW.Container_ID;
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_container_item_calc_update
		AFTER UPDATE ON Container_Item
		BEGIN
			-- 1. Calculate item fields
			UPDATE Container_Item
			SET Item_ID = (SELECT Item_ID FROM PO_Item WHERE PO_Item_ID = NEW.PO_Item_ID),
				Gross_Weight = (SELECT Gross_Weight FROM Item_Master im JOIN PO_Item pi ON im.Item_ID = pi.Item_ID WHERE pi.PO_Item_ID = NEW.PO_Item_ID) * NEW.Load_Qty,
				Net_Weight = (SELECT Net_Weight FROM Item_Master im JOIN PO_Item pi ON im.Item_ID = pi.Item_ID WHERE pi.PO_Item_ID = NEW.PO_Item_ID) * NEW.Load_Qty,
				Cbm = (SELECT CBM FROM Item_Master im JOIN PO_Item pi ON im.Item_ID = pi.Item_ID WHERE pi.PO_Item_ID = NEW.PO_Item_ID) * NEW.Load_Qty
			WHERE Container_Item_ID = NEW.Container_Item_ID;

			-- 2. Aggregate to NEW Container
			UPDATE Container
			SET Total_CBM = (SELECT SUM(Cbm) FROM Container_Item WHERE Container_ID = NEW.Container_ID),
				Total_Net_Wgt = (SELECT SUM(Net_Weight) FROM Container_Item WHERE Container_ID = NEW.Container_ID),
				Total_Gross_Wgt = (SELECT SUM(Gross_Weight) FROM Container_Item WHERE Container_ID = NEW.Container_ID)
			WHERE Container_ID = NEW.Container_ID;

			-- 3. Aggregate to OLD Container (if changed)
			UPDATE Container
			SET Total_CBM = (SELECT SUM(Cbm) FROM Container_Item WHERE Container_ID = OLD.Container_ID),
				Total_Net_Wgt = (SELECT SUM(Net_Weight) FROM Container_Item WHERE Container_ID = OLD.Container_ID),
				Total_Gross_Wgt = (SELECT SUM(Gross_Weight) FROM Container_Item WHERE Container_ID = OLD.Container_ID)
			WHERE Container_ID = OLD.Container_ID AND OLD.Container_ID <> NEW.Container_ID;
		END;`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration error: %w", err)
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
	rows, err := r.db.Query("SELECT Item_ID, SKU_Code, Name, Vendor_ID, IFNULL(CBM, 0), IFNULL(Net_Weight, 0), IFNULL(Gross_Weight, 0), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Item_Master")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.ItemMaster{}
	for rows.Next() {
		var i models.ItemMaster
		err := rows.Scan(&i.ID, &i.SKUCode, &i.Name, &i.VendorID, &i.CBM, &i.NetWeight, &i.GrossWeight, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetItemByID(id int) (*models.ItemMaster, error) {
	var i models.ItemMaster
	err := r.db.QueryRow("SELECT Item_ID, SKU_Code, Name, Vendor_ID, IFNULL(CBM, 0), IFNULL(Net_Weight, 0), IFNULL(Gross_Weight, 0), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Item_Master WHERE Item_ID = ?", id).
		Scan(&i.ID, &i.SKUCode, &i.Name, &i.VendorID, &i.CBM, &i.NetWeight, &i.GrossWeight, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &i, err
}

func (r *SQLiteProcurementRepository) SaveItem(i *models.ItemMaster) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Item_Master (SKU_Code, Name, Vendor_ID, CBM, Net_Weight, Gross_Weight, Remark, Created_By) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			i.SKUCode, i.Name, i.VendorID, i.CBM, i.NetWeight, i.GrossWeight, i.Remark, i.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Item_Master SET SKU_Code=?, Name=?, Vendor_ID=?, CBM=?, Net_Weight=?, Gross_Weight=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE Item_ID=?",
		i.SKUCode, i.Name, i.VendorID, i.CBM, i.NetWeight, i.GrossWeight, i.Remark, i.UpdatedBy, i.ID)
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
	rows, err := r.db.Query("SELECT PO_Item_ID, PO_ID, Item_ID, IFNULL(PO_Qty, 0), IFNULL(Unit_Price, 0), Status, IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM PO_Item WHERE PO_ID = ?", poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.POItem{}
	for rows.Next() {
		var i models.POItem
		err := rows.Scan(&i.ID, &i.POID, &i.ItemID, &i.POQty, &i.UnitPrice, &i.Status, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SavePOItem(i *models.POItem) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO PO_Item (PO_ID, Item_ID, PO_Qty, Unit_Price, Status, Remark, Created_By) VALUES (?, ?, ?, ?, ?, ?, ?)",
			i.POID, i.ItemID, i.POQty, i.UnitPrice, i.Status, i.Remark, i.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE PO_Item SET Item_ID=?, PO_Qty=?, Unit_Price=?, Status=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE PO_Item_ID=?",
		i.ItemID, i.POQty, i.UnitPrice, i.Status, i.Remark, i.UpdatedBy, i.ID)
	return err
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
	query := `SELECT ci.Item_ID, im.Name, SUM(ci.Load_Qty), SUM(ci.Load_Qty * ci.Unit_Price), ci.Currency
	          FROM Container_Item ci
	          LEFT JOIN Item_Master im ON ci.Item_ID = im.Item_ID
	          WHERE ci.CI_ID = ?
	          GROUP BY ci.Item_ID, im.Name, ci.Currency`
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
	rows, err := r.db.Query("SELECT AP_ID, Vendor_ID, AP_No, IFNULL(Amount, 0), IFNULL(Currency, ''), IFNULL(Local_Amount, 0), IFNULL(Allocation_Type, ''), IFNULL(Reference_UUID, ''), IFNULL(Reference_Type, ''), Due_Date, Date_of_Payment, Status, IFNULL(Allocation_Status, ''), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At, UUID FROM Account_Payable")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.AccountPayable{}
	for rows.Next() {
		var i models.AccountPayable
		var dueDate, dateOfPayment *time.Time
		err := rows.Scan(&i.ID, &i.VendorID, &i.APNo, &i.Amount, &i.Currency, &i.LocalAmount, &i.AllocationType, &i.ReferenceUUID, &i.ReferenceType, &dueDate, &dateOfPayment, &i.Status, &i.AllocationStatus, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt, &i.UUID)
		if err != nil {
			return nil, err
		}
		if dueDate != nil { i.DueDate = *dueDate }
		if dateOfPayment != nil { i.DateOfPayment = *dateOfPayment }
		list = append(list, i)
	}
	return list, nil
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
	rows, err := r.db.Query("SELECT BL_ID, BL_No, ETD, ETA, IFNULL(POL, ''), IFNULL(POD, ''), IFNULL(Carrier, ''), IFNULL(Vessel_Name, ''), Status, IFNULL(Remark, ''), UUID FROM BL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.BL{}
	for rows.Next() {
		var i models.BL
		rows.Scan(&i.ID, &i.BLNo, &i.ETD, &i.ETA, &i.POL, &i.POD, &i.Carrier, &i.VesselName, &i.Status, &i.Remark, &i.UUID)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveBL(i *models.BL) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO BL (BL_No, ETD, ETA, POL, POD, Carrier, Vessel_Name, Status, Remark, UUID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", 
			i.BLNo, i.ETD, i.ETA, i.POL, i.POD, i.Carrier, i.VesselName, i.Status, i.Remark, i.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE BL SET BL_No=?, ETD=?, ETA=?, POL=?, POD=?, Carrier=?, Vessel_Name=?, Status=?, Remark=? WHERE BL_ID=?", 
		i.BLNo, i.ETD, i.ETA, i.POL, i.POD, i.Carrier, i.VesselName, i.Status, i.Remark, i.ID)
	return err
}

// Goods Receipt & Inventory
func (r *SQLiteProcurementRepository) GetGoodsReceipts() ([]models.GoodsReceipt, error) {
	rows, err := r.db.Query("SELECT GR_ID, Container_ID, BL_ID, Receive_Date, IFNULL(Remark, ''), IFNULL(Created_By, ''), UUID FROM Goods_Receipt")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.GoodsReceipt{}
	for rows.Next() {
		var i models.GoodsReceipt
		rows.Scan(&i.ID, &i.ContainerID, &i.BLID, &i.ReceiveDate, &i.Remark, &i.CreatedBy, &i.UUID)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveGoodsReceipt(i *models.GoodsReceipt) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Goods_Receipt (Container_ID, BL_ID, Receive_Date, Remark, Created_By, UUID) VALUES (?, ?, ?, ?, ?, ?)",
			i.ContainerID, i.BLID, i.ReceiveDate, i.Remark, i.CreatedBy, i.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Goods_Receipt SET Container_ID=?, BL_ID=?, Receive_Date=?, Remark=? WHERE GR_ID=?", i.ContainerID, i.BLID, i.ReceiveDate, i.Remark, i.ID)
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
		rows.Scan(&i.ID, &i.GRID, &i.ContainerItemID, &i.LotNo, &i.ExpiryDate, &i.Qty, &i.LandedCostPerUnit, &i.QuarantineStatus, &i.QuarantineRemark, &i.Remark, &i.UUID)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) GetInventoryLotsByGRID(grID int) ([]models.InventoryLot, error) {
	rows, err := r.db.Query("SELECT Lot_ID, GR_ID, Container_Item_ID, Lot_No, Expiry_Date, IFNULL(Qty, 0), IFNULL(Landed_Cost_Per_Unit, 0), IFNULL(Quarantine_Status, ''), IFNULL(Remark, ''), UUID FROM Inventory_Lot WHERE GR_ID = ?", grID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.InventoryLot{}
	for rows.Next() {
		var i models.InventoryLot
		var expiryStr *string
		rows.Scan(&i.ID, &i.GRID, &i.ContainerItemID, &i.LotNo, &expiryStr, &i.Qty, &i.LandedCostPerUnit, &i.QuarantineStatus, &i.Remark, &i.UUID)
		if expiryStr != nil {
			i.ExpiryDate, _ = time.Parse("2006-01-02", *expiryStr)
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveInventoryLot(i *models.InventoryLot) error {
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Inventory_Lot (GR_ID, Container_Item_ID, Lot_No, Expiry_Date, Qty, Landed_Cost_Per_Unit, Quarantine_Status, Quarantine_Remark, Remark, UUID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			i.GRID, i.ContainerItemID, i.LotNo, i.ExpiryDate, i.Qty, i.LandedCostPerUnit, i.QuarantineStatus, i.QuarantineRemark, i.Remark, i.UUID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Inventory_Lot SET GR_ID=?, Container_Item_ID=?, Lot_No=?, Expiry_Date=?, Qty=?, Landed_Cost_Per_Unit=?, Quarantine_Status=?, Quarantine_Remark=?, Remark=? WHERE Lot_ID=?",
		i.GRID, i.ContainerItemID, i.LotNo, i.ExpiryDate, i.Qty, i.LandedCostPerUnit, i.QuarantineStatus, i.QuarantineRemark, i.Remark, i.ID)
	return err
}

// Cost Allocation
func (r *SQLiteProcurementRepository) GetCostAllocations() ([]models.CostAllocation, error) {
	rows, err := r.db.Query("SELECT Cost_Allocation_ID, Allocation_date, IFNULL(Total_Allocated_Amount, 0), IFNULL(Remark, ''), IFNULL(Created_By, ''), Created_At, IFNULL(Updated_By, ''), Updated_At FROM Cost_Allocation")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.CostAllocation{}
	for rows.Next() {
		var i models.CostAllocation
		rows.Scan(&i.ID, &i.AllocationDate, &i.TotalAllocatedAmount, &i.Remark, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt)
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveCostAllocation(ca *models.CostAllocation) error {
	if ca.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Cost_Allocation (Allocation_date, Total_Allocated_Amount, Remark, Created_By) VALUES (?, ?, ?, ?)",
			ca.AllocationDate, ca.TotalAllocatedAmount, ca.Remark, ca.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		ca.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Cost_Allocation SET Allocation_date=?, Total_Allocated_Amount=?, Remark=?, Updated_By=?, Updated_At=CURRENT_TIMESTAMP WHERE Cost_Allocation_ID=?",
		ca.AllocationDate, ca.TotalAllocatedAmount, ca.Remark, ca.UpdatedBy, ca.ID)
	return err
}

// Container Items
func (r *SQLiteProcurementRepository) GetContainerItemsByContainerID(containerID int) ([]models.ContainerItem, error) {
	rows, err := r.db.Query("SELECT Container_Item_ID, PO_Item_ID, IFNULL(Container_ID, 0), IFNULL(CI_ID, 0), IFNULL(BL_ID, 0), Item_ID, IFNULL(Unit_Price, 0), IFNULL(Currency, ''), IFNULL(Load_Qty, 0), IFNULL(Gross_Weight, 0), IFNULL(Net_Weight, 0), IFNULL(Cbm, 0), IFNULL(Remark, '') FROM Container_Item WHERE Container_ID = ?", containerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.ContainerItem{}
	for rows.Next() {
		var i models.ContainerItem
		err := rows.Scan(&i.ID, &i.POItemID, &i.ContainerID, &i.CIID, &i.BLID, &i.ItemID, &i.UnitPrice, &i.Currency, &i.LoadQty, &i.GrossWeight, &i.NetWeight, &i.CBM, &i.Remark)
		if err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *SQLiteProcurementRepository) SaveContainerItem(i *models.ContainerItem) error {
	var err error
	if i.ID == 0 {
		_, err = r.db.Exec(`INSERT INTO Container_Item (PO_Item_ID, Container_ID, CI_ID, BL_ID, Item_ID, Unit_Price, Currency, Load_Qty, Gross_Weight, Net_Weight, Cbm, Remark) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			i.POItemID, i.ContainerID, i.CIID, i.BLID, i.ItemID, i.UnitPrice, i.Currency, i.LoadQty, i.GrossWeight, i.NetWeight, i.CBM, i.Remark)
	} else {
		_, err = r.db.Exec(`UPDATE Container_Item SET PO_Item_ID=?, Container_ID=?, CI_ID=?, BL_ID=?, Item_ID=?, Unit_Price=?, Currency=?, Load_Qty=?, Gross_Weight=?, Net_Weight=?, Cbm=?, Remark=? WHERE Container_Item_ID=?`,
			i.POItemID, i.ContainerID, i.CIID, i.BLID, i.ItemID, i.UnitPrice, i.Currency, i.LoadQty, i.GrossWeight, i.NetWeight, i.CBM, i.Remark, i.ID)
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
	if i.ID == 0 {
		res, err := r.db.Exec("INSERT INTO Cost_Allocation_Item (Cost_Allocation_ID, Lot_ID, Allocated_Amount, AP_ID) VALUES (?, ?, ?, ?)",
			i.CostAllocationID, i.LotID, i.AllocatedAmount, i.APID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		i.ID = int(id)
		return nil
	}
	_, err := r.db.Exec("UPDATE Cost_Allocation_Item SET Cost_Allocation_ID=?, Lot_ID=?, Allocated_Amount=?, AP_ID=? WHERE Cost_Allocation_Item_ID=?",
		i.CostAllocationID, i.LotID, i.AllocatedAmount, i.APID, i.ID)
	return err
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

func (r *SQLiteProcurementRepository) GetBookings() ([]models.BookingView, error) {
	query := `
		SELECT 
			ci.Container_Item_ID,
			IFNULL(ci.Container_ID, 0),
			IFNULL(c.Container_No, ''),
			IFNULL(c.Status, ''),
			IFNULL(c.Total_CBM, 0),
			IFNULL(c.Total_Net_Wgt, 0),
			IFNULL(c.Total_Gross_Wgt, 0),
			IFNULL(ci.BL_ID, 0),
			IFNULL(b.BL_No, ''),
			IFNULL(b.Status, '') as BL_Status,
			b.ETD,
			b.ETA,
			IFNULL(b.POL, ''),
			IFNULL(b.POD, ''),
			IFNULL(b.Carrier, ''),
			IFNULL(b.Vessel_Name, ''),
			IFNULL(ci.PO_Item_ID, 0),
			IFNULL(ci.Item_ID, 0),
			IFNULL(im.Name, '') as Item_Name,
			IFNULL(ci.CI_ID, 0),
			IFNULL(ci.Load_Qty, 0),
			IFNULL(ci.Unit_Price, 0),
			IFNULL(ci.Currency, ''),
			IFNULL(ci.Gross_Weight, 0),
			IFNULL(ci.Net_Weight, 0),
			IFNULL(ci.Cbm, 0),
			IFNULL(ci.Remark, '')
		FROM Container_Item ci
		LEFT JOIN Container c ON ci.Container_ID = c.Container_ID
		LEFT JOIN BL b ON ci.BL_ID = b.BL_ID
		LEFT JOIN Item_Master im ON ci.Item_ID = im.Item_ID
		ORDER BY 
			CASE c.Status 
				WHEN 'Loaded' THEN 1 
				WHEN 'Shipping' THEN 2 
				WHEN 'Arrived' THEN 3 
				ELSE 4 
			END,
			b.ETA ASC,
			c.Container_No ASC
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
		err := rows.Scan(
			&b.ContainerItemID, &b.ContainerID, &b.ContainerNo, &b.Status,
			&b.TotalCBM, &b.TotalNetWgt, &b.TotalGrossWgt,
			&b.BLID, &b.BLNo, &b.BLStatus, &etd, &eta,
			&b.POL, &b.POD, &b.Carrier, &b.VesselName,
			&b.POItemID, &b.ItemID, &b.ItemName, &b.CIID, &b.LoadQty, &b.UnitPrice, &b.Currency,
			&b.GrossWeight, &b.NetWeight, &b.CBM, &b.Remark,
		)
		if err != nil {
			return nil, err
		}
		if etd != nil { b.ETD = *etd }
		if eta != nil { b.ETA = *eta }
		list = append(list, b)
	}
	return list, nil
}
