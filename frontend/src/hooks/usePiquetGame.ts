import { useCallback, useEffect } from 'react';
import { piquetApi } from '../api/gameApi';
import type { PiquetConfig } from '../types/card';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Piquet game configuration. */
export const DEFAULT_PIQUET_CONFIG: PiquetConfig = {
  cpuDifficulty: 1,
  dealsPerPartie: 6,
};

/** CPU difficulty level options for Piquet. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Deal-count options for Piquet partie length. */
export const DEALS_PER_PARTIE_OPTIONS = [1, 3, 6] as const;

/** Hook that manages Piquet game state and player actions. */
export function usePiquetGame() {
  const { config: piquetConfig, handleConfigChange } = useGameConfig<PiquetConfig>(DEFAULT_PIQUET_CONFIG);

  const { state, loading, error, exec: rawExec, retry } = useGameApi(piquetApi.exec);

  const gameExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    gameExec('reset', undefined, undefined, DEFAULT_PIQUET_CONFIG);
  }, [gameExec]);

  const handleReset = useCallback(() => {
    gameExec('reset', undefined, undefined, piquetConfig);
  }, [gameExec, piquetConfig]);

  const handleExchangeElder = useCallback(
    (discardIndices: number[]) => {
      gameExec('e', undefined, discardIndices);
    },
    [gameExec],
  );

  const handleExchangeYounger = useCallback(
    (discardIndices: number[]) => {
      gameExec('y', undefined, discardIndices);
    },
    [gameExec],
  );

  const handleResolveDeclaration = useCallback(() => {
    gameExec('d');
  }, [gameExec]);

  const handlePlay = useCallback(
    (cardIndex: number) => {
      gameExec('p', cardIndex);
    },
    [gameExec],
  );

  const handleNextDeal = useCallback(() => {
    gameExec('nd');
  }, [gameExec]);

  const handleHint = useCallback(() => {
    gameExec('h');
  }, [gameExec]);

  return {
    state,
    loading,
    error,
    exec: rawExec,
    piquetConfig,
    handleConfigChange,
    handleReset,
    handleExchangeElder,
    handleExchangeYounger,
    handleResolveDeclaration,
    handlePlay,
    handleNextDeal,
    handleHint,
    retry,
  };
}
