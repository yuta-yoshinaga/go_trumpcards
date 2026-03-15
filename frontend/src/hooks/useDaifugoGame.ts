import { useCallback, useEffect, useRef, useState } from 'react';
import { daifugoApi } from '../api/gameApi';
import type { DaifugoConfigInput, DaifugoResponse } from '../types/card';
import { runReplay, shouldSkipReplay } from './gameReplay';
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

/** Build the display state right after human's action, before any CPU actions. */
function buildDaifugoHumanActionState(finalState: DaifugoResponse): DaifugoResponse | null {
  if (!finalState.humanAction || (finalState.cpuActions?.length ?? 0) === 0) return null;

  const counts = finalState.players.map((p) => p.cardCount);
  for (let i = finalState.cpuActions.length - 1; i >= 0; i--) {
    const a = finalState.cpuActions[i];
    counts[a.playerIdx] += a.playedCards?.length ?? 0;
  }

  const [firstCpuAction] = finalState.cpuActions;
  const ha = finalState.humanAction;
  return {
    ...finalState,
    players: finalState.players.map((p, idx) => ({
      ...p,
      cardCount: Math.max(0, counts[idx]),
    })),
    currentTurn: firstCpuAction.playerIdx,
    tableCards: ha.playedCards?.length ? ha.playedCards : finalState.tableCards,
    cpuActions: [],
    gameEndFlag: false,
    message: '',
  };
}

/** Compute intermediate display states, one per CPU action. */
function buildDaifugoReplayStates(finalState: DaifugoResponse): DaifugoResponse[] {
  const actions = finalState.cpuActions ?? [];
  if (actions.length === 0) return [];

  const counts = finalState.players.map((p) => p.cardCount);
  for (let i = actions.length - 1; i >= 0; i--) {
    const a = actions[i];
    counts[a.playerIdx] += a.playedCards?.length ?? 0;
  }

  let currentTableCards = finalState.humanAction?.playedCards?.length
    ? finalState.humanAction.playedCards
    : ([] as typeof finalState.tableCards);

  const states: DaifugoResponse[] = [];
  for (let i = 0; i < actions.length; i++) {
    const a = actions[i];
    counts[a.playerIdx] = Math.max(0, counts[a.playerIdx] - (a.playedCards?.length ?? 0));
    if (a.playedCards?.length) {
      currentTableCards = a.playedCards;
    }
    const isLastAction = i === actions.length - 1;
    states.push({
      ...finalState,
      players: finalState.players.map((p, idx) => ({
        ...p,
        cardCount: Math.max(0, counts[idx]),
      })),
      currentTurn: a.playerIdx,
      tableCards: currentTableCards,
      lastPlayPlayerIdx: a.playerIdx,
      cpuActions: actions.slice(0, i + 1),
      gameEndFlag: isLastAction ? finalState.gameEndFlag : false,
      message: isLastAction ? finalState.message : '',
    });
  }
  return states;
}

export function useDaifugoGame() {
  const {
    selected: selectedIndices,
    toggle: toggleCardSelection,
    clear: clearSelection,
    setSelected: setSelectedIndices,
  } = useCardSelection();
  const [configInput, setConfigInput] = useState<DaifugoConfigInput>(defaultConfigInput);
  const [displayState, setDisplayState] = useState<DaifugoResponse | null>(null);

  const lastReplayedActionsRef = useRef<DaifugoResponse['cpuActions']>(undefined);

  const onSuccess = useCallback(
    async (res: DaifugoResponse) => {
      clearSelection();
      if (shouldSkipReplay(res.cpuActions ?? [], lastReplayedActionsRef, res, setDisplayState)) {
        return;
      }
      await runReplay(res, setDisplayState, {
        buildReplayStates: buildDaifugoReplayStates,
        buildHumanActionState: buildDaifugoHumanActionState,
      });
    },
    [clearSelection],
  );

  const { loading, error, exec } = useGameApi(daifugoApi.exec, { onSuccess });

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
    state: displayState,
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
