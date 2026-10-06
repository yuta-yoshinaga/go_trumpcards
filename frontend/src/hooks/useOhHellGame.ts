import { useCallback } from 'react';
import { ohHellApi } from '../api/gameApi';
import type { OhHellConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Oh Hell game configuration. */
export const DEFAULT_OH_HELL_CONFIG: OhHellConfig = {
  cpuDifficulty: 1,
  maxHandSize: 10,
  scoringVariant: 0,
  roundDirection: 1,
};

/** CPU difficulty level options for Oh Hell. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available max hand size options for Oh Hell. */
export const MAX_HAND_SIZE_OPTIONS = [3, 5, 7, 10, 13] as const;

/** Scoring variant options for Oh Hell. */
export const SCORING_VARIANT_OPTIONS = [
  { value: 0, labelKey: 'scoringStandard' },
  { value: 1, labelKey: 'scoringPenalty' },
] as const;

/** Round direction options for Oh Hell. */
export const ROUND_DIRECTION_OPTIONS = [
  { value: 0, labelKey: 'roundDownOnly' },
  { value: 1, labelKey: 'roundDownAndUp' },
] as const;

/** Hook that manages Oh Hell game state, bidding, and player actions. */
export function useOhHellGame() {
  const {
    exec,
    config: ohHellConfig,
    ...base
  } = useTrickGameBase({
    apiFn: ohHellApi.exec,
    defaultConfig: DEFAULT_OH_HELL_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const handleBid = useCallback(
    (bid: number) => {
      exec('bid', bid);
    },
    [exec],
  );

  return {
    ...base,
    exec,
    ohHellConfig,
    handleBid,
  };
}
