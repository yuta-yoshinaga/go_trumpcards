import { useCallback, useState } from 'react';
import { braidApi } from '../api/gameApi';
import type { BraidHint, BraidMoveZone } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Braid game state, source selection, hints, and moves. */
export function useBraidGame() {
  const [selectedSource, setSelectedSource] = useState<BraidMoveZone | null>(null);
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
  } = useSolitaireGameBase<Awaited<ReturnType<typeof braidApi.exec>>, Parameters<typeof braidApi.exec>, BraidHint>(
    braidApi.exec,
    {
      onClearSelection,
      hintApi: () => braidApi.exec('hint'),
      selectHint: (res) => res.hint,
    },
  );
  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiCall('draw');
  }, [apiCall, setHint]);

  /**
   * Fix the direction the foundations build in. Legal exactly once per game;
   * the backend rejects a second call, so the button disappears once set.
   */
  const handleChooseDirection = useCallback(
    (ascending: boolean) => {
      setSelectedSource(null);
      setHint(null);
      apiCall('dir', undefined, undefined, undefined, ascending);
    },
    [apiCall, setHint],
  );

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: BraidMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) return null;
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: BraidMoveZone) => {
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
    handleChooseDirection,
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
