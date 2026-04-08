import { useCallback, useEffect, useState } from 'react';
import { durakApi } from '../api/gameApi';
import type { DurakConfig } from '../types/card';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Durak game configuration. */
export const DEFAULT_DURAK_CONFIG: DurakConfig = {
  playerCount: 4,
  cpuDifficulty: 1,
  transferEnabled: false,
};

/** CPU difficulty options for Durak settings. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
] as const;

/** Hook that manages Durak game state, card selection, and player actions. */
export function useDurakGame() {
  const [selectedCardIdx, setSelectedCardIdx] = useState<number | null>(null);
  const [selectedAttackIdx, setSelectedAttackIdx] = useState<number | null>(null);
  const {
    config: durakConfig,
    handleConfigChange,
    handleToggle: handleConfigToggle,
  } = useGameConfig<DurakConfig>(DEFAULT_DURAK_CONFIG);

  const onSuccess = useCallback(() => {
    setSelectedCardIdx(null);
    setSelectedAttackIdx(null);
  }, []);

  const { state, loading, error, exec: gameExec, retry } = useGameApi(durakApi.exec, { onSuccess });

  useEffect(() => {
    gameExec('reset', undefined, undefined, DEFAULT_DURAK_CONFIG);
  }, [gameExec]);

  const handleAttack = useCallback(() => {
    if (selectedCardIdx === null) return;
    gameExec('attack', selectedCardIdx);
  }, [gameExec, selectedCardIdx]);

  const handleDefend = useCallback(() => {
    if (selectedCardIdx === null || selectedAttackIdx === null) return;
    gameExec('defend', selectedCardIdx, selectedAttackIdx);
  }, [gameExec, selectedCardIdx, selectedAttackIdx]);

  const handlePass = useCallback(() => {
    gameExec('pass');
  }, [gameExec]);

  const handleTake = useCallback(() => {
    gameExec('take');
  }, [gameExec]);

  const handleSort = useCallback(
    (mode: number) => {
      gameExec('sort', undefined, undefined, undefined, mode);
    },
    [gameExec],
  );

  return {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    durakConfig,
    selectedCardIdx,
    setSelectedCardIdx,
    selectedAttackIdx,
    setSelectedAttackIdx,
    handleConfigChange,
    handleConfigToggle,
    handleAttack,
    handleDefend,
    handlePass,
    handleTake,
    handleSort,
  };
}
