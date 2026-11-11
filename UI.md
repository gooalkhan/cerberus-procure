# Cerberus Procure - User Interface Design

Cerberus Procure is built with a modern, dark-themed interface designed for efficiency in procurement and logistics management. The UI utilizes a premium aesthetic with deep blue backgrounds, sleek panels, and vibrant accent colors.

## 1. Global Layout & Theme

- **Color Palette**: 
  - Background: `#0f172a` (Slate Deep Blue)
  - Panels: `#1e293b` (Slate Dark)
  - Primary Accent: `#38bdf8` (Sky Blue)
  - Success: `#10b981` (Emerald Green)
  - Danger: `#ef4444` (Rose Red)
- **Typography**: Uses 'Inter' for body text and 'Outfit' for headings to provide a clean, professional look.
- **Atmosphere**: Utilizes glassmorphism (backdrop-blur), smooth transitions, and radial gradients for a premium feel.

## 2. Navigation (Sidebar)

The sidebar provides central navigation across all modules:
- ** Shipment Tracking Monitor**: An overview of shipments(Planned) 
- **📊 AP Aging Report**: Financial overview (Planned/Implemented).
- **📦 Item Master**: Core product database.
- **🤝 Vendor Master**: Supplier management.
- **📝 Purchase Orders**: Managing procurement requests.
- **🚢 Logistics (Bookings)**: Tracking shipments and container items.
- **📄 BL Management**: Bill of Lading lifecycle.
- **📥 Container Master**: Physical shipping container tracking.
- **🧾 Commercial Invoices**: Supplier billing.
- **💰 Account Payables**: Unified debt management.
- **🏭 Landed Goods**: Inventory receipt and lot unpacking.
- **⚖️ Cost Allocations**: Landing cost distribution.

## 3. Core Component: CrudPage

Most modules utilize the `CrudPage` component, which provides a standardized workflow:
- **Filter Panel**: Top-area filters for searching and narrowing down data.
- **Data Table**: A clean list view with hover effects and sorting capabilities.
- **Edit Modal**: A comprehensive popup form for creating or modifying records.
- **Detail View**: An expandable section within the modal for complex sub-relationships (e.g., items within a PO).

## 4. Specialized UI Components

### 🛒 PO Item Detail
- Integrated within the Purchase Order edit modal.
- Provides a sub-table for managing individual line items with automatic PO total calculation.

### 🔍 UUID Search Modal
- A dedicated lookup tool used whenever a "Reference UUID" is required.
- Allows users to search across different document types (PO, CI, BL, etc.) and select records to establish links.

### Container_Item Merge & Split
- In order to facilitate user's manual data entry and correction, the Container_Item table supports merge and split operations.
- If there is an AP attached to the Container_Item and user tries to merge or split it, the UI warns user that the AP will be orphaned after the merge or split operation.

### 📥 Bulk Booking Import via Excel (Container Item Entry)
- Located on the **Logistics (Bookings)** page alongside the `+ New Entry` button.
- **Pre-filled Template Download**:
  - Automatically queries all `Open` status Purchase Orders and their child line items (`PO_Item`).
  - Calculates the remaining unbooked quantity for each item (`Remaining Qty = PO_Qty - SUM(Load_Qty)` across existing `Container_Item` records).
  - Generates and downloads a pre-populated Excel (`.xlsx`) sheet containing only items with remaining unbooked quantities (`Remaining Qty > 0`).
  - Pre-filled columns:
    - **PO Item ID** (`po_item_id`, Mandatory, system identifier)
    - **PO No**, **SKU Code**, **Item Name**, **Vendor** (Reference only)
    - **Ordered Qty**, **Remaining Qty** (Reference only)
    - **Load Qty** (Mandatory, defaulted to remaining unbooked quantity, user-editable)
    - **Unit Price**, **Currency** (Pre-filled from PO, editable)
    - **Container No / ID** (Optional/Input)
    - **BL No / ID** (Optional/Input)
    - **Temporary ETA** (Optional, date format `YYYY-MM-DD`)
    - **Remark** (Optional)
- **Batch Upload & Preview Confirmation**:
  - Allows users to drag-and-drop or upload the completed Excel file.
  - Automatically parses the data and performs row-by-row validation:
    - Validates that `PO Item ID` exists and belongs to an active Open PO.
    - Validates that `Load Qty` is positive and warns if it exceeds remaining unbooked quantity.
    - Resolves Container / BL by identifier or number.
  - Displays a **Preview & Confirmation Modal** summarizing rows to be imported, warnings (e.g., over-booking), and any error rows.
  - Upon user confirmation, bulk-creates `Container_Item` records, which automatically triggers the database status updates for `PO_Item` and parent `Purchase_Order`.

### 🏭 Landed Goods (Unpacking)
- A specialized interface for Goods Receipt (GR).
- Supports "Unpacking" logic where a single shipment is split into multiple `Inventory Lots` based on SKU or physical attributes.

### 🌊 Logistics Relationship Flow
- A visual "Step" tracker found in the Booking detail.
- Displays the full path: `Source PO` → `PO Item` → `Container` → `B/L` → `Invoice (CI)`.

### 🧾 CI Aggregation & AP Linkage
- Shows aggregated items loaded in a Commercial Invoice.
- Allows direct creation and linkage of `Account Payable` records to the invoice.

### 💰 AP Allocation Settings
- Advanced settings for financial reconciliation.
- Supports selection of allocation basis (Weight, Volume, Quantity, Value, Row) and polymorphic references.
- AP Allocation can be done on a single lot, which is the minimal unit of AP allocation. User can use CI, BL, Container, GR, Lot as a reference of batch import of its lots.

### AP Association
- AP can be associated with PO, CI, BL, Container, Container_Item, GR, Lot, or **AP Target Group**. User must explicitly specify the association. the association is based on UUID which is given to the target of the AP linkage.
- **AP Target Group**: A dedicated interface allows users to group **multiple documents of the same type** (multiple POs, multiple CIs, multiple BLs, multiple Containers, multiple Container_Items, multiple GRs, or multiple Lots) into a single logical AP target with its own UUID. The group is defined by a selected `Reference_Type`, and only references of that type can be added. This grouped UUID can then be selected when issuing a single AP that covers multiple underlying documents of the same kind. The group detail view lists all grouped references and their allocated amounts.
- Whenever the target is deleted, the System will block the deletion and list the APs that reference it (the AP would otherwise be orphaned). A revision (edit) to PO, CI, BL, Container, Container_Item, Lot does not change its UUID; UUIDs remain stable across quantity changes because Cost Allocation records amounts per Lot (snapshot) and recomputes proportions from current base values.

### Shipment Tracking Monitor
- User can see whole shipments that are not closed. Shipments are divided into stages(POed, loaded, shipped, discharged). User can see the overall progress of each shipment by week or by month.

### 💳 AP Aging Report
- User can see all APs that are not paid. APs are divided into buckets based on the due date. User can see the overall progress of each AP.

## 5. User Experience Features

- **Responsive Modals**: Large forms are organized into grids and are scrollable to fit various screen sizes.
- **Interactive Feedback**: Hover states on table rows, smooth transitions on modal opening, and color-coded status badges (e.g., 'Paid', 'Open').
- **Authentication Flow**: A clean, centralized login screen with session management and user profile persistence.

## 6. Administrative Constraint UI (Constraint Profile & Instance Chain)

Implements the model defined in `BUSINESS_LOGIC.md` §4–§5: administrative procedures (요건승인 / 검역 / 통관 …) are governed by a **SKU-level profile** (rule), selected per shipment, and **cloned per `Lot` at GR** (instance) where start/end dates cascade into the **Cost Allocation Ready date**.

> Note: this section is the **frontend change plan** (what to add/change). Backend endpoints and WASM bindings must be added in parallel; the frontend talks to them only through `procureApi.ts`.

### 6.1 Navigation Changes

* Add a sidebar item **⚙️ Constraint Profiles** (`constraint_profiles`) under a "Master Data" group.
* (Optional) **🧬 Constraint Types** catalog (요건승인/검역/통관/불시검문검사 + default lead time).

### 6.2 New Screen: Constraint Profile Manager

Standard `CrudPage` list (Profile No, Name, Step count) with a detail modal editing the profile and its ordered steps.

```
Constraint Profile
┌────────────────────────────────────────────────────────────────┐
│ Name: [식품 표준                ]   Remark: [................. ] │
├────┬────────────────┬──────────────┬──────────┬───────────────┤
│ #  │ Type           │ Lead (days)  │ Optional │ Actions       │
├────┼────────────────┼──────────────┼──────────┼───────────────┤
│ 1  │ 요건승인        │ [ 3 ]        │ [ ]      │ ▲ ▼ ✕         │
│ 2  │ 검역           │ [ 2 ]        │ [ ]      │ ▲ ▼ ✕         │
│ 3  │ 통관           │ [ 4 ]        │ [ ]      │ ▲ ▼ ✕         │
│ 4  │ 불시검문검사     │ [ 0 ]        │ [x]      │ ▲ ▼ ✕         │
└────┴────────────────┴──────────────┴──────────┴───────────────┘
[+ Add Step]                                    [Cancel]  [Save Profile]
```

* Row **▲ ▼** = reorder, **✕** = remove, `+ Add Step` = append.
* `Optional` steps are pre-declared but not instantiated until enabled (§5.9).

### 6.3 Item Master (change)

* Add `default_constraint_profile_id` (select) to the form/columns.
* Display the profile **name** in the table.

```
Item Master  ·  edit
┌──────────────────────────────────────────────────────────┐
│ SKU Code: [ITEM-001]   Name: [Premium Coffee Beans]       │
│ Vendor: [Global Trading Co.]                              │
│ Default Constraint Profile: [식품 표준 ▾]                 │
│ … packaging fields …                                      │
└──────────────────────────────────────────────────────────┘
```

### 6.4 Bookings / Container_Item (change)

* Add `constraint_profile_id` select, **auto-filled from the SKU default**, overridable.
* Show a read-only **projected** availability from `BL.ETA` (fallback: `Temporary_ETA`) (no persistence, §5.4).

```
Container_Item  ·  edit
┌──────────────────────────────────────────────────────────────┐
│ PO Item: VT260410 / ITEM-001     Load Qty: [ 100 ]            │
│ Temporary ETA: [2026-02-01]                                   │
│ Constraint Profile: [식품 표준 ▾]   (SKU default — overridable)│
│ ─────────────────────────────────────────────────────────────│
│ ▶ Expected Cost Allocation Ready: 2026-02-11   (projected)    │
└──────────────────────────────────────────────────────────────┘
```

### 6.5 Landed Goods (GR + Lot) — main change

Inside `LandedGoodsDetail`, add a **per-Lot constraint chain editor**. Instances are cloned from the shipment profile at GR/Lot creation.

```
Lot: LOT-COFFEE-001   (GR-0001 / ITEM-001)
┌────┬────────────────┬──────────────┬──────────────┬───────────────┐
│ #  │ Constraint     │ Start        │ End          │ Actions       │
├────┼────────────────┼──────────────┼──────────────┼───────────────┤
│ 1  │ 요건승인        │ [2026-02-02] │ [2026-02-05] │ ▲ ▼ ✕  [+ ins]│
│ 2  │ 검역           │ [2026-02-05] │ (lead 2d)    │ ▲ ▼ ✕  [+ ins]│
│ 3  │ 통관           │ (auto)       │ (auto)       │ ▲ ▼ ✕         │
└────┴────────────────┴──────────────┴──────────────┴───────────────┘
[ Enable optional: 불시검문검사 ]   [ Apply to all Lots of this shipment ]
▶ Cost Allocation Ready: 2026-02-11  (actual)     Status: In Progress
```

* **Insert between steps** = the `[+ ins]` action (relink, §5.8.1).
* **Reorder** = ▲ ▼; **Remove** = ✕ (§5.8.2/5.8.3).
* Editing **Start** or **End** triggers **cascade**; downstream auto dates refresh immediately.
* **Optional enable** adds an `Is_Optional` step to this Lot's chain.
* **Apply to all Lots of this shipment** = bulk edit keyed by `Source_Container_Item_ID` (§5.8.5).
* Display per-Lot **Cost Allocation Ready** (projected vs actual) with an `AvailabilityBadge`.

### 6.6 Cost Allocations (optional change)

* Show a **gate/warning** when a lot's Cost Allocation Ready date has not been reached.

### 6.7 Reusable Components

| Component | Purpose |
| :--- | :--- |
| `ConstraintChainEditor` | Step list + insert/delete/reorder + cascade; shared by Lot editor and Container_Item **preview** |
| `AvailabilityBadge` | Renders `Pending / In Progress / Ready` with projected-vs-actual styling |

### 6.8 Frontend Model & API Additions

* **`models.ts`**: `ConstraintProfile`, `ConstraintProfileStep`, `ConstraintInstance`; `ItemMaster.default_constraint_profile_id`; `ContainerItem.constraint_profile_id`; lot availability field.
* **`procureApi.ts`**: profile CRUD, instance CRUD (`getConstraintInstances(lotId)`, save/delete, insert/reorder), `bulkApplyConstraints`, `getExpectedAvailability`.
* **Server + WASM**: matching endpoints/bindings (transparent to the frontend via `procureApi`).

### 6.9 Phased Rollout

1. **Projection**: Constraint Profile Manager (§6.2) + Item Master (§6.3) + Bookings profile select & projected availability (§6.4). No instance persistence.
2. **Instance**: Landed Goods per-Lot chain editor + cascade (§6.5) + reusable components (§6.7) + bulk edit.
3. **Gate**: Cost Allocation availability gate (§6.6).
