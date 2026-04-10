import { useCallback, useEffect, useRef } from 'react';
import { speedApi } from '../api/gameApi';
import type { SpeedConfig } from '../types/card';
import { SpeedPhase } from '../types/phases';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Speed game configuration. */
export const DEFAULT_SPEED_CONFIG: SpeedConfig = {
  cpuDifficulty: 1,
  autoFlip: true,
};

/** Delay in milliseconds before an automatic flip fires while stuck. */
export const AUTO_FLIP_DELAY_MS = 2500;

/** CPU difficulty options for Speed settings. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
] as const;

/** Hook that manages Speed game state and player actions. */
export function useSpeedGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: speedConfig, handleConfigChange, handleToggle } = useGameConfig<SpeedConfig>(DEFAULT_SPEED_CONFIG);

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

  // Auto-flip: when stuck and enabled, schedule a flip after a short delay
  // so players don't have to manually press the button. Cleared on phase
  // change, unmount, or when the toggle flips off.
  const autoFlipEnabled = speedConfig.autoFlip;
  const phase = state?.phase;
  const handleFlipRef = useRef(handleFlip);
  useEffect(() => {
    handleFlipRef.current = handleFlip;
  }, [handleFlip]);
  useEffect(() => {
    if (!autoFlipEnabled) return;
    if (phase !== SpeedPhase.STUCK) return;
    const timerId = setTimeout(() => {
      handleFlipRef.current();
    }, AUTO_FLIP_DELAY_MS);
    return () => clearTimeout(timerId);
  }, [autoFlipEnabled, phase]);

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
    handleToggle,
    handlePlay,
    handleFlip,
    handleHint,
    retry,
  };
}
