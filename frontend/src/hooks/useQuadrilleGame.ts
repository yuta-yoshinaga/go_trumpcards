import { useCallback } from 'react';
import { type QuadrilleConfigInput, quadrilleApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Quadrille game configuration. */
export const DEFAULT_QUADRILLE_CONFIG: Required<QuadrilleConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 5,
};

/** CPU difficulty level options for Quadrille. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target deal-count options for Quadrille (match length; highest cumulative score wins). */
export const TARGET_ROUNDS_OPTIONS = [3, 5, 7] as const;

/**
 * Hook that manages Quadrille game state and its player actions: bid
 * (pass/entrar/solo with a chosen trump suit), play a card, plus trick/round
 * advancement.
 *
 * Quadrille is a 3-player soloist-vs-coalition trick-taker on a 40-card Spanish
 * deck. A Bid phase decides the Quadrille (bid winner), who names a trump suit and
 * plays alone against the coalition of the other two players. The command set
 * is built directly on {@link useGameApi}.
 */
export function useQuadrilleGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<QuadrilleConfigInput>>(DEFAULT_QUADRILLE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(quadrilleApi.exec, { onSuccess });

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

  /**
   * Calls the King of `suit` in the KingCall phase, giving the bidder a partner.
   *
   * **Without this the board freezes right after the auction**: play is
   * rejected until a king is called.
   */
  const handleCallKing = useCallback(
    (suit: number) => {
      void exec('king', { kingSuit: suit });
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
    quadrilleConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleBid,
    handleCallKing,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
