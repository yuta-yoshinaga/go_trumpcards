import { useCallback, useEffect } from 'react';
import { nainjauneApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Le Nain Jaune game state.
 *
 * No local rule state: whether a card continues the run (by rank, ignoring
 * suit), which box a card claims, and what a hand is worth in points are all
 * decided by the server.
 */
export function useNainJauneGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(nainjauneApi.exec);

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
