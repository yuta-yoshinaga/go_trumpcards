import { useCallback, useEffect } from 'react';
import { settemezzoApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/** Hook that manages Sette e Mezzo game state and the round's actions. */
export function useSetteEMezzoGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(settemezzoApi.exec);

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

  /**
   * Set the matta's value. The argument is in HALVES, matching the wire format,
   * so the caller never has to round a half-point through a float.
   */
  const handleMatta = useCallback(
    (halves: number) => {
      runApi('matta', halves);
    },
    [runApi],
  );

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
    handleMatta,
    handleBankerHit,
    handleBankerStand,
    retry,
  };
}
