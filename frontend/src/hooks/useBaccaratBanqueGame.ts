import { useCallback, useEffect } from 'react';
import { baccaratbanqueApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Baccarat Banque settings. */
export const DEFAULT_BACCARAT_BANQUE_CONFIG = { cpuDifficulty: 1, startChips: 1000 };

/** Available CPU difficulty levels for Baccarat Banque. */
export const BACCARAT_BANQUE_DIFFICULTY_OPTIONS = [0, 1, 2] as const;

/** Available starting chip amounts for Baccarat Banque. */
export const BACCARAT_BANQUE_START_CHIPS_OPTIONS = [500, 1000, 5000] as const;

/** Manages Baccarat Banque state and settings. */
export function useBaccaratBanqueGame() {
  const { config, handleConfigChange } = useGameConfig(DEFAULT_BACCARAT_BANQUE_CONFIG);
  const game = useGameApi(baccaratbanqueApi.exec);
  useEffect(() => {
    void game.exec('reset');
  }, [game.exec]);
  const handleResetWithConfig = useCallback(() => {
    void game.exec('reset', config);
  }, [game.exec, config]);
  return { ...game, config, handleConfigChange, handleResetWithConfig };
}
