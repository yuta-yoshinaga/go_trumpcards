import { useCallback, useState } from 'react';
import { type PenguinMoveZone, penguinApi } from '../api/gameApi';
import type { PenguinHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Penguin game state, source selection, hints, and moves. */
export function usePenguinGame() {
  const [selectedSource, setSelectedSource] = useState<PenguinMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof penguinApi.exec>>,
    Parameters<typeof penguinApi.exec>,
    PenguinHint
  >(penguinApi.exec, {
    onClearSelection,
    hintApi: () => penguinApi.exec('hint'),
  });

  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: PenguinMoveZone) => {
    setSelectedSource((prev) => {
      if (
        prev &&
        prev.zone === zone.zone &&
        prev.col === zone.col &&
        prev.cell === zone.cell &&
        prev.cardIndex === zone.cardIndex
      ) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: PenguinMoveZone) => {
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
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
  };
}
