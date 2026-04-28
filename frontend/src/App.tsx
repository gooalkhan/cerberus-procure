import React, { useState, useEffect } from 'react'
import { User } from './api/authApi'
import { procureApi } from './api/procureApi'
import Login from './components/Login'
import Sidebar from './components/Sidebar'
import CrudPage, { Column } from './components/CrudPage'
import POItemDetail from './components/POItemDetail'

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
              { key: 'vendor_id', label: 'Vendor ID', type: 'number' },
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
            ]}
            fetchData={procureApi.getVendors}
            onSave={procureApi.saveVendor}
            emptyItem={{ vendor_id: 0, name: '', category: 'Supplier', business_reg_no: '', bank_account: '', remark: '' }}
          />
        )
      case 'pos':
        return (
          <CrudPage 
            title="Purchase Orders"
            columns={[
              { key: 'po_no', label: 'PO No' },
              { key: 'po_date', label: 'PO Date', type: 'date' },
              { key: 'vendor_id', label: 'Vendor ID', type: 'number' },
              { key: 'currency', label: 'Currency' },
              { key: 'total_amount', label: 'Total Amount', type: 'number', filterType: 'none' },
              { key: 'status', label: 'Status', filterType: 'select', filterOptions: ['Open', 'Closed'] },
            ]}
            fetchData={procureApi.getPurchaseOrders}
            onSave={procureApi.savePurchaseOrder}
            emptyItem={{ po_id: 0, po_no: '', po_date: new Date().toISOString(), vendor_id: 0, currency: 'USD', total_amount: 0, status: 'Open', uuid: '' }}
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
              { key: 'container_id', label: 'Container ID', type: 'number', tableHidden: true },
              { key: 'bl_id', label: 'BL ID', type: 'number', tableHidden: true },
              { key: 'po_item_id', label: 'PO Item ID', type: 'number', tableHidden: true },
              { key: 'ci_id', label: 'CI ID', type: 'number', tableHidden: true },
              { key: 'item_id', label: 'Item ID', type: 'number', tableHidden: true, formHidden: true },
              { key: 'load_qty', label: 'Load Qty', type: 'number', filterType: 'none' },
              { key: 'unit_price', label: 'Price', type: 'number', filterType: 'none' },
              { key: 'currency', label: 'Currency' },
              { key: 'gross_weight', label: 'Gross Wgt', type: 'number', formHidden: true },
              { key: 'net_weight', label: 'Net Wgt', type: 'number', formHidden: true },
              { key: 'cbm', label: 'CBM', type: 'number', formHidden: true },
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
              });
            }}
            emptyItem={{ container_item_id: 0, container_id: 0, container_no: '', status: 'Loaded', total_cbm: 0, total_net_wgt: 0, total_gross_wgt: 0, bl_id: 0, bl_no: '', bl_status: 'Released', etd: '', eta: '', pol: '', pod: '', carrier: '', vessel_name: '', po_item_id: 0, item_id: 0, ci_id: 0, load_qty: 0, unit_price: 0, currency: 'USD', gross_weight: 0, net_weight: 0, cbm: 0 }}
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
              { key: 'remark', label: 'Remark' },
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
              { key: 'remark', label: 'Remark' },
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
              { key: 'vendor_id', label: 'Vendor ID', type: 'number' },
              { key: 'currency', label: 'Currency' },
              { key: 'total_amount', label: 'Total Amount', type: 'number' },
              { key: 'status', label: 'Status', filterType: 'select', filterOptions: ['Draft', 'Open', 'Closed'] },
            ]}
            fetchData={procureApi.getCommercialInvoices}
            onSave={procureApi.saveCommercialInvoice}
            emptyItem={{ ci_id: 0, ci_no: '', invoice_date: new Date().toISOString(), vendor_id: 0, currency: 'USD', total_amount: 0, status: 'Draft', uuid: '' }}
            renderDetail={(ci) => <CIDetail ci={ci} />}
          />
        )
      case 'aps':
        return (
          <CrudPage 
            title="Account Payables"
            columns={[
              { key: 'ap_no', label: 'AP No' },
              { key: 'vendor_id', label: 'Vendor ID', type: 'number' },
              { key: 'amount', label: 'Amount', type: 'number' },
              { key: 'currency', label: 'Currency' },
              { key: 'local_amount', label: 'Local Amount', type: 'number' },
              { key: 'due_date', label: 'Due Date', type: 'date' },
              { key: 'allocation_status', label: 'Status', filterType: 'select', filterOptions: ['Draft', 'Open', 'Closed'] },
            ]}
            fetchData={procureApi.getAccountPayables}
            onSave={procureApi.saveAccountPayable}
            emptyItem={{ ap_id: 0, vendor_id: 0, ap_no: '', amount: 0, currency: 'USD', local_amount: 0, allocation_type: 'Amount', reference_uuid: '', reference_type: 'PO', due_date: new Date().toISOString(), allocation_status: 'Draft', uuid: '' }}
            renderDetail={(ap, onChange) => <APDetail ap={ap} onChange={onChange} />}
          />
        )
      case 'inventory':
        return (
          <CrudPage 
            title="Inventory (GR/Lot)"
            columns={[
              { key: 'lot_no', label: 'Lot No' },
              { key: 'qty', label: 'Qty', type: 'number' },
              { key: 'landed_cost_per_unit', label: 'Landed Cost', type: 'number' },
              { key: 'quarantine_status', label: 'Status' },
            ]}
            fetchData={procureApi.getInventoryLots}
            onSave={procureApi.saveInventoryLot}
            emptyItem={{ lot_id: 0, gr_id: 0, container_item_id: 0, lot_no: '', qty: 0, landed_cost_per_unit: 0, quarantine_status: 'Pending', uuid: '' }}
          />
        )
      case 'allocations':
        return (
          <CrudPage 
            title="Cost Allocations"
            columns={[
              { key: 'allocation_date', label: 'Date', type: 'date' },
              { key: 'total_allocated_amount', label: 'Total Amount', type: 'number' },
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
  const [status, setStatus] = useState('')

  const handleSearch = async () => {
    setStatus('Searching...')
    try {
      let foundUuid = ''
      switch (ap.reference_type) {
        case 'PO':
          const pos = await procureApi.getPurchaseOrders()
          const po = pos.find(p => p.po_no === refNo)
          if (po) foundUuid = po.uuid
          break
        case 'CI':
          const cis = await procureApi.getCommercialInvoices()
          const ci = cis.find(c => c.ci_no === refNo)
          if (ci) foundUuid = ci.uuid
          break
        case 'Container':
          const containers = await procureApi.getContainers()
          const cont = containers.find(c => c.container_no === refNo)
          if (cont) foundUuid = cont.uuid
          break
        case 'BL':
          const bls = await procureApi.getBLs()
          const bl = bls.find(b => b.bl_no === refNo)
          if (bl) foundUuid = bl.uuid
          break
        // Add more as needed (GR, Lot)
      }

      if (foundUuid) {
        onChange({ ...ap, reference_uuid: foundUuid })
        setStatus(`Found UUID: ${foundUuid}`)
      } else {
        setStatus('Reference not found.')
      }
    } catch (e) {
      setStatus('Error during search.')
    }
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
          <label>Reference No Search</label>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <input 
              type="text" 
              placeholder={`Enter ${ap.reference_type} No`}
              value={refNo}
              onChange={e => setRefNo(e.target.value)}
            />
            <button type="button" className="secondary" onClick={handleSearch}>Find UUID</button>
          </div>
          {status && <div style={{ fontSize: '0.85rem', marginTop: '0.25rem', opacity: 0.8 }}>{status}</div>}
        </div>
        <div className="form-group">
          <label>Resolved Reference UUID</label>
          <input type="text" value={ap.reference_uuid} readOnly style={{ opacity: 0.6 }} />
        </div>
      </div>
    </div>
  )
}

function CIDetail({ ci }: { ci: any }) {
  if (!ci.ci_id) return null;
  const [items, setItems] = useState<any[]>([]);
  useEffect(() => {
    procureApi.getCIAggregatedItems(ci.ci_id).then(setItems);
  }, [ci.ci_id]);

  return (
    <div style={{ marginTop: '2rem' }}>
      <h3>Aggregated Loaded Items</h3>
      <table className="sub-table">
        <thead>
          <tr>
            <th>Item ID</th>
            <th>Total Qty</th>
            <th>Total Amount</th>
          </tr>
        </thead>
        <tbody>
          {(!items || items.length === 0) ? (
            <tr>
              <td colSpan={3} style={{ textAlign: 'center', opacity: 0.6 }}>No items associated with this invoice.</td>
            </tr>
          ) : items.map((it, idx) => (
            <tr key={idx}>
              <td>{it.item_id}</td>
              <td>{it.total_qty}</td>
              <td>{it.amount.toLocaleString()} {it.currency}</td>
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

export default App
