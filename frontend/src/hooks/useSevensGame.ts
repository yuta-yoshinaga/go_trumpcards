import { useCallback, useEffect, useRef, useState } from 'react';
import { type SevensConfigInput, sevensApi } from '../api/gameApi';
import type { SevensAction, SevensResponse } from '../types/card';
import { buildReplayStates } from '../utils/replayBuilder';
import { runReplay, shouldSkipReplay } from './gameReplay';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useIsMounted } from './useIsMounted';

/**
 * Local config state shape for the Sevens settings panel. Mirrors the
 * SevensConfigInput API shape (with all fields required) so handleManualReset
 * can pass `config` straight to the api call without per-field plumbing.
 */
export type SevensConfigState = Required<SevensConfigInput>;

const DEFAULT_SEVENS_CONFIG: SevensConfigState = {
  tunnelEnabled: false,
  tunnelSkipWidth: 0,
  jokerCount: 0,
  cpuStrategy: 0,
  maxPasses: 5,
  noJokerFinish: false,
  jokerReclaim: false,
  endStop: false,
  jokerConsecutiveBanned: false,
};

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

/** Hook that manages Sevens game state, joker placement, configuration, and CPU replay. */
export function useSevensGame() {
  const isMounted = useIsMounted();

  const [jokerCardIdx, setJokerCardIdx] = useState<number | null>(null);
  const { config, setConfig, handleConfigChange, handleToggle } =
    useGameConfig<SevensConfigState>(DEFAULT_SEVENS_CONFIG);
  const [displayState, setDisplayState] = useState<SevensResponse | null>(null);

  const lastReplayedActionsRef = useRef<SevensResponse['cpuActions']>(undefined);

  const onSuccess = useCallback(
    async (res: SevensResponse) => {
      setJokerCardIdx(null);
      // Server response uses jokerReclaimEnabled / endStopEnabled, but the API
      // request uses jokerReclaim / endStop — translate while writing into
      // local state so handleManualReset can pass `config` straight to the api.
      setConfig({
        tunnelEnabled: res.config.tunnelEnabled,
        tunnelSkipWidth: res.config.tunnelSkipWidth,
        jokerCount: res.config.jokerCount,
        cpuStrategy: res.config.cpuStrategy,
        maxPasses: res.config.maxPasses,
        noJokerFinish: res.config.noJokerFinish,
        jokerReclaim: res.config.jokerReclaimEnabled,
        endStop: res.config.endStopEnabled,
        jokerConsecutiveBanned: res.config.jokerConsecutiveBanned,
      });
      if (shouldSkipReplay(res.cpuActions ?? [], lastReplayedActionsRef, res, setDisplayState)) {
        return;
      }
      await runReplay(res, setDisplayState, {
        isMounted,
        buildReplayStates: buildSevensReplayStates,
      });
    },
    [setConfig, isMounted],
  );

  const { loading, error, exec, retry } = useGameApi(sevensApi.exec, { onSuccess });

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
    config,
    handleConfigChange,
    handleToggle,
    handleCardPlay,
    handleJokerPlace,
    retry,
  };
}
