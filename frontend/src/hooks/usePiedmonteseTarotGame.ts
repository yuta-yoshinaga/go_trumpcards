import { useCallback } from 'react';
import { type PiedmonteseTarotConfigInput, piedmonteseTarotApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Tarocco Piemontese game configuration. */
export const DEFAULT_PIEDMONTESE_TAROT_CONFIG: Required<PiedmonteseTarotConfigInput> = {
  seats: 4,
  cpuDifficulty: 1,
  targetDeals: 4,
};

/** CPU difficulty level options. */
export const PIEDMONTESE_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Table sizes the game deals for.
 *
 * **Only 3 and 4.** The deal is part of the rules — 25 cards each with a
 * three-card talon, or 19 each with a two-card talon — so a size the rules do
 * not name has no deal at all, and the server refuses it rather than rounding.
 */
export const PIEDMONTESE_SEAT_OPTIONS = [3, 4] as const;

/** Available match lengths (highest cumulative score wins). */
export const PIEDMONTESE_TARGET_DEALS_OPTIONS = [2, 4, 6] as const;

/**
 * Hook that manages Tarocco Piemontese state and its player actions: the
 * dealer's scarto (burying the talon), playing a card, and trick/deal
 * advancement.
 */
export function usePiedmonteseTarotGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<PiedmonteseTarotConfigInput>>(
    DEFAULT_PIEDMONTESE_TAROT_CONFIG,
  );

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(piedmonteseTarotApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Buries the selected cards.
   *
   * **The count comes from the table, not from a constant** — four seats bury
   * two cards, three seats bury three.
   */
  const handleScarto = useCallback(
    (talonSize: number) => {
      if (selectedCardIndices.length !== talonSize) return;
      void exec('scarto', { cardIndices: [...selectedCardIndices] });
    },
    [exec, selectedCardIndices],
  );

  /** Plays the single currently-selected card. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Advances to the next trick. */
  const handleNextTrick = useCallback(() => {
    void exec('next');
  }, [exec]);

  /** Advances to the next deal. */
  const handleNextRound = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    piedmonteseConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleScarto,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
