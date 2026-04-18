import { useCallback, useEffect, useRef } from 'react';
import { memoryApi } from '../api/gameApi';
import type { MemoryConfig } from '../types/card';
import { MemoryPhase } from '../types/phases';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useLocalStorageNumber } from './useLocalStorageToggle';

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

/** Auto-next delay options for Memory result phase (0 = manual advance). */
export const AUTO_NEXT_DELAY_OPTIONS = [
  { value: 0, label: 'manual' },
  { value: 1000, label: 'fast' },
  { value: 2000, label: 'slow' },
] as const;

/** Default auto-next delay in milliseconds. */
export const DEFAULT_AUTO_NEXT_DELAY_MS = 1500;

const AUTO_NEXT_DELAY_STORAGE_KEY = 'memory_auto_next_delay_ms';

/** Hook that manages Memory game state, card flipping, and configuration. */
export function useMemoryGame() {
  const { config: memoryConfig, handleConfigChange } = useGameConfig<MemoryConfig>(DEFAULT_MEMORY_CONFIG);
  const [autoNextDelayMs, setAutoNextDelayMs] = useLocalStorageNumber(
    AUTO_NEXT_DELAY_STORAGE_KEY,
    DEFAULT_AUTO_NEXT_DELAY_MS,
  );

  // NOTE: rawExec is the game API exec function from useGameApi, not a shell exec.
  const { state, loading, error, exec: rawExec, retry } = useGameApi(memoryApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset', undefined, DEFAULT_MEMORY_CONFIG);
  }, [runApi]);

  const handleFlip = useCallback(
    (pos: number) => {
      runApi('flip', pos);
    },
    [runApi],
  );

  const handleNext = useCallback(() => {
    runApi('next');
  }, [runApi]);

  // Auto-advance after the result phase so players don't need to tap "Next" every turn.
  // The delay is user-configurable; 0 disables auto-advance.
  const handleNextRef = useRef(handleNext);
  useEffect(() => {
    handleNextRef.current = handleNext;
  }, [handleNext]);
  const isResultPhase = state?.phase === MemoryPhase.RESULT;
  const isGameEnd = state?.gameEndFlag ?? false;
  useEffect(() => {
    if (!isResultPhase || isGameEnd || loading || autoNextDelayMs <= 0) return;
    const timerId = setTimeout(() => {
      handleNextRef.current();
    }, autoNextDelayMs);
    return () => clearTimeout(timerId);
  }, [isResultPhase, isGameEnd, loading, autoNextDelayMs]);

  return {
    state,
    loading,
    error,
    exec: runApi,
    memoryConfig,
    handleConfigChange,
    autoNextDelayMs,
    setAutoNextDelayMs,
    handleFlip,
    handleNext,
    retry,
  };
}
