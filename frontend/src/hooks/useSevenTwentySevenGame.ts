import { useCallback } from 'react';
import { type SevenTwentySevenConfigInput, sevenTwentySevenApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default SevenTwentySeven game configuration. */
export const DEFAULT_SEVENTWENTYSEVEN_CONFIG: Required<SevenTwentySevenConfigInput> = {
  playerCount: 4,
  ante: 10,
  startingChips: 200,
  targetRounds: 10,
};

/** Available player-count options for SevenTwentySeven. */
export const PLAYER_COUNT_OPTIONS = [2, 3, 4, 5, 6, 7] as const;

/** Available ante options for SevenTwentySeven. */
export const ANTE_OPTIONS = [5, 10, 20, 50] as const;

/** Available starting-chips options for SevenTwentySeven. */
export const STARTING_CHIPS_OPTIONS = [100, 200, 500] as const;

/** Available target-rounds options for SevenTwentySeven. */
export const TARGET_ROUNDS_OPTIONS = [5, 10, 20] as const;

/**
 * Hook that manages SevenTwentySeven game state and the draw actions.
 *
 * Each pass the human either takes another card or stands pat. **Standing is
 * final for the round** — the server plays out the remaining CPU draws and
 * settles, so there is nothing more to send until `nextround`. Config (player
 * count, ante, starting chips, target rounds) is only accepted on `reset`.
 */
export function useSevenTwentySevenGame() {
  const { config, handleConfigChange } = useGameConfig<Required<SevenTwentySevenConfigInput>>(
    DEFAULT_SEVENTWENTYSEVEN_CONFIG,
  );

  const { state, loading, error, exec, retry } = useGameApi(sevenTwentySevenApi.exec);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', config);
  }, [exec, config]);

  /** Takes one more card. */
  const handleTakeCard = useCallback(() => {
    void exec('card');
  }, [exec]);

  /** Stands pat — no more cards this round. */
  const handleStand = useCallback(() => {
    void exec('stand');
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
    sevenTwentySevenConfig: config,
    handleConfigChange,
    reset,
    handleTakeCard,
    handleStand,
    handleNextRound,
  };
}
