import { useCallback } from 'react';
import { gongzhuApi } from '../api/gameApi';
import type { GongZhuConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Gong Zhu game configuration. */
export const DEFAULT_GONGZHU_CONFIG: GongZhuConfig = {
  cpuDifficulty: 1,
  pointLimit: 1000,
};

/** CPU difficulty level options for Gong Zhu. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available end-game point limit options for Gong Zhu. */
export const POINT_LIMIT_OPTIONS = [500, 1000, 1500, 2000] as const;

/** Hook that manages Gong Zhu game state and player actions. */
export function useGongZhuGame() {
  const { exec, selectedCardIndices, config, ...rest } = useTrickGameBase({
    apiFn: gongzhuApi.exec,
    defaultConfig: DEFAULT_GONGZHU_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const handleExpose = useCallback(() => {
    exec('expose', selectedCardIndices);
  }, [exec, selectedCardIndices]);

  return { ...rest, exec, gongzhuConfig: config, selectedCardIndices, handleExpose };
}
