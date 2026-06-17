import { useCallback } from 'react';
import { type ThreeCardBragConfigInput, threeCardBragApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Three Card Brag game configuration. */
export const DEFAULT_THREE_CARD_BRAG_CONFIG: Required<ThreeCardBragConfigInput> = {
  cpuDifficulty: 1,
  ante: 1,
  startingChips: 100,
};

/** CPU difficulty level options for Three Card Brag. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available ante (per-deal stake) options for Three Card Brag. */
export const ANTE_OPTIONS = [1, 2, 5] as const;

/** Available starting-chips options for Three Card Brag. */
export const STARTING_CHIPS_OPTIONS = [50, 100, 200] as const;

/**
 * Hook that manages Three Card Brag game state and the vying (betting) player
 * actions.
 *
 * Three Card Brag is not trick-taking, so like Mus it builds its command set
 * directly on {@link useGameApi}. On the human's Betting turn the player may
 * See (reveal their hand, Blind→Seen), Bet (call the stake), Raise (increase
 * the stake), Fold, or Show (force a showdown when two players remain). `next`
 * advances to the following deal.
 */
export function useThreeCardBragGame() {
  const { config, handleConfigChange } =
    useGameConfig<Required<ThreeCardBragConfigInput>>(DEFAULT_THREE_CARD_BRAG_CONFIG);

  const { state, loading, error, exec, retry } = useGameApi(threeCardBragApi.exec);

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
    threeCardBragConfig: config,
    handleConfigChange,
    reset,
    handleSee,
    handleBet,
    handleRaise,
    handleFold,
    handleShow,
    handleNextRound,
  };
}
