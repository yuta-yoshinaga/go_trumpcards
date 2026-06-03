import { useCallback, useEffect } from 'react';
import { fiveHundredApi } from '../api/gameApi';
import type { FiveHundredConfig } from '../types/card';
import { FiveHundredContract } from '../types/phases';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default 500 (Five Hundred) game configuration. */
export const DEFAULT_FIVE_HUNDRED_CONFIG: FiveHundredConfig = {
  cpuDifficulty: 1,
  targetScore: 500,
};

/** CPU difficulty level options for 500. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target score options for 500. */
export const TARGET_SCORE_OPTIONS = [300, 500, 700] as const;

/** Hook that manages 500 (Five Hundred) game state, bidding, kitty exchange, and play. */
export function useFiveHundredGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<FiveHundredConfig>(DEFAULT_FIVE_HUNDRED_CONFIG);

  const onSuccess = useCallback(() => clearSelection(), [clearSelection]);
  const { state, loading, error, exec, retry } = useGameApi(fiveHundredApi.exec, { onSuccess });
  const apiCall = useCallback((...args: Parameters<typeof exec>) => exec(...args), [exec]);

  useEffect(() => {
    apiCall('reset', { config: DEFAULT_FIVE_HUNDRED_CONFIG });
  }, [apiCall]);

  const bidSuit = useCallback(
    (tricks: number, suit: number) =>
      apiCall('bid', { bidKind: FiveHundredContract.SUIT, bidTricks: tricks, bidSuit: suit }),
    [apiCall],
  );
  const bidNoTrump = useCallback(
    (tricks: number) => apiCall('bid', { bidKind: FiveHundredContract.NO_TRUMP, bidTricks: tricks }),
    [apiCall],
  );
  const bidMisere = useCallback(() => apiCall('bid', { bidKind: FiveHundredContract.MISERE }), [apiCall]);
  const bidOpenMisere = useCallback(() => apiCall('bid', { bidKind: FiveHundredContract.OPEN_MISERE }), [apiCall]);
  const pass = useCallback(() => apiCall('pass'), [apiCall]);
  const exchange = useCallback((discardIndices: number[]) => apiCall('exchange', { discardIndices }), [apiCall]);
  const play = useCallback(
    (cardIndex: number, jokerSuit?: number) => apiCall('play', { cardIndex, jokerSuit }),
    [apiCall],
  );
  const nextTrick = useCallback(() => apiCall('next'), [apiCall]);
  const nextRound = useCallback(() => apiCall('nextround'), [apiCall]);
  const reset = useCallback(() => apiCall('reset', { config }), [apiCall, config]);

  return {
    state,
    loading,
    error,
    retry,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    config,
    handleConfigChange,
    apiCall,
    bidSuit,
    bidNoTrump,
    bidMisere,
    bidOpenMisere,
    pass,
    exchange,
    play,
    nextTrick,
    nextRound,
    reset,
  };
}
