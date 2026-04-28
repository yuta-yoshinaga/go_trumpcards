import { useCallback, useState } from 'react';
import { type KlondikeConfigInput, type KlondikeMoveZone, klondikeApi } from '../api/gameApi';
import type { KlondikeHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Klondike game state, source selection, hints, and moves. */
export function useKlondikeGame() {
  const [selectedSource, setSelectedSource] = useState<KlondikeMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const base = useSolitaireGameBase<
    Awaited<ReturnType<typeof klondikeApi.exec>>,
    Parameters<typeof klondikeApi.exec>,
    KlondikeHint
  >(klondikeApi.exec, {
    onClearSelection,
    hintApi: () => klondikeApi.exec('hint'),
  });

  const handleDraw = useCallback(() => base.runAction('draw'), [base.runAction]);
  const handleResetWithConfig = useCallback(
    (config: KlondikeConfigInput) => base.runAction('reset', undefined, undefined, config),
    [base.runAction],
  );
  const handleUndoEscape = useCallback(
    (n: number) => base.runAction('undo_n', undefined, undefined, undefined, n),
    [base.runAction],
  );

  const handleSelectSource = useCallback((zone: KlondikeMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: KlondikeMoveZone) => {
      if (!selectedSource) return;
      base.setHint(null);
      void base.apiCall('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, base],
  );

  return {
    state: base.state,
    loading: base.loading,
    error: base.error,
    hintError: base.hintError,
    exec: base.apiCall,
    selectedSource,
    hint: base.hint,
    handleDraw,
    handleReset: base.handleReset,
    handleResetWithConfig,
    handleGiveUp: base.handleGiveUp,
    handleHint: base.handleHint,
    handleAutoComplete: base.handleAutoComplete,
    handleUndo: base.handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting: base.isAutoCompleting,
    retry: base.retry,
  };
}
