import { useCallback, useEffect, useState } from 'react';
import { bridgeApi } from '../api/gameApi';
import type { BridgeConfig, BridgeHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useHintRequest } from './useHintRequest';

/** Default Bridge game configuration. */
export const DEFAULT_BRIDGE_CONFIG: BridgeConfig = {
  cpuDifficulty: 1,
};

/** CPU difficulty level options for Bridge. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Hook that manages Bridge game state, bidding, and player actions. */
export function useBridgeGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: bridgeConfig, handleConfigChange } = useGameConfig<BridgeConfig>(DEFAULT_BRIDGE_CONFIG);
  const [hint, setHint] = useState<BridgeHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(bridgeApi.exec, { onSuccess });

  const apiExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    apiExec('reset', undefined, undefined, undefined, undefined, DEFAULT_BRIDGE_CONFIG);
  }, [apiExec]);

  const handleBid = useCallback(
    (bidType: number, bidLevel?: number, bidSuit?: number) => {
      apiExec('bid', undefined, bidType, bidLevel, bidSuit);
    },
    [apiExec],
  );

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    apiExec('play', selectedCardIndices[0]);
  }, [apiExec, selectedCardIndices]);

  const handleNextTrick = useCallback(() => {
    apiExec('next');
  }, [apiExec]);

  const handleNextRound = useCallback(() => {
    apiExec('nextround');
  }, [apiExec]);

  const handleHint = useHintRequest({
    fetchHint: () => bridgeApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
    setHintLoading,
  });

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
    apiExec,
    bridgeConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
