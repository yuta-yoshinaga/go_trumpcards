import { useCallback, useState } from 'react';
import { type MrsMopConfigInput, type MrsMopMoveZone, mrsMopApi } from '../api/gameApi';
import type { MrsMopHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages MrsMop Solitaire game state, source selection, hints, and moves. */
export function useMrsMopGame() {
  const [selectedSource, setSelectedSource] = useState<MrsMopMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const { apiCall, runAction, setHint, ...rest } = useSolitaireGameBase<
    Awaited<ReturnType<typeof mrsMopApi.exec>>,
    Parameters<typeof mrsMopApi.exec>,
    MrsMopHint
  >(mrsMopApi.exec, {
    onClearSelection,
    hintApi: () => mrsMopApi.exec('hint'),
  });

  const handleResetWithConfig = useCallback(
    (config: MrsMopConfigInput) => runAction('reset', undefined, undefined, config),
    [runAction],
  );
  const handleUndoEscape = useCallback(
    (n: number) => runAction('undo_n', undefined, undefined, undefined, n),
    [runAction],
  );

  const handleSelectSource = useCallback((zone: MrsMopMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: MrsMopMoveZone) => {
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
    handleResetWithConfig,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
  };
}
