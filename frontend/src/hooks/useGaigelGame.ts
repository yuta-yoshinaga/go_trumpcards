import { useCallback } from 'react';
import { gaigelApi } from '../api/gameApi';
import type { GaigelConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Gaigel game configuration. */
export const DEFAULT_GAIGEL_CONFIG: GaigelConfig = {
  cpuDifficulty: 1,
  targetScore: 101,
};

/** CPU difficulty level options for Gaigel. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-score options for Gaigel. */
export const TARGET_SCORE_OPTIONS = [101, 201, 301] as const;

/** Hook that manages Gaigel game state and player actions. */
export function useGaigelGame() {
  const { exec, config, ...rest } = useTrickGameBase({
    apiFn: gaigelApi.exec,
    defaultConfig: DEFAULT_GAIGEL_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const handleMarriage = useCallback(
    (cardIndex: number) => {
      void (exec as unknown as (command: string, a1?: number, ci?: number) => Promise<void>)(
        'marriage',
        undefined,
        cardIndex,
      );
    },
    [exec],
  );

  return { ...rest, exec, gaigelConfig: config, handleMarriage };
}
