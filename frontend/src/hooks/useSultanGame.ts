import { useCallback } from 'react';
import { type SultanMoveZone, sultanApi } from '../api/gameApi';
import type { SultanHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Sultan of Turkey game state, source selection, hints, redeal, and moves. */
export function useSultanGame() {
  const {
    apiCall: exec,
    runAction,
    state,
    loading,
    error,
    retry,
    hint,
    hintError,
    isAutoCompleting,
    setHint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
  } = useSolitaireGameBase<Awaited<ReturnType<typeof sultanApi.exec>>, Parameters<typeof sultanApi.exec>, SultanHint>(
    sultanApi.exec,
    {
      hintApi: () => sultanApi.exec('hint'),
      selectHint: (res) => res.hint,
    },
  );

  const handleDraw = useCallback(() => {
    setHint(null);
    exec('draw');
  }, [exec, setHint]);

  const handleRedeal = useCallback(() => {
    setHint(null);
    exec('redeal');
  }, [exec, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  /** Play a divan slot or the waste top onto its matching foundation. */
  const handlePlay = useCallback(
    (source: SultanMoveZone) => {
      setHint(null);
      exec('move', source);
    },
    [exec, setHint],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec,
    hint,
    handleDraw,
    handleRedeal,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handlePlay,
    isAutoCompleting,
    retry,
  };
}
