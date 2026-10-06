import { useCallback } from 'react';
import { golfApi } from '../api/gameApi';
import type { GolfHint } from '../types/card';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Golf Solitaire game state, hints, and card removal actions. */
export function useGolfGame() {
  const {
    state,
    loading,
    error,
    retry,
    apiCall: exec,
    hint,
    hintError,
    setHint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
  } = useSolitaireGameBase<Awaited<ReturnType<typeof golfApi.exec>>, Parameters<typeof golfApi.exec>, GolfHint>(
    golfApi.exec,
    { hintApi: () => golfApi.exec('hint'), selectHint: (res) => res.hint },
  );

  const handleDraw = useCallback(() => {
    setHint(null);
    exec('draw');
  }, [exec, setHint]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      setHint(null);
      exec('undo_n', undefined, n);
    },
    [exec, setHint],
  );

  const handleSelectCard = useCallback(
    (col: number) => {
      setHint(null);
      exec('remove', col);
    },
    [exec, setHint],
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
