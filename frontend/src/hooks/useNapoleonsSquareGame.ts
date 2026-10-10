import { useCallback, useState } from 'react';
import { napoleonsSquareApi } from '../api/gameApi';
import type { NapoleonsSquareHint, NapoleonsSquareMoveZone } from '../types/card';
import { type NapoleonsSquareMoveSource, napoleonsSquareLegalTargets } from '../utils/napoleonsSquareLegalTargets';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Napoleon's Square game state, source selection, hints, and moves. */
export function useNapoleonsSquareGame() {
  const [selectedSource, setSelectedSource] = useState<NapoleonsSquareMoveSource | null>(null);

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
    Awaited<ReturnType<typeof napoleonsSquareApi.exec>>,
    Parameters<typeof napoleonsSquareApi.exec>,
    NapoleonsSquareHint
  >(napoleonsSquareApi.exec, {
    onClearSelection,
    hintApi: () => napoleonsSquareApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  const legalTargets = state && selectedSource ? napoleonsSquareLegalTargets(state, selectedSource) : [];

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

  const handleSelectSource = useCallback((zone: NapoleonsSquareMoveZone) => {
    const source: NapoleonsSquareMoveSource | null =
      zone.zone === 'waste'
        ? { zone: 'waste' }
        : zone.zone === 'tableau' && zone.col !== undefined && zone.cardIndex !== undefined
          ? { zone: 'tableau', col: zone.col, cardIndex: zone.cardIndex }
          : null;
    if (!source) return;
    setSelectedSource((prev) => {
      if (
        prev &&
        prev.zone === source.zone &&
        prev.zone === 'tableau' &&
        source.zone === 'tableau' &&
        prev.col === source.col &&
        prev.cardIndex === source.cardIndex
      ) {
        return null;
      }
      if (prev?.zone === 'waste' && source.zone === 'waste') return null;
      return source;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: NapoleonsSquareMoveZone) => {
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
    legalTargets,
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
