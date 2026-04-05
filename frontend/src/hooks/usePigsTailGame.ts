import { useCallback, useEffect } from 'react';
import { pigtailApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/** Hook that manages Pig's Tail game state and player actions. */
export function usePigsTailGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(pigtailApi.exec);

  const gameExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    gameExec('reset');
  }, [gameExec]);

  const handleDraw = useCallback(() => {
    gameExec('draw');
  }, [gameExec]);

  const handleReset = useCallback(() => {
    gameExec('reset');
  }, [gameExec]);

  return { state, loading, error, gameExec, handleDraw, handleReset, retry };
}
