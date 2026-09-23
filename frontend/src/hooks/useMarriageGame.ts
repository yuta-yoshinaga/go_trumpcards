import { useCallback, useEffect } from 'react';
import { marriageApi } from '../api/games/marriage';
import type { MarriageConfig } from '../types/games/marriage';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Marriage game configuration. */
export const DEFAULT_MARRIAGE_CONFIG: MarriageConfig = {
  playerCount: 5,
  cpuDifficulty: 1,
  targetRounds: 3,
};

/** CPU difficulty level options for Marriage. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available player-count options for Marriage (2-5 players). */
export const PLAYER_COUNT_OPTIONS = [2, 3, 4, 5] as const;

/** Available target-round options for Marriage. */
export const TARGET_ROUNDS_OPTIONS = [1, 3, 5, 10] as const;

/** Hook that manages Marriage game state and player actions. */
export function useMarriageGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: marriageConfig, handleConfigChange } = useGameConfig<MarriageConfig>(DEFAULT_MARRIAGE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(marriageApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    void exec('reset', undefined, DEFAULT_MARRIAGE_CONFIG);
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
    marriageConfig,
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
