import { useCallback, useState } from 'react';
import { type SpiderConfigInput, type SpiderMoveZone, spiderApi } from '../api/gameApi';
import type { SpiderHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Spider Solitaire game state, source selection, hints, and moves. */
export function useSpiderGame() {
  const [selectedSource, setSelectedSource] = useState<SpiderMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof spiderApi.exec>>,
    Parameters<typeof spiderApi.exec>,
    SpiderHint
  >(spiderApi.exec, {
    onClearSelection,
    hintApi: () => spiderApi.exec('hint'),
  });

  const handleDeal = useCallback(() => runAction('deal'), [runAction]);
  const handleResetWithConfig = useCallback(
    (config: SpiderConfigInput) => runAction('reset', undefined, undefined, config),
    [runAction],
  );
  const handleUndoEscape = useCallback(
    (n: number) => runAction('undo_n', undefined, undefined, undefined, n),
    [runAction],
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
      setHint(null);
      void apiCall('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, apiCall, setHint],
  );

  return {
    ...rest,
    runAction,
    setHint,
    exec: apiCall,
    selectedSource,
    handleDeal,
    handleResetWithConfig,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
  };
}
