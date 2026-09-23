import { useCallback } from 'react';
import { batakApi } from '../api/gameApi';
import type { BatakConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Batak game configuration (5 fixed rounds, Normal CPU). */
export const DEFAULT_BATAK_CONFIG: BatakConfig = {
  cpuDifficulty: 1,
  maxRounds: 5,
};

/** CPU difficulty level options for Batak. */
export const BATAK_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Total-round options for Batak (default 5 = standard rule). */
export const BATAK_MAX_ROUNDS_OPTIONS = [3, 5, 7, 10] as const;

/** Hook that manages Batak game state, bidding, and player actions. */
export function useBatakGame() {
  const { exec, config, ...rest } = useTrickGameBase({
    apiFn: batakApi.exec,
    defaultConfig: DEFAULT_BATAK_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const handleBid = useCallback(
    (bid: number) => {
      exec('bid', bid);
    },
    [exec],
  );

  return { ...rest, exec, batakConfig: config, handleBid };
}
