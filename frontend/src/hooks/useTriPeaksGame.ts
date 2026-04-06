import { useCallback, useEffect, useState } from 'react';
import { tripeaksApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { TriPeaksHint } from '../types/card';
import { useGameApi } from './useGameApi';

/** Hook that manages TriPeaks game state, hints, and card removal actions. */
export function useTriPeaksGame() {
  const { state, loading, error, exec, retry } = useGameApi(tripeaksApi.exec);
  const [hint, setHint] = useState<TriPeaksHint | null>(null);
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
      const res = await tripeaksApi.exec('hint');
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

  const handleSelectCard = useCallback(
    (row: number, col: number) => {
      setHint(null);
      exec('remove', row, col);
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
    handleSelectCard,
    retry,
  };
}
