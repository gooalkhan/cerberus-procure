import React, { useState, useEffect, useCallback } from 'react';
import { POItem, PurchaseOrder, POAllocationWarning, POItemAdjustment } from '../api/models';
import { procureApi } from '../api/procureApi';
import SearchModal from './SearchModal';

function parseNumberInput(value: string): number {
  const cleaned = value.replace(/,/g, '').trim();
  if (cleaned === '' || cleaned === '.') return 0;
  const num = Number(cleaned);
  return isNaN(num) ? 0 : num;
}

interface POItemDetailProps {
  po: PurchaseOrder;
  onChange: (updatedPo: PurchaseOrder) => void;
}

const POItemDetail: React.FC<POItemDetailProps> = ({ po, onChange }) => {
  const [items, setItems] = useState<POItem[]>(po.items || []);
  const [loading, setLoading] = useState(false);
  const [aps, setAps] = useState<any[]>([]);
  const [newAp, setNewAp] = useState({ ap_no: '', currency: po.currency || 'USD', amount: 0, due_date: null as string | null });
  const [itemSearchOpen, setItemSearchOpen] = useState(false);
  const [itemMasterMap, setItemMasterMap] = useState<Map<number, { name: string; sku_code: string }>>(new Map());
  const [allocationWarning, setAllocationWarning] = useState<POAllocationWarning | null>(null);
  const [lockInfo, setLockInfo] = useState<{ locked: boolean; has_ap: boolean; has_container_item: boolean } | null>(null);
  const [adjustItem, setAdjustItem] = useState<POItem | null>(null);
  const [adjustDelta, setAdjustDelta] = useState('');
  const [adjustReason, setAdjustReason] = useState('');
  const [adjustHistory, setAdjustHistory] = useState<POItemAdjustment[]>([]);
  const [adjustSaving, setAdjustSaving] = useState(false);
  const [adjustmentsMap, setAdjustmentsMap] = useState<Map<number, POItemAdjustment[]>>(new Map());
  const [adjustRefresh, setAdjustRefresh] = useState(0);
  const [expandedItems, setExpandedItems] = useState<Set<number>>(new Set());

  const isLocked = !!lockInfo?.locked;

  const loadAdjustments = async (_list: POItem[]) => {
    if (!po.po_id) {
      setAdjustmentsMap(new Map());
      return;
    }
    try {
      const all = await procureApi.getPOItemAdjustmentsByPO(po.po_id);
      const m = new Map<number, POItemAdjustment[]>();
      all.forEach(a => {
        const arr = m.get(a.po_item_id) || [];
        arr.push(a);
        m.set(a.po_item_id, arr);
      });
      setAdjustmentsMap(m);
    } catch (e) {
      // ignore
    }
  };

  useEffect(() => {
    loadAdjustments(items);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [items.map(i => i.po_item_id).join(','), adjustRefresh]);

  useEffect(() => {
    if (po.po_id && (!po.items || po.items.length === 0)) {
      loadItems();
    }
  }, [po.po_id]);

  useEffect(() => {
    if (po.po_id) {
      procureApi.getPOAllocationWarning(po.po_id).then(setAllocationWarning).catch(() => setAllocationWarning(null));
      procureApi.getPOEditLock(po.po_id).then(setLockInfo).catch(() => setLockInfo(null));
    } else {
      setAllocationWarning(null);
      setLockInfo(null);
    }
  }, [po.po_id]);

  // Load item master data for display
  useEffect(() => {
    procureApi.getItems().then(allItems => {
      const map = new Map<number, { name: string; sku_code: string }>();
      allItems.forEach(it => map.set(it.item_id, { name: it.name, sku_code: it.sku_code }));
      setItemMasterMap(map);
    });
  }, []);

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
    const total = newItems
      .filter(item => item.status !== 'Cancelled')
      .reduce((acc, item) => acc + (item.po_qty * item.unit_price), 0);
    onChange({ ...po, items: newItems, total_amount: total });
  };

  const handleItemChange = (idx: number, field: keyof POItem, value: any) => {
    const newItems = [...items];
    newItems[idx] = { ...newItems[idx], [field]: value };
    setItems(newItems);
    updateParent(newItems);
  };

  const handleAddItem = () => {
    setItemSearchOpen(true);
  };

  const handleSelectItem = (selectedItem: any) => {
    const newItem: POItem = {
      po_item_id: 0,
      po_id: po.po_id,
      item_id: selectedItem.item_id,
      base_qty: 0,
      po_qty: 0,
      unit_price: 0,
      status: 'Not Shipped',
      remark: ''
    };
    const newItems = [...items, newItem];
    setItems(newItems);
    updateParent(newItems);
    setItemSearchOpen(false);
  };

  const handleRemoveItem = (idx: number) => {
    const newItems = items.filter((_, i) => i !== idx);
    setItems(newItems);
    updateParent(newItems);
  };

  const handleOpenAdjust = async (item: POItem) => {
    if (!item.po_item_id) {
      alert('먼저 PO를 저장해주세요.');
      return;
    }
    setAdjustItem(item);
    setAdjustDelta('');
    setAdjustReason('');
    setAdjustHistory([]);
    try {
      const hist = await procureApi.getPOItemAdjustments(item.po_item_id);
      setAdjustHistory(hist);
    } catch (e) {
      // ignore
    }
  };

  const handleAdjustConfirm = async () => {
    if (!adjustItem) return;
    const delta = parseNumberInput(adjustDelta);
    if (delta === 0) {
      alert('조정 수량(증분)을 입력해주세요. (증가 +, 감소 -)');
      return;
    }
    setAdjustSaving(true);
    try {
      await procureApi.adjustPOItemQty({
        adjustment_id: 0,
        po_item_id: adjustItem.po_item_id,
        kind: 'QTY',
        qty_delta: delta,
        old_value: 0,
        new_value: 0,
        reason: adjustReason,
      });
      setAdjustItem(null);
      await loadItems();
      setAdjustRefresh(k => k + 1);
      if (po.po_id) procureApi.getPOEditLock(po.po_id).then(setLockInfo).catch(() => {});
    } catch (e: any) {
      alert(`조정 실패: ${e.message || e}`);
    } finally {
      setAdjustSaving(false);
    }
  };

  const getItemName = (itemId: number) => {
    const item = itemMasterMap.get(itemId);
    return item ? `${item.sku_code} - ${item.name}` : `Item ID: ${itemId}`;
  };

  return (
    <div style={{ marginTop: '2rem', borderTop: '1px solid var(--border-color)', paddingTop: '1.5rem' }}>
      {allocationWarning?.has_allocation && (
        <div style={{ 
          background: '#fff7ed', 
          border: '1px solid #f97316', 
          borderRadius: '8px', 
          padding: '1rem', 
          marginBottom: '1.5rem',
          color: '#c2410c'
        }}>
          <strong>⚠️ Cost Allocation Warning</strong>
          <div style={{ marginTop: '0.25rem', fontSize: '0.9rem' }}>
            This PO has already been used in Cost Allocation.
            {' '}Allocated lots: <strong>{allocationWarning.allocated_lot_count}</strong>,
            {' '}Total allocated amount: <strong>{allocationWarning.total_allocated_amount.toLocaleString()}</strong>.
            <br />
            Changes to PO items may affect existing allocation ratios. Please proceed with caution.
          </div>
        </div>
      )}
      <div style={{ marginBottom: '0.5rem' }}>
        <h3 style={{ color: 'var(--accent-color)', margin: 0 }}>
          Current PO Items
          <span style={{ fontSize: '0.75rem', fontWeight: 400, color: 'var(--text-secondary)', marginLeft: '0.5rem' }}>
            (현행 최종 PO — Added PO Items 합산 · 읽기전용)
          </span>
        </h3>
      </div>
      {loading ? (
        <p>Loading items...</p>
      ) : (
        <div className="table-container" style={{ marginTop: '0.5rem' }}>
          <table className="sub-table">
            <thead>
              <tr>
                <th>Item</th>
                <th style={{ textAlign: 'right' }}>Qty</th>
                <th style={{ textAlign: 'right' }}>Unit Price</th>
                <th style={{ textAlign: 'right' }}>Amount</th>
                <th>Status</th>
                <th>Remark</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item, idx) => (
                <tr key={idx}>
                  <td style={{ minWidth: '180px' }}>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>ID: {item.item_id}</div>
                    <div style={{ fontWeight: 500 }}>{getItemName(item.item_id)}</div>
                  </td>
                  <td style={{ textAlign: 'right' }}>{item.po_qty?.toLocaleString()}</td>
                  <td style={{ textAlign: 'right' }}>{item.unit_price?.toLocaleString()}</td>
                  <td style={{ textAlign: 'right', fontWeight: 600 }}>{(item.po_qty * item.unit_price).toLocaleString()}</td>
                  <td>{item.status}</td>
                  <td>{item.remark || '-'}</td>
                </tr>
              ))}
              {items.length > 0 && (
                <tr style={{ background: 'rgba(14, 165, 233, 0.05)' }}>
                  <td colSpan={3} style={{ textAlign: 'right', fontWeight: 700 }}>Total:</td>
                  <td style={{ textAlign: 'right', fontWeight: 700, color: 'var(--accent-color)' }}>
                    {items.reduce((acc, i) => acc + (i.po_qty * i.unit_price), 0).toLocaleString()}
                  </td>
                  <td colSpan={2}></td>
                </tr>
              )}
              {items.length === 0 && (
                <tr>
                  <td colSpan={6} style={{ textAlign: 'center', padding: '1rem' }}>No items. Use "+ Add Item" in Added PO Items.</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Added PO Items — 입력/변경 원장 */}
      <div style={{ marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
          <h3 style={{ color: 'var(--accent-color)', margin: 0, display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
            Added PO Items
            <span style={{ fontSize: '0.75rem', fontWeight: 400, color: 'var(--text-secondary)' }}>
              (추가/증분 입력 — 사유 기록)
            </span>
            {isLocked && (
              <span style={{ fontSize: '0.75rem', fontWeight: 400, color: '#b45309' }}>
                🔒 참조(AP / Container Item)가 있어 기존 품목의 수량/단가는 조정(증분)으로만 변경됩니다 (신규 품목 추가는 가능)
              </span>
            )}
          </h3>
          <button className="secondary" style={{ padding: '0.3rem 0.8rem', fontSize: '0.8rem' }} onClick={handleAddItem}>+ Add Item</button>
        </div>
        {items.length === 0 ? (
          <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', padding: '1rem', textAlign: 'center', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
            추가된 항목이 없습니다. "+ Add Item"을 눌러 추가하세요.
          </div>
        ) : (
          <div className="table-container">
            <table className="sub-table">
              <thead>
                <tr>
                  <th>Item</th>
                  <th>구분</th>
                  <th>수량 / 단가</th>
                  <th>사유</th>
                  <th>작성자</th>
                  <th>일시</th>
                  <th style={{ width: '120px' }}></th>
                </tr>
              </thead>
              <tbody>
                {items.flatMap((item, idx) => {
                  const adj = adjustmentsMap.get(item.po_item_id) || [];
                  const expanded = expandedItems.has(item.po_item_id);
                  const toggle = () => setExpandedItems(prev => {
                    const s = new Set(prev);
                    if (s.has(item.po_item_id)) s.delete(item.po_item_id); else s.add(item.po_item_id);
                    return s;
                  });
                  const baseQty = item.base_qty ?? item.po_qty ?? 0;
                  const editable = !isLocked || item.po_item_id === 0;
                  const rows: JSX.Element[] = [];
                  rows.push(
                    <tr key={`${item.po_item_id}-base`}>
                      <td>
                        <button
                          type="button"
                          className="secondary"
                          style={{ padding: '0 0.35rem', fontSize: '0.7rem', minWidth: 'auto', marginRight: '0.35rem' }}
                          onClick={toggle}
                          disabled={adj.length === 0}
                        >
                          {expanded ? '▾' : '▸'}
                        </button>
                        {getItemName(item.item_id)}
                      </td>
                      <td>기준</td>
                      <td>
                        {!editable ? (
                          <span style={{ fontWeight: 600 }}>
                            {baseQty.toLocaleString()}
                            <span style={{ color: 'var(--text-secondary)', fontWeight: 400 }}> ({item.unit_price?.toLocaleString()})</span>
                          </span>
                        ) : (
                          <div style={{ display: 'flex', gap: '0.3rem', alignItems: 'center' }}>
                            <input
                              type="text"
                              inputMode="decimal"
                              value={item.po_qty || ''}
                              style={{ width: '70px', padding: '0.2rem' }}
                              onChange={(e) => handleItemChange(idx, 'po_qty', parseNumberInput(e.target.value))}
                            />
                            {item.po_item_id === 0 ? (
                              <input
                                type="text"
                                inputMode="decimal"
                                value={item.unit_price || ''}
                                placeholder="단가"
                                style={{ width: '80px', padding: '0.2rem' }}
                                onChange={(e) => handleItemChange(idx, 'unit_price', parseNumberInput(e.target.value))}
                              />
                            ) : (
                              <span style={{ color: 'var(--text-secondary)', minWidth: '60px' }}>{item.unit_price?.toLocaleString()}</span>
                            )}
                          </div>
                        )}
                      </td>
                      <td>
                        {editable ? (
                          <input
                            type="text"
                            value={item.remark || ''}
                            placeholder="Remark..."
                            style={{ width: '120px', padding: '0.2rem' }}
                            onChange={(e) => handleItemChange(idx, 'remark', e.target.value)}
                          />
                        ) : (item.remark || '-')}
                      </td>
                      <td>{item.created_by || '-'}</td>
                      <td>{item.created_at ? String(item.created_at).split('T')[0] : '-'}</td>
                      <td>
                        {item.po_item_id > 0 && (
                          <button
                            type="button"
                            className="secondary"
                            style={{ padding: '0.15rem 0.45rem', fontSize: '0.7rem', minWidth: 'auto', marginRight: '0.3rem' }}
                            onClick={() => handleOpenAdjust(item)}
                          >
                            조정
                          </button>
                        )}
                        {editable && (
                          <button
                            className="btn-danger secondary"
                            style={{ padding: '0.15rem 0.45rem', fontSize: '0.7rem', minWidth: 'auto' }}
                            onClick={() => handleRemoveItem(idx)}
                          >
                            ✕
                          </button>
                        )}
                      </td>
                    </tr>
                  );
                  if (expanded) {
                    adj.forEach(a => {
                      rows.push(
                        <tr key={`adj-${a.adjustment_id}`}>
                          <td></td>
                          <td>{a.kind === 'PRICE' ? '단가' : '수량'}</td>
                          <td style={{ fontWeight: 600 }}>
                            {a.kind === 'PRICE'
                              ? `${a.old_value?.toLocaleString()} → ${a.new_value?.toLocaleString()}`
                              : <span style={{ color: a.qty_delta < 0 ? '#dc2626' : '#059669' }}>{a.qty_delta > 0 ? `+${a.qty_delta.toLocaleString()}` : a.qty_delta.toLocaleString()}</span>}
                          </td>
                          <td>{a.reason || '-'}</td>
                          <td>{a.created_by || '-'}</td>
                          <td>{a.created_at ? String(a.created_at).split('T')[0] : '-'}</td>
                          <td></td>
                        </tr>
                      );
                    });
                  }
                  if (adj.length > 0) {
                    rows.push(
                      <tr key={`${item.po_item_id}-net`} style={{ background: 'var(--bg-secondary)' }}>
                        <td></td>
                        <td>현행</td>
                        <td style={{ fontWeight: 700 }}>
                          {item.po_qty?.toLocaleString()}
                          <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}> ({item.unit_price?.toLocaleString()})</span>
                        </td>
                        <td colSpan={4}></td>
                      </tr>
                    );
                  }
                  return rows;
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {itemSearchOpen && (
        <SearchModal
          type="Item"
          searchTerm=""
          onClose={() => setItemSearchOpen(false)}
          onSelect={handleSelectItem}
        />
      )}

      {adjustItem && (
        <div className="modal-overlay" onClick={() => setAdjustItem(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()} style={{ maxWidth: '560px' }}>
            <div className="modal-header">
              <h3>수량 조정 — {getItemName(adjustItem.item_id)}</h3>
              <button className="secondary" onClick={() => setAdjustItem(null)}>✕</button>
            </div>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '0.75rem' }}>
              현재 수량: <strong>{adjustItem.po_qty?.toLocaleString()}</strong>
              {'  |  '}단가: <strong>{adjustItem.unit_price?.toLocaleString()}</strong> (불변)
            </div>
            <div className="form-grid">
              <div className="form-group">
                <label>조정 수량 (증가 +, 감소 -)</label>
                <input
                  type="text"
                  inputMode="decimal"
                  value={adjustDelta}
                  onChange={e => setAdjustDelta(e.target.value)}
                  placeholder="예: 100 또는 -50"
                />
              </div>
              <div className="form-group">
                <label>사유</label>
                <input
                  type="text"
                  value={adjustReason}
                  onChange={e => setAdjustReason(e.target.value)}
                  placeholder="예: 발주 수량 증량 / 취소"
                />
              </div>
            </div>
            {adjustDelta && (
              <div style={{ fontSize: '0.85rem', marginBottom: '0.75rem' }}>
                조정 후 수량: <strong>{((adjustItem.po_qty || 0) + parseNumberInput(adjustDelta)).toLocaleString()}</strong>
              </div>
            )}
            {adjustHistory.length > 0 && (
              <div style={{ maxHeight: '150px', overflow: 'auto', borderTop: '1px solid var(--border-color)', paddingTop: '0.5rem' }}>
                <div style={{ fontSize: '0.8rem', fontWeight: 600, marginBottom: '0.25rem' }}>조정 이력</div>
                <table className="sub-table" style={{ fontSize: '0.75rem', margin: 0 }}>
                  <tbody>
                    {adjustHistory.map(h => (
                      <tr key={h.adjustment_id}>
                        <td style={{ color: h.qty_delta >= 0 ? '#059669' : '#dc2626' }}>{h.qty_delta > 0 ? `+${h.qty_delta}` : h.qty_delta}</td>
                        <td>{h.reason || '-'}</td>
                        <td>{h.created_at?.split('T')[0]}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            <div className="modal-actions">
              <button className="secondary" onClick={() => setAdjustItem(null)}>Cancel</button>
              <button className="btn-success" onClick={handleAdjustConfirm} disabled={adjustSaving}>
                {adjustSaving ? '저장 중...' : 'Confirm'}
              </button>
            </div>
          </div>
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

        <div className="form-grid" style={{ background: '#f8fafc', padding: '1.5rem', borderRadius: '12px', border: '1px solid #e2e8f0' }}>
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
            <input type="text" inputMode="decimal" value={newAp.amount} onChange={e => setNewAp({ ...newAp, amount: parseNumberInput(e.target.value) })} />
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
