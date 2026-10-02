import React, { useState, useEffect, useCallback } from 'react';
import { AP_TargetGroup, AP_TargetGroupItem, APTargetGroupReference, AccountPayable } from '../api/models';
import { procureApi } from '../api/procureApi';

interface AP_TargetGroupDetailProps {
  group: AP_TargetGroup;
  onChange: (updatedGroup: AP_TargetGroup) => void;
}

export default function AP_TargetGroupDetail({ group, onChange }: AP_TargetGroupDetailProps) {
  const [items, setItems] = useState<AP_TargetGroupItem[]>(group.items || []);
  const [references, setReferences] = useState<APTargetGroupReference[]>([]);
  const [loading, setLoading] = useState(false);
  const [aps, setAps] = useState<AccountPayable[]>([]);
  const [showRefSelector, setShowRefSelector] = useState(false);
  const [selectedRefType, setSelectedRefType] = useState(group.reference_type || 'PO');
  const [selectedRefUUID, setSelectedRefUUID] = useState('');
  const [allocatedAmount, setAllocatedAmount] = useState('');

  const groupType = group.reference_type || 'PO';
  const REF_TYPES = ['PO', 'CI', 'BL', 'Container', 'Container Item', 'GR', 'Lot'];

  const loadReferences = useCallback(async () => {
    try {
      const refs = await procureApi.getAPTargetGroupReferences();
      setReferences(refs || []);
    } catch (e) {
      console.error('Failed to load references', e);
    }
  }, []);

  const loadAPs = useCallback(async () => {
    try {
      const allAps = await procureApi.getAccountPayables();
      setAps(allAps.filter(ap => ap.reference_uuid === group.uuid && ap.reference_type === 'AP_Target_Group'));
    } catch (e) {
      console.error(e);
    }
  }, [group.uuid]);

  useEffect(() => {
    if (group.ap_target_group_id && (!group.items || group.items.length === 0)) {
      loadItems();
    }
    loadReferences();
    if (group.uuid) loadAPs();
  }, [group.ap_target_group_id, group.uuid]);

  useEffect(() => {
    setSelectedRefType(groupType);
  }, [groupType]);

  const loadItems = async () => {
    setLoading(true);
    try {
      const loaded = await procureApi.getAPTargetGroups();
      const found = loaded.find(g => g.ap_target_group_id === group.ap_target_group_id);
      if (found && found.items) {
        setItems(found.items);
        onChange({ ...group, items: found.items });
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const updateParent = (newItems: AP_TargetGroupItem[]) => {
    onChange({ ...group, items: newItems });
  };

  const handleAddReference = async () => {
    if (!selectedRefUUID) {
      alert('참조를 선택해주세요.');
      return;
    }

    const targetGroupId = group.ap_target_group_id;
    if (!targetGroupId) {
      alert('먼저 그룹을 저장해주세요.');
      return;
    }

    if (selectedRefType !== groupType) {
      alert(`이 그룹은 ${groupType} 타입만 포함할 수 있습니다.`);
      return;
    }

    const newItem: AP_TargetGroupItem = {
      ap_target_group_item_id: 0,
      ap_target_group_id: targetGroupId,
      reference_uuid: selectedRefUUID,
      allocated_amount: Number(allocatedAmount) || 0,
      remark: '',
    };

    try {
      await procureApi.saveAPTargetGroupItem(newItem);
      const updated = [...items, newItem];
      setItems(updated);
      updateParent(updated);
      setShowRefSelector(false);
      setSelectedRefUUID('');
      setAllocatedAmount('');
    } catch (e: any) {
      alert(`추가 실패: ${e.message}`);
    }
  };

  const handleRemoveItem = async (idx: number, item: AP_TargetGroupItem) => {
    try {
      if (item.ap_target_group_item_id) {
        await procureApi.deleteAPTargetGroupItem(item.ap_target_group_item_id);
      }
      const updated = items.filter((_, i) => i !== idx);
      setItems(updated);
      updateParent(updated);
    } catch (e: any) {
      alert(`삭제 실패: ${e.message}`);
    }
  };

  const getRefDisplay = (uuid: string) => {
    const ref = references.find(r => r.reference_uuid === uuid);
    if (ref) {
      return `${ref.reference_no}${ref.description ? ` (${ref.description})` : ''}`;
    }
    return `${groupType}: ${uuid.substring(0, 8)}`;
  };

  return (
    <div style={{ marginTop: '2rem', borderTop: '1px solid var(--border-color)', paddingTop: '1.5rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h3 style={{ color: 'var(--accent-color)' }}>Grouped AP References</h3>
        <button className="secondary" style={{ padding: '0.3rem 0.8rem', fontSize: '0.8rem' }} onClick={() => { setSelectedRefType(groupType); setSelectedRefUUID(''); setAllocatedAmount(''); setShowRefSelector(true); }}>
          + Add Reference
        </button>
      </div>

      {loading ? (
        <p>Loading...</p>
      ) : (
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>Reference Type</th>
                <th>Reference</th>
                <th style={{ textAlign: 'right' }}>Allocated Amount</th>
                <th style={{ width: '50px' }}></th>
              </tr>
            </thead>
            <tbody>
              {items.map((item, idx) => (
                <tr key={idx}>
                  <td>{groupType}</td>
                  <td style={{ fontSize: '0.85rem' }}>{getRefDisplay(item.reference_uuid)}</td>
                  <td style={{ textAlign: 'right' }}>{item.allocated_amount?.toLocaleString()}</td>
                  <td>
                    <button className="btn-danger secondary" style={{ padding: '0.2rem 0.5rem' }} onClick={() => handleRemoveItem(idx, item)}>✕</button>
                  </td>
                </tr>
              ))}
              {items.length > 0 && (
                <tr style={{ background: 'rgba(14, 165, 233, 0.05)' }}>
                  <td colSpan={2} style={{ textAlign: 'right', fontWeight: 700 }}>Total Allocated:</td>
                  <td style={{ textAlign: 'right', fontWeight: 700, color: 'var(--accent-color)' }}>
                    {items.reduce((acc, it) => acc + (it.allocated_amount || 0), 0).toLocaleString()}
                  </td>
                  <td></td>
                </tr>
              )}
              {items.length === 0 && (
                <tr>
                  <td colSpan={4} style={{ textAlign: 'center', padding: '1rem' }}>No references grouped. Click "+ Add Reference" to add PO, CI, BL, etc.</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {showRefSelector && (
        <div className="modal-overlay" style={{ zIndex: 1100 }} onClick={() => setShowRefSelector(false)}>
          <div className="modal-content" style={{ maxWidth: '600px' }} onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Add Reference to Group</h3>
              <button className="secondary" onClick={() => setShowRefSelector(false)}>✕</button>
            </div>
            <div className="form-grid" style={{ gridTemplateColumns: '1fr' }}>
              <div className="form-group">
                <label>Reference Type</label>
                <select value={selectedRefType} onChange={e => { setSelectedRefType(e.target.value); setSelectedRefUUID(''); }}>
                  {REF_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
                </select>
                {selectedRefType !== groupType && (
                  <small style={{ color: '#fca5a5' }}>This group only allows {groupType} references.</small>
                )}
              </div>
              <div className="form-group">
                <label>{selectedRefType} Reference</label>
                <select value={selectedRefUUID} onChange={e => setSelectedRefUUID(e.target.value)}>
                  <option value="">Select {selectedRefType}...</option>
                  {references.filter(r => r.reference_type === selectedRefType).map(ref => (
                    <option key={ref.reference_uuid} value={ref.reference_uuid}>
                      {ref.reference_no} {ref.description ? `- ${ref.description}` : ''} {ref.amount > 0 ? `(${ref.amount.toLocaleString()} ${ref.currency})` : ''}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Allocated Amount (Optional)</label>
                <input type="number" value={allocatedAmount} onChange={e => setAllocatedAmount(e.target.value)} placeholder="0" />
              </div>
            </div>
            <div className="modal-actions">
              <button className="secondary" onClick={() => setShowRefSelector(false)}>Cancel</button>
              <button className="btn-success" onClick={handleAddReference}>Add Reference</button>
            </div>
          </div>
        </div>
      )}

      {/* Associated APs */}
      <div style={{ marginTop: '2rem' }}>
        <h3 style={{ color: 'var(--accent-color)', marginBottom: '1rem' }}>Associated Account Payables</h3>
        <table className="sub-table">
          <thead>
            <tr>
              <th>AP No</th>
              <th>Amount</th>
              <th>Currency</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {aps.length === 0 ? (
              <tr><td colSpan={4} style={{ textAlign: 'center', opacity: 0.6 }}>No APs linked to this group. Create an AP with Reference Type "AP Target Group".</td></tr>
            ) : aps.map((ap, idx) => (
              <tr key={idx}>
                <td>{ap.ap_no}</td>
                <td>{ap.amount.toLocaleString()}</td>
                <td>{ap.currency}</td>
                <td><span className={`badge ${ap.status}`}>{ap.status}</span></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
