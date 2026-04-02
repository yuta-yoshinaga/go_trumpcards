import { useCallback, useEffect } from 'react';
import { canastaApi } from '../api/gameApi';
import type { CanastaConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Canasta game configuration. */
export const DEFAULT_CANASTA_CONFIG: CanastaConfig = {
  cpuDifficulty: 1,
  pointLimit: 5000,
};

/** CPU difficulty level options for Canasta. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Canasta. */
export const POINT_LIMIT_OPTIONS = [3000, 5000, 7500, 10000] as const;

/** Hook that manages Canasta game state and player actions. */
export function useCanastaGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: canastaConfig, handleConfigChange } = useGameConfig<CanastaConfig>(DEFAULT_CANASTA_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec } = useGameApi(canastaApi.exec, { onSuccess });

  const gameExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    gameExec('reset', undefined, DEFAULT_CANASTA_CONFIG);
  }, [gameExec]);

  const handleDrawStock = useCallback(() => {
    gameExec('drawstock');
  }, [gameExec]);

  const handleDrawDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 2) return;
    gameExec('drawdiscard', undefined, undefined, selectedCardIndices);
  }, [gameExec, selectedCardIndices]);

  const handleMeldSelected = useCallback(() => {
    if (selectedCardIndices.length < 3) return;
    gameExec('meld', undefined, undefined, undefined, [selectedCardIndices]);
  }, [gameExec, selectedCardIndices]);

  const handleSkipMeld = useCallback(() => {
    gameExec('skipmeld');
  }, [gameExec]);

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    gameExec('discard', selectedCardIndices[0]);
  }, [gameExec, selectedCardIndices]);

  const handleGoOut = useCallback(() => {
    gameExec('goout');
  }, [gameExec]);

  const handleNextRound = useCallback(() => {
    gameExec('nextround');
  }, [gameExec]);

  const handleReset = useCallback(() => {
    gameExec('reset', undefined, canastaConfig);
  }, [gameExec, canastaConfig]);

  return {
    state,
    loading,
    error,
    gameExec,
    canastaConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleDrawStock,
    handleDrawDiscard,
    handleMeldSelected,
    handleSkipMeld,
    handleDiscard,
    handleGoOut,
    handleNextRound,
    handleReset,
  };
}
