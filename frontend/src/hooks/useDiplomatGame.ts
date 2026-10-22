import { useCallback, useState } from 'react';
import { diplomatApi } from '../api/gameApi';
import type { DiplomatHint, DiplomatMoveZone } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Diplomat game state, source selection, hints, and moves. */
export function useDiplomatGame() {
  const [selectedSource, setSelectedSource] = useState<DiplomatMoveZone | null>(null);
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
    Awaited<ReturnType<typeof diplomatApi.exec>>,
    Parameters<typeof diplomatApi.exec>,
    DiplomatHint
  >(diplomatApi.exec, {
    onClearSelection,
    hintApi: () => diplomatApi.exec('hint'),
    selectHint: (res) => res.hint,
  });
  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiCall('draw');
  }, [apiCall, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: DiplomatMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) return null;
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: DiplomatMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      apiCall('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, apiCall, setHint],
  );

  /**
   * Double-click / double-tap shortcut: send an exposed card to its server supplied foundation target.
   */
  const handleFoundationShortcut = useCallback(
    (source: DiplomatMoveZone) => {
      const key = source.zone === 'waste' ? 'waste' : source.col === undefined ? null : `tableau-${source.col}`;
      if (key === null) return;
      const target = state?.legalTargets[key]?.foundation[0];
      if (target === undefined) return;
      setHint(null);
      apiCall('move', source, { zone: 'foundation', col: target });
      setSelectedSource(null);
    },
    [state?.legalTargets, apiCall, setHint],
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
    handleFoundationShortcut,
    isAutoCompleting,
    retry,
  };
}
