import { useCallback, useEffect } from 'react';
import { toepenApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Toepen game state.
 *
 * There is no local selection state: a card is played on a single click, and
 * which cards are legal comes from the server's `validPlayIndices`.
 */
export function useToepenGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(toepenApi.exec);

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

  const handleToep = useCallback(() => {
    runApi('toep');
  }, [runApi]);

  /** Answer an outstanding toep: stay in, or fold. */
  const handleRespond = useCallback(
    (stay: boolean) => {
      runApi('answer', undefined, stay);
    },
    [runApi],
  );

  const handleNextHand = useCallback(() => {
    runApi('next');
  }, [runApi]);

  return {
    state,
    loading,
    error,
    exec: runApi,
    handleReset,
    handlePlay,
    handleToep,
    handleRespond,
    handleNextHand,
    retry,
  };
}
