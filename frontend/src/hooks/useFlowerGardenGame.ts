import { useCallback, useState } from 'react';
import { type FlowerGardenMoveZone, flowerGardenApi } from '../api/gameApi';
import type { FlowerGardenHint } from '../types/card';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Flower Garden game state, source selection, hints, and moves. */
export function useFlowerGardenGame() {
  const [selectedSource, setSelectedSource] = useState<FlowerGardenMoveZone | null>(null);

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
    Awaited<ReturnType<typeof flowerGardenApi.exec>>,
    Parameters<typeof flowerGardenApi.exec>,
    FlowerGardenHint
  >(flowerGardenApi.exec, {
    onClearSelection,
    hintApi: () => flowerGardenApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      runApi('undo_n', undefined, undefined, n);
    },
    [runApi, setHint],
  );

  const handleSelectSource = useCallback((zone: FlowerGardenMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: FlowerGardenMoveZone) => {
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
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
    retry,
  };
}
