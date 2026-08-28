import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { shortdeckApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
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
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarning } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { placeholderCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ShortDeckResponse } from '../types/card';
import { HoldemPhase, HoldemRebuyPhaseType } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseShortdeckCommand, SHORTDECK_HELP } from '../utils/cli/commands/shortdeckCommands';
import { formatShortdeckState } from '../utils/cli/formatters/shortdeckFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName } from '../utils/playerUtils';
import { shortDeckBestFive } from '../utils/shortDeckBestFive';

/** Short Deck Hold'em tutorial step definitions. */
const SD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sd-community-cards"]',
    messageKey: 'tutorial.communityCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-learning-mode"]',
    messageKey: 'tutorial.learningMode',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SHORTDECK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [HoldemPhase.PRE_FLOP]: 'preFlop',
  [HoldemPhase.FLOP]: 'flop',
  [HoldemPhase.TURN]: 'turn',
  [HoldemPhase.RIVER]: 'river',
  [HoldemPhase.SHOWDOWN]: 'showdown',
  [HoldemPhase.END]: 'end',
  [HoldemPhase.REBUY]: 'rebuy',
};

/** Renders the Short Deck Hold'em game page with community cards, betting, and showdown. */
export const ShortDeckPage = withTutorial(ShortDeckPageContent, 'shortdeck', SD_TUTORIAL_STEPS);
/** Inner content of the Short Deck Hold'em page, wrapped by TutorialProvider. */
function ShortDeckPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('shortdeck');
  const phaseNames = usePhaseNames('shortdeck', SHORTDECK_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();
  const { state, loading, error, exec: execApi, retry } = useGameApi(shortdeckApi.exec);
  const [betAmount, setBetAmount] = useState(20);
  const [learningMode, setLearningMode] = useState(false);
  // Touch/keyboard-accessible toggle for the ★ rank-override note (title alone is hover-only).
  const [showRuleNote, setShowRuleNote] = useState(false);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  // The page already renders hand/level progress when tournamentMode is on, but
  // nothing could turn it on -- the state was displayable and unreachable.
  const [tournamentMode, setTournamentMode] = useState(false);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('shortdeck', state);
  const turnStartRef = useRef(0);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('shortdeck');
  const cliConfig: CliGameConfig<ShortDeckResponse, Parameters<typeof shortdeckApi.exec>> = useMemo(
    () => ({
      gameName: 'shortdeck',
      parseCommand: parseShortdeckCommand,
      formatResponse: formatShortdeckState,
      helpText: SHORTDECK_HELP,
      localCommand: hintLocalCommand(hint),
    }),
    [hint],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void execApi('reset', undefined, { cpuMetaAI, tournamentMode });
  }, [execApi, hideActionLog, cpuMetaAI, tournamentMode]);

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

  const phase = state?.phase ?? HoldemPhase.INIT;
  const isActive = phase >= HoldemPhase.PRE_FLOP && phase <= HoldemPhase.RIVER;
  const isShowdown = phase === HoldemPhase.SHOWDOWN || phase === HoldemPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isActive && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  // Short Deck reorders the hand rankings, so the standard holdemBestFive would
  // sometimes mark five cards the server did not score — most visibly on the
  // A-6-7-8-9 wheel, which it cannot see at all (#4684).
  const showdownBest5 = useMemo(() => {
    const empty = { holeSet: new Set<number>(), boardSet: new Set<number>() };
    if (!isShowdown || !humanPlayer || humanPlayer.folded) return empty;
    const hole = humanPlayer.cards ?? [];
    const board = state?.communityCards ?? [];
    if (hole.length !== 2 || board.length < 5) return empty;
    const combined = [...hole, ...board.slice(0, 5)];
    const holeSet = new Set<number>();
    const boardSet = new Set<number>();
    for (const i of shortDeckBestFive(combined) ?? []) {
      if (i < hole.length) holeSet.add(i);
      else boardSet.add(i - hole.length);
    }
    return { holeSet, boardSet };
  }, [isShowdown, humanPlayer, state?.communityCards]);

  const minRaise = state?.minRaise ?? 0;
  const isMuckPhase = phase === HoldemPhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase = phase === HoldemPhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.REBUY;
  const isAddonPhase = phase === HoldemPhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.ADDON;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const humanRebuyCount = state?.rebuyCounts?.[humanIdx] ?? 0;
  const cpuPlayers = useMemo(() => state?.players?.filter((p) => !p.isHuman) ?? [], [state?.players]);

  const actionBindings = useMemo(
    () => [
      {
        key: 'c',
        action: () => execApi('call', undefined, undefined, getElapsed()),
        enabled: hasOutstandingBet,
        label: 'call',
      },
      {
        key: 'r',
        action: () =>
          hasOutstandingBet
            ? execApi('raise', betAmount, undefined, getElapsed())
            : execApi('bet', betAmount, undefined, getElapsed()),
        label: 'raiseOrBet',
      },
      {
        key: 'k',
        action: () => execApi('check', undefined, undefined, getElapsed()),
        enabled: !hasOutstandingBet,
        label: 'check',
      },
      { key: 'f', action: () => execApi('fold', undefined, undefined, getElapsed()), label: 'fold' },
      { key: 'a', action: () => execApi('allin', undefined, undefined, getElapsed()), label: 'allin' },
    ],
    [execApi, hasOutstandingBet, betAmount, getElapsed],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: canAct && !loading,
  });

  if (!state)
    return (
      <GameSkeleton
        gameKey="shortdeck"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 2, footerHandSize: 2 }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.shortdeck')}
      gameThemeBg={gameTheme.shortdeck.bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct}
      gamePath="/shortdeck"
      gameEndFlag={phase === HoldemPhase.SHOWDOWN || phase === HoldemPhase.END}
      winShow={phase === HoldemPhase.END}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="sd-pot-display">
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
                  <div className="text-ds-text-primary text-lg mb-1.5 flex items-center gap-2 flex-wrap">
                    <span>{t('communityCards')}</span>
                    {/* Use the project's badgeWarning state-token style so the contrast ratios
                        in DESIGN.md hold on the green poker felt — mixing the warning token with
                        Tailwind opacity suffixes silently breaks WCAG AA. The size overrides shadow
                        the default px/py/text-size for a compact in-line chip. */}
                    <span
                      data-testid="shortdeck-rank-watermark"
                      className={`${badgeWarning} px-2 py-0.5 text-[11px] uppercase tracking-wider`}
                      title={t('rankOverrideReminder')}
                    >
                      {t('rankWatermark')}
                    </span>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {state?.communityCards?.length
                      ? state.communityCards.map((card, idx) => {
                          const inBest = showdownBest5.boardSet.has(idx);
                          const dim = isShowdown && showdownBest5.boardSet.size > 0 && !inBest;
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
              const cpuPlayerCards = cpuPlayers.map((p) => (
                <CpuPlayerCard
                  key={p.id}
                  player={p}
                  showCards={isShowdown}
                  faceDownCount={2}
                  showHandName={isShowdown}
                  extraInfo={
                    p.totalHands > 0 ? (
                      <HudStats namespace="shortdeck" vpip={p.vpip} pfr={p.pfr} threeBet={p.threeBet} af={p.af} />
                    ) : undefined
                  }
                />
              ));

              if (!isMobile) {
                return (
                  <PokerTableLayout
                    communityCardsTutorial="sd-community-cards"
                    cpuAreaTutorial="sd-cpu-area"
                    communityCards={communityCardsContent}
                    cpuPlayers={cpuPlayerCards}
                  />
                );
              }

              return (
                <>
                  <div
                    className="sticky top-0 z-10 bg-game-bg-green-poker pb-1 shadow-sm"
                    data-tutorial="sd-community-cards"
                  >
                    {communityCardsContent}
                  </div>
                  <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="sd-cpu-area">
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
          <GameFooter className={`${gameTheme.shortdeck.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="sd-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
                  {t('yourHand')}
                  <span className="ml-3 text-xs">
                    {tc('betting.chips')} {humanPlayer.chips}
                  </span>
                  {humanPlayer.totalHands > 0 && (
                    <HudStats
                      namespace="shortdeck"
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
                    <span
                      className={`inline-flex items-center gap-1 ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}
                    >
                      {humanPlayer.handName}
                      {/* Button (not a hover-only title span) so touch + keyboard users can
                          read the rank-override note; desktop hover title is preserved. */}
                      <button
                        type="button"
                        data-testid="shortdeck-handname-rule"
                        aria-label={t('rankOverrideReminder')}
                        aria-expanded={showRuleNote}
                        title={t('rankOverrideReminder')}
                        onClick={() => setShowRuleNote((v) => !v)}
                        className="cursor-help text-ds-warning"
                      >
                        ★
                      </button>
                    </span>
                  )}
                  {isShowdown && !humanPlayer.folded && humanPlayer.handName && showRuleNote && (
                    <span className="ml-2 text-xs text-ds-warning" data-testid="shortdeck-rule-note" role="status">
                      {t('rankOverrideReminder')}
                    </span>
                  )}
                </div>
                <div className="flex flex-wrap gap-1.5 mb-2">
                  {humanPlayer.cards?.length
                    ? humanPlayer.cards.map((card, idx) => {
                        const inBest = showdownBest5.holeSet.has(idx);
                        const dim = isShowdown && showdownBest5.holeSet.size > 0 && !inBest;
                        return (
                          <div
                            key={`${card.design}-${card.value}`}
                            className={`transition-all ${
                              inBest ? '-translate-y-1 ring-2 ring-ds-success motion-safe:animate-pulse' : ''
                            } ${dim ? 'opacity-50' : ''}`}
                            data-best5-hole={inBest || undefined}
                          >
                            <AnimatedCard card={card} width={cardWidth} style={placeholderCardStyle} />
                          </div>
                        );
                      })
                    : !humanPlayer.folded &&
                      Array.from({ length: 2 }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
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
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {/* Betting controls */}
            {canAct && (
              <div data-tutorial="sd-action-buttons">
                <BettingControls
                  inputId="shortdeckBetAmount"
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
            <details className="mb-1" data-tutorial="sd-learning-mode" open={learningMode || undefined}>
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
                  <label className="text-ds-text-primary text-sm flex items-center gap-1 min-h-[44px]">
                    <input
                      type="checkbox"
                      checked={tournamentMode}
                      data-testid="sd-tournament-toggle"
                      onChange={(e) => setTournamentMode(e.target.checked)}
                    />
                    {t('settings.tournamentMode')}
                  </label>
                </div>
              </div>
            </details>
            <details className="mb-1" data-testid="sd-handrank-reference">
              <summary className="cursor-pointer select-none text-ds-text-primary text-sm font-bold py-1">
                {t('handRank.title')}
              </summary>
              <ol className="list-decimal list-inside text-ds-text-muted text-xs py-1 space-y-0.5">
                <li>{t('handRank.straightFlush')}</li>
                <li>{t('handRank.fourOfAKind')}</li>
                <li className="text-ds-text-primary font-semibold">{t('handRank.flush')}</li>
                <li className="text-ds-text-primary font-semibold">{t('handRank.fullHouse')}</li>
                <li>{t('handRank.straight')}</li>
                <li>{t('handRank.threeOfAKind')}</li>
                <li>{t('handRank.twoPair')}</li>
                <li>{t('handRank.onePair')}</li>
                <li>{t('handRank.highCard')}</li>
              </ol>
              <p className="text-ds-info text-xs pt-1">{t('handRank.note')}</p>
            </details>
            <GameResetButton
              isGameEnd={phase === HoldemPhase.SHOWDOWN || phase === HoldemPhase.END}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
              dataTutorial="sd-reset-button"
              className="min-w-[90px]"
            />
            <ActionShortcutsPanel bindings={actionBindings} data-testid="short-deck-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
