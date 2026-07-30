import { useCallback, useEffect } from 'react';
import { skitgubbeApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Skitgubbe game state.
 *
 * No local selection state: which cards may be played and whether the pile may
 * be picked up both come from the server, which owns the beat rule.
 */
export function useSkitgubbeGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(skitgubbeApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  const handlePlay = useCallback(
    (handIdx: number) => {
      runApi('play', handIdx);
    },
    [runApi],
  );

  const handlePickUp = useCallback(() => {
    runApi('pickup');
  }, [runApi]);

  return { state, loading, error, exec: runApi, handleReset, handlePlay, handlePickUp, retry };
}
