import { useCallback, useEffect } from 'react';
import { sambaApi } from '../api/gameApi';
import type { SambaConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Samba game configuration. */
export const DEFAULT_SAMBA_CONFIG: SambaConfig = {
  cpuDifficulty: 1,
  pointLimit: 10000,
};

/** CPU difficulty level options for Samba. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Samba (Samba is played to a higher target than Canasta). */
export const POINT_LIMIT_OPTIONS = [5000, 7500, 10000, 15000] as const;

/** Hook that manages Samba game state and player actions. */
export function useSambaGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: sambaConfig, handleConfigChange } = useGameConfig<SambaConfig>(DEFAULT_SAMBA_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(sambaApi.exec, { onSuccess });

  const gameExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    gameExec('reset', undefined, DEFAULT_SAMBA_CONFIG);
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

  return {
    state,
    loading,
    error,
    gameExec,
    sambaConfig,
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
    retry,
  };
}
