import { useCallback } from 'react';
import { type FortyFivesConfigInput, fortyFivesApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Auction Forty-Fives game configuration. */
export const DEFAULT_FORTY_FIVES_CONFIG: Required<FortyFivesConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 45,
};

/** CPU difficulty level options for Auction Forty-Fives. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Auction Forty-Fives (first team to reach wins). */
export const TARGET_POINTS_OPTIONS = [30, 45, 60] as const;

/**
 * Hook that manages Auction Forty-Fives game state, the bid action, the play
 * action, and trick/round advancement.
 *
 * Auction Forty-Fives is an Irish/Canadian 4-player, 2-team (seats 0&2 vs 1&3)
 * trick-taker with a bidding phase. Each player bids once (Pass/15/20/25); the
 * highest bidder's team declares trump and plays five tricks. Each trick is
 * worth 5 points. The first team to reach the target points (default 45) wins.
 */
export function useFortyFivesGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<FortyFivesConfigInput>>(DEFAULT_FORTY_FIVES_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(fortyFivesApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Submits a bid (0=Pass 15 20 25). */
  const handleBid = useCallback(
    (bid: number) => {
      void exec('bid', { bid });
    },
    [exec],
  );

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Advances to the next trick. */
  const handleNextTrick = useCallback(() => {
    void exec('next');
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
    fortyFivesConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
