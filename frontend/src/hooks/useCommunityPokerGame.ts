import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { HoldemLikeExec } from '../api/gameApi';
import { useSound } from '../providers/SoundProvider';
import type { OmahaResponse } from '../types/card';
import { OmahaPhase, OmahaRebuyPhaseType } from '../types/phases';
import type { CliGameConfig } from '../utils/cli/types';
import { useActionKeyboardNav } from './useActionKeyboardNav';
import { useCardDimensions, useIsLargeDesktop, useIsMobile } from './useCardDimensions';
import { useCliGame } from './useCliGame';
import { useCliMode } from './useCliMode';
import { useGameApi } from './useGameApi';
import { useGameHint } from './useGameHint';
import { useGamePageSetup } from './useGamePageSetup';
import { useMountReset } from './useMountReset';
import { usePhaseNames } from './usePhaseNames';

/** The exec function shared by every community-poker game API. */
type CommunityPokerExec = HoldemLikeExec<OmahaResponse>;

/** Community-card poker games sharing the Hold'em response shape + phase enum. */
export type CommunityPokerName = 'holdem' | 'omaha' | 'omahahilo' | 'bigo' | 'bigohilo' | 'shortdeck';

/** Config for {@link useCommunityPokerGame}. */
export interface CommunityPokerGameConfig {
  /** Game key (i18n namespace, hint registry, CLI mode, useGamePageSetup). */
  game: CommunityPokerName;
  /** The game's REST exec function (e.g. `omahaApi.exec`). */
  exec: CommunityPokerExec;
  /** Phase → i18n key map for `usePhaseNames`. */
  phaseKeys: Readonly<Record<number, string>>;
  /** CLI command parser / formatter / help for the terminal fallback. */
  cli: Pick<
    CliGameConfig<OmahaResponse, Parameters<CommunityPokerExec>>,
    'parseCommand' | 'formatResponse' | 'helpText'
  >;
}

/**
 * Shared orchestration for the community-card poker pages (Omaha, Big O, their
 * Hi-Lo variants, Hold'em, Short Deck) — issue #4301. Centralizes the identical
 * ~120 lines each page copied: server state, bet-amount / learning / meta-AI
 * state, the meta-AI elapsed-time tracking, CLI wiring, mount reset, keyboard
 * shortcuts, and the derived phase/turn booleans the JSX branches on. Each page
 * still supplies its own layout (tutorial prefixes, hole-card count, Hi-Lo
 * display, best-five evaluator differ by variant).
 */
export function useCommunityPokerGame(config: CommunityPokerGameConfig) {
  const { game, exec, phaseKeys, cli } = config;

  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup(game);
  const phaseNames = usePhaseNames(game, phaseKeys);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const isMobile = useIsMobile();
  const isLargeDesktop = useIsLargeDesktop();
  const { state, loading, error, exec: execApi, retry } = useGameApi(exec);
  const [betAmount, setBetAmount] = useState(20);
  const [learningMode, setLearningMode] = useState(false);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const { hint, hintEnabled, setHintEnabled } = useGameHint(game, state);
  const turnStartRef = useRef(0);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode(game);
  const cliConfig: CliGameConfig<OmahaResponse, Parameters<CommunityPokerExec>> = useMemo<
    CliGameConfig<OmahaResponse, Parameters<CommunityPokerExec>>
  >(
    () => ({
      gameName: game,
      parseCommand: cli.parseCommand,
      formatResponse: cli.formatResponse,
      helpText: cli.helpText,
    }),
    [game, cli.parseCommand, cli.formatResponse, cli.helpText],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void execApi('reset', undefined, { cpuMetaAI });
  }, [execApi, hideActionLog, cpuMetaAI]);

  useEffect(() => {
    if (state?.minRaise && state.minRaise > 0) {
      setBetAmount(state.minRaise);
    } else if (state) {
      setBetAmount(20);
    }
  }, [state]);

  useEffect(() => {
    if (state && state.currentTurn === state.players?.find((p) => p.isHuman)?.id) {
      turnStartRef.current = Date.now();
    }
  }, [state]);

  const getElapsed = useCallback(() => {
    if (!cpuMetaAI || turnStartRef.current === 0) return 0;
    const elapsed = Date.now() - turnStartRef.current;
    turnStartRef.current = 0;
    return elapsed;
  }, [cpuMetaAI]);

  const phase = state?.phase ?? OmahaPhase.INIT;
  const isActive = phase >= OmahaPhase.PRE_FLOP && phase <= OmahaPhase.RIVER;
  const isShowdown = phase === OmahaPhase.SHOWDOWN || phase === OmahaPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isActive && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;
  const isMuckPhase = phase === OmahaPhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase = phase === OmahaPhase.REBUY && state?.rebuyPhaseType === OmahaRebuyPhaseType.REBUY;
  const isAddonPhase = phase === OmahaPhase.REBUY && state?.rebuyPhaseType === OmahaRebuyPhaseType.ADDON;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const humanRebuyCount = state?.rebuyCounts?.[humanIdx] ?? 0;
  const cpuPlayers = useMemo(() => state?.players?.filter((player) => !player.isHuman) ?? [], [state?.players]);

  const actionBindings = useMemo(
    () => [
      { key: 'c', action: () => execApi('call', undefined, undefined, getElapsed()), enabled: hasOutstandingBet },
      {
        key: 'r',
        action: () =>
          hasOutstandingBet
            ? execApi('raise', betAmount, undefined, getElapsed())
            : execApi('bet', betAmount, undefined, getElapsed()),
      },
      { key: 'k', action: () => execApi('check', undefined, undefined, getElapsed()), enabled: !hasOutstandingBet },
      { key: 'f', action: () => execApi('fold', undefined, undefined, getElapsed()) },
      { key: 'a', action: () => execApi('allin', undefined, undefined, getElapsed()) },
    ],
    [execApi, hasOutstandingBet, betAmount, getElapsed],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: canAct && !loading });

  return {
    t,
    tc,
    actionLog,
    showActionLog,
    hideActionLog,
    confirmOpen,
    requestConfirm,
    confirmReset,
    cancelReset,
    phaseNames,
    cardWidth,
    playSound,
    isMobile,
    isLargeDesktop,
    state,
    loading,
    error,
    execApi,
    retry,
    betAmount,
    setBetAmount,
    learningMode,
    setLearningMode,
    cpuMetaAI,
    setCpuMetaAI,
    hint,
    hintEnabled,
    setHintEnabled,
    getElapsed,
    cliEnabled,
    toggleCli,
    logEntries,
    handleCommand,
    handleManualReset,
    phase,
    isActive,
    isShowdown,
    humanPlayer,
    humanFolded,
    humanAllIn,
    canAct,
    hasOutstandingBet,
    minRaise,
    isMuckPhase,
    isRebuyPhase,
    isAddonPhase,
    humanRebuyCount,
    cpuPlayers,
  };
}
