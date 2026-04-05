import { useCallback, useEffect, useState } from 'react';
import { goFishApi } from '../api/gameApi';
import type { GoFishConfig } from '../types/card';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Go Fish game configuration. */
export const DEFAULT_GOFISH_CONFIG: GoFishConfig = {
  cpuDifficulty: 1,
};

/** CPU difficulty level options for Go Fish. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Hook that manages Go Fish game state and player actions. */
export function useGoFishGame() {
  const { config: goFishConfig, handleConfigChange } = useGameConfig<GoFishConfig>(DEFAULT_GOFISH_CONFIG);
  const [selectedTarget, setSelectedTarget] = useState<number | null>(null);
  const [selectedRank, setSelectedRank] = useState<number | null>(null);

  const onSuccess = useCallback(() => {
    setSelectedTarget(null);
    setSelectedRank(null);
  }, []);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(goFishApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, undefined, DEFAULT_GOFISH_CONFIG);
  }, [exec]);

  const handleSelectTarget = useCallback((idx: number) => {
    setSelectedTarget((prev) => (prev === idx ? null : idx));
  }, []);

  const handleSelectRank = useCallback((rank: number) => {
    setSelectedRank((prev) => (prev === rank ? null : rank));
  }, []);

  const handleAsk = useCallback(() => {
    if (selectedTarget === null || selectedRank === null) return;
    exec('ask', selectedTarget, selectedRank);
  }, [exec, selectedTarget, selectedRank]);

  return {
    state,
    loading,
    error,
    exec,
    goFishConfig,
    handleConfigChange,
    selectedTarget,
    selectedRank,
    handleSelectTarget,
    handleSelectRank,
    handleAsk,
    retry,
  };
}
