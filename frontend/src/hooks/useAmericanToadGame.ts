import { useCallback, useState } from 'react';
import { americanToadApi } from '../api/gameApi';
import type { AmericanToadHint, AmericanToadMoveZone } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages American Toad game state, source selection, hints, and moves. */
export function useAmericanToadGame() {
  const [selectedSource, setSelectedSource] = useState<AmericanToadMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);
  const {
    apiCall,
    runAction,
    state,
    loading,
    error,
    retry,
    hint,
    hintError,
    isAutoCompleting,
    setHint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
  } = useSolitaireGameBase<
    Awaited<ReturnType<typeof americanToadApi.exec>>,
    Parameters<typeof americanToadApi.exec>,
    AmericanToadHint
  >(americanToadApi.exec, {
    onClearSelection,
    hintApi: () => americanToadApi.exec('hint'),
    selectHint: (res) => res.hint,
  });
  /** Turn one card. When the stock is out this recycles the waste — once only. */
  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiCall('draw');
  }, [apiCall, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: AmericanToadMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: AmericanToadMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      apiCall('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, apiCall, setHint],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec: apiCall,
    selectedSource,
    hint,
    handleReset,
    handleDraw,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
    retry,
  };
}
