import { useCallback, useState } from 'react';
import { type WillOTheWispMoveZone, willothewispApi } from '../api/games/willothewisp';
import type { WillOTheWispHint } from '../types/games/willothewisp';
import { useSolitaireGameBase } from './useSolitaireGameBase';

const runCommand = willothewispApi.exec;

/** Hook that manages WillOTheWisp game state, source selection, hints, and moves. */
export function useWillOTheWispGame() {
  const [selectedSource, setSelectedSource] = useState<WillOTheWispMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof runCommand>>,
    Parameters<typeof runCommand>,
    WillOTheWispHint
  >(runCommand, {
    onClearSelection,
    hintApi: () => runCommand('hint'),
  });

  const handleDeal = useCallback(() => runAction('deal'), [runAction]);
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: WillOTheWispMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: WillOTheWispMoveZone) => {
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
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
  };
}
