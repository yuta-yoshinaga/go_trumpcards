import { useCallback, useEffect } from 'react';
import { sevenBridgeApi } from '../api/gameApi';
import type { SevenBridgeConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Seven Bridge game configuration. */
export const DEFAULT_SEVENBRIDGE_CONFIG: SevenBridgeConfig = {
  cpuDifficulty: 1,
  pointLimit: 100,
};

/** CPU difficulty level options for Seven Bridge. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Seven Bridge. */
export const POINT_LIMIT_OPTIONS = [50, 100, 150, 200] as const;

/** Hook that manages Seven Bridge game state and player actions. */
export function useSevenBridgeGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: sevenBridgeConfig, handleConfigChange } =
    useGameConfig<SevenBridgeConfig>(DEFAULT_SEVENBRIDGE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawApi, retry } = useGameApi(sevenBridgeApi.exec, { onSuccess });

  const callApi = useCallback((...args: Parameters<typeof rawApi>) => rawApi(...args), [rawApi]);

  useEffect(() => {
    callApi('reset', undefined, DEFAULT_SEVENBRIDGE_CONFIG);
  }, [callApi]);

  const handleDrawStock = useCallback(() => {
    callApi('drawstock');
  }, [callApi]);

  const handlePon = useCallback(() => {
    if (selectedCardIndices.length !== 2) return;
    callApi('pon', undefined, undefined, selectedCardIndices);
  }, [callApi, selectedCardIndices]);

  const handleChi = useCallback(() => {
    if (selectedCardIndices.length !== 2) return;
    callApi('chi', undefined, undefined, selectedCardIndices);
  }, [callApi, selectedCardIndices]);

  const handleMeld = useCallback(() => {
    if (selectedCardIndices.length < 3) return;
    callApi('meld', undefined, undefined, selectedCardIndices);
  }, [callApi, selectedCardIndices]);

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    callApi('discard', selectedCardIndices[0]);
  }, [callApi, selectedCardIndices]);

  const handleLayoff = useCallback(
    (targetPlayerIdx: number, meldIdx: number) => {
      if (selectedCardIndices.length !== 1) return;
      callApi('layoff', selectedCardIndices[0], undefined, undefined, targetPlayerIdx, meldIdx);
    },
    [callApi, selectedCardIndices],
  );

  const handleNextRound = useCallback(() => {
    callApi('nextround');
  }, [callApi]);

  return {
    state,
    loading,
    error,
    callApi,
    sevenBridgeConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handlePon,
    handleChi,
    handleMeld,
    handleDiscard,
    handleLayoff,
    handleNextRound,
    retry,
  };
}
