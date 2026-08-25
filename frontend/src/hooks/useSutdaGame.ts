import { useCallback } from 'react';
import { type SutdaConfigInput, sutdaApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Sutda configuration. */
export const DEFAULT_SUTDA_CONFIG: Required<SutdaConfigInput> = {
  cpuDifficulty: 1,
  seats: 3,
  startChips: 1000,
};

/** CPU difficulty options. */
export const SUTDA_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Seat-count options. */
export const SUTDA_SEAT_OPTIONS = [2, 3, 4, 5] as const;

/** Starting-chip options. */
export const SUTDA_CHIP_OPTIONS = [500, 1000, 2000, 5000] as const;

/**
 * Hook that manages Sutda state and its player actions: calling, raising,
 * folding, and moving to the next hand.
 */
export function useSutdaGame() {
  const { config, handleConfigChange } = useGameConfig<Required<SutdaConfigInput>>(DEFAULT_SUTDA_CONFIG);
  const { state, loading, error, exec, retry } = useGameApi(sutdaApi.exec);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Calls the current bet (a check when nothing is owed). */
  const handleCall = useCallback(() => {
    void exec('call');
  }, [exec]);

  /** Raises by one unit. */
  const handleRaise = useCallback(() => {
    void exec('raise');
  }, [exec]);

  /** Folds out of the hand. */
  const handleFold = useCallback(() => {
    void exec('fold');
  }, [exec]);

  /** Advances to the next hand. */
  const handleNextHand = useCallback(() => {
    void exec('nexthand');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    sutdaConfig: config,
    handleConfigChange,
    reset,
    handleCall,
    handleRaise,
    handleFold,
    handleNextHand,
  };
}
