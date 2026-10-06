import { useCallback, useState } from 'react';
import { bigBenApi } from '../api/gameApi';
import type { BigBenHint, BigBenMoveZone } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Big Ben game state, source selection, hints, and moves. */
export function useBigBenGame() {
  const [selectedSource, setSelectedSource] = useState<BigBenMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);
  const {
    apiCall,
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
  } = useSolitaireGameBase<Awaited<ReturnType<typeof bigBenApi.exec>>, Parameters<typeof bigBenApi.exec>, BigBenHint>(
    bigBenApi.exec,
    {
      onClearSelection,
      hintApi: () => bigBenApi.exec('hint'),
      selectHint: (res) => res.hint,
    },
  );
  // **補充がこのゲームの逃げ道。**手が尽きたら各列を 3 枚まで戻す。
  const handleDeal = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiCall('deal');
  }, [apiCall, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: BigBenMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: BigBenMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      apiCall('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, apiCall, setHint],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec: apiCall,
    selectedSource,
    hint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleDeal,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
    retry,
  };
}
