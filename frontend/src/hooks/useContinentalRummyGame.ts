import { useCallback, useEffect } from 'react';
import { continentalrummyApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Continental Rummy settings. */
export const DEFAULT_CONTINENTAL_RUMMY_CONFIG = { cpuDifficulty: 1, totalRounds: 3 };

/** Available CPU difficulty levels for Continental Rummy. */
export const CONTINENTAL_RUMMY_DIFFICULTY_OPTIONS = [0, 1, 2] as const;
/** Minimum supported number of Continental Rummy rounds. */
export const CONTINENTAL_RUMMY_MIN_ROUNDS = 1;
/** Maximum supported number of Continental Rummy rounds. */
export const CONTINENTAL_RUMMY_MAX_ROUNDS = 10;

/** Manages Continental Rummy state and settings. */
export function useContinentalRummyGame() {
  const { config, handleConfigChange } = useGameConfig(DEFAULT_CONTINENTAL_RUMMY_CONFIG);
  const game = useGameApi(continentalrummyApi.exec);
  useEffect(() => {
    void game.exec('reset');
  }, [game.exec]);
  const handleResetWithConfig = useCallback(() => {
    void game.exec('reset', { config });
  }, [game.exec, config]);
  return { ...game, config, handleConfigChange, handleResetWithConfig };
}
