import { useCallback, useEffect, useState } from 'react';
import { omiApi } from '../api/gameApi';
import type { OmiConfig, OmiHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useHintRequest } from './useHintRequest';

/** Default Omi game configuration. */
export const DEFAULT_OMI_CONFIG: OmiConfig = {
  cpuDifficulty: 1,
  pointLimit: 10,
};

/** CPU difficulty level options for Omi. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Omi. */
export const POINT_LIMIT_OPTIONS = [5, 7, 10, 15, 21] as const;

/** Hook that manages Omi game state, trump calling, and trick play. */
export function useOmiGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: omiConfig, handleConfigChange } = useGameConfig<OmiConfig>(DEFAULT_OMI_CONFIG);
  const [hint, setHint] = useState<OmiHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(omiApi.exec, { onSuccess });

  const apiExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    apiExec('reset', undefined, undefined, undefined, DEFAULT_OMI_CONFIG);
  }, [apiExec]);

  const handleCallTrump = useCallback(
    (suit: number) => {
      apiExec('calltrump', undefined, suit);
    },
    [apiExec],
  );

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    apiExec('play', selectedCardIndices[0]);
  }, [apiExec, selectedCardIndices]);

  const handleNextTrick = useCallback(() => {
    apiExec('next');
  }, [apiExec]);

  const handleNextRound = useCallback(() => {
    apiExec('nextround');
  }, [apiExec]);

  const handleHint = useHintRequest({
    fetchHint: () => omiApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
    setHintLoading,
  });

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
    apiExec,
    omiConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleCallTrump,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
