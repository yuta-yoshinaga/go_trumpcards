import { useCallback, useEffect, useRef, useState } from 'react';
import { daifugoApi } from '../api/gameApi';
import type { Card, DaifugoAction, DaifugoConfigInput, DaifugoResponse } from '../types/card';
import { buildHumanActionState, buildReplayStates } from '../utils/replayBuilder';
import { runReplay, shouldSkipReplay } from './gameReplay';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useIsMounted } from './useIsMounted';

const defaultConfigInput: DaifugoConfigInput = {
  jokerCount: 2,
  eightCutEnabled: true,
  suitLockMode: 2,
  elevenBackEnabled: true,
  sequenceEnabled: true,
  cardExchangeEnabled: true,
  blindExchangeEnabled: false,
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
  sequenceLockEnabled: false,
  illegalFinishEnabled: false,
  queenBomberEnabled: false,
  cpuDifficulty: 0,
};

interface DaifugoCountsCtx {
  counts: number[];
}

interface DaifugoCtx extends DaifugoCountsCtx {
  currentTableCards: Card[];
}

const daifugoInitContext = (fs: DaifugoResponse): DaifugoCtx => ({
  counts: fs.players.map((p) => p.cardCount),
  currentTableCards: fs.humanAction?.playedCards?.length ? fs.humanAction.playedCards : ([] as Card[]),
});

const daifugoReverseAction = (ctx: DaifugoCountsCtx, a: DaifugoAction) => {
  ctx.counts[a.playerIdx] += a.playedCards?.length ?? 0;
};

const daifugoApplyAction = (ctx: DaifugoCtx, a: DaifugoAction) => {
  ctx.counts[a.playerIdx] = Math.max(0, ctx.counts[a.playerIdx] - (a.playedCards?.length ?? 0));
  if (a.playedCards?.length) {
    ctx.currentTableCards = a.playedCards;
  }
};

function buildDaifugoHumanActionState(finalState: DaifugoResponse): DaifugoResponse | null {
  if (!finalState.humanAction || (finalState.cpuActions?.length ?? 0) === 0) return null;

  const [firstCpuAction] = finalState.cpuActions;
  const ha = finalState.humanAction;
  return buildHumanActionState({
    actions: finalState.cpuActions,
    finalState,
    initContext: (fs) => ({ counts: fs.players.map((p) => p.cardCount) }),
    reverseAction: daifugoReverseAction,
    buildState: (fs, ctx) => ({
      ...fs,
      players: fs.players.map((p, idx) => ({
        ...p,
        cardCount: Math.max(0, ctx.counts[idx]),
      })),
      currentTurn: firstCpuAction.playerIdx,
      tableCards: ha.playedCards?.length ? ha.playedCards : fs.tableCards,
      cpuActions: [],
      gameEndFlag: false,
      message: '',
    }),
  });
}

function buildDaifugoReplayStates(finalState: DaifugoResponse): DaifugoResponse[] {
  return buildReplayStates({
    actions: finalState.cpuActions ?? [],
    finalState,
    initContext: daifugoInitContext,
    reverseAction: daifugoReverseAction,
    applyAction: daifugoApplyAction,
    buildState: (fs, ctx, a, processedActions, isLast) => ({
      ...fs,
      players: fs.players.map((p, idx) => ({
        ...p,
        cardCount: Math.max(0, ctx.counts[idx]),
      })),
      currentTurn: a.playerIdx,
      tableCards: ctx.currentTableCards,
      lastPlayPlayerIdx: a.playerIdx,
      cpuActions: processedActions,
      gameEndFlag: isLast ? fs.gameEndFlag : false,
      message: isLast ? fs.message : '',
    }),
  });
}

/** Hook that manages Daifugo game state, card selection, and CPU replay. */
export function useDaifugoGame() {
  const {
    selected: selectedIndices,
    toggle: toggleCardSelection,
    clear: clearSelection,
    setSelected: setSelectedIndices,
  } = useCardSelection();
  const isMounted = useIsMounted();

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
        isMounted,
        buildReplayStates: buildDaifugoReplayStates,
        buildHumanActionState: buildDaifugoHumanActionState,
      });
    },
    [clearSelection, isMounted],
  );

  const { loading, error, exec, retry } = useGameApi(daifugoApi.exec, { onSuccess });

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
    clearSelection,
    configInput,
    handleDragCard,
    handleDrop,
    handleConfigChange,
    retry,
  };
}
