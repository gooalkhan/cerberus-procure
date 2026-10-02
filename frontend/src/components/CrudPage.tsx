import React, { useState, useEffect } from 'react';
import SearchModal from './SearchModal';
import { procureApi } from '../api/procureApi';

export interface Column {
  key: string;
  label: string;
  type?: 'text' | 'number' | 'date' | 'select';
  options?: string[];
  filterType?: 'text' | 'select' | 'none';
  filterOptions?: string[];
  formHidden?: boolean;
  tableHidden?: boolean;
  fullWidth?: boolean;
  searchType?: string;
  divider?: boolean;
}

interface CrudPageProps<T> {
  title: string;
  columns: Column[];
  fetchData: () => Promise<T[]>;
  onSave: (data: T) => Promise<void>;
  emptyItem: T;
  renderDetail?: (item: T, onChange: (updatedItem: T) => void) => React.ReactNode;
  tableName?: string;
  idField?: string;
  onFieldChange?: (field: string, value: any, currentItem: T) => Promise<Partial<T> | null>;
}

interface DeleteModalState {
  item: any;
  tableName: string;
  id: number;
  references: any[];
  canDelete: boolean;
  totalRefs: number;
  loading: boolean;
}

function CrudPage<T extends { [key: string]: any }>({ title, columns, fetchData, onSave, emptyItem, renderDetail, tableName, idField, onFieldChange }: CrudPageProps<T>) {
  const [data, setData] = useState<T[]>([]);
  const [filteredData, setFilteredData] = useState<T[]>([]);
  const [filters, setFilters] = useState<{ [key: string]: any }>({});
  const [selectedItem, setSelectedItem] = useState<T | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [sortConfig, setSortConfig] = useState<{ key: string, direction: 'asc' | 'desc' } | null>(null);
  const [searchConfig, setSearchConfig] = useState<{ type: string, field: string } | null>(null);
  const [deleteModal, setDeleteModal] = useState<DeleteModalState | null>(null);

  useEffect(() => {
    loadData();
  }, [fetchData]);

  const loadData = async () => {
    const res = await fetchData();
    const list = Array.isArray(res) ? res : [];
    setData(list);
    setFilteredData(list);
  };

  const handleSort = (key: string) => {
    let direction: 'asc' | 'desc' = 'asc';
    if (sortConfig && sortConfig.key === key && sortConfig.direction === 'asc') {
      direction = 'desc';
    }
    setSortConfig({ key, direction });
  };

  useEffect(() => {
    let result = [...data];

    Object.keys(filters).forEach(key => {
      const val = filters[key];
      if (!val) return;

      if (key.endsWith('_start')) {
        const field = key.replace('_start', '');
        result = result.filter(item => !item[field] || new Date(item[field]) >= new Date(val));
      } else if (key.endsWith('_end')) {
        const field = key.replace('_end', '');
        result = result.filter(item => !item[field] || new Date(item[field]) <= new Date(val));
      } else {
        result = result.filter(item =>
          String(item[key]).toLowerCase().includes(String(val).toLowerCase())
        );
      }
    });

    if (sortConfig) {
      result.sort((a, b) => {
        let aVal = a[sortConfig.key];
        let bVal = b[sortConfig.key];

        if (aVal === null || aVal === undefined) return 1;
        if (bVal === null || bVal === undefined) return -1;

        if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
      });
    }

    setFilteredData(result);
  }, [filters, data, sortConfig]);

  const handleRowClick = (item: T) => {
    setSelectedItem({ ...item });
    setIsModalOpen(true);
  };

  const handleAddNew = () => {
    setSelectedItem({ ...emptyItem });
    setIsModalOpen(true);
  };

  const handleSave = async () => {
    if (selectedItem) {
      await onSave(selectedItem);
      alert('저장되었습니다.');
      setIsModalOpen(false);
      loadData();
    }
  };

  const getRecordId = (item: any): number => {
    if (idField) return item[idField];
    // Auto-detect ID field
    for (const col of columns) {
      if (col.key.endsWith('_id') && !col.key.includes('vendor') && !col.key.includes('container') && !col.key.includes('bl') && !col.key.includes('po') && !col.key.includes('ci') && !col.key.includes('gr') && !col.key.includes('lot') && !col.key.includes('ap')) {
        return item[col.key];
      }
    }
    // Fallback: first column ending with _id
    for (const col of columns) {
      if (col.key.endsWith('_id')) return item[col.key];
    }
    return 0;
  };

  const handleDeleteClick = async (e: React.MouseEvent, item: any) => {
    e.stopPropagation();

    if (!tableName) {
      alert('삭제 기능이 지원되지 않습니다.');
      return;
    }

    const id = getRecordId(item);
    if (!id) {
      alert('레코드 ID를 찾을 수 없습니다.');
      return;
    }

    setDeleteModal({
      item,
      tableName,
      id,
      references: [],
      canDelete: true,
      totalRefs: 0,
      loading: true,
    });

    try {
      const result = await procureApi.checkReferences(tableName, id);
      setDeleteModal(prev => prev ? {
        ...prev,
        references: result.references || [],
        canDelete: result.can_delete,
        totalRefs: result.total_refs,
        loading: false,
      } : null);
    } catch (err: any) {
      setDeleteModal(prev => prev ? { ...prev, loading: false } : null);
    }
  };

  const handleConfirmDelete = async () => {
    if (!deleteModal || !deleteModal.canDelete) return;

    try {
      await procureApi.deleteRecord(deleteModal.tableName, deleteModal.id);
      setDeleteModal(null);
      loadData();
    } catch (err: any) {
      alert(`삭제 실패: ${err.message}`);
    }
  };

  const getItemDisplayName = (item: any): string => {
    // Try to find a meaningful display name
    if (item.po_no) return item.po_no;
    if (item.ci_no) return item.ci_no;
    if (item.ap_no) return item.ap_no;
    if (item.bl_no) return item.bl_no;
    if (item.container_no) return item.container_no;
    if (item.name) return item.name;
    if (item.sku_code) return item.sku_code;
    if (item.lot_no) return item.lot_no;
    return `ID: ${getRecordId(item)}`;
  };

  const formatCellValue = (value: any, col: Column): string => {
    if (value === null || value === undefined) return '';
    if (col.type === 'date' && value) {
      return new Date(value).toLocaleDateString();
    }
    if (typeof value === 'number') {
      return value.toLocaleString();
    }
    return String(value);
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h2 className="page-title">{title}</h2>
        <button onClick={handleAddNew}>+ New Entry</button>
      </div>

      <div className="filter-panel">
        {columns.map(col => {
          if (col.filterType === 'none') return null;

          if (col.type === 'date') {
            return (
              <React.Fragment key={col.key}>
                <div className="filter-group">
                  <label>{col.label} (Start)</label>
                  <input
                    type="date"
                    value={filters[`${col.key}_start`] || ''}
                    onChange={(e) => setFilters({ ...filters, [`${col.key}_start`]: e.target.value })}
                  />
                </div>
                <div className="filter-group">
                  <label>{col.label} (End)</label>
                  <input
                    type="date"
                    value={filters[`${col.key}_end`] || ''}
                    onChange={(e) => setFilters({ ...filters, [`${col.key}_end`]: e.target.value })}
                  />
                </div>
              </React.Fragment>
            );
          }

          if (col.filterType === 'select') {
            return (
              <div key={col.key} className="filter-group">
                <label>{col.label}</label>
                <select
                  value={filters[col.key] || ''}
                  onChange={(e) => setFilters({ ...filters, [col.key]: e.target.value })}
                >
                  <option value="">All {col.label}</option>
                  {col.filterOptions?.map(opt => (
                    <option key={opt} value={opt}>{opt}</option>
                  ))}
                </select>
              </div>
            );
          }

          const isDefaultShown = columns.indexOf(col) < 5;
          if (!col.filterType && !isDefaultShown) return null;

          return (
            <div key={col.key} className="filter-group">
              <label>{col.label}</label>
              <input
                placeholder={`Search ${col.label}...`}
                value={filters[col.key] || ''}
                onChange={(e) => setFilters({ ...filters, [col.key]: e.target.value })}
              />
            </div>
          );
        })}
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              {tableName && <th style={{ width: '50px', textAlign: 'center', cursor: 'default' }}></th>}
              {columns.map(col => {
                if (col.tableHidden) return null;
                const isSorted = sortConfig?.key === col.key;
                return (
                  <th
                    key={col.key}
                    onClick={() => handleSort(col.key)}
                    style={{ cursor: 'pointer', userSelect: 'none', position: 'relative', paddingRight: '20px' }}
                  >
                    {col.label}
                    <span style={{ marginLeft: '4px', opacity: isSorted ? 1 : 0.3, fontSize: '0.8rem' }}>
                      {isSorted ? (sortConfig.direction === 'asc' ? '↑' : '↓') : '↕'}
                    </span>
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {filteredData.map((item, idx) => (
              <tr key={idx} onClick={() => handleRowClick(item)}>
                {tableName && (
                  <td style={{ textAlign: 'center', padding: '0.4rem' }} onClick={e => e.stopPropagation()}>
                    <button
                      className="btn-danger"
                      onClick={(e) => handleDeleteClick(e, item)}
                      style={{
                        padding: '2px 8px',
                        fontSize: '0.75rem',
                        minWidth: 'auto',
                        borderRadius: '4px',
                        opacity: 0.8,
                      }}
                      title="삭제"
                    >
                      ✕
                    </button>
                  </td>
                )}
                {columns.map(col => {
                  if (col.tableHidden) return null;
                  return (
                    <td key={col.key}>
                      {formatCellValue(item[col.key], col)}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Edit Modal */}
      {isModalOpen && selectedItem && (
        <div className="modal-overlay" onClick={() => setIsModalOpen(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2>{selectedItem.id || selectedItem[columns[0].key] ? 'Edit' : 'New'} {title}</h2>
              <button className="secondary" onClick={() => setIsModalOpen(false)}>✕</button>
            </div>
            <div className="form-grid">
              {columns.map(col => {
                if (col.formHidden) return null;
                if (col.divider) {
                  return <div key={col.key} style={{ gridColumn: 'span 3', borderTop: '1px solid var(--border-color)', margin: '1rem 0', opacity: 0.3 }}></div>;
                }
                return (
                  <div key={col.key} className={`form-group ${col.fullWidth ? 'full-width' : ''}`}>
                    <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      {col.label}
                      {col.searchType && (
                        <span
                          style={{
                            fontSize: '0.65rem',
                            background: 'var(--accent-color)',
                            color: 'white',
                            padding: '2px 6px',
                            borderRadius: '4px',
                            cursor: 'pointer',
                            textTransform: 'uppercase',
                            fontWeight: 'bold'
                          }}
                          onClick={() => setSearchConfig({ type: col.searchType!, field: col.key })}
                        >
                          search
                        </span>
                      )}
                    </label>
                    {col.type === 'select' ? (
                      <select
                        value={selectedItem[col.key] || ''}
                        onChange={async (e) => {
                          const val = e.target.value;
                          let updatedItem = { ...selectedItem, [col.key]: val };
                          if (onFieldChange) {
                            const additionalUpdates = await onFieldChange(col.key, val, updatedItem);
                            if (additionalUpdates) {
                              updatedItem = { ...updatedItem, ...additionalUpdates };
                            }
                          }
                          setSelectedItem(updatedItem);
                        }}
                      >
                        {(col.options || []).map(opt => (
                          <option key={opt} value={opt}>{opt}</option>
                        ))}
                      </select>
                    ) : (
                      <input
                        type={col.type || 'text'}
                        value={col.type === 'date' && selectedItem[col.key]
                          ? new Date(selectedItem[col.key]).toISOString().split('T')[0]
                          : selectedItem[col.key] || ''}
                        onChange={async (e) => {
                          let val: any = e.target.value;
                          if (col.type === 'number') {
                            val = Number(val);
                          } else if (col.type === 'date') {
                            val = val ? new Date(val).toISOString() : null;
                          }
                          let updatedItem = { ...selectedItem, [col.key]: val };
                          if (onFieldChange) {
                            const additionalUpdates = await onFieldChange(col.key, val, updatedItem);
                            if (additionalUpdates) {
                              updatedItem = { ...updatedItem, ...additionalUpdates };
                            }
                          }
                          setSelectedItem(updatedItem);
                        }}
                      />
                    )}
                  </div>
                );
              })}
            </div>
            {renderDetail && selectedItem && renderDetail(selectedItem, setSelectedItem)}
            <div className="modal-actions">
              <button className="btn-danger secondary" onClick={() => setIsModalOpen(false)}>Cancel</button>
              <button className="btn-success" onClick={handleSave}>Save Changes</button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteModal && (
        <div className="modal-overlay" onClick={() => setDeleteModal(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()} style={{ maxWidth: '600px' }}>
            <div className="modal-header">
              <h2>🗑️ 삭제 확인</h2>
              <button className="secondary" onClick={() => setDeleteModal(null)}>✕</button>
            </div>

            <div style={{ marginBottom: '1.5rem' }}>
              <p style={{ color: 'var(--text-primary)', marginBottom: '0.5rem' }}>
                <strong>"{getItemDisplayName(deleteModal.item)}"</strong> 레코드를 삭제하시겠습니까?
              </p>
              <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem' }}>
                테이블: {deleteModal.tableName} | ID: {deleteModal.id}
              </p>
            </div>

            {deleteModal.loading ? (
              <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>
                참조 데이터 확인 중...
              </div>
            ) : deleteModal.references.length > 0 ? (
              <div>
                <div style={{
                  background: 'rgba(239, 68, 68, 0.1)',
                  border: '1px solid rgba(239, 68, 68, 0.3)',
                  borderRadius: '8px',
                  padding: '1rem',
                  marginBottom: '1rem'
                }}>
                  <p style={{ color: '#fca5a5', fontWeight: 600, marginBottom: '0.5rem' }}>
                    ⚠️ 이 레코드는 {deleteModal.totalRefs}개의 다른 데이터에서 참조하고 있어 삭제할 수 없습니다.
                  </p>
                  <p style={{ color: '#fca5a5', fontSize: '0.85rem' }}>
                    아래 데이터를 먼저 삭제한 후 다시 시도해 주세요.
                  </p>
                </div>

                {deleteModal.references.map((ref, idx) => (
                  <div key={idx} style={{
                    background: 'rgba(255,255,255,0.03)',
                    border: '1px solid var(--border-color)',
                    borderRadius: '8px',
                    padding: '1rem',
                    marginBottom: '0.75rem'
                  }}>
                    <h4 style={{ color: 'var(--accent-color)', marginBottom: '0.75rem', fontSize: '0.9rem' }}>
                      {ref.table_name} ({ref.count}건)
                    </h4>
                    <div style={{ overflowX: 'auto' }}>
                      <table className="sub-table" style={{ margin: 0 }}>
                        <thead>
                          <tr>
                            {ref.records && ref.records[0] && Object.keys(ref.records[0]).map(key => (
                              <th key={key}>{key.replace(/_/g, ' ')}</th>
                            ))}
                          </tr>
                        </thead>
                        <tbody>
                          {ref.records?.slice(0, 5).map((record: any, rIdx: number) => (
                            <tr key={rIdx}>
                              {Object.values(record).map((val: any, vIdx) => (
                                <td key={vIdx}>{val !== null && val !== undefined ? String(val) : '-'}</td>
                              ))}
                            </tr>
                          ))}
                          {ref.records && ref.records.length > 5 && (
                            <tr>
                              <td colSpan={Object.keys(ref.records[0]).length} style={{ textAlign: 'center', opacity: 0.6 }}>
                                ... 및 {ref.records.length - 5}건 더
                              </td>
                            </tr>
                          )}
                        </tbody>
                      </table>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div style={{
                background: 'rgba(16, 185, 129, 0.1)',
                border: '1px solid rgba(16, 185, 129, 0.3)',
                borderRadius: '8px',
                padding: '1rem',
                marginBottom: '1rem'
              }}>
                <p style={{ color: '#6ee7b7' }}>
                  ✅ 이 레코드를 참조하는 다른 데이터가 없어 안전하게 삭제할 수 있습니다.
                </p>
              </div>
            )}

            <div className="modal-actions">
              <button className="secondary" onClick={() => setDeleteModal(null)}>취소</button>
              <button
                className="btn-danger"
                onClick={handleConfirmDelete}
                disabled={!deleteModal.canDelete || deleteModal.loading}
                style={{
                  opacity: (!deleteModal.canDelete || deleteModal.loading) ? 0.5 : 1,
                  cursor: (!deleteModal.canDelete || deleteModal.loading) ? 'not-allowed' : 'pointer'
                }}
              >
                삭제하기
              </button>
            </div>
          </div>
        </div>
      )}

      {searchConfig && (
        <SearchModal
          type={searchConfig.type}
          searchTerm=""
          onClose={() => setSearchConfig(null)}
          onSelect={async (item) => {
            const getID = (item: any, type: string) => {
              switch (type) {
                case 'Vendor': return item.vendor_id;
                case 'PO': return item.po_id;
                case 'CI': return item.ci_id;
                case 'Container': return item.container_id;
                case 'BL': return item.bl_id;
                case 'PO Item': return item.po_item_id || item.id;
                case 'Item': return item.item_id;
                case 'GR': return item.gr_id;
                case 'Lot': return item.lot_id;
                case 'Container Item': return item.container_item_id;
                default: return item.id || item.vendor_id || item.po_id || item.ci_id || item.container_id || item.bl_id;
              }
            };
            const id = getID(item, searchConfig.type);
            if (selectedItem) {
              let updatedItem = { ...selectedItem, [searchConfig.field]: id };
              if (onFieldChange) {
                const additionalUpdates = await onFieldChange(searchConfig.field, id, updatedItem);
                if (additionalUpdates) {
                  updatedItem = { ...updatedItem, ...additionalUpdates };
                }
              }
              setSelectedItem(updatedItem);
            }
            setSearchConfig(null);
          }}
        />
      )}
    </div>
  );
}

export default CrudPage;
