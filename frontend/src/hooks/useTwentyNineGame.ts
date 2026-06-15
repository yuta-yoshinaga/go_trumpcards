import { useCallback } from 'react';
import { type TwentyNineConfigInput, twentyNineApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Twenty-Nine (29) game configuration. */
export const DEFAULT_TWENTY_NINE_CONFIG: Required<TwentyNineConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 6,
};

/** CPU difficulty level options for Twenty-Nine (29). */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target game-point options for Twenty-Nine (29) (first team to reach wins). */
export const TARGET_POINTS_OPTIONS = [4, 6, 8] as const;

/**
 * Hook that manages Twenty-Nine (29) game state, the bid action, the play
 * action, and trick/round advancement.
 *
 * Twenty-Nine is an Indian/Bangladeshi 4-player, 2-team (seats 0&2 vs 1&3)
 * trick-taker with a bidding phase and a hidden trump. Each player bids once
 * (Pass/16/20/24/28); the highest bidder's team picks a hidden trump suit
 * (revealed only mid-play) and plays eight tricks. Card points (J=3, 9=2, A=1,
 * 10=1) plus the last trick total 29. A made bid scores +1 game point; a
 * failed bid gives +1 to the other team. The first team to reach the target
 * game points (default 6) wins.
 */
export function useTwentyNineGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<TwentyNineConfigInput>>(DEFAULT_TWENTY_NINE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(twentyNineApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Submits a bid (0=Pass 16 20 24 28). */
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
    twentyNineConfig: config,
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
