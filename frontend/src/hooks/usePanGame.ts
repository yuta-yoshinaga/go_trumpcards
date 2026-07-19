import { useCallback, useEffect } from 'react';
import { panApi } from '../api/gameApi';
import type { PanConfig } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Panguingue (Pan) game configuration. */
export const DEFAULT_PAN_CONFIG: PanConfig = {
  playerCount: 4,
  cpuDifficulty: 1,
  targetRounds: 3,
};

/** CPU difficulty level options for Panguingue. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available player-count options for Panguingue (3-6 players). */
export const PLAYER_COUNT_OPTIONS = [3, 4, 5, 6] as const;

/** Available target-round options for Panguingue. */
export const TARGET_ROUNDS_OPTIONS = [1, 3, 5, 10] as const;

/** Hook that manages Panguingue (Pan) game state and player actions. */
export function usePanGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection, setSelected } = useCardSelection();
  const { config: panConfig, handleConfigChange } = useGameConfig<PanConfig>(DEFAULT_PAN_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(panApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, DEFAULT_PAN_CONFIG);
  }, [exec]);

  const handleDrawStock = useCallback(() => {
    exec('drawstock');
  }, [exec]);

  const handleDrawDiscard = useCallback(() => {
    exec('drawdiscard');
  }, [exec]);

  const handleMeld = useCallback(() => {
    if (selectedCardIndices.length < 3) return;
    exec('meld', { cardIndices: [...selectedCardIndices] });
  }, [exec, selectedCardIndices]);

  const handleLayoff = useCallback(
    (meldOwner: number, meldIdx: number) => {
      if (selectedCardIndices.length !== 1) return;
      exec('layoff', { meldOwner, meldIdx, cardIndex: selectedCardIndices[0] });
    },
    [exec, selectedCardIndices],
  );

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('discard', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  const handleNextRound = useCallback(() => {
    exec('nextround');
  }, [exec]);

  /** Replaces the current selection with the exact given hand indices (used by meld-candidate hints). */
  const selectCards = useCallback(
    (indices: number[]) => {
      setSelected([...indices]);
    },
    [setSelected],
  );

  return {
    state,
    loading,
    error,
    exec,
    panConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    selectCards,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleMeld,
    handleLayoff,
    handleDiscard,
    handleNextRound,
    retry,
  };
}
