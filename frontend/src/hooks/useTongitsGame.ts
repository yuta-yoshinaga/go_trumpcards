import { useCallback, useEffect } from 'react';
import { tongitsApi } from '../api/gameApi';
import type { TongitsConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Tongits game configuration. */
export const DEFAULT_TONGITS_CONFIG: TongitsConfig = {
  cpuDifficulty: 1,
  pointLimit: 50,
};

/** CPU difficulty level options for Tongits. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Tongits. */
export const POINT_LIMIT_OPTIONS = [25, 50, 100, 150] as const;

/** Hook that manages Tongits game state and player actions. */
export function useTongitsGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: tongitsConfig, handleConfigChange } = useGameConfig<TongitsConfig>(DEFAULT_TONGITS_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const apiCall = tongitsApi.exec;
  const { state, loading, error, exec: rawExec, retry } = useGameApi(apiCall, { onSuccess });

  const sendCommand = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    sendCommand('reset', undefined, DEFAULT_TONGITS_CONFIG);
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

  const handleMeld = useCallback(() => {
    if (selectedCardIndices.length < 3) return;
    sendCommand('meld', undefined, undefined, selectedCardIndices);
  }, [sendCommand, selectedCardIndices]);

  const handleSapaw = useCallback(
    (targetPlayerIdx: number, meldIdx: number) => {
      if (selectedCardIndices.length !== 1) return;
      sendCommand('sapaw', selectedCardIndices[0], undefined, undefined, targetPlayerIdx, meldIdx);
    },
    [sendCommand, selectedCardIndices],
  );

  const handleChallenge = useCallback(() => {
    sendCommand('challenge', undefined, undefined, undefined, undefined, undefined, [true, true]);
  }, [sendCommand]);

  const handleNextRound = useCallback(() => {
    sendCommand('nextround');
  }, [sendCommand]);

  return {
    state,
    loading,
    error,
    exec: sendCommand,
    tongitsConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleMeld,
    handleSapaw,
    handleChallenge,
    handleNextRound,
    retry,
  };
}
