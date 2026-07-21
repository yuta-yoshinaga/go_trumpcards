import { useCallback } from 'react';
import { spadesApi } from '../api/gameApi';
import type { SpadesConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Spades game configuration. */
export const DEFAULT_SPADES_CONFIG: SpadesConfig = {
  cpuDifficulty: 1,
  pointLimit: 500,
  nilBonus: 100,
  bagPenaltyThreshold: 10,
};

/** CPU difficulty level options for Spades. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Spades. */
export const POINT_LIMIT_OPTIONS = [200, 300, 500, 750, 1000] as const;

/** Available nil-bonus options for Spades (backend accepts 0-500). */
export const NIL_BONUS_OPTIONS = [50, 100, 150, 200] as const;

/** Available bag-penalty threshold options for Spades (backend accepts 1-100). */
export const BAG_PENALTY_THRESHOLD_OPTIONS = [5, 10, 15] as const;

/** Hook that manages Spades game state, bidding, and player actions. */
export function useSpadesGame() {
  const base = useTrickGameBase({
    apiFn: spadesApi.exec,
    defaultConfig: DEFAULT_SPADES_CONFIG,
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
    spadesConfig: base.config,
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
