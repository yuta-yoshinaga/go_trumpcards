import { madrassoApi } from '../api/gameApi';
import type { MadrassoConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Madrasso game configuration. */
export const DEFAULT_MADRASSO_CONFIG: MadrassoConfig = {
  cpuDifficulty: 1,
  targetPoints: 21,
};

/** CPU difficulty level options for Madrasso. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-point options for Madrasso. */
export const TARGET_POINTS_OPTIONS = [11, 21, 31, 41] as const;

/** Hook that manages Madrasso game state and player actions. */
export function useMadrassoGame() {
  const { config, ...rest } = useTrickGameBase({
    apiFn: madrassoApi.exec,
    defaultConfig: DEFAULT_MADRASSO_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  return { ...rest, madrassoConfig: config };
}
