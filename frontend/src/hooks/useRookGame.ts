import { useCallback, useEffect } from 'react';
import { rookApi } from '../api/gameApi';
import type { RookConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Rook (ルーク) game configuration. */
export const DEFAULT_ROOK_CONFIG: RookConfig = {
  cpuDifficulty: 1,
  targetScore: 500,
};

/** CPU difficulty level options for Rook. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target score options for Rook. */
export const TARGET_SCORE_OPTIONS = [300, 500, 700] as const;

/** Hook that manages Rook (ルーク) game state, bidding, nest exchange, and play. */
export function useRookGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<RookConfig>(DEFAULT_ROOK_CONFIG);

  const onSuccess = useCallback(() => clearSelection(), [clearSelection]);
  const { state, loading, error, exec, retry } = useGameApi(rookApi.exec, { onSuccess });
  const apiCall = useCallback((...args: Parameters<typeof exec>) => exec(...args), [exec]);

  useEffect(() => {
    apiCall('reset', { config: DEFAULT_ROOK_CONFIG });
  }, [apiCall]);

  const bid = useCallback((value: number) => apiCall('bid', { bid: value }), [apiCall]);
  const pass = useCallback(() => apiCall('pass'), [apiCall]);
  const exchange = useCallback(
    (discardIndices: number[], trumpColor: number) => apiCall('exchange', { discardIndices, trumpColor }),
    [apiCall],
  );
  const play = useCallback((cardIndex: number) => apiCall('play', { cardIndex }), [apiCall]);
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
    bid,
    pass,
    exchange,
    play,
    nextTrick,
    nextRound,
    reset,
  };
}
