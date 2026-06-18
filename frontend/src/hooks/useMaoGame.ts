import { useCallback, useEffect } from 'react';
import { maoApi } from '../api/gameApi';
import type { MaoConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Mao game configuration. */
export const DEFAULT_MAO_CONFIG: MaoConfig = {
  cpuDifficulty: 1,
  pointLimit: 200,
};

/** CPU difficulty level options for Mao. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Mao. */
export const POINT_LIMIT_OPTIONS = [100, 200, 300, 500] as const;

/** Hook that manages Mao game state and player actions, including the hidden-rule "say word" action. */
export function useMaoGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: maoConfig, handleConfigChange } = useGameConfig<MaoConfig>(DEFAULT_MAO_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(maoApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, undefined, DEFAULT_MAO_CONFIG);
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

  const handleDeclareWord = useCallback(
    (word: string) => {
      const trimmed = word.trim();
      if (trimmed.length === 0) return;
      exec('declareword', undefined, undefined, undefined, trimmed);
    },
    [exec],
  );

  return {
    state,
    loading,
    error,
    exec,
    maoConfig,
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
    handleDeclareWord,
    retry,
  };
}
