import { useCallback, useState } from 'react';
import { type SpideretteMoveZone, spideretteApi } from '../api/gameApi';
import type { SpideretteHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

const runCommand = spideretteApi.exec;

/** Hook that manages Spiderette game state, source selection, hints, and moves. */
export function useSpideretteGame() {
  const [selectedSource, setSelectedSource] = useState<SpideretteMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof runCommand>>,
    Parameters<typeof runCommand>,
    SpideretteHint
  >(runCommand, {
    onClearSelection,
    hintApi: () => runCommand('hint'),
  });

  const handleDeal = useCallback(() => runAction('deal'), [runAction]);
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

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
