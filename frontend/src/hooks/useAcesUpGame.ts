import { useCallback, useEffect, useRef, useState } from 'react';
import { acesupApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { AcesUpCard, AcesUpHint, AcesUpResponse } from '../types/card';
import { useGameApi } from './useGameApi';
import { useIsMounted } from './useIsMounted';

/** Upper bound on batch-removal iterations (one per non-ace card) to guard against loops. */
const MAX_REMOVE_ALL_STEPS = 52;

/** Index of the first column whose top card is currently removable, or -1 if none. */
function firstRemovableCol(columns: AcesUpCard[][]): number {
  return columns.findIndex((col) => col.length > 0 && col[col.length - 1]?.removable === true);
}

/** Hook that manages Aces Up game state, hints, deal/remove/move actions. */
export function useAcesUpGame() {
  const { state, setState, loading, error, exec, retry } = useGameApi(acesupApi.exec);
  const [hint, setHint] = useState<AcesUpHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [isRemovingAll, setIsRemovingAll] = useState(false);
  // Mirror `state` in a ref so handleRemoveAll can read the freshest board at
  // click time without listing `state` in its deps (keeps the callback stable),
  // plus a re-entry guard so rapid double-clicks can't start two batch loops.
  const stateRef = useRef<AcesUpResponse | null>(state);
  stateRef.current = state;
  const removingAllRef = useRef(false);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleDraw = useCallback(() => {
    setHint(null);
    exec('draw');
  }, [exec]);

  const handleReset = useCallback(() => {
    setHint(null);
    exec('reset');
  }, [exec]);

  const handleGiveUp = useCallback(() => {
    setHint(null);
    exec('giveup');
  }, [exec]);

  const isMounted = useIsMounted();

  const handleHint = useCallback(async () => {
    try {
      const res = await acesupApi.exec('hint');
      // Navigating away mid-request must not write to a gone component (#4447).
      if (!isMounted()) return;
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, [isMounted]);

  const handleUndo = useCallback(() => {
    setHint(null);
    exec('undo');
  }, [exec]);

  /** Batch undo to escape stalemate. */
  const handleUndoEscape = useCallback(
    (n: number) => {
      setHint(null);
      exec('undo_n', undefined, n);
    },
    [exec],
  );

  const handleRemove = useCallback(
    (col: number) => {
      setHint(null);
      exec('remove', col);
    },
    [exec],
  );

  const handleMove = useCallback(
    (col: number) => {
      setHint(null);
      exec('move', col);
    },
    [exec],
  );

  /**
   * Discards every currently-removable top card in one action. Removability is
   * order-dependent (a card can be discarded only while a higher card of the
   * same suit sits on top of another column), so this drives the existing
   * single `remove` action sequentially, re-reading the fresh board after each
   * discard and stopping once nothing is removable — which also sweeps up cards
   * that only become removable once the ones above them are gone. The re-entry
   * guard blocks a second concurrent loop, and `isRemovingAll` lets the page
   * disable its controls while the batch runs.
   */
  const handleRemoveAll = useCallback(async () => {
    if (removingAllRef.current) return;
    removingAllRef.current = true;
    setIsRemovingAll(true);
    setHint(null);
    let col = -1;
    try {
      let columns = stateRef.current?.columns;
      for (let step = 0; step < MAX_REMOVE_ALL_STEPS && columns; step++) {
        col = firstRemovableCol(columns);
        if (col < 0) break;
        const res = await acesupApi.exec('remove', col);
        // Abandon the remove-all loop if the player left mid-sequence (#4447).
        if (!isMounted()) return;
        setState(res);
        columns = res.columns;
      }
    } catch {
      // Re-issue the failing removal through the shared exec so the network
      // failure surfaces via the standard error/retry channel.
      if (col >= 0) await exec('remove', col);
    } finally {
      removingAllRef.current = false;
      if (isMounted()) setIsRemovingAll(false);
    }
  }, [exec, setState, isMounted]);

  return {
    state,
    loading,
    error,
    exec,
    hintError,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleUndoEscape,
    handleRemove,
    handleRemoveAll,
    handleMove,
    isRemovingAll,
    retry,
  };
}
