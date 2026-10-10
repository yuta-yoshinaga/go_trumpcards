import { useCallback, useState } from 'react';
import { type StHelenaMoveZone, stHelenaApi } from '../api/gameApi';
import type { StHelenaHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages StHelena Solitaire state, source selection, hints, and moves. */
export function useStHelenaGame() {
  const [selectedSource, setSelectedSource] = useState<StHelenaMoveZone | null>(null);
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
    Awaited<ReturnType<typeof stHelenaApi.exec>>,
    Parameters<typeof stHelenaApi.exec>,
    StHelenaHint
  >(stHelenaApi.exec, {
    onClearSelection,
    hintApi: () => stHelenaApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  const handleRedeal = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('redeal');
  }, [exec, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: StHelenaMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: StHelenaMoveZone) => {
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
