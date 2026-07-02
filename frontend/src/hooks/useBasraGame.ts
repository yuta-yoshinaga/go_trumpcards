import { useCallback, useState } from 'react';
import { type BasraConfigInput, basraApi } from '../api/gameApi';
import type { BasraResponse } from '../types/card';
import { useGameApi } from './useGameApi';

/** Default Basra (Bastra) game configuration (CPU difficulty only). */
export const DEFAULT_BASRA_CONFIG: Required<BasraConfigInput> = {
  cpuDifficulty: 1,
};

/** CPU difficulty level options for Basra. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Hook that manages Basra (Bastra) game state, selections, and action dispatch.
 *
 * Basra is a 4-player 52-card fishing/capture game. On the human's turn the
 * player selects one hand card and optionally a set of table cards to capture,
 * then plays. A number card captures same-rank cards and/or table subsets
 * summing to its value; a Jack sweeps the whole table (except other Jacks);
 * playing with no table cards selected trails the card. The command set is built
 * directly on {@link useGameApi}.
 */
export function useBasraGame() {
  const [handIndex, setHandIndex] = useState<number | null>(null);
  const [tableIndices, setTableIndices] = useState<number[]>([]);
  const [configInput, setConfigInput] = useState<Required<BasraConfigInput>>(DEFAULT_BASRA_CONFIG);

  /** Clears the current hand-card and table-card selection. */
  const clearSelection = useCallback(() => {
    setHandIndex(null);
    setTableIndices([]);
  }, []);

  /** Toggles a table card index in/out of the capture selection. */
  const toggleTable = useCallback((idx: number) => {
    setTableIndices((prev) => (prev.includes(idx) ? prev.filter((x) => x !== idx) : [...prev, idx]));
  }, []);

  const onSuccess = useCallback(
    async (_res: BasraResponse) => {
      clearSelection();
    },
    [clearSelection],
  );

  const { loading, error, state, exec: callApi, retry } = useGameApi(basraApi.exec, { onSuccess });

  /** Updates a single config field (CPU difficulty). */
  const handleConfigChange = useCallback((key: keyof BasraConfigInput, value: number) => {
    setConfigInput((prev) => ({ ...prev, [key]: value }));
  }, []);

  /** Plays the selected hand card, capturing the currently selected table cards. */
  const playCard = useCallback(() => {
    if (handIndex === null) return;
    callApi('play', {
      cardIndex: handIndex,
      tableIndices: [...tableIndices].sort((a, b) => a - b),
    });
  }, [callApi, handIndex, tableIndices]);

  /** Starts a fresh game once the previous one has ended. */
  const handleNextGame = useCallback(() => {
    callApi('nextround');
  }, [callApi]);

  /** Resets the game, applying the current config. */
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
    clearSelection,
    configInput,
    handleConfigChange,
    playCard,
    handleNextGame,
    handleResetWithConfig,
  };
}
