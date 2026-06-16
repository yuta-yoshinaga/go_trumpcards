import { useCallback } from 'react';
import { type BeziqueConfigInput, beziqueApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Bezique game configuration. */
export const DEFAULT_BEZIQUE_CONFIG: Required<BeziqueConfigInput> = {
  cpuDifficulty: 1,
  targetScore: 1000,
};

/** CPU difficulty level options for Bezique. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-score options for Bezique (first player to reach wins). */
export const TARGET_SCORE_OPTIONS = [500, 1000, 1500] as const;

/**
 * Hook that manages Bezique game state, the play action, meld declaration /
 * skip during the Meld phase, and deal advancement.
 *
 * Bezique is a 2-player ancestor of Pinochle played with a 64-card deck. While
 * the stock remains (phase 1), the trick winner may declare one meld for points
 * or skip; both then draw. When the stock empties, phase 2 is strict must-follow
 * with no melds and a +10 last-trick bonus. Scores accumulate across deals to a
 * target (default 1000); the higher total wins.
 */
export function useBeziqueGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<BeziqueConfigInput>>(DEFAULT_BEZIQUE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(beziqueApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Declares the meld at the given index during the Meld phase. */
  const handleMeld = useCallback(
    (meldIndex: number) => {
      void exec('meld', { meldIndex });
    },
    [exec],
  );

  /** Skips declaring a meld during the Meld phase. */
  const handleSkipMeld = useCallback(() => {
    void exec('skip');
  }, [exec]);

  /** Advances to the next deal after a round (deal) ends. */
  const handleNextRound = useCallback(() => {
    void exec('next');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    beziqueConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handlePlay,
    handleMeld,
    handleSkipMeld,
    handleNextRound,
  };
}
