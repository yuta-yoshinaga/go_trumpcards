import { useCallback } from 'react';
import { type ManilleConfigInput, manilleApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Manille game configuration. */
export const DEFAULT_MANILLE_CONFIG: Required<ManilleConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 101,
};

/** CPU difficulty level options for Manille. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Manille (first team to reach wins). */
export const TARGET_POINTS_OPTIONS = [51, 101, 151] as const;

/**
 * Hook that manages Manille game state and the single player action (play a
 * card) plus trick/round advancement.
 *
 * Manille is a French/Belgian 4-player (2 vs 2) trump trick-taker. The only
 * legal move is playing a card (must follow suit / trump when void unless the
 * partner already holds the trick); there are no declarations and no Roem. The
 * command set is built directly on {@link useGameApi}.
 */
export function useManilleGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<ManilleConfigInput>>(DEFAULT_MANILLE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(manilleApi.exec, { onSuccess });

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
    manilleConfig: config,
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
