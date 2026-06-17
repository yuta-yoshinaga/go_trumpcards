import { useCallback, useEffect, useState } from 'react';
import { type EscobaConfigInput, escobaApi } from '../api/gameApi';
import type { EscobaResponse } from '../types/card';
import { useGameApi } from './useGameApi';

const defaultConfigInput: EscobaConfigInput = {
  targetScore: 10,
  cpuDifficulty: 1,
};

/** Hook that manages Escoba game state, selections, and action dispatch. */
export function useEscobaGame() {
  const [handIndex, setHandIndex] = useState<number | null>(null);
  const [tableIndices, setTableIndices] = useState<number[]>([]);
  const [configInput, setConfigInput] = useState<EscobaConfigInput>(defaultConfigInput);

  const clearSelection = useCallback(() => {
    setHandIndex(null);
    setTableIndices([]);
  }, []);

  const toggleTable = useCallback((idx: number) => {
    setTableIndices((prev) => (prev.includes(idx) ? prev.filter((x) => x !== idx) : [...prev, idx]));
  }, []);

  const onSuccess = useCallback(
    async (_res: EscobaResponse) => {
      clearSelection();
    },
    [clearSelection],
  );

  const { loading, error, state, exec: callApi, retry } = useGameApi(escobaApi.exec, { onSuccess });

  useEffect(() => {
    callApi('r');
  }, [callApi]);

  const handleConfigChange = useCallback((key: keyof EscobaConfigInput, value: number) => {
    setConfigInput((prev) => ({ ...prev, [key]: value }));
  }, []);

  const play = useCallback(() => {
    if (handIndex === null) return;
    callApi('p', {
      handIndex,
      tableIndices: [...tableIndices].sort((a, b) => a - b),
    });
  }, [callApi, handIndex, tableIndices]);

  const handleNextRound = useCallback(() => {
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
    tableIndices,
    toggleTable,
    clearSelection,
    configInput,
    handleConfigChange,
    play,
    handleNextRound,
    handleResetWithConfig,
  };
}
