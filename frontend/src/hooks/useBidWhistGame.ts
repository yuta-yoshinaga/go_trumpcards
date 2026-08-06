import { useCallback, useEffect } from 'react';
import { bidWhistApi } from '../api/gameApi';
import type { BidWhistConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Bid Whist game configuration. */
export const DEFAULT_BID_WHIST_CONFIG: BidWhistConfig = {
  cpuDifficulty: 1,
  targetScore: 7,
};

/** CPU difficulty level options for Bid Whist. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target score options for Bid Whist. */
export const TARGET_SCORE_OPTIONS = [7, 9, 11] as const;

/** Hook that manages Bid Whist game state: bidding, trump declaration, kitty exchange, and play. */
export function useBidWhistGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<BidWhistConfig>(DEFAULT_BID_WHIST_CONFIG);

  const onSuccess = useCallback(() => clearSelection(), [clearSelection]);
  const { state, loading, error, exec, retry } = useGameApi(bidWhistApi.exec, { onSuccess });
  const apiCall = useCallback((...args: Parameters<typeof exec>) => exec(...args), [exec]);

  useEffect(() => {
    apiCall('reset', { config: DEFAULT_BID_WHIST_CONFIG });
  }, [apiCall]);

  const bid = useCallback(
    (tricks: number, direction: number) => apiCall('bid', { bidTricks: tricks, bidDirection: direction }),
    [apiCall],
  );
  const pass = useCallback(() => apiCall('pass'), [apiCall]);
  const declareTrump = useCallback((suit: number) => apiCall('trump', { trumpSuit: suit }), [apiCall]);
  const exchange = useCallback((discardIndices: number[]) => apiCall('exchange', { discardIndices }), [apiCall]);
  const play = useCallback((cardIndex: number) => apiCall('play', { cardIndex }), [apiCall]);
  const nextTrick = useCallback(() => apiCall('next'), [apiCall]);
  const nextRound = useCallback(() => apiCall('nextround'), [apiCall]);
  const reset = useCallback(() => apiCall('reset', { config }), [apiCall, config]);
  // サーバーのヒントを要求する。CUI には HintOutput があるのに、Web からも CLI
  // からも到達できず事実上のデッドコードだった (#4814)。
  const requestHint = useCallback(() => apiCall('hint'), [apiCall]);

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
    requestHint,
    bid,
    pass,
    declareTrump,
    exchange,
    play,
    nextTrick,
    nextRound,
    reset,
  };
}
