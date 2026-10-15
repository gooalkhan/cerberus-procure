import React, { useState, useEffect, useMemo } from 'react';
import { procureApi } from '../api/procureApi';
import {
  CostAllocation,
  CostAllocationItem,
  CostAllocationLotCandidate,
  CostAllocationProposal,
} from '../api/models';

// AP별 배분 내역 (LOT별 금액)
interface APAllocLot {
  amount: number;
  itemId: number; // 기존 저장된 Cost_Allocation_Item ID (업데이트용, 없으면 0)
}
interface APAlloc {
  ap_id: number;
  ap_no: string;
  allocation_type: string;
  currency: string;
  local_amount: number;
  lots: { [lotId: number]: APAllocLot };
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
  // 분배 대상 Lot (이 섹션에는 저장될 항목 내용을 두지 않음)
  const [lots, setLots] = useState<number[]>([]);
  // AP별 배분 내역 (Referenced APs 편집 대상, 유저에게는 요약만 보임)
  const [apAllocs, setApAllocs] = useState<APAlloc[]>([]);
  const [editingApId, setEditingApId] = useState<number | null>(null);
  const [modalAmounts, setModalAmounts] = useState<{ [lotId: number]: string }>({});

  const [showLotSelector, setShowLotSelector] = useState(false);
  const [availableLots, setAvailableLots] = useState<CostAllocationLotCandidate[]>([]);
  const [selectedLotIds, setSelectedLotIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(false);
  const [importLoading, setImportLoading] = useState(false);
  const [apMap, setApMap] = useState<Map<number, { ap_no: string; allocation_type: string; currency: string; local_amount: number; reference_type: string; reference_uuid: string }>>(new Map());
  const [groupTypeMap, setGroupTypeMap] = useState<Map<string, string>>(new Map());

  const lotById = (id: number) => availableLots.find(l => l.lot_id === id);

  // ---- 부모로 저장될 항목(flat) 동기화 ----
  const buildItems = (apArg: APAlloc[]): CostAllocationItem[] => {
    const rows: CostAllocationItem[] = [];
    apArg.forEach(a => {
      Object.keys(a.lots).forEach(k => {
        const lotId = Number(k);
        const entry = a.lots[lotId];
        if (!entry) return;
        if (entry.amount === 0 && !entry.itemId) return;
        rows.push({
          cost_allocation_item_id: entry.itemId || 0,
          cost_allocation_id: costAllocation.cost_allocation_id || 0,
          lot_id: lotId,
          allocated_amount: entry.amount,
          ap_id: a.ap_id,
        });
      });
    });
    return rows;
  };

  const syncToParent = (apArg: APAlloc[]) => {
    const items = buildItems(apArg);
    const total = items.reduce((s, it) => s + (it.allocated_amount || 0), 0);
    onChange({ ...costAllocation, total_allocated_amount: total, items });
  };

  // ---- Lot 상세 로드 ----
  const loadLotDetails = async (lotIds: number[]) => {
    const ids = Array.from(new Set(lotIds.filter(id => id > 0)));
    if (ids.length === 0) return;
    try {
      const candidates = await procureApi.getCostAllocationLotCandidatesByIDs(ids);
      setAvailableLots(prev => {
        const map = new Map(prev.map(c => [c.lot_id, c]));
        candidates.forEach(c => map.set(c.lot_id, c));
        return Array.from(map.values());
      });
    } catch (err) {
      console.error('Failed to load lot details:', err);
    }
  };

  // 저장된 항목(flat)으로부터 Lot 목록 + AP별 내역 복원
  const buildFromItems = (items: CostAllocationItem[]) => {
    const lotSet = new Set<number>();
    const map = new Map<number, APAlloc>();
    items.forEach(it => {
      lotSet.add(it.lot_id);
      const key = it.ap_id || 0;
      if (!map.has(key)) {
        map.set(key, { ap_id: key, ap_no: '', allocation_type: '', currency: '', local_amount: 0, lots: {} });
      }
      map.get(key)!.lots[it.lot_id] = { amount: it.allocated_amount, itemId: it.cost_allocation_item_id };
    });
    const lotsArr = Array.from(lotSet);
    setLots(lotsArr);
    setApAllocs(Array.from(map.values()));
    loadLotDetails(lotsArr);
  };

  const loadExisting = async () => {
    try {
      const existing = await procureApi.getCostAllocationItems(costAllocation.cost_allocation_id);
      buildFromItems(existing);
    } catch (err) {
      console.error('Failed to load cost allocation items:', err);
    }
  };

  useEffect(() => {
    if (costAllocation.cost_allocation_id > 0) {
      loadExisting();
    } else {
      buildFromItems(costAllocation.items || []);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [costAllocation.cost_allocation_id]);

  // AP 메타정보(AP No 등) 로드
  useEffect(() => {
    procureApi.getAccountPayables().then(list => {
      const m = new Map<number, { ap_no: string; allocation_type: string; currency: string; local_amount: number; reference_type: string; reference_uuid: string }>();
      list.forEach(ap => m.set(ap.ap_id, { ap_no: ap.ap_no, allocation_type: ap.allocation_type, currency: ap.currency, local_amount: ap.local_amount, reference_type: ap.reference_type, reference_uuid: ap.reference_uuid }));
      setApMap(m);
    }).catch(() => {});
  }, []);

  // AP Target Group의 Reference_Type 로드 (그룹 AP 표시용: GROUP-PO 등)
  useEffect(() => {
    procureApi.getAPTargetGroups().then(list => {
      const m = new Map<string, string>();
      list.forEach(g => m.set(g.uuid, g.reference_type));
      setGroupTypeMap(m);
    }).catch(() => {});
  }, []);

  const apInfo = (a: APAlloc) => {
    const info = apMap.get(a.ap_id);
    const refType = info?.reference_type || '';
    const refUUID = info?.reference_uuid || '';
    let reference = refType;
    if (refType === 'AP_Target_Group') {
      reference = `GROUP-${groupTypeMap.get(refUUID) || '?'}`;
    }
    return {
      ap_no: a.ap_no || info?.ap_no || (a.ap_id ? `AP #${a.ap_id}` : '—'),
      allocation_type: a.allocation_type || info?.allocation_type || '',
      currency: a.currency || info?.currency || '',
      local_amount: a.local_amount || info?.local_amount || 0,
      reference,
    };
  };

  // ---- 분배 대상 Lot 선택 ----
  const handleOpenLotSelector = async () => {
    setShowLotSelector(true);
    setLoading(true);
    try {
      const all = await procureApi.getAvailableCostAllocationLots(costAllocation.is_late_cost_allocation || false);
      setAvailableLots(prev => {
        const map = new Map(prev.map(c => [c.lot_id, c]));
        all.forEach(c => map.set(c.lot_id, c));
        return Array.from(map.values());
      });
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
    const newLots = Array.from(new Set([...lots, ...selectedLotIds]));
    setLots(newLots);
    loadLotDetails(newLots);
    setShowLotSelector(false);
    setSelectedLotIds([]);
  };

  const handleRemoveLot = (lotId: number) => {
    const newLots = lots.filter(id => id !== lotId);
    const newApAllocs = apAllocs.map(a => {
      const l = { ...a.lots };
      delete l[lotId];
      return { ...a, lots: l };
    });
    setLots(newLots);
    setApAllocs(newApAllocs);
    syncToParent(newApAllocs);
  };

  // ---- AP 가져오기 (자동 분배 기본값 생성) ----
  const handleImportAPs = async () => {
    if (lots.length === 0) {
      alert('먼저 분배 대상 Lot을 추가해주세요.');
      return;
    }
    setImportLoading(true);
    try {
      const fetched = await procureApi.calculateCostAllocation(lots);
      const prevMap = new Map(apAllocs.map(a => [a.ap_id, a]));
      const newApAllocs: APAlloc[] = fetched.map(p => {
        const existing = prevMap.get(p.ap_id);
        const lotsMap: { [lotId: number]: APAllocLot } = {};
        p.proposals.forEach(pr => {
          if (pr.lot_id <= 0) return;
          const ex = existing?.lots[pr.lot_id];
          // 이미 편집한 값이 있으면 유지, 없으면 자동 분배 기본값
          lotsMap[pr.lot_id] = { amount: ex ? ex.amount : pr.allocated_amount, itemId: ex?.itemId || 0 };
        });
        return {
          ap_id: p.ap_id,
          ap_no: p.ap_no,
          allocation_type: p.allocation_type,
          currency: p.currency,
          local_amount: p.local_amount,
          lots: lotsMap,
        };
      });
      setApAllocs(newApAllocs);
      syncToParent(newApAllocs);
    } catch (err: any) {
      alert(`AP 불러오기 실패: ${err.message || err}`);
    } finally {
      setImportLoading(false);
    }
  };

  // ---- AP별 LOT 금액 편집 모달 ----
  const openApModal = (apId: number) => {
    const a = apAllocs.find(x => x.ap_id === apId);
    if (!a) return;
    const amounts: { [lotId: number]: string } = {};
    lots.forEach(lotId => {
      const v = a.lots[lotId]?.amount;
      amounts[lotId] = v === undefined ? '' : String(Math.round((v + Number.EPSILON) * 10000) / 10000);
    });
    setModalAmounts(amounts);
    setEditingApId(apId);
  };

  const handleApModalSave = () => {
    if (editingApId === null) return;
    const a = apAllocs.find(x => x.ap_id === editingApId);
    if (!a) { setEditingApId(null); return; }
    const newLots: { [lotId: number]: APAllocLot } = {};
    lots.forEach(lotId => {
      const raw = modalAmounts[lotId];
      const amount = (raw === undefined || raw === '') ? 0 : parseNumberInput(raw);
      const existing = a.lots[lotId];
      if (amount === 0 && !existing) return;
      newLots[lotId] = { amount, itemId: existing?.itemId || 0 };
    });
    const newApAllocs = apAllocs.map(x => x.ap_id === editingApId ? { ...x, lots: newLots } : x);
    setApAllocs(newApAllocs);
    syncToParent(newApAllocs);
    setEditingApId(null);
  };

  const apRows = useMemo(() => {
    return apAllocs.map(a => {
      const info = apInfo(a);
      const distributed = Object.values(a.lots).reduce((s, e) => s + (e.amount || 0), 0);
      return { alloc: a, ...info, distributed };
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apAllocs, apMap, groupTypeMap]);

  const grandTotal = apRows.reduce((s, r) => s + r.distributed, 0);
  const editingAlloc = editingApId !== null ? apAllocs.find(a => a.ap_id === editingApId) : undefined;
  const editingInfo = editingAlloc ? apInfo(editingAlloc) : undefined;

  return (
    <div style={{ marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1rem' }}>Target Lots (분배 대상)</h3>
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
            disabled={importLoading || lots.length === 0}
            style={{ fontSize: '0.85rem' }}
          >
            {importLoading ? '불러오는 중...' : '가져오기'}
          </button>
          <button type="button" onClick={handleOpenLotSelector} style={{ fontSize: '0.85rem' }}>
            + Add Line
          </button>
        </div>
      </div>

      {lots.length === 0 ? (
        <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', padding: '1rem', textAlign: 'center', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
          분배 대상 Lot이 없습니다. "+ Add Line" 버튼을 눌러 추가하세요.
        </div>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <table className="sub-table" style={{ margin: 0, fontSize: '0.8rem' }}>
            <thead>
              <tr>
                <th style={{ width: '50px' }}>Lot ID</th>
                <th>Lot No</th>
                <th>Item</th>
                <th>PO No</th>
                <th>BL No</th>
                <th>Container No</th>
                <th>Qty</th>
                <th style={{ width: '60px' }}></th>
              </tr>
            </thead>
            <tbody>
              {lots.map(lotId => {
                const lot = lotById(lotId);
                return (
                  <tr key={lotId}>
                    <td>{lotId}</td>
                    <td>{lot?.lot_no || '-'}</td>
                    <td>{lot?.item_name || '-'}</td>
                    <td>{lot?.po_no || '-'}</td>
                    <td>{lot?.bl_no || '-'}</td>
                    <td>{lot?.container_no || '-'}</td>
                    <td>{lot?.qty?.toLocaleString() || '-'}</td>
                    <td>
                      <button
                        type="button"
                        className="btn-danger"
                        onClick={() => handleRemoveLot(lotId)}
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
      )}

      {/* Referenced APs — 클릭하여 LOT별 금액 편집 */}
      <div style={{ marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1rem' }}>
        <h3 style={{ margin: '0 0 0.75rem 0', fontSize: '1rem' }}>
          Referenced APs
          <span style={{ fontSize: '0.75rem', fontWeight: 400, color: 'var(--text-secondary)', marginLeft: '0.5rem' }}>
            (AP 행을 클릭하면 LOT별 분배 금액을 편집할 수 있습니다)
          </span>
        </h3>
        {apRows.length === 0 ? (
          <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', padding: '1rem', textAlign: 'center', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
            '가져오기' 버튼을 누르면 관련 AP들이 여기에 표시됩니다.
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className="sub-table" style={{ margin: 0, fontSize: '0.8rem' }}>
              <thead>
                <tr>
                  <th>AP No</th>
                  <th>Reference</th>
                  <th>Allocation Type</th>
                  <th>Local Amount</th>
                  <th>Distributed Amount</th>
                </tr>
              </thead>
              <tbody>
                {apRows.map(row => (
                  <tr
                    key={row.alloc.ap_id}
                    onClick={() => openApModal(row.alloc.ap_id)}
                    style={{ cursor: 'pointer' }}
                  >
                    <td>{row.ap_no}</td>
                    <td>{row.reference || '-'}</td>
                    <td>{row.allocation_type}</td>
                    <td>{formatNumber(row.local_amount)}</td>
                    <td>{formatNumber(row.distributed)}</td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr style={{ fontWeight: 600, background: 'var(--bg-secondary)' }}>
                  <td colSpan={4} style={{ textAlign: 'right' }}>Total Allocated Amount:</td>
                  <td>{formatNumber(grandTotal)}</td>
                </tr>
              </tfoot>
            </table>
          </div>
        )}
      </div>

      {/* AP별 LOT 금액 편집 모달 */}
      {editingAlloc && editingInfo && (
        <div className="modal-overlay" onClick={() => setEditingApId(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()} style={{ maxWidth: '760px', maxHeight: '80vh', display: 'flex', flexDirection: 'column' }}>
            <div className="modal-header">
              <h2>{editingInfo.ap_no} — LOT별 분배 금액</h2>
              <button className="secondary" onClick={() => setEditingApId(null)}>✕</button>
            </div>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '0.75rem' }}>
              Allocation Type: <strong>{editingInfo.allocation_type}</strong>
              {'  |  '}Local Amount: <strong>{formatNumber(editingInfo.local_amount)} {editingInfo.currency}</strong>
            </div>
            <div style={{ flex: 1, overflow: 'auto' }}>
              <table className="sub-table" style={{ margin: 0, fontSize: '0.8rem' }}>
                <thead>
                  <tr>
                    <th style={{ width: '60px' }}>Lot ID</th>
                    <th>Lot No</th>
                    <th>Item</th>
                    <th>Qty</th>
                    <th>Allocated Amount</th>
                  </tr>
                </thead>
                <tbody>
                  {lots.map(lotId => {
                    const lot = lotById(lotId);
                    return (
                      <tr key={lotId}>
                        <td>{lotId}</td>
                        <td>{lot?.lot_no || '-'}</td>
                        <td>{lot?.item_name || '-'}</td>
                        <td>{lot?.qty?.toLocaleString() || '-'}</td>
                        <td>
                          <input
                            type="text"
                            inputMode="decimal"
                            value={modalAmounts[lotId] ?? ''}
                            onChange={(e) => setModalAmounts({ ...modalAmounts, [lotId]: e.target.value })}
                            style={{ width: '140px', fontSize: '0.8rem' }}
                          />
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            <div className="modal-actions">
              <button className="secondary" onClick={() => setEditingApId(null)}>Cancel</button>
              <button className="btn-success" onClick={handleApModalSave}>Confirm</button>
            </div>
          </div>
        </div>
      )}

      {/* Lot 선택 모달 */}
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
