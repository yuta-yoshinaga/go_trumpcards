import { useCallback, useState } from 'react';
import { type SpiderConfigInput, type SpiderMoveZone, spiderApi } from '../api/gameApi';
import type { SpiderHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Spider Solitaire game state, source selection, hints, and moves. */
export function useSpiderGame() {
  const [selectedSource, setSelectedSource] = useState<SpiderMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const base = useSolitaireGameBase<
    Awaited<ReturnType<typeof spiderApi.exec>>,
    Parameters<typeof spiderApi.exec>,
    SpiderHint
  >(spiderApi.exec, {
    onClearSelection,
    hintApi: () => spiderApi.exec('hint'),
  });

  const handleDeal = useCallback(() => base.runAction('deal'), [base.runAction]);
  const handleResetWithConfig = useCallback(
    (config: SpiderConfigInput) => base.runAction('reset', undefined, undefined, config),
    [base.runAction],
  );
  const handleUndoEscape = useCallback(
    (n: number) => base.runAction('undo_n', undefined, undefined, undefined, n),
    [base.runAction],
  );

  const handleSelectSource = useCallback((zone: SpiderMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: SpiderMoveZone) => {
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
