import { useCallback, useEffect } from 'react';
import { popejoanApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Pope Joan game state.
 *
 * No local rule state: whether a card continues the run, which compartment a
 * card pays, and who is excused the per-card payment are all decided by the
 * server.
 */
export function usePopeJoanGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(popejoanApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  const handlePlay = useCallback(
    (cardIndex: number) => {
      runApi('play', cardIndex);
    },
    [runApi],
  );

  const handleNextDeal = useCallback(() => {
    runApi('next');
  }, [runApi]);

  return { state, loading, error, exec: runApi, handleReset, handlePlay, handleNextDeal, retry };
}
