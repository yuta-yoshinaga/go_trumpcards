import { useCallback, useEffect } from 'react';
import { speedApi } from '../api/gameApi';
import type { SpeedConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Speed game configuration. */
export const DEFAULT_SPEED_CONFIG: SpeedConfig = {
  cpuDifficulty: 1,
};

/** CPU difficulty options for Speed settings. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
] as const;

/** Hook that manages Speed game state and player actions. */
export function useSpeedGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: speedConfig, handleConfigChange } = useGameConfig<SpeedConfig>(DEFAULT_SPEED_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  // NOTE: exec here is the game API exec function from useGameApi, not child_process.exec
  const { state, loading, error, exec: gameExec, retry } = useGameApi(speedApi.exec, { onSuccess });

  useEffect(() => {
    gameExec('reset', undefined, undefined, DEFAULT_SPEED_CONFIG);
  }, [gameExec]);

  const handlePlay = useCallback(
    (pileIndex: number) => {
      if (selectedCardIndices.length !== 1) return;
      gameExec('play', selectedCardIndices[0], pileIndex);
    },
    [gameExec, selectedCardIndices],
  );

  const handleFlip = useCallback(() => {
    gameExec('flip');
  }, [gameExec]);

  const handleHint = useCallback(() => {
    gameExec('hint');
  }, [gameExec]);

  return {
    state,
    loading,
    error,
    exec: gameExec,
    speedConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleFlip,
    handleHint,
    retry,
  };
}
