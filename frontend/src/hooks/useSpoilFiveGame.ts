import { useCallback } from 'react';
import { type SpoilFiveConfigInput, spoilFiveApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Spoil Five game configuration (CPU difficulty only). */
export const DEFAULT_SPOIL_FIVE_CONFIG: Required<SpoilFiveConfigInput> = {
  cpuDifficulty: 1,
};

/** CPU difficulty level options for Spoil Five. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Hook that manages Spoil Five game state and the single player action
 * (play a card) plus trick/round advancement.
 *
 * Spoil Five (Maw) is an Irish play-only trick-taker for 5 players on a 52-card
 * deck (5 cards each). Trump is the turned-up card; the trump 5, trump J, and
 * ♥A are the fixed top trumps and may be held back (Reneging). The first player
 * to win 3 of the 5 tricks takes the pot; otherwise the round is a Spoil and
 * the pot carries over. First player to the target score wins. The command set
 * is built directly on {@link useGameApi}.
 */
export function useSpoilFiveGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<SpoilFiveConfigInput>>(DEFAULT_SPOIL_FIVE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(spoilFiveApi.exec, { onSuccess });

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
    spoilFiveConfig: config,
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
