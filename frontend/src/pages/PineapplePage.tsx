import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { crazyPineappleApi, irishPokerApi, pineappleApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CpuAccordion } from '../components/CpuAccordion';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuActionToast } from '../components/CpuActionToast';
import { CpuPlayerCard } from '../components/CpuPlayerCard';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { EquityDisplay } from '../components/EquityDisplay';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HudStats } from '../components/HudStats';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { PokerTableLayout } from '../components/PokerTableLayout';
import { RoundResults } from '../components/RoundResults';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { placeholderCardStyle, selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PineappleResponse } from '../types/card';
import { HoldemRebuyPhaseType, PineapplePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PINEAPPLE_HELP, parsePineappleCommand } from '../utils/cli/commands/pineappleCommands';
import { formatPineappleState } from '../utils/cli/formatters/pineappleFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { holdemBestFive } from '../utils/holdemBestFive';
import { type PineappleKeepFeature, pineappleKeepFeatures } from '../utils/pineappleDiscardHint';
import { findPlayerName } from '../utils/playerUtils';
import { evaluateFiveCardHand, type PokerHandRank, pokerHandKey } from '../utils/pokerSquaresUtils';

/** Pineapple Poker tutorial step definitions. */
const PN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pn-community-cards"]',
    messageKey: 'tutorial.communityCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-discard-controls"]',
    messageKey: 'tutorial.discardPhase',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-learning-mode"]',
    messageKey: 'tutorial.learningMode',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PINEAPPLE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PineapplePhase.PRE_FLOP]: 'preFlop',
  [PineapplePhase.FLOP]: 'flop',
  [PineapplePhase.TURN]: 'turn',
  [PineapplePhase.RIVER]: 'river',
  [PineapplePhase.SHOWDOWN]: 'showdown',
  [PineapplePhase.END]: 'end',
  [PineapplePhase.REBUY]: 'rebuy',
  [PineapplePhase.DISCARD]: 'discard',
};

/** Variant of the Pineapple page: standard "pineapple" or the Crazy Pineapple
 * variant (discard happens after the flop betting round instead of before). */
export type PineappleVariant = 'pineapple' | 'crazypineapple' | 'irishpoker';

/** Renders the Pineapple Poker game page with community cards, discard phase, betting, and showdown. */
export function PineapplePage({ variant = 'pineapple' }: { variant?: PineappleVariant } = {}) {
  return (
    <TutorialWrapper gameName={variant} steps={PN_TUTORIAL_STEPS}>
      <PineapplePageContent variant={variant} />
    </TutorialWrapper>
  );
}

/** Inner content of the Pineapple Poker page, wrapped by TutorialProvider. */
function PineapplePageContent({ variant }: { variant: PineappleVariant }) {
  const apiClient =
    variant === 'irishpoker' ? irishPokerApi : variant === 'crazypineapple' ? crazyPineappleApi : pineappleApi;
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup(variant);
  const phaseNames = usePhaseNames(variant, PINEAPPLE_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const isMobile = useIsMobile();
  const { state, loading, error, exec: apiExec, retry } = useGameApi(apiClient.exec);
  const [betAmount, setBetAmount] = useState(20);
  const [learningMode, setLearningMode] = useState(false);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const { hint, hintEnabled, setHintEnabled } = useGameHint(variant, state);
  // Indices selected for discard. Pineapple/Crazy Pineapple discard 1 (3→2);
  // Irish Poker discards 2 (4→2). Capped at `discardCount` below.
  const [selectedDiscards, setSelectedDiscards] = useState<number[]>([]);
  // Two-step discard: pressing "discard" enters a confirm step before committing.
  const [discardConfirming, setDiscardConfirming] = useState(false);
  // Feedback when a click is ignored because the discard cap is already reached:
  // a message for the live region and a nonce that re-fires it (and the shake)
  // on every repeated over-limit attempt, even with identical text.
  const [limitAnnounce, setLimitAnnounce] = useState('');
  const [limitNonce, setLimitNonce] = useState(0);
  // Mirror the selection in a ref so toggleDiscard can read it without listing
  // selectedDiscards as a dependency — otherwise the callback (and the global
  // keydown listener in useCardKeyboardNav) would be re-created on every toggle.
  const selectedDiscardsRef = useRef<number[]>(selectedDiscards);
  selectedDiscardsRef.current = selectedDiscards;
  const turnStartRef = useRef(0);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode(variant);
  const cliConfig: CliGameConfig<PineappleResponse, Parameters<typeof apiClient.exec>> = useMemo(
    () => ({
      gameName: variant,
      parseCommand: parsePineappleCommand,
      formatResponse: formatPineappleState,
      helpText: PINEAPPLE_HELP,
    }),
    [variant],
  );
  const { handleCommand } = useCliGame(apiExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(apiExec);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void apiExec('reset', undefined, { cpuMetaAI });
  }, [apiExec, hideActionLog, cpuMetaAI]);

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

  // Reset discard selection (and any pending confirm / limit banner) when leaving
  // the discard phase, so a stale over-limit warning can't outlive it.
  useEffect(() => {
    if (!state?.isDiscardPhase) {
      setSelectedDiscards([]);
      setDiscardConfirming(false);
      setLimitAnnounce('');
    }
  }, [state?.isDiscardPhase]);

  const getElapsed = useCallback(() => {
    if (!cpuMetaAI || turnStartRef.current === 0) return 0;
    const elapsed = Date.now() - turnStartRef.current;
    turnStartRef.current = 0;
    return elapsed;
  }, [cpuMetaAI]);

  const phase = state?.phase ?? PineapplePhase.INIT;
  const isActive = phase >= PineapplePhase.PRE_FLOP && phase <= PineapplePhase.RIVER;
  const isShowdown = phase === PineapplePhase.SHOWDOWN || phase === PineapplePhase.END;
  const isDiscardPhase = state?.isDiscardPhase ?? false;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isActive && !humanFolded && !humanAllIn && !isDiscardPhase && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  // Best-5 highlight at showdown: after the discard, the hand is Hold'em-style
  // (2 hole + 5 board), so holdemBestFive marks the winning five cards. Indices
  // 0..1 map to the hole cards, 2..6 to the board.
  const showdownBest5 = useMemo(() => {
    const hole = humanPlayer?.cards ?? [];
    const board = state?.communityCards ?? [];
    if (!isShowdown || !humanPlayer || humanPlayer.folded || hole.length !== 2 || board.length < 5) {
      return { holeSet: new Set<number>(), boardSet: new Set<number>() };
    }
    const picked = holdemBestFive([...hole, ...board.slice(0, 5)]) ?? [];
    const holeSet = new Set<number>();
    const boardSet = new Set<number>();
    for (const i of picked) {
      if (i < hole.length) holeSet.add(i);
      else boardSet.add(i - hole.length);
    }
    return { holeSet, boardSet };
  }, [isShowdown, humanPlayer, state?.communityCards]);
  const minRaise = state?.minRaise ?? 0;
  const isMuckPhase = phase === PineapplePhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase = phase === PineapplePhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.REBUY;
  const isAddonPhase = phase === PineapplePhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.ADDON;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const humanRebuyCount = state?.rebuyCounts?.[humanIdx] ?? 0;
  const cpuPlayers = useMemo(() => state?.players?.filter((p) => !p.isHuman) ?? [], [state?.players]);
  const humanDiscardDone = state?.discardDone?.[humanIdx] ?? false;
  const canDiscard = isDiscardPhase && !humanDiscardDone;
  // Cards the player must discard down to 2: Pineapple/Crazy = 1, Irish = 2.
  const discardCount = Math.max(1, (state?.initialDealCount ?? 3) - 2);
  // Irish Poker discards 2 of 4 hole cards; while choosing, preview the 2 cards
  // that would be kept plus the best hand they make with the current board.
  const discardPreview = useMemo(() => {
    const isIrishDiscardChoice = variant === 'irishpoker' && isDiscardPhase && selectedDiscards.length === discardCount;
    if (!isIrishDiscardChoice) return null;
    const kept = (humanPlayer?.cards ?? []).filter((_, i) => !selectedDiscards.includes(i));
    // holdemBestFive picks the best 5 of (kept + board); it returns null for <5 cards.
    const all = [...kept, ...(state?.communityCards ?? [])];
    const picked = holdemBestFive(all);
    const rank = picked ? evaluateFiveCardHand(picked.map((i) => all[i])) : null;
    return { kept, handKey: rank == null ? null : pokerHandKey(rank) };
  }, [variant, isDiscardPhase, humanPlayer, state?.communityCards, selectedDiscards, discardCount]);
  // Crazy Pineapple discards 1 of 3 after the flop; annotate each hole card with
  // the best hand the OTHER two would make with the board if that card is the
  // one discarded, so the player can compare keeps before committing. Each entry
  // carries both the i18n hand key and the raw rank, so the strongest keep can be
  // flagged as recommended below.
  const candidatePreviews = useMemo<({ handKey: string; rank: PokerHandRank } | null)[] | null>(() => {
    if (variant !== 'crazypineapple' || !isDiscardPhase) return null;
    const hole = humanPlayer?.cards ?? [];
    const board = state?.communityCards ?? [];
    return hole.map((_, discardIdx) => {
      const all = [...hole.filter((_, i) => i !== discardIdx), ...board];
      const picked = holdemBestFive(all);
      const rank = picked ? evaluateFiveCardHand(picked.map((i) => all[i])) : null;
      return rank == null ? null : { handKey: pokerHandKey(rank), rank };
    });
  }, [variant, isDiscardPhase, humanPlayer, state?.communityCards]);
  // The Crazy Pineapple discard(s) whose removal leaves the strongest resulting
  // hand — flagged with a "recommended" badge and ring. When several discards
  // tie for the best rank, all of them are flagged.
  const recommendedDiscards = useMemo<Set<number>>(() => {
    const out = new Set<number>();
    if (!candidatePreviews) return out;
    let best = -1;
    for (const p of candidatePreviews) {
      if (p && p.rank > best) best = p.rank;
    }
    if (best < 0) return out;
    candidatePreviews.forEach((p, i) => {
      if (p && p.rank === best) out.add(i);
    });
    return out;
  }, [candidatePreviews]);
  // Irish Poker discards 2 of 4 hole cards. Once the FIRST card is chosen, annotate
  // each still-selectable card with the best hand the OTHER two kept cards would make
  // with the board if that card became the second discard — turning the C(3,1)=3
  // remaining choices into a side-by-side comparison before the pair is committed.
  // (At 2 selected, the full `discardPreview` above takes over instead.)
  const irishCandidatePreviews = useMemo<(string | null)[] | null>(() => {
    if (variant !== 'irishpoker' || !isDiscardPhase) return null;
    if (selectedDiscards.length !== discardCount - 1) return null;
    const hole = humanPlayer?.cards ?? [];
    const board = state?.communityCards ?? [];
    return hole.map((_, idx) => {
      // The already-selected card is marked as selected, not annotated.
      if (selectedDiscards.includes(idx)) return null;
      const kept = hole.filter((_, i) => i !== idx && !selectedDiscards.includes(i));
      const all = [...kept, ...board];
      const picked = holdemBestFive(all);
      const rank = picked ? evaluateFiveCardHand(picked.map((i) => all[i])) : null;
      return rank == null ? null : pokerHandKey(rank);
    });
  }, [variant, isDiscardPhase, humanPlayer, state?.communityCards, selectedDiscards, discardCount]);
  // Plain Pineapple discards 1 of 3 preflop (before any board), so a board-based
  // preview is impossible. Instead annotate each hole card with the qualitative
  // shape (pair / suited / connector / high card) the OTHER two would keep, a
  // board-free judgment the player can use to pick which card to throw.
  const keepFeaturePreviews = useMemo<(PineappleKeepFeature[] | null)[] | null>(() => {
    if (variant !== 'pineapple' || !isDiscardPhase) return null;
    const hole = humanPlayer?.cards ?? [];
    if (hole.length !== 3) return null;
    return hole.map((_, discardIdx) => {
      const [a, b] = hole.filter((_, i) => i !== discardIdx);
      return pineappleKeepFeatures(a, b);
    });
  }, [variant, isDiscardPhase, humanPlayer]);

  const actionBindings = useMemo(
    () => [
      { key: 'c', action: () => apiExec('call', undefined, undefined, getElapsed()), enabled: hasOutstandingBet },
      {
        key: 'r',
        action: () =>
          hasOutstandingBet
            ? apiExec('raise', betAmount, undefined, getElapsed())
            : apiExec('bet', betAmount, undefined, getElapsed()),
      },
      { key: 'k', action: () => apiExec('check', undefined, undefined, getElapsed()), enabled: !hasOutstandingBet },
      { key: 'f', action: () => apiExec('fold', undefined, undefined, getElapsed()) },
      { key: 'a', action: () => apiExec('allin', undefined, undefined, getElapsed()) },
    ],
    [apiExec, hasOutstandingBet, betAmount, getElapsed],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: canAct && !loading,
  });

  // Discard phase control shared by mouse and keyboard: toggle a hole card into
  // the discard selection. Guarded so nothing changes once the two-stage confirm
  // has started (the player must commit or cancel), keeping both input methods
  // consistent. Keyboard: number keys toggle, Enter steps select -> confirm ->
  // commit, Escape backs out.
  const discardCardCount = humanPlayer?.cards?.length ?? 0;
  const toggleDiscard = useCallback(
    (idx: number) => {
      if (!canDiscard || discardConfirming) return;
      const current = selectedDiscardsRef.current;
      const isSelected = current.includes(idx);
      // Selecting a new card while already at the cap is ignored — but tell the
      // user why (live region + shake) instead of silently swallowing the click.
      if (!isSelected && current.length >= discardCount) {
        setLimitAnnounce(t('discard.limitReached', { count: discardCount }));
        setLimitNonce((n) => n + 1);
        return;
      }
      setLimitAnnounce('');
      setSelectedDiscards((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
    },
    [canDiscard, discardConfirming, discardCount, t],
  );
  const discardConfirm = useCallback(() => {
    if (selectedDiscards.length !== discardCount) return;
    if (discardConfirming) {
      void apiExec('discard', undefined, { cardIdxs: [...selectedDiscards] });
      setSelectedDiscards([]);
      setDiscardConfirming(false);
    } else {
      setDiscardConfirming(true);
    }
  }, [selectedDiscards, discardCount, discardConfirming, apiExec]);
  const discardClear = useCallback(() => {
    if (discardConfirming) setDiscardConfirming(false);
    else setSelectedDiscards([]);
  }, [discardConfirming]);
  useCardKeyboardNav({
    cardCount: discardCardCount,
    onToggle: toggleDiscard,
    onConfirm: discardConfirm,
    onClear: discardClear,
    enabled: canDiscard && !loading,
  });

  if (!state)
    return (
      <GameSkeleton
        gameKey="pineapple"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 2, footerHandSize: 2 }}
      />
    );

  return (
    <GamePageShell
      title={tc(`nav.${variant}`)}
      gameThemeBg={gameTheme[variant].bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct || canDiscard}
      gamePath={`/${variant}`}
      gameEndFlag={!!state?.gameEndFlag}
      winShow={phase === PineapplePhase.END}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="pn-pot-display">
            {tc('label.pot')} <strong>{state?.pot ?? 0}</strong>
          </span>
          <span>
            SB/BB:{' '}
            <strong>
              {state?.smallBlind ?? 0}/{state?.bigBlind ?? 0}
            </strong>
          </span>
          <span>
            {tc('label.dealer')} <strong>{findPlayerName(state.players, state.dealerIdx)}</strong>
          </span>
          {state?.tournamentMode && (
            <span>{t('handNumber', { count: state.handCount, level: state.blindLevelHands })}</span>
          )}
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Scrollable: community cards + CPU players */}
          <div className={`flex-1 overflow-y-auto pt-4 px-5 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Community cards + CPU players (poker table layout on desktop, accordion on mobile) */}
            {(() => {
              const communityCardsContent = (
                <>
                  <div className="text-ds-text-primary text-lg mb-1.5">{t('communityCards')}</div>
                  <div className="flex flex-wrap gap-2">
                    {state?.communityCards?.length
                      ? state.communityCards.map((card, idx) => {
                          const inBest = showdownBest5.boardSet.has(idx);
                          const dim = showdownBest5.boardSet.size > 0 && !inBest;
                          return (
                            <div
                              key={`${card.design}-${card.value}`}
                              className={`transition-all ${inBest ? '-translate-y-1 rounded-lg ring-2 ring-ds-success motion-safe:animate-pulse' : ''} ${dim ? 'opacity-50' : ''}`}
                              data-testid={inBest ? 'pn-best5-card' : undefined}
                            >
                              <AnimatedCard card={card} width={cardWidth} style={placeholderCardStyle} />
                            </div>
                          );
                        })
                      : Array.from({ length: 5 }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
                  </div>
                </>
              );
              const cpuPlayerCards = cpuPlayers.map((p) => (
                <CpuPlayerCard
                  key={p.id}
                  player={p}
                  showCards={isShowdown}
                  faceDownCount={2}
                  showHandName={isShowdown}
                  extraInfo={
                    p.totalHands > 0 ? (
                      <HudStats namespace={variant} vpip={p.vpip} pfr={p.pfr} threeBet={p.threeBet} af={p.af} />
                    ) : undefined
                  }
                />
              ));

              if (!isMobile) {
                return (
                  <PokerTableLayout
                    communityCardsTutorial="pn-community-cards"
                    cpuAreaTutorial="pn-cpu-area"
                    communityCards={communityCardsContent}
                    cpuPlayers={cpuPlayerCards}
                  />
                );
              }

              return (
                <>
                  <div
                    className="sticky top-0 z-10 bg-game-bg-green-poker pb-1 shadow-sm"
                    data-tutorial="pn-community-cards"
                  >
                    {communityCardsContent}
                  </div>
                  <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="pn-cpu-area">
                    {cpuPlayerCards}
                  </CpuAccordion>
                </>
              );
            })()}

            {/* CPU actions: toast on mobile, inline log on desktop */}
            {isMobile ? <CpuActionToast actions={state?.cpuActions} /> : <CpuActionLog actions={state?.cpuActions} />}

            {/* Round results */}
            {isShowdown && <RoundResults results={state?.roundResults} players={state?.players ?? []} />}

            {/* Action log */}
            <ActionLogSection
              isEndPhase={!!state?.gameEndFlag}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Sticky footer: player hand + buttons */}
          <GameFooter className={`${gameTheme[variant].footer} px-5 py-3`}>
            {/* Crazy Pineapple and Irish Poker discard after the flop betting round
                (unlike plain Pineapple's immediate discard) — forewarn the player during
                the flop bet so they can factor the upcoming discard into their decision. */}
            {(variant === 'crazypineapple' || variant === 'irishpoker') && phase === PineapplePhase.FLOP && (
              <div
                data-testid="cp-discard-upcoming-banner"
                className="mb-2 rounded border border-ds-info bg-ds-surface px-3 py-1.5 text-center text-ds-info text-sm"
              >
                {t('discardUpcoming')}
              </div>
            )}
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="pn-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
                  {t('yourHand')}
                  <span className="ml-3 text-xs">
                    {tc('betting.chips')} {humanPlayer.chips}
                  </span>
                  {humanPlayer.totalHands > 0 && (
                    <HudStats
                      namespace={variant}
                      vpip={humanPlayer.vpip}
                      pfr={humanPlayer.pfr}
                      threeBet={humanPlayer.threeBet}
                      af={humanPlayer.af}
                    />
                  )}
                  {humanPlayer.currentBet > 0 && (
                    <span className="ml-2 text-xs">
                      {tc('betting.currentBet')} {humanPlayer.currentBet}
                    </span>
                  )}
                  {humanPlayer.folded && <span className="ml-2 text-ds-error text-xs">[{tc('status.folded')}]</span>}
                  {humanPlayer.allIn && <span className="ml-2 text-ds-warning text-xs">[{tc('status.allIn')}]</span>}
                  {isShowdown && !humanPlayer.folded && humanPlayer.handName && (
                    <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                      {humanPlayer.handName}
                    </span>
                  )}
                </div>
                {/* Screen-reader description of the discard cap, associated with the
                    hand group so AT conveys the limit up front. */}
                {canDiscard && (
                  <span id="pn-discard-limit-desc" className="sr-only">
                    {t('discard.limitDescription', { count: discardCount })}
                  </span>
                )}
                <div
                  className="flex flex-wrap gap-1.5 mb-2"
                  aria-describedby={canDiscard ? 'pn-discard-limit-desc' : undefined}
                >
                  {humanPlayer.cards?.length
                    ? humanPlayer.cards.map((card, idx) => {
                        const isSelected = selectedDiscards.includes(idx);
                        const inBest = showdownBest5.holeSet.has(idx);
                        const dim = showdownBest5.holeSet.size > 0 && !inBest;
                        const candKey = candidatePreviews?.[idx]?.handKey ?? null;
                        const isRecommendedDiscard = recommendedDiscards.has(idx);
                        const irishCandKey = irishCandidatePreviews?.[idx] ?? null;
                        const keepFeatures = keepFeaturePreviews?.[idx] ?? null;
                        return (
                          <div key={`${card.design}-${card.value}`} className="flex flex-col items-center">
                            <button
                              type="button"
                              onClick={() => toggleDiscard(idx)}
                              aria-pressed={canDiscard ? isSelected : undefined}
                              className={`${canDiscard ? 'cursor-pointer' : 'cursor-default'} ${inBest ? 'rounded-lg ring-2 ring-ds-success motion-safe:animate-pulse' : ''} ${isRecommendedDiscard ? 'rounded-lg ring-2 ring-ds-info' : ''} ${dim ? 'opacity-50' : ''}`}
                              disabled={!canDiscard}
                              style={selectedCardStyle(canDiscard && isSelected)}
                              data-testid={inBest ? 'pn-best5-card' : undefined}
                            >
                              <AnimatedCard card={card} width={cardWidth} />
                            </button>
                            {candKey && (
                              <span
                                className="mt-0.5 text-[10px] text-ds-text-muted"
                                data-testid="cp-discard-candidate"
                              >
                                {`${t('discard.candidateHand')}: ${t(`hand.${candKey}`)}`}
                              </span>
                            )}
                            {isRecommendedDiscard && (
                              <span
                                className="mt-0.5 rounded-full bg-ds-info/20 px-1.5 text-[10px] font-semibold text-ds-info"
                                data-testid="cp-discard-recommended"
                              >
                                {t('discard.recommended')}
                              </span>
                            )}
                            {irishCandKey && (
                              <span
                                className="mt-0.5 text-[10px] text-ds-text-muted"
                                data-testid="irishpoker-discard-candidate"
                              >
                                {`${t('discard.candidateHand')}: ${t(`hand.${irishCandKey}`)}`}
                              </span>
                            )}
                            {keepFeatures && (
                              <span
                                className="mt-0.5 text-[10px] text-ds-text-muted"
                                data-testid="pn-discard-keep-feature"
                              >
                                {`${t('discard.keepLabel')}: ${keepFeatures
                                  .map((f) => t(`discard.feature${f.charAt(0).toUpperCase()}${f.slice(1)}`))
                                  .join('・')}`}
                              </span>
                            )}
                          </div>
                        );
                      })
                    : !humanPlayer.folded &&
                      Array.from({ length: state?.initialDealCount ?? 3 }).map((_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth} />
                      ))}
                </div>
                {/* Over-limit feedback: visible (color-blind-safe) + announced to AT.
                    Keyed on the nonce so a repeated over-limit click re-fires the
                    announcement and restarts the shake even with identical text. */}
                {limitAnnounce && (
                  <div
                    key={limitNonce}
                    role="status"
                    aria-live="polite"
                    data-testid="pn-discard-limit-announce"
                    className="text-center text-ds-warning text-xs mb-2 motion-safe:animate-shake"
                  >
                    {limitAnnounce}
                  </div>
                )}
              </div>
            )}

            {/* Message */}
            <GameMessageBox
              message={state?.message}
              messageCode={state?.messageCode}
              messageParams={state?.messageParams}
              alwaysVisible
            />

            <ErrorAlert message={error} onRetry={retry} />

            {/* Discard controls */}
            {canDiscard && (
              <div className="mb-2 text-center" data-testid="discard-controls" data-tutorial="pn-discard-controls">
                {discardPreview && (
                  <div className="mb-2 text-sm" data-testid="irishpoker-discard-preview">
                    <span className="text-ds-text-muted">{`${t('discard.keepLabel')}: `}</span>
                    <span className="text-ds-text-primary font-semibold">
                      {discardPreview.kept.map((c) => cardAlt(c)).join('  ')}
                    </span>
                    {discardPreview.handKey && (
                      <span className="ml-2 inline-block rounded-full bg-ds-accent/30 px-2 py-0.5 text-ds-text-primary font-semibold">
                        {`${t('discard.previewHand')}: ${t(`hand.${discardPreview.handKey}`)}`}
                      </span>
                    )}
                  </div>
                )}
                {discardConfirming && selectedDiscards.length === discardCount && humanPlayer ? (
                  <div data-testid="discard-confirm">
                    <p className="text-ds-text-primary mb-2">
                      {t('discard.confirm', {
                        card: selectedDiscards.map((i) => cardAlt(humanPlayer.cards[i])).join(', '),
                      })}
                    </p>
                    <div className="flex justify-center gap-2">
                      <button
                        type="button"
                        className={`${btnPrimary} min-w-[90px]`}
                        disabled={loading}
                        onClick={() => {
                          apiExec('discard', undefined, { cardIdxs: [...selectedDiscards] });
                          setSelectedDiscards([]);
                          setDiscardConfirming(false);
                        }}
                      >
                        {t('discard.confirmYes')}
                      </button>
                      <button
                        type="button"
                        className={`${btnSecondary} min-w-[90px]`}
                        disabled={loading}
                        onClick={() => setDiscardConfirming(false)}
                      >
                        {t('discard.confirmNo')}
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <p className="text-ds-text-primary mb-2">{t('discard.select')}</p>
                    {discardCount > 1 && (
                      <p className="text-ds-text-muted text-xs mb-2" data-testid="discard-count">
                        {t('discard.selectedCount', { n: selectedDiscards.length, total: discardCount })}
                      </p>
                    )}
                    <button
                      type="button"
                      className={`${btnPrimary} min-w-[90px]`}
                      disabled={loading || selectedDiscards.length !== discardCount}
                      onClick={() => setDiscardConfirming(true)}
                    >
                      {t('discard.prompt')}
                    </button>
                  </>
                )}
              </div>
            )}

            {/* Muck/Show controls */}
            {isMuckPhase && (
              <div className="mb-2 text-center" data-testid="muck-controls">
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={`${btnPrimary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('muck')}
                  >
                    {t('muck.muck')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('show')}
                  >
                    {t('muck.show')}
                  </button>
                </div>
              </div>
            )}

            {/* Rebuy/Addon controls */}
            {isRebuyPhase && (
              <div className="mb-2 text-center" data-testid="rebuy-controls">
                <p className="text-ds-text-primary mb-2">
                  {t('rebuy.prompt', { chips: state?.rebuyChips, used: humanRebuyCount, max: state?.rebuyMaxCount })}
                </p>
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={`${btnPrimary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('rebuy')}
                  >
                    {t('rebuy.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('skiprebuy')}
                  >
                    {t('rebuy.skip')}
                  </button>
                </div>
              </div>
            )}
            {isAddonPhase && (
              <div className="mb-2 text-center" data-testid="addon-controls">
                <p className="text-ds-text-primary mb-2">{t('addon.prompt', { chips: state?.addonChips })}</p>
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={`${btnPrimary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('addon')}
                  >
                    {t('addon.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('skipaddon')}
                  >
                    {t('addon.skip')}
                  </button>
                </div>
              </div>
            )}

            {/* Hint */}
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {/* Betting controls */}
            {canAct && (
              <div data-tutorial="pn-action-buttons">
                <BettingControls
                  inputId="pineappleBetAmount"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={minRaise}
                  maxBetAmount={state?.maxBetAmount}
                  potSize={state?.pot}
                  hasOutstandingBet={hasOutstandingBet}
                  loading={loading}
                  onCall={() => apiExec('call', undefined, undefined, getElapsed())}
                  onRaise={() => apiExec('raise', betAmount, undefined, getElapsed())}
                  onBet={() => apiExec('bet', betAmount, undefined, getElapsed())}
                  onCheck={() => apiExec('check', undefined, undefined, getElapsed())}
                  onFold={() => apiExec('fold', undefined, undefined, getElapsed())}
                  onAllIn={() => apiExec('allin', undefined, undefined, getElapsed())}
                />
              </div>
            )}

            {/* Settings + Reset */}
            <details className="mb-1" data-tutorial="pn-learning-mode" open={learningMode || undefined}>
              <summary className="cursor-pointer select-none text-ds-text-primary text-sm font-bold py-1">
                {tc('settings.title')}
              </summary>
              <div className="flex flex-col gap-2 py-1">
                <div className="flex items-center gap-2" data-testid="learning-mode-toggle">
                  <label htmlFor="learningModeCheckbox" className="text-ds-text-primary text-sm cursor-pointer">
                    {t('learning.toggle')}
                  </label>
                  <input
                    id="learningModeCheckbox"
                    type="checkbox"
                    checked={learningMode}
                    onChange={(e) => setLearningMode(e.target.checked)}
                  />
                </div>
                {learningMode && state?.equity && state.potOdds != null && (
                  <EquityDisplay equity={state.equity} potOdds={state.potOdds} />
                )}
                <div className="flex items-center gap-3">
                  <label className="text-ds-text-primary text-sm flex items-center gap-1">
                    <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
                    {tc('hint.toggle', { ns: 'tutorial' })}
                  </label>
                  <label className="text-ds-text-primary text-sm flex items-center gap-1">
                    <input type="checkbox" checked={cpuMetaAI} onChange={(e) => setCpuMetaAI(e.target.checked)} />
                    {t('settings.cpuMetaAI')}
                  </label>
                </div>
              </div>
            </details>
            <GameResetButton
              isGameEnd={phase === PineapplePhase.SHOWDOWN || phase === PineapplePhase.END}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
              dataTutorial="pn-reset-button"
              className="min-w-[90px]"
            />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
