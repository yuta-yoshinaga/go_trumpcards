import { useCallback } from 'react';
import { type MariasConfigInput, mariasApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Mariáš game configuration. */
export const DEFAULT_MARIAS_CONFIG: Required<MariasConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 10,
};

/** CPU difficulty level options for Mariáš. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target game-point options for Mariáš (first player to reach wins). */
export const TARGET_POINTS_OPTIONS = [5, 10, 15] as const;

/**
 * Hook that manages Mariáš game state and the single player action (play a
 * card) plus trick/round advancement.
 *
 * Mariáš is a Czech/Slovak 3-player 32-card trump trick-taker. Each round one
 * rotating Soloist plays alone against the 2 Defenders; trump is the Soloist's
 * longest suit (auto). The only legal move is playing a card (must follow suit
 * / trump when void); there are no declarations. The command set is built
 * directly on {@link useGameApi}.
 */
export function useMariasGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<MariasConfigInput>>(DEFAULT_MARIAS_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(mariasApi.exec, { onSuccess });

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
    mariasConfig: config,
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
