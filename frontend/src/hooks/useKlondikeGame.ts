import { useCallback, useState } from 'react';
import { type KlondikeConfigInput, type KlondikeMoveZone, klondikeApi } from '../api/gameApi';
import type { KlondikeHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Klondike game state, source selection, hints, and moves. */
export function useKlondikeGame() {
  const [selectedSource, setSelectedSource] = useState<KlondikeMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof klondikeApi.exec>>,
    Parameters<typeof klondikeApi.exec>,
    KlondikeHint
  >(klondikeApi.exec, {
    onClearSelection,
    hintApi: () => klondikeApi.exec('hint'),
  });

  const handleDraw = useCallback(() => runAction('draw'), [runAction]);
  const handleResetWithConfig = useCallback(
    (config: KlondikeConfigInput) => runAction('reset', undefined, undefined, config),
    [runAction],
  );
  const handleUndoEscape = useCallback(
    (n: number) => runAction('undo_n', undefined, undefined, undefined, n),
    [runAction],
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
