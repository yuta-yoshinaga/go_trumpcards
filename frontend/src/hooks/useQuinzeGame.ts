import { useCallback, useEffect } from 'react';
import { quinzeApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/** Hook that manages Quinze game state and the round's actions. */
export function useQuinzeGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(quinzeApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  const handleBet = useCallback(
    (amount: number) => {
      runApi('bet', amount);
    },
    [runApi],
  );

  /** Deal the round the human is banking. The banker places no stake. */
  const handleDeal = useCallback(() => {
    runApi('deal');
  }, [runApi]);

  const handleHit = useCallback(() => {
    runApi('hit');
  }, [runApi]);

  const handleStand = useCallback(() => {
    runApi('stand');
  }, [runApi]);

  const handleBankerHit = useCallback(() => {
    runApi('bankerhit');
  }, [runApi]);

  const handleBankerStand = useCallback(() => {
    runApi('bankerstand');
  }, [runApi]);

  return {
    state,
    loading,
    error,
    exec: runApi,
    handleReset,
    handleBet,
    handleDeal,
    handleHit,
    handleStand,
    handleBankerHit,
    handleBankerStand,
    retry,
  };
}
