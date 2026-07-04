import { useCallback } from 'react';
import { type OmbreConfigInput, ombreApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Ombre (Hombre) game configuration. */
export const DEFAULT_OMBRE_CONFIG: Required<OmbreConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 5,
};

/** CPU difficulty level options for Ombre. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target deal-count options for Ombre (match length; highest cumulative score wins). */
export const TARGET_ROUNDS_OPTIONS = [3, 5, 7] as const;

/**
 * Hook that manages Ombre (Hombre) game state and its player actions: bid
 * (pass/entrar/solo with a chosen trump suit), play a card, plus trick/round
 * advancement.
 *
 * Ombre is a 3-player soloist-vs-coalition trick-taker on a 40-card Spanish
 * deck. A Bid phase decides the Ombre (bid winner), who names a trump suit and
 * plays alone against the coalition of the other two players. The command set
 * is built directly on {@link useGameApi}.
 */
export function useOmbreGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<OmbreConfigInput>>(DEFAULT_OMBRE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(ombreApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Declares a bid in the Bid phase (0=pass, 1=entrar, 2=solo). An entrar/solo
   * declaration carries the chosen trump suit (1=♠ 2=♣ 3=♥ 4=♦).
   */
  const handleBid = useCallback(
    (bid: number, trumpSuit?: number) => {
      void exec('bid', { bid, trumpSuit });
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
    ombreConfig: config,
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
