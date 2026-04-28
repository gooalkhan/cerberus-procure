export interface ItemMaster {
  item_id: number;
  sku_code: string;
  name: string;
  vendor_id: number;
  cbm: number;
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
  items?: POItem[];
}

export interface POItem {
  po_item_id: number;
  po_id: number;
  item_id: number;
  po_qty: number;
  unit_price: number;
  status: string;
  remark: string;
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
}

export interface CIAggregatedItem {
  item_id: number;
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
  due_date: string;
  uuid: string;
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
  item_id: number;
  unit_price: number;
  currency: string;
  load_qty: number;
  gross_weight: number;
  net_weight: number;
  cbm: number;
}

export interface BL {
  bl_id: number;
  bl_no: string;
  etd: string;
  eta: string;
  pol: string;
  pod: string;
  carrier: string;
  vessel_name: string;
  status: string;
  remark: string;
  uuid: string;
}

export interface GoodsReceipt {
  gr_id: number;
  container_id: number;
  bl_id: number;
  receive_date: string;
  remark: string;
  uuid: string;
}

export interface InventoryLot {
  lot_id: number;
  gr_id: number;
  container_item_id: number;
  lot_no: string;
  expiry_date?: string;
  qty: number;
  landed_cost_per_unit: number;
  quarantine_status: string;
  uuid: string;
}

export interface CostAllocation {
  cost_allocation_id: number;
  allocation_date: string;
  total_allocated_amount: number;
  remark: string;
}

export interface CostAllocationItem {
  cost_allocation_item_id: number;
  cost_allocation_id: number;
  lot_id: number;
  allocated_amount: number;
  ap_id: number;
}

export interface BookingView {
  container_item_id: number;
  container_id: number;
  container_no: string;
  status: string;
  total_cbm: number;
  total_net_wgt: number;
  total_gross_wgt: number;
  bl_id: number;
  bl_no: string;
  bl_status: string;
  etd: string;
  eta: string;
  pol: string;
  pod: string;
  carrier: string;
  vessel_name: string;
  po_item_id: number;
  ci_id: number;
  load_qty: number;
  unit_price: number;
  currency: string;
}
