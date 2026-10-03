import { useCallback, useEffect, useState } from 'react';
import { napoleonsSquareApi } from '../api/gameApi';
import type { NapoleonsSquareHint, NapoleonsSquareMoveZone } from '../types/card';
import { type NapoleonsSquareMoveSource, napoleonsSquareLegalTargets } from '../utils/napoleonsSquareLegalTargets';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';
import { useHintRequest } from './useHintRequest';

/** Hook that manages Napoleon's Square game state, source selection, hints, and moves. */
export function useNapoleonsSquareGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(napoleonsSquareApi.exec);
  const [selectedSource, setSelectedSource] = useState<NapoleonsSquareMoveSource | null>(null);
  const [hint, setHint] = useState<NapoleonsSquareHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const { isAutoCompleting, startAutoComplete } = useAutoCompleteState();

  const legalTargets = state && selectedSource ? napoleonsSquareLegalTargets(state, selectedSource) : [];

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    runApi('reset');
  }, [runApi]);

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    runApi('draw');
  }, [runApi]);

  const handleGiveUp = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    runApi('giveup');
  }, [runApi]);

  const handleHint = useHintRequest({
    fetchHint: () => napoleonsSquareApi.exec('hint'),
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

  const handleSelectSource = useCallback((zone: NapoleonsSquareMoveZone) => {
    const source: NapoleonsSquareMoveSource | null =
      zone.zone === 'waste'
        ? { zone: 'waste' }
        : zone.zone === 'tableau' && zone.col !== undefined && zone.cardIndex !== undefined
          ? { zone: 'tableau', col: zone.col, cardIndex: zone.cardIndex }
          : null;
    if (!source) return;
    setSelectedSource((prev) => {
      if (
        prev &&
        prev.zone === source.zone &&
        prev.zone === 'tableau' &&
        source.zone === 'tableau' &&
        prev.col === source.col &&
        prev.cardIndex === source.cardIndex
      ) {
        return null;
      }
      if (prev?.zone === 'waste' && source.zone === 'waste') return null;
      return source;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: NapoleonsSquareMoveZone) => {
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
    legalTargets,
    hint,
    handleReset,
    handleDraw,
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
