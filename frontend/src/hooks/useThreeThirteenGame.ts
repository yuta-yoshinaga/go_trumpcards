import { useCallback, useEffect } from 'react';
import { threethirteenApi } from '../api/gameApi';
import type { ThreeThirteenConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Three Thirteen game configuration. */
export const DEFAULT_THREETHIRTEEN_CONFIG: ThreeThirteenConfig = {
  cpuDifficulty: 1,
  playerCount: 2,
};

/** CPU difficulty level options for Three Thirteen. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available player count options for Three Thirteen. */
export const PLAYER_COUNT_OPTIONS = [2, 3, 4] as const;

/** Hook that manages Three Thirteen game state and player actions. */
export function useThreeThirteenGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: threeThirteenConfig, handleConfigChange } =
    useGameConfig<ThreeThirteenConfig>(DEFAULT_THREETHIRTEEN_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(threethirteenApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, DEFAULT_THREETHIRTEEN_CONFIG);
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

  const handleNextRound = useCallback(() => {
    exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    threeThirteenConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleKnock,
    handleNextRound,
    retry,
  };
}
