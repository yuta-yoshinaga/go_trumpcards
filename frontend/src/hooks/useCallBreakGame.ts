import { useCallback } from 'react';
import { callBreakApi } from '../api/gameApi';
import type { CallBreakConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Call Break game configuration (5 fixed rounds, Normal CPU). */
export const DEFAULT_CALLBREAK_CONFIG: CallBreakConfig = {
  cpuDifficulty: 1,
  maxRounds: 5,
};

/** CPU difficulty level options for Call Break. */
export const CALLBREAK_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Total-round options for Call Break (default 5 = standard rule). */
export const CALLBREAK_MAX_ROUNDS_OPTIONS = [3, 5, 7, 10] as const;

/** Hook that manages Call Break game state, bidding, and player actions. */
export function useCallBreakGame() {
  const base = useTrickGameBase({
    apiFn: callBreakApi.exec,
    defaultConfig: DEFAULT_CALLBREAK_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const { exec } = base;

  const handleBid = useCallback(
    (bid: number) => {
      exec('bid', bid);
    },
    [exec],
  );

  return {
    state: base.state,
    loading: base.loading,
    error: base.error,
    hint: base.hint,
    hintError: base.hintError,
    hintLoading: base.hintLoading,
    exec: base.exec,
    callBreakConfig: base.config,
    selectedCardIndices: base.selectedCardIndices,
    toggleCard: base.toggleCard,
    clearSelection: base.clearSelection,
    handleConfigChange: base.handleConfigChange,
    handleBid,
    handlePlay: base.handlePlay,
    handleNextTrick: base.handleNextTrick,
    handleNextRound: base.handleNextRound,
    handleHint: base.handleHint,
    retry: base.retry,
  };
}
