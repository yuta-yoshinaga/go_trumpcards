import { useCallback, useState } from 'react';
import { type SeahavenTowersMoveZone, seahaventowersApi } from '../api/gameApi';
import type { SeahavenTowersHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Seahaven Towers game state, source selection, hints, and moves. */
export function useSeahavenTowersGame() {
  const [selectedSource, setSelectedSource] = useState<SeahavenTowersMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof seahaventowersApi.exec>>,
    Parameters<typeof seahaventowersApi.exec>,
    SeahavenTowersHint
  >(seahaventowersApi.exec, {
    onClearSelection,
    hintApi: () => seahaventowersApi.exec('hint'),
  });

  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: SeahavenTowersMoveZone) => {
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
    (zone: SeahavenTowersMoveZone) => {
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
