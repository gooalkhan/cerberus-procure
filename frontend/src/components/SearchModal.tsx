import React, { useState, useEffect } from 'react';
import { procureApi } from '../api/procureApi';

interface SearchModalProps {
  type: string;
  searchTerm: string;
  onClose: () => void;
  onSelect: (item: any) => void;
  availableTypes?: string[];
}

const SearchModal: React.FC<SearchModalProps> = ({
  type: initialType,
  searchTerm: initialSearch,
  onClose,
  onSelect,
  availableTypes = ['PO', 'CI', 'Container', 'BL', 'GR', 'Lot', 'PO Item', 'Vendor', 'Item']
}) => {
  const [list, setList] = useState<any[]>([])
  const [search, setSearch] = useState(initialSearch)
  const [currentType, setCurrentType] = useState(initialType || 'PO')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadData()
  }, [currentType])

  const loadData = async () => {
    setLoading(true)
    try {
      let data: any[] = []
      switch (currentType) {
        case 'PO': data = await procureApi.getPurchaseOrders(); break;
        case 'CI': data = await procureApi.getCommercialInvoices(); break;
        case 'Container': data = await procureApi.getContainers(); break;
        case 'BL': data = await procureApi.getBLs(); break;
        case 'GR': data = await procureApi.getGoodsReceipts(); break;
        case 'Lot': data = await procureApi.getInventoryLots(); break;
        case 'Vendor': data = await procureApi.getVendors(); break;
        case 'Item': data = await procureApi.getItems(); break;
        case 'PO Item':
          const pos = await procureApi.getPurchaseOrders();
          const allItems: any[] = [];
          for (const po of pos) {
            const items = await procureApi.getPOItems(po.po_id);
            allItems.push(...items.map(it => ({ ...it, po_no: po.po_no, id: it.po_item_id })));
          }
          data = allItems;
          break;
      }
      setList(data)
    } finally {
      setLoading(false)
    }
  }

  const filteredList = list.filter(item => {
    const s = search.toLowerCase()
    if (!s) return true
    switch (currentType) {
      case 'PO': return (item.po_no || '').toLowerCase().includes(s)
      case 'CI': return (item.ci_no || '').toLowerCase().includes(s)
      case 'Container': return (item.container_no || '').toLowerCase().includes(s)
      case 'BL': return (item.bl_no || '').toLowerCase().includes(s)
      case 'GR': return (item.uuid || '').toLowerCase().includes(s)
      case 'Lot': return (item.lot_no || '').toLowerCase().includes(s)
      case 'Vendor': return (item.name || '').toLowerCase().includes(s) || (item.business_reg_no || '').toLowerCase().includes(s)
      case 'Item': return (item.sku_code || '').toLowerCase().includes(s) || (item.name || '').toLowerCase().includes(s)
      case 'PO Item': return (item.po_no || '').toLowerCase().includes(s) || String(item.item_id).includes(s)
      default: return false
    }
  })

  return (
    <div className="modal-overlay" style={{ zIndex: 1100 }} onClick={onClose}>
      <div className="modal-content" style={{ width: '600px', maxHeight: '80vh' }} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Select {currentType}</h2>
          <button className="secondary" onClick={onClose}>✕</button>
        </div>

        <div style={{ padding: '1rem' }}>
          <div className="form-grid" style={{ marginBottom: '1rem' }}>
            <div className="form-group">
              <label>Reference Type</label>
              <select
                value={currentType}
                onChange={e => setCurrentType(e.target.value)}
              >
                {availableTypes.map(t => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>Search Text</label>
              <input
                autoFocus
                type="text"
                placeholder={`Filter ${currentType} by Name/No...`}
                value={search}
                onChange={e => setSearch(e.target.value)}
              />
            </div>
          </div>

          <div style={{ overflowY: 'auto', maxHeight: '400px', border: '1px solid var(--border-color)', borderRadius: '4px' }}>
            {loading ? (
              <div style={{ padding: '2rem', textAlign: 'center' }}>Loading...</div>
            ) : (
              <table className="sub-table" style={{ margin: 0 }}>
                <thead>
                  <tr>
                    <th>Reference No / Name</th>
                    <th>Info</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredList.length === 0 ? (
                    <tr><td colSpan={3} style={{ textAlign: 'center', opacity: 0.6 }}>No matches found.</td></tr>
                  ) : filteredList.map((item, idx) => {
                    let no = '', info = ''
                    switch (currentType) {
                      case 'PO': no = item.po_no; info = `${(item.po_date || '').split('T')[0]} / ${item.total_amount} ${item.currency}`; break;
                      case 'CI': no = item.ci_no; info = `${(item.invoice_date || '').split('T')[0]} / ${item.total_amount} ${item.currency}`; break;
                      case 'Container': no = item.container_no; info = item.status; break;
                      case 'BL': no = item.bl_no; info = `${item.vessel_name} (ETA: ${(item.eta || '').split('T')[0]})`; break;
                      case 'GR': no = (item.uuid || '').substring(0, 8); info = (item.receive_date || '').split('T')[0]; break;
                      case 'Lot': no = item.lot_no; info = `Qty: ${item.qty}`; break;
                      case 'Vendor': no = item.name; info = `Reg: ${item.business_reg_no}`; break;
                      case 'Item': no = item.sku_code; info = item.name; break;
                      case 'PO Item': no = `${item.po_no} - Item ${item.item_id}`; info = `Qty: ${item.po_qty} / ${item.status}`; break;
                    }
                    return (
                      <tr key={idx} style={{ cursor: 'pointer' }} onClick={() => onSelect(item)}>
                        <td><strong>{no}</strong></td>
                        <td style={{ fontSize: '0.85rem', opacity: 0.8 }}>{info}</td>
                        <td><button className="secondary small">Select</button></td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default SearchModal;
