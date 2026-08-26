import { trappolaApi } from '../api/gameApi';
import type { TrappolaConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Trappola game configuration. */
export const DEFAULT_TRAPPOLA_CONFIG: TrappolaConfig = {
  cpuDifficulty: 1,
  targetPoints: 21,
};

/** CPU difficulty level options for Trappola. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-point options for Trappola. */
export const TARGET_POINTS_OPTIONS = [11, 21, 31, 41] as const;

/** Hook that manages Trappola game state and player actions. */
export function useTrappolaGame() {
  const { config, ...rest } = useTrickGameBase({
    apiFn: trappolaApi.exec,
    defaultConfig: DEFAULT_TRAPPOLA_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  return { ...rest, trappolaConfig: config };
}
