import { useCallback } from 'react';
import { type TarocchiniConfigInput, tarocchiniApi } from '../api/gameApi';
import { TAROCCHINI_SURPLUS } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Tarocchini game configuration. */
export const DEFAULT_TAROCCHINI_CONFIG: Required<TarocchiniConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 4,
};

/** CPU difficulty level options for Tarocchini. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Round-count options for Tarocchini.
 *
 * **Multiples of the player count only.** The deal rotates each round, so a
 * non-multiple ends the match with someone having taken the scarto more often
 * than the rest — the backend rejects it outright.
 */
export const TARGET_ROUNDS_OPTIONS = [4, 8, 12] as const;

/**
 * Hook that manages Tarocchini game state: the dealer's scarto, the play
 * action, and trick/round advancement.
 *
 * Tarocchini is a 4-player Bolognese trick-taker in fixed 2v2 teams on a
 * 62-card tarot deck. There is no bidding — the dealer buries 2 surplus cards
 * and play begins.
 */
export function useTarocchiniGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<TarocchiniConfigInput>>(DEFAULT_TAROCCHINI_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(tarocchiniApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Buries the selected surplus cards. Requires exactly TAROCCHINI_SURPLUS. */
  const handleScarto = useCallback(() => {
    if (selectedCardIndices.length !== TAROCCHINI_SURPLUS) return;
    void exec('scarto', { cardIndices: selectedCardIndices });
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
    tarocchiniConfig: config,
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
