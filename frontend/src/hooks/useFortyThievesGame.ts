import { useCallback, useEffect, useState } from 'react';
import { type FortyThievesMoveZone, fortyThievesApi } from '../api/gameApi';
import type { Card, FortyThievesHint } from '../types/card';
import { fortyThievesFoundationTarget } from '../utils/fortyThievesFoundationTarget';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';
import { useHintRequest } from './useHintRequest';

/** Hook that manages Forty Thieves game state, source selection, hints, and moves. */
export function useFortyThievesGame() {
  const { state, loading, error, exec: rawExec, retry } = useGameApi(fortyThievesApi.exec);
  const [selectedSource, setSelectedSource] = useState<FortyThievesMoveZone | null>(null);
  const [hint, setHint] = useState<FortyThievesHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const { isAutoCompleting, startAutoComplete } = useAutoCompleteState();

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('draw');
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
    fetchHint: () => fortyThievesApi.exec('hint'),
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

  const handleSelectSource = useCallback((zone: FortyThievesMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: FortyThievesMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      exec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, exec],
  );

  /**
   * Double-click / double-tap shortcut: auto-send an exposed top card (from the
   * waste or a tableau column) straight to a foundation when a legal target
   * exists; otherwise do nothing (no error, selection cleared). Mirrors the
   * Easthaven foundation shortcut. `source` is the card's own zone.
   */
  const handleFoundationShortcut = useCallback(
    (source: FortyThievesMoveZone, card: Card) => {
      const target = fortyThievesFoundationTarget(card, state?.foundation ?? []);
      if (!target) return;
      setHint(null);
      exec('move', source, target);
      setSelectedSource(null);
    },
    [state?.foundation, exec],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec,
    selectedSource,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    handleFoundationShortcut,
    isAutoCompleting,
    retry,
  };
}
