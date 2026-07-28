import { useCallback, useEffect, useState } from 'react';
import { golfApi } from '../api/gameApi';
import type { GolfHint } from '../types/card';
import { useGameApi } from './useGameApi';
import { useHintRequest } from './useHintRequest';

/** Hook that manages Golf Solitaire game state, hints, and card removal actions. */
export function useGolfGame() {
  const { state, loading, error, exec, retry } = useGameApi(golfApi.exec);
  const [hint, setHint] = useState<GolfHint | null>(null);
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

  const handleHint = useHintRequest({
    fetchHint: () => golfApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
  });

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

  const handleSelectCard = useCallback(
    (col: number) => {
      setHint(null);
      exec('remove', col);
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
    handleSelectCard,
    retry,
  };
}
