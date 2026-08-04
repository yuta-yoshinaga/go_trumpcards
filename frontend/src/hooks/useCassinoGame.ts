import { useCallback, useEffect, useState } from 'react';
import { type CassinoConfigInput, cassinoApi } from '../api/gameApi';
import type { CassinoResponse } from '../types/card';
import { useGameApi } from './useGameApi';

const defaultConfigInput: CassinoConfigInput = {
  targetScore: 21,
  multiBuildEnabled: true,
  sweepBonusEnabled: true,
  cpuDifficulty: 1,
};

/** Hook that manages Cassino game state, selections, and action dispatch. */
export function useCassinoGame() {
  const [handIndex, setHandIndex] = useState<number | null>(null);
  const [tableIndices, setTableIndices] = useState<number[]>([]);
  const [buildIndices, setBuildIndices] = useState<number[]>([]);
  const [declaredValue, setDeclaredValue] = useState<number>(8);
  const [configInput, setConfigInput] = useState<CassinoConfigInput>(defaultConfigInput);

  const clearSelection = useCallback(() => {
    setHandIndex(null);
    setTableIndices([]);
    setBuildIndices([]);
  }, []);

  const toggleTable = useCallback((idx: number) => {
    setTableIndices((prev) => (prev.includes(idx) ? prev.filter((x) => x !== idx) : [...prev, idx]));
  }, []);

  const toggleBuild = useCallback((idx: number) => {
    setBuildIndices((prev) => (prev.includes(idx) ? prev.filter((x) => x !== idx) : [...prev, idx]));
  }, []);

  const onSuccess = useCallback(
    async (_res: CassinoResponse) => {
      clearSelection();
    },
    [clearSelection],
  );

  const { loading, error, state, exec: callApi, retry } = useGameApi(cassinoApi.exec, { onSuccess });

  useEffect(() => {
    callApi('reset');
  }, [callApi]);

  const handleConfigChange = useCallback((key: keyof CassinoConfigInput, value: boolean | number) => {
    setConfigInput((prev) => ({ ...prev, [key]: value }));
  }, []);

  const playTake = useCallback(() => {
    if (handIndex === null) return;
    callApi('take', {
      handIndex,
      tableIndices: [...tableIndices].sort((a, b) => a - b),
      buildIndices: [...buildIndices].sort((a, b) => a - b),
    });
  }, [callApi, handIndex, tableIndices, buildIndices]);

  const playBuild = useCallback(() => {
    if (handIndex === null) return;
    callApi('build', {
      handIndex,
      tableIndices: [...tableIndices].sort((a, b) => a - b),
      declaredValue,
    });
  }, [callApi, handIndex, tableIndices, declaredValue]);

  const playTrail = useCallback(() => {
    if (handIndex === null) return;
    callApi('trail', { handIndex });
  }, [callApi, handIndex]);

  const handleResetWithConfig = useCallback(() => {
    callApi('reset', { config: configInput });
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
    buildIndices,
    toggleBuild,
    declaredValue,
    setDeclaredValue,
    clearSelection,
    configInput,
    handleConfigChange,
    playTake,
    playBuild,
    playTrail,
    handleResetWithConfig,
  };
}
