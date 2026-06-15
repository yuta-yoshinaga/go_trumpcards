import { useCallback } from 'react';
import { type SoloWhistConfigInput, soloWhistApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Solo Whist game configuration. */
export const DEFAULT_SOLO_WHIST_CONFIG: Required<SoloWhistConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 21,
};

/** CPU difficulty level options for Solo Whist. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Solo Whist (first player to reach wins). */
export const TARGET_POINTS_OPTIONS = [11, 21, 31] as const;

/**
 * Hook that manages Solo Whist game state, the bid action, the play action, and
 * trick/round advancement.
 *
 * Solo Whist is a British 4-player 52-card trick-taker with a bidding phase.
 * Each player bids once (Pass/Solo/Misère/Abundance); the highest bidder becomes
 * the declarer who plays alone against the other 3 defenders. Trump (for
 * Solo/Abundance) is the declarer's longest suit, chosen automatically.
 */
export function useSoloWhistGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<SoloWhistConfigInput>>(DEFAULT_SOLO_WHIST_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(soloWhistApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Submits a bid (0=Pass 1=Solo 2=Misère 3=Abundance). */
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
    soloWhistConfig: config,
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
