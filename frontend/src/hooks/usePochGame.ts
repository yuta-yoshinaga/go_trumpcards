import { useCallback, useEffect } from 'react';
import { pochApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Poch game state.
 *
 * No local rule state: which pools a hand claims, which same-rank set wins the
 * pochen, and whether a card continues the run are all decided by the server.
 */
export function usePochGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(pochApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  const handleBet = useCallback(() => {
    runApi('bet');
  }, [runApi]);

  const handleFold = useCallback(() => {
    runApi('fold');
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

  return {
    state,
    loading,
    error,
    exec: runApi,
    handleReset,
    handleBet,
    handleFold,
    handlePlay,
    handleNextDeal,
    retry,
  };
}
