import { useCallback, useEffect, useState } from 'react';
import { type KlondikeMoveZone, klondikeApi } from '../api/gameApi';
import type { KlondikeHint } from '../types/card';
import { useGameApi } from './useGameApi';

export function useKlondikeGame() {
  const { state, loading, error, exec: rawExec } = useGameApi(klondikeApi.exec);
  const [selectedSource, setSelectedSource] = useState<KlondikeMoveZone | null>(null);
  const [hint, setHint] = useState<KlondikeHint | null>(null);

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleDraw = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('draw');
  }, [exec]);

  const handleReset = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('reset');
  }, [exec]);

  const handleGiveUp = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('giveup');
  }, [exec]);

  const handleHint = useCallback(async () => {
    const res = await klondikeApi.exec('hint');
    setHint(res.hint ?? null);
  }, []);

  const handleAutoComplete = useCallback(() => {
    setSelectedSource(null);
    setHint(null);
    exec('autocomplete');
  }, [exec]);

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
      exec('move', selectedSource, zone);
      setSelectedSource(null);
    },
    [selectedSource, exec],
  );

  return {
    state,
    loading,
    error,
    exec,
    selectedSource,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleSelectSource,
    handleSelectTarget,
  };
}
