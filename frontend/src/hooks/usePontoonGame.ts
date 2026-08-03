import { useCallback, useEffect } from 'react';
import { pontoonApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/** Hook that manages Pontoon game state and the round's actions. */
export function usePontoonGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(pontoonApi.exec);

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

  const handleStick = useCallback(() => {
    runApi('stick');
  }, [runApi]);

  const handleTwist = useCallback(() => {
    runApi('twist');
  }, [runApi]);

  const handleBuy = useCallback(
    (extra: number) => {
      runApi('buy', extra);
    },
    [runApi],
  );

  const handleSplit = useCallback(() => {
    runApi('split');
  }, [runApi]);

  const handleBankerTwist = useCallback(() => {
    runApi('bankertwist');
  }, [runApi]);

  const handleBankerStay = useCallback(() => {
    runApi('bankerstay');
  }, [runApi]);

  return {
    state,
    loading,
    error,
    exec: runApi,
    handleReset,
    handleBet,
    handleDeal,
    handleStick,
    handleTwist,
    handleBuy,
    handleSplit,
    handleBankerTwist,
    handleBankerStay,
    retry,
  };
}
