import { useCallback, useState } from 'react';
import { grandfathersClockApi } from '../api/gameApi';
import type { GrandfathersClockHint, GrandfathersClockMoveZone } from '../types/card';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Grandfather's Clock game state, source selection, hints, and moves. */
export function useGrandfathersClockGame() {
  const [selectedSource, setSelectedSource] = useState<GrandfathersClockMoveZone | null>(null);

  const onClearSelection = useCallback(() => setSelectedSource(null), []);
  const {
    state,
    loading,
    error,
    retry,
    apiCall: runApi,
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
    Awaited<ReturnType<typeof grandfathersClockApi.exec>>,
    Parameters<typeof grandfathersClockApi.exec>,
    GrandfathersClockHint
  >(grandfathersClockApi.exec, {
    onClearSelection,
    hintApi: () => grandfathersClockApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  const handleRedo = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    runApi('redo');
  }, [runApi, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      runApi('undo_n', undefined, undefined, n);
    },
    [runApi, setHint],
  );

  const handleSelectSource = useCallback((zone: GrandfathersClockMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: GrandfathersClockMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      runApi('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, runApi, setHint],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec: runApi,
    selectedSource,
    hint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleRedo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
    retry,
  };
}
