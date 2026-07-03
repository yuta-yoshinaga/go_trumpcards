import { useCallback } from 'react';
import { type BouillotteConfigInput, bouillotteApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Bouillotte game configuration. */
export const DEFAULT_BOUILLOTTE_CONFIG: Required<BouillotteConfigInput> = {
  playerCount: 4,
  ante: 10,
  startingChips: 200,
  targetRounds: 10,
};

/** Available player-count options for Bouillotte. */
export const PLAYER_COUNT_OPTIONS = [3, 4] as const;

/** Available ante options for Bouillotte. */
export const ANTE_OPTIONS = [5, 10, 20, 50] as const;

/** Available starting-chips options for Bouillotte. */
export const STARTING_CHIPS_OPTIONS = [100, 200, 500] as const;

/** Available target-rounds options for Bouillotte. */
export const TARGET_ROUNDS_OPTIONS = [5, 10, 20] as const;

/**
 * Hook that manages Bouillotte game state and the pot-vying betting actions.
 *
 * Bouillotte is not trick-taking, so like Guts and Three Card Brag it builds
 * its command set directly on {@link useGameApi}. On the human's Betting turn
 * the player calls, raises (vie), or folds; when betting closes the round
 * resolves. `nextround` deals the following round (chips persist server-side).
 * Config (player count, ante, starting chips, target rounds) is only accepted
 * on `reset`.
 */
export function useBouillotteGame() {
  const { config, handleConfigChange } = useGameConfig<Required<BouillotteConfigInput>>(DEFAULT_BOUILLOTTE_CONFIG);

  const { state, loading, error, exec, retry } = useGameApi(bouillotteApi.exec);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', undefined, config);
  }, [exec, config]);

  /** Calls — the human matches the current bet and stays in the round. */
  const handleCall = useCallback(() => {
    void exec('bet', 'call');
  }, [exec]);

  /** Raises (vie) — the human increases the bet by a fixed increment. */
  const handleRaise = useCallback(() => {
    void exec('bet', 'raise');
  }, [exec]);

  /** Folds — the human drops out of the current round. */
  const handleFold = useCallback(() => {
    void exec('bet', 'fold');
  }, [exec]);

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
    bouillotteConfig: config,
    handleConfigChange,
    reset,
    handleCall,
    handleRaise,
    handleFold,
    handleNextRound,
  };
}
