import { useCallback } from 'react';
import { type ScartoConfigInput, scartoApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Number of cards the dealer buries in the scarto (discard). */
export const SCARTO_DISCARD_COUNT = 3;

/** Default Scarto (スカルト) game configuration. */
export const DEFAULT_SCARTO_CONFIG: Required<ScartoConfigInput> = {
  cpuDifficulty: 1,
  targetDeals: 5,
};

/** CPU difficulty level options for Scarto. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target deal-count options (match length; highest cumulative score wins). */
export const TARGET_DEALS_OPTIONS = [3, 5, 7] as const;

/**
 * Hook that manages Scarto (スカルト) game state and its player actions: perform
 * the dealer's scarto (bury three low pip cards), play a card, and trick/round
 * advancement.
 *
 * Scarto is a simple 3-player Italian tarocchi trick-taker on the 78-card tarot
 * deck. The human is seat 0. There is no bidding, no chien exchange, and no
 * partnership — the command set is built directly on {@link useGameApi}.
 */
export function useScartoGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<ScartoConfigInput>>(DEFAULT_SCARTO_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(scartoApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Buries the 3 currently-selected scarto cards (dealer only, Scarto phase). */
  const handleScarto = useCallback(() => {
    if (selectedCardIndices.length !== SCARTO_DISCARD_COUNT) return;
    void exec('scarto', { cardIndices: [...selectedCardIndices] });
  }, [exec, selectedCardIndices]);

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
    scartoConfig: config,
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
