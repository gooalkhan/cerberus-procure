import { 
  ItemMaster, VendorMaster, PurchaseOrder, POItem, 
  CommercialInvoice, CIAggregatedItem, AccountPayable, Container, 
  ContainerItem, BL, GoodsReceipt, InventoryLot, 
  CostAllocation, CostAllocationItem, CostAllocationLotCandidate, CostAllocationProposal,
  BookingView, BookingTemplateRow, BulkImportRow,
  AP_TargetGroup, AP_TargetGroupItem, APTargetGroupReference
} from './models';

declare global {
  interface Window {
    procureApi: {
      seedData: () => Promise<void>;
      getItems: () => Promise<string>;
      saveItem: (json: string) => Promise<void>;
      getVendors: () => Promise<string>;
      saveVendor: (json: string) => Promise<void>;
      getPurchaseOrders: () => Promise<string>;
      savePurchaseOrder: (json: string) => Promise<void>;
      getPOItems: (poId: number) => Promise<string>;
      savePOItem: (json: string) => Promise<void>;
      deletePOItem: (id: number) => Promise<void>;
      getCommercialInvoices: () => Promise<string>;
      getCIAggregatedItems: (ciId: number) => Promise<string>;
      saveCommercialInvoice: (json: string) => Promise<void>;
      getAccountPayables: () => Promise<string>;
      saveAccountPayable: (json: string) => Promise<void>;
      getContainers: () => Promise<string>;
      saveContainer: (json: string) => Promise<void>;
      getBLs: () => Promise<string>;
      saveBL: (json: string) => Promise<void>;
      getContainersByBLID: (id: number) => Promise<string>;
      saveContainerItem: (json: string) => Promise<void>;
      getContainerItemsByContainerID: (id: number) => Promise<string>;
      getGoodsReceipts: () => Promise<string>;
      saveGoodsReceipt: (json: string) => Promise<void>;
      getInventoryLots: () => Promise<string>;
      getInventoryLotsByGRID: (grId: number) => Promise<string>;
      saveInventoryLot: (json: string) => Promise<void>;
      getCostAllocations: () => Promise<string>;
      saveCostAllocation: (json: string) => Promise<void>;
      getCostAllocationItems: (costAllocationId: number) => Promise<string>;
      getAvailableCostAllocationLots: (late: boolean) => Promise<string>;
      getCostAllocationLotCandidatesByIDs: (json: string) => Promise<string>;
      calculateCostAllocation: (json: string) => Promise<string>;
      getBookings: () => Promise<string>;
      getBookingTemplate: () => Promise<string>;
      bulkImportBookings: (json: string) => Promise<string>;
      checkReferences: (table: string, id: number) => Promise<string>;
      deleteRecord: (table: string, id: number) => Promise<void>;
      getAPTargetGroups: () => Promise<string>;
      saveAPTargetGroup: (json: string) => Promise<string>;
      saveAPTargetGroupItem: (json: string) => Promise<string>;
      deleteAPTargetGroupItem: (id: number) => Promise<void>;
      getAPTargetGroupReferences: () => Promise<string>;
    };
  }
}

const isWasm = () => !!(window as any).procureApi;

function fixDates(obj: any): any {
  if (obj === null || typeof obj !== 'object') return obj;
  if (Array.isArray(obj)) return obj.map(fixDates);
  
  const newObj = { ...obj };
  for (const key in newObj) {
    const val = newObj[key];
    if (typeof val === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(val)) {
      newObj[key] = `${val}T00:00:00Z`;
    } else if (typeof val === 'object') {
      newObj[key] = fixDates(val);
    }
  }
  return newObj;
}

async function request<T>(path: string, method: string = 'GET', body?: any): Promise<T> {
  const processedBody = body ? fixDates(body) : body;
  const res = await fetch(`/api${path}`, {
    method,
    headers: processedBody ? { 'Content-Type': 'application/json' } : undefined,
    body: processedBody ? JSON.stringify(processedBody) : undefined,
  });
  if (!res.ok) throw new Error(await res.text());
  
  const text = await res.text();
  if (!text) return null as any;
  try {
    return JSON.parse(text);
  } catch (e) {
    return null as any;
  }
}

export const procureApi = {
  seedData: async (): Promise<void> => {
    if (isWasm()) return window.procureApi.seedData();
    return request('/seed', 'POST'); // Server might not implement this yet, but that's okay
  },
  // Items
  getItems: async (): Promise<ItemMaster[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getItems());
    return request<ItemMaster[]>('/items');
  },
  saveItem: async (item: ItemMaster): Promise<void> => {
    const processed = fixDates(item);
    if (isWasm()) return window.procureApi.saveItem(JSON.stringify(processed));
    return request('/items', 'POST', processed);
  },

  // Vendors
  getVendors: async (): Promise<VendorMaster[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getVendors());
    return request<VendorMaster[]>('/vendors');
  },
  saveVendor: async (v: VendorMaster): Promise<void> => {
    const processed = fixDates(v);
    if (isWasm()) return window.procureApi.saveVendor(JSON.stringify(processed));
    return request('/vendors', 'POST', processed);
  },

  // Purchase Orders
  getPurchaseOrders: async (): Promise<PurchaseOrder[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getPurchaseOrders());
    return request<PurchaseOrder[]>('/pos');
  },
  savePurchaseOrder: async (p: PurchaseOrder): Promise<void> => {
    const processed = fixDates(p);
    if (isWasm()) return window.procureApi.savePurchaseOrder(JSON.stringify(processed));
    return request('/pos', 'POST', processed);
  },

  // PO Items
  getPOItems: async (poId: number): Promise<POItem[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getPOItems(poId));
    return request<POItem[]>(`/pos/items?poId=${poId}`);
  },
  savePOItem: async (i: POItem): Promise<void> => {
    const processed = fixDates(i);
    if (isWasm()) return window.procureApi.savePOItem(JSON.stringify(processed));
    return request('/pos/items', 'POST', processed);
  },
  deletePOItem: async (id: number): Promise<void> => {
    if (isWasm()) return window.procureApi.deletePOItem(id);
    return request(`/pos/items?id=${id}`, 'DELETE');
  },

  // Invoices
  getCommercialInvoices: async (): Promise<CommercialInvoice[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getCommercialInvoices());
    return request<CommercialInvoice[]>('/invoices');
  },
  getCIAggregatedItems: async (ciId: number): Promise<CIAggregatedItem[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getCIAggregatedItems(ciId));
    return request<CIAggregatedItem[]>(`/invoices/items?ciId=${ciId}`);
  },
  saveCommercialInvoice: async (i: CommercialInvoice): Promise<void> => {
    const processed = fixDates(i);
    if (isWasm()) return window.procureApi.saveCommercialInvoice(JSON.stringify(processed));
    return request('/invoices', 'POST', processed);
  },

  // AP
  getAccountPayables: async (): Promise<AccountPayable[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getAccountPayables());
    return request<AccountPayable[]>('/aps');
  },
  saveAccountPayable: async (i: AccountPayable): Promise<void> => {
    const processed = fixDates(i);
    if (isWasm()) return window.procureApi.saveAccountPayable(JSON.stringify(processed));
    return request('/aps', 'POST', processed);
  },

  // Containers
  getContainers: async (): Promise<Container[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getContainers());
    return request<Container[]>('/containers');
  },
  saveContainer: async (i: Container): Promise<void> => {
    const processed = fixDates(i);
    if (isWasm()) return window.procureApi.saveContainer(JSON.stringify(processed));
    return request('/containers', 'POST', processed);
  },
  saveContainerItem: async (i: ContainerItem): Promise<void> => {
    const processed = fixDates(i);
    if (isWasm()) return window.procureApi.saveContainerItem(JSON.stringify(processed));
    return request('/containers/items', 'POST', processed);
  },
  getContainerItemsByContainerID: async (containerId: number): Promise<ContainerItem[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getContainerItemsByContainerID(containerId));
    return request<ContainerItem[]>(`/containers/items?containerId=${containerId}`);
  },

  // BL
  getBLs: async (): Promise<BL[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getBLs());
    return request<BL[]>('/bls');
  },
  saveBL: async (i: BL): Promise<void> => {
    const processed = fixDates(i);
    if (isWasm()) return window.procureApi.saveBL(JSON.stringify(processed));
    return request('/bls', 'POST', processed);
  },
  getContainersByBLID: async (blId: number): Promise<Container[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getContainersByBLID(blId));
    return request<Container[]>(`/containers/bl?blId=${blId}`);
  },

  // Goods Receipt
  getGoodsReceipts: async (): Promise<GoodsReceipt[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getGoodsReceipts());
    return request<GoodsReceipt[]>('/grs');
  },
  getInventoryLotsByGRID: async (grId: number): Promise<InventoryLot[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getInventoryLotsByGRID(grId));
    return request<InventoryLot[]>(`/lots/gr?grId=${grId}`);
  },
  saveGoodsReceipt: async (i: GoodsReceipt): Promise<void> => {
    const processed = fixDates(i);
    if (isWasm()) return window.procureApi.saveGoodsReceipt(JSON.stringify(processed));
    return request('/grs', 'POST', processed);
  },

  // Inventory Lot
  getInventoryLots: async (): Promise<InventoryLot[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getInventoryLots());
    return request<InventoryLot[]>('/lots');
  },
  saveInventoryLot: async (i: InventoryLot): Promise<void> => {
    const processed = fixDates(i);
    if (isWasm()) return window.procureApi.saveInventoryLot(JSON.stringify(processed));
    return request('/lots', 'POST', processed);
  },

  // Cost Allocation
  getCostAllocations: async (): Promise<CostAllocation[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getCostAllocations());
    return request<CostAllocation[]>('/allocations');
  },
  saveCostAllocation: async (i: CostAllocation): Promise<void> => {
    const processed = fixDates(i);
    if (isWasm()) return window.procureApi.saveCostAllocation(JSON.stringify(processed));
    return request('/allocations', 'POST', processed);
  },
  getCostAllocationItems: async (costAllocationId: number): Promise<CostAllocationItem[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getCostAllocationItems(costAllocationId));
    return request<CostAllocationItem[]>(`/allocations/items?costAllocationId=${costAllocationId}`);
  },
  getAvailableCostAllocationLots: async (late: boolean = false): Promise<CostAllocationLotCandidate[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getAvailableCostAllocationLots(late));
    return request<CostAllocationLotCandidate[]>(`/allocations/available-lots?late=${late ? 'true' : 'false'}`);
  },
  getCostAllocationLotCandidatesByIDs: async (lotIds: number[]): Promise<CostAllocationLotCandidate[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getCostAllocationLotCandidatesByIDs(JSON.stringify(lotIds)));
    return request<CostAllocationLotCandidate[]>('/allocations/lot-candidates', 'POST', lotIds);
  },
  calculateCostAllocation: async (lotIds: number[], late: boolean = false): Promise<CostAllocationProposal[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.calculateCostAllocation(JSON.stringify({ lot_ids: lotIds, late })));
    return request<CostAllocationProposal[]>('/allocations/calculate', 'POST', { lot_ids: lotIds, late });
  },

  // Unified Views
  getBookings: async (): Promise<BookingView[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getBookings());
    return request<BookingView[]>('/bookings');
  },

  // Booking Template
  getBookingTemplate: async (): Promise<BookingTemplateRow[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getBookingTemplate());
    return request<BookingTemplateRow[]>('/bookings/template');
  },

  // Bulk Import
  bulkImportBookings: async (rows: BulkImportRow[]): Promise<{ created: number; total: number }> => {
    if (isWasm()) return JSON.parse(await window.procureApi.bulkImportBookings(JSON.stringify(rows)));
    return request<{ created: number; total: number }>('/bookings/bulk-import', 'POST', rows);
  },

  // References & Delete
  checkReferences: async (table: string, id: number): Promise<{ can_delete: boolean; references: any[]; total_refs: number }> => {
    if (isWasm()) return JSON.parse(await window.procureApi.checkReferences(table, id));
    return request(`/references?table=${encodeURIComponent(table)}&id=${id}`);
  },

  deleteRecord: async (table: string, id: number): Promise<void> => {
    if (isWasm()) return window.procureApi.deleteRecord(table, id);
    return request(`/delete?table=${encodeURIComponent(table)}&id=${id}`, 'DELETE');
  },

  // AP Target Group
  getAPTargetGroups: async (): Promise<AP_TargetGroup[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getAPTargetGroups());
    return request<AP_TargetGroup[]>('/ap-target-groups');
  },

  saveAPTargetGroup: async (g: AP_TargetGroup): Promise<AP_TargetGroup> => {
    if (isWasm()) return JSON.parse(await window.procureApi.saveAPTargetGroup(JSON.stringify(g)));
    return request<AP_TargetGroup>('/ap-target-groups', 'POST', g);
  },

  saveAPTargetGroupItem: async (item: AP_TargetGroupItem): Promise<AP_TargetGroupItem> => {
    if (isWasm()) return JSON.parse(await window.procureApi.saveAPTargetGroupItem(JSON.stringify(item)));
    return request<AP_TargetGroupItem>('/ap-target-groups/items', 'POST', item);
  },

  deleteAPTargetGroupItem: async (id: number): Promise<void> => {
    if (isWasm()) return window.procureApi.deleteAPTargetGroupItem(id);
    return request(`/ap-target-groups/items?id=${id}`, 'DELETE');
  },

  getAPTargetGroupReferences: async (): Promise<APTargetGroupReference[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getAPTargetGroupReferences());
    return request<APTargetGroupReference[]>('/ap-target-groups/references');
  },
};
