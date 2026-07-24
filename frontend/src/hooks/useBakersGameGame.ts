import { useCallback, useState } from 'react';
import { bakersgameApi, type FreeCellMoveZone } from '../api/gameApi';
import type { FreeCellHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/**
 * Hook that manages Baker's Game state, source selection, hints, and moves.
 * Baker's Game shares the FreeCell wire shape (FreeCellMoveZone / FreeCellHint);
 * only the server-side same-suit stacking rule differs.
 */
export function useBakersGameGame() {
  const [selectedSource, setSelectedSource] = useState<FreeCellMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const base = useSolitaireGameBase<
    Awaited<ReturnType<typeof bakersgameApi.exec>>,
    Parameters<typeof bakersgameApi.exec>,
    FreeCellHint
  >(bakersgameApi.exec, {
    onClearSelection,
    hintApi: () => bakersgameApi.exec('hint'),
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

  // Double-click / double-tap shortcut: dispatch a pre-computed auto-move (a
  // foundation move, or an empty-free-cell fallback) for the given source
  // without the two-step target click, and clear any pending selection so it
  // stays in lockstep.
  const handleAutoMove = useCallback(
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
    handleAutoMove,
    isAutoCompleting: base.isAutoCompleting,
    retry: base.retry,
  };
}
