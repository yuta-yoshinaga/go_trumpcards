import { useCallback, useState } from 'react';
import { congressApi } from '../api/gameApi';
import type { CongressHint, CongressMoveZone } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Congress game state, source selection, hints, and moves. */
export function useCongressGame() {
  const [selectedSource, setSelectedSource] = useState<CongressMoveZone | null>(null);
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
    Awaited<ReturnType<typeof congressApi.exec>>,
    Parameters<typeof congressApi.exec>,
    CongressHint
  >(congressApi.exec, {
    onClearSelection,
    hintApi: () => congressApi.exec('hint'),
    selectHint: (res) => res.hint,
  });
  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiCall('draw');
  }, [apiCall, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: CongressMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) return null;
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: CongressMoveZone) => {
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
