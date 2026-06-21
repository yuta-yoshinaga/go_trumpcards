import { useCallback } from 'react';
import { type KnockoutWhistConfigInput, knockoutWhistApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Knockout Whist game configuration (CPU difficulty only). */
export const DEFAULT_KNOCKOUT_WHIST_CONFIG: Required<KnockoutWhistConfigInput> = {
  cpuDifficulty: 1,
};

/** CPU difficulty level options for Knockout Whist. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Hook that manages Knockout Whist game state and the single player action
 * (play a card) plus trick/round advancement.
 *
 * Knockout Whist is a British play-only survival trick-taker for 4 players.
 * Each round the hand shrinks by one card; the previous round's winner's
 * longest suit becomes trump (auto). Must-follow, Ace-high. Winning zero
 * tricks in a round costs a Dogbone token — or eliminates a player who has
 * none left. Last player standing wins. The command set is built directly on
 * {@link useGameApi}.
 */
export function useKnockoutWhistGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } =
    useGameConfig<Required<KnockoutWhistConfigInput>>(DEFAULT_KNOCKOUT_WHIST_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(knockoutWhistApi.exec, { onSuccess });

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
    knockoutWhistConfig: config,
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
