import { useCallback, useEffect } from 'react';
import { boliviaApi } from '../api/gameApi';
import type { BoliviaConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/**
 * Default Bolivia game configuration.
 *
 * **The target is 15000, not Samba's 10000.** This value is sent on reset, so
 * leaving the cloned 10000 here made the Web GUI play a different game from
 * the CUI and from every line of the docs — the domain default
 * (`BoliviaDefaultPointLimit`) is 15000.
 */
export const DEFAULT_BOLIVIA_CONFIG: BoliviaConfig = {
  cpuDifficulty: 1,
  pointLimit: 15000,
};

/** CPU difficulty level options for Bolivia. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Bolivia (Bolivia is played to a higher target than Canasta). */
export const POINT_LIMIT_OPTIONS = [5000, 7500, 10000, 15000] as const;

/** Hook that manages Bolivia game state and player actions. */
export function useBoliviaGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: boliviaConfig, handleConfigChange } = useGameConfig<BoliviaConfig>(DEFAULT_BOLIVIA_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(boliviaApi.exec, { onSuccess });

  const gameExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    gameExec('reset', undefined, DEFAULT_BOLIVIA_CONFIG);
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
    boliviaConfig,
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
