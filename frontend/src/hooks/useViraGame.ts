import { useCallback } from 'react';
import { type ViraConfigInput, viraApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Vira game configuration. */
export const DEFAULT_PREFERENCE_CONFIG: Required<ViraConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 30,
};

/** CPU difficulty level options for Vira. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Vira (first player to reach wins). */
export const TARGET_ROUNDS_OPTIONS = [20, 30, 40] as const;

/**
 * Hook that manages Vira game state, the bid action, the play action, and
 * trick/round advancement.
 *
 * Vira is a Russian/Austrian 3-player 32-card trick-taker with a bidding
 * phase. Each player bids once (Pass/Six/Misère/Seven/Eight); the highest bidder
 * becomes the declarer who plays alone against the other 2 defenders. Trump (for
 * Six/Seven/Eight) is the declarer's longest suit, chosen automatically; Misère
 * has no trump.
 */
export function useViraGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<ViraConfigInput>>(DEFAULT_PREFERENCE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(viraApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Submits a bid (0=Pass 1=Six 2=Misère 3=Seven 4=Eight). */
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
    viraConfig: config,
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
