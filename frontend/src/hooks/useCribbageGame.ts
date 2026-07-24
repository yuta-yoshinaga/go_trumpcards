import { useCallback, useEffect } from 'react';
import { cribbageApi } from '../api/gameApi';
import type { CribbageConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Cribbage game configuration. */
export const DEFAULT_CRIBBAGE_CONFIG: CribbageConfig = {
  cpuDifficulty: 1,
  pointLimit: 121,
};

/** CPU difficulty level options for Cribbage. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Cribbage. */
export const POINT_LIMIT_OPTIONS = [61, 121] as const;

/** Hook that manages Cribbage game state and player actions. */
export function useCribbageGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: cribbageConfig, handleConfigChange } = useGameConfig<CribbageConfig>(DEFAULT_CRIBBAGE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(cribbageApi.exec, { onSuccess });

  const gameExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    gameExec('reset', undefined, undefined, DEFAULT_CRIBBAGE_CONFIG);
  }, [gameExec]);

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 2) return;
    gameExec('discard', undefined, selectedCardIndices);
  }, [gameExec, selectedCardIndices]);

  const handleCut = useCallback(() => {
    gameExec('cut');
  }, [gameExec]);

  const handlePeg = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    gameExec('peg', selectedCardIndices[0]);
  }, [gameExec, selectedCardIndices]);

  const handleGo = useCallback(() => {
    gameExec('go');
  }, [gameExec]);

  const handleShowNext = useCallback(() => {
    gameExec('shownext');
  }, [gameExec]);

  const handleNextRound = useCallback(() => {
    gameExec('nextround');
  }, [gameExec]);

  return {
    state,
    loading,
    error,
    exec: gameExec,
    cribbageConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDiscard,
    handleCut,
    handlePeg,
    handleGo,
    handleShowNext,
    handleNextRound,
    retry,
  };
}
