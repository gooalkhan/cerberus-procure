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
- Supports selection of allocation basis (Weight, Volume, Quantity, Value, Unit) and polymorphic references.
- AP Allocation can be done on a single lot, which is the minimal unit of AP allocation. User can use CI, BL, Container, GR, Lot as a reference of batch import of its lots.

### AP Association
- AP can be associated with PO, CI, BL, Container, Container_Item, GR, Lot. User must explicitly specify the association. the association is based on UUID which is given to the target of the AP linkage.
- Whenever the target is deleted, the System will warn the user that the AP will be orphaned after the deletion. A revision to PO, CI, BL, Container does not effect its given UUID, but a revision to Container_Item, Lot will effect its given UUID. So the System should warn the user that the AP will be orphaned after the revision.

### Shipment Tracking Monitor
- User can see whole shipments that are not closed. Shipments are divided into stages(POed, loaded, shipped, discharged). User can see the overall progress of each shipment by week or by month.

### 💳 AP Aging Report
- User can see all APs that are not paid. APs are divided into buckets based on the due date. User can see the overall progress of each AP.

## 5. User Experience Features

- **Responsive Modals**: Large forms are organized into grids and are scrollable to fit various screen sizes.
- **Interactive Feedback**: Hover states on table rows, smooth transitions on modal opening, and color-coded status badges (e.g., 'Paid', 'Open').
- **Authentication Flow**: A clean, centralized login screen with session management and user profile persistence.
