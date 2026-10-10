import { useCallback, useState } from 'react';
import { type FortyAndEightMoveZone, fortyAndEightApi } from '../api/gameApi';
import type { FortyAndEightHint } from '../types/card';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Forty and Eight game state, source selection, hints, redeal, and moves. */
export function useFortyAndEightGame() {
  const [selectedSource, setSelectedSource] = useState<FortyAndEightMoveZone | null>(null);

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
    Awaited<ReturnType<typeof fortyAndEightApi.exec>>,
    Parameters<typeof fortyAndEightApi.exec>,
    FortyAndEightHint
  >(fortyAndEightApi.exec, {
    onClearSelection,
    hintApi: () => fortyAndEightApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('draw');
  }, [exec, setHint]);

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

  const handleSelectSource = useCallback((zone: FortyAndEightMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: FortyAndEightMoveZone) => {
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
    handleDraw,
    handleRedeal,
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
