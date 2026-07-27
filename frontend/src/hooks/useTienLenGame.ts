import { useCallback, useEffect, useRef, useState } from 'react';
import { tienlenApi } from '../api/gameApi';
import type { Card, TienLenAction, TienLenConfigInput, TienLenResponse } from '../types/card';
import { buildHumanActionState, buildReplayStates } from '../utils/replayBuilder';
import { runReplay, shouldSkipReplay } from './gameReplay';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useIsMounted } from './useIsMounted';

// 0 = Normal, matching the server's DefaultTienLenConfig() so the settings
// panel's initial value agrees with the on-mount reset (which sends no config).
const defaultConfigInput: TienLenConfigInput = {
  cpuDifficulty: 0,
};

interface TienLenCountsCtx {
  counts: number[];
}

interface TienLenCtx extends TienLenCountsCtx {
  currentTableCards: Card[];
}

const tienLenInitContext = (fs: TienLenResponse): TienLenCtx => ({
  counts: fs.players.map((p) => p.cardCount),
  currentTableCards: fs.humanAction?.playedCards?.length ? fs.humanAction.playedCards : ([] as Card[]),
});

const tienLenReverseAction = (ctx: TienLenCountsCtx, a: TienLenAction) => {
  ctx.counts[a.playerIdx] += a.playedCards?.length ?? 0;
};

const tienLenApplyAction = (ctx: TienLenCtx, a: TienLenAction) => {
  ctx.counts[a.playerIdx] = Math.max(0, ctx.counts[a.playerIdx] - (a.playedCards?.length ?? 0));
  if (a.playedCards?.length) {
    ctx.currentTableCards = a.playedCards;
  }
};

function buildTienLenHumanActionState(finalState: TienLenResponse): TienLenResponse | null {
  if (!finalState.humanAction || (finalState.cpuActions?.length ?? 0) === 0) return null;
  const [firstCpuAction] = finalState.cpuActions;
  const ha = finalState.humanAction;
  return buildHumanActionState({
    actions: finalState.cpuActions,
    finalState,
    initContext: (fs) => ({ counts: fs.players.map((p) => p.cardCount) }),
    reverseAction: tienLenReverseAction,
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

function buildTienLenReplayStates(finalState: TienLenResponse): TienLenResponse[] {
  return buildReplayStates({
    actions: finalState.cpuActions ?? [],
    finalState,
    initContext: tienLenInitContext,
    reverseAction: tienLenReverseAction,
    applyAction: tienLenApplyAction,
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

/** Hook that manages Tien Len game state, card selection, and CPU replay. */
export function useTienLenGame() {
  const isMounted = useIsMounted();

  const { selected: selectedIndices, toggle: toggleCardSelection, clear: clearSelection } = useCardSelection();
  const [configInput, setConfigInput] = useState<TienLenConfigInput>(defaultConfigInput);
  const [displayState, setDisplayState] = useState<TienLenResponse | null>(null);

  const lastReplayedActionsRef = useRef<TienLenResponse['cpuActions']>(undefined);

  const onSuccess = useCallback(
    async (res: TienLenResponse) => {
      clearSelection();
      if (shouldSkipReplay(res.cpuActions ?? [], lastReplayedActionsRef, res, setDisplayState)) {
        return;
      }
      await runReplay(res, setDisplayState, {
        isMounted,
        buildReplayStates: buildTienLenReplayStates,
        buildHumanActionState: buildTienLenHumanActionState,
      });
    },
    [clearSelection, isMounted],
  );

  const { loading, error, exec: callApi, retry } = useGameApi(tienlenApi.exec, { onSuccess });

  useEffect(() => {
    callApi('reset');
  }, [callApi]);

  const handleConfigChange = useCallback((key: keyof TienLenConfigInput, value: number) => {
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
