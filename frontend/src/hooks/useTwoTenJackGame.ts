import { useCallback } from 'react';
import { twoTenJackApi } from '../api/gameApi';
import type { TwoTenJackConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Two Ten Jack game configuration. */
export const DEFAULT_TWOTENJACK_CONFIG: TwoTenJackConfig = {
  cpuDifficulty: 1,
  pointLimit: 50,
};

/** CPU difficulty level options for Two Ten Jack. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Two Ten Jack (cumulative total to win). */
export const POINT_LIMIT_OPTIONS = [30, 50, 75, 100] as const;

/**
 * Hook that manages Two Ten Jack game state and player actions.
 * Wraps {@link useTrickGameBase} and exposes a trump suit declaration handler.
 */
export function useTwoTenJackGame() {
  const {
    exec: dispatch,
    config,
    ...rest
  } = useTrickGameBase({
    apiFn: twoTenJackApi.exec,
    defaultConfig: DEFAULT_TWOTENJACK_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const handleDeclare = useCallback(
    (trumpSuit: number) => {
      dispatch('declare', trumpSuit);
    },
    [dispatch],
  );

  return { ...rest, exec: dispatch, twoTenJackConfig: config, handleDeclare };
}
