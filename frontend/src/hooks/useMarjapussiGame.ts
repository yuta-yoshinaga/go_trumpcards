import { useCallback } from 'react';
import { type MarjapussiConfigInput, marjapussiApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Marjapussi game configuration. */
export const DEFAULT_MARJAPUSSI_CONFIG: Required<MarjapussiConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 500,
};

/** CPU difficulty level options for Marjapussi. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Marjapussi (first team to reach wins). */
export const TARGET_POINTS_OPTIONS = [300, 500, 1000] as const;

/**
 * Hook that manages Marjapussi game state and its player actions: play a card,
 * plus trick/round advancement.
 *
 * Marjapussi is a Finnish 4-player 2-vs-2 trick-taker. The command set is built
 * directly on {@link useGameApi}.
 */
export function useMarjapussiGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<MarjapussiConfigInput>>(DEFAULT_MARJAPUSSI_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(marjapussiApi.exec, { onSuccess });

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
    marjapussiConfig: config,
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
