import { useCallback, useEffect } from 'react';
import { niuniuApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/** Hook that manages Niu Niu game state. The round resolves at the bet. */
export function useNiuNiuGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(niuniuApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  /** Bet, deal and settle in one call -- the game offers no decisions after this. */
  const handleBet = useCallback(
    (amount: number) => {
      runApi('bet', amount);
    },
    [runApi],
  );

  return { state, loading, error, exec: runApi, handleReset, handleBet, retry };
}
