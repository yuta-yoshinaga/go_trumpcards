import { useCallback, useEffect, useState } from 'react';
import { golfApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { GolfHint } from '../types/card';
import { useGameApi } from './useGameApi';

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

  const handleHint = useCallback(async () => {
    try {
      const res = await golfApi.exec('hint');
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
    handleSelectCard,
    retry,
  };
}
