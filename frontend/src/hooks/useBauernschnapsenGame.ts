import { useCallback } from 'react';
import { bauernschnapsenApi } from '../api/gameApi';
import type { BauernschnapsenConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Bauernschnapsen game configuration. */
export const DEFAULT_BAUERNSCHNAPSEN_CONFIG: BauernschnapsenConfig = {
  cpuDifficulty: 1,
  targetScore: 24,
};

/** CPU difficulty level options for Bauernschnapsen. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Available target-score options for Bauernschnapsen.
 *
 * Scores are **game values** (Rufer 2, Farbenzwang 6, Bettel 5), not card
 * points, so the ladder is far shorter than the clone source's 101/201/301.
 */
export const TARGET_SCORE_OPTIONS = [24, 36, 48] as const;

/** Hook that manages Bauernschnapsen game state and player actions. */
export function useBauernschnapsenGame() {
  const { exec, config, ...rest } = useTrickGameBase({
    apiFn: bauernschnapsenApi.exec,
    defaultConfig: DEFAULT_BAUERNSCHNAPSEN_CONFIG,
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

  /**
   * Declares a contract for the human seat.
   *
   * **Without this the board is stuck**: a fresh round opens in the contract
   * phase, and play is rejected until a contract is settled.
   */
  const handleContract = useCallback(
    (contract: number, trumpSuit: number) => {
      void (
        exec as unknown as (command: string, a1?: number, ci?: number, cfg?: undefined, suit?: number) => Promise<void>
      )('contract', contract, undefined, undefined, trumpSuit);
    },
    [exec],
  );

  return { ...rest, exec, bauernschnapsenConfig: config, handleMarriage, handleContract };
}
