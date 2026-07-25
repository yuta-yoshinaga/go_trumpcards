import { useMemo } from 'react';
import { omahaApi } from '../api/gameApi';
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
import { useCommunityPokerGame } from '../hooks/useCommunityPokerGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { placeholderCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { OmahaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { OMAHA_HELP, parseOmahaCommand } from '../utils/cli/commands/omahaCommands';
import { formatOmahaState } from '../utils/cli/formatters/omahaFormatter';
import { omahaBestFive } from '../utils/omahaBestFive';
import { findPlayerName } from '../utils/playerUtils';
import { evaluateFiveCardHand, pokerHandKey } from '../utils/pokerSquaresUtils';

/** Omaha Hold'em tutorial step definitions. */
const OH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="oh-community-cards"]',
    messageKey: 'tutorial.communityCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-combination-rule"]',
    messageKey: 'tutorial.combinationRule',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const OMAHA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [OmahaPhase.PRE_FLOP]: 'preFlop',
  [OmahaPhase.FLOP]: 'flop',
  [OmahaPhase.TURN]: 'turn',
  [OmahaPhase.RIVER]: 'river',
  [OmahaPhase.SHOWDOWN]: 'showdown',
  [OmahaPhase.END]: 'end',
  [OmahaPhase.REBUY]: 'rebuy',
};

/** Renders the Omaha Hold'em game page with community cards, betting, and showdown. */
export const OmahaPage = withTutorial(OmahaPageContent, 'omaha', OH_TUTORIAL_STEPS);
/** Inner content of the Omaha Hold'em page, wrapped by TutorialProvider. */
function OmahaPageContent() {
  const {
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
    canAct,
    hasOutstandingBet,
    minRaise,
    isMuckPhase,
    isRebuyPhase,
    isAddonPhase,
    humanRebuyCount,
    cpuPlayers,
  } = useCommunityPokerGame({
    game: 'omaha',
    exec: omahaApi.exec,
    phaseKeys: OMAHA_PHASE_KEYS,
    cli: { parseCommand: parseOmahaCommand, formatResponse: formatOmahaState, helpText: OMAHA_HELP },
  });

  // At showdown, highlight the human's winning 5 cards under Omaha's
  // must-use-exactly-2-hole + 3-board rule (dim the rest).
  const showdownBest5 = useMemo(() => {
    const empty = { holeSet: new Set<number>(), boardSet: new Set<number>() };
    if (!isShowdown || !humanPlayer || humanPlayer.folded) return empty;
    const hole = humanPlayer.cards ?? [];
    const board = state?.communityCards ?? [];
    const best = omahaBestFive(hole, board);
    if (!best) return empty;
    return { holeSet: new Set(best.holeIdx), boardSet: new Set(best.boardIdx) };
  }, [isShowdown, humanPlayer, state?.communityCards]);
  // During play (flop..river), preview the human's current best hand name under
  // Omaha's must-use-exactly-2-hole + 3-board rule. Returns null pre-flop (fewer
  // than 3 board cards) or at showdown (where the winning hand is already shown).
  const liveBestHandKey = useMemo(() => {
    if (!isActive || isShowdown || !humanPlayer || humanPlayer.folded) return null;
    const hole = humanPlayer.cards ?? [];
    const board = state?.communityCards ?? [];
    const best = omahaBestFive(hole, board);
    if (!best) return null;
    const five = [...best.holeIdx.map((i) => hole[i]), ...best.boardIdx.map((i) => board[i])];
    const rank = evaluateFiveCardHand(five);
    return rank == null ? null : pokerHandKey(rank);
  }, [isActive, isShowdown, humanPlayer, state?.communityCards]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="omaha"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 4, footerHandSize: 4 }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.omaha')}
      gameThemeBg={gameTheme.omaha.bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct}
      gamePath="/omaha"
      gameEndFlag={phase === OmahaPhase.END}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="oh-pot-display">
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
                  faceDownCount={4}
                  showHandName={isShowdown}
                  extraInfo={
                    player.totalHands > 0 ? (
                      <HudStats
                        namespace="omaha"
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
                    communityCardsTutorial="oh-community-cards"
                    cpuAreaTutorial="oh-cpu-area"
                    communityCards={communityCardsContent}
                    cpuPlayers={cpuPlayerCards}
                  />
                );
              }

              return (
                <>
                  <div
                    className={`mb-4 ${isMobile ? 'sticky top-0 z-10 bg-game-bg-green-poker pb-1 shadow-sm' : ''}`}
                    data-tutorial="oh-community-cards"
                  >
                    {communityCardsContent}
                  </div>
                  <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="oh-cpu-area">
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
          <GameFooter className={`${gameTheme.omaha.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="oh-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
                  {t('yourHand')}
                  <span className="ml-3 text-xs">
                    {tc('betting.chips')} {humanPlayer.chips}
                  </span>
                  {humanPlayer.totalHands > 0 && (
                    <HudStats
                      namespace="omaha"
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
                <div
                  className="mb-1 inline-flex items-center gap-1.5 rounded-full bg-ds-info/20 px-2 py-0.5 text-[11px] font-semibold text-ds-info"
                  data-testid="omaha-rule-badge"
                  title={t('mandatoryRuleAria')}
                >
                  <span aria-hidden="true">🎯</span>
                  {t('mandatoryRule')}
                </div>
                {liveBestHandKey && (
                  <div className="mb-1" data-testid="omaha-live-besthand">
                    <span className="text-ds-text-primary text-xs">{t('livePreview')}</span>
                    <span
                      className={`inline-block ml-1.5 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}
                      data-testid="omaha-live-besthand-name"
                    >
                      {t(`hand.${liveBestHandKey}`)}
                    </span>
                  </div>
                )}
                <div className="flex flex-wrap gap-1.5 mb-2" data-tutorial="oh-combination-rule">
                  {humanPlayer.cards?.length
                    ? humanPlayer.cards.map((card, idx) => {
                        const inBest = showdownBest5.holeSet.has(idx);
                        const dim = showdownBest5.holeSet.size > 0 && !inBest;
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
                      Array.from({ length: 4 }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
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
              <div data-tutorial="oh-action-buttons">
                <BettingControls
                  inputId="omahaBetAmount"
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
              isGameEnd={phase === OmahaPhase.SHOWDOWN || phase === OmahaPhase.END}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
              dataTutorial="oh-reset-button"
              className="min-w-[90px]"
            />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
