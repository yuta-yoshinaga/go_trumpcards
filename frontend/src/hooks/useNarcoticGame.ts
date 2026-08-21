import { useCallback, useEffect, useState } from 'react';
import { narcoticApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { NarcoticHint } from '../types/card';
import { useGameApi } from './useGameApi';
import { useIsMounted } from './useIsMounted';

/** Hook that manages Narcotic game state, hints, and the deal/discard/stack/redeal actions. */
export function useNarcoticGame() {
  const { state, loading, error, exec, retry } = useGameApi(narcoticApi.exec);
  const [hint, setHint] = useState<NarcoticHint | null>(null);
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

  const isMounted = useIsMounted();

  const handleHint = useCallback(async () => {
    try {
      const res = await narcoticApi.exec('hint');
      // Navigating away mid-request must not write to a gone component (#4447).
      if (!isMounted()) return;
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      if (!isMounted()) return;
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, [isMounted]);

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

  // **列を取らない。**揃った4枚をまとめて捨てるので、選ぶ余地が無い。
  const handleRemove = useCallback(() => {
    setHint(null);
    exec('remove');
  }, [exec]);

  // **クローン元 (Aces Up) には無い手。**山札が尽きても場を集めれば続けられる。
  const handleRedeal = useCallback(() => {
    setHint(null);
    exec('redeal');
  }, [exec]);

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
    handleRedeal,
    handleMove,
    retry,
  };
}
