import { useCallback } from 'react';
import { type CometConfigInput, cometApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Comet configuration. */
export const DEFAULT_COMET_CONFIG: Required<CometConfigInput> = {
  cpuDifficulty: 1,
  players: 4,
  targetScore: 100,
};

/** CPU difficulty options. */
export const COMET_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Seat-count options. */
export const COMET_PLAYER_OPTIONS = [2, 3, 4, 5] as const;

/** Target-score options. */
export const COMET_TARGET_OPTIONS = [20, 50, 100, 200] as const;

/**
 * Hook that manages Comet state and its player actions: playing a card, or
 * passing when nothing is playable.
 *
 * **Which cards are playable comes from the server.** The Comet (9 of
 * diamonds) stands in for any rank, so the page reads `playableIdxs` rather
 * than re-deriving the legal plays.
 */
export function useCometGame() {
  const { config, handleConfigChange } = useGameConfig<Required<CometConfigInput>>(DEFAULT_COMET_CONFIG);
  const { state, loading, error, exec, retry } = useGameApi(cometApi.exec);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Plays the hand card at `idx`. */
  const play = useCallback(
    (idx: number) => {
      void exec('play', { handIndex: idx });
    },
    [exec],
  );

  /** Passes. Refused by the server when a playable card is held. */
  const pass = useCallback(() => {
    void exec('pass');
  }, [exec]);

  /** Advances to the next round. */
  const handleNextRound = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    cometConfig: config,
    handleConfigChange,
    play,
    pass,
    handleNextRound,
    reset,
  };
}
