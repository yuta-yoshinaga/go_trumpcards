import { useCallback, useEffect, useState } from 'react';
import { acesupApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { AcesUpHint } from '../types/card';
import { useGameApi } from './useGameApi';

/** Hook that manages Aces Up game state, hints, deal/remove/move actions. */
export function useAcesUpGame() {
  const { state, loading, error, exec, retry } = useGameApi(acesupApi.exec);
  const [hint, setHint] = useState<AcesUpHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleDraw = useCallback(() => {
    setHint(null);
    exec('draw');
  }, [exec]);

  const handleReset = useCallback(() => {
    setHint(null);
    exec('reset');
  }, [exec]);

  const handleGiveUp = useCallback(() => {
    setHint(null);
    exec('giveup');
  }, [exec]);

  const handleHint = useCallback(async () => {
    try {
      const res = await acesupApi.exec('hint');
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, []);

  const handleUndo = useCallback(() => {
    setHint(null);
    exec('undo');
  }, [exec]);

  /** Batch undo to escape stalemate. */
  const handleUndoEscape = useCallback(
    (n: number) => {
      setHint(null);
      exec('undo_n', undefined, n);
    },
    [exec],
  );

  const handleRemove = useCallback(
    (col: number) => {
      setHint(null);
      exec('remove', col);
    },
    [exec],
  );

  const handleMove = useCallback(
    (col: number) => {
      setHint(null);
      exec('move', col);
    },
    [exec],
  );

  return {
    state,
    loading,
    error,
    exec,
    hintError,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleUndoEscape,
    handleRemove,
    handleMove,
    retry,
  };
}
