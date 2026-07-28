import { useCallback, useState } from 'react';
import { type PyramidRemoveCard, pyramidApi } from '../api/gameApi';
import type { PyramidHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Selected card position in the pyramid or waste. */
export interface PyramidSelection {
  zone: 'pyramid' | 'waste';
  row?: number;
  col?: number;
}

/** Hook that manages Pyramid game state, card selection, hints, and removal actions. */
export function usePyramidGame() {
  const [selectedCard, setSelectedCard] = useState<PyramidSelection | null>(null);
  const onClearSelection = useCallback(() => setSelectedCard(null), []);

  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof pyramidApi.exec>>,
    Parameters<typeof pyramidApi.exec>,
    PyramidHint
  >(pyramidApi.exec, {
    onClearSelection,
    hintApi: () => pyramidApi.exec('hint'),
  });

  const handleDraw = useCallback(() => runAction('draw'), [runAction]);
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const selectionToRemoveCard = useCallback((sel: PyramidSelection): PyramidRemoveCard => {
    return { zone: sel.zone, row: sel.row, col: sel.col };
  }, []);

  const handleSelectCard = useCallback(
    (sel: PyramidSelection, cardValue?: number) => {
      // Any card interaction consumes the current hint.
      setHint(null);
      // King (value 13) - remove solo immediately
      if (cardValue === 13) {
        void apiCall('remove', selectionToRemoveCard(sel));
        setSelectedCard(null);
        return;
      }

      if (selectedCard === null) {
        // First card selected
        setSelectedCard(sel);
        return;
      }

      // If clicking the same card again, deselect
      if (selectedCard.zone === sel.zone && selectedCard.row === sel.row && selectedCard.col === sel.col) {
        setSelectedCard(null);
        return;
      }

      // Second card selected - attempt to remove pair
      void apiCall('remove', selectionToRemoveCard(selectedCard), selectionToRemoveCard(sel));
      setSelectedCard(null);
    },
    [selectedCard, apiCall, setHint, selectionToRemoveCard],
  );

  return { ...rest, runAction, setHint, exec: apiCall, selectedCard, handleDraw, handleUndoEscape, handleSelectCard };
}
