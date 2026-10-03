import { useCallback, useEffect, useRef, useState } from 'react';
import { type BakersDozenMoveZone, bakersDozenApi } from '../api/gameApi';
import type { BakersDozenHint } from '../types/card';
import { BakersDozenPhase } from '../types/phases';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';
import { useHintRequest } from './useHintRequest';

/** Hook that manages Baker's Dozen game state, source selection, hints, and moves. */
export function useBakersDozenGame() {
  const [autoCompleteStatus, setAutoCompleteStatus] = useState<'inProgress' | 'finished' | 'interrupted' | null>(null);
  const {
    state,
    loading,
    error,
    exec: rawExec,
    retry: rawRetry,
  } = useGameApi(bakersDozenApi.exec, {
    onSuccess: (res, args) => {
      if (args[0] === 'autocomplete') {
        setAutoCompleteStatus(res.phase === BakersDozenPhase.GAME_CLEAR ? 'finished' : 'interrupted');
      }
    },
  });
  const previousError = useRef(error);
  useEffect(() => {
    if (error && !loading && !previousError.current && autoCompleteStatus === 'inProgress') {
      setAutoCompleteStatus('interrupted');
    }
    previousError.current = error;
  }, [error, loading, autoCompleteStatus]);
  const [selectedSource, setSelectedSource] = useState<BakersDozenMoveZone | null>(null);
  const [hint, setHint] = useState<BakersDozenHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const { isAutoCompleting, startAutoComplete } = useAutoCompleteState();
  const lastCommand = useRef<string | undefined>(undefined);

  const exec = useCallback(
    (...args: Parameters<typeof rawExec>) => {
      lastCommand.current = args[0];
      if (args[0] === 'reset') setAutoCompleteStatus(null);
      const request = rawExec(...args);
      if (args[0] === 'autocomplete') setAutoCompleteStatus('inProgress');
      return request;
    },
    [rawExec],
  );
  const retry = useCallback(() => {
    if (lastCommand.current === 'autocomplete') setAutoCompleteStatus('inProgress');
    return rawRetry();
  }, [rawRetry]);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleReset = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('reset');
  }, [exec]);

  const handleGiveUp = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('giveup');
  }, [exec]);

  const handleHint = useHintRequest({
    fetchHint: () => bakersDozenApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
  });

  const handleAutoComplete = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    startAutoComplete();
    exec('autocomplete');
  }, [exec, startAutoComplete]);

  const handleUndo = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('undo');
  }, [exec]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      exec('undo_n', undefined, undefined, n);
    },
    [exec],
  );

  const handleSelectSource = useCallback((zone: BakersDozenMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: BakersDozenMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      exec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, exec],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec,
    selectedSource,
    hint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
    autoCompleteStatus,
    retry,
  };
}
