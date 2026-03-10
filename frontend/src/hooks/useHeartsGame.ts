import { useCallback, useEffect, useState } from 'react';
import { heartsApi } from '../api/gameApi';
import type { HeartsConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';

export const DEFAULT_HEARTS_CONFIG: HeartsConfig = {
  cpuDifficulty: 1,
  pointLimit: 100,
};

export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

export const POINT_LIMIT_OPTIONS = [50, 100, 150, 200] as const;

export function useHeartsGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const [heartsConfig, setHeartsConfig] = useState<HeartsConfig>(DEFAULT_HEARTS_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec } = useGameApi(heartsApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, undefined, DEFAULT_HEARTS_CONFIG);
  }, [exec]);

  const handleConfigChange = useCallback((key: keyof HeartsConfig, value: string) => {
    const parsed = Number(value);
    if (!Number.isNaN(parsed)) {
      setHeartsConfig((prev) => ({ ...prev, [key]: parsed }));
    }
  }, []);

  const handlePass = useCallback(() => {
    exec('pass', selectedCardIndices);
  }, [exec, selectedCardIndices]);

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
    heartsConfig,
    selectedCardIndices,
    toggleCard,
    handleConfigChange,
    handlePass,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
