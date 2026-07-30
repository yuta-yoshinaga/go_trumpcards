import { useCallback, useEffect } from 'react';
import { mushiApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Mushi game state.
 *
 * There is no local selection state to keep: a capture choice is a single
 * click on a field card the SERVER has already marked selectable, so the page
 * never has to decide what may be taken.
 */
export function useMushiGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(mushiApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  /** Play a card from the hand. */
  const handlePlay = useCallback(
    (handIdx: number) => {
      runApi('play', handIdx);
    },
    [runApi],
  );

  /** Take a field card — used for both the same-month choice and the wild's target. */
  const handleSelect = useCallback(
    (fieldIdx: number) => {
      runApi('select', undefined, fieldIdx);
    },
    [runApi],
  );

  const handleNextRound = useCallback(() => {
    runApi('next');
  }, [runApi]);

  return { state, loading, error, exec: runApi, handleReset, handlePlay, handleSelect, handleNextRound, retry };
}
