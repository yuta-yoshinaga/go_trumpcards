import { useCallback, useEffect, useRef, useState } from 'react';
import { oldmaidApi } from '../api/gameApi';
import type { Card, OldMaidResponse } from '../types/card';
import { useGameApi } from './useGameApi';

const REPLAY_DELAY_MS = 800;

export const OldMaidMode = {
  Normal: 0,
  JijiNuki: 1,
} as const;

const delay = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms));

/** Compute intermediate player card counts by reversing all CPU actions from the final state,
 *  then replay forward. Returns one OldMaidResponse per CPU action (state after each action). */
function buildReplayStates(finalState: OldMaidResponse): OldMaidResponse[] {
  const actions = finalState.cpuActions;

  // Work backwards to get counts before all CPU actions
  const counts = finalState.players.map((p) => p.cardCount);
  for (let i = actions.length - 1; i >= 0; i--) {
    const a = actions[i];
    counts[a.drawPlayerIdx] = counts[a.drawPlayerIdx] + 2 * a.discardedPairs - 1;
    counts[a.drawFromIdx] = counts[a.drawFromIdx] + 1;
  }

  // Play forward, building a display state after each CPU action
  const states: OldMaidResponse[] = [];
  for (let i = 0; i < actions.length; i++) {
    const a = actions[i];
    counts[a.drawFromIdx] -= 1;
    counts[a.drawPlayerIdx] += 1 - 2 * a.discardedPairs;

    const isLastAction = i === actions.length - 1;
    // Intermediate replay states carry the full final drawHistory intentionally:
    // history entries don't reveal card contents, so showing all entries is safe.
    states.push({
      ...finalState,
      players: finalState.players.map((p, idx) => ({
        ...p,
        cardCount: Math.max(0, counts[idx]),
        isFinished: counts[idx] <= 0,
      })),
      currentTurn: a.drawPlayerIdx,
      hasDrawn: true,
      lastDrawPlayerIdx: a.drawPlayerIdx,
      lastDrawFromIdx: a.drawFromIdx,
      lastDrawCard: a.drawnCard,
      lastDiscardedPairs: a.discardedPairs,
      lastDiscardedCards: a.discardedCards ?? [],
      cpuActions: actions.slice(0, i + 1),
      gameEndFlag: isLastAction ? finalState.gameEndFlag : false,
      message: isLastAction ? finalState.message : '',
      nextDrawTargetIdx: isLastAction ? finalState.nextDrawTargetIdx : actions[i + 1].drawFromIdx,
    });
  }
  return states;
}

/** Build the display state right after human's draw, before any CPU actions. */
function buildHumanDrawState(finalState: OldMaidResponse): OldMaidResponse | null {
  const ha = finalState.humanAction;
  if (!ha) return null;

  const counts = finalState.players.map((p) => p.cardCount);
  for (let i = finalState.cpuActions.length - 1; i >= 0; i--) {
    const a = finalState.cpuActions[i];
    counts[a.drawPlayerIdx] = counts[a.drawPlayerIdx] + 2 * a.discardedPairs - 1;
    counts[a.drawFromIdx] = counts[a.drawFromIdx] + 1;
  }

  const [firstCpuAction] = finalState.cpuActions;

  return {
    ...finalState,
    players: finalState.players.map((p, idx) => ({
      ...p,
      cardCount: Math.max(0, counts[idx]),
      isFinished: counts[idx] <= 0,
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
  };
}

export function useOldMaidGame() {
  const [displayState, setDisplayState] = useState<OldMaidResponse | null>(null);
  const [setupMode, setSetupMode] = useState<number>(OldMaidMode.Normal);
  const [setupStrategy, setSetupStrategy] = useState(false);
  const [setupMemoryAI, setSetupMemoryAI] = useState(false);
  const [gameSettings, setGameSettings] = useState<{
    mode: number;
    cpuPlacementStrategy: boolean;
    cpuMemoryAI: boolean;
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

  const onSuccess = useCallback(async (res: OldMaidResponse) => {
    const humanDrawState = buildHumanDrawState(res);
    if (humanDrawState) {
      setDisplayState(humanDrawState);
      await delay(REPLAY_DELAY_MS);
    }
    const replayStates = buildReplayStates(res);
    if (replayStates.length === 0) {
      setDisplayState(res);
      return;
    }
    for (const step of replayStates) {
      setDisplayState(step);
      await delay(REPLAY_DELAY_MS);
    }
    setDisplayState(res);
  }, []);

  const { loading, error, exec } = useGameApi(oldmaidApi.exec, { onSuccess });

  const handleStart = useCallback(() => {
    const settings = { mode: setupMode, cpuPlacementStrategy: setupStrategy, cpuMemoryAI: setupMemoryAI };
    setGameSettings(settings);
    setSuspectPins(new Set());
    exec('reset', undefined, settings.mode, settings.cpuPlacementStrategy, undefined, settings.cpuMemoryAI);
  }, [exec, setupMode, setupStrategy, setupMemoryAI]);

  const handleReset = useCallback(() => {
    setSuspectPins(new Set());
    if (gameSettings) {
      exec(
        'reset',
        undefined,
        gameSettings.mode,
        gameSettings.cpuPlacementStrategy,
        undefined,
        gameSettings.cpuMemoryAI,
      );
    }
  }, [exec, gameSettings]);

  const handleReorder = useCallback(
    (indices: number[]) => {
      exec('reorder', undefined, undefined, undefined, indices);
    },
    [exec],
  );

  return {
    displayState,
    setupMode,
    setupStrategy,
    setupMemoryAI,
    gameSettings,
    suspectPins,
    setSuspectPins,
    shakeKey,
    revealedCard,
    loading,
    error,
    exec,
    handleStart,
    handleReset,
    handleReorder,
    setSetupMode,
    setSetupStrategy,
    setSetupMemoryAI,
    setGameSettings,
  };
}
