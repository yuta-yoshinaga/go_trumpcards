import { useCallback, useEffect, useState } from 'react';
import { type StHelenaMoveZone, stHelenaApi } from '../api/gameApi';
import type { StHelenaHint } from '../types/card';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';
import { useHintRequest } from './useHintRequest';

/** Hook that manages StHelena Solitaire state, source selection, hints, and moves. */
export function useStHelenaGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(stHelenaApi.exec);
  const [selectedSource, setSelectedSource] = useState<StHelenaMoveZone | null>(null);
  const [hint, setHint] = useState<StHelenaHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const { isAutoCompleting, startAutoComplete } = useAutoCompleteState();

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleReset = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('reset');
  }, [exec]);

  const handleRedeal = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('redeal');
  }, [exec]);

  const handleGiveUp = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('giveup');
  }, [exec]);

  const handleHint = useHintRequest({
    fetchHint: () => stHelenaApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
  });

  const handleAutoComplete = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    startAutoComplete();
    exec('autocomplete');
  }, [exec, startAutoComplete]);

  const handleUndo = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('undo');
  }, [exec]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      exec('undo_n', undefined, undefined, n);
    },
    [exec],
  );

  const handleSelectSource = useCallback((zone: StHelenaMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: StHelenaMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      exec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, exec],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec,
    selectedSource,
    hint,
    handleReset,
    handleRedeal,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
    retry,
  };
}
