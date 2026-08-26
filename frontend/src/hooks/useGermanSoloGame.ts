import { useCallback } from 'react';
import { type GermanSoloConfigInput, germansoloApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default German Solo game configuration. */
export const DEFAULT_GERMAN_SOLO_CONFIG: Required<GermanSoloConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 5,
};

/** CPU difficulty level options for German Solo. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target deal-count options for German Solo (match length; highest cumulative score wins). */
export const TARGET_ROUNDS_OPTIONS = [3, 5, 7] as const;

/**
 * Hook that manages German Solo game state and its player actions: bid
 * (pass / Frage / Solo / Tout with a chosen trump suit), call the partner's ace,
 * play a card, plus trick/round advancement.
 *
 * German Solo is a 4-player trick-taker on a 32-card Skat pack. A Frage (or the
 * forced Mussfrage) lets the declarer call an ace they do not hold, and its
 * holder becomes a hidden partner, so play is two sides of two; Solo and Tout
 * are played alone against three. The command set is built directly on
 * {@link useGameApi}.
 */
export function useGermanSoloGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<GermanSoloConfigInput>>(DEFAULT_GERMAN_SOLO_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(germansoloApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Declares a bid in the Bid phase (0=pass, 2=Frage, 3=Solo, 4=Tout). Every
   * declaration but a pass carries the chosen trump suit (1=♠ 2=♣ 3=♥ 4=♦).
   *
   * **1 (Mussfrage) is never sent from here**: it is what the server forces on
   * the holder of Spadille when every seat passes.
   */
  const handleBid = useCallback(
    (bid: number, trumpSuit?: number) => {
      void exec('bid', { bid, trumpSuit });
    },
    [exec],
  );

  /**
   * Calls the ace of `suit` in the AceCall phase, giving the declarer a partner.
   *
   * **Without this the board freezes right after a Frage**: play is rejected
   * until an ace is called.
   */
  const handleCallAce = useCallback(
    (suit: number) => {
      void exec('ace', { aceSuit: suit });
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
    germansoloConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleBid,
    handleCallAce,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
