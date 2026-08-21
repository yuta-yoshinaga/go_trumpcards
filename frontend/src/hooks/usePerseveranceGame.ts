import { useCallback, useEffect, useState } from 'react';
import { type PerseveranceMoveZone, perseveranceApi } from '../api/gameApi';
import type { PerseveranceHint } from '../types/card';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';
import { useHintRequest } from './useHintRequest';

/** Hook that manages Perseverance game state, source selection, hints, and moves. */
export function usePerseveranceGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(perseveranceApi.exec);
  const [selectedSource, setSelectedSource] = useState<PerseveranceMoveZone | null>(null);
  const [hint, setHint] = useState<PerseveranceHint | null>(null);
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

  const handleGiveUp = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('giveup');
  }, [exec]);

  const handleHint = useHintRequest({
    fetchHint: () => perseveranceApi.exec('hint'),
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

  // **リディールは Perseverance だけの救済手段** (クローン元の Baker's Dozen には無い)。
  // 選択とヒントを落としてから投げる ── 盤が総入れ替えになるので、残した索引は
  // 全部別の札を指す。
  const handleRedeal = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('redeal');
  }, [exec]);

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

  const handleSelectSource = useCallback((zone: PerseveranceMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: PerseveranceMoveZone) => {
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
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleRedeal,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
    retry,
  };
}
