import { useCallback } from 'react';
import { type GutsConfigInput, gutsApi } from '../api/gameApi';
import { GutsDeclaration } from '../types/phases';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Guts game configuration. */
export const DEFAULT_GUTS_CONFIG: Required<GutsConfigInput> = {
  playerCount: 4,
  ante: 10,
  startingChips: 200,
  targetRounds: 10,
};

/** Available player-count options for Guts. */
export const PLAYER_COUNT_OPTIONS = [2, 3, 4, 5, 6, 7] as const;

/** Available ante options for Guts. */
export const ANTE_OPTIONS = [5, 10, 20, 50] as const;

/** Available starting-chips options for Guts. */
export const STARTING_CHIPS_OPTIONS = [100, 200, 500] as const;

/** Available target-rounds options for Guts. */
export const TARGET_ROUNDS_OPTIONS = [5, 10, 20] as const;

/**
 * Hook that manages Guts game state and the pot-vying player actions.
 *
 * Guts is not trick-taking, so like Three Card Brag it builds its command set
 * directly on {@link useGameApi}. On the human's Declare turn the player calls
 * In (stay) or Out (fold), which resolves the round; `nextround` deals the
 * following round (chips persist server-side). Config (player count, ante,
 * starting chips, target rounds) is only accepted on `reset`.
 */
export function useGutsGame() {
  const { config, handleConfigChange } = useGameConfig<Required<GutsConfigInput>>(DEFAULT_GUTS_CONFIG);

  const { state, loading, error, exec, retry } = useGameApi(gutsApi.exec);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', undefined, config);
  }, [exec, config]);

  /** Declares In (stay) — the human keeps their hand and vies for the pot. */
  const handleIn = useCallback(() => {
    void exec('declare', GutsDeclaration.IN);
  }, [exec]);

  /** Declares Out (fold) — the human drops out of the round. */
  const handleOut = useCallback(() => {
    void exec('declare', GutsDeclaration.OUT);
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
    gutsConfig: config,
    handleConfigChange,
    reset,
    handleIn,
    handleOut,
    handleNextRound,
  };
}
