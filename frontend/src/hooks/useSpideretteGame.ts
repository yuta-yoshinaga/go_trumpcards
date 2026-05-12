import { useCallback, useState } from 'react';
import { type SpideretteMoveZone, spideretteApi } from '../api/gameApi';
import type { SpideretteHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

const runCommand = spideretteApi.exec;

/** Hook that manages Spiderette game state, source selection, hints, and moves. */
export function useSpideretteGame() {
  const [selectedSource, setSelectedSource] = useState<SpideretteMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const base = useSolitaireGameBase<
    Awaited<ReturnType<typeof runCommand>>,
    Parameters<typeof runCommand>,
    SpideretteHint
  >(runCommand, {
    onClearSelection,
    hintApi: () => runCommand('hint'),
  });

  const handleDeal = useCallback(() => base.runAction('deal'), [base.runAction]);
  const handleUndoEscape = useCallback(
    (n: number) => base.runAction('undo_n', undefined, undefined, n),
    [base.runAction],
  );

  const handleSelectSource = useCallback((zone: SpideretteMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: SpideretteMoveZone) => {
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
    exec: base.apiCall,
    hintError: base.hintError,
    selectedSource,
    hint: base.hint,
    handleDeal,
    handleReset: base.handleReset,
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
