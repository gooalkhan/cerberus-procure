import { 
  ItemMaster, VendorMaster, PurchaseOrder, POItem, 
  CommercialInvoice, CIAggregatedItem, AccountPayable, Container, 
  ContainerItem, BL, GoodsReceipt, InventoryLot, 
  CostAllocation, CostAllocationItem, BookingView 
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
      getBookings: () => Promise<string>;
    };
  }
}

const isWasm = () => !!(window as any).procureApi;

async function request<T>(path: string, method: string = 'GET', body?: any): Promise<T> {
  const res = await fetch(`/api${path}`, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
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
    if (isWasm()) return window.procureApi.saveItem(JSON.stringify(item));
    return request('/items', 'POST', item);
  },

  // Vendors
  getVendors: async (): Promise<VendorMaster[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getVendors());
    return request<VendorMaster[]>('/vendors');
  },
  saveVendor: async (v: VendorMaster): Promise<void> => {
    if (isWasm()) return window.procureApi.saveVendor(JSON.stringify(v));
    return request('/vendors', 'POST', v);
  },

  // Purchase Orders
  getPurchaseOrders: async (): Promise<PurchaseOrder[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getPurchaseOrders());
    return request<PurchaseOrder[]>('/pos');
  },
  savePurchaseOrder: async (p: PurchaseOrder): Promise<void> => {
    if (isWasm()) return window.procureApi.savePurchaseOrder(JSON.stringify(p));
    return request('/pos', 'POST', p);
  },

  // PO Items
  getPOItems: async (poId: number): Promise<POItem[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getPOItems(poId));
    return request<POItem[]>(`/pos/items?poId=${poId}`);
  },
  savePOItem: async (i: POItem): Promise<void> => {
    if (isWasm()) return window.procureApi.savePOItem(JSON.stringify(i));
    return request('/pos/items', 'POST', i);
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
    if (isWasm()) return window.procureApi.saveCommercialInvoice(JSON.stringify(i));
    return request('/invoices', 'POST', i);
  },

  // AP
  getAccountPayables: async (): Promise<AccountPayable[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getAccountPayables());
    return request<AccountPayable[]>('/aps');
  },
  saveAccountPayable: async (i: AccountPayable): Promise<void> => {
    if (isWasm()) return window.procureApi.saveAccountPayable(JSON.stringify(i));
    return request('/aps', 'POST', i);
  },

  // Containers
  getContainers: async (): Promise<Container[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getContainers());
    return request<Container[]>('/containers');
  },
  saveContainer: async (i: Container): Promise<void> => {
    if (isWasm()) return window.procureApi.saveContainer(JSON.stringify(i));
    return request('/containers', 'POST', i);
  },
  saveContainerItem: async (i: ContainerItem): Promise<void> => {
    if (isWasm()) return window.procureApi.saveContainerItem(JSON.stringify(i));
    return request('/containers/items', 'POST', i);
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
    if (isWasm()) return window.procureApi.saveBL(JSON.stringify(i));
    return request('/bls', 'POST', i);
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
    if (isWasm()) return window.procureApi.saveGoodsReceipt(JSON.stringify(i));
    return request('/grs', 'POST', i);
  },

  // Inventory Lot
  getInventoryLots: async (): Promise<InventoryLot[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getInventoryLots());
    return request<InventoryLot[]>('/lots');
  },
  saveInventoryLot: async (i: InventoryLot): Promise<void> => {
    if (isWasm()) return window.procureApi.saveInventoryLot(JSON.stringify(i));
    return request('/lots', 'POST', i);
  },

  // Cost Allocation
  getCostAllocations: async (): Promise<CostAllocation[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getCostAllocations());
    return request<CostAllocation[]>('/allocations');
  },
  saveCostAllocation: async (i: CostAllocation): Promise<void> => {
    if (isWasm()) return window.procureApi.saveCostAllocation(JSON.stringify(i));
    return request('/allocations', 'POST', i);
  },

  // Unified Views
  getBookings: async (): Promise<BookingView[]> => {
    if (isWasm()) return JSON.parse(await window.procureApi.getBookings());
    return request<BookingView[]>('/bookings');
  },
};
