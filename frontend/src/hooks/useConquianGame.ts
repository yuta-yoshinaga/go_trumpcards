import { useCallback, useEffect } from 'react';
import { conquianApi } from '../api/gameApi';
import type { ConquianConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Conquian game configuration. */
export const DEFAULT_CONQUIAN_CONFIG: ConquianConfig = {
  cpuDifficulty: 1,
  targetWins: 3,
};

/** CPU difficulty level options for Conquian. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-wins (match length) options for Conquian. */
export const TARGET_WINS_OPTIONS = [1, 3, 5, 7] as const;

/** Hook that manages Conquian game state and player actions. */
export function useConquianGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: conquianConfig, handleConfigChange } = useGameConfig<ConquianConfig>(DEFAULT_CONQUIAN_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(conquianApi.exec, { onSuccess });

  const gameExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    gameExec('reset', undefined, DEFAULT_CONQUIAN_CONFIG);
  }, [gameExec]);

  const handleDrawStock = useCallback(() => {
    gameExec('drawstock');
  }, [gameExec]);

  const handleDrawDiscard = useCallback(() => {
    gameExec('drawdiscard');
  }, [gameExec]);

  const handleMeldSelected = useCallback(
    (layoffTarget?: number) => {
      // 1 card extends an existing table meld; 3+ cards lay a new set/run.
      if (selectedCardIndices.length !== 1 && selectedCardIndices.length < 3) return;
      // 延長先はプレイヤーが選べる。♠5 は「5 のセット」も「♠4-6-7 のラン」も
      // 延長できるので、先頭一致で決め打たれると意図した側を選べない (#4837)。
      const targets = layoffTarget === undefined ? undefined : [layoffTarget];
      gameExec('meld', undefined, undefined, [selectedCardIndices], targets);
    },
    [gameExec, selectedCardIndices],
  );

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    gameExec('discard', selectedCardIndices[0]);
  }, [gameExec, selectedCardIndices]);

  const handleNextRound = useCallback(() => {
    gameExec('nextround');
  }, [gameExec]);

  return {
    state,
    loading,
    error,
    gameExec,
    conquianConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleDrawStock,
    handleDrawDiscard,
    handleMeldSelected,
    handleDiscard,
    handleNextRound,
    retry,
  };
}
