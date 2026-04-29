import { useCallback } from 'react';
import { tripeaksApi } from '../api/gameApi';
import type { TriPeaksHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages TriPeaks game state, hints, and card removal actions. */
export function useTriPeaksGame() {
  const base = useSolitaireGameBase<
    Awaited<ReturnType<typeof tripeaksApi.exec>>,
    Parameters<typeof tripeaksApi.exec>,
    TriPeaksHint
  >(tripeaksApi.exec, {
    hintApi: () => tripeaksApi.exec('hint'),
  });

  const handleDraw = useCallback(() => base.runAction('draw'), [base.runAction]);
  const handleUndoEscape = useCallback(
    (n: number) => base.runAction('undo_n', undefined, undefined, n),
    [base.runAction],
  );

  const handleSelectCard = useCallback(
    (row: number, col: number) => {
      base.setHint(null);
      void base.apiCall('remove', row, col);
    },
    [base],
  );

  return {
    state: base.state,
    loading: base.loading,
    error: base.error,
    exec: base.apiCall,
    hintError: base.hintError,
    hint: base.hint,
    handleDraw,
    handleReset: base.handleReset,
    handleGiveUp: base.handleGiveUp,
    handleHint: base.handleHint,
    handleUndo: base.handleUndo,
    handleUndoEscape,
    handleSelectCard,
    retry: base.retry,
  };
}
