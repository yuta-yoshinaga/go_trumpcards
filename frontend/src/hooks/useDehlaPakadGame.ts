import { useCallback } from 'react';
import { type DehlaPakadConfigInput, dehlaPakadApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Dehla Pakad configuration. */
export const DEFAULT_DEHLA_PAKAD_CONFIG: Required<DehlaPakadConfigInput> = {
  cpuDifficulty: 1,
  targetKots: 2,
};

/** CPU difficulty options. */
export const DEHLA_PAKAD_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Match lengths, in kots. */
export const DEHLA_PAKAD_KOT_OPTIONS = [1, 2, 3, 5] as const;

/** Trump suits, in the order the domain numbers them. */
export const DEHLA_PAKAD_SUITS = [
  { value: 1, key: 'spade' },
  { value: 2, key: 'club' },
  { value: 3, key: 'heart' },
  { value: 4, key: 'diamond' },
] as const;

/**
 * Hook that manages Dehla Pakad state and its player actions: calling the
 * trump from the first five cards, playing a card, and moving to the next hand.
 */
export function useDehlaPakadGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<DehlaPakadConfigInput>>(DEFAULT_DEHLA_PAKAD_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(dehlaPakadApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Calls the trump suit. */
  const handleSelectTrump = useCallback(
    (trumpSuit: number) => {
      void exec('trump', { trumpSuit });
    },
    [exec],
  );

  /** Plays the selected card. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Advances to the next hand. */
  const handleNextHand = useCallback(() => {
    void exec('nexthand');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    dehlaPakadConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleSelectTrump,
    handlePlay,
    handleNextHand,
  };
}
