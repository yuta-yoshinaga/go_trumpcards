import { useCallback, useEffect, useState } from 'react';
import { type BarbuConfigInput, barbuApi } from '../api/gameApi';
import type { BarbuResponse } from '../types/card';
import { useGameApi } from './useGameApi';

const defaultConfigInput: BarbuConfigInput = {
  cpuDifficulty: 1,
};

/** Hook that manages Barbu game state, selections, and action dispatch. */
export function useBarbuGame() {
  const [handIndex, setHandIndex] = useState<number | null>(null);
  const [configInput, setConfigInput] = useState<BarbuConfigInput>(defaultConfigInput);

  const clearSelection = useCallback(() => {
    setHandIndex(null);
  }, []);

  const onSuccess = useCallback(
    async (_res: BarbuResponse) => {
      clearSelection();
    },
    [clearSelection],
  );

  const { loading, error, state, exec: callApi, retry } = useGameApi(barbuApi.exec, { onSuccess });

  useEffect(() => {
    callApi('r');
  }, [callApi]);

  const handleConfigChange = useCallback((key: keyof BarbuConfigInput, value: number) => {
    setConfigInput((prev) => ({ ...prev, [key]: value }));
  }, []);

  /** Dealer picks a contract (trumpSuit only used for the Trumps contract). */
  const selectContract = useCallback(
    (contract: number, trumpSuit = -1) => {
      callApi('c', { contract, trumpSuit });
    },
    [callApi],
  );

  /** Play the currently selected hand card. */
  const play = useCallback(() => {
    if (handIndex === null) return;
    callApi('p', { handIndex });
  }, [callApi, handIndex]);

  /** Pass (only legal in Dominoes when no card is playable). */
  const pass = useCallback(() => {
    callApi('p', { handIndex: -1 });
  }, [callApi]);

  const handleNextDeal = useCallback(() => {
    callApi('n');
  }, [callApi]);

  const handleResetWithConfig = useCallback(() => {
    callApi('r', { config: configInput });
  }, [callApi, configInput]);

  return {
    state,
    loading,
    error,
    retry,
    callApi,
    handIndex,
    setHandIndex,
    clearSelection,
    configInput,
    handleConfigChange,
    selectContract,
    play,
    pass,
    handleNextDeal,
    handleResetWithConfig,
  };
}
