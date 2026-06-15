import { useCallback } from 'react';
import { type SedmaConfigInput, sedmaApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Sedma game configuration. */
export const DEFAULT_SEDMA_CONFIG: Required<SedmaConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 101,
};

/** CPU difficulty level options for Sedma. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Sedma (first team to reach wins). */
export const TARGET_POINTS_OPTIONS = [51, 101, 151] as const;

/**
 * Hook that manages Sedma game state and the single player action (play a
 * card) plus trick/round advancement.
 *
 * Sedma is a Czech/Slovak 4-player (2 vs 2) no-trump capture trick-taker. Any
 * card may be played (there is no follow obligation); a card captures the
 * trick when its rank equals the lead card's rank or it is a 7 (wild). There
 * are no declarations. The command set is built directly on {@link useGameApi}.
 */
export function useSedmaGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<SedmaConfigInput>>(DEFAULT_SEDMA_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(sedmaApi.exec, { onSuccess });

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
    sedmaConfig: config,
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
