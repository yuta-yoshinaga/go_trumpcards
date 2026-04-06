import { useCallback, useEffect } from 'react';
import { pigtailApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/** Hook that manages Pig's Tail game state and player actions. */
export function usePigsTailGame() {
  const { state, loading, error, exec, retry } = useGameApi(pigtailApi.exec);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleDraw = useCallback(() => {
    exec('draw');
  }, [exec]);

  const handleReset = useCallback(() => {
    exec('reset');
  }, [exec]);

  return { state, loading, error, handleDraw, handleReset, retry };
}
