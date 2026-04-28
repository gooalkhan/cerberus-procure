import React, { useState, useEffect } from 'react';
import { POItem, PurchaseOrder } from '../api/models';
import { procureApi } from '../api/procureApi';

interface POItemDetailProps {
  po: PurchaseOrder;
  onChange: (updatedPo: PurchaseOrder) => void;
}

const POItemDetail: React.FC<POItemDetailProps> = ({ po, onChange }) => {
  const [items, setItems] = useState<POItem[]>(po.items || []);
  const [loading, setLoading] = useState(false);

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
    </div>
  );
};

export default POItemDetail;
