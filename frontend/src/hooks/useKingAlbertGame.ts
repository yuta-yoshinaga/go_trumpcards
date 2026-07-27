import { useCallback, useEffect, useState } from 'react';
import { type KingAlbertMoveZone, kingAlbertApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { KingAlbertHint } from '../types/card';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';
import { useIsMounted } from './useIsMounted';

/** Hook that manages King Albert game state, source selection, hints, and moves. */
export function useKingAlbertGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(kingAlbertApi.exec);
  const [selectedSource, setSelectedSource] = useState<KingAlbertMoveZone | null>(null);
  const [hint, setHint] = useState<KingAlbertHint | null>(null);
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

  const isMounted = useIsMounted();

  const handleHint = useCallback(async () => {
    try {
      const res = await kingAlbertApi.exec('hint');
      // Navigating away mid-request must not write to a gone component (#4447).
      if (!isMounted()) return;
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      if (!isMounted()) return;
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, [isMounted]);

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

  const handleSelectSource = useCallback((zone: KingAlbertMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: KingAlbertMoveZone) => {
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
