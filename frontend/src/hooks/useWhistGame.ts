import { whistApi } from '../api/gameApi';
import type { WhistConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Whist game configuration. */
export const DEFAULT_WHIST_CONFIG: WhistConfig = {
  cpuDifficulty: 1,
  pointLimit: 5,
};

/** CPU difficulty level options for Whist. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Whist. */
export const POINT_LIMIT_OPTIONS = [3, 5, 7, 10] as const;

/** Hook that manages Whist game state and player actions. */
export function useWhistGame() {
  const { config, ...rest } = useTrickGameBase({
    apiFn: whistApi.exec,
    defaultConfig: DEFAULT_WHIST_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  return { ...rest, whistConfig: config };
}
