import { useCallback, useState } from 'react';
import { type PerseveranceMoveZone, perseveranceApi } from '../api/gameApi';
import type { PerseveranceHint } from '../types/card';
import { useSolitaireGameBase } from './useSolitaireGameBase';

/** Hook that manages Perseverance game state, source selection, hints, and moves. */
export function usePerseveranceGame() {
  const [selectedSource, setSelectedSource] = useState<PerseveranceMoveZone | null>(null);
  const onClearSelection = useCallback(() => setSelectedSource(null), []);
  const {
    apiCall: exec,
    runAction,
    state,
    loading,
    error,
    retry,
    hint,
    hintError,
    isAutoCompleting,
    setHint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
  } = useSolitaireGameBase<
    Awaited<ReturnType<typeof perseveranceApi.exec>>,
    Parameters<typeof perseveranceApi.exec>,
    PerseveranceHint
  >(perseveranceApi.exec, {
    onClearSelection,
    hintApi: () => perseveranceApi.exec('hint'),
    selectHint: (res) => res.hint,
  });

  // **リディールは Perseverance だけの救済手段** (クローン元の Baker's Dozen には無い)。
  // 選択とヒントを落としてから投げる ── 盤が総入れ替えになるので、残した索引は
  // 全部別の札を指す。
  const handleRedeal = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('redeal');
  }, [exec, setHint]);

  /** Undo N moves at once to escape a stalemate. */
  const handleUndoEscape = useCallback((n: number) => runAction('undo_n', undefined, undefined, n), [runAction]);

  const handleSelectSource = useCallback((zone: PerseveranceMoveZone) => {
    setSelectedSource((prev) => {
      if (prev && prev.zone === zone.zone && prev.col === zone.col && prev.cardIndex === zone.cardIndex) {
        return null;
      }
      return zone;
    });
  }, []);

  const handleSelectTarget = useCallback(
    (zone: PerseveranceMoveZone) => {
      if (!selectedSource) return;
      setHint(null);
      exec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, exec, setHint],
  );

  return {
    state,
    loading,
    error,
    hintError,
    exec,
    selectedSource,
    hint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleRedeal,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
    retry,
  };
}
