import { useCallback, useState } from 'react';
import { toggleArrayItem } from '../utils/arrayUtils';

/** Hook that manages card selection state with toggle and clear operations. */
export function useCardSelection(initialSelection: number[] = []) {
  const [selected, setSelected] = useState<number[]>(initialSelection);

  const toggle = useCallback((idx: number) => {
    setSelected((prev) => toggleArrayItem(prev, idx));
  }, []);

  const clear = useCallback(() => setSelected([]), []);

  return { selected, toggle, clear, setSelected } as const;
}
