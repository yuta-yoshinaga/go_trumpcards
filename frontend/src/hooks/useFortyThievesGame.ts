import { useCallback, useState } from 'react';
import { type FortyThievesMoveZone, fortyThievesApi } from '../api/gameApi';
import type { Card, FortyThievesHint } from '../types/card';
import { fortyThievesFoundationTarget } from '../utils/fortyThievesFoundationTarget';

import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Forty Thieves game state, source selection, hints, and moves. */
export function useFortyThievesGame() {
  const [selectedSource, setSelectedSource] = useState<FortyThievesMoveZone | null>(null);

  const onClearSelection = useCallback(() => setSelectedSource(null), []);
  const {
    state,
    loading,
    error,
    retry,
    apiCall: exec,
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
    Awaited<ReturnType<typeof fortyThievesApi.exec>>,
    Parameters<typeof fortyThievesApi.exec>,
    FortyThievesHint
  >(fortyThievesApi.exec, {
    onClearSelection,
    hintApi: () => fortyThievesApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('draw');
  }, [exec, setHint]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      setSelectedSource(null);
      setHint(null);
      exec('undo_n', undefined, undefined, n);
    },
    [exec, setHint],
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
    [selectedSource, exec, setHint],
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
