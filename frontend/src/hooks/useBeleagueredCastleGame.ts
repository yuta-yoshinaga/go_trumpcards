import { useCallback, useState } from 'react';
import { type BeleagueredCastleMoveZone, beleagueredCastleApi } from '../api/gameApi';
import type { BeleagueredCastleHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Beleaguered Castle game state, source selection, hints, and moves. */
export function useBeleagueredCastleGame() {
  const [selectedSource, setSelectedSource] = useState<BeleagueredCastleMoveZone | null>(null);
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
    Awaited<ReturnType<typeof beleagueredCastleApi.exec>>,
    Parameters<typeof beleagueredCastleApi.exec>,
    BeleagueredCastleHint
  >(beleagueredCastleApi.exec, {
    onClearSelection,
    hintApi: () => beleagueredCastleApi.exec('hint'),
    selectHint: (res) => res.hint,
  });
  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: BeleagueredCastleMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: BeleagueredCastleMoveZone) => {
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
