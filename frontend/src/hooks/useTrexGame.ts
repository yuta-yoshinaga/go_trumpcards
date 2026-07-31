import { useCallback, useEffect } from 'react';
import { trexApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Trex game state.
 *
 * No local state: which contracts are still available, which cards may be
 * played and whether a pass is possible all come from the server, which owns
 * the per-contract rules.
 */
export function useTrexGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(trexApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  const handleChoose = useCallback(
    (contract: number) => {
      // Contract 0 is the king of hearts, so it must reach the server as a value.
      runApi('choose', contract);
    },
    [runApi],
  );

  const handlePlay = useCallback(
    (handIdx: number) => {
      runApi('play', undefined, handIdx);
    },
    [runApi],
  );

  const handlePass = useCallback(() => {
    runApi('pass');
  }, [runApi]);

  const handleNextDeal = useCallback(() => {
    runApi('next');
  }, [runApi]);

  return {
    state,
    loading,
    error,
    exec: runApi,
    handleReset,
    handleChoose,
    handlePlay,
    handlePass,
    handleNextDeal,
    retry,
  };
}
