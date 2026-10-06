import { useCallback, useState } from 'react';
import { type RankAndFileMoveZone, rankAndFileApi } from '../api/gameApi';
import type { Card, RankAndFileHint } from '../types/card';
import { rankAndFileFoundationTarget } from '../utils/rankAndFileFoundationTarget';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Rank and File game state, source selection, hints, and moves. */
export function useRankAndFileGame() {
  const [selectedSource, setSelectedSource] = useState<RankAndFileMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);
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
  } = useSolitaireGameBase<
    Awaited<ReturnType<typeof rankAndFileApi.exec>>,
    Parameters<typeof rankAndFileApi.exec>,
    RankAndFileHint
  >(rankAndFileApi.exec, {
    onClearSelection,
    hintApi: () => rankAndFileApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('draw');
  }, [exec, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: RankAndFileMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: RankAndFileMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      exec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, exec, setHint],
  );

  /**
   * Double-click / double-tap shortcut: auto-send an exposed top card (from the
   * waste or a tableau column) straight to a foundation when a legal target
   * exists; otherwise do nothing (no error, selection cleared). Mirrors the
   * Easthaven foundation shortcut. `source` is the card's own zone.
   */
  const handleFoundationShortcut = useCallback(
    (source: RankAndFileMoveZone, card: Card) => {
      const target = rankAndFileFoundationTarget(card, state?.foundation ?? []);
      if (!target) return;
      setHint(null);
      exec('move', source, target);
      setSelectedSource(null);
    },
    [state?.foundation, exec, setHint],
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
