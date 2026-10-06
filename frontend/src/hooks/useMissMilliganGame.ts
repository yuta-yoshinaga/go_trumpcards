import { useCallback, useState } from 'react';
import { missMilliganApi } from '../api/gameApi';
import type { MissMilliganHint, MissMilliganMoveZone } from '../types/card';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Miss Milligan game state, source selection, hints, and moves. */
export function useMissMilliganGame() {
  const [selectedSource, setSelectedSource] = useState<MissMilliganMoveZone | null>(null);

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
    Awaited<ReturnType<typeof missMilliganApi.exec>>,
    Parameters<typeof missMilliganApi.exec>,
    MissMilliganHint
  >(missMilliganApi.exec, {
    onClearSelection,
    hintApi: () => missMilliganApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  const handleDeal = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    runApi('deal');
  }, [runApi, setHint]);

  /** Lift a run aside. Only legal once the stock is gone and nothing is held. */
  const handleWaive = useCallback(
    (col: number, cardIndex?: number) => {
      setSelectedSource(null);
      setHint(null);
      runApi('waive', { zone: 'tableau', col, cardIndex });
    },
    [runApi, setHint],
  );

  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      runApi('undo_n', undefined, undefined, n);
    },
    [runApi, setHint],
  );

  const handleSelectSource = useCallback((zone: MissMilliganMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: MissMilliganMoveZone) => {
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
    handleDeal,
    handleWaive,
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
