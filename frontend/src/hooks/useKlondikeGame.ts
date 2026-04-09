import { useCallback, useEffect, useState } from 'react';
import { type KlondikeConfigInput, type KlondikeMoveZone, klondikeApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { KlondikeHint } from '../types/card';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';

/** Hook that manages Klondike game state, source selection, hints, and moves. */
export function useKlondikeGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(klondikeApi.exec);
  const [selectedSource, setSelectedSource] = useState<KlondikeMoveZone | null>(null);
  const [hint, setHint] = useState<KlondikeHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const { isAutoCompleting, startAutoComplete } = useAutoCompleteState();

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('draw');
  }, [exec]);

  const handleReset = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('reset');
  }, [exec]);

  const handleResetWithConfig = useCallback(
    (config: KlondikeConfigInput) => {
      setSelectedSource(null);
      setHint(null);
      exec('reset', undefined, undefined, config);
    },
    [exec],
  );

  const handleGiveUp = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('giveup');
  }, [exec]);

  const handleHint = useCallback(async () => {
    try {
      const res = await klondikeApi.exec('hint');
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, []);

  const handleAutoComplete = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    startAutoComplete();
    exec('autocomplete');
  }, [exec, startAutoComplete]);

  const handleUndo = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('undo');
  }, [exec]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      exec('undo_n', undefined, undefined, undefined, n);
    },
    [exec],
  );

  const handleSelectSource = useCallback((zone: KlondikeMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: KlondikeMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      exec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, exec],
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
    handleReset,
    handleResetWithConfig,
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
