import { useCallback, useEffect } from 'react';
import { machiavelliApi } from '../api/gameApi';
import type { MachiavelliConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Machiavelli game configuration. */
export const DEFAULT_MACHIAVELLI_CONFIG: MachiavelliConfig = {
  playerCount: 4,
  cpuDifficulty: 1,
  targetRounds: 3,
};

/** CPU difficulty level options for Machiavelli. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available player-count options for Machiavelli (2-5 players). */
export const PLAYER_COUNT_OPTIONS = [2, 3, 4, 5] as const;

/** Available target-round options for Machiavelli. */
export const TARGET_ROUNDS_OPTIONS = [1, 3, 5, 10] as const;

/** Hook that manages Machiavelli game state and player actions. */
export function useMachiavelliGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: machiavelliConfig, handleConfigChange } =
    useGameConfig<MachiavelliConfig>(DEFAULT_MACHIAVELLI_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(machiavelliApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, DEFAULT_MACHIAVELLI_CONFIG);
  }, [exec]);

  const handleDraw = useCallback(() => {
    exec('draw');
  }, [exec]);

  const handleNewMeld = useCallback(() => {
    if (selectedCardIndices.length < 3) return;
    exec('newmeld', { handIndices: [...selectedCardIndices] });
  }, [exec, selectedCardIndices]);

  const handleLayoff = useCallback(
    (meldIdx: number) => {
      if (selectedCardIndices.length !== 1) return;
      exec('layoff', { meldIdx, handIndex: selectedCardIndices[0] });
    },
    [exec, selectedCardIndices],
  );

  const handleRearrange = useCallback(
    (tableMelds: { design: number; value: number }[][], handIndices: number[]) => {
      if (handIndices.length < 1 || tableMelds.length < 1) return;
      exec('play', { tableMelds, handIndices });
    },
    [exec],
  );

  const handleNextRound = useCallback(() => {
    exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    machiavelliConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDraw,
    handleNewMeld,
    handleLayoff,
    handleRearrange,
    handleNextRound,
    retry,
  };
}
