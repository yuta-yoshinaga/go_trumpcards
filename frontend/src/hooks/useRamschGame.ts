import { useCallback, useEffect, useState } from 'react';
import { ramschApi } from '../api/gameApi';
import type { RamschConfig, RamschHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useHintRequest } from './useHintRequest';

/** Default Ramsch game configuration. */
export const DEFAULT_RAMSCH_CONFIG: RamschConfig = {
  cpuDifficulty: 1,
  targetScore: 500,
};

/** CPU difficulty options for Ramsch. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-score options. */
export const TARGET_SCORE_OPTIONS = [100, 250, 500, 1000] as const;

/** Hook that manages Ramsch game state and player actions. */
export function useRamschGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: ramschConfig, handleConfigChange } = useGameConfig<RamschConfig>(DEFAULT_RAMSCH_CONFIG);
  const [hint, setHint] = useState<RamschHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);

  const { state, loading, error, exec: dispatch, retry } = useGameApi(ramschApi.exec, { onSuccess });

  /** Resets the game, applying the current config (CPU difficulty, target score). */
  const reset = useCallback(() => {
    dispatch('reset', { config: ramschConfig });
  }, [dispatch, ramschConfig]);

  // Fetch a fresh game on mount using the initial (default) config.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run reset once on mount with the initial config.
  useEffect(() => {
    reset();
  }, []);

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    dispatch('play', { cardIndex: selectedCardIndices[0] });
  }, [dispatch, selectedCardIndices]);

  const handleNextTrick = useCallback(() => {
    dispatch('next');
  }, [dispatch]);

  const handleNextRound = useCallback(() => {
    dispatch('nextround');
  }, [dispatch]);

  const handleHint = useHintRequest({
    fetchHint: () => ramschApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
    setHintLoading,
  });

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
    dispatch,
    ramschConfig,
    reset,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
