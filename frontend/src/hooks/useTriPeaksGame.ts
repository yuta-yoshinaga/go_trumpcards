import { useCallback } from 'react';
import { tripeaksApi } from '../api/gameApi';
import type { TriPeaksHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages TriPeaks game state, hints, and card removal actions. */
export function useTriPeaksGame() {
  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof tripeaksApi.exec>>,
    Parameters<typeof tripeaksApi.exec>,
    TriPeaksHint
  >(tripeaksApi.exec, {
    hintApi: () => tripeaksApi.exec('hint'),
  });

  const handleDraw = useCallback(() => runAction('draw'), [runAction]);
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectCard = useCallback(
    (row: number, col: number) => {
      setHint(null);
      void apiCall('remove', row, col);
    },
    [apiCall, setHint],
  );

  return { ...rest, runAction, setHint, exec: apiCall, handleDraw, handleUndoEscape, handleSelectCard };
}
