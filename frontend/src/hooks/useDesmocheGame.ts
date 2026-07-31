import { useCallback, useEffect } from 'react';
import { desmocheApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/**
 * Hook that manages Desmoche game state.
 *
 * No local rule state: whether a set of cards is a legal meld, whether a card
 * fits an existing one, and whether pulling a card out of a meld leaves it
 * valid are all decided by the server.
 */
export function useDesmocheGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(desmocheApi.exec);

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

  /**
   * Moves card `cardIndex` of `fromMeldIndex` into `toMeldIndex` — the move the
   * game is named after. `cardIndex` here indexes into the source meld, not the
   * hand.
   */
  const handleDesmoche = useCallback(
    (fromMeldIndex: number, cardIndex: number, toMeldIndex: number) => {
      runApi('desmoche', cardIndex, undefined, undefined, { fromMeldIndex, toMeldIndex });
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
    handleDesmoche,
    handleDiscard,
    handleNextRound,
    retry,
  };
}
