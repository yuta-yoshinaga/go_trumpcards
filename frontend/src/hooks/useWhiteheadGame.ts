import { useCallback, useState } from 'react';
import { type WhiteheadConfigInput, type WhiteheadMoveZone, whiteheadApi } from '../api/gameApi';
import type { WhiteheadHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Whitehead game state, source selection, hints, and moves. */
export function useWhiteheadGame() {
  const [selectedSource, setSelectedSource] = useState<WhiteheadMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof whiteheadApi.exec>>,
    Parameters<typeof whiteheadApi.exec>,
    WhiteheadHint
  >(whiteheadApi.exec, {
    onClearSelection,
    hintApi: () => whiteheadApi.exec('hint'),
  });

  const handleDraw = useCallback(() => runAction('draw'), [runAction]);
  const handleResetWithConfig = useCallback(
    (config: WhiteheadConfigInput) => runAction('reset', undefined, undefined, config),
    [runAction],
  );
  const handleUndoEscape = useCallback(
    (n: number) => runAction('undo_n', undefined, undefined, undefined, n),
    [runAction],
  );

  const handleSelectSource = useCallback((zone: WhiteheadMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: WhiteheadMoveZone) => {
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
    handleDraw,
    handleResetWithConfig,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
  };
}
