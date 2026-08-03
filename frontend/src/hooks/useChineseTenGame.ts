import { useCallback, useEffect } from 'react';
import { chinesetenApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Chinese Ten game state.
 *
 * No local selection state: a capture choice is one click on a layout card the
 * SERVER has already marked selectable, so the page never decides what may be
 * taken.
 */
export function useChineseTenGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(chinesetenApi.exec);

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

  const handleSelect = useCallback(
    (layoutIdx: number) => {
      runApi('select', undefined, layoutIdx);
    },
    [runApi],
  );

  return { state, loading, error, exec: runApi, handleReset, handlePlay, handleSelect, retry };
}
