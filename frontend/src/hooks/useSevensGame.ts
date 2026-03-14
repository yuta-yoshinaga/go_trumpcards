import { useCallback, useEffect, useState } from 'react';
import { sevensApi } from '../api/gameApi';
import type { SevensResponse } from '../types/card';
import { useGameApi } from './useGameApi';
import { runReplay } from './useGameReplay';

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

/** Compute intermediate display states, one per CPU action. */
function buildSevensReplayStates(finalState: SevensResponse): SevensResponse[] {
  const actions = finalState.cpuActions ?? [];
  if (actions.length === 0) return [];

  // Work backwards to reconstruct state before all CPU actions
  const counts = finalState.players.map((p) => p.cardCount);
  const tablePlaced = [...finalState.tablePlaced];
  for (let i = actions.length - 1; i >= 0; i--) {
    const a = actions[i];
    if (!a.forcedPass && a.playedCard !== null && a.targetSuit > 0 && a.targetValue > 0) {
      counts[a.playerIdx] += 1;
      tablePlaced[a.targetSuit] &= ~(1 << a.targetValue);
    }
  }

  // Play forward, building a display state after each CPU action
  const currentPlaced = [...tablePlaced];
  const states: SevensResponse[] = [];
  for (let i = 0; i < actions.length; i++) {
    const a = actions[i];
    if (!a.forcedPass && a.playedCard !== null && a.targetSuit > 0 && a.targetValue > 0) {
      counts[a.playerIdx] = Math.max(0, counts[a.playerIdx] - 1);
      currentPlaced[a.targetSuit] |= 1 << a.targetValue;
    }
    const isLastAction = i === actions.length - 1;
    states.push({
      ...finalState,
      players: finalState.players.map((p, idx) => ({
        ...p,
        cardCount: Math.max(0, counts[idx]),
      })),
      currentTurn: a.playerIdx,
      tablePlaced: [...currentPlaced],
      tableMinVals: computeTableMinVals(currentPlaced),
      tableMaxVals: computeTableMaxVals(currentPlaced),
      cpuActions: actions.slice(0, i + 1),
      gameEndFlag: isLastAction ? finalState.gameEndFlag : false,
      message: isLastAction ? finalState.message : '',
    });
  }
  return states;
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
