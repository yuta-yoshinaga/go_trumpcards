import { useCallback, useEffect } from 'react';
import { shitheadApi } from '../api/gameApi';
import type { ShitheadConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Shithead game configuration. */
export const DEFAULT_SHITHEAD_CONFIG: ShitheadConfig = {
  magicTwo: true,
  magicSeven: true,
  magicEight: true,
  magicTen: true,
  fourOfAKindBurn: true,
  cpuDifficulty: 1,
};

/** CPU difficulty options for Shithead. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Hook that manages Shithead game state and player actions. */
export function useShitheadGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: shitheadConfig, handleConfigChange } = useGameConfig<ShitheadConfig>(DEFAULT_SHITHEAD_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec: dispatch, retry } = useGameApi(shitheadApi.exec, { onSuccess });

  useEffect(() => {
    dispatch('reset', { config: DEFAULT_SHITHEAD_CONFIG });
  }, [dispatch]);

  const handlePlay = useCallback(() => {
    dispatch('play', { indices: selectedCardIndices });
  }, [dispatch, selectedCardIndices]);

  const handlePickup = useCallback(() => {
    dispatch('play', { indices: [] });
  }, [dispatch]);

  const handleResetWithConfig = useCallback(
    (config: ShitheadConfig) => {
      dispatch('reset', { config });
    },
    [dispatch],
  );

  const handleReset = useCallback(() => {
    dispatch('reset', { config: DEFAULT_SHITHEAD_CONFIG });
  }, [dispatch]);

  return {
    state,
    loading,
    error,
    dispatch,
    shitheadConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handlePickup,
    handleResetWithConfig,
    handleReset,
    retry,
  };
}
