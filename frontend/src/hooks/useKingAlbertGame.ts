import { useCallback, useState } from 'react';
import { type KingAlbertMoveZone, kingAlbertApi } from '../api/gameApi';
import type { KingAlbertHint } from '../types/card';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages King Albert game state, source selection, hints, and moves. */
export function useKingAlbertGame() {
  const [selectedSource, setSelectedSource] = useState<KingAlbertMoveZone | null>(null);

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
    Awaited<ReturnType<typeof kingAlbertApi.exec>>,
    Parameters<typeof kingAlbertApi.exec>,
    KingAlbertHint
  >(kingAlbertApi.exec, { onClearSelection, hintApi: () => kingAlbertApi.exec('hint'), selectHint: (res) => res.hint });

  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      runApi('undo_n', undefined, undefined, n);
    },
    [runApi, setHint],
  );

  const handleSelectSource = useCallback((zone: KingAlbertMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: KingAlbertMoveZone) => {
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
