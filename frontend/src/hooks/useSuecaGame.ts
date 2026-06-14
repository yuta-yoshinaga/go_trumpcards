import { useCallback } from 'react';
import { type SuecaConfigInput, suecaApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Sueca game configuration. */
export const DEFAULT_SUECA_CONFIG: Required<SuecaConfigInput> = {
  cpuDifficulty: 1,
  targetGamePoints: 4,
};

/** CPU difficulty level options for Sueca. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target game-point options for Sueca (first team to reach wins the match). */
export const TARGET_GAME_POINTS_OPTIONS = [2, 4, 6, 8] as const;

/**
 * Hook that manages Sueca game state and the single player action (play a card)
 * plus trick/round advancement.
 *
 * Sueca is a Portuguese/Brazilian 4-player (2 vs 2) trump trick-taker. The only
 * legal move is playing a card (must follow suit); unlike Tute there are no
 * marriage or Tute declarations. The command set is built directly on
 * {@link useGameApi}.
 */
export function useSuecaGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<SuecaConfigInput>>(DEFAULT_SUECA_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(suecaApi.exec, { onSuccess });

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
    suecaConfig: config,
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
