import { useEffect, useRef, useState } from 'react';

/**
 * Track a "chain" of consecutive remove actions (no stock draw between them).
 *
 * The hook watches `moveCount` and `stockCount` from the server response and
 * increments an internal counter whenever `moveCount` rises while `stockCount`
 * stays the same. Any change to `stockCount` (i.e. the player drew from the
 * stock) resets the chain back to zero.
 *
 * Used by TriPeaks and Golf to surface a streak/Combo badge.
 */
export function useChainCombo(moveCount: number | undefined, stockCount: number | undefined): number {
  const [combo, setCombo] = useState(0);
  const prevMove = useRef<number | undefined>(moveCount);
  const prevStock = useRef<number | undefined>(stockCount);

  useEffect(() => {
    if (moveCount === undefined || stockCount === undefined) {
      setCombo(0);
      prevMove.current = moveCount;
      prevStock.current = stockCount;
      return;
    }
    const lastMove = prevMove.current;
    const lastStock = prevStock.current;
    if (lastMove === undefined || lastStock === undefined) {
      prevMove.current = moveCount;
      prevStock.current = stockCount;
      return;
    }
    if (moveCount === 0) {
      setCombo(0);
    } else if (stockCount !== lastStock) {
      setCombo(0);
    } else if (moveCount > lastMove) {
      setCombo((c) => c + 1);
    }
    prevMove.current = moveCount;
    prevStock.current = stockCount;
  }, [moveCount, stockCount]);

  return combo;
}
