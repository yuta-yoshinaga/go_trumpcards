import { useCallback } from 'react';
import { type KlaverjasConfigInput, klaverjasApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Klaverjas game configuration. */
export const DEFAULT_KLAVERJAS_CONFIG: Required<KlaverjasConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 1501,
};

/** CPU difficulty level options for Klaverjas. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Klaverjas (first team to reach wins). */
export const TARGET_POINTS_OPTIONS = [1001, 1501, 2001] as const;

/**
 * Hook that manages Klaverjas game state and the single player action (play a
 * card) plus trick/round advancement.
 *
 * Klaverjas is a Dutch 4-player (2 vs 2) trump trick-taker with Roem bonuses.
 * The only legal move is playing a card (must follow suit / overtrump); there
 * are no declarations. The command set is built directly on {@link useGameApi}.
 */
export function useKlaverjasGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<KlaverjasConfigInput>>(DEFAULT_KLAVERJAS_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(klaverjasApi.exec, { onSuccess });

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
    klaverjasConfig: config,
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
