export interface ItemMaster {
  item_id: number;
  sku_code: string;
  name: string;
  vendor_id: number;
  default_constraint_profile_id: number;
  cbm: number;
  box_qty: number;
  boxes_per_pallet: number;
  box_width: number;
  box_length: number;
  box_height: number;
  net_weight: number;
  gross_weight: number;
  remark: string;
  created_at?: string;
  updated_at?: string;
}

export interface VendorMaster {
  vendor_id: number;
  name: string;
  category: string;
  business_reg_no: string;
  bank_account: string;
  remark: string;
}

export interface PurchaseOrder {
  po_id: number;
  po_date: string;
  po_no: string;
  vendor_id: number;
  currency: string;
  total_amount: number;
  status: string;
  uuid: string;
  remark: string;
  items?: POItem[];
}

export interface POItem {
  po_item_id: number;
  po_id: number;
  item_id: number;
  base_qty: number;
  po_qty: number;
  unit_price: number;
  status: string;
  remark: string;
  created_by?: string;
  created_at?: string;
}

export interface POAllocationWarning {
  has_allocation: boolean;
  allocated_lot_count: number;
  total_allocated_amount: number;
}

export interface POItemAdjustment {
  adjustment_id: number;
  po_item_id: number;
  kind: string;
  qty_delta: number;
  old_value: number;
  new_value: number;
  reason: string;
  created_by?: string;
  created_at?: string;
}

export interface CommercialInvoice {
  ci_id: number;
  ci_no: string;
  invoice_date: string;
  vendor_id: number;
  currency: string;
  total_amount: number;
  status: string;
  uuid: string;
  remark: string;
}

export interface CIAggregatedItem {
  item_id: number;
  item_name: string;
  total_qty: number;
  amount: number;
  currency: string;
}

export interface AccountPayable {
  ap_id: number;
  vendor_id: number;
  ap_no: string;
  amount: number;
  currency: string;
  local_amount: number;
  allocation_type: string;
  reference_uuid: string;
  reference_type: string;
  allocation_status: string;
  due_date: string | null;
  date_of_payment?: string | null;
  status: string;
  uuid: string;
  remark: string;
  has_allocation?: boolean;
}

export interface Container {
  container_id: number;
  container_no: string;
  remark: string;
  total_cbm: number;
  total_net_wgt: number;
  total_gross_wgt: number;
  status: string;
  uuid: string;
}

export interface ContainerItem {
  container_item_id: number;
  po_item_id: number;
  container_id: number;
  ci_id: number;
  bl_id: number;
  constraint_profile_id: number;
  load_qty: number;
  gross_weight: number;
  net_weight: number;
  cbm: number;
  temporary_eta: string | null;
  uuid: string;
  remark: string;
  status?: string;
}

export interface BL {
  bl_id: number;
  bl_no: string;
  etd: string | null;
  eta: string | null;
  pol: string;
  pod: string;
  carrier: string;
  vessel_name: string;
  remark: string;
  uuid: string;
}

export interface GoodsReceipt {
  gr_id: number;
  gr_no: string;
  receive_date: string;
  status?: string;
  remark: string;
  uuid: string;
}

export interface InventoryLot {
  lot_id: number;
  gr_id: number;
  container_item_id: number;
  lot_no: string;
  expiry_date?: string | null;
  qty: number;
  landed_cost_per_unit: number;
  quarantine_status: string;
  uuid: string;
}

export interface CostAllocation {
  cost_allocation_id: number;
  allocation_date: string;
  total_allocated_amount: number;
  is_late_cost_allocation: boolean;
  remark: string;
  items?: CostAllocationItem[];
}

export interface CostAllocationItem {
  cost_allocation_item_id: number;
  cost_allocation_id: number;
  lot_id: number;
  allocated_amount: number;
  ap_id: number;
}

export interface CostAllocationLotCandidate {
  lot_id: number;
  lot_no: string;
  gr_id: number;
  container_item_id: number;
  ci_id: number;
  container_id: number;
  container_no: string;
  bl_id: number;
  bl_no: string;
  po_id: number;
  po_no: string;
  item_id: number;
  item_name: string;
  qty: number;
  gross_weight: number;
  net_weight: number;
  cbm: number;
  unit_price: number;
  currency: string;
}

export interface CostAllocationItemProposal {
  lot_id: number;
  allocated_amount: number;
}

export interface CostAllocationProposal {
  ap_id: number;
  ap_no: string;
  amount: number;
  local_amount: number;
  currency: string;
  allocation_type: string;
  proposals: CostAllocationItemProposal[];
}

export interface BookingView {
  container_item_id: number;
  uuid: string;
  container_id: number;
  container_no: string;
  status: string;
  total_cbm: number;
  total_net_wgt: number;
  total_gross_wgt: number;
  bl_id: number;
  bl_no: string;
  etd: string | null;
  eta: string | null;
  pol: string;
  pod: string;
  carrier: string;
  vessel_name: string;
  po_item_id: number;
  po_id: number;
  po_no: string;
  item_id: number;
  item_name: string;
  ci_id: number;
  ci_no: string;
  load_qty: number;
  gross_weight: number;
  net_weight: number;
  cbm: number;
  master_cbm: number;
  unit_cbm: number;
  constraint_profile_id: number;
  temporary_eta: string | null;
  remark: string;
}

// Bulk Booking Import
export interface BookingTemplateRow {
  po_item_id: number;
  po_no: string;
  sku_code: string;
  item_name: string;
  vendor_name: string;
  ordered_qty: number;
  remaining_qty: number;
  load_qty: number;
  container_no: string;
  container_id: number;
  bl_no: string;
  bl_id: number;
  temporary_eta: string;
  remark: string;
}

export interface BulkImportRow {
  po_item_id: number;
  load_qty: number;
  container_no: string;
  container_id: number;
  bl_no: string;
  bl_id: number;
  temporary_eta: string;
  remark: string;
}

export interface BulkImportPreview {
  valid_rows: BulkImportPreviewRow[];
  error_rows: BulkImportPreviewRow[];
  warn_rows: BulkImportPreviewRow[];
  total_count: number;
  valid_count: number;
  error_count: number;
  warn_count: number;
}

export interface BulkImportPreviewRow {
  row_number: number;
  po_item_id: number;
  po_no: string;
  item_name: string;
  load_qty: number;
  remaining_qty: number;
  container_no: string;
  container_id: number;
  bl_no: string;
  bl_id: number;
  temporary_eta: string;
  remark: string;
  status: 'valid' | 'error' | 'warn';
  message: string;
}

// AP Target Group
export interface AP_TargetGroup {
  ap_target_group_id: number;
  group_no: string;
  group_name: string;
  reference_type: string;
  status: string;
  remark: string;
  uuid: string;
  items?: AP_TargetGroupItem[];
  created_at?: string;
  updated_at?: string;
}

export interface AP_TargetGroupItem {
  ap_target_group_item_id: number;
  ap_target_group_id: number;
  reference_uuid: string;
  remark: string;
}

export interface APTargetGroupReference {
  reference_uuid: string;
  reference_type: string;
  reference_no: string;
  description: string;
  amount: number;
  currency: string;
}

export interface ConstraintProfileStep {
  constraint_profile_step_id: number;
  constraint_profile_id: number;
  constraint_id: number;
  sequence: number;
  constraint_type: string;
  lead_time_days: number;
  is_optional: boolean;
  remark: string;
}

export interface ConstraintMaster {
  constraint_id: number;
  name: string;
  lead_time_days: number;
  remark: string;
  created_at?: string;
  updated_at?: string;
}

export interface ConstraintProfile {
  constraint_profile_id: number;
  profile_no: string;
  name: string;
  remark: string;
  uuid: string;
  steps?: ConstraintProfileStep[];
  created_at?: string;
  updated_at?: string;
}

export interface ConstraintAvailability {
  container_item_id: number;
  constraint_profile_id: number;
  profile_name: string;
  anchor_date: string | null;
  anchor_source: string;
  arrival_to_gr_days: number;
  total_lead_days: number;
  expected_ready_date: string | null;
  steps?: ConstraintProfileStep[];
}

export interface ConstraintInstance {
  constraint_instance_id: number;
  lot_id: number;
  source_container_item_id: number;
  constraint_profile_id: number;
  predecessor_id: number;
  sequence: number;
  constraint_type: string;
  lead_time_days: number;
  is_optional: boolean;
  planned_start?: string | null;
  actual_start?: string | null;
  planned_end?: string | null;
  actual_end?: string | null;
  status: string;
  remark: string;
  computed_start?: string | null;
  computed_end?: string | null;
}

export interface LotConstraintSchedule {
  lot_id: number;
  anchor_date: string | null;
  instances: ConstraintInstance[];
  ready_date: string | null;
  ready_is_actual: boolean;
}
