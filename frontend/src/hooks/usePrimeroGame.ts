import { useCallback } from 'react';
import { type PrimeroConfigInput, primeroApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Primero game configuration. */
export const DEFAULT_PRIMERO_CONFIG: Required<PrimeroConfigInput> = {
  playerCount: 4,
  ante: 10,
  startingChips: 200,
  targetRounds: 10,
};

/** Available player-count options for Primero. */
export const PLAYER_COUNT_OPTIONS = [2, 3, 4, 5, 6] as const;

/** Available ante options for Primero. */
export const ANTE_OPTIONS = [5, 10, 20, 50] as const;

/** Available starting-chips options for Primero. */
export const STARTING_CHIPS_OPTIONS = [100, 200, 500] as const;

/** Available target-rounds options for Primero. */
export const TARGET_ROUNDS_OPTIONS = [5, 10, 20] as const;

/**
 * Hook that manages Primero game state and the pot-vying betting actions.
 *
 * Primero is not trick-taking, so like Guts and Bouillotte it builds its
 * command set directly on {@link useGameApi}. On the human's Betting turn the
 * player calls, raises (vie), or folds; when betting closes the round resolves.
 * `nextround` deals the following round (chips persist server-side). Config
 * (player count, ante, starting chips, target rounds) is only accepted on
 * `reset`.
 */
export function usePrimeroGame() {
  const { config, handleConfigChange } = useGameConfig<Required<PrimeroConfigInput>>(DEFAULT_PRIMERO_CONFIG);

  const { state, loading, error, exec, retry } = useGameApi(primeroApi.exec);

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
    primeroConfig: config,
    handleConfigChange,
    reset,
    handleCall,
    handleRaise,
    handleFold,
    handleNextRound,
  };
}
