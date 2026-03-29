import { useCallback, useEffect, useState } from 'react';
import { type SpiderConfigInput, type SpiderMoveZone, spiderApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { SpiderHint } from '../types/card';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';

/** Hook that manages Spider Solitaire game state, source selection, hints, and moves. */
export function useSpiderGame() {
  const { state, loading, error, exec: rawExec } = useGameApi(spiderApi.exec);
  const [selectedSource, setSelectedSource] = useState<SpiderMoveZone | null>(null);
  const [hint, setHint] = useState<SpiderHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const { isAutoCompleting, startAutoComplete } = useAutoCompleteState();

  const apiExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    apiExec('reset');
  }, [apiExec]);

  const handleDeal = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiExec('deal');
  }, [apiExec]);

  const handleReset = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiExec('reset');
  }, [apiExec]);

  const handleResetWithConfig = useCallback(
    (config: SpiderConfigInput) => {
      setSelectedSource(null);
      setHint(null);
      apiExec('reset', undefined, undefined, config);
    },
    [apiExec],
  );

  const handleGiveUp = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiExec('giveup');
  }, [apiExec]);

  const handleHint = useCallback(async () => {
    try {
      const res = await spiderApi.exec('hint');
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
    apiExec('autocomplete');
  }, [apiExec, startAutoComplete]);

  const handleUndo = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    apiExec('undo');
  }, [apiExec]);

  const handleSelectSource = useCallback((zone: SpiderMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: SpiderMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      apiExec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, apiExec],
  );

  return {
    state,
    loading,
    error,
    hintError,
    selectedSource,
    hint,
    handleDeal,
    handleReset,
    handleResetWithConfig,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
  };
}
