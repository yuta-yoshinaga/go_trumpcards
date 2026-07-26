import { useCallback, useState } from 'react';
import { type EightOffMoveZone, eightoffApi } from '../api/gameApi';
import type { EightOffHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Eight Off game state, source selection, hints, and moves. */
export function useEightOffGame() {
  const [selectedSource, setSelectedSource] = useState<EightOffMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);

  const {
    handleHint: baseHandleHint,
    apiCall,
    runAction,
    setHint,
    ...rest
  } = useSolitaireGameBase<
    Awaited<ReturnType<typeof eightoffApi.exec>>,
    Parameters<typeof eightoffApi.exec>,
    EightOffHint
  >(eightoffApi.exec, {
    onClearSelection,
    hintApi: () => eightoffApi.exec('hint'),
  });

  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  // Bumped every time a hint is requested so the page can announce the result
  // (a move, or "no moves available") even when two requests yield the same
  // hint value — a null hint after a request is indistinguishable from the
  // initial null without this signal.
  const [hintNonce, setHintNonce] = useState(0);
  // Depend on the stable baseHandleHint reference, not the base result object
  // (which is re-created whenever state/loading change), so this callback stays stable.
  const handleHint = useCallback(async () => {
    await baseHandleHint();
    setHintNonce((n) => n + 1);
  }, [baseHandleHint]);

  const handleSelectSource = useCallback((zone: EightOffMoveZone) => {
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
    (zone: EightOffMoveZone) => {
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
    hintNonce,
    handleHint,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
  };
}
