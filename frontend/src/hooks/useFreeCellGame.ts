import { useCallback, useEffect, useState } from 'react';
import { type FreeCellMoveZone, freecellApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { FreeCellHint } from '../types/card';
import { useGameApi } from './useGameApi';

export function useFreeCellGame() {
  const { state, loading, error, exec: rawExec } = useGameApi(freecellApi.exec);
  const [selectedSource, setSelectedSource] = useState<FreeCellMoveZone | null>(null);
  const [hint, setHint] = useState<FreeCellHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);

  const callExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    callExec('reset');
  }, [callExec]);

  const handleReset = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    callExec('reset');
  }, [callExec]);

  const handleGiveUp = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    callExec('giveup');
  }, [callExec]);

  const handleHint = useCallback(async () => {
    try {
      const res = await freecellApi.exec('hint');
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, []);

  const handleAutoComplete = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    callExec('autocomplete');
  }, [callExec]);

  const handleUndo = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    callExec('undo');
  }, [callExec]);

  const handleSelectSource = useCallback((zone: FreeCellMoveZone) => {
    setSelectedSource((prev) => {
      if (
        prev &&
        prev.zone === zone.zone &&
        prev.col === zone.col &&
        prev.cell === zone.cell &&
        prev.cardIndex === zone.cardIndex
      ) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: FreeCellMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      callExec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, callExec],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec: callExec,
    selectedSource,
    hint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleSelectSource,
    handleSelectTarget,
  };
}
