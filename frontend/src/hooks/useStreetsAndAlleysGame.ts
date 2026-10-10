import { useCallback, useState } from 'react';
import { type StreetsAndAlleysMoveZone, streetsAndAlleysApi } from '../api/gameApi';
import type { StreetsAndAlleysHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Streets and Alleys game state, source selection, hints, and moves. */
export function useStreetsAndAlleysGame() {
  const [selectedSource, setSelectedSource] = useState<StreetsAndAlleysMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);
  const {
    apiCall: exec,
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
    Awaited<ReturnType<typeof streetsAndAlleysApi.exec>>,
    Parameters<typeof streetsAndAlleysApi.exec>,
    StreetsAndAlleysHint
  >(streetsAndAlleysApi.exec, {
    onClearSelection,
    hintApi: () => streetsAndAlleysApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: StreetsAndAlleysMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: StreetsAndAlleysMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      exec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, exec, setHint],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec,
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
