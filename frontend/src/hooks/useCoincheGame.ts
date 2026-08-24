import { useCallback } from 'react';
import { coincheApi } from '../api/gameApi';
import type { CoincheConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Coinche game configuration. */
export const DEFAULT_COINCHE_CONFIG: CoincheConfig = {
  cpuDifficulty: 1,
  targetScore: 1000,
  dixDeDer: 10,
  enableBeloteRebelote: true,
};

/** CPU difficulty level options for Coinche. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-score options for Coinche. */
export const TARGET_SCORE_OPTIONS = [500, 750, 1000, 1500] as const;

/** Hook that manages Coinche game state and player actions. */
export function useCoincheGame() {
  const { exec, config, ...rest } = useTrickGameBase({
    apiFn: coincheApi.exec,
    defaultConfig: DEFAULT_COINCHE_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  /**
   * Bids a contract.
   *
   * **Both halves travel together.** A contract is a target *and* a trump
   * suit; sending one without the other would leave the server to fill in
   * the rest, which is a different contract from the one that was clicked.
   */
  const handleBid = useCallback(
    (points: number, suit: number) => {
      // 共通フックと同じ (command, arg1, arg2, config) の並び。bid は
      // arg1=目標点 / arg2=切り札スート。
      void (exec as unknown as (command: string, a1?: number, a2?: number) => Promise<void>)('bid', points, suit);
    },
    [exec],
  );

  const handlePass = useCallback(() => {
    void (exec as unknown as (command: string) => Promise<void>)('pass');
  }, [exec]);

  const handleCoinche = useCallback(() => {
    void (exec as unknown as (command: string) => Promise<void>)('coinche');
  }, [exec]);

  const handleSurcoinche = useCallback(() => {
    void (exec as unknown as (command: string) => Promise<void>)('surcoinche');
  }, [exec]);

  const handleDeclineDouble = useCallback(() => {
    void (exec as unknown as (command: string) => Promise<void>)('decline');
  }, [exec]);

  return {
    ...rest,
    exec,
    coincheConfig: config,
    handleBid,
    handlePass,
    handleCoinche,
    handleSurcoinche,
    handleDeclineDouble,
  };
}
