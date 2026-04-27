# Database Schema

## 1. 마스터 정보 (Master Data)
시스템 운영의 기반이 되는 아이템 및 협력사 정보를 관리합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Item_Master** | Item_ID (PK), SKU_Code, Name, CBM, Net_Weight, Gross_Weight, Remark, Created_By, Created_At, Updated_By, Updated_At | 아이템 기본 속성. CBM 및 중량은 랜딩 코스트 안분의 필수 데이터임. |
| **Vendor_Master** | Vendor_ID (PK), Name, Category, Business_Reg_No, Bank_Account, Remark, Created_By, Created_At, Updated_By, Updated_At | Category: Supplier(공급자), Forwarder(포워더), Customs_Broker(관세사) 등 통합 관리. |

## 2. 발주 및 통합 매입채무 (PO & Unified AP)
확정 데이터 중심 원칙에 따라 모든 지출 의무를 단일 창구에서 관리합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Purchase_Order** | PO_ID (PK), PO_Date, PO_No, Vendor_ID (FK), Currency, Status, Due_Date, Remark, Created_By, Created_At, Updated_By, Updated_At, UUID (Unique) | 수입 발주 기본 정보. |
| **PO_Item** | PO_Item_ID (PK), PO_ID (FK), Item_ID (FK), PO_Qty, Unit_Price, Status, Remark, Created_By, Created_At, Updated_By, Updated_At | 발주 품목 정보. |
| **Commercial_Invoice** | CI_ID (PK), CI_No, Invoice_Date, Vendor_ID (FK), Currency, Total_Amount, Status, Due_Date, Remark, Created_By, Created_At, Updated_By, Updated_At, UUID (Unique)| 상업 송장 기본 정보. |
| **Account_Payable** | AP_ID (PK), Vendor_ID (FK), AP_No, Amount, Currency, Local_Amount, Allocation_Type, Reference_UUID, Reference_Type, Due_Date, Allocation_Status, Remark, Created_By, Created_At, Updated_By, Updated_At, UUID (Unique) | Allocation_Type: 'Weight', 'Volume', 'Quantity', 'Value', 'Unit(1/N)' 5종 지원. | Reference_UUID: 참조 정보: 참조 대상의 UUID (CI, PO, BL 등) | Reference_Type: 참조 정보: BL, Container, PO, CI, GR, Lot, etc |

## 3. 물류 및 선적 관리 (Logistics & Shipping)
N:N:N 아이템 매핑을 통해 복잡한 혼적 상황을 해결합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Container** | Container_ID (PK), Container_No, Remark, UUID(Unique)| 운송 단위 정보. |
| **Container_Item** | Container_Item_ID (PK), Container_ID (FK), CI_ID (FK), PO_Item_ID (FK), BL_ID(FK), Unit_Price, Currency, Load_Qty | N:N:N 매핑 핵심: 특정 PO에서 발주되어 특정 CI로 청구된 아이템이 어느 컨테이너에 실렸는지 추적. |
| **BL** | BL_ID (PK), BL_No, ETD, ETA, Vessel_Name, Remark, UUID(Unique)| BL 정보. |

## 4. 입고 및 로트 재고 (GR & Inventory Lot)
입고 시 유통기한별 로트 분할(Unpacking)을 수행하며, 최종 귀속 원가를 관리합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Goods_Receipt** | GR_ID (PK), Container_ID (FK), BL_ID (FK), Receive_Date, Remark, Created_By, UUID(Unique) | 컨테이너 입고 이벤트 기록. |
| **Inventory_Lot** | Lot_ID (PK), GR_ID (FK), Container_Item_ID (FK), Lot_No, Expiry_Date, Qty, Landed_Cost_Per_Unit, Quarantine_Status, Quarantine_Remark, Remark, UUID(Unique) | 재고 최소 단위. |

## 5. 랜딩 코스트 부대비용 (Landing Cost)
5대 배분 기준에 따라 상위 단위의 비용을 로트로 자동 안분합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Cost_Allocation** | Cost_Allocation_ID (PK), Allocation_date, Total_Allocated_Amount, Remark, Created_By, Created_At, Updated_By, Updated_At |
| **Cost_Allocation_Item** | Cost_Allocation_Item_ID (PK), Cost_Allocation_ID (FK), Lot_ID (FK), Allocated_Amount, AP_ID (FK) |

## 6. 사용자 정보 (Users)

| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| User_ID | INTEGER | PRIMARY KEY AUTOINCREMENT | Unique identifier for the user |
| Username | TEXT | UNIQUE NOT NULL | User's login name |
| Password | TEXT | NOT NULL | Hashed password |
| Display_Name | TEXT | | Optional display name |
| Created_At | DATETIME | DEFAULT CURRENT_TIMESTAMP | Account creation time |
| Updated_At | DATETIME | DEFAULT CURRENT_TIMESTAMP | Last update time |
