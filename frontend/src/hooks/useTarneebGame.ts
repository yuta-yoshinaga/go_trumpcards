import { useCallback } from 'react';
import { tarneebApi } from '../api/gameApi';
import type { TarneebConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Tarneeb configuration: Normal CPU, 31-point target, minimum bid 7. */
export const DEFAULT_TARNEEB_CONFIG: TarneebConfig = {
  cpuDifficulty: 1,
  pointLimit: 31,
  minBid: 7,
};

/** CPU difficulty options exposed in the settings panel. */
export const TARNEEB_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Selectable point-limit targets for Tarneeb (31 is the canonical default). */
export const TARNEEB_POINT_LIMIT_OPTIONS = [21, 31, 41, 51] as const;

/** Selectable minimum-bid values for Tarneeb (7 = standard, half + 1 of 13 tricks). */
export const TARNEEB_MIN_BID_OPTIONS = [5, 6, 7, 8] as const;

/** Hook that manages Tarneeb game state, bidding, trump declaration, and play. */
export function useTarneebGame() {
  const base = useTrickGameBase({
    apiFn: tarneebApi.exec,
    defaultConfig: DEFAULT_TARNEEB_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const runAction = base.exec;

  const handleBid = useCallback(
    (bid: number) => {
      runAction('bid', bid);
    },
    [runAction],
  );

  const handleDeclareTrump = useCallback(
    (suit: number) => {
      runAction('trump', suit);
    },
    [runAction],
  );

  return {
    state: base.state,
    loading: base.loading,
    error: base.error,
    hint: base.hint,
    hintError: base.hintError,
    hintLoading: base.hintLoading,
    exec: base.exec,
    tarneebConfig: base.config,
    selectedCardIndices: base.selectedCardIndices,
    toggleCard: base.toggleCard,
    clearSelection: base.clearSelection,
    handleConfigChange: base.handleConfigChange,
    handleBid,
    handleDeclareTrump,
    handlePlay: base.handlePlay,
    handleNextTrick: base.handleNextTrick,
    handleNextRound: base.handleNextRound,
    handleHint: base.handleHint,
    retry: base.retry,
  };
}
