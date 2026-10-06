import { useCallback, useState } from 'react';
import { matrimonyApi } from '../api/games/matrimony';
import type { MatrimonyHint, MatrimonyMoveZone } from '../types/games/matrimony';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Matrimony game state, source selection, hints, and moves. */
export function useMatrimonyGame() {
  const [selectedSource, setSelectedSource] = useState<MatrimonyMoveZone | null>(null);

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
    Awaited<ReturnType<typeof matrimonyApi.exec>>,
    Parameters<typeof matrimonyApi.exec>,
    MatrimonyHint
  >(matrimonyApi.exec, { onClearSelection, hintApi: () => matrimonyApi.exec('hint'), selectHint: (res) => res.hint });

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    runApi('draw');
  }, [runApi, setHint]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      runApi('undo_n', undefined, undefined, n);
    },
    [runApi, setHint],
  );

  const handleSelectSource = useCallback((zone: MatrimonyMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) return null;
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: MatrimonyMoveZone) => {
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
