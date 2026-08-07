import { useCallback, useEffect, useState } from 'react';
import { crazyeightsApi } from '../api/gameApi';
import type { CrazyEightsConfig, CrazyEightsHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useHintRequest } from './useHintRequest';

/** Default Crazy Eights game configuration. */
export const DEFAULT_CRAZYEIGHTS_CONFIG: CrazyEightsConfig = {
  cpuDifficulty: 1,
  pointLimit: 200,
};

/** CPU difficulty level options for Crazy Eights. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Crazy Eights. */
export const POINT_LIMIT_OPTIONS = [100, 200, 300, 500] as const;

/** Hook that manages Crazy Eights game state and player actions. */
export function useCrazyEightsGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: crazyEightsConfig, handleConfigChange } =
    useGameConfig<CrazyEightsConfig>(DEFAULT_CRAZYEIGHTS_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(crazyeightsApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  // **Hearts / Spades はサーバー計算の理由付きヒントを返すのに、CrazyEights は
  // フロント完結の簡易ヒューリスティックしか無かった (#4737)。**
  const [hint, setHint] = useState<CrazyEightsHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const handleHint = useHintRequest({
    fetchHint: () => crazyeightsApi.exec('hint'),
    selectHint: (res) => res.hint ?? null,
    setHint,
    setHintError,
  });

  useEffect(() => {
    exec('reset', undefined, undefined, DEFAULT_CRAZYEIGHTS_CONFIG);
  }, [exec]);

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('play', selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleDraw = useCallback(() => {
    exec('draw');
  }, [exec]);

  const handleChooseSuit = useCallback(
    (suit: number) => {
      exec('suit', undefined, suit);
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
    crazyEightsConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
    handleChooseSuit,
    handleNextRound,
    hint,
    hintError,
    handleHint,
    retry,
  };
}
