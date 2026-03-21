import { useCallback, useEffect } from 'react';
import { memoryApi } from '../api/gameApi';
import type { MemoryConfig } from '../types/card';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Memory game configuration. */
export const DEFAULT_MEMORY_CONFIG: MemoryConfig = {
  cpuDifficulty: 1,
};

/** CPU difficulty level options for Memory. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Hook that manages Memory game state, card flipping, and configuration. */
export function useMemoryGame() {
  const { config: memoryConfig, handleConfigChange } = useGameConfig<MemoryConfig>(DEFAULT_MEMORY_CONFIG);

  const { state, loading, error, exec: rawExec } = useGameApi(memoryApi.exec);

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, DEFAULT_MEMORY_CONFIG);
  }, [exec]);

  const handleFlip = useCallback(
    (pos: number) => {
      exec('flip', pos);
    },
    [exec],
  );

  const handleNext = useCallback(() => {
    exec('next');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    memoryConfig,
    handleConfigChange,
    handleFlip,
    handleNext,
  };
}
