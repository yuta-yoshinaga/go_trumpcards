import { useCallback, useEffect } from 'react';
import { tonkApi } from '../api/gameApi';
import type { TonkConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Tonk game configuration. */
export const DEFAULT_TONK_CONFIG: TonkConfig = {
  cpuDifficulty: 1,
  pointLimit: 50,
};

/** CPU difficulty level options for Tonk. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Tonk. */
export const POINT_LIMIT_OPTIONS = [25, 50, 100, 150] as const;

/** Hook that manages Tonk game state and player actions. */
export function useTonkGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: tonkConfig, handleConfigChange } = useGameConfig<TonkConfig>(DEFAULT_TONK_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const apiCall = tonkApi.exec;
  const { state, loading, error, exec: rawExec, retry } = useGameApi(apiCall, { onSuccess });

  const sendCommand = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    sendCommand('reset', undefined, DEFAULT_TONK_CONFIG);
  }, [sendCommand]);

  const handleDrawStock = useCallback(() => {
    sendCommand('drawstock');
  }, [sendCommand]);

  const handleDrawDiscard = useCallback(() => {
    sendCommand('drawdiscard');
  }, [sendCommand]);

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    sendCommand('discard', selectedCardIndices[0]);
  }, [sendCommand, selectedCardIndices]);

  const handleKnock = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    sendCommand('knock', selectedCardIndices[0]);
  }, [sendCommand, selectedCardIndices]);

  const handleNextRound = useCallback(() => {
    sendCommand('nextround');
  }, [sendCommand]);

  return {
    state,
    loading,
    error,
    exec: sendCommand,
    tonkConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleKnock,
    handleNextRound,
    retry,
  };
}
