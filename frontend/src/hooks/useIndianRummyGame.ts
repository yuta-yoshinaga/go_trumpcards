import { useCallback, useEffect } from 'react';
import { indianRummyApi } from '../api/gameApi';
import type { IndianRummyConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Indian Rummy game configuration. */
export const DEFAULT_INDIANRUMMY_CONFIG: IndianRummyConfig = {
  playerCount: 4,
  cpuDifficulty: 1,
  targetRounds: 3,
};

/** CPU difficulty level options for Indian Rummy. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available player-count options for Indian Rummy (2-4 players). */
export const PLAYER_COUNT_OPTIONS = [2, 3, 4] as const;

/** Available target-round options for Indian Rummy. */
export const TARGET_ROUNDS_OPTIONS = [1, 3, 5, 10] as const;

/** Hook that manages Indian Rummy game state and player actions. */
export function useIndianRummyGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: indianRummyConfig, handleConfigChange } =
    useGameConfig<IndianRummyConfig>(DEFAULT_INDIANRUMMY_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(indianRummyApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, DEFAULT_INDIANRUMMY_CONFIG);
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

  const handleDeclare = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('declare', selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleNextRound = useCallback(() => {
    exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    indianRummyConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleDeclare,
    handleNextRound,
    retry,
  };
}
