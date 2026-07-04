import { useCallback } from 'react';
import { type MichiganConfigInput, michiganApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Michigan game configuration. */
export const DEFAULT_MICHIGAN_CONFIG: Required<MichiganConfigInput> = {
  playerCount: 4,
  ante: 8,
  startingChips: 200,
  targetRounds: 10,
};

/** Available player-count options for Michigan. */
export const PLAYER_COUNT_OPTIONS = [3, 4, 5, 6, 8] as const;

/** Available ante options for Michigan (distributed across the four boodles). */
export const ANTE_OPTIONS = [4, 8, 12, 20] as const;

/** Available starting-chips options for Michigan. */
export const STARTING_CHIPS_OPTIONS = [100, 200, 500] as const;

/** Available target-rounds options for Michigan. */
export const TARGET_ROUNDS_OPTIONS = [5, 10, 20] as const;

/**
 * Hook that manages Michigan (Newmarket) game state and its "stops" actions.
 *
 * Michigan is a chip-betting stops game, so like Guts and Bouillotte it builds
 * its command set directly on {@link useGameApi}. On the human's Bet turn the
 * player distributes chips across the four boodles; during the Play phase the
 * player plays cards in ascending same-suit sequences. `nextround` deals the
 * following round (chips persist server-side). Config (player count, ante,
 * starting chips, target rounds) is only accepted on `reset`.
 */
export function useMichiganGame() {
  const { config, handleConfigChange } = useGameConfig<Required<MichiganConfigInput>>(DEFAULT_MICHIGAN_CONFIG);

  const { state, loading, error, exec, retry } = useGameApi(michiganApi.exec);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', undefined, undefined, config);
  }, [exec, config]);

  /** Places the human's boodle bets — a length-4 chip distribution summing to betBudget. */
  const handleBet = useCallback(
    (boodleBets: number[]) => {
      void exec('bet', boodleBets);
    },
    [exec],
  );

  /** Plays the hand card at the given index (must be a legal, playable index). */
  const handlePlay = useCallback(
    (cardIndex: number) => {
      void exec('play', undefined, cardIndex);
    },
    [exec],
  );

  /** Advances to the next round after the current one resolves. */
  const handleNextRound = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    michiganConfig: config,
    handleConfigChange,
    reset,
    handleBet,
    handlePlay,
    handleNextRound,
  };
}
