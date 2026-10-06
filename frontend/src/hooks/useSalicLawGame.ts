import { useCallback, useState } from 'react';
import { salicLawApi } from '../api/gameApi';
import type { SalicLawHint, SalicLawMoveZone } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages SalicLaw game state, source selection, hints, and moves. */
export function useSalicLawGame() {
  const [selectedSource, setSelectedSource] = useState<SalicLawMoveZone | null>(null);
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
    Awaited<ReturnType<typeof salicLawApi.exec>>,
    Parameters<typeof salicLawApi.exec>,
    SalicLawHint
  >(salicLawApi.exec, {
    onClearSelection,
    hintApi: () => salicLawApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('draw');
  }, [exec, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: SalicLawMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col) return null;
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: SalicLawMoveZone) => {
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
