import { useCallback, useEffect, useState } from 'react';
import { type PyramidRemoveCard, pyramidApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { PyramidHint } from '../types/card';
import { useGameApi } from './useGameApi';

/** Selected card position in the pyramid or waste. */
export interface PyramidSelection {
  zone: 'pyramid' | 'waste';
  row?: number;
  col?: number;
}

/** Hook that manages Pyramid game state, card selection, hints, and removal actions. */
export function usePyramidGame() {
  const { state, loading, error, exec: rawExec } = useGameApi(pyramidApi.exec);
  const [selectedCard, setSelectedCard] = useState<PyramidSelection | null>(null);
  const [hint, setHint] = useState<PyramidHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);

  const callApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    callApi('reset');
  }, [callApi]);

  const handleDraw = useCallback(() => {
    setSelectedCard(null);
    setHint(null);
    callApi('draw');
  }, [callApi]);

  const handleReset = useCallback(() => {
    setSelectedCard(null);
    setHint(null);
    callApi('reset');
  }, [callApi]);

  const handleGiveUp = useCallback(() => {
    setSelectedCard(null);
    setHint(null);
    callApi('giveup');
  }, [callApi]);

  const handleHint = useCallback(async () => {
    try {
      const res = await pyramidApi.exec('hint');
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, []);

  const handleUndo = useCallback(() => {
    setSelectedCard(null);
    setHint(null);
    callApi('undo');
  }, [callApi]);

  const selectionToRemoveCard = useCallback((sel: PyramidSelection): PyramidRemoveCard => {
    return { zone: sel.zone, row: sel.row, col: sel.col };
  }, []);

  const handleSelectCard = useCallback(
    (sel: PyramidSelection, cardValue?: number) => {
      // King (value 13) - remove solo immediately
      if (cardValue === 13) {
        setHint(null);
        callApi('remove', selectionToRemoveCard(sel));
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
      setHint(null);
      callApi('remove', selectionToRemoveCard(selectedCard), selectionToRemoveCard(sel));
      setSelectedCard(null);
    },
    [selectedCard, callApi, selectionToRemoveCard],
  );

  return {
    state,
    loading,
    error,
    hintError,
    selectedCard,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleSelectCard,
  };
}
