import { useCallback, useState } from 'react';
import { type SeahavenTowersMoveZone, seahaventowersApi } from '../api/gameApi';
import type { SeahavenTowersHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Seahaven Towers game state, source selection, hints, and moves. */
export function useSeahavenTowersGame() {
  const [selectedSource, setSelectedSource] = useState<SeahavenTowersMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const base = useSolitaireGameBase<
    Awaited<ReturnType<typeof seahaventowersApi.exec>>,
    Parameters<typeof seahaventowersApi.exec>,
    SeahavenTowersHint
  >(seahaventowersApi.exec, {
    onClearSelection,
    hintApi: () => seahaventowersApi.exec('hint'),
  });

  const handleUndoEscape = useCallback(
    (n: number) => base.runAction('undo_n', undefined, undefined, n),
    [base.runAction],
  );

  const handleSelectSource = useCallback((zone: SeahavenTowersMoveZone) => {
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
    (zone: SeahavenTowersMoveZone) => {
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
