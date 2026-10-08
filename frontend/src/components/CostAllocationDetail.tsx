import React, { useState, useEffect, useMemo } from 'react';
import { procureApi } from '../api/procureApi';
import {
  CostAllocation,
  CostAllocationItem,
  CostAllocationLotCandidate,
  CostAllocationProposal,
} from '../api/models';

interface APDistributionRow {
  ap_id: number;
  ap_no: string;
  allocation_type: string;
  local_amount: number;
  distributed_amount: number;
  currency: string;
}

interface CostAllocationDetailProps {
  costAllocation: CostAllocation;
  onChange: (updated: CostAllocation) => void;
}

function formatNumber(value: number | undefined | null): string {
  if (value === undefined || value === null) return '';
  return value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function parseNumberInput(value: string): number {
  const cleaned = value.replace(/,/g, '').trim();
  if (cleaned === '' || cleaned === '.') return 0;
  const num = Number(cleaned);
  return isNaN(num) ? 0 : num;
}

const CostAllocationDetail: React.FC<CostAllocationDetailProps> = ({ costAllocation, onChange }) => {
  const [items, setItems] = useState<CostAllocationItem[]>(costAllocation.items || []);
  const [showLotSelector, setShowLotSelector] = useState(false);
  const [availableLots, setAvailableLots] = useState<CostAllocationLotCandidate[]>([]);
  const [selectedLotIds, setSelectedLotIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(false);
  const [importLoading, setImportLoading] = useState(false);
  const [proposals, setProposals] = useState<CostAllocationProposal[]>([]);

  const totalAllocatedAmount = useMemo(() => {
    return items.reduce((sum, item) => sum + (item.allocated_amount || 0), 0);
  }, [items]);

  const apDistribution = useMemo<APDistributionRow[]>(() => {
    const originalTotal = proposals.reduce((sum, p) => sum + p.proposals.reduce((s, prop) => s + prop.allocated_amount, 0), 0);
    const scale = originalTotal > 0 ? totalAllocatedAmount / originalTotal : 0;
    return proposals.map(p => {
      const originalDistributed = p.proposals.reduce((sum, prop) => sum + prop.allocated_amount, 0);
      return {
        ap_id: p.ap_id,
        ap_no: p.ap_no,
        allocation_type: p.allocation_type,
        local_amount: p.local_amount,
        distributed_amount: originalTotal > 0 ? originalDistributed * scale : 0,
        currency: p.currency,
      };
    });
  }, [proposals, totalAllocatedAmount]);

  const apDistributionTotal = useMemo(() => {
    return apDistribution.reduce((sum, row) => sum + row.distributed_amount, 0);
  }, [apDistribution]);

  useEffect(() => {
    setItems(costAllocation.items || []);
    if (costAllocation.cost_allocation_id > 0 && (!costAllocation.items || costAllocation.items.length === 0)) {
      loadExistingItems();
    }
  }, [costAllocation.cost_allocation_id]);

  useEffect(() => {
    const lotIds = items.map(i => i.lot_id).filter(id => id > 0);
    if (lotIds.length > 0) {
      loadLotDetails(lotIds);
    }
  }, [items.map(i => i.lot_id).join(',')]);

  const loadExistingItems = async () => {
    try {
      const existingItems = await procureApi.getCostAllocationItems(costAllocation.cost_allocation_id);
      setItems(existingItems);
      onChange({ ...costAllocation, items: existingItems });
    } catch (err) {
      console.error('Failed to load cost allocation items:', err);
    }
  };

  const loadLotDetails = async (lotIds: number[]) => {
    try {
      const candidates = await procureApi.getCostAllocationLotCandidatesByIDs(lotIds);
      setAvailableLots(prev => {
        const map = new Map(prev.map(c => [c.lot_id, c]));
        candidates.forEach(c => map.set(c.lot_id, c));
        return Array.from(map.values());
      });
    } catch (err) {
      console.error('Failed to load lot details:', err);
    }
  };

  const updateItems = (newItems: CostAllocationItem[]) => {
    setItems(newItems);
    const total = newItems.reduce((sum, item) => sum + (item.allocated_amount || 0), 0);
    onChange({ ...costAllocation, total_allocated_amount: total, items: newItems });
  };

  const handleOpenLotSelector = async () => {
    setShowLotSelector(true);
    setLoading(true);
    try {
      const lots = await procureApi.getAvailableCostAllocationLots(costAllocation.is_late_cost_allocation || false);
      setAvailableLots(lots);
    } catch (err) {
      alert('로트 목록을 불러오는데 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleLot = (lotId: number) => {
    setSelectedLotIds(prev =>
      prev.includes(lotId) ? prev.filter(id => id !== lotId) : [...prev, lotId]
    );
  };

  const handleConfirmLots = () => {
    const existingLotIds = new Set(items.map(i => i.lot_id));
    const newItems: CostAllocationItem[] = [];
    selectedLotIds.forEach(lotId => {
      if (!existingLotIds.has(lotId)) {
        newItems.push({
          cost_allocation_item_id: 0,
          cost_allocation_id: costAllocation.cost_allocation_id || 0,
          lot_id: lotId,
          allocated_amount: 0,
          ap_id: 0,
        });
      }
    });
    updateItems([...items, ...newItems]);
    setShowLotSelector(false);
    setSelectedLotIds([]);
  };

  const handleImportAPs = async () => {
    if (items.length === 0) {
      alert('먼저 로트 아이템을 추가해주세요.');
      return;
    }
    setImportLoading(true);
    try {
      const lotIds = items.map(i => i.lot_id);
      const fetchedProposals = await procureApi.calculateCostAllocation(lotIds, costAllocation.is_late_cost_allocation || false);
      setProposals(fetchedProposals);
      mergeProposals(fetchedProposals);
    } catch (err: any) {
      alert(`AP 불러오기 실패: ${err.message || err}`);
    } finally {
      setImportLoading(false);
    }
  };

  const mergeProposals = (proposals: CostAllocationProposal[]) => {
    const lotAmountMap: { [lotId: number]: number } = {};
    proposals.forEach(proposal => {
      proposal.proposals.forEach(p => {
        lotAmountMap[p.lot_id] = (lotAmountMap[p.lot_id] || 0) + p.allocated_amount;
      });
    });

    const updatedItems = items.map(item => ({
      ...item,
      allocated_amount: lotAmountMap[item.lot_id] || item.allocated_amount,
    }));

    updateItems(updatedItems);
  };

  const handleAmountChange = (index: number, value: string) => {
    const updated = [...items];
    updated[index].allocated_amount = parseNumberInput(value);
    updateItems(updated);
  };

  const handleRemoveItem = (index: number) => {
    const updated = items.filter((_, i) => i !== index);
    updateItems(updated);
  };



  return (
    <div style={{ marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1rem' }}>Cost Allocation Items</h3>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', marginTop: '0.5rem', fontSize: '0.85rem', color: 'var(--text-secondary)', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={costAllocation.is_late_cost_allocation || false}
              onChange={(e) => onChange({ ...costAllocation, is_late_cost_allocation: e.target.checked })}
            />
            Late Cost Allocation (이미 배분된 Lot도 추가 가능)
          </label>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            type="button"
            className="secondary"
            onClick={handleImportAPs}
            disabled={importLoading || items.length === 0}
            style={{ fontSize: '0.85rem' }}
          >
            {importLoading ? '불러오는 중...' : '가져오기'}
          </button>
          <button
            type="button"
            onClick={handleOpenLotSelector}
            style={{ fontSize: '0.85rem' }}
          >
            + Add Line
          </button>
        </div>
      </div>

      {items.length === 0 ? (
        <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', padding: '1rem', textAlign: 'center', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
          추가된 로트 아이템이 없습니다. "+ Add Line" 버튼을 눌러 추가하세요.
        </div>
      ) : (
        <>
          <div style={{ overflowX: 'auto' }}>
            <table className="sub-table" style={{ margin: 0, fontSize: '0.8rem' }}>
              <thead>
                <tr>
                  <th style={{ width: '50px' }}>Lot ID</th>
                  <th>Item</th>
                  <th>PO No</th>
                  <th>BL No</th>
                  <th>Container No</th>
                  <th>Qty</th>
                  <th>Allocated Amount</th>
                  <th style={{ width: '60px' }}></th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, index) => {
                  const lot = availableLots.find(l => l.lot_id === item.lot_id);
                  return (
                    <tr key={index}>
                      <td>{item.lot_id}</td>
                      <td>{lot?.item_name || '-'}</td>
                      <td>{lot?.po_no || '-'}</td>
                      <td>{lot?.bl_no || '-'}</td>
                      <td>{lot?.container_no || '-'}</td>
                      <td>{lot?.qty.toLocaleString() || '-'}</td>
                      <td>
                        <input
                          type="text"
                          inputMode="decimal"
                          value={formatNumber(item.allocated_amount)}
                          onChange={(e) => handleAmountChange(index, e.target.value)}
                          style={{ width: '120px', fontSize: '0.8rem' }}
                        />
                      </td>
                      <td>
                        <button
                          type="button"
                          className="btn-danger"
                          onClick={() => handleRemoveItem(index)}
                          style={{ padding: '2px 6px', fontSize: '0.7rem', minWidth: 'auto' }}
                        >
                          ✕
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
          <div style={{ textAlign: 'right', marginTop: '0.75rem', fontWeight: 600, fontSize: '0.9rem' }}>
            Total Allocated Amount: {formatNumber(totalAllocatedAmount)}
          </div>
        </>
      )}

      <div style={{ marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1rem' }}>
        <h3 style={{ margin: '0 0 0.75rem 0', fontSize: '1rem' }}>Referenced APs</h3>
        {apDistribution.length === 0 ? (
          <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', padding: '1rem', textAlign: 'center', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
            '가져오기' 버튼을 누른 관련 AP들이 여기에 표시됩니다.
          </div>
        ) : (
          <>
            <div style={{ overflowX: 'auto' }}>
              <table className="sub-table" style={{ margin: 0, fontSize: '0.8rem' }}>
                <thead>
                  <tr>
                    <th>AP ID</th>
                    <th>AP No</th>
                    <th>Allocation Type</th>
                    <th>Currency</th>
                    <th>Local Amount</th>
                    <th>Distributed Amount</th>
                  </tr>
                </thead>
                <tbody>
                  {apDistribution.map(row => (
                    <tr key={row.ap_id}>
                      <td>{row.ap_id}</td>
                      <td>{row.ap_no}</td>
                      <td>{row.allocation_type}</td>
                      <td>{row.currency}</td>
                      <td>{formatNumber(row.local_amount)}</td>
                      <td>{formatNumber(row.distributed_amount)}</td>
                    </tr>
                  ))}
                </tbody>
                <tfoot>
                  <tr style={{ fontWeight: 600, background: 'var(--bg-secondary)' }}>
                    <td colSpan={5} style={{ textAlign: 'right' }}>Total Allocated Amount:</td>
                    <td>{formatNumber(apDistributionTotal)}</td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </>
        )}
      </div>

      {showLotSelector && (
        <div className="modal-overlay" onClick={() => setShowLotSelector(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()} style={{ maxWidth: '900px', maxHeight: '80vh', display: 'flex', flexDirection: 'column' }}>
            <div className="modal-header">
              <h2>Select Lot Items</h2>
              <button className="secondary" onClick={() => setShowLotSelector(false)}>✕</button>
            </div>
            <div style={{ flex: 1, overflow: 'auto', margin: '1rem 0' }}>
              {loading ? (
                <div style={{ textAlign: 'center', padding: '2rem' }}>Loading...</div>
              ) : availableLots.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>
                  사용 가능한 로트 아이템이 없습니다.
                </div>
              ) : (
                <table className="sub-table" style={{ margin: 0, fontSize: '0.8rem' }}>
                  <thead>
                    <tr>
                      <th style={{ width: '40px' }}></th>
                      <th>Lot ID</th>
                      <th>Lot No</th>
                      <th>Item</th>
                      <th>PO No</th>
                      <th>BL No</th>
                      <th>Container No</th>
                      <th>Qty</th>
                    </tr>
                  </thead>
                  <tbody>
                    {availableLots.map(lot => (
                      <tr
                        key={lot.lot_id}
                        onClick={() => handleToggleLot(lot.lot_id)}
                        style={{ cursor: 'pointer', background: selectedLotIds.includes(lot.lot_id) ? 'rgba(59, 130, 246, 0.1)' : undefined }}
                      >
                        <td style={{ textAlign: 'center' }}>
                          <input
                            type="checkbox"
                            checked={selectedLotIds.includes(lot.lot_id)}
                            onChange={() => handleToggleLot(lot.lot_id)}
                            onClick={e => e.stopPropagation()}
                          />
                        </td>
                        <td>{lot.lot_id}</td>
                        <td>{lot.lot_no}</td>
                        <td>{lot.item_name}</td>
                        <td>{lot.po_no}</td>
                        <td>{lot.bl_no}</td>
                        <td>{lot.container_no}</td>
                        <td>{lot.qty.toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            <div className="modal-actions">
              <button className="secondary" onClick={() => setShowLotSelector(false)}>Cancel</button>
              <button
                className="btn-success"
                onClick={handleConfirmLots}
                disabled={selectedLotIds.length === 0}
              >
                Confirm ({selectedLotIds.length})
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default CostAllocationDetail;
