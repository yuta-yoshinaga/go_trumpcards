import { useCallback, useEffect, useRef } from 'react';
import { memoryApi } from '../api/gameApi';
import type { MemoryConfig } from '../types/card';
import { MemoryPhase } from '../types/phases';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useLocalStorageNumber } from './useLocalStorageToggle';

/** Pair-count options, from a quick board to the full deck (ADR-0035). */
export const PAIR_COUNT_OPTIONS = [6, 10, 15, 20, 26] as const;

/** Full deck: 26 pairs = 52 cards. */
export const FULL_DECK_PAIRS = 26;

/**
 * Pairs a narrow screen starts with.
 *
 * 52 cards cannot fit a 375x667 viewport while every card keeps its 44x44 tap
 * target: 335px of width allows at most 7 columns, so 52 cards need 8 rows = 366px
 * against a 286px board budget. 20 pairs fit in 6 rows with slack. The player can
 * still choose 26 from the settings. See ADR-0035.
 */
export const MOBILE_DEFAULT_PAIRS = 20;

/** Below this width the board cannot show a full deck; see MOBILE_DEFAULT_PAIRS. */
const NARROW_VIEWPORT_PX = 640;

/** Pairs to open with, given the current viewport. */
export function initialPairCount(): number {
  if (typeof window === 'undefined') return FULL_DECK_PAIRS;
  return window.innerWidth < NARROW_VIEWPORT_PX ? MOBILE_DEFAULT_PAIRS : FULL_DECK_PAIRS;
}

/** Default Memory game configuration. */
export const DEFAULT_MEMORY_CONFIG: MemoryConfig = {
  cpuDifficulty: 1,
  pairCount: FULL_DECK_PAIRS,
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

/** Default auto-next delay in milliseconds — must match one of AUTO_NEXT_DELAY_OPTIONS so the settings <select> has a matching option. */
export const DEFAULT_AUTO_NEXT_DELAY_MS = 1000;

const AUTO_NEXT_DELAY_STORAGE_KEY = 'memory_auto_next_delay_ms';

/** Hook that manages Memory game state, card flipping, and configuration. */
export function useMemoryGame() {
  // One source of truth for the opening deal. The settings state and the config sent
  // to the server MUST come from the same value: seeding the config with the default
  // 26 while dealing initialPairCount() left the select reading "26 pairs" over a
  // 40-card board, and the plain Reset button then re-dealt 52 — silently undoing
  // ADR-0035 on the very screen it exists for.
  const initialConfigRef = useRef<MemoryConfig>({ ...DEFAULT_MEMORY_CONFIG, pairCount: initialPairCount() });
  const { config: memoryConfig, handleConfigChange } = useGameConfig<MemoryConfig>(initialConfigRef.current);
  const [autoNextDelayMs, setAutoNextDelayMs] = useLocalStorageNumber(
    AUTO_NEXT_DELAY_STORAGE_KEY,
    DEFAULT_AUTO_NEXT_DELAY_MS,
  );

  // NOTE: rawExec is the game API exec function from useGameApi, not a shell exec.
  const { state, loading, error, exec: rawExec, retry } = useGameApi(memoryApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset', undefined, initialConfigRef.current);
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
