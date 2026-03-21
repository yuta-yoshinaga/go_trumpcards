import { useCallback, useEffect } from 'react';
import { spadesApi } from '../api/gameApi';
import type { SpadesConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Spades game configuration. */
export const DEFAULT_SPADES_CONFIG: SpadesConfig = {
  cpuDifficulty: 1,
  pointLimit: 500,
  nilBonus: 100,
  bagPenaltyThreshold: 10,
};

/** CPU difficulty level options for Spades. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Spades. */
export const POINT_LIMIT_OPTIONS = [200, 300, 500, 750, 1000] as const;

/** Hook that manages Spades game state, bidding, and player actions. */
export function useSpadesGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: spadesConfig, handleConfigChange } = useGameConfig<SpadesConfig>(DEFAULT_SPADES_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec } = useGameApi(spadesApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, undefined, DEFAULT_SPADES_CONFIG);
  }, [exec]);

  const handleBid = useCallback(
    (bid: number) => {
      exec('bid', bid);
    },
    [exec],
  );

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('play', undefined, selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleNextTrick = useCallback(() => {
    exec('next');
  }, [exec]);

  const handleNextRound = useCallback(() => {
    exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    spadesConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
