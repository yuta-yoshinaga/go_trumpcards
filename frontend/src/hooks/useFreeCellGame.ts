import { useCallback, useState } from 'react';
import { type FreeCellMoveZone, freecellApi } from '../api/gameApi';
import type { FreeCellHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages FreeCell game state, source selection, hints, and moves. */
export function useFreeCellGame() {
  const [selectedSource, setSelectedSource] = useState<FreeCellMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const base = useSolitaireGameBase<
    Awaited<ReturnType<typeof freecellApi.exec>>,
    Parameters<typeof freecellApi.exec>,
    FreeCellHint
  >(freecellApi.exec, {
    onClearSelection,
    hintApi: () => freecellApi.exec('hint'),
  });

  const handleUndoEscape = useCallback(
    (n: number) => base.runAction('undo_n', undefined, undefined, n),
    [base.runAction],
  );

  const handleSelectSource = useCallback((zone: FreeCellMoveZone) => {
    setSelectedSource((prev) => {
      if (
        prev &&
        prev.zone === zone.zone &&
        prev.col === zone.col &&
        prev.cell === zone.cell &&
        prev.cardIndex === zone.cardIndex
      ) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: FreeCellMoveZone) => {
      if (!selectedSource) return;
      base.setHint(null);
      void base.apiCall('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, base],
  );

  // Double-click / double-tap shortcut: dispatch a pre-computed foundation
  // move for the given source without requiring the two-step target click, and
  // clear any pending selection so it stays in lockstep.
  const handleAutoFoundation = useCallback(
    (source: FreeCellMoveZone, target: FreeCellMoveZone) => {
      base.setHint(null);
      void base.apiCall('move', source, target);
      setSelectedSource(null);
    },
    [base],
  );

  return {
    state: base.state,
    loading: base.loading,
    error: base.error,
    hintError: base.hintError,
    exec: base.apiCall,
    selectedSource,
    hint: base.hint,
    handleReset: base.handleReset,
    handleGiveUp: base.handleGiveUp,
    handleHint: base.handleHint,
    handleAutoComplete: base.handleAutoComplete,
    handleUndo: base.handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    handleAutoFoundation,
    isAutoCompleting: base.isAutoCompleting,
    retry: base.retry,
  };
}
