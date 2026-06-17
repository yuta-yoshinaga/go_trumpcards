import { useCallback } from 'react';
import { type TeenPattiConfigInput, teenPattiApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Teen Patti game configuration. */
export const DEFAULT_TEEN_PATTI_CONFIG: Required<TeenPattiConfigInput> = {
  cpuDifficulty: 1,
  ante: 1,
  startingChips: 100,
};

/** CPU difficulty level options for Teen Patti. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available ante (per-deal stake) options for Teen Patti. */
export const ANTE_OPTIONS = [1, 2, 5] as const;

/** Available starting-chips options for Teen Patti. */
export const STARTING_CHIPS_OPTIONS = [50, 100, 200] as const;

/**
 * Hook that manages Teen Patti game state and the vying (betting) player
 * actions.
 *
 * Teen Patti is the Indian variant of Three Card Brag; like Mus it builds its
 * command set directly on {@link useGameApi}. On the human's Betting turn the
 * player may See (reveal their hand, Blind→Seen), Bet (call the stake), Raise
 * (increase the stake), Fold, Show (force a showdown when two players remain),
 * or request a Side Show (a private hand comparison with the previous Seen
 * player). When a Side Show is requested of the human, the human responds by
 * accepting or declining it. `next` advances to the following deal.
 */
export function useTeenPattiGame() {
  const { config, handleConfigChange } = useGameConfig<Required<TeenPattiConfigInput>>(DEFAULT_TEEN_PATTI_CONFIG);

  const { state, loading, error, exec, retry } = useGameApi(teenPattiApi.exec);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Reveals the human's hand (Blind → Seen). */
  const handleSee = useCallback(() => {
    void exec('see');
  }, [exec]);

  /** Bets — calls the current stake (Blind pays the stake, Seen pays double). */
  const handleBet = useCallback(() => {
    void exec('bet');
  }, [exec]);

  /** Raises the stake to the given amount. */
  const handleRaise = useCallback(
    (raiseStake: number) => {
      void exec('raise', { raiseStake });
    },
    [exec],
  );

  /** Folds out of the current deal. */
  const handleFold = useCallback(() => {
    void exec('fold');
  }, [exec]);

  /** Forces a showdown (only legal for a Seen player when two players remain). */
  const handleShow = useCallback(() => {
    void exec('show');
  }, [exec]);

  /** Requests a Side Show with the previous Seen player. */
  const handleSideShow = useCallback(() => {
    void exec('sideshow');
  }, [exec]);

  /** Responds to a Side Show request (accept or decline). */
  const handleRespondSideShow = useCallback(
    (accept: boolean) => {
      void exec('respond', { accept });
    },
    [exec],
  );

  /** Advances to the next deal after a deal ends. */
  const handleNextRound = useCallback(() => {
    void exec('next');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    teenPattiConfig: config,
    handleConfigChange,
    reset,
    handleSee,
    handleBet,
    handleRaise,
    handleFold,
    handleShow,
    handleSideShow,
    handleRespondSideShow,
    handleNextRound,
  };
}
