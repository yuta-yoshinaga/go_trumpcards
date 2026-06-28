import { useCallback, useEffect } from 'react';
import { prsiApi } from '../api/gameApi';
import type { PrsiConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Prší game configuration (no point limit — first to empty hand wins). */
export const DEFAULT_PRSI_CONFIG: PrsiConfig = {
  cpuDifficulty: 1,
};

/** CPU difficulty level options for Prší. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Hook that manages Prší game state and player actions. */
export function usePrsiGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: prsiConfig, handleConfigChange } = useGameConfig<PrsiConfig>(DEFAULT_PRSI_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec, retry } = useGameApi(prsiApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset', undefined, DEFAULT_PRSI_CONFIG);
  }, [exec]);

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('play', selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleDraw = useCallback(() => {
    exec('draw');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    prsiConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
    retry,
  };
}
