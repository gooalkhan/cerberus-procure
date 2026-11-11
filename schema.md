# Database Schema

## 1. 마스터 정보 (Master Data)
시스템 운영의 기반이 되는 아이템 및 협력사 정보를 관리합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Item_Master** | Item_ID (PK), SKU_Code (UK, NN), Name (NN), Vendor_ID(FK, NN), Default_Constraint_Profile_ID (FK), CBM (>=0), Net_Weight (>=0), Gross_Weight (>=0), Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default) | 아이템 기본 속성. `Default_Constraint_Profile_ID`=SKU의 기본 행정 제약 프로파일. |
| **Vendor_Master** | Vendor_ID (PK), Name (NN), Category (Check), Business_Reg_No, Bank_Account, Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default) | Category: Supplier, Forwarder, Customs_Broker, Etc. |

## 2. 발주 및 통합 매입채무 (PO & Unified AP)
확정 데이터 중심 원칙에 따라 모든 지출 의무를 단일 창구에서 관리합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Purchase_Order** | PO_ID (PK), PO_Date (NN), PO_No (UK, NN), Vendor_ID (FK, NN), Currency (NN), Total_Amount (NN), Status (Default 'Open'), Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default), UUID (UK, NN) | 수입 발주 기본 정보. Status: Open, Closed |
| **PO_Item** | PO_Item_ID (PK), PO_ID (FK, NN), Item_ID (FK, NN), Base_Qty, PO_Qty (>=0), Unit_Price (NN, **불변**), Status(Default 'Not Shipped'), Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default) | 발주 품목(=**동일 SKU+동일 단가 버킷**). `UNIQUE(PO_ID, Item_ID, Unit_Price)`. `Base_Qty`=최초 기준 수량(불변), `PO_Qty`=현행(Base+Σ증분). Status: Shipped, Partially Shipped, Over Shipped, Not Shipped (취소=음수 증분). **파생값(Status·Container 집계)은 애플리케이션(유스케이스) 레이어에서 계산**(SQL 트리거 미사용). 모든 품목이 Shipped/Over Shipped이면 PO Closed. |
| **PO_Item_Adjustment** | Adjustment_ID (PK), PO_Item_ID (FK, NN), Kind ('QTY'), Qty_Delta, Old_Value, New_Value, Reason, Created_By, Created_At (Default) | PO 품목 **수량 증분 원장**. 참조 있는 PO는 수량을 Delta로만 변경(사유·작성자·시각). (A1: 단가는 불변이라 PRICE 미사용) |
| **Commercial_Invoice** | CI_ID (PK), CI_No (UK, NN), Invoice_Date (NN), Vendor_ID (FK, NN), Currency (NN), Total_Amount (NN), Status, Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default), UUID (UK, NN)| 상업 송장 기본 정보. Status: Draft, Open, Closed |
| **Account_Payable** | AP_ID (PK), Vendor_ID (FK, NN), AP_No (UK, NN), Amount (NN), Currency (NN), Local_Amount, Allocation_Type (Check), Reference_UUID, Reference_Type, Due_Date, Date_of_Payment, Allocation_Status, Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default), UUID (UK, NN) | AP 통합 관리. Status: Paid, Unpaid \|Allocation_Type: Weight, Volume, Quantity, Value, Row. Reference_Type에 `AP_Target_Group` 추가 지원. **Allocation_Status(파생): Local_Amount 전액 배분 시 `Closed`, 일부 배분 시 `Open`** (배분 원장 기준 조회 시 재계산 + 저장 시 반영). |

### 2.1 AP 타겟 그룹 (AP Target Group)
복수의 원천 문서(PO, CI, BL, Container 등)를 하나의 논리적 AP 타겟으로 묶어 1회 청구형 AP를 발행할 수 있습니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **AP_Target_Group** | Group_ID (PK), Group_No (UK, NN), Group_Name, Reference_Type (NN), Status, Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default), UUID (UK, NN) | AP 타겟 그룹 헤더. 하나의 UUID로 동일 종류의 여러 문서를 대표. 금액/통화/벤더는 그룹에 속한 기저 문서에서 집계됨. Reference_Type: PO, CI, BL, Container, Container_Item, GR, Lot. Status: Draft, Open, Closed |
| **AP_Target_Group_Item** | Group_Item_ID (PK), Group_ID (FK, NN), Reference_UUID (NN), Remark | 그룹에 포함된 개별 참조. 그룹의 Reference_Type과 동일한 종류의 문서만 포함 가능. |

### 2.2 행정 제약 프로파일 (Constraint Profile)
SKU별 기본 제약 조합을 정의하고, 선적 시 선택되어 GR/Lot 인스턴스로 전개됩니다(`BUSINESS_LOGIC.md` §5).

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Constraint_Profile** | Profile_ID (PK), Profile_No (UK, NN), Name (NN), Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default), UUID (UK, NN) | 행정 제약 프로파일(템플릿). 예: 식품 표준 / 일반. |
| **Constraint_Master** | Constraint_ID (PK), Name (UK, NN), Lead_Time_Days, Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default) | 개별 제약(step) 카탈로그. **이름 + 기본 소요기간(일)**. `Constraints` 메뉴에서 유저가 정의. |
| **Constraint_Profile_Step** | Step_ID (PK), Profile_ID (FK, NN), Constraint_ID (FK), Sequence (NN), Constraint_Type (NN), Lead_Time_Days, Is_Optional, Remark | 프로파일 단계 = **카탈로그 제약 참조 + 오버라이드**. `Constraint_ID`=참조, `Constraint_Type`=이름 스냅샷, `Lead_Time_Days`=유효값(기본 카탈로그, 변경 가능). `Is_Optional`=필요 시 enable. |
| **Constraint_Instance** | Instance_ID (PK), Lot_ID (FK, NN), Source_Container_Item_ID (FK), Profile_ID (FK), Predecessor_ID (FK), Sequence, Constraint_Type (NN), Lead_Time_Days, Is_Optional, Planned_Start, Actual_Start, Planned_End, Actual_End, Status, Remark, Created_By, Created_At, Updated_By, Updated_At | **Lot별 제약 인스턴스**. GR/Lot 생성 시 `Container_Item`의 프로파일에서 클론(선택 단계 제외). `Predecessor_ID`=체인 연결(연결 리스트). Status: Planned, InProgress, Done, Waived. 삽입/삭제/재정렬은 `Predecessor_ID` 재연결(relink) + cascade. |

> 파생: `Container_Item`의 **예상 Cost Allocation 가능일** = 앵커(`BL.ETA` 우선, 없으면 `Temporary_ETA`) + 도착→입고 리드 + Σ(프로파일 단계 리드). **Lot의 Cost Allocation 가능일** = GR일 앵커 + 인스턴스 체인 cascade(`start = Actual_Start ?? 선행 end/앵커`, `end = Actual_End ?? start + lead`). (앱 레이어 계산). 릴리즈·이벤트 로그/통계는 **범위 제외**(`BUSINESS_LOGIC.md` §5.11).

## 3. 물류 및 선적 관리 (Logistics & Shipping)
N:N:N 아이템 매핑을 통해 복잡한 혼적 상황을 해결합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Container** | Container_ID (PK), Container_No (NN), Remark, UUID (UK, NN), Total_CBM, Total_Net_Wgt, Total_Gross_Wgt, Status| 운송 단위 정보. Status: Loaded, Shipping, Arrived|
| **Container_Item** | Container_Item_ID (PK), PO_Item_ID (FK, NN), Container_ID (FK), CI_ID (FK), BL_ID(FK), Constraint_Profile_ID (FK), Load_Qty, Gross_Weight, Net_Weight, Cbm, Temporary_ETA, UUID(UK, NN) Remark, **Status(계산값)** | 컨테이너 적재 품목 매핑. 단가/통화는 PO_Item에서 파생. `Constraint_Profile_ID`=선적 시 선택된 행정 제약 프로파일(기본값: SKU). **Status는 저장하지 않고 조회 시 계산**: `Expected`(BL 미연결) / `BL Released`(now<ETD) / `Shipping`(ETD≤now<ETA) / `Arrived`(now≥ETA, GR 없음) / `Partially Landed`(일부 GR) / `Landed`(전체 GR). |
| **BL** | BL_ID (PK), BL_No (UK, NN), ETD (NN), ETA (NN, >=ETD), POL, POD, Carrier, Vessel_Name, Remark, UUID (UK, NN)| BL 정보. (Status 컬럼 없음 — 상태는 `Container_Item`에서 계산) |

## 4. 입고 및 로트 재고 (GR & Inventory Lot)
입고 시 유통기한별 로트 분할(Unpacking)을 수행하며, 최종 귀속 원가를 관리합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Goods_Receipt** | GR_ID (PK), GR_No, Receive_Date (NN), Remark, Created_By, UUID (UK, NN) | 컨테이너 입고 기록. |
| **Inventory_Lot** | Lot_ID (PK), GR_ID (FK, NN), Container_Item_ID (FK, NN), Lot_No (NN), Expiry_Date (Nullable), Qty (NN), Landed_Cost_Per_Unit, Quarantine_Status, Quarantine_Remark, Remark, UUID (UK, NN) | 재고 최소 단위. |

## 5. 랜딩 코스트 부대비용 (Landing Cost)
5대 배분 기준에 따라 상위 단위의 비용을 로트로 자동 안분합니다.

| Table | Columns | Description |
| :--- | :--- | :--- |
| **Cost_Allocation** | Cost_Allocation_ID (PK), Allocation_date (NN), Total_Allocated_Amount (NN), Remark, Created_By, Created_At (Default), Updated_By, Updated_At (Default) | 배분 헤더. |
| **Cost_Allocation_Item** | Cost_Allocation_Item_ID (PK), Cost_Allocation_ID (FK, NN), Lot_ID (FK, NN), Allocated_Amount (NN), AP_ID (FK, Nullable) | 배분 상세. **AP별 × Lot별** 단위로 1행씩 기록 (같은 Lot에 여러 AP가 배분되면 각각 행). AP_ID가 NULL이면 수동 입력 라인. 감사 정보는 부모 `Cost_Allocation`에서 관리. |

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
