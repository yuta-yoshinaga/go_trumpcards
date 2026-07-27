import { useCallback, useEffect, useRef, useState } from 'react';
import { type PresidentConfigInput, presidentApi } from '../api/gameApi';
import type { Card, PresidentAction, PresidentResponse } from '../types/card';
import { buildHumanActionState, buildReplayStates } from '../utils/replayBuilder';
import { runReplay, shouldSkipReplay } from './gameReplay';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useIsMounted } from './useIsMounted';

const defaultConfigInput: PresidentConfigInput = {
  revolutionEnabled: true,
  cardExchangeEnabled: true,
  passFieldFlushEnabled: true,
  cpuDifficulty: 1,
};

interface PresidentCountsCtx {
  counts: number[];
}

interface PresidentCtx extends PresidentCountsCtx {
  currentTableCards: Card[];
}

const presidentInitContext = (fs: PresidentResponse): PresidentCtx => ({
  counts: fs.players.map((p) => p.cardCount),
  currentTableCards: fs.humanAction?.playedCards?.length ? fs.humanAction.playedCards : ([] as Card[]),
});

const presidentReverseAction = (ctx: PresidentCountsCtx, a: PresidentAction) => {
  ctx.counts[a.playerIdx] += a.playedCards?.length ?? 0;
};

const presidentApplyAction = (ctx: PresidentCtx, a: PresidentAction) => {
  ctx.counts[a.playerIdx] = Math.max(0, ctx.counts[a.playerIdx] - (a.playedCards?.length ?? 0));
  if (a.playedCards?.length) {
    ctx.currentTableCards = a.playedCards;
  }
};

function buildPresidentHumanActionState(finalState: PresidentResponse): PresidentResponse | null {
  if (!finalState.humanAction || (finalState.cpuActions?.length ?? 0) === 0) return null;
  const [firstCpuAction] = finalState.cpuActions;
  const ha = finalState.humanAction;
  return buildHumanActionState({
    actions: finalState.cpuActions,
    finalState,
    initContext: (fs) => ({ counts: fs.players.map((p) => p.cardCount) }),
    reverseAction: presidentReverseAction,
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

function buildPresidentReplayStates(finalState: PresidentResponse): PresidentResponse[] {
  return buildReplayStates({
    actions: finalState.cpuActions ?? [],
    finalState,
    initContext: presidentInitContext,
    reverseAction: presidentReverseAction,
    applyAction: presidentApplyAction,
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

/** Hook that manages President game state, card selection, and CPU replay. */
export function usePresidentGame() {
  const isMounted = useIsMounted();

  const { selected: selectedIndices, toggle: toggleCardSelection, clear: clearSelection } = useCardSelection();
  const [configInput, setConfigInput] = useState<PresidentConfigInput>(defaultConfigInput);
  const [displayState, setDisplayState] = useState<PresidentResponse | null>(null);

  const lastReplayedActionsRef = useRef<PresidentResponse['cpuActions']>(undefined);

  const onSuccess = useCallback(
    async (res: PresidentResponse) => {
      clearSelection();
      if (shouldSkipReplay(res.cpuActions ?? [], lastReplayedActionsRef, res, setDisplayState)) {
        return;
      }
      await runReplay(res, setDisplayState, {
        isMounted,
        buildReplayStates: buildPresidentReplayStates,
        buildHumanActionState: buildPresidentHumanActionState,
      });
    },
    [clearSelection, isMounted],
  );

  const { loading, error, exec: callApi, retry } = useGameApi(presidentApi.exec, { onSuccess });

  useEffect(() => {
    callApi('reset');
  }, [callApi]);

  const handleConfigChange = useCallback((key: keyof PresidentConfigInput, value: boolean | number) => {
    setConfigInput((prev: PresidentConfigInput) => ({ ...prev, [key]: value }));
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
