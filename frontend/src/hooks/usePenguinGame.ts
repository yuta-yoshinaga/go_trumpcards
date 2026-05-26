import { useCallback, useState } from 'react';
import { type PenguinMoveZone, penguinApi } from '../api/gameApi';
import type { PenguinHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Penguin game state, source selection, hints, and moves. */
export function usePenguinGame() {
  const [selectedSource, setSelectedSource] = useState<PenguinMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const base = useSolitaireGameBase<
    Awaited<ReturnType<typeof penguinApi.exec>>,
    Parameters<typeof penguinApi.exec>,
    PenguinHint
  >(penguinApi.exec, {
    onClearSelection,
    hintApi: () => penguinApi.exec('hint'),
  });

  const handleUndoEscape = useCallback(
    (n: number) => base.runAction('undo_n', undefined, undefined, n),
    [base.runAction],
  );

  const handleSelectSource = useCallback((zone: PenguinMoveZone) => {
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
    (zone: PenguinMoveZone) => {
      if (!selectedSource) return;
      base.setHint(null);
      void base.apiCall('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, base],
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
    isAutoCompleting: base.isAutoCompleting,
    retry: base.retry,
  };
}
