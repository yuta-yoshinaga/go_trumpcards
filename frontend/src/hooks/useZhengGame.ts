import { useCallback, useEffect, useRef, useState } from 'react';
import { zhengApi } from '../api/gameApi';
import type { Card, ZhengAction, ZhengConfigInput, ZhengResponse } from '../types/card';
import { buildHumanActionState, buildReplayStates } from '../utils/replayBuilder';
import { runReplay, shouldSkipReplay } from './gameReplay';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';

// 0 = Normal, matching the server's DefaultZhengConfig() so the settings
// panel's initial value agrees with the on-mount reset (which sends no config).
const defaultConfigInput: ZhengConfigInput = {
  cpuDifficulty: 0,
};

interface ZhengCountsCtx {
  counts: number[];
}

interface ZhengCtx extends ZhengCountsCtx {
  currentTableCards: Card[];
}

const zhengInitContext = (fs: ZhengResponse): ZhengCtx => ({
  counts: fs.players.map((p) => p.cardCount),
  currentTableCards: fs.humanAction?.playedCards?.length ? fs.humanAction.playedCards : ([] as Card[]),
});

const zhengReverseAction = (ctx: ZhengCountsCtx, a: ZhengAction) => {
  ctx.counts[a.playerIdx] += a.playedCards?.length ?? 0;
};

const zhengApplyAction = (ctx: ZhengCtx, a: ZhengAction) => {
  ctx.counts[a.playerIdx] = Math.max(0, ctx.counts[a.playerIdx] - (a.playedCards?.length ?? 0));
  if (a.playedCards?.length) {
    ctx.currentTableCards = a.playedCards;
  }
};

function buildZhengHumanActionState(finalState: ZhengResponse): ZhengResponse | null {
  if (!finalState.humanAction || (finalState.cpuActions?.length ?? 0) === 0) return null;
  const [firstCpuAction] = finalState.cpuActions;
  const ha = finalState.humanAction;
  return buildHumanActionState({
    actions: finalState.cpuActions,
    finalState,
    initContext: (fs) => ({ counts: fs.players.map((p) => p.cardCount) }),
    reverseAction: zhengReverseAction,
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

function buildZhengReplayStates(finalState: ZhengResponse): ZhengResponse[] {
  return buildReplayStates({
    actions: finalState.cpuActions ?? [],
    finalState,
    initContext: zhengInitContext,
    reverseAction: zhengReverseAction,
    applyAction: zhengApplyAction,
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

/** Hook that manages Zheng Shangyou game state, card selection, and CPU replay. */
export function useZhengGame() {
  const { selected: selectedIndices, toggle: toggleCardSelection, clear: clearSelection } = useCardSelection();
  const [configInput, setConfigInput] = useState<ZhengConfigInput>(defaultConfigInput);
  const [displayState, setDisplayState] = useState<ZhengResponse | null>(null);

  const lastReplayedActionsRef = useRef<ZhengResponse['cpuActions']>(undefined);

  const onSuccess = useCallback(
    async (res: ZhengResponse) => {
      clearSelection();
      if (shouldSkipReplay(res.cpuActions ?? [], lastReplayedActionsRef, res, setDisplayState)) {
        return;
      }
      await runReplay(res, setDisplayState, {
        buildReplayStates: buildZhengReplayStates,
        buildHumanActionState: buildZhengHumanActionState,
      });
    },
    [clearSelection],
  );

  const { loading, error, exec: callApi, retry } = useGameApi(zhengApi.exec, { onSuccess });

  useEffect(() => {
    callApi('reset');
  }, [callApi]);

  const handleConfigChange = useCallback((key: keyof ZhengConfigInput, value: number) => {
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
