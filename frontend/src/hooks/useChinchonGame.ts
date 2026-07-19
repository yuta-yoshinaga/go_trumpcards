import { useCallback, useEffect } from 'react';
import { chinchonApi } from '../api/gameApi';
import type { ChinchonConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Chinchón game configuration. */
export const DEFAULT_CHINCHON_CONFIG: ChinchonConfig = {
  cpuDifficulty: 1,
  playerCount: 2,
  knockThreshold: 5,
  eliminationLimit: 100,
};

/** CPU difficulty level options for Chinchón. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available player count options for Chinchón. */
export const PLAYER_COUNT_OPTIONS = [2, 3, 4] as const;

/** Available knock (deadwood) threshold options for Chinchón. */
export const KNOCK_THRESHOLD_OPTIONS = [0, 3, 5, 10] as const;

/** Available elimination limit options for Chinchón. */
export const ELIMINATION_LIMIT_OPTIONS = [50, 100, 150, 200] as const;

/** Hook that manages Chinchón game state and player actions. */
export function useChinchonGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: chinchonConfig, handleConfigChange } = useGameConfig<ChinchonConfig>(DEFAULT_CHINCHON_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(chinchonApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, DEFAULT_CHINCHON_CONFIG);
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
    chinchonConfig,
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
    retry,
  };
}
