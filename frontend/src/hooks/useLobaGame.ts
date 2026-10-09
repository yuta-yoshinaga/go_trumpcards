import { useCallback, useEffect, useRef, useState } from 'react';
import { lobaApi } from '../api/gameApi';
import { isRejectedAction, useGameApi } from './useGameApi';

/**
 * Hook that manages Loba game state.
 *
 * No local rule state: whether a set of cards is a legal meld, and whether a
 * card fits an existing one, are both decided by the server.
 */
export function useLobaGame() {
  const pendingDiscardIndex = useRef<number | null>(null);
  const actionAccepted = useRef<boolean | null>(null);
  const [drawnDiscardIndex, setDrawnDiscardIndex] = useState<number | null>(null);
  const {
    state,
    loading,
    error,
    exec: rawExec,
    retry,
  } = useGameApi(lobaApi.exec, {
    onSuccess: (response, args) => {
      if (args[0] === 'meld' || args[0] === 'layoff') actionAccepted.current = !isRejectedAction(response);
      const index = pendingDiscardIndex.current;
      pendingDiscardIndex.current = null;
      if (args[0] !== 'drawdiscard' || index === null) return;
      const human = response.players.find((player) => player.isHuman);
      if (human && human.cards.length > index) setDrawnDiscardIndex(index);
    },
  });

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
    const human = state?.players.find((player) => player.isHuman);
    pendingDiscardIndex.current = human?.cards.length ?? null;
    setDrawnDiscardIndex(null);
    runApi('drawdiscard');
  }, [runApi, state]);

  const handleMeld = useCallback(
    async (cardIndices: number[]) => {
      actionAccepted.current = null;
      await runApi('meld', undefined, undefined, cardIndices);
      return actionAccepted.current === true;
    },
    [runApi],
  );

  const handleLayOff = useCallback(
    async (cardIndex: number, meldIndex: number) => {
      actionAccepted.current = null;
      await runApi('layoff', cardIndex, meldIndex);
      return actionAccepted.current === true;
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
    drawnDiscardIndex,
    handleMeld,
    handleLayOff,
    handleDiscard,
    handleNextRound,
    retry,
  };
}
