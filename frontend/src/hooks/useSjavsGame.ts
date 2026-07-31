import { useCallback, useEffect } from 'react';
import { sjavsApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Sjavs game state.
 *
 * No local state: what may be bid, what may be played and whether the hand is
 * over all come from the server, which owns the permanent-trump rules.
 */
export function useSjavsGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(sjavsApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  const handleBid = useCallback(
    (length: number) => {
      // 0 is a pass, and it must still reach the server as a value.
      runApi('bid', length);
    },
    [runApi],
  );

  const handlePlay = useCallback(
    (handIdx: number) => {
      runApi('play', undefined, handIdx);
    },
    [runApi],
  );

  const handleNextHand = useCallback(() => {
    runApi('next');
  }, [runApi]);

  return { state, loading, error, exec: runApi, handleReset, handleBid, handlePlay, handleNextHand, retry };
}
