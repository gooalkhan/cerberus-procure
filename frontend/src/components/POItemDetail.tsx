import React, { useState, useEffect, useCallback } from 'react';
import { POItem, PurchaseOrder } from '../api/models';
import { procureApi } from '../api/procureApi';

interface POItemDetailProps {
  po: PurchaseOrder;
  onChange: (updatedPo: PurchaseOrder) => void;
}

const POItemDetail: React.FC<POItemDetailProps> = ({ po, onChange }) => {
  const [items, setItems] = useState<POItem[]>(po.items || []);
  const [loading, setLoading] = useState(false);
  const [aps, setAps] = useState<any[]>([]);
  const [newAp, setNewAp] = useState({ ap_no: '', currency: po.currency || 'USD', amount: 0, due_date: null as string | null });

  useEffect(() => {
    if (po.po_id && (!po.items || po.items.length === 0)) {
      loadItems();
    }
  }, [po.po_id]);

  const loadItems = async () => {
    setLoading(true);
    try {
      const res = await procureApi.getPOItems(po.po_id);
      setItems(res || []);
      updateParent(res || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const loadAPs = useCallback(() => {
    procureApi.getAccountPayables().then(list => {
      setAps(list.filter(ap => ap.reference_uuid === po.uuid && ap.reference_type === 'PO'));
    });
  }, [po.uuid]);

  useEffect(() => {
    if (po.uuid) {
      loadAPs();
    }
  }, [po.uuid, loadAPs]);

  const handleAddAp = async () => {
    if (!newAp.ap_no || !newAp.amount || !newAp.due_date) {
      alert("Please fill in all AP fields");
      return;
    }
    await procureApi.saveAccountPayable({
      ...newAp,
      ap_id: 0,
      vendor_id: po.vendor_id,
      reference_uuid: po.uuid,
      reference_type: 'PO',
      status: 'unpaid',
      allocation_status: 'Open',
      allocation_type: 'Value',
      local_amount: newAp.amount
    } as any);
    setNewAp({ ap_no: '', currency: po.currency || 'USD', amount: 0, due_date: null });
    loadAPs();
  };

  const updateParent = (newItems: POItem[]) => {
    const total = newItems.reduce((acc, item) => acc + (item.po_qty * item.unit_price), 0);
    onChange({ ...po, items: newItems, total_amount: total });
  };

  const handleItemChange = (idx: number, field: keyof POItem, value: any) => {
    const newItems = [...items];
    newItems[idx] = { ...newItems[idx], [field]: value };
    setItems(newItems);
    updateParent(newItems);
  };

  const handleAddItem = () => {
    const newItem: POItem = {
      po_item_id: 0,
      po_id: po.po_id,
      item_id: 0,
      po_qty: 0,
      unit_price: 0,
      status: 'Not Shipped',
      remark: ''
    };
    const newItems = [...items, newItem];
    setItems(newItems);
    updateParent(newItems);
  };

  const handleRemoveItem = (idx: number) => {
    const newItems = items.filter((_, i) => i !== idx);
    setItems(newItems);
    updateParent(newItems);
  };

  return (
    <div style={{ marginTop: '2rem', borderTop: '1px solid var(--border-color)', paddingTop: '1.5rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h3 style={{ color: 'var(--accent-color)' }}>PO Items</h3>
        <button className="secondary" style={{ padding: '0.3rem 0.8rem', fontSize: '0.8rem' }} onClick={handleAddItem}>+ Add Item</button>
      </div>
      {loading ? (
        <p>Loading items...</p>
      ) : (
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>Item ID</th>
                <th>Qty</th>
                <th>Price</th>
                <th>Amount</th>
                <th>Status</th>
                <th>Remark</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {items.map((item, idx) => (
                <tr key={idx}>
                  <td>
                    <input 
                      type="number" 
                      value={item.item_id || ''} 
                      style={{ width: '80px', padding: '0.3rem' }}
                      onChange={(e) => handleItemChange(idx, 'item_id', Number(e.target.value))}
                    />
                  </td>
                  <td>
                    <input 
                      type="number" 
                      value={item.po_qty || ''} 
                      style={{ width: '80px', padding: '0.3rem' }}
                      onChange={(e) => handleItemChange(idx, 'po_qty', Number(e.target.value))}
                    />
                  </td>
                  <td>
                    <input 
                      type="number" 
                      value={item.unit_price || ''} 
                      style={{ width: '100px', padding: '0.3rem' }}
                      onChange={(e) => handleItemChange(idx, 'unit_price', Number(e.target.value))}
                    />
                  </td>
                  <td style={{ textAlign: 'right', paddingRight: '1rem', fontWeight: 600 }}>
                    {(item.po_qty * item.unit_price).toLocaleString()}
                  </td>
                  <td>
                    <select 
                      value={item.status || 'Not Shipped'} 
                      style={{ padding: '0.3rem' }}
                      onChange={(e) => handleItemChange(idx, 'status', e.target.value)}
                    >
                      <option value="Not Shipped">Not Shipped</option>
                      <option value="Shipped">Shipped</option>
                      <option value="Partially Shipped">Partially Shipped</option>
                      <option value="Cancelled">Cancelled</option>
                    </select>
                  </td>
                  <td>
                    <input 
                      type="text" 
                      value={item.remark || ''} 
                      placeholder="Remark..."
                      style={{ width: '120px', padding: '0.3rem' }}
                      onChange={(e) => handleItemChange(idx, 'remark', e.target.value)}
                    />
                  </td>
                  <td>
                    <button className="btn-danger secondary" style={{ padding: '0.2rem 0.5rem' }} onClick={() => handleRemoveItem(idx)}>✕</button>
                  </td>
                </tr>
              ))}
              {items.length > 0 && (
                <tr style={{ background: 'rgba(14, 165, 233, 0.05)' }}>
                  <td colSpan={3} style={{ textAlign: 'right', fontWeight: 700 }}>Total:</td>
                  <td style={{ textAlign: 'right', paddingRight: '1rem', fontWeight: 700, color: 'var(--accent-color)' }}>
                    {items.reduce((acc, i) => acc + (i.po_qty * i.unit_price), 0).toLocaleString()}
                  </td>
                  <td colSpan={3}></td>
                </tr>
              )}
              {items.length === 0 && (
                <tr>
                  <td colSpan={7} style={{ textAlign: 'center', padding: '1rem' }}>No items added</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Associated Account Payables Section */}
      <div style={{ marginTop: '3rem' }}>
        <h3 style={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center',
          borderTop: '1px solid var(--border-color)', 
          paddingTop: '1.5rem', 
          marginBottom: '1.5rem' 
        }}>
          Associated Account Payables
          <span style={{ fontSize: '0.8rem', opacity: 0.6, fontWeight: 400 }}>Reference: {po.po_no}</span>
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
              <tr><td colSpan={5} style={{ textAlign: 'center', opacity: 0.6 }}>No APs linked to this PO.</td></tr>
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
            <input type="text" value={newAp.ap_no} onChange={e => setNewAp({ ...newAp, ap_no: e.target.value })} placeholder="e.g. AP-PO-001" />
          </div>
          <div className="form-group">
            <label>Currency</label>
            <input type="text" value={newAp.currency} onChange={e => setNewAp({ ...newAp, currency: e.target.value })} />
          </div>
          <div className="form-group">
            <label>Amount</label>
            <input type="number" value={newAp.amount} onChange={e => setNewAp({ ...newAp, amount: Number(e.target.value) })} />
          </div>
          <div className="form-group">
            <label>Due Date</label>
            <input type="date" value={newAp.due_date || ''} onChange={e => setNewAp({ ...newAp, due_date: e.target.value })} />
          </div>
          <div className="form-group" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'flex-end' }}>
            <label>&nbsp;</label>
            <button className="btn-success" onClick={handleAddAp} style={{ width: '100%', height: '38px' }}>Add AP to PO</button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default POItemDetail;
