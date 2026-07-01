import { useCallback } from 'react';
import { type LooConfigInput, looApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Loo (Lanterloo) game configuration. */
export const DEFAULT_LOO_CONFIG: Required<LooConfigInput> = {
  cpuDifficulty: 1,
  ante: 3,
};

/** CPU difficulty level options for Loo. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available ante (per-deal pot contribution) options for Loo. */
export const ANTE_OPTIONS = [1, 3, 5] as const;

/**
 * Hook that manages Loo (Lanterloo) game state and its player actions: decide
 * (play/pass), play a card, plus deal advancement.
 *
 * Loo is a 4-player 52-card pot-based gambling trick-taker. Trump is set from the
 * turn-up card (no bidding, no trump naming). Each player decides to play or pass;
 * participants fight five must-follow / must-head tricks. There is no game-over
 * target — it is a repeating deal loop. The command set is built directly on
 * {@link useGameApi}.
 */
export function useLooGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<LooConfigInput>>(DEFAULT_LOO_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(looApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Declares the participation decision in the Decide phase (true=play, false=pass). */
  const handleDecide = useCallback(
    (play: boolean) => {
      void exec('decide', { play });
    },
    [exec],
  );

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Advances to the next deal. */
  const handleNextDeal = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    looConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleDecide,
    handlePlay,
    handleNextDeal,
  };
}
