import { useCallback, useState } from 'react';
import { duchessApi } from '../api/gameApi';
import type { DuchessHint, DuchessMoveZone } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Duchess game state, source selection, hints, and moves. */
export function useDuchessGame() {
  const [selectedSource, setSelectedSource] = useState<DuchessMoveZone | null>(null);
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
    Awaited<ReturnType<typeof duchessApi.exec>>,
    Parameters<typeof duchessApi.exec>,
    DuchessHint
  >(duchessApi.exec, {
    onClearSelection,
    hintApi: () => duchessApi.exec('hint'),
    selectHint: (res) => res.hint,
  });
  /** Fix the rank all four foundations start from, taken off a reserve fan. */
  const handleChooseBase = useCallback(
    (fanIdx: number) => {
      setSelectedSource(null);
      setHint(null);
      apiCall('base', { zone: 'reserve', col: fanIdx });
    },
    [apiCall, setHint],
  );

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiCall('draw');
  }, [apiCall, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: DuchessMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: DuchessMoveZone) => {
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
    handleChooseBase,
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
