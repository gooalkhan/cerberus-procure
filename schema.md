# Database Schema

## 1. 마스터 정보 (Master Data)
시스템 운영의 기반이 되는 아이템 및 협력사 정보를 관리합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Item_Master** | Item_ID (PK), SKU_Code (UK, NN), Name (NN), Vendor_ID(FK, NN), CBM (>=0), Net_Weight (>=0), Gross_Weight (>=0), Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default) | 아이템 기본 속성. |
| **Vendor_Master** | Vendor_ID (PK), Name (NN), Category (Check), Business_Reg_No, Bank_Account, Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default) | Category: Supplier, Forwarder, Customs_Broker, Etc. |

## 2. 발주 및 통합 매입채무 (PO & Unified AP)
확정 데이터 중심 원칙에 따라 모든 지출 의무를 단일 창구에서 관리합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Purchase_Order** | PO_ID (PK), PO_Date (NN), PO_No (UK, NN), Vendor_ID (FK, NN), Currency (NN), Total_Amount (NN), Status (Default 'Open'), Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default), UUID (UK, NN) | 수입 발주 기본 정보. Status: Open, Closed |
| **PO_Item** | PO_Item_ID (PK), PO_ID (FK, NN), Item_ID (FK, NN), PO_Qty (>0), Unit_Price (NN), Status(Default 'Not Shipped'), Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default) | 발주 품목 정보. Status: Shipped, Partially Shipped, Not Shipped, Cancelled. **Trigger**: 모든 품목이 Shipped/Cancelled이면 PO를 Closed로, 하나라도 미완료면 Open으로 자동 변경. |
| **Commercial_Invoice** | CI_ID (PK), CI_No (UK, NN), Invoice_Date (NN), Vendor_ID (FK, NN), Currency (NN), Total_Amount (NN), Status, Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default), UUID (UK, NN)| 상업 송장 기본 정보. Status: Draft, Open, Closed |
| **Account_Payable** | AP_ID (PK), Vendor_ID (FK, NN), AP_No (UK, NN), Amount (NN), Currency (NN), Local_Amount, Allocation_Type (Check), Reference_UUID, Reference_Type, Due_Date, Allocation_Status, Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default), UUID (UK, NN) | AP 통합 관리. Status: Draft, Open, Closed |Allocation_Type: Amount, Quantity, CBM, Weight, lot, item|

## 3. 물류 및 선적 관리 (Logistics & Shipping)
N:N:N 아이템 매핑을 통해 복잡한 혼적 상황을 해결합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Container** | Container_ID (PK), Container_No (NN), Remark, UUID (UK, NN), Total_CBM, Total_Net_Wgt, Total_Gross_Wgt, Status| 운송 단위 정보. Status: Loaded, Shipping, Arrived|
| **Container_Item** | Container_Item_ID (PK), PO_Item_ID (FK, NN), Container_ID (FK), CI_ID (FK), BL_ID(FK), Item_ID(FK, NN), Unit_Price, Currency, Load_Qty, Gross_Weight, Net_Weight, Cbm | 컨테이너 적재 품목 매핑. |
| **BL** | BL_ID (PK), BL_No (UK, NN), ETD (NN), ETA (NN, >=ETD), POL, POD, Carrier, Vessel_Name, Status, Remark, UUID (UK, NN)| BL 정보. Status: Released, Partially Shipping, Shipping, Partially Arrived, Arrived|

## 4. 입고 및 로트 재고 (GR & Inventory Lot)
입고 시 유통기한별 로트 분할(Unpacking)을 수행하며, 최종 귀속 원가를 관리합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Goods_Receipt** | GR_ID (PK), Container_ID (FK, NN), BL_ID (FK, NN), Receive_Date (NN), Remark, Created_By, UUID (UK, NN) | 컨테이너 입고 기록. |
| **Inventory_Lot** | Lot_ID (PK), GR_ID (FK, NN), Container_Item_ID (FK, NN), Lot_No (NN), Expiry_Date (Nullable), Qty (>=0), Landed_Cost_Per_Unit, Quarantine_Status, Quarantine_Remark, Remark, UUID (UK, NN) | 재고 최소 단위. |

## 5. 랜딩 코스트 부대비용 (Landing Cost)
5대 배분 기준에 따라 상위 단위의 비용을 로트로 자동 안분합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Cost_Allocation** | Cost_Allocation_ID (PK), Allocation_date (NN), Total_Allocated_Amount (NN), Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default) | 배분 헤더. |
| **Cost_Allocation_Item** | Cost_Allocation_Item_ID (PK), Cost_Allocation_ID (FK, NN), Lot_ID (FK, NN), Allocated_Amount (NN), AP_ID (FK, NN) | 배분 상세. |

## 6. 사용자 정보 (Users)

| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| User_ID | INTEGER | PRIMARY KEY AUTOINCREMENT | Unique identifier |
| Username | TEXT | UNIQUE NOT NULL | User's login name |
| Password | TEXT | NOT NULL | Hashed password |
| Display_Name | TEXT | | Optional display name |
| Created_At | DATETIME | DEFAULT CURRENT_TIMESTAMP | Creation time |
| Updated_At | DATETIME | DEFAULT CURRENT_TIMESTAMP | Last update time |

---
**Abbreviation Key:**
*   **PK**: Primary Key
*   **FK**: Foreign Key
*   **NN**: Not Null
*   **UK**: Unique Key
*   **Check**: Check Constraint (Enum or Range)
*   **Default**: Default Value (Timestamp or Status)
