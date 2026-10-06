import { useCallback, useState } from 'react';
import { type CrescentMoveZone, crescentApi } from '../api/gameApi';
import type { CrescentHint } from '../types/card';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Crescent Solitaire state, source selection, hints, and moves. */
export function useCrescentGame() {
  const [selectedSource, setSelectedSource] = useState<CrescentMoveZone | null>(null);

  const onClearSelection = useCallback(() => setSelectedSource(null), []);
  const {
    state,
    loading,
    error,
    retry,
    apiCall: exec,
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
    Awaited<ReturnType<typeof crescentApi.exec>>,
    Parameters<typeof crescentApi.exec>,
    CrescentHint
  >(crescentApi.exec, { onClearSelection, hintApi: () => crescentApi.exec('hint'), selectHint: (res) => res.hint });

  const handleRedeal = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('redeal');
  }, [exec, setHint]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      exec('undo_n', undefined, undefined, n);
    },
    [exec, setHint],
  );

  const handleSelectSource = useCallback((zone: CrescentMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: CrescentMoveZone) => {
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
    handleRedeal,
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
