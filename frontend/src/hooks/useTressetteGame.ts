import { tressetteApi } from '../api/gameApi';
import type { TressetteConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Tressette game configuration. */
export const DEFAULT_TRESSETTE_CONFIG: TressetteConfig = {
  cpuDifficulty: 1,
  targetPoints: 21,
};

/** CPU difficulty level options for Tressette. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-point options for Tressette. */
export const TARGET_POINTS_OPTIONS = [11, 21, 31, 41] as const;

/** Hook that manages Tressette game state and player actions. */
export function useTressetteGame() {
  const { config, ...rest } = useTrickGameBase({
    apiFn: tressetteApi.exec,
    defaultConfig: DEFAULT_TRESSETTE_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  return { ...rest, tressetteConfig: config };
}
