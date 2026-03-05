import { useCallback, useState } from 'react';

export function useCardSelection(initialSelection: number[] = []) {
  const [selected, setSelected] = useState<number[]>(initialSelection);

  const toggle = useCallback((idx: number) => {
    setSelected((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  }, []);

  const clear = useCallback(() => setSelected([]), []);

  return { selected, toggle, clear, setSelected } as const;
}
