import { useCallback, useEffect, useState } from 'react';
import { durakApi } from '../api/gameApi';
import type { DurakConfig, DurakHint } from '../types/card';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useHintRequest } from './useHintRequest';

/** Default Durak game configuration. */
export const DEFAULT_DURAK_CONFIG: DurakConfig = {
  playerCount: 4,
  cpuDifficulty: 0,
  transferEnabled: false,
};

/** CPU difficulty options for Durak settings (0=Normal, 1=Easy, 2=Hard matches backend). */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'normal' },
  { value: 1, label: 'easy' },
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

  // **他のトリック系はサーバー計算の理由付きヒントを持つのに、Durak は
  // クライアント完結の簡易ヒューリスティックだけだった (#4740)。**
  const [hint, setHint] = useState<DurakHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const handleHint = useHintRequest({
    fetchHint: () => durakApi.exec('hint'),
    selectHint: (res) => res.hint ?? null,
    setHint,
    setHintError,
  });

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

  const handleTransfer = useCallback(() => {
    if (selectedCardIdx === null) return;
    gameExec('transfer', selectedCardIdx);
  }, [gameExec, selectedCardIdx]);

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
    handleTransfer,
    handleSort,
    hint,
    hintError,
    handleHint,
  };
}
