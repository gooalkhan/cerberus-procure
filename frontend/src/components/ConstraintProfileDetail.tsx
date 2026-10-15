import React, { useState, useEffect } from 'react';
import { ConstraintProfile, ConstraintProfileStep, ConstraintMaster } from '../api/models';
import { procureApi } from '../api/procureApi';

interface Props {
  profile: ConstraintProfile;
  onChange: (updated: ConstraintProfile) => void;
}

const ConstraintProfileDetail: React.FC<Props> = ({ profile, onChange }) => {
  const [constraints, setConstraints] = useState<ConstraintMaster[]>([]);

  useEffect(() => {
    procureApi.getConstraints().then(setConstraints).catch(() => setConstraints([]));
  }, []);

  const steps = profile.steps || [];
  const update = (arr: ConstraintProfileStep[]) => onChange({ ...profile, steps: arr });

  const addStep = () => {
    update([
      ...steps,
      {
        constraint_profile_step_id: 0,
        constraint_profile_id: profile.constraint_profile_id || 0,
        constraint_id: 0,
        sequence: steps.length + 1,
        constraint_type: '',
        lead_time_days: 0,
        is_optional: false,
        remark: '',
      },
    ]);
  };

  const changeStep = (i: number, patch: Partial<ConstraintProfileStep>) => {
    update(steps.map((s, idx) => (idx === i ? { ...s, ...patch } : s)));
  };

  const removeStep = (i: number) => update(steps.filter((_, idx) => idx !== i));

  const move = (i: number, dir: -1 | 1) => {
    const j = i + dir;
    if (j < 0 || j >= steps.length) return;
    const arr = steps.map(s => ({ ...s }));
    [arr[i], arr[j]] = [arr[j], arr[i]];
    arr.forEach((s, idx) => { s.sequence = idx + 1; });
    update(arr);
  };

  const selectConstraint = (i: number, idStr: string) => {
    const id = Number(idStr);
    const c = constraints.find(x => x.constraint_id === id);
    if (c) {
      changeStep(i, { constraint_id: id, constraint_type: c.name, lead_time_days: c.lead_time_days });
    } else {
      changeStep(i, { constraint_id: 0 });
    }
  };

  return (
    <div style={{ marginTop: '1rem', borderTop: '1px solid var(--border-color)', paddingTop: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
        <h3 style={{ color: 'var(--accent-color)', margin: 0, fontSize: '1rem' }}>
          제약 단계 (Steps)
          <span style={{ fontSize: '0.75rem', fontWeight: 400, color: 'var(--text-secondary)', marginLeft: '0.5rem' }}>
            Constraints 메뉴에서 정의한 제약을 골라 순서를 구성합니다
          </span>
        </h3>
        <button className="secondary" type="button" style={{ padding: '0.3rem 0.8rem', fontSize: '0.8rem' }} onClick={addStep}>
          + Add Step
        </button>
      </div>

      {steps.length === 0 ? (
        <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', padding: '1rem', textAlign: 'center', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
          단계가 없습니다. "+ Add Step"으로 추가하세요.
        </div>
      ) : (
        <div className="table-container">
          <table className="sub-table">
            <thead>
              <tr>
                <th style={{ width: '40px' }}>#</th>
                <th>Constraint</th>
                <th style={{ width: '100px' }}>Lead (days)</th>
                <th style={{ width: '80px' }}>Optional</th>
                <th>Remark</th>
                <th style={{ width: '140px' }}></th>
              </tr>
            </thead>
            <tbody>
              {steps.map((s, i) => (
                <tr key={i}>
                  <td>{i + 1}</td>
                  <td>
                    <select value={s.constraint_id || 0} onChange={e => selectConstraint(i, e.target.value)} style={{ width: '100%' }}>
                      <option value={0}>— 제약 선택 —</option>
                      {constraints.map(c => (
                        <option key={c.constraint_id} value={c.constraint_id}>{c.name}</option>
                      ))}
                    </select>
                    {s.constraint_id === 0 && s.constraint_type ? (
                      <div style={{ fontSize: '0.7rem', color: 'var(--text-secondary)' }}>기존: {s.constraint_type}</div>
                    ) : null}
                  </td>
                  <td>
                    <input
                      type="text"
                      inputMode="numeric"
                      style={{ width: '70px' }}
                      value={s.lead_time_days}
                      onChange={e => changeStep(i, { lead_time_days: Number(e.target.value.replace(/[^0-9]/g, '')) || 0 })}
                    />
                  </td>
                  <td style={{ textAlign: 'center' }}>
                    <input
                      type="checkbox"
                      checked={!!s.is_optional}
                      onChange={e => changeStep(i, { is_optional: e.target.checked })}
                    />
                  </td>
                  <td>
                    <input type="text" value={s.remark || ''} onChange={e => changeStep(i, { remark: e.target.value })} />
                  </td>
                  <td style={{ whiteSpace: 'nowrap' }}>
                    <button className="secondary" type="button" style={{ padding: '0.15rem 0.4rem', fontSize: '0.7rem', minWidth: 'auto', marginRight: '0.2rem' }} onClick={() => move(i, -1)}>▲</button>
                    <button className="secondary" type="button" style={{ padding: '0.15rem 0.4rem', fontSize: '0.7rem', minWidth: 'auto', marginRight: '0.2rem' }} onClick={() => move(i, 1)}>▼</button>
                    <button className="btn-danger secondary" type="button" style={{ padding: '0.15rem 0.4rem', fontSize: '0.7rem', minWidth: 'auto' }} onClick={() => removeStep(i)}>✕</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default ConstraintProfileDetail;
