import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { dramahaApi } from '../api/gameApi';
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
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useIsLargeDesktop, useIsMobile } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeInfoColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { placeholderCardStyle, selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DramahaResponse } from '../types/card';
import { DramahaPhase, DramahaRebuyPhaseType } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { DRAMAHA_HELP, parseDramahaCommand } from '../utils/cli/commands/dramahaCommands';
import { formatDramahaState } from '../utils/cli/formatters/dramahaFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { DRAMAHA_HOLE_CARDS, dramahaHands } from '../utils/dramahaBestFive';
import { findPlayerName } from '../utils/playerUtils';

/** Dramaha tutorial step definitions. */
const DR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dr-community-cards"]',
    messageKey: 'tutorial.communityCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dr-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dr-split-rule"]',
    messageKey: 'tutorial.splitRule',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dr-draw-round"]',
    messageKey: 'tutorial.drawRound',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dr-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dr-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dr-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/**
 * Phase → i18n key. DRAW is numerically last (8) but happens between FLOP and
 * TURN; the map is keyed, not ordered, so its position here is cosmetic.
 */
const DRAMAHA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [DramahaPhase.PRE_FLOP]: 'preFlop',
  [DramahaPhase.FLOP]: 'flop',
  [DramahaPhase.DRAW]: 'draw',
  [DramahaPhase.TURN]: 'turn',
  [DramahaPhase.RIVER]: 'river',
  [DramahaPhase.SHOWDOWN]: 'showdown',
  [DramahaPhase.END]: 'end',
  [DramahaPhase.REBUY]: 'rebuy',
};

/** Renders the Dramaha game page: five hole cards, a draw round, and a pot that always splits. */
export const DramahaPage = withTutorial(DramahaPageContent, 'dramaha', DR_TUTORIAL_STEPS);
/** Inner content of the Dramaha page, wrapped by TutorialProvider. */
function DramahaPageContent() {
  // Wired explicitly rather than through `useCommunityPokerGame`: that hook
  // types `exec` as the shared Hold'em command set, which has no `draw` and no
  // `indices`. Pineapple — the other community-poker variant with an extra
  // phase and an extra command — is wired the same way for the same reason.
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('dramaha');
  const phaseNames = usePhaseNames('dramaha', DRAMAHA_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();
  const isLargeDesktop = useIsLargeDesktop();
  const { state, loading, error, exec: execApi, retry } = useGameApi(dramahaApi.exec);
  const [betAmount, setBetAmount] = useState(20);
  const [learningMode, setLearningMode] = useState(false);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('dramaha', state);
  const turnStartRef = useRef(0);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('dramaha');
  const cliConfig: CliGameConfig<DramahaResponse, Parameters<typeof dramahaApi.exec>> = useMemo(
    () => ({
      gameName: 'dramaha',
      parseCommand: parseDramahaCommand,
      formatResponse: formatDramahaState,
      helpText: DRAMAHA_HELP,
      localCommand: hintLocalCommand(hint),
    }),
    [hint],
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

  const phase = state?.phase ?? DramahaPhase.INIT;
  const isActive = phase >= DramahaPhase.PRE_FLOP && phase <= DramahaPhase.RIVER;
  const isShowdown = phase === DramahaPhase.SHOWDOWN || phase === DramahaPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const canAct =
    isActive && !humanPlayer?.folded && !humanPlayer?.allIn && state?.currentTurn === humanPlayer?.id && !!humanPlayer;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;
  const isMuckPhase = phase === DramahaPhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase = phase === DramahaPhase.REBUY && state?.rebuyPhaseType === DramahaRebuyPhaseType.REBUY;
  const isAddonPhase = phase === DramahaPhase.REBUY && state?.rebuyPhaseType === DramahaRebuyPhaseType.ADDON;
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

  // The draw round sits between the flop betting and the turn. `isActive` from
  // the shared hook covers PRE_FLOP..RIVER only, so the draw phase is
  // deliberately outside it: no betting is legal here.
  const isDrawPhase = phase === DramahaPhase.DRAW;
  const canDraw = isDrawPhase && !!humanPlayer && !humanPlayer.folded;
  const [selectedDraw, setSelectedDraw] = useState<number[]>([]);

  // Drop a stale selection when the draw round ends, so the cards the player
  // ticked cannot reappear as selected on a later hand.
  useEffect(() => {
    if (!isDrawPhase) setSelectedDraw([]);
  }, [isDrawPhase]);

  const toggleDraw = useCallback(
    (idx: number) => {
      if (!canDraw) return;
      setSelectedDraw((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx].sort()));
    },
    [canDraw],
  );

  const submitDraw = useCallback(
    (indices: number[]) => {
      // Standing pat and exchanging are the same call with an empty list —
      // exactly as the backend reads it, which treats omitted and empty alike.
      void execApi('draw', undefined, { indices });
      setSelectedDraw([]);
    },
    [execApi],
  );

  // Both halves of the split, from the same five cards. Recomputed on every
  // board change: the Omaha half moves with the board, the draw half never
  // does. Hidden for a folded seat, which is playing for neither half.
  const hands = useMemo(() => {
    if (!humanPlayer || humanPlayer.folded) return { omaha: null, draw: null };
    return dramahaHands(humanPlayer.cards ?? [], state?.communityCards ?? []);
  }, [humanPlayer, state?.communityCards]);

  // At showdown, ring the five cards that took the Omaha half. The draw half is
  // all five hole cards by definition, so there is nothing to single out there.
  const showdownBest5 = useMemo(() => {
    const empty = { holeSet: new Set<number>(), boardSet: new Set<number>() };
    if (!isShowdown || !hands.omaha) return empty;
    return { holeSet: new Set(hands.omaha.holeIdx), boardSet: new Set(hands.omaha.boardIdx) };
  }, [isShowdown, hands.omaha]);

  // Which half (or halves) each seat took, straight from the round results.
  // `hiWonAmount` is the Omaha half and `lowWonAmount` the draw half — the
  // backend reuses the Hi-Lo field names for Dramaha's two sides.
  const splitResults = useMemo(
    () =>
      (state?.roundResults ?? []).map((r) => ({
        playerIdx: r.playerIdx,
        omahaWon: (r.hiWonAmount ?? 0) > 0,
        drawWon: (r.lowWonAmount ?? 0) > 0,
      })),
    [state?.roundResults],
  );

  if (!state)
    return (
      <GameSkeleton
        gameKey="dramaha"
        layout={{
          kind: 'community-poker',
          community: 5,
          opponents: 3,
          opponentCards: DRAMAHA_HOLE_CARDS,
          footerHandSize: DRAMAHA_HOLE_CARDS,
        }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.dramaha')}
      gameThemeBg={gameTheme.dramaha.bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct || canDraw}
      gamePath="/dramaha"
      gameEndFlag={phase === DramahaPhase.END}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="dr-pot-display">
            {tc('label.pot')} <strong>{state?.pot ?? 0}</strong>
          </span>
          <span>
            {tc('label.blinds')}{' '}
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
          <div className={`relative flex-1 overflow-y-auto pt-4 px-5 lg:px-8 ${lgCardAreaConstraint}`}>
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
                              className={`transition-all ${
                                inBest ? '-translate-y-1 ring-2 ring-ds-success motion-safe:animate-pulse' : ''
                              } ${dim ? 'opacity-50' : ''}`}
                              data-best5-board={inBest || undefined}
                            >
                              <AnimatedCard card={card} width={cardWidth} style={placeholderCardStyle} />
                            </div>
                          );
                        })
                      : Array.from({ length: 5 }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
                  </div>
                </>
              );
              const cpuPlayerCards = cpuPlayers.map((player) => (
                <CpuPlayerCard
                  key={player.id}
                  player={player}
                  showCards={isShowdown}
                  faceDownCount={DRAMAHA_HOLE_CARDS}
                  showHandName={isShowdown}
                  extraInfo={
                    player.totalHands > 0 ? (
                      <HudStats
                        namespace="dramaha"
                        vpip={player.vpip}
                        pfr={player.pfr}
                        threeBet={player.threeBet}
                        af={player.af}
                      />
                    ) : undefined
                  }
                />
              ));

              if (isLargeDesktop) {
                return (
                  <PokerTableLayout
                    communityCardsTutorial="dr-community-cards"
                    cpuAreaTutorial="dr-cpu-area"
                    communityCards={communityCardsContent}
                    cpuPlayers={cpuPlayerCards}
                  />
                );
              }

              return (
                <>
                  <div
                    className={`mb-4 ${isMobile ? 'sticky top-0 z-10 bg-game-bg-green-poker pb-1 shadow-sm' : ''}`}
                    data-tutorial="dr-community-cards"
                  >
                    {communityCardsContent}
                  </div>
                  <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="dr-cpu-area">
                    {cpuPlayerCards}
                  </CpuAccordion>
                </>
              );
            })()}

            {/* CPU actions: toast on mobile, inline log on desktop */}
            {isMobile ? <CpuActionToast actions={state?.cpuActions} /> : <CpuActionLog actions={state?.cpuActions} />}

            {/* Round results */}
            {isShowdown && <RoundResults results={state?.roundResults} players={state?.players ?? []} />}

            {/* Which half of the split each seat took. The pot always divides,
                so "who won" is two answers, not one. */}
            {isShowdown && splitResults.length > 0 && (
              <div className="bg-black/30 rounded p-2 mb-3 text-white text-xs" data-testid="dramaha-split-results">
                {splitResults.map((r) => {
                  const halves = [
                    ...(r.omahaWon ? [t('result.omahaHalf')] : []),
                    ...(r.drawWon ? [t('result.drawHalf')] : []),
                  ];
                  return (
                    <div key={r.playerIdx} data-testid={`dramaha-split-result-${r.playerIdx}`}>
                      {state.players[r.playerIdx]?.isHuman ? tc('player.you') : `CPU ${r.playerIdx}`}
                      {': '}
                      {halves.length > 0 ? t('result.wonHalves', { halves: halves.join(' + ') }) : t('result.wonNone')}
                      {r.omahaWon && r.drawWon && (
                        <span className="ml-1 text-ds-warning font-bold" data-testid="dramaha-scoop">
                          {t('result.scoop')}
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
            )}

            {/* Action log */}
            <ActionLogSection
              isEndPhase={!!state?.gameEndFlag}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Sticky footer: player hand + buttons */}
          <GameFooter className={`${gameTheme.dramaha.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="dr-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
                  {t('yourHand')}
                  <span className="ml-3 text-xs">
                    {tc('betting.chips')} {humanPlayer.chips}
                  </span>
                  {humanPlayer.totalHands > 0 && (
                    <HudStats
                      namespace="dramaha"
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
                </div>
                <div
                  className={`mb-1 inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-semibold ${badgeInfoColors}`}
                  data-testid="dramaha-split-rule-badge"
                  data-tutorial="dr-split-rule"
                  title={t('splitRuleAria')}
                >
                  <span aria-hidden="true">🔀</span>
                  {t('splitRule')}
                </div>
                {/* Both halves, side by side. Showing one alone would hide half
                    of what the player is playing for. */}
                <div className="mb-1 flex flex-wrap gap-2" data-testid="dramaha-hands">
                  <div className="flex items-center gap-1" data-testid="dramaha-omaha-hand">
                    <span className="text-ds-text-primary text-xs">
                      {t('omahaHand')} <span className="text-ds-text-muted">({t('omahaHandRule')})</span>
                    </span>
                    <span
                      className={`inline-block text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}
                      data-testid="dramaha-omaha-hand-name"
                    >
                      {hands.omaha ? t(`hand.${hands.omaha.key}`) : t('handsUnavailable')}
                    </span>
                  </div>
                  <div className="flex items-center gap-1" data-testid="dramaha-draw-hand">
                    <span className="text-ds-text-primary text-xs">
                      {t('drawHand')} <span className="text-ds-text-muted">({t('drawHandRule')})</span>
                    </span>
                    <span
                      className={`inline-block text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}
                      data-testid="dramaha-draw-hand-name"
                    >
                      {hands.draw ? t(`hand.${hands.draw.key}`) : t('handsUnavailable')}
                    </span>
                  </div>
                </div>
                <div className="flex flex-wrap gap-1.5 mb-2">
                  {humanPlayer.cards?.length
                    ? humanPlayer.cards.map((card, idx) => {
                        const isSelected = selectedDraw.includes(idx);
                        const inBest = showdownBest5.holeSet.has(idx);
                        const dim = showdownBest5.holeSet.size > 0 && !inBest;
                        return (
                          <button
                            type="button"
                            key={`${card.design}-${card.value}`}
                            onClick={() => toggleDraw(idx)}
                            disabled={!canDraw || loading}
                            aria-pressed={canDraw ? isSelected : undefined}
                            className={`transition-all ${canDraw ? 'cursor-pointer' : 'cursor-default'} ${
                              inBest ? 'rounded-lg ring-2 ring-ds-success motion-safe:animate-pulse' : ''
                            } ${dim ? 'opacity-50' : ''}`}
                            style={selectedCardStyle(canDraw && isSelected)}
                            data-best5-hole={inBest || undefined}
                            data-testid={`dramaha-hole-card-${idx}`}
                          >
                            <AnimatedCard card={card} width={cardWidth} />
                          </button>
                        );
                      })
                    : !humanPlayer.folded &&
                      Array.from({ length: DRAMAHA_HOLE_CARDS }).map((_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth} />
                      ))}
                </div>
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

            {/* Draw controls */}
            {canDraw && (
              <div className="mb-2 text-center" data-testid="draw-controls" data-tutorial="dr-draw-round">
                <p className="text-ds-text-primary font-bold mb-1">{t('draw.title')}</p>
                <p className="text-ds-text-primary mb-1">{t('draw.prompt')}</p>
                <p className="text-ds-text-muted text-xs mb-2" data-testid="dramaha-draw-once">
                  {t('draw.onceOnly')}
                </p>
                <p className="text-ds-text-primary text-xs mb-2" data-testid="dramaha-draw-selected">
                  {t('draw.selected', { count: selectedDraw.length })}
                </p>
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={`${btnPrimary} min-w-[90px]`}
                    disabled={loading || selectedDraw.length === 0}
                    onClick={() => submitDraw([...selectedDraw])}
                    data-testid="dramaha-draw-exchange"
                  >
                    {t('draw.exchange', { count: selectedDraw.length })}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => submitDraw([])}
                    data-testid="dramaha-draw-standpat"
                  >
                    {t('draw.standPat')}
                  </button>
                </div>
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
                    onClick={() => execApi('muck')}
                  >
                    {t('muck.muck')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => execApi('show')}
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
                    onClick={() => execApi('rebuy')}
                  >
                    {t('rebuy.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => execApi('skiprebuy')}
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
                    onClick={() => execApi('addon')}
                  >
                    {t('addon.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => execApi('skipaddon')}
                  >
                    {t('addon.skip')}
                  </button>
                </div>
              </div>
            )}

            {/* Hint */}
            {hintEnabled && hint && (
              <HintTooltip reason={t(hint.reason, hint.reasonParams)} confidence={hint.confidence} />
            )}

            {/* Betting controls */}
            {canAct && (
              <div data-tutorial="dr-action-buttons">
                <BettingControls
                  inputId="dramahaBetAmount"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={minRaise}
                  maxBetAmount={state?.maxBetAmount}
                  potSize={state?.pot}
                  hasOutstandingBet={hasOutstandingBet}
                  loading={loading}
                  onCall={() => execApi('call', undefined, undefined, getElapsed())}
                  onRaise={() => execApi('raise', betAmount, undefined, getElapsed())}
                  onBet={() => execApi('bet', betAmount, undefined, getElapsed())}
                  onCheck={() => execApi('check', undefined, undefined, getElapsed())}
                  onFold={() => execApi('fold', undefined, undefined, getElapsed())}
                  onAllIn={() => execApi('allin', undefined, undefined, getElapsed())}
                />
              </div>
            )}

            {/* Settings + Reset */}
            <details className="mb-1" open={learningMode || undefined}>
              <summary className="cursor-pointer select-none text-ds-text-primary text-sm font-bold py-1">
                {tc('settings.title')}
              </summary>
              <div className="flex flex-col gap-2 py-1">
                <div className="flex items-center gap-2" data-testid="learning-mode-toggle">
                  <label
                    htmlFor="learningModeCheckbox"
                    className="text-ds-text-primary text-sm cursor-pointer flex items-center gap-2 min-h-[44px]"
                  >
                    {t('learning.toggle')}
                    <input
                      id="learningModeCheckbox"
                      type="checkbox"
                      checked={learningMode}
                      onChange={(e) => setLearningMode(e.target.checked)}
                    />
                  </label>
                </div>
                {learningMode && state?.equity && state.potOdds != null && (
                  <EquityDisplay equity={state.equity} potOdds={state.potOdds} />
                )}
                <div className="flex items-center gap-3">
                  <label className="text-ds-text-primary text-sm flex items-center gap-1 min-h-[44px]">
                    <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
                    {tc('hint.toggle', { ns: 'tutorial' })}
                  </label>
                  <label className="text-ds-text-primary text-sm flex items-center gap-1 min-h-[44px]">
                    <input type="checkbox" checked={cpuMetaAI} onChange={(e) => setCpuMetaAI(e.target.checked)} />
                    {t('settings.cpuMetaAI')}
                  </label>
                </div>
              </div>
            </details>
            <GameResetButton
              isGameEnd={phase === DramahaPhase.SHOWDOWN || phase === DramahaPhase.END}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
              dataTutorial="dr-reset-button"
              className="min-w-[90px]"
            />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
