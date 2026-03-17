import { useCallback, useEffect, useRef, useState } from 'react';
import { sevensApi } from '../api/gameApi';
import type { SevensAction, SevensResponse } from '../types/card';
import { buildReplayStates } from '../utils/replayBuilder';
import { runReplay, shouldSkipReplay } from './gameReplay';
import { useGameApi } from './useGameApi';

function computeTableMinVals(tablePlaced: number[]): number[] {
  const result = [0, 0, 0, 0, 0];
  for (let suit = 1; suit <= 4; suit++) {
    for (let v = 1; v <= 13; v++) {
      if (tablePlaced[suit] & (1 << v)) {
        result[suit] = v;
        break;
      }
    }
  }
  return result;
}

function computeTableMaxVals(tablePlaced: number[]): number[] {
  const result = [0, 0, 0, 0, 0];
  for (let suit = 1; suit <= 4; suit++) {
    for (let v = 13; v >= 1; v--) {
      if (tablePlaced[suit] & (1 << v)) {
        result[suit] = v;
        break;
      }
    }
  }
  return result;
}

const isCardPlay = (a: SevensAction): boolean =>
  !a.forcedPass && a.playedCard !== null && a.targetSuit > 0 && a.targetValue > 0;

function buildSevensReplayStates(finalState: SevensResponse): SevensResponse[] {
  return buildReplayStates({
    actions: finalState.cpuActions ?? [],
    finalState,
    initContext: (fs) => ({
      counts: fs.players.map((p) => p.cardCount),
      currentPlaced: [...fs.tablePlaced],
    }),
    reverseAction: (ctx, a) => {
      if (isCardPlay(a)) {
        ctx.counts[a.playerIdx] += 1;
        ctx.currentPlaced[a.targetSuit] &= ~(1 << a.targetValue);
      }
    },
    applyAction: (ctx, a) => {
      if (isCardPlay(a)) {
        ctx.counts[a.playerIdx] = Math.max(0, ctx.counts[a.playerIdx] - 1);
        ctx.currentPlaced[a.targetSuit] |= 1 << a.targetValue;
      }
    },
    buildState: (fs, ctx, a, processedActions, isLast) => ({
      ...fs,
      players: fs.players.map((p, idx) => ({
        ...p,
        cardCount: Math.max(0, ctx.counts[idx]),
      })),
      currentTurn: a.playerIdx,
      tablePlaced: [...ctx.currentPlaced],
      tableMinVals: computeTableMinVals(ctx.currentPlaced),
      tableMaxVals: computeTableMaxVals(ctx.currentPlaced),
      cpuActions: processedActions,
      gameEndFlag: isLast ? fs.gameEndFlag : false,
      message: isLast ? fs.message : '',
    }),
  });
}

export function useSevensGame() {
  const [jokerCardIdx, setJokerCardIdx] = useState<number | null>(null);
  const [cfgTunnel, setCfgTunnel] = useState(false);
  const [cfgTunnelSkipWidth, setCfgTunnelSkipWidth] = useState(0);
  const [cfgJokerCount, setCfgJokerCount] = useState(0);
  const [cfgCpuStrategy, setCfgCpuStrategy] = useState(0);
  const [cfgMaxPasses, setCfgMaxPasses] = useState(5);
  const [cfgNoJokerFinish, setCfgNoJokerFinish] = useState(false);
  const [cfgJokerReclaim, setCfgJokerReclaim] = useState(false);
  const [cfgEndStop, setCfgEndStop] = useState(false);
  const [cfgJokerConsBan, setCfgJokerConsBan] = useState(false);
  const [displayState, setDisplayState] = useState<SevensResponse | null>(null);

  const lastReplayedActionsRef = useRef<SevensResponse['cpuActions']>(undefined);

  const onSuccess = useCallback(async (res: SevensResponse) => {
    setJokerCardIdx(null);
    setCfgTunnel(res.config.tunnelEnabled);
    setCfgTunnelSkipWidth(res.config.tunnelSkipWidth);
    setCfgJokerCount(res.config.jokerCount);
    setCfgCpuStrategy(res.config.cpuStrategy);
    setCfgMaxPasses(res.config.maxPasses);
    setCfgNoJokerFinish(res.config.noJokerFinish);
    setCfgJokerReclaim(res.config.jokerReclaimEnabled);
    setCfgEndStop(res.config.endStopEnabled);
    setCfgJokerConsBan(res.config.jokerConsecutiveBanned);
    if (shouldSkipReplay(res.cpuActions ?? [], lastReplayedActionsRef, res, setDisplayState)) {
      return;
    }
    await runReplay(res, setDisplayState, {
      buildReplayStates: buildSevensReplayStates,
    });
  }, []);

  const { loading, error, exec } = useGameApi(sevensApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleCardPlay = useCallback(
    (idx: number) => {
      const humanPlayer = displayState?.players.find((p) => p.isHuman);
      const card = humanPlayer?.cards?.[idx];
      if (card?.design === 'JOKER') {
        setJokerCardIdx(idx);
      } else {
        exec('play', idx);
      }
    },
    [displayState, exec],
  );

  const handleJokerPlace = useCallback(
    (suit: number, value: number) => {
      exec('joker', jokerCardIdx as number, suit, value);
    },
    [exec, jokerCardIdx],
  );

  return {
    state: displayState,
    loading,
    error,
    exec,
    jokerCardIdx,
    setJokerCardIdx,
    cfgTunnel,
    setCfgTunnel,
    cfgTunnelSkipWidth,
    setCfgTunnelSkipWidth,
    cfgJokerCount,
    setCfgJokerCount,
    cfgCpuStrategy,
    setCfgCpuStrategy,
    cfgMaxPasses,
    setCfgMaxPasses,
    cfgNoJokerFinish,
    setCfgNoJokerFinish,
    cfgJokerReclaim,
    setCfgJokerReclaim,
    cfgEndStop,
    setCfgEndStop,
    cfgJokerConsBan,
    setCfgJokerConsBan,
    handleCardPlay,
    handleJokerPlace,
  };
}
