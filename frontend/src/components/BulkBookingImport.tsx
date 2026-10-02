import React, { useState, useRef } from 'react';
import * as XLSX from 'xlsx';
import { procureApi } from '../api/procureApi';
import { BookingTemplateRow, BulkImportRow } from '../api/models';

interface BulkBookingImportProps {
  onComplete: () => void;
}

interface PreviewRow {
  row_number: number;
  po_item_id: number;
  po_no: string;
  item_name: string;
  load_qty: number;
  remaining_qty: number;
  container_no: string;
  container_id: number;
  bl_no: string;
  bl_id: number;
  temporary_eta: string;
  remark: string;
  status: 'valid' | 'error' | 'warn';
  message: string;
}

export default function BulkBookingImport({ onComplete }: BulkBookingImportProps) {
  const [loading, setLoading] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [previewData, setPreviewData] = useState<PreviewRow[] | null>(null);
  const [uploadResult, setUploadResult] = useState<{ created: number; total: number } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 템플릿 다운로드
  const handleDownloadTemplate = async () => {
    setDownloading(true);
    setError(null);
    try {
      const templateData = await procureApi.getBookingTemplate();

      if (templateData.length === 0) {
        setError('No unbooked PO items found. All items may be fully booked or no Open POs exist.');
        return;
      }

      // 엑셀 워크북 생성
      const wb = XLSX.utils.book_new();

      // 헤더 정의
      const headers = [
        'PO Item ID (Mandatory)',
        'PO No (Reference)',
        'SKU Code (Reference)',
        'Item Name (Reference)',
        'Vendor (Reference)',
        'Ordered Qty (Reference)',
        'Remaining Qty (Reference)',
        'Load Qty (Mandatory)',
        'Container No / ID (Optional)',
        'BL No / ID (Optional)',
        'Temporary ETA (YYYY-MM-DD)',
        'Remark'
      ];

      // 데이터 변환
      const rows = templateData.map((row: BookingTemplateRow) => [
        row.po_item_id,
        row.po_no,
        row.sku_code,
        row.item_name,
        row.vendor_name,
        row.ordered_qty,
        row.remaining_qty,
        row.load_qty,
        '',
        '',
        '',
        ''
      ]);

      const ws = XLSX.utils.aoa_to_sheet([headers, ...rows]);

      // 컬럼 너비 설정
      ws['!cols'] = [
        { wch: 18 }, // PO Item ID
        { wch: 15 }, // PO No
        { wch: 15 }, // SKU Code
        { wch: 25 }, // Item Name
        { wch: 20 }, // Vendor
        { wch: 14 }, // Ordered Qty
        { wch: 16 }, // Remaining Qty
        { wch: 12 }, // Load Qty
        { wch: 22 }, // Container No/ID
        { wch: 18 }, // BL No/ID
        { wch: 20 }, // Temporary ETA
        { wch: 20 }, // Remark
      ];

      XLSX.utils.book_append_sheet(wb, ws, 'Booking Template');

      // 다운로드
      const today = new Date().toISOString().split('T')[0];
      XLSX.writeFile(wb, `Booking_Template_${today}.xlsx`);
    } catch (err: any) {
      setError(`Failed to download template: ${err.message}`);
    } finally {
      setDownloading(false);
    }
  };

  // 파일 업로드 및 파싱
  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setLoading(true);
    setError(null);
    setPreviewData(null);
    setUploadResult(null);

    try {
      const data = await file.arrayBuffer();
      const wb = XLSX.read(data, { type: 'array' });
      const ws = wb.Sheets[wb.SheetNames[0]];
      const jsonData = XLSX.utils.sheet_to_json<any>(ws, { header: 1 });

      if (jsonData.length < 2) {
        setError('Excel file is empty or has no data rows.');
        return;
      }

      // 헤더 스킵, 데이터 행부터 처리
      const rows = jsonData.slice(1) as any[][];
      const preview: PreviewRow[] = [];

      // 기존 템플릿 데이터를 가져와서 remaining_qty 확인
      const templateData = await procureApi.getBookingTemplate();
      const templateMap = new Map<number, BookingTemplateRow>();
      templateData.forEach((t: BookingTemplateRow) => templateMap.set(t.po_item_id, t));

      for (let i = 0; i < rows.length; i++) {
        const row = rows[i];
        const rowNum = i + 2; // 엑셀 행 번호 (헤더 포함)

        // 빈 행 건너뛰기
        if (!row || row.length === 0 || (row[0] === undefined && row[7] === undefined)) {
          continue;
        }

        const poItemID = Number(row[0]);
        const loadQty = Number(row[7]);
        const containerInput = row[8] !== undefined ? String(row[8]).trim() : '';
        const blInput = row[9] !== undefined ? String(row[9]).trim() : '';
        const temporaryETA = row[10] !== undefined ? String(row[10]).trim() : '';
        const remark = row[11] !== undefined ? String(row[11]).trim() : '';

        const templateRow = templateMap.get(poItemID);
        const poNo = templateRow?.po_no || String(row[1] || '');
        const itemName = templateRow?.item_name || String(row[3] || '');
        const remainingQty = templateRow?.remaining_qty || 0;

        // 컨테이너/BL 파싱 (ID 또는 No로 입력 가능)
        let containerID = 0;
        let containerNo = '';
        let blID = 0;
        let blNo = '';

        if (containerInput) {
          const numVal = Number(containerInput);
          if (!isNaN(numVal) && numVal > 0) {
            containerID = numVal;
          } else {
            containerNo = containerInput;
          }
        }

        if (blInput) {
          const numVal = Number(blInput);
          if (!isNaN(numVal) && numVal > 0) {
            blID = numVal;
          } else {
            blNo = blInput;
          }
        }

        // 유효성 검사
        let status: 'valid' | 'error' | 'warn' = 'valid';
        let messages: string[] = [];

        if (!poItemID || poItemID <= 0) {
          status = 'error';
          messages.push('Invalid PO Item ID');
        } else if (!templateRow) {
          status = 'error';
          messages.push('PO Item ID not found or not in an Open PO');
        }

        if (!loadQty || loadQty <= 0) {
          status = 'error';
          messages.push('Load Qty must be positive');
        } else if (loadQty > remainingQty && status !== 'error') {
          status = 'warn';
          messages.push(`Load Qty (${loadQty}) exceeds remaining (${remainingQty})`);
        }

        preview.push({
          row_number: rowNum,
          po_item_id: poItemID,
          po_no: poNo,
          item_name: itemName,
          load_qty: loadQty,
          remaining_qty: remainingQty,
          container_no: containerNo,
          container_id: containerID,
          bl_no: blNo,
          bl_id: blID,
          temporary_eta: temporaryETA,
          remark: remark,
          status: status,
          message: messages.join('; ') || 'OK',
        });
      }

      setPreviewData(preview);
    } catch (err: any) {
      setError(`Failed to parse Excel file: ${err.message}`);
    } finally {
      setLoading(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  // 확인 후 가져오기 실행
  const handleConfirmImport = async () => {
    if (!previewData) return;

    const validRows = previewData.filter(r => r.status === 'valid' || r.status === 'warn');
    if (validRows.length === 0) {
      setError('No valid rows to import.');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const importRows: BulkImportRow[] = validRows.map(row => ({
        po_item_id: row.po_item_id,
        load_qty: row.load_qty,
        container_no: row.container_no,
        container_id: row.container_id,
        bl_no: row.bl_no,
        bl_id: row.bl_id,
        temporary_eta: row.temporary_eta,
        remark: row.remark,
      }));

      const result = await procureApi.bulkImportBookings(importRows);
      setUploadResult(result);
      setPreviewData(null);
    } catch (err: any) {
      setError(`Import failed: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    setPreviewData(null);
    setUploadResult(null);
    setError(null);
  };

  const validCount = previewData?.filter(r => r.status === 'valid').length || 0;
  const warnCount = previewData?.filter(r => r.status === 'warn').length || 0;
  const errorCount = previewData?.filter(r => r.status === 'error').length || 0;

  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      background: 'rgba(15,23,42,0.4)', display: 'flex', alignItems: 'center', justifyContent: 'center',
      zIndex: 1000, backdropFilter: 'blur(4px)'
    }}>
      <div style={{
        background: '#ffffff', borderRadius: '16px', padding: '2rem',
        width: '90%', maxWidth: '1100px', maxHeight: '90vh', overflow: 'auto',
        border: '1px solid #0f172a', boxShadow: '0 25px 50px rgba(15,23,42,0.15)'
      }}>
        {/* 헤더 */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <h2 style={{ margin: 0, color: '#0f172a' }}>📥 Bulk Booking Import via Excel</h2>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            {previewData && (
              <button className="secondary" onClick={handleReset}>Reset</button>
            )}
            <button className="secondary" onClick={() => { handleReset(); onComplete(); }}>Close</button>
          </div>
        </div>

        {/* 에러 메시지 */}
        {error && (
          <div style={{
            background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.25)',
            borderRadius: '8px', padding: '0.75rem 1rem', marginBottom: '1rem', color: '#b91c1c'
          }}>
            {error}
          </div>
        )}

        {/* 성공 메시지 */}
        {uploadResult && (
          <div style={{
            background: 'rgba(16,185,129,0.08)', border: '1px solid rgba(16,185,129,0.25)',
            borderRadius: '8px', padding: '1rem', marginBottom: '1rem', color: '#047857'
          }}>
            <strong>✅ Import Complete!</strong> {uploadResult.created} of {uploadResult.total} records created successfully.
            <div style={{ marginTop: '0.75rem' }}>
              <button className="btn-success" onClick={() => { handleReset(); onComplete(); }}>Done</button>
            </div>
          </div>
        )}

        {/* 단계 1: 템플릿 다운로드 & 업로드 */}
        {!previewData && !uploadResult && (
          <div>
            <div style={{ marginBottom: '1.5rem' }}>
              <h3 style={{ color: '#0f172a', marginBottom: '0.5rem' }}>Step 1: Download Pre-filled Template</h3>
              <p style={{ color: '#64748b', fontSize: '0.85rem', marginBottom: '1rem' }}>
                Downloads an Excel template pre-filled with all unbooked items from Open Purchase Orders.
                Each row shows the remaining quantity available for booking.
              </p>
              <button
                className="btn-success"
                onClick={handleDownloadTemplate}
                disabled={downloading}
                style={{ minWidth: '200px' }}
              >
                {downloading ? 'Generating...' : '⬇️ Download Template (.xlsx)'}
              </button>
            </div>

            <div style={{
              borderTop: '1px solid #e2e8f0', paddingTop: '1.5rem'
            }}>
              <h3 style={{ color: '#0f172a', marginBottom: '0.5rem' }}>Step 2: Upload Completed Excel</h3>
              <p style={{ color: '#64748b', fontSize: '0.85rem', marginBottom: '1rem' }}>
                After filling in the Load Qty and optional fields (Container, BL, ETA, Remark),
                upload the file for validation and preview.
              </p>
              <input
                ref={fileInputRef}
                type="file"
                accept=".xlsx,.xls"
                onChange={handleFileUpload}
                style={{ display: 'none' }}
              />
              <button
                className="secondary"
                onClick={() => fileInputRef.current?.click()}
                disabled={loading}
                style={{ minWidth: '200px' }}
              >
                {loading ? 'Parsing...' : '📤 Upload Excel File'}
              </button>
            </div>
          </div>
        )}

        {/* 단계 3: 미리보기 & 확인 */}
        {previewData && !uploadResult && (
          <div>
            <h3 style={{ color: '#0f172a', marginBottom: '0.75rem' }}>Preview & Confirmation</h3>

            {/* 요약 */}
            <div style={{
              display: 'flex', gap: '1rem', marginBottom: '1rem',
              background: '#f8fafc', padding: '0.75rem 1rem', borderRadius: '8px'
            }}>
              <span style={{ color: '#64748b' }}>Total: <strong style={{ color: '#0f172a' }}>{previewData.length}</strong></span>
              <span style={{ color: '#047857' }}>Valid: <strong>{validCount}</strong></span>
              <span style={{ color: '#b45309' }}>Warnings: <strong>{warnCount}</strong></span>
              <span style={{ color: '#b91c1c' }}>Errors: <strong>{errorCount}</strong></span>
            </div>

            {/* 미리보기 테이블 */}
            <div style={{ overflowX: 'auto', marginBottom: '1rem' }}>
              <table className="sub-table" style={{ minWidth: '900px' }}>
                <thead>
                  <tr>
                    <th style={{ width: '40px' }}>#</th>
                    <th>Status</th>
                    <th>PO Item ID</th>
                    <th>PO No</th>
                    <th>Item Name</th>
                    <th>Load Qty</th>
                    <th>Remaining</th>
                    <th>Container</th>
                    <th>BL</th>
                    <th>Message</th>
                  </tr>
                </thead>
                <tbody>
                  {previewData.map((row, idx) => (
                    <tr key={idx} style={{
                      background: row.status === 'error' ? 'rgba(239,68,68,0.06)' :
                        row.status === 'warn' ? 'rgba(251,191,36,0.06)' : 'transparent'
                    }}>
                      <td style={{ textAlign: 'center', opacity: 0.6 }}>{row.row_number}</td>
                      <td>
                        <span className={`badge ${row.status === 'valid' ? 'Open' : row.status === 'warn' ? 'Partially Shipped' : 'Closed'}`}>
                          {row.status === 'valid' ? '✓' : row.status === 'warn' ? '⚠' : '✕'}
                        </span>
                      </td>
                      <td>{row.po_item_id}</td>
                      <td>{row.po_no}</td>
                      <td>{row.item_name}</td>
                      <td style={{ textAlign: 'right' }}>{row.load_qty}</td>
                      <td style={{ textAlign: 'right' }}>{row.remaining_qty}</td>
                      <td>{row.container_id ? `ID:${row.container_id}` : row.container_no || '-'}</td>
                      <td>{row.bl_id ? `ID:${row.bl_id}` : row.bl_no || '-'}</td>
                      <td style={{ fontSize: '0.8rem', color: row.status === 'error' ? '#b91c1c' : row.status === 'warn' ? '#b45309' : '#64748b' }}>
                        {row.message}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* 가져오기 버튼 */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
              <button className="secondary" onClick={handleReset}>Cancel</button>
              <button
                className="btn-success"
                onClick={handleConfirmImport}
                disabled={loading || (validCount + warnCount) === 0}
                style={{ minWidth: '180px' }}
              >
                {loading ? 'Importing...' : `Confirm Import (${validCount + warnCount} rows)`}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}