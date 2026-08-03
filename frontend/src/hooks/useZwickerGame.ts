import { useCallback, useEffect } from 'react';
import { zwickerApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Zwicker game state.
 *
 * No local rule state: which combinations add up to a played value, and whether
 * a build is legal, are both decided by the server. The matching values arrive
 * on each card, so the page never re-derives them either.
 */
export function useZwickerGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(zwickerApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  /**
   * Plays `cardIndex` as `playedValue` and captures the selected table cards
   * and builds. The value is separate from the card because an ace or court
   * card has two.
   */
  const handleTake = useCallback(
    (cardIndex: number, playedValue: number, tableIndices: number[], buildIndices: number[]) => {
      runApi('take', { cardIndex, playedValue, tableIndices, buildIndices });
    },
    [runApi],
  );

  const handleBuild = useCallback(
    (cardIndex: number, tableIndices: number[], declaredValue: number) => {
      runApi('build', { cardIndex, tableIndices, declaredValue });
    },
    [runApi],
  );

  const handleTrail = useCallback(
    (cardIndex: number) => {
      runApi('trail', { cardIndex });
    },
    [runApi],
  );

  const handleNextRound = useCallback(() => {
    runApi('next');
  }, [runApi]);

  return {
    state,
    loading,
    error,
    exec: runApi,
    handleReset,
    handleTake,
    handleBuild,
    handleTrail,
    handleNextRound,
    retry,
  };
}
