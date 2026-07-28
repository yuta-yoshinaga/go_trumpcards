import { useCallback, useEffect, useRef, useState } from 'react';
import { bigtwoApi } from '../api/gameApi';
import type { BigTwoAction, BigTwoConfigInput, BigTwoResponse, Card } from '../types/card';
import { buildHumanActionState, buildReplayStates } from '../utils/replayBuilder';
import { runReplay, shouldSkipReplay } from './gameReplay';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useIsMounted } from './useIsMounted';

const defaultConfigInput: BigTwoConfigInput = {
  cpuDifficulty: 1,
};

interface BigTwoCountsCtx {
  counts: number[];
}

interface BigTwoCtx extends BigTwoCountsCtx {
  currentTableCards: Card[];
}

const bigTwoInitContext = (fs: BigTwoResponse): BigTwoCtx => ({
  counts: fs.players.map((p) => p.cardCount),
  currentTableCards: fs.humanAction?.playedCards?.length ? fs.humanAction.playedCards : ([] as Card[]),
});

const bigTwoReverseAction = (ctx: BigTwoCountsCtx, a: BigTwoAction) => {
  ctx.counts[a.playerIdx] += a.playedCards?.length ?? 0;
};

const bigTwoApplyAction = (ctx: BigTwoCtx, a: BigTwoAction) => {
  ctx.counts[a.playerIdx] = Math.max(0, ctx.counts[a.playerIdx] - (a.playedCards?.length ?? 0));
  if (a.playedCards?.length) {
    ctx.currentTableCards = a.playedCards;
  }
};

function buildBigTwoHumanActionState(finalState: BigTwoResponse): BigTwoResponse | null {
  if (!finalState.humanAction || (finalState.cpuActions?.length ?? 0) === 0) return null;
  const [firstCpuAction] = finalState.cpuActions;
  const ha = finalState.humanAction;
  return buildHumanActionState({
    actions: finalState.cpuActions,
    finalState,
    initContext: (fs) => ({ counts: fs.players.map((p) => p.cardCount) }),
    reverseAction: bigTwoReverseAction,
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

function buildBigTwoReplayStates(finalState: BigTwoResponse): BigTwoResponse[] {
  return buildReplayStates({
    actions: finalState.cpuActions ?? [],
    finalState,
    initContext: bigTwoInitContext,
    reverseAction: bigTwoReverseAction,
    applyAction: bigTwoApplyAction,
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

/** Hook that manages Big Two game state, card selection, and CPU replay. */
export function useBigTwoGame() {
  const isMounted = useIsMounted();

  const { selected: selectedIndices, toggle: toggleCardSelection, clear: clearSelection } = useCardSelection();
  const [configInput, setConfigInput] = useState<BigTwoConfigInput>(defaultConfigInput);
  const [displayState, setDisplayState] = useState<BigTwoResponse | null>(null);

  const lastReplayedActionsRef = useRef<BigTwoResponse['cpuActions']>(undefined);

  const onSuccess = useCallback(
    async (res: BigTwoResponse) => {
      clearSelection();
      if (shouldSkipReplay(res.cpuActions ?? [], lastReplayedActionsRef, res, setDisplayState)) {
        return;
      }
      await runReplay(res, setDisplayState, {
        isMounted,
        buildReplayStates: buildBigTwoReplayStates,
        buildHumanActionState: buildBigTwoHumanActionState,
      });
    },
    [clearSelection, isMounted],
  );

  const { loading, error, exec: callApi, retry } = useGameApi(bigtwoApi.exec, { onSuccess });

  useEffect(() => {
    callApi('reset');
  }, [callApi]);

  const handleConfigChange = useCallback((key: keyof BigTwoConfigInput, value: number) => {
    setConfigInput((prev) => ({ ...prev, [key]: value }));
  }, []);

  const handlePlay = useCallback(() => {
    callApi(
      'play',
      [...selectedIndices].sort((a, b) => a - b),
    );
  }, [callApi, selectedIndices]);

  const handlePass = useCallback(() => {
    callApi('play', []);
  }, [callApi]);

  const handleResetWithConfig = useCallback(() => {
    callApi('reset', undefined, configInput);
  }, [callApi, configInput]);

  return {
    state: displayState,
    loading,
    error,
    callApi,
    selectedIndices,
    toggleCardSelection,
    clearSelection,
    configInput,
    handleConfigChange,
    handlePlay,
    handlePass,
    handleResetWithConfig,
    retry,
  };
}
