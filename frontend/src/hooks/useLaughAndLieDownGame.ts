import { useCallback, useEffect } from 'react';
import { laughandliedownApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Laugh and Lie Down game state.
 *
 * No local selection state: which cards match, and which of them can take
 * three, both come from the server, which owns the capture rule.
 */
export function useLaughAndLieDownGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(laughandliedownApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  const handlePlay = useCallback(
    (handIdx: number, takeCount = 1) => {
      runApi('play', handIdx, takeCount);
    },
    [runApi],
  );

  return { state, loading, error, exec: runApi, handleReset, handlePlay, retry };
}
