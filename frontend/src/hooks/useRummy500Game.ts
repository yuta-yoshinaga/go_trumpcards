import { useCallback, useEffect } from 'react';
import { rummy500Api } from '../api/gameApi';
import type { Rummy500Config } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Rummy 500 game configuration. */
export const DEFAULT_RUMMY500_CONFIG: Rummy500Config = {
  cpuDifficulty: 1,
  pointLimit: 500,
};

/** CPU difficulty level options for Rummy 500. */
export const RUMMY500_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Rummy 500. */
export const RUMMY500_POINT_LIMIT_OPTIONS = [200, 300, 500, 750, 1000] as const;

/** Hook that manages Rummy 500 game state and player actions. */
export function useRummy500Game() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: rummy500Config, handleConfigChange } = useGameConfig<Rummy500Config>(DEFAULT_RUMMY500_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawApi, retry } = useGameApi(rummy500Api.exec, { onSuccess });

  const runApi = useCallback((...args: Parameters<typeof rawApi>) => rawApi(...args), [rawApi]);

  useEffect(() => {
    runApi('reset', undefined, DEFAULT_RUMMY500_CONFIG);
  }, [runApi]);

  const handleDrawStock = useCallback(() => {
    runApi('drawstock');
  }, [runApi]);

  const handleDrawDiscard = useCallback(
    (idx?: number) => {
      runApi('drawdiscard', undefined, undefined, undefined, idx ?? -1);
    },
    [runApi],
  );

  const handleMeld = useCallback(() => {
    if (selectedCardIndices.length < 3) return;
    runApi('meld', undefined, undefined, selectedCardIndices);
  }, [runApi, selectedCardIndices]);

  const handleLayoff = useCallback(
    (meldOwner: number, meldIdx: number) => {
      if (selectedCardIndices.length !== 1) return;
      runApi('layoff', undefined, undefined, undefined, undefined, {
        meldOwner,
        meldIdx,
        cardIndex: selectedCardIndices[0],
      });
    },
    [runApi, selectedCardIndices],
  );

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    runApi('discard', selectedCardIndices[0]);
  }, [runApi, selectedCardIndices]);

  const handleNextRound = useCallback(() => {
    runApi('nextround');
  }, [runApi]);

  return {
    state,
    loading,
    error,
    exec: runApi,
    rummy500Config,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleMeld,
    handleLayoff,
    handleDiscard,
    handleNextRound,
    retry,
  };
}
