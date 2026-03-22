import { useCallback, useEffect, useState } from 'react';
import { heartsApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { HeartsConfig, HeartsHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Hearts game configuration. */
export const DEFAULT_HEARTS_CONFIG: HeartsConfig = {
  cpuDifficulty: 1,
  pointLimit: 100,
  omnibusJD: false,
};

/** CPU difficulty level options for Hearts. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Hearts. */
export const POINT_LIMIT_OPTIONS = [50, 100, 150, 200] as const;

/** Hook that manages Hearts game state and player actions. */
export function useHeartsGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: heartsConfig, handleConfigChange, handleToggle } = useGameConfig<HeartsConfig>(DEFAULT_HEARTS_CONFIG);
  const [hint, setHint] = useState<HeartsHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec } = useGameApi(heartsApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, undefined, DEFAULT_HEARTS_CONFIG);
  }, [exec]);

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

  const handleHint = useCallback(async () => {
    try {
      const res = await heartsApi.exec('hint');
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, []);

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    exec,
    heartsConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleToggle,
    handlePass,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
  };
}
