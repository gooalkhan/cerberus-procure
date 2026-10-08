import React, { useState, useEffect, useCallback } from 'react'
import { User, getSession, logout } from './api/authApi'
import { procureApi } from './api/procureApi'
import Login from './components/Login'
import Sidebar from './components/Sidebar'
import CrudPage from './components/CrudPage'
import POItemDetail from './components/POItemDetail'
import SearchModal from './components/SearchModal'
import BulkBookingImport from './components/BulkBookingImport'
import AP_TargetGroupDetail from './components/APTargetGroupDetail'
import CostAllocationDetail from './components/CostAllocationDetail'

function parseNumberInput(value: string): number {
  const cleaned = value.replace(/,/g, '').trim();
  if (cleaned === '' || cleaned === '.') return 0;
  const num = Number(cleaned);
  return isNaN(num) ? 0 : num;
}

function App() {
  const [user, setUser] = useState<User | null>(null)
  const [activeMenu, setActiveMenu] = useState('items')
  const [loading, setLoading] = useState(true)
  const [showBulkImport, setShowBulkImport] = useState(false)
  const [bookingsRefreshKey, setBookingsRefreshKey] = useState(0)

  useEffect(() => {
    getSession().then(u => {
      if (u) setUser(u)
      setLoading(false)
    })
  }, [])

  const handleLogin = async (u: User) => {
    await procureApi.seedData();
    setUser(u)
  }

  if (loading) return null;

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
              { key: 'item_id', label: 'ID', formHidden: true, filterType: 'none' },
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
            tableName="Item_Master"
            idField="item_id"
          />
        )
      case 'vendors':
        return (
          <CrudPage
            title="Vendor Master"
            columns={[
              { key: 'vendor_id', label: 'ID', formHidden: true, filterType: 'none' },
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
            tableName="Vendor_Master"
            idField="vendor_id"
          />
        )
      case 'pos':
        return (
          <CrudPage
            title="Purchase Orders"
            columns={[
              { key: 'po_id', label: 'ID', formHidden: true, filterType: 'none' },
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
            tableName="Purchase_Order"
            idField="po_id"
          />
        )
      case 'logistics':
        return (
          <CrudPage
            key={`bookings-${bookingsRefreshKey}`}
            title="Bookings"
            headerActions={
              <button
                className="secondary"
                onClick={() => setShowBulkImport(true)}
                style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}
              >
                📥 Bulk Import via Excel
              </button>
            }
              columns={[
              { key: 'container_item_id', label: 'ID', formHidden: true, filterType: 'none' },
              { key: 'container_no', label: 'Container No', formHidden: true },
              { key: 'bl_no', label: 'BL No', formHidden: true },
              { key: 'status', label: 'Status', filterType: 'select', filterOptions: ['Loaded', 'Shipping', 'Arrived'], formHidden: true },
              { key: 'eta', label: 'ETA', type: 'date', formHidden: true },
              { key: 'etd', label: 'ETD', type: 'date', formHidden: true },
              { key: 'container_id', label: 'Container ID', type: 'number', tableHidden: true, searchType: 'Container' },
              { key: 'bl_id', label: 'BL ID', type: 'number', tableHidden: true, searchType: 'BL' },
              { key: 'po_item_id', label: 'PO Item ID', type: 'number', tableHidden: true, searchType: 'PO Item' },
              { key: 'ci_id', label: 'CI ID', type: 'number', tableHidden: true, searchType: 'CI' },
              { key: 'item_name', label: 'Item Name', formHidden: true },
              { key: 'divider_1', label: '', divider: true },
              { key: 'load_qty', label: 'Load Qty', type: 'number', filterType: 'none' },
              { key: 'cbm', label: 'CBM', type: 'number', formHidden: true },
              { key: 'temporary_eta', label: 'Temporary ETA', type: 'date', filterType: 'none', tableHidden: true },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getBookings}
            onSave={async (booking: any) => {
              await procureApi.saveContainerItem({
                container_item_id: booking.container_item_id,
                container_id: booking.container_id,
                bl_id: booking.bl_id,
                po_item_id: booking.po_item_id,
                ci_id: booking.ci_id || 0,
                load_qty: booking.load_qty,
                gross_weight: booking.gross_weight || 0,
                net_weight: booking.net_weight || 0,
                cbm: booking.cbm || 0,
                temporary_eta: booking.temporary_eta || null,
                uuid: booking.uuid || '',
                remark: booking.remark || '',
              });
            }}
            emptyItem={{ container_item_id: 0, container_id: 0, container_no: '', status: 'Loaded', total_cbm: 0, total_net_wgt: 0, total_gross_wgt: 0, bl_id: 0, bl_no: '', bl_status: 'Released', etd: null, eta: null, pol: '', pod: '', carrier: '', vessel_name: '', po_item_id: 0, item_id: 0, ci_id: 0, load_qty: 0, gross_weight: 0, net_weight: 0, cbm: 0, temporary_eta: null, uuid: '', remark: '' }}
            renderDetail={(booking) => <BookingFlow booking={booking} />}
            tableName="Container_Item"
            idField="container_item_id"
          />
        )
      case 'bls':
        return (
          <CrudPage
            title="BL Management"
            columns={[
              { key: 'bl_id', label: 'ID', formHidden: true, filterType: 'none' },
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
            tableName="BL"
            idField="bl_id"
          />
        )
      case 'containers':
        return (
          <CrudPage
            title="Container Master"
            columns={[
              { key: 'container_id', label: 'ID', formHidden: true, filterType: 'none' },
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
            tableName="Container"
            idField="container_id"
          />
        )
      case 'invoices':
        return (
          <CrudPage
            title="Commercial Invoices"
            columns={[
              { key: 'ci_id', label: 'ID', formHidden: true, filterType: 'none' },
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
            tableName="Commercial_Invoice"
            idField="ci_id"
          />
        )
      case 'aps':
        return (
          <CrudPage
            title="Account Payables"
            columns={[
              { key: 'ap_id', label: 'ID', formHidden: true, filterType: 'none' },
              { key: 'ap_no', label: 'AP No' },
              { key: 'vendor_id', label: 'Vendor ID', type: 'number', searchType: 'Vendor' },
              { key: 'amount', label: 'Amount', type: 'number', filterType: 'none' },
              { key: 'currency', label: 'Billing Currency' },
              { key: 'local_amount', label: 'Local Amount (Korean Won)', type: 'number', filterType: 'none' },
              { key: 'due_date', label: 'Due Date', type: 'date' },
              { key: 'date_of_payment', label: 'Payment Date', type: 'date', filterType: 'none' },
              { key: 'status', label: 'Pay Status', filterType: 'select', filterOptions: ['paid', 'unpaid'] },
              { key: 'allocation_status', label: 'Alloc Status', filterType: 'select', filterOptions: ['Draft', 'Open', 'Closed'] },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getAccountPayables}
            onSave={procureApi.saveAccountPayable}
            emptyItem={{ ap_id: 0, vendor_id: 0, ap_no: '', amount: 0, currency: 'USD', local_amount: 0, allocation_type: 'Value', reference_uuid: '', reference_type: 'PO', due_date: new Date().toISOString(), date_of_payment: null, status: 'unpaid', allocation_status: 'Draft', remark: '', uuid: '' }}
            renderDetail={(ap, onChange) => <APDetail ap={ap} onChange={onChange} />}
            tableName="Account_Payable"
            idField="ap_id"
          />
        )
      case 'ap_target_groups':
        return (
          <CrudPage
            title="AP Target Groups"
            columns={[
              { key: 'ap_target_group_id', label: 'ID', formHidden: true, filterType: 'none' },
              { key: 'group_no', label: 'Group No' },
              { key: 'group_name', label: 'Group Name' },
              { key: 'reference_type', label: 'Reference Type', type: 'select', options: ['PO', 'CI', 'BL', 'Container', 'Container Item', 'GR', 'Lot'], filterType: 'select', filterOptions: ['PO', 'CI', 'BL', 'Container', 'Container Item', 'GR', 'Lot'] },
              { key: 'status', label: 'Status', filterType: 'select', filterOptions: ['Draft', 'Open', 'Closed'] },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getAPTargetGroups}
            onSave={async (group: any) => {
              await procureApi.saveAPTargetGroup({
                ap_target_group_id: group.ap_target_group_id,
                group_no: group.group_no,
                group_name: group.group_name,
                reference_type: group.reference_type,
                status: group.status,
                remark: group.remark,
                uuid: group.uuid,
                items: group.items,
              });
            }}
            emptyItem={{ ap_target_group_id: 0, group_no: '', group_name: '', reference_type: 'PO', status: 'Draft', remark: '', uuid: '', items: [] }}
            renderDetail={(group, onChange) => <AP_TargetGroupDetail group={group} onChange={onChange} />}
            tableName="AP_Target_Group"
            idField="ap_target_group_id"
          />
        )
      case 'inventory':
        return (
          <CrudPage
            title="Landed Goods"
            columns={[
              { key: 'gr_id', label: 'ID', formHidden: true, filterType: 'none' },
              { key: 'receive_date', label: 'Receive Date', type: 'date' },
              { key: 'remark', label: 'Remark', fullWidth: true },
              { key: 'divider_lots', label: '', divider: true },
            ]}
            fetchData={procureApi.getGoodsReceipts}
            onSave={async (gr: any) => {
              await procureApi.saveGoodsReceipt(gr);
              if (gr.lots) {
                for (const lot of gr.lots) {
                  let targetGrId = gr.gr_id;
                  if (!targetGrId) {
                    const allGrs = await procureApi.getGoodsReceipts();
                    const lastGr = allGrs[allGrs.length - 1];
                    targetGrId = lastGr.gr_id;
                  }
                  await procureApi.saveInventoryLot({ ...lot, gr_id: targetGrId });
                }
              }
            }}
            emptyItem={{ gr_id: 0, receive_date: new Date().toISOString(), remark: '', lots: [] }}
            renderDetail={(gr, onChange) => <LandedGoodsDetail gr={gr} onChange={onChange} />}
            tableName="Goods_Receipt"
            idField="gr_id"
          />
        )
      case 'allocations':
        return (
          <CrudPage
            title="Cost Allocations"
            columns={[
              { key: 'cost_allocation_id', label: 'ID', formHidden: true, filterType: 'none' },
              { key: 'allocation_date', label: 'Date', type: 'date' },
              { key: 'total_allocated_amount', label: 'Total Amount', type: 'number' },
              { key: 'remark', label: 'Remark', fullWidth: true },
            ]}
            fetchData={procureApi.getCostAllocations}
            onSave={procureApi.saveCostAllocation}
            emptyItem={{ cost_allocation_id: 0, allocation_date: new Date().toISOString(), total_allocated_amount: 0, is_late_cost_allocation: false, remark: '', items: [] }}
            renderDetail={(item, onChange) => (
              <CostAllocationDetail
                costAllocation={item}
                onChange={(updated) => onChange(updated as any)}
              />
            )}
            tableName="Cost_Allocation"
            idField="cost_allocation_id"
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
          <button className="secondary" onClick={logout}>Logout</button>
        </header>
        <div className="content-area">
          {renderContent()}
        </div>
      </div>
      {showBulkImport && (
        <BulkBookingImport onComplete={() => {
          setShowBulkImport(false);
          setBookingsRefreshKey(prev => prev + 1);
        }} />
      )}
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
            {['Weight', 'Volume', 'Quantity', 'Value', 'Unit'].map(t => (
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
            {['PO', 'CI', 'Container', 'BL', 'GR', 'Lot', 'Container Item', 'AP Target Group'].map(t => (
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
  const [newAp, setNewAp] = useState({ ap_no: '', currency: ci.currency || 'USD', amount: 0, due_date: null as string | null });

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
      allocation_type: 'Value',
      local_amount: newAp.amount // Defaulting local amount to same for simplicity
    } as any);
    setNewAp({ ap_no: '', currency: ci.currency || 'USD', amount: 0, due_date: null });
    loadData();
  };

  return (
    <div style={{ marginTop: '2rem' }}>
      <div style={{ marginBottom: '2.5rem' }}>
        <h3 style={{ marginBottom: '1.5rem' }}>Aggregated Loaded Items</h3>
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
              <tr style={{ fontWeight: 'bold', background: '#f1f5f9' }}>
                <td colSpan={2} style={{ textAlign: 'right' }}>Total:</td>
                <td>{items.reduce((sum, it) => sum + it.total_qty, 0)}</td>
                <td>{items.reduce((sum, it) => sum + it.amount, 0).toLocaleString()} {items[0]?.currency}</td>
              </tr>
            </tfoot>
          )}
        </table>
      </div>

      <div>
        <h3 style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderTop: '1px solid #333', paddingTop: '1.5rem', marginBottom: '1.5rem' }}>
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

        <div className="form-grid" style={{ background: '#f8fafc', padding: '1.5rem', borderRadius: '12px', border: '1px solid #e2e8f0' }}>
          <div className="form-group">
            <label>AP No</label>
            <input type="text" value={newAp.ap_no} onChange={e => setNewAp({ ...newAp, ap_no: e.target.value })} placeholder="e.g. AP-INV-001" />
          </div>
          <div className="form-group">
            <label>Currency</label>
            <input type="text" value={newAp.currency} onChange={e => setNewAp({ ...newAp, currency: e.target.value })} />
          </div>
          <div className="form-group">
            <label>Amount</label>
            <input type="text" inputMode="decimal" value={newAp.amount} onChange={e => setNewAp({ ...newAp, amount: parseNumberInput(e.target.value) })} />
          </div>
          <div className="form-group">
            <label>Due Date</label>
            <input type="date" value={newAp.due_date || ''} onChange={e => setNewAp({ ...newAp, due_date: e.target.value })} />
          </div>
          <div className="form-group" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'flex-end' }}>
            <label>&nbsp;</label>
            <button className="btn-success" onClick={handleAddAp} style={{ width: '100%', height: '38px' }}>Add AP to Invoice</button>
          </div>
        </div>
      </div>
    </div>
  );
}

function AssociatedAPs({ referenceUuid, referenceType }: { referenceUuid: string, referenceType: string }) {
  const [aps, setAps] = useState<any[]>([]);

  useEffect(() => {
    if (!referenceUuid) return;
    procureApi.getAccountPayables().then(list => {
      setAps(list.filter(ap => ap.reference_uuid === referenceUuid && ap.reference_type === referenceType));
    });
  }, [referenceUuid, referenceType]);

  return (
    <div style={{ marginTop: '2rem' }}>
      <h3>Associated Account Payables</h3>
      <table className="sub-table">
        <thead>
          <tr>
            <th>AP No</th>
            <th>Amount</th>
            <th>Currency</th>
            <th>Local Amount</th>
            <th>Status</th>
            <th>Remark</th>
          </tr>
        </thead>
        <tbody>
          {aps.length === 0 ? (
            <tr>
              <td colSpan={6} style={{ textAlign: 'center', opacity: 0.6 }}>No APs linked to this {referenceType}.</td>
            </tr>
          ) : aps.map((ap, idx) => (
            <tr key={idx}>
              <td>{ap.ap_no}</td>
              <td style={{ textAlign: 'right' }}>{ap.amount.toLocaleString()}</td>
              <td>{ap.currency}</td>
              <td style={{ textAlign: 'right' }}>{ap.local_amount.toLocaleString()}</td>
              <td><span className={`badge ${ap.status}`}>{ap.status}</span></td>
              <td>{ap.remark}</td>
            </tr>
          ))}
        </tbody>
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
      <AssociatedAPs referenceUuid={bl.uuid} referenceType="BL" />
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
            <th>Weights (G/N)</th>
            <th>CBM</th>
          </tr>
        </thead>
        <tbody>
          {(!items || items.length === 0) ? (
            <tr>
              <td colSpan={4} style={{ textAlign: 'center', opacity: 0.6 }}>No items loaded in this container.</td>
            </tr>
          ) : items.map((it, idx) => (
            <tr key={idx}>
              <td>{it.item_id}</td>
              <td>{it.load_qty}</td>
              <td>{it.gross_weight} / {it.net_weight}</td>
              <td>{it.cbm}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <AssociatedAPs referenceUuid={container.uuid} referenceType="Container" />
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
  const Step = ({ label, value, color }: { label: string, value: string, color: string }) => (
    <div style={{
      flex: 1,
      background: '#f8fafc',
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
        <Step label="Source PO" value={booking.po_no} color="#f59e0b" />
        <Arrow />
        <Step label="PO Item" value={booking.item_name} color="#10b981" />
        <Arrow />
        <Step label="Container" value={booking.container_no} color="#3b82f6" />
        <Arrow />
        <Step label="B/L" value={booking.bl_no} color="#8b5cf6" />
        <Arrow />
        <Step label="Invoice (CI)" value={booking.ci_no} color="#ec4899" />
      </div>
      <div style={{ marginTop: '1rem', fontSize: '0.8rem', opacity: 0.6, textAlign: 'center' }}>
        Linking: PO #{booking.po_no || '...'} contains {booking.load_qty} units of {booking.item_name || '...'} loaded in Container {booking.container_no || '...'} under B/L {booking.bl_no || '...'} for Invoice {booking.ci_no || '...'}.
      </div>
    </div>
  );
}

function LandedGoodsDetail({ gr, onChange }: { gr: any, onChange: (updated: any) => void }) {
  const [lots, setLots] = useState<any[]>(gr.lots || []);
  const [containerItems, setContainerItems] = useState<any[]>([]);
  const [bls, setBls] = useState<any[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [selectedBL, setSelectedBL] = useState<number>(0);
  const [selectedItems, setSelectedItems] = useState<Set<number>>(new Set());

  useEffect(() => {
    if (gr.gr_id) {
      procureApi.getInventoryLotsByGRID(gr.gr_id).then(setLots);
    }
  }, [gr.gr_id]);

  useEffect(() => {
    procureApi.getBookings().then(setContainerItems);
    procureApi.getBLs().then(setBls);
  }, []);

  const handleAddLot = () => {
    setSelectedBL(0);
    setSelectedItems(new Set());
    setShowModal(true);
  };

  const handleConfirmAdd = () => {
    if (selectedItems.size === 0) {
      alert('Please select at least one item.');
      return;
    }
    const newLots = Array.from(selectedItems).map(containerItemId => {
      const item = containerItems.find((ci: any) => ci.container_item_id === containerItemId);
      return {
        lot_id: 0,
        container_item_id: containerItemId,
        lot_no: '',
        expiry_date: null,
        qty: item?.load_qty || 0,
        remark: '',
      };
    });
    const updated = [...lots, ...newLots];
    setLots(updated);
    onChange({ ...gr, lots: updated });
    setShowModal(false);
    setSelectedBL(0);
    setSelectedItems(new Set());
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

  const toggleItemSelection = (containerItemId: number) => {
    const next = new Set(selectedItems);
    if (next.has(containerItemId)) {
      next.delete(containerItemId);
    } else {
      next.add(containerItemId);
    }
    setSelectedItems(next);
  };

  const filteredItems = selectedBL > 0
    ? containerItems.filter((ci: any) => ci.bl_id === selectedBL)
    : [];

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
                  type="text"
                  inputMode="decimal"
                  value={lot.qty}
                  onChange={e => handleUpdateLot(idx, 'qty', parseNumberInput(e.target.value))}
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

      {showModal && (
        <div className="modal-overlay" style={{ zIndex: 1100 }} onClick={() => setShowModal(false)}>
          <div className="modal-content" style={{ maxWidth: '700px' }} onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Add Lot Lines by B/L</h3>
              <button className="secondary" onClick={() => setShowModal(false)}>✕</button>
            </div>

            <div className="form-group" style={{ marginBottom: '1rem' }}>
              <label>B/L No</label>
              <select value={selectedBL} onChange={e => { setSelectedBL(Number(e.target.value)); setSelectedItems(new Set()); }}>
                <option value={0}>Select B/L...</option>
                {bls.map((bl: any) => (
                  <option key={bl.bl_id} value={bl.bl_id}>
                    {bl.bl_no}
                  </option>
                ))}
              </select>
            </div>

            {selectedBL > 0 && (
              <div style={{ maxHeight: '300px', overflowY: 'auto', border: '1px solid var(--border-color)', borderRadius: '8px' }}>
                {filteredItems.length === 0 ? (
                  <p style={{ padding: '1rem', textAlign: 'center', opacity: 0.6 }}>No container items found for this B/L.</p>
                ) : (
                  <table className="sub-table" style={{ margin: 0 }}>
                    <thead>
                      <tr>
                        <th style={{ width: '40px' }}></th>
                        <th>Container Item ID</th>
                        <th>PO No</th>
                        <th>Item Name</th>
                        <th style={{ textAlign: 'right' }}>Ship Qty</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredItems.map((item: any) => (
                        <tr key={item.container_item_id} onClick={() => toggleItemSelection(item.container_item_id)} style={{ cursor: 'pointer' }}>
                          <td style={{ textAlign: 'center' }}>
                            <input
                              type="checkbox"
                              checked={selectedItems.has(item.container_item_id)}
                              onChange={() => toggleItemSelection(item.container_item_id)}
                              onClick={e => e.stopPropagation()}
                            />
                          </td>
                          <td>{item.container_item_id}</td>
                          <td>{item.po_no}</td>
                          <td>{item.item_name}</td>
                          <td style={{ textAlign: 'right' }}>{item.load_qty}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            )}

            <div className="modal-actions">
              <button className="secondary" onClick={() => setShowModal(false)}>Cancel</button>
              <button className="btn-success" onClick={handleConfirmAdd} disabled={selectedItems.size === 0}>
                Add {selectedItems.size > 0 ? `${selectedItems.size}` : ''} Lot Line{selectedItems.size > 1 ? 's' : ''}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default App
