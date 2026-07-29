import { useCallback, useEffect, useState } from 'react';
import { buraApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Bura game state and the multi-card selection a lead needs.
 *
 * Selection lives here rather than in the page because a play clears it, and
 * leaving that to each call site is how a stale index survives into the next
 * trick and plays the wrong card.
 */
export function useBuraGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(buraApi.exec);
  const [selected, setSelected] = useState<number[]>([]);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    setSelected([]);
    runApi('reset');
  }, [runApi]);

  /** Toggle one hand index in or out of the current selection. */
  const toggleCard = useCallback((idx: number) => {
    setSelected((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  }, []);

  const clearSelection = useCallback(() => {
    setSelected([]);
  }, []);

  const handlePlay = useCallback(() => {
    if (selected.length === 0) return;
    setSelected([]);
    runApi('play', selected);
  }, [runApi, selected]);

  const handleClaim = useCallback(() => {
    setSelected([]);
    runApi('claim');
  }, [runApi]);

  const handleDeclare = useCallback(() => {
    setSelected([]);
    runApi('declare');
  }, [runApi]);

  return {
    state,
    loading,
    error,
    exec: runApi,
    selected,
    toggleCard,
    clearSelection,
    handleReset,
    handlePlay,
    handleClaim,
    handleDeclare,
    retry,
  };
}
