import { useCallback } from 'react';
import { type AnacondaConfigInput, anacondaApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Anaconda game configuration. */
export const DEFAULT_ANACONDA_CONFIG: Required<AnacondaConfigInput> = {
  playerCount: 4,
  ante: 10,
  startingChips: 200,
  targetRounds: 10,
};

/** Available player-count options for Anaconda. */
export const PLAYER_COUNT_OPTIONS = [3, 4, 5, 6, 7] as const;

/** Available ante options for Anaconda. */
export const ANTE_OPTIONS = [5, 10, 20, 50] as const;

/** Available starting-chips options for Anaconda. */
export const STARTING_CHIPS_OPTIONS = [100, 200, 500] as const;

/** Available target-rounds options for Anaconda. */
export const TARGET_ROUNDS_OPTIONS = [5, 10, 20] as const;

/**
 * Hook that manages Anaconda (Pass the Trash) game state and the poker
 * pot-vying player actions.
 *
 * Anaconda flows through four phases: Pass (select `passCount` cards to pass
 * left), Set (keep exactly 5 cards, discarding 2), Roll (reveal cards one at a
 * time with a call/raise/fold betting round between reveals), and Result. The
 * best 5-card poker hand wins the pot; `nextround` deals the next round (chips
 * persist server-side). Config (player count, ante, starting chips, target
 * rounds) is only accepted on `reset`.
 */
export function useAnacondaGame() {
  const { config, handleConfigChange } = useGameConfig<Required<AnacondaConfigInput>>(DEFAULT_ANACONDA_CONFIG);

  const { state, loading, error, exec, retry } = useGameApi(anacondaApi.exec);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', undefined, undefined, config);
  }, [exec, config]);

  /** Passes the selected cards left (3→2→1). */
  const handlePass = useCallback(
    (indices: number[]) => {
      void exec('pass', indices);
    },
    [exec],
  );

  /** Keeps exactly 5 cards, discarding the other 2. */
  const handleKeep = useCallback(
    (indices: number[]) => {
      void exec('keep', indices);
    },
    [exec],
  );

  /** Calls (or checks) the current bet during the Roll phase. */
  const handleCall = useCallback(() => {
    void exec('bet', undefined, 'call');
  }, [exec]);

  /** Raises the current bet during the Roll phase. */
  const handleRaise = useCallback(() => {
    void exec('bet', undefined, 'raise');
  }, [exec]);

  /** Folds out of the current round during the Roll phase. */
  const handleFold = useCallback(() => {
    void exec('bet', undefined, 'fold');
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
    anacondaConfig: config,
    handleConfigChange,
    reset,
    handlePass,
    handleKeep,
    handleCall,
    handleRaise,
    handleFold,
    handleNextRound,
  };
}
