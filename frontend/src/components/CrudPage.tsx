import React, { useState, useEffect } from 'react';
import SearchModal from './SearchModal';

export interface Column {
  key: string;
  label: string;
  type?: 'text' | 'number' | 'date';
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
}

function CrudPage<T extends { [key: string]: any }>({ title, columns, fetchData, onSave, emptyItem, renderDetail }: CrudPageProps<T>) {
  const [data, setData] = useState<T[]>([]);
  const [filteredData, setFilteredData] = useState<T[]>([]);
  const [filters, setFilters] = useState<{ [key: string]: any }>({});
  const [selectedItem, setSelectedItem] = useState<T | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [sortConfig, setSortConfig] = useState<{ key: string, direction: 'asc' | 'desc' } | null>(null);
  const [searchConfig, setSearchConfig] = useState<{ type: string, field: string } | null>(null);

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

    // Apply filters
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

    // Apply sorting
    if (sortConfig) {
      result.sort((a, b) => {
        let aVal = a[sortConfig.key];
        let bVal = b[sortConfig.key];

        // Handle nulls
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

          // Default text filter (only show for first 5 columns if not specified)
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
                {columns.map(col => {
                  if (col.tableHidden) return null;
                  return (
                    <td key={col.key}>
                      {col.type === 'date' && item[col.key]
                        ? new Date(item[col.key]).toLocaleDateString()
                        : item[col.key]}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

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
                    <input
                      type={col.type || 'text'}
                      value={col.type === 'date' && selectedItem[col.key]
                        ? new Date(selectedItem[col.key]).toISOString().split('T')[0]
                        : selectedItem[col.key] || ''}
                      onChange={(e) => {
                        let val: any = e.target.value;
                        if (col.type === 'number') {
                          val = Number(val);
                        } else if (col.type === 'date') {
                          val = val ? new Date(val).toISOString() : null;
                        }
                        setSelectedItem({ ...selectedItem, [col.key]: val });
                      }}
                    />
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

      {searchConfig && (
        <SearchModal
          type={searchConfig.type}
          searchTerm=""
          onClose={() => setSearchConfig(null)}
          onSelect={(item) => {
            const id = item.vendor_id || item.po_id || item.ci_id || item.container_id || item.bl_id || item.id;
            if (selectedItem) {
              setSelectedItem({ ...selectedItem, [searchConfig.field]: id });
            }
            setSearchConfig(null);
          }}
        />
      )}
    </div>
  );
}

export default CrudPage;
