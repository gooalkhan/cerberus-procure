import React, { useState, useEffect, useCallback } from 'react'
import { User } from './api/authApi'
import { procureApi } from './api/procureApi'
import Login from './components/Login'
import Sidebar from './components/Sidebar'
import CrudPage from './components/CrudPage'
import POItemDetail from './components/POItemDetail'
import SearchModal from './components/SearchModal'

function App() {
  const [user, setUser] = useState<User | null>(null)
  const [activeMenu, setActiveMenu] = useState('items')

  const handleLogin = async (u: User) => {
    await procureApi.seedData();
    setUser(u)
  }

  if (!user) {
    return <Login onLogin={handleLogin} />
  }

  const renderContent = () => {
    switch (activeMenu) {
      case 'items':
        return (
          <CrudPage 
            title="Item Master"
            columns={[
              { key: 'sku_code', label: 'SKU Code' },
              { key: 'name', label: 'Name' },
              { key: 'vendor_id', label: 'Vendor ID', type: 'number', searchType: 'Vendor' },
              { key: 'cbm', label: 'CBM', type: 'number' },
              { key: 'net_weight', label: 'Net Weight', type: 'number' },
              { key: 'gross_weight', label: 'Gross Weight', type: 'number' },
            ]}
            fetchData={procureApi.getItems}
            onSave={procureApi.saveItem}
            emptyItem={{ item_id: 0, sku_code: '', name: '', vendor_id: 0, cbm: 0, net_weight: 0, gross_weight: 0, remark: '' }}
          />
        )
      case 'vendors':
        return (
          <CrudPage 
            title="Vendor Master"
            columns={[
              { key: 'name', label: 'Name' },
              { key: 'category', label: 'Category' },
              { key: 'business_reg_no', label: 'Reg No' },
              { key: 'bank_account', label: 'Bank Account' },
              { key: 'remark', label: 'Remark', fullWidth: true, filterType: 'none' },
            ]}
            fetchData={procureApi.getVendors}
            onSave={procureApi.saveVendor}
            emptyItem={{ vendor_id: 0, name: '', category: 'Supplier', business_reg_no: '', bank_account: '', remark: '' }}
            renderDetail={(vendor) => <VendorDetail vendor={vendor} />}
          />
        )
      case 'pos':
        return (
          <CrudPage 
            title="Purchase Orders"
            columns={[
              { key: 'po_no', label: 'PO No' },
              { key: 'po_date', label: 'PO Date', type: 'date' },
              { key: 'vendor_id', label: 'Vendor ID', type: 'number', searchType: 'Vendor' },
              { key: 'currency', label: 'Currency' },
              { key: 'total_amount', label: 'Total Amount', type: 'number', filterType: 'none' },
              { key: 'status', label: 'Status', filterType: 'select', filterOptions: ['Open', 'Closed'] },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getPurchaseOrders}
            onSave={procureApi.savePurchaseOrder}
            emptyItem={{ po_id: 0, po_no: '', po_date: new Date().toISOString(), vendor_id: 0, currency: 'USD', total_amount: 0, status: 'Open', remark: '', uuid: '' }}
            renderDetail={(po, onChange) => <POItemDetail po={po} onChange={onChange} />}
          />
        )
      case 'logistics':
        return (
          <CrudPage 
            title="Bookings"
            columns={[
              { key: 'container_no', label: 'Container No', formHidden: true },
              { key: 'bl_no', label: 'BL No', formHidden: true },
              { key: 'status', label: 'Status', filterType: 'select', filterOptions: ['Loaded', 'Shipping', 'Arrived'], formHidden: true },
              { key: 'eta', label: 'ETA', type: 'date', formHidden: true },
              { key: 'etd', label: 'ETD', type: 'date', formHidden: true },
              { key: 'container_id', label: 'Container ID', type: 'number', tableHidden: true, searchType: 'Container' },
              { key: 'bl_id', label: 'BL ID', type: 'number', tableHidden: true, searchType: 'BL' },
              { key: 'po_item_id', label: 'PO Item ID', type: 'number', tableHidden: true, searchType: 'PO Item' },
              { key: 'ci_id', label: 'CI ID', type: 'number', tableHidden: true, searchType: 'CI' },
              { key: 'item_id', label: 'Item ID', type: 'number', tableHidden: true, searchType: 'Item' },
              { key: 'item_name', label: 'Item Name', formHidden: true },
              { key: 'divider_1', label: '', divider: true },
              { key: 'load_qty', label: 'Load Qty', type: 'number', filterType: 'none' },
              { key: 'cbm', label: 'CBM', type: 'number', formHidden: true },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getBookings}
            onSave={async (booking: any) => {
              await procureApi.saveContainerItem({
                container_item_id: booking.container_item_id,
                container_id: booking.container_id,
                bl_id: booking.bl_id,
                po_item_id: booking.po_item_id,
                item_id: booking.item_id || 0,
                ci_id: booking.ci_id || 0,
                load_qty: booking.load_qty,
                unit_price: booking.unit_price,
                currency: booking.currency,
                gross_weight: booking.gross_weight || 0,
                net_weight: booking.net_weight || 0,
                cbm: booking.cbm || 0,
                remark: booking.remark || '',
              });
            }}
            emptyItem={{ container_item_id: 0, container_id: 0, container_no: '', status: 'Loaded', total_cbm: 0, total_net_wgt: 0, total_gross_wgt: 0, bl_id: 0, bl_no: '', bl_status: 'Released', etd: '', eta: '', pol: '', pod: '', carrier: '', vessel_name: '', po_item_id: 0, item_id: 0, ci_id: 0, load_qty: 0, unit_price: 0, currency: 'USD', gross_weight: 0, net_weight: 0, cbm: 0, remark: '' }}
            renderDetail={(booking) => <BookingFlow booking={booking} />}
          />
        )
      case 'bls':
        return (
          <CrudPage 
            title="BL Management"
            columns={[
              { key: 'bl_no', label: 'BL No' },
              { key: 'status', label: 'Status', filterType: 'select', filterOptions: ['Released', 'Partially Shipping', 'Shipping', 'Partially Arrived', 'Arrived'] },
              { key: 'etd', label: 'ETD', type: 'date' },
              { key: 'eta', label: 'ETA', type: 'date' },
              { key: 'pol', label: 'POL' },
              { key: 'pod', label: 'POD', filterType: 'text' },
              { key: 'carrier', label: 'Carrier' },
              { key: 'vessel_name', label: 'Vessel Name' },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getBLs}
            onSave={procureApi.saveBL}
            emptyItem={{ bl_id: 0, bl_no: '', etd: new Date().toISOString(), eta: new Date().toISOString(), pol: '', pod: '', carrier: '', vessel_name: '', status: 'Released', remark: '', uuid: '' }}
            renderDetail={(bl) => <BLDetail bl={bl} />}
          />
        )
      case 'containers':
        return (
          <CrudPage 
            title="Container Master"
            columns={[
              { key: 'container_no', label: 'Container No' },
              { key: 'status', label: 'Status', filterType: 'select', filterOptions: ['Loaded', 'Shipping', 'Arrived'] },
              { key: 'total_cbm', label: 'Total CBM', type: 'number', formHidden: true },
              { key: 'total_net_wgt', label: 'Net Wgt', type: 'number', formHidden: true },
              { key: 'total_gross_wgt', label: 'Gross Wgt', type: 'number', formHidden: true },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getContainers}
            onSave={procureApi.saveContainer}
            emptyItem={{ container_id: 0, container_no: '', status: 'Loaded', total_cbm: 0, total_net_wgt: 0, total_gross_wgt: 0, remark: '', uuid: '' }}
            renderDetail={(container) => <ContainerDetail container={container} />}
          />
        )
      case 'invoices':
        return (
          <CrudPage 
            title="Commercial Invoices"
            columns={[
              { key: 'ci_no', label: 'Invoice No' },
              { key: 'invoice_date', label: 'Invoice Date', type: 'date' },
              { key: 'vendor_id', label: 'Vendor ID', type: 'number', searchType: 'Vendor' },
              { key: 'currency', label: 'Currency' },
              { key: 'total_amount', label: 'Total Amount', type: 'number' },
              { key: 'status', label: 'Status', filterType: 'select', filterOptions: ['Draft', 'Open', 'Closed'] },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getCommercialInvoices}
            onSave={procureApi.saveCommercialInvoice}
            emptyItem={{ ci_id: 0, ci_no: '', invoice_date: new Date().toISOString(), vendor_id: 0, currency: 'USD', total_amount: 0, status: 'Draft', remark: '', uuid: '' }}
            renderDetail={(ci) => <CIDetail ci={ci} />}
          />
        )
      case 'aps':
        return (
          <CrudPage 
            title="Account Payables"
            columns={[
              { key: 'ap_no', label: 'AP No' },
              { key: 'vendor_id', label: 'Vendor ID', type: 'number', searchType: 'Vendor' },
              { key: 'amount', label: 'Amount', type: 'number', filterType: 'none' },
              { key: 'currency', label: 'Currency' },
              { key: 'local_amount', label: 'Local Amount', type: 'number', filterType: 'none' },
              { key: 'due_date', label: 'Due Date', type: 'date' },
              { key: 'date_of_payment', label: 'Payment Date', type: 'date', filterType: 'none' },
              { key: 'status', label: 'Pay Status', filterType: 'select', filterOptions: ['paid', 'unpaid'] },
              { key: 'allocation_status', label: 'Alloc Status', filterType: 'select', filterOptions: ['Draft', 'Open', 'Closed'] },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getAccountPayables}
            onSave={procureApi.saveAccountPayable}
            emptyItem={{ ap_id: 0, vendor_id: 0, ap_no: '', amount: 0, currency: 'USD', local_amount: 0, allocation_type: 'Amount', reference_uuid: '', reference_type: 'PO', due_date: new Date().toISOString(), date_of_payment: '', status: 'unpaid', allocation_status: 'Draft', remark: '', uuid: '' }}
            renderDetail={(ap, onChange) => <APDetail ap={ap} onChange={onChange} />}
          />
        )
      case 'inventory':
        return (
          <CrudPage 
            title="Landed Goods"
            columns={[
              { key: 'container_id', label: 'Container ID', type: 'number', searchType: 'Container' },
              { key: 'bl_id', label: 'BL ID', type: 'number', searchType: 'BL' },
              { key: 'receive_date', label: 'Receive Date', type: 'date' },
              { key: 'remark', label: 'Remark', fullWidth: true },
              { key: 'divider_lots', label: '', divider: true },
            ]}
            fetchData={procureApi.getGoodsReceipts}
            onSave={async (gr: any) => {
              // Simple save: GR first
              await procureApi.saveGoodsReceipt(gr);
              // Lots are handled in the component via renderDetail's onChange
              // But we need to make sure they are saved too.
              // For simplicity, we'll save them one by one if they have changes.
              if (gr.lots) {
                for (const lot of gr.lots) {
                  // We need the gr_id. If it's a new GR, we might need a better way.
                  // For now, assume gr_id is present if it's an edit, 
                  // or find the last GR if it's new.
                  let targetGrId = gr.gr_id;
                  if (!targetGrId) {
                    const allGrs = await procureApi.getGoodsReceipts();
                    const lastGr = allGrs[allGrs.length - 1]; // Naive approach
                    targetGrId = lastGr.gr_id;
                  }
                  await procureApi.saveInventoryLot({ ...lot, gr_id: targetGrId });
                }
              }
            }}
            emptyItem={{ gr_id: 0, container_id: 0, bl_id: 0, receive_date: new Date().toISOString(), remark: '', lots: [] }}
            renderDetail={(gr, onChange) => <LandedGoodsDetail gr={gr} onChange={onChange} />}
          />
        )
      case 'allocations':
        return (
          <CrudPage 
            title="Cost Allocations"
            columns={[
              { key: 'allocation_date', label: 'Date', type: 'date' },
              { key: 'total_allocated_amount', label: 'Total Amount', type: 'number' },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getCostAllocations}
            onSave={procureApi.saveCostAllocation}
            emptyItem={{ cost_allocation_id: 0, allocation_date: new Date().toISOString(), total_allocated_amount: 0, remark: '' }}
          />
        )
      default:
        return <div>Select a menu</div>
    }
  }

  return (
    <div className="app-container">
      <Sidebar activeMenu={activeMenu} setActiveMenu={setActiveMenu} />
      <div className="main-content">
        <header>
          <div className="user-info">
            Welcome, <strong>{user?.display_name}</strong> ({user?.username})
          </div>
          <button className="secondary" onClick={() => setUser(null)}>Logout</button>
        </header>
        <div className="content-area">
          {renderContent()}
        </div>
      </div>
    </div>
  )
}

function APDetail({ ap, onChange }: { ap: any, onChange: (updated: any) => void }) {
  const [refNo, setRefNo] = useState('')
  const [isSearchOpen, setIsSearchOpen] = useState(false)

  const handleOpenSearch = () => {
    setIsSearchOpen(true)
  }

  const handleSelect = (item: any) => {
    onChange({ ...ap, reference_uuid: item.uuid })
    setIsSearchOpen(false)
  }

  return (
    <div style={{ marginTop: '2rem', borderTop: '1px solid #333', paddingTop: '1.5rem' }}>
      <h3>Allocation & Reference Settings</h3>
      <div className="form-grid">
        <div className="form-group">
          <label>Allocation Type</label>
          <select 
            value={ap.allocation_type} 
            onChange={e => onChange({ ...ap, allocation_type: e.target.value })}
          >
            {['Amount', 'Quantity', 'CBM', 'Weight', 'Lot', 'Item'].map(t => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
        </div>
        <div className="form-group">
          <label>Reference Type</label>
          <select 
            value={ap.reference_type} 
            onChange={e => onChange({ ...ap, reference_type: e.target.value })}
          >
            {['PO', 'CI', 'Container', 'BL', 'GR', 'Lot'].map(t => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
        </div>
        <div className="form-group">
          <label>Reference Search</label>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <input 
              type="text" 
              placeholder={`Search ${ap.reference_type}...`}
              value={refNo}
              onChange={e => setRefNo(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleOpenSearch()}
          />
            <button type="button" className="secondary" onClick={handleOpenSearch}>Find UUID</button>
          </div>
        </div>
        <div className="form-group">
          <label>Resolved Reference UUID</label>
          <input type="text" value={ap.reference_uuid} readOnly style={{ opacity: 0.6 }} />
        </div>
      </div>

      {isSearchOpen && (
        <SearchModal 
          type={ap.reference_type} 
          searchTerm={refNo} 
          onClose={() => setIsSearchOpen(false)} 
          onSelect={handleSelect} 
        />
      )}
    </div>
  )
}


function CIDetail({ ci }: { ci: any }) {
  if (!ci.ci_id) return null;
  const [items, setItems] = useState<any[]>([]);
  const [aps, setAps] = useState<any[]>([]);
  const [newAp, setNewAp] = useState({ ap_no: '', currency: ci.currency || 'USD', amount: 0, due_date: '' });

  const loadData = useCallback(() => {
    procureApi.getCIAggregatedItems(ci.ci_id).then(setItems);
    procureApi.getAccountPayables().then(list => {
      setAps(list.filter(ap => ap.reference_uuid === ci.uuid && ap.reference_type === 'CI'));
    });
  }, [ci.ci_id, ci.uuid]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleAddAp = async () => {
    if (!newAp.ap_no || !newAp.amount || !newAp.due_date) {
      alert("Please fill in all AP fields");
      return;
    }
    await procureApi.saveAccountPayable({
      ...newAp,
      ap_id: 0,
      vendor_id: ci.vendor_id,
      reference_uuid: ci.uuid,
      reference_type: 'CI',
      status: 'unpaid',
      allocation_status: 'Open',
      local_amount: newAp.amount // Defaulting local amount to same for simplicity
    } as any);
    setNewAp({ ap_no: '', currency: ci.currency || 'USD', amount: 0, due_date: '' });
    loadData();
  };

  return (
    <div style={{ marginTop: '2rem' }}>
      <div style={{ marginBottom: '2.5rem' }}>
        <h3 style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          Associated Account Payables
          <span style={{ fontSize: '0.8rem', opacity: 0.6 }}>Reference: {ci.ci_no}</span>
        </h3>
        <table className="sub-table" style={{ marginBottom: '1.5rem' }}>
          <thead>
            <tr>
              <th>AP No</th>
              <th>Currency</th>
              <th>Amount</th>
              <th>Due Date</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {aps.length === 0 ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', opacity: 0.6 }}>No APs linked to this CI.</td></tr>
            ) : aps.map((ap, idx) => (
              <tr key={idx}>
                <td>{ap.ap_no}</td>
                <td>{ap.currency}</td>
                <td>{ap.amount.toLocaleString()}</td>
                <td>{ap.due_date?.split('T')[0]}</td>
                <td><span className={`badge ${ap.status}`}>{ap.status}</span></td>
              </tr>
            ))}
          </tbody>
        </table>

        <div className="form-grid" style={{ background: 'rgba(255,255,255,0.03)', padding: '1.5rem', borderRadius: '12px', border: '1px solid rgba(255,255,255,0.1)' }}>
          <div className="form-group">
            <label>AP No</label>
            <input type="text" value={newAp.ap_no} onChange={e => setNewAp({...newAp, ap_no: e.target.value})} placeholder="e.g. AP-INV-001" />
          </div>
          <div className="form-group">
            <label>Currency</label>
            <input type="text" value={newAp.currency} onChange={e => setNewAp({...newAp, currency: e.target.value})} />
          </div>
          <div className="form-group">
            <label>Amount</label>
            <input type="number" value={newAp.amount} onChange={e => setNewAp({...newAp, amount: Number(e.target.value)})} />
          </div>
          <div className="form-group">
            <label>Due Date</label>
            <input type="date" value={newAp.due_date} onChange={e => setNewAp({...newAp, due_date: e.target.value})} />
          </div>
          <div className="form-group" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'flex-end' }}>
            <label>&nbsp;</label>
            <button className="btn-success" onClick={handleAddAp} style={{ width: '100%', height: '38px' }}>Add AP to Invoice</button>
          </div>
        </div>
      </div>

      <h3 style={{ borderTop: '1px solid #333', paddingTop: '1.5rem', marginBottom: '1.5rem' }}>Aggregated Loaded Items</h3>
      <table className="sub-table">
        <thead>
          <tr>
            <th>Item ID</th>
            <th>Item Name</th>
            <th>Total Qty</th>
            <th>Total Amount</th>
          </tr>
        </thead>
        <tbody>
          {(!items || items.length === 0) ? (
            <tr>
              <td colSpan={4} style={{ textAlign: 'center', opacity: 0.6 }}>No items associated with this invoice.</td>
            </tr>
          ) : items.map((it, idx) => (
            <tr key={idx}>
              <td>{it.item_id}</td>
              <td><strong>{it.item_name}</strong></td>
              <td>{it.total_qty}</td>
              <td>{it.amount.toLocaleString()} {it.currency}</td>
            </tr>
          ))}
        </tbody>
        {items && items.length > 0 && (
          <tfoot>
            <tr style={{ fontWeight: 'bold', background: 'rgba(255,255,255,0.05)' }}>
              <td colSpan={2} style={{ textAlign: 'right' }}>Total:</td>
              <td>{items.reduce((sum, it) => sum + it.total_qty, 0)}</td>
              <td>{items.reduce((sum, it) => sum + it.amount, 0).toLocaleString()} {items[0]?.currency}</td>
            </tr>
          </tfoot>
        )}
      </table>
    </div>
  );
}

function BLDetail({ bl }: { bl: any }) {
  if (!bl.bl_id) return null;
  const [containers, setContainers] = useState<any[]>([]);
  useEffect(() => {
    procureApi.getContainersByBLID(bl.bl_id).then(setContainers);
  }, [bl.bl_id]);

  return (
    <div style={{ marginTop: '2rem' }}>
      <h3>Associated Containers</h3>
      <table className="sub-table">
        <thead>
          <tr>
            <th>Container No</th>
            <th>Status</th>
            <th>Total CBM</th>
            <th>Net Wgt</th>
            <th>Gross Wgt</th>
          </tr>
        </thead>
        <tbody>
          {(!containers || containers.length === 0) ? (
            <tr>
              <td colSpan={5} style={{ textAlign: 'center', opacity: 0.6 }}>No containers associated with this B/L.</td>
            </tr>
          ) : containers.map((c, idx) => (
            <tr key={idx}>
              <td>{c.container_no}</td>
              <td>{c.status}</td>
              <td>{c.total_cbm}</td>
              <td>{c.total_net_wgt}</td>
              <td>{c.total_gross_wgt}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function ContainerDetail({ container }: { container: any }) {
  if (!container.container_id) return null;
  const [items, setItems] = useState<any[]>([]);
  useEffect(() => {
    procureApi.getContainerItemsByContainerID(container.container_id).then(setItems);
  }, [container.container_id]);

  return (
    <div style={{ marginTop: '2rem' }}>
      <h3>Loaded Items</h3>
      <table className="sub-table">
        <thead>
          <tr>
            <th>Item ID</th>
            <th>Qty</th>
            <th>Price</th>
            <th>Weights (G/N)</th>
            <th>CBM</th>
          </tr>
        </thead>
        <tbody>
          {(!items || items.length === 0) ? (
            <tr>
              <td colSpan={5} style={{ textAlign: 'center', opacity: 0.6 }}>No items loaded in this container.</td>
            </tr>
          ) : items.map((it, idx) => (
            <tr key={idx}>
              <td>{it.item_id}</td>
              <td>{it.load_qty}</td>
              <td>{it.unit_price} {it.currency}</td>
              <td>{it.gross_weight} / {it.net_weight}</td>
              <td>{it.cbm}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function VendorDetail({ vendor }: { vendor: any }) {
  if (!vendor.vendor_id) return null;
  const [unpaidAPs, setUnpaidAPs] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    procureApi.getAccountPayables()
      .then(aps => {
        const filtered = aps.filter(ap => ap.vendor_id === vendor.vendor_id && ap.status === 'unpaid');
        setUnpaidAPs(filtered);
      })
      .finally(() => setLoading(false));
  }, [vendor.vendor_id]);

  return (
    <div style={{ marginTop: '2rem', borderTop: '1px solid #333', paddingTop: '1.5rem' }}>
      <h3>Unpaid Account Payables</h3>
      {loading ? (
        <div style={{ opacity: 0.6 }}>Loading APs...</div>
      ) : (
        <table className="sub-table">
          <thead>
            <tr>
              <th>AP No</th>
              <th>Currency</th>
              <th>Amount</th>
              <th>Local Amount</th>
              <th>Due Date</th>
            </tr>
          </thead>
          <tbody>
            {unpaidAPs.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ textAlign: 'center', opacity: 0.6 }}>No unpaid APs found for this vendor.</td>
              </tr>
            ) : unpaidAPs.map((ap, idx) => (
              <tr key={idx}>
                <td><strong>{ap.ap_no}</strong></td>
                <td>{ap.currency}</td>
                <td>{ap.amount.toLocaleString()}</td>
                <td>{ap.local_amount.toLocaleString()}</td>
                <td>{ap.due_date ? new Date(ap.due_date).toLocaleDateString() : '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}


function BookingFlow({ booking }: { booking: any }) {
  const [details, setDetails] = useState<any>(null);

  useEffect(() => {
    const load = async () => {
      const data: any = {};
      if (booking.po_item_id) {
        const poItem = await procureApi.getPOItems(0); // Mock/Generic fetch
        // In a real app, we'd have a specific "getPOByItemID" or similar
        const pos = await procureApi.getPurchaseOrders();
        const po = pos.find(p => p.po_id === booking.po_id); // Assuming po_id is in booking
        data.po_no = po?.po_no || 'PO-' + booking.po_item_id;
      }
      if (booking.ci_id) {
        const cis = await procureApi.getCommercialInvoices();
        const ci = cis.find(c => c.ci_id === booking.ci_id);
        data.ci_no = ci?.ci_no || 'CI-' + booking.ci_id;
      }
      if (booking.item_id) {
        const items = await procureApi.getItems();
        const item = items.find(i => i.item_id === booking.item_id);
        data.item_name = item?.name || 'Item-' + booking.item_id;
      }
      setDetails(data);
    };
    load();
  }, [booking]);

  const Step = ({ label, value, color }: { label: string, value: string, color: string }) => (
    <div style={{ 
      flex: 1, 
      background: 'rgba(255,255,255,0.03)', 
      padding: '1rem', 
      borderRadius: '8px', 
      borderLeft: `4px solid ${color}`,
      textAlign: 'center'
    }}>
      <div style={{ fontSize: '0.7rem', color: 'var(--text-secondary)', marginBottom: '0.2rem', textTransform: 'uppercase' }}>{label}</div>
      <div style={{ fontWeight: 600, fontSize: '0.9rem' }}>{value || '-'}</div>
    </div>
  );

  const Arrow = () => (
    <div style={{ display: 'flex', alignItems: 'center', color: 'var(--text-secondary)', padding: '0 0.5rem' }}>→</div>
  );

  return (
    <div style={{ marginTop: '2rem', borderTop: '1px solid #333', paddingTop: '1.5rem' }}>
      <h3 style={{ marginBottom: '1.5rem' }}>Logistics Relationship Flow</h3>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'stretch' }}>
        <Step label="Source PO" value={details?.po_no} color="#f59e0b" />
        <Arrow />
        <Step label="PO Item" value={details?.item_name} color="#10b981" />
        <Arrow />
        <Step label="Container" value={booking.container_no} color="#3b82f6" />
        <Arrow />
        <Step label="B/L" value={booking.bl_no} color="#8b5cf6" />
        <Arrow />
        <Step label="Invoice (CI)" value={details?.ci_no} color="#ec4899" />
      </div>
      <div style={{ marginTop: '1rem', fontSize: '0.8rem', opacity: 0.6, textAlign: 'center' }}>
        Linking: PO #{details?.po_no || '...'} contains {booking.load_qty} units of {details?.item_name || '...'} loaded in Container {booking.container_no || '...'} under B/L {booking.bl_no || '...'} for Invoice {details?.ci_no || '...'}.
      </div>
    </div>
  );
}

function LandedGoodsDetail({ gr, onChange }: { gr: any, onChange: (updated: any) => void }) {
  const [lots, setLots] = useState<any[]>(gr.lots || []);
  const [containerItems, setContainerItems] = useState<any[]>([]);

  useEffect(() => {
    if (gr.gr_id) {
      procureApi.getInventoryLotsByGRID(gr.gr_id).then(setLots);
    }
  }, [gr.gr_id]);

  useEffect(() => {
    if (gr.container_id) {
      procureApi.getContainerItemsByContainerID(gr.container_id).then(setContainerItems);
    }
  }, [gr.container_id]);

  const handleAddLot = () => {
    const newLot = { lot_id: 0, container_item_id: 0, lot_no: '', expiry_date: '', qty: 0, remark: '' };
    const updated = [...lots, newLot];
    setLots(updated);
    onChange({ ...gr, lots: updated });
  };

  const handleUpdateLot = (idx: number, field: string, value: any) => {
    const updated = [...lots];
    updated[idx] = { ...updated[idx], [field]: value };
    setLots(updated);
    onChange({ ...gr, lots: updated });
  };

  const handleRemoveLot = (idx: number) => {
    const updated = lots.filter((_, i) => i !== idx);
    setLots(updated);
    onChange({ ...gr, lots: updated });
  };

  return (
    <div style={{ marginTop: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h3 style={{ margin: 0 }}>Unpacking / Inventory Lots</h3>
        <button className="secondary" onClick={handleAddLot}>+ Add Lot Line</button>
      </div>
      <table className="sub-table">
        <thead>
          <tr>
            <th>Container Item (ID/Qty)</th>
            <th>Lot No</th>
            <th>Expiry Date</th>
            <th>Qty</th>
            <th>Remark</th>
            <th style={{ width: '40px' }}></th>
          </tr>
        </thead>
        <tbody>
          {lots.length === 0 ? (
            <tr><td colSpan={5} style={{ textAlign: 'center', opacity: 0.6 }}>No lots defined. Click "Add Lot Line" to split the receipt.</td></tr>
          ) : lots.map((lot, idx) => (
            <tr key={idx}>
              <td>
                <select 
                  value={lot.container_item_id} 
                  onChange={e => handleUpdateLot(idx, 'container_item_id', Number(e.target.value))}
                  className="select-cell"
                  style={{ width: '100%', background: 'transparent', color: 'inherit', border: 'none', padding: '4px' }}
                >
                  <option value={0}>Select Item...</option>
                  {containerItems.map(item => (
                    <option key={item.container_item_id} value={item.container_item_id}>
                      ID: {item.container_item_id} (Ship Qty: {item.load_qty})
                    </option>
                  ))}
                </select>
              </td>
              <td>
                <input 
                  type="text" 
                  value={lot.lot_no} 
                  onChange={e => handleUpdateLot(idx, 'lot_no', e.target.value)}
                  style={{ width: '100%', background: 'transparent', border: 'none', color: 'inherit' }}
                  placeholder="e.g. LOT-A01"
                />
              </td>
              <td>
                <input 
                  type="date" 
                  value={lot.expiry_date?.split('T')[0] || ''} 
                  onChange={e => handleUpdateLot(idx, 'expiry_date', e.target.value)}
                  style={{ width: '100%', background: 'transparent', border: 'none', color: 'inherit' }}
                />
              </td>
              <td>
                <input 
                  type="number" 
                  value={lot.qty} 
                  onChange={e => handleUpdateLot(idx, 'qty', Number(e.target.value))}
                  style={{ width: '80px', background: 'transparent', border: 'none', color: 'inherit' }}
                />
              </td>
              <td>
                <input 
                  type="text" 
                  value={lot.remark} 
                  onChange={e => handleUpdateLot(idx, 'remark', e.target.value)}
                  style={{ width: '100%', background: 'transparent', border: 'none', color: 'inherit' }}
                />
              </td>
              <td>
                <button className="secondary" onClick={() => handleRemoveLot(idx)} style={{ padding: '2px 8px' }}>✕</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <div style={{ marginTop: '0.5rem', fontSize: '0.8rem', opacity: 0.6 }}>
        Total Lot Qty: {lots.reduce((sum, l) => sum + (l.qty || 0), 0)}
      </div>
    </div>
  );
}

export default App
