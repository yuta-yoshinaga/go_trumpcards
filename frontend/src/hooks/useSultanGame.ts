import { useCallback, useEffect, useState } from 'react';
import { type SultanMoveZone, sultanApi } from '../api/gameApi';
import type { SultanHint } from '../types/card';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';
import { useHintRequest } from './useHintRequest';

/** Hook that manages Sultan of Turkey game state, source selection, hints, redeal, and moves. */
export function useSultanGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(sultanApi.exec);
  const [hint, setHint] = useState<SultanHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const { isAutoCompleting, startAutoComplete } = useAutoCompleteState();

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleDraw = useCallback(() => {
    setHint(null);
    exec('draw');
  }, [exec]);

  const handleRedeal = useCallback(() => {
    setHint(null);
    exec('redeal');
  }, [exec]);

  const handleReset = useCallback(() => {
    setHint(null);
    exec('reset');
  }, [exec]);

  const handleGiveUp = useCallback(() => {
    setHint(null);
    exec('giveup');
  }, [exec]);

  const handleHint = useHintRequest({
    fetchHint: () => sultanApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
  });

  const handleAutoComplete = useCallback(() => {
    setHint(null);
    startAutoComplete();
    exec('autocomplete');
  }, [exec, startAutoComplete]);

  const handleUndo = useCallback(() => {
    setHint(null);
    exec('undo');
  }, [exec]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback(
    (n: number) => {
      setHint(null);
      exec('undo_n', undefined, undefined, n);
    },
    [exec],
  );

  /** Play a divan slot or the waste top onto its matching foundation. */
  const handlePlay = useCallback(
    (source: SultanMoveZone) => {
      setHint(null);
      exec('move', source);
    },
    [exec],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec,
    hint,
    handleDraw,
    handleRedeal,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handlePlay,
    isAutoCompleting,
    retry,
  };
}
