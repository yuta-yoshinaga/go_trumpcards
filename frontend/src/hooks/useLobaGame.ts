import { useCallback, useEffect } from 'react';
import { lobaApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Loba game state.
 *
 * No local rule state: whether a set of cards is a legal meld, and whether a
 * card fits an existing one, are both decided by the server.
 */
export function useLobaGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(lobaApi.exec);

  const runApi = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    runApi('reset');
  }, [runApi]);

  const handleReset = useCallback(() => {
    runApi('reset');
  }, [runApi]);

  const handleDrawStock = useCallback(() => {
    runApi('drawstock');
  }, [runApi]);

  const handleDrawDiscard = useCallback(() => {
    runApi('drawdiscard');
  }, [runApi]);

  const handleMeld = useCallback(
    (cardIndices: number[]) => {
      runApi('meld', undefined, undefined, cardIndices);
    },
    [runApi],
  );

  const handleLayOff = useCallback(
    (cardIndex: number, meldIndex: number) => {
      runApi('layoff', cardIndex, meldIndex);
    },
    [runApi],
  );

  const handleDiscard = useCallback(
    (cardIndex: number) => {
      runApi('discard', cardIndex);
    },
    [runApi],
  );

  const handleNextRound = useCallback(() => {
    runApi('next');
  }, [runApi]);

  return {
    state,
    loading,
    error,
    exec: runApi,
    handleReset,
    handleDrawStock,
    handleDrawDiscard,
    handleMeld,
    handleLayOff,
    handleDiscard,
    handleNextRound,
    retry,
  };
}
