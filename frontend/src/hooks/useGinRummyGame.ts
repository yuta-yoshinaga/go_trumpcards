import { useCallback, useEffect } from 'react';
import { ginrummyApi } from '../api/gameApi';
import type { GinRummyConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Gin Rummy game configuration. */
export const DEFAULT_GINRUMMY_CONFIG: GinRummyConfig = {
  cpuDifficulty: 1,
  pointLimit: 100,
};

/** CPU difficulty level options for Gin Rummy. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Gin Rummy. */
export const POINT_LIMIT_OPTIONS = [50, 100, 150, 200] as const;

/** Hook that manages Gin Rummy game state and player actions. */
export function useGinRummyGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: ginRummyConfig, handleConfigChange } = useGameConfig<GinRummyConfig>(DEFAULT_GINRUMMY_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec } = useGameApi(ginrummyApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, DEFAULT_GINRUMMY_CONFIG);
  }, [exec]);

  const handleDrawStock = useCallback(() => {
    exec('drawstock');
  }, [exec]);

  const handleDrawDiscard = useCallback(() => {
    exec('drawdiscard');
  }, [exec]);

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('discard', selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleKnock = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('knock', selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleLayoff = useCallback(() => {
    exec('layoff', undefined, undefined, selectedCardIndices);
  }, [exec, selectedCardIndices]);

  const handleSkipLayoff = useCallback(() => {
    exec('layoff', undefined, undefined, []);
  }, [exec]);

  const handleNextRound = useCallback(() => {
    exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    ginRummyConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleKnock,
    handleLayoff,
    handleSkipLayoff,
    handleNextRound,
  };
}
