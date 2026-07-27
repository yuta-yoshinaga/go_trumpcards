import { useCallback, useEffect, useState } from 'react';
import { type FlowerGardenMoveZone, flowerGardenApi } from '../api/gameApi';
import type { FlowerGardenHint } from '../types/card';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';
import { useHintRequest } from './useHintRequest';

/** Hook that manages Flower Garden game state, source selection, hints, and moves. */
export function useFlowerGardenGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(flowerGardenApi.exec);
  const [selectedSource, setSelectedSource] = useState<FlowerGardenMoveZone | null>(null);
  const [hint, setHint] = useState<FlowerGardenHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const { isAutoCompleting, startAutoComplete } = useAutoCompleteState();

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    runApi('reset');
  }, [runApi]);

  const handleGiveUp = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    runApi('giveup');
  }, [runApi]);

  const handleHint = useHintRequest({
    fetchHint: () => flowerGardenApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
  });

  const handleAutoComplete = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    startAutoComplete();
    runApi('autocomplete');
  }, [runApi, startAutoComplete]);

  const handleUndo = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    runApi('undo');
  }, [runApi]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      runApi('undo_n', undefined, undefined, n);
    },
    [runApi],
  );

  const handleSelectSource = useCallback((zone: FlowerGardenMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: FlowerGardenMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      runApi('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, runApi],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec: runApi,
    selectedSource,
    hint,
    handleReset,
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
