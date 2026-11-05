import React, { useEffect, useState } from 'react';

// 숫자 문자열 → number (콤마 허용, 입력 중 상태도 안전)
export function parseNumberInput(value: string): number {
  const cleaned = (value ?? '').replace(/,/g, '').trim();
  if (cleaned === '' || cleaned === '-' || cleaned === '.' || cleaned === '-.') return 0;
  const num = Number(cleaned);
  return isNaN(num) ? 0 : num;
}

interface NumberInputProps {
  value: number | null | undefined;
  onChange: (value: number) => void;
  onBlur?: () => void;
  className?: string;
  style?: React.CSSProperties;
  placeholder?: string;
  disabled?: boolean;
  autoFocus?: boolean;
  title?: string;
}

/**
 * 소수점 입력이 가능한 숫자 입력 컴포넌트.
 * 입력 중에는 원본 문자열("12.")을 유지하고, 모델에는 파싱된 number를 전달한다.
 */
const NumberInput: React.FC<NumberInputProps> = ({ value, onChange, onBlur, ...rest }) => {
  const toText = (v: number | null | undefined) =>
    v === null || v === undefined || Number.isNaN(v) ? '' : String(v);

  const [text, setText] = useState<string>(toText(value));

  useEffect(() => {
    const parsed = parseNumberInput(text);
    if (value !== null && value !== undefined && value !== parsed) {
      setText(toText(value));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  return (
    <input
      type="text"
      inputMode="decimal"
      value={text}
      onChange={e => {
        const raw = e.target.value;
        // 숫자/부호/콤마/소수점만 허용 (예: 1,234.56)
        if (!/^-?[0-9,]*(\.[0-9]*)?$/.test(raw)) return;
        setText(raw);
        onChange(parseNumberInput(raw));
      }}
      onBlur={() => {
        onBlur?.();
        const n = parseNumberInput(text);
        setText(text === '' ? '' : String(n));
        onChange(n);
      }}
      {...rest}
    />
  );
};

export default NumberInput;
