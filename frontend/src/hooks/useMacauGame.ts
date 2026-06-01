import { useCallback, useEffect } from 'react';
import { macauApi } from '../api/gameApi';
import type { MacauConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Macau game configuration. */
export const DEFAULT_MACAU_CONFIG: MacauConfig = {
  cpuDifficulty: 1,
  pointLimit: 200,
};

/** CPU difficulty level options for Macau. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Macau. */
export const POINT_LIMIT_OPTIONS = [100, 200, 300, 500] as const;

/** Hook that manages Macau game state and player actions. */
export function useMacauGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: macauConfig, handleConfigChange } = useGameConfig<MacauConfig>(DEFAULT_MACAU_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(macauApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, undefined, DEFAULT_MACAU_CONFIG);
  }, [exec]);

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('play', selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleDraw = useCallback(() => {
    exec('draw');
  }, [exec]);

  const handleChooseSuit = useCallback(
    (suit: number) => {
      exec('suit', undefined, suit);
    },
    [exec],
  );

  const handleDeclare = useCallback(() => {
    exec('declare');
  }, [exec]);

  const handleSkipDeclare = useCallback(() => {
    exec('skipdeclare');
  }, [exec]);

  const handleNextRound = useCallback(() => {
    exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    macauConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
    handleChooseSuit,
    handleDeclare,
    handleSkipDeclare,
    handleNextRound,
    retry,
  };
}
