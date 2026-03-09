import { useCallback, useEffect, useState } from 'react';
import { daifugoApi } from '../api/gameApi';
import type { DaifugoConfigInput } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';

const defaultConfigInput: DaifugoConfigInput = {
  jokerCount: 2,
  eightCutEnabled: true,
  suitLockMode: 2,
  elevenBackEnabled: true,
  sequenceEnabled: true,
  cardExchangeEnabled: true,
  fiveSkipEnabled: false,
  fiveSkipCount: 1,
  sevenPassEnabled: false,
  tenDiscardEnabled: false,
  spadeThreeEnabled: false,
  capitalFallEnabled: false,
  nineReverseEnabled: false,
  coupDetatEnabled: false,
  numberLockEnabled: false,
  sandstormEnabled: false,
  emperorEnabled: false,
  sequenceRevolutionEnabled: false,
  illegalFinishEnabled: false,
  queenBomberEnabled: false,
  cpuDifficulty: 0,
};

export function useDaifugoGame() {
  const {
    selected: selectedIndices,
    toggle: toggleCardSelection,
    clear: clearSelection,
    setSelected: setSelectedIndices,
  } = useCardSelection();
  const [configInput, setConfigInput] = useState<DaifugoConfigInput>(defaultConfigInput);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec } = useGameApi(daifugoApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleDragCard = useCallback(
    (idx: number) => {
      setSelectedIndices((prev) => (prev.includes(idx) ? prev : [...prev, idx]));
    },
    [setSelectedIndices],
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      const draggedIdx = parseInt(e.dataTransfer.getData('cardIndex'), 10);
      if (Number.isNaN(draggedIdx)) {
        return;
      }
      const toPlay = selectedIndices.includes(draggedIdx) ? selectedIndices : [draggedIdx];
      exec(
        'play',
        [...toPlay].sort((a, b) => a - b),
      );
    },
    [exec, selectedIndices],
  );

  const handleConfigChange = useCallback((key: keyof DaifugoConfigInput, value: boolean | number) => {
    setConfigInput((prev) => ({ ...prev, [key]: value }));
  }, []);

  return {
    state,
    loading,
    error,
    exec,
    selectedIndices,
    toggleCardSelection,
    configInput,
    handleDragCard,
    handleDrop,
    handleConfigChange,
  };
}
