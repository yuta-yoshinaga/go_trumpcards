import { useCallback, useEffect } from 'react';
import { pageoneApi } from '../api/gameApi';
import type { PageOneConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Page One game configuration. */
export const DEFAULT_PAGEONE_CONFIG: PageOneConfig = {
  cpuDifficulty: 1,
  pointLimit: 200,
};

/** CPU difficulty level options for Page One. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Page One. */
export const POINT_LIMIT_OPTIONS = [100, 200, 300, 500, 1000] as const;

/** Hook that manages Page One game state and player actions. */
export function usePageOneGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: pageOneConfig, handleConfigChange } = useGameConfig<PageOneConfig>(DEFAULT_PAGEONE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const apiResult = useGameApi(pageoneApi.exec, { onSuccess });
  const { state, loading, error, retry } = apiResult;
  const rawCall = apiResult.exec;

  const call = useCallback((...args: Parameters<typeof rawCall>) => rawCall(...args), [rawCall]);

  useEffect(() => {
    call('reset', undefined, DEFAULT_PAGEONE_CONFIG);
  }, [call]);

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    call('play', selectedCardIndices[0]);
  }, [call, selectedCardIndices]);

  const handleDraw = useCallback(() => {
    call('draw');
  }, [call]);

  const handleDeclare = useCallback(() => {
    call('declare');
  }, [call]);

  const handleSkipDeclare = useCallback(() => {
    call('skip');
  }, [call]);

  const handleNextRound = useCallback(() => {
    call('nextround');
  }, [call]);

  return {
    state,
    loading,
    error,
    exec: call,
    pageOneConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
    handleDeclare,
    handleSkipDeclare,
    handleNextRound,
    retry,
  };
}
