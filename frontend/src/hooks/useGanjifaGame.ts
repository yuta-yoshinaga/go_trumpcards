import { useCallback } from 'react';
import { type GanjifaConfigInput, ganjifaApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Ganjifa game configuration. */
export const DEFAULT_GANJIFA_CONFIG: Required<GanjifaConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 3,
};

/** CPU difficulty level options for Ganjifa. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available round-count options for Ganjifa (highest total after the last round wins). */
export const TARGET_ROUNDS_OPTIONS = [3, 6, 9] as const;

/**
 * Hook that manages Ganjifa game state, the play action, and trick/round
 * advancement.
 *
 * Ganjifa is a 3-player Mughal-era Indian trick-taker on a circular 96-card
 * deck (8 suits x 12 ranks), 32 tricks per round. There is no bidding: trump
 * is auto-declared from the dealer's longest suit each round, so `play` is the
 * only decision the human makes.
 */
export function useGanjifaGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<GanjifaConfigInput>>(DEFAULT_GANJIFA_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(ganjifaApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

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
    ganjifaConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
