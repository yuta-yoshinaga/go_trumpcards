import { useCallback, useEffect, useRef, useState } from 'react';
import { oldmaidApi } from '../api/gameApi';
import type { Card, CpuAction, OldMaidResponse } from '../types/card';
import { buildHumanActionState, buildReplayStates } from '../utils/replayBuilder';
import { REPLAY_DELAY_MS, runReplay, shouldSkipReplay } from './gameReplay';
import { useGameApi } from './useGameApi';
import { useIsMounted } from './useIsMounted';

/** Old Maid game mode constants (Normal or JijiNuki). */
export const OldMaidMode = {
  Normal: 0,
  JijiNuki: 1,
} as const;

interface OldMaidCtx {
  counts: number[];
}

const oldMaidInitContext = (fs: OldMaidResponse): OldMaidCtx => ({
  counts: fs.players.map((p) => p.cardCount),
});

const oldMaidReverseAction = (ctx: OldMaidCtx, a: CpuAction) => {
  ctx.counts[a.drawPlayerIdx] = ctx.counts[a.drawPlayerIdx] + 2 * a.discardedPairs - 1;
  ctx.counts[a.drawFromIdx] = ctx.counts[a.drawFromIdx] + 1;
};

const oldMaidApplyAction = (ctx: OldMaidCtx, a: CpuAction) => {
  ctx.counts[a.drawFromIdx] -= 1;
  ctx.counts[a.drawPlayerIdx] += 1 - 2 * a.discardedPairs;
};

function buildOldMaidReplayStates(finalState: OldMaidResponse): OldMaidResponse[] {
  const actions = finalState.cpuActions;
  return buildReplayStates({
    actions,
    finalState,
    initContext: oldMaidInitContext,
    reverseAction: oldMaidReverseAction,
    applyAction: oldMaidApplyAction,
    buildState: (fs, ctx, a, processedActions, isLast) => ({
      ...fs,
      players: fs.players.map((p, idx) => ({
        ...p,
        cardCount: Math.max(0, ctx.counts[idx]),
        isFinished: ctx.counts[idx] <= 0,
      })),
      currentTurn: a.drawPlayerIdx,
      hasDrawn: true,
      lastDrawPlayerIdx: a.drawPlayerIdx,
      lastDrawFromIdx: a.drawFromIdx,
      lastDrawCard: a.drawnCard,
      lastDiscardedPairs: a.discardedPairs,
      lastDiscardedCards: a.discardedCards ?? [],
      cpuActions: processedActions,
      gameEndFlag: isLast ? fs.gameEndFlag : false,
      message: isLast ? fs.message : '',
      nextDrawTargetIdx: isLast ? fs.nextDrawTargetIdx : actions[processedActions.length].drawFromIdx,
    }),
  });
}

function buildHumanDrawState(finalState: OldMaidResponse): OldMaidResponse | null {
  const ha = finalState.humanAction;
  if (!ha) return null;

  const [firstCpuAction] = finalState.cpuActions;

  return buildHumanActionState({
    actions: finalState.cpuActions,
    finalState,
    initContext: oldMaidInitContext,
    reverseAction: oldMaidReverseAction,
    buildState: (fs, ctx) => ({
      ...fs,
      players: fs.players.map((p, idx) => ({
        ...p,
        cardCount: Math.max(0, ctx.counts[idx]),
        isFinished: ctx.counts[idx] <= 0,
      })),
      hasDrawn: true,
      lastDrawPlayerIdx: ha.drawPlayerIdx,
      lastDrawFromIdx: ha.drawFromIdx,
      lastDrawCard: ha.drawnCard,
      lastDiscardedPairs: ha.discardedPairs,
      lastDiscardedCards: ha.discardedCards ?? [],
      cpuActions: [],
      ...(firstCpuAction && {
        currentTurn: firstCpuAction.drawPlayerIdx,
        gameEndFlag: false,
        message: '',
        nextDrawTargetIdx: firstCpuAction.drawFromIdx,
      }),
    }),
  });
}

/** Hook that manages Old Maid game state, setup, CPU replay, and card reveal. */
export function useOldMaidGame() {
  const isMounted = useIsMounted();

  const [displayState, setDisplayState] = useState<OldMaidResponse | null>(null);
  const [setupMode, setSetupMode] = useState<number>(OldMaidMode.Normal);
  const [setupStrategy, setSetupStrategy] = useState(false);
  const [setupMemoryAI, setSetupMemoryAI] = useState(false);
  const [setupHesitation, setSetupHesitation] = useState(false);
  const [setupMetaAI, setSetupMetaAI] = useState(false);
  const [gameSettings, setGameSettings] = useState<{
    mode: number;
    cpuPlacementStrategy: boolean;
    cpuMemoryAI: boolean;
    cpuHesitationEnabled: boolean;
    cpuMetaAI: boolean;
  } | null>(null);
  const [suspectPins, setSuspectPins] = useState<Set<number>>(new Set());
  const [shakeKey, setShakeKey] = useState(0);
  const [revealedCard, setRevealedCard] = useState<Card | null>(null);
  const revealTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Card reveal suspense: show card-back for 600ms, then flip to actual card
  useEffect(() => {
    const card = displayState?.lastDrawCard;
    if (!card) {
      setRevealedCard(null);
      return;
    }
    setRevealedCard(null);
    revealTimerRef.current = setTimeout(() => {
      setRevealedCard(card);
      revealTimerRef.current = null;
      if (card.design === 'JOKER') {
        setShakeKey((k) => k + 1);
      }
    }, 600);
    return () => {
      if (revealTimerRef.current !== null) {
        clearTimeout(revealTimerRef.current);
        revealTimerRef.current = null;
      }
    };
  }, [displayState?.lastDrawCard]);

  const lastReplayedActionsRef = useRef<OldMaidResponse['cpuActions']>(undefined);

  const onSuccess = useCallback(
    async (res: OldMaidResponse) => {
      if (shouldSkipReplay(res.cpuActions ?? [], lastReplayedActionsRef, res, setDisplayState)) {
        return;
      }
      await runReplay(res, setDisplayState, {
        isMounted,
        buildReplayStates: buildOldMaidReplayStates,
        buildHumanActionState: buildHumanDrawState,
        // hesitationMs is 0 when disabled; || falls back to REPLAY_DELAY_MS (min enabled value is 300ms)
        getActionDelay: (state, i) => state.cpuActions[i]?.hesitationMs || REPLAY_DELAY_MS,
      });
    },
    [isMounted],
  );

  const { loading, error, exec: gameExec, retry } = useGameApi(oldmaidApi.exec, { onSuccess });

  const handleStart = useCallback(() => {
    const settings = {
      mode: setupMode,
      cpuPlacementStrategy: setupStrategy,
      cpuMemoryAI: setupMemoryAI,
      cpuHesitationEnabled: setupHesitation,
      cpuMetaAI: setupMetaAI,
    };
    setGameSettings(settings);
    setSuspectPins(new Set());
    gameExec(
      'reset',
      undefined,
      settings.mode,
      settings.cpuPlacementStrategy,
      undefined,
      settings.cpuMemoryAI,
      settings.cpuHesitationEnabled,
      settings.cpuMetaAI,
    );
  }, [gameExec, setupMode, setupStrategy, setupMemoryAI, setupHesitation, setupMetaAI]);

  // Auto-start with default settings on first mount
  const autoStartedRef = useRef(false);
  useEffect(() => {
    if (!autoStartedRef.current) {
      autoStartedRef.current = true;
      handleStart();
    }
  }, [handleStart]);

  const handleReset = useCallback(() => {
    setSuspectPins(new Set());
    if (gameSettings) {
      gameExec(
        'reset',
        undefined,
        gameSettings.mode,
        gameSettings.cpuPlacementStrategy,
        undefined,
        gameSettings.cpuMemoryAI,
        gameSettings.cpuHesitationEnabled,
        gameSettings.cpuMetaAI,
      );
    }
  }, [gameExec, gameSettings]);

  const handleReorder = useCallback(
    (indices: number[]) => {
      gameExec('reorder', undefined, undefined, undefined, indices);
    },
    [gameExec],
  );

  return {
    displayState,
    setupMode,
    setupStrategy,
    setupMemoryAI,
    setupHesitation,
    setupMetaAI,
    gameSettings,
    suspectPins,
    setSuspectPins,
    shakeKey,
    revealedCard,
    loading,
    error,
    gameExec,
    handleStart,
    handleReset,
    handleReorder,
    setSetupMode,
    setSetupStrategy,
    setSetupMemoryAI,
    setSetupHesitation,
    setSetupMetaAI,
    retry,
  };
}
