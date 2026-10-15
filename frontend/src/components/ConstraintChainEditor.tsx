import React, { useState, useEffect } from 'react';
import { procureApi } from '../api/procureApi';

interface Props {
  lotId: number;
  lotNo?: string;
  onClose: () => void;
  onChanged?: () => void;
}

const TYPES = ['요건승인', '검역', '통관', '불시검문검사'];
const fmt = (d: any) => (d ? String(d).split('T')[0] : '');
const toISO = (v: string) => (v ? `${v}T00:00:00Z` : null);

const ConstraintChainEditor: React.FC<Props> = ({ lotId, lotNo, onClose, onChanged }) => {
  const [insts, setInsts] = useState<any[]>([]);
  const [anchor, setAnchor] = useState<any>(null);
  const [ready, setReady] = useState<any>(null);
  const [readyActual, setReadyActual] = useState(false);
  const [loading, setLoading] = useState(false);
  const [bulkType, setBulkType] = useState('불시검문검사');
  const [bulkLead, setBulkLead] = useState(1);

  const load = async () => {
    setLoading(true);
    try {
      const sched = await procureApi.getConstraintSchedule(lotId);
      setInsts(sched.instances || []);
      setAnchor(sched.anchor_date);
      setReady(sched.ready_date);
      setReadyActual(sched.ready_is_actual);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [lotId]);

  const saveAll = async (arr: any[]) => {
    for (let k = 0; k < arr.length; k++) {
      arr[k].sequence = k + 1;
      arr[k].predecessor_id = k === 0 ? 0 : arr[k - 1].constraint_instance_id;
      arr[k] = await procureApi.saveConstraintInstance(arr[k]);
    }
    onChanged?.();
  };

  const commit = async (arr: any[]) => {
    await saveAll(arr);
    await load();
  };

  const changeField = (i: number, field: string, value: any) => {
    setInsts(insts.map((x, idx) => (idx === i ? { ...x, [field]: value } : x)));
  };

  const persistRow = async (i: number) => {
    await procureApi.saveConstraintInstance(insts[i]);
    onChanged?.();
    await load();
  };

  const addStep = async (afterIdx: number | null) => {
    const base = insts[0] || {};
    const newInst: any = {
      constraint_instance_id: 0,
      lot_id: lotId,
      source_container_item_id: base.source_container_item_id || 0,
      constraint_profile_id: base.constraint_profile_id || 0,
      predecessor_id: 0,
      sequence: 0,
      constraint_type: '통관',
      lead_time_days: 0,
      is_optional: false,
      status: 'Planned',
      remark: '',
    };
    const saved: any = await procureApi.saveConstraintInstance(newInst);
    const arr = insts.map(x => ({ ...x }));
    const pos = afterIdx === null ? arr.length : afterIdx + 1;
    arr.splice(pos, 0, saved);
    await commit(arr);
  };

  const removeStep = async (i: number) => {
    const arr = insts.map(x => ({ ...x }));
    const removed = arr.splice(i, 1)[0];
    if (removed.constraint_instance_id) {
      await procureApi.deleteConstraintInstance(removed.constraint_instance_id);
    }
    await commit(arr);
  };

  const move = async (i: number, dir: -1 | 1) => {
    const j = i + dir;
    if (j < 0 || j >= insts.length) return;
    const arr = insts.map(x => ({ ...x }));
    [arr[i], arr[j]] = [arr[j], arr[i]];
    await commit(arr);
  };

  const bulkAppend = async () => {
    const ci = insts[0]?.source_container_item_id;
    if (!ci) {
      alert('선적(Container Item) 정보가 없어 일괄 적용할 수 없습니다.');
      return;
    }
    await procureApi.bulkAppendConstraint(ci, bulkType, bulkLead);
    alert('이 선적의 모든 Lot에 단계가 추가되었습니다.');
    await load();
  };

  return (
    <div className="modal-overlay" style={{ zIndex: 1200 }} onClick={onClose}>
      <div className="modal-content" style={{ maxWidth: '960px' }} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h3>행정 제약 체인 · Lot {lotNo || lotId}</h3>
          <button className="secondary" onClick={onClose}>✕</button>
        </div>

        <div style={{ fontSize: '0.85rem', marginBottom: '0.75rem' }}>
          앵커(GR일): <strong>{fmt(anchor) || '-'}</strong>
          {' · '}Cost Allocation 가능일: <strong style={{ color: 'var(--accent-color)' }}>{fmt(ready) || '미정'}</strong>{' '}
          <span style={{ color: 'var(--text-secondary)' }}>{readyActual ? '(실제)' : '(예상)'}</span>
        </div>

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            시작/완료 실제일을 입력하면 후속 단계로 자동 cascade 됩니다.
          </span>
          <button className="secondary" style={{ padding: '0.3rem 0.8rem', fontSize: '0.8rem' }} onClick={() => addStep(null)}>+ Add Step</button>
        </div>

        <div className="table-container">
          <table className="sub-table">
            <thead>
              <tr>
                <th style={{ width: '36px' }}>#</th>
                <th>Type</th>
                <th style={{ width: '80px' }}>Lead</th>
                <th style={{ width: '140px' }}>Start (예상)</th>
                <th style={{ width: '140px' }}>End (예상)</th>
                <th style={{ width: '150px' }}>Actual Start</th>
                <th style={{ width: '150px' }}>Actual End</th>
                <th style={{ width: '130px' }}></th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan={8} style={{ textAlign: 'center', padding: '1rem' }}>Loading...</td></tr>
              ) : insts.length === 0 ? (
                <tr><td colSpan={8} style={{ textAlign: 'center', padding: '1rem', opacity: 0.6 }}>
                  제약 단계가 없습니다. (Lot 저장 시 프로파일이 있으면 자동 생성됩니다)
                </td></tr>
              ) : insts.map((s, i) => (
                <tr key={s.constraint_instance_id || `new-${i}`}>
                  <td>{i + 1}</td>
                  <td>
                    <select value={s.constraint_type} onChange={e => changeField(i, 'constraint_type', e.target.value)} onBlur={() => persistRow(i)}>
                      {TYPES.map(t => <option key={t} value={t}>{t}</option>)}
                    </select>
                  </td>
                  <td>
                    <input
                      type="text"
                      inputMode="numeric"
                      style={{ width: '50px' }}
                      value={s.lead_time_days}
                      onChange={e => changeField(i, 'lead_time_days', Number(e.target.value.replace(/[^0-9]/g, '')) || 0)}
                      onBlur={() => persistRow(i)}
                    />
                  </td>
                  <td>{fmt(s.computed_start) || '-'}</td>
                  <td>{fmt(s.computed_end) || '-'}</td>
                  <td>
                    <input
                      type="date"
                      value={fmt(s.actual_start)}
                      onChange={e => { changeField(i, 'actual_start', toISO(e.target.value)); }}
                      onBlur={() => persistRow(i)}
                    />
                  </td>
                  <td>
                    <input
                      type="date"
                      value={fmt(s.actual_end)}
                      onChange={e => { changeField(i, 'actual_end', toISO(e.target.value)); }}
                      onBlur={() => persistRow(i)}
                    />
                  </td>
                  <td style={{ whiteSpace: 'nowrap' }}>
                    <button className="secondary" style={{ padding: '0.15rem 0.4rem', fontSize: '0.7rem', minWidth: 'auto', marginRight: '0.2rem' }} onClick={() => move(i, -1)}>▲</button>
                    <button className="secondary" style={{ padding: '0.15rem 0.4rem', fontSize: '0.7rem', minWidth: 'auto', marginRight: '0.2rem' }} onClick={() => move(i, 1)}>▼</button>
                    <button className="secondary" style={{ padding: '0.15rem 0.4rem', fontSize: '0.7rem', minWidth: 'auto', marginRight: '0.2rem' }} onClick={() => addStep(i)}>+</button>
                    <button className="btn-danger secondary" style={{ padding: '0.15rem 0.4rem', fontSize: '0.7rem', minWidth: 'auto' }} onClick={() => removeStep(i)}>✕</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div style={{ marginTop: '1rem', borderTop: '1px solid var(--border-color)', paddingTop: '0.75rem', display: 'flex', gap: '0.5rem', alignItems: 'center', flexWrap: 'wrap' }}>
          <span style={{ fontSize: '0.85rem' }}>선적 일괄 적용:</span>
          <select value={bulkType} onChange={e => setBulkType(e.target.value)}>
            {TYPES.map(t => <option key={t} value={t}>{t}</option>)}
          </select>
          <input type="text" inputMode="numeric" style={{ width: '60px' }} value={bulkLead}
            onChange={e => setBulkLead(Number(e.target.value.replace(/[^0-9]/g, '')) || 0)} />
          <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>일</span>
          <button className="secondary" onClick={bulkAppend}>이 선적의 모든 Lot에 추가</button>
        </div>

        <div className="modal-actions">
          <button className="secondary" onClick={onClose}>닫기</button>
        </div>
      </div>
    </div>
  );
};

export default ConstraintChainEditor;
