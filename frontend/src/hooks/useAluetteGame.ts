import { useCallback } from 'react';
import { type AluetteConfigInput, aluetteApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Aluette game configuration. */
export const DEFAULT_ALUETTE_CONFIG: Required<AluetteConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 5,
};

/** CPU difficulty level options for Aluette. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Match-length options for Aluette, counted in menes won. */
export const TARGET_POINTS_OPTIONS = [3, 5, 7] as const;

/**
 * Hook that manages Aluette game state: the play action and trick/mene
 * advancement.
 *
 * Aluette is a 4-player Breton trick-taker in fixed 2v2 teams on a 48-card
 * Spanish-suited deck. There is no trump suit, no bidding and no follow
 * obligation — six named cards (the luettes) outrank everything else.
 */
export function useAluetteGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<AluetteConfigInput>>(DEFAULT_ALUETTE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(aluetteApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Plays the single currently-selected card. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Advances to the next trick. */
  const handleNextTrick = useCallback(() => {
    void exec('next');
  }, [exec]);

  /** Advances to the next mene. */
  const handleNextRound = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    aluetteConfig: config,
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
