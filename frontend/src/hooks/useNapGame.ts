import { useCallback } from 'react';
import { type NapConfigInput, napApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Nap (Napoleon) game configuration. */
export const DEFAULT_NAP_CONFIG: Required<NapConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 20,
};

/** CPU difficulty level options for Nap. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target chip-count options for Nap (first player to reach wins). */
export const TARGET_POINTS_OPTIONS = [10, 20, 30] as const;

/**
 * Hook that manages Nap (Napoleon) game state, the bid action, the play action,
 * and trick/round advancement.
 *
 * Nap is a British 4-player 5-card gambling trick-taker with a bidding phase.
 * Each player bids once (Pass/Two/Three/Four/Nap = how many of the 5 tricks they
 * will take); the highest bidder becomes the declarer who picks trump (auto =
 * longest suit) and leads. Make the bid to win chips; fail and the opponents
 * each win the bid value. First player to targetPoints chips wins.
 */
export function useNapGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<NapConfigInput>>(DEFAULT_NAP_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(napApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Submits a bid (0=Pass 2=Two 3=Three 4=Four 5=Nap). */
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
    napConfig: config,
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
