import { useCallback, useEffect, useState } from 'react';
import { sevensApi } from '../api/gameApi';
import type { SevensResponse } from '../types/card';
import { useGameApi } from './useGameApi';

export function useSevensGame() {
  const [jokerCardIdx, setJokerCardIdx] = useState<number | null>(null);
  const [cfgTunnel, setCfgTunnel] = useState(false);
  const [cfgJokerCount, setCfgJokerCount] = useState(0);
  const [cfgCpuStrategy, setCfgCpuStrategy] = useState(false);
  const [cfgMaxPasses, setCfgMaxPasses] = useState(5);
  const [cfgNoJokerFinish, setCfgNoJokerFinish] = useState(false);
  const [cfgJokerReclaim, setCfgJokerReclaim] = useState(false);
  const [cfgEndStop, setCfgEndStop] = useState(false);
  const [cfgJokerConsBan, setCfgJokerConsBan] = useState(false);

  const onSuccess = useCallback((res: SevensResponse) => {
    setJokerCardIdx(null);
    setCfgTunnel(res.config.tunnelEnabled);
    setCfgJokerCount(res.config.jokerCount);
    setCfgCpuStrategy(res.config.cpuStrategy);
    setCfgMaxPasses(res.config.maxPasses);
    setCfgNoJokerFinish(res.config.noJokerFinish);
    setCfgJokerReclaim(res.config.jokerReclaimEnabled);
    setCfgEndStop(res.config.endStopEnabled);
    setCfgJokerConsBan(res.config.jokerConsecutiveBanned);
  }, []);
  const { state, loading, error, exec } = useGameApi(sevensApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleCardPlay = (idx: number) => {
    const humanPlayer = state?.players.find((p) => p.isHuman);
    const card = humanPlayer?.cards?.[idx];
    if (card?.design === 'JOKER') {
      setJokerCardIdx(idx);
    } else {
      exec('play', idx);
    }
  };

  const handleJokerPlace = (suit: number, value: number) => {
    exec('joker', jokerCardIdx as number, suit, value);
  };

  return {
    state,
    loading,
    error,
    exec,
    jokerCardIdx,
    setJokerCardIdx,
    cfgTunnel,
    setCfgTunnel,
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
