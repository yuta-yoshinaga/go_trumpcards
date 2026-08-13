import { useMemo } from 'react';
import { courchevelApi } from '../api/gameApi';
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
import { badgeInfoColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { placeholderCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { OmahaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { OMAHA_HELP, parseOmahaCommand } from '../utils/cli/commands/omahaCommands';
import { formatOmahaState } from '../utils/cli/formatters/omahaFormatter';
import { omahaLivePreviewKey } from '../utils/livePokerPreview';
import { omahaBestFive } from '../utils/omahaBestFive';
import { findPlayerName } from '../utils/playerUtils';

/** Courchevel tutorial step definitions. */
const COURCHEVEL_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bo-community-cards"]',
    messageKey: 'tutorial.communityCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bo-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bo-combination-rule"]',
    messageKey: 'tutorial.combinationRule',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bo-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bo-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bo-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bo-reset-button"]',
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

/**
 * Renders the Courchevel game page with community cards, betting and showdown.
 *
 * **Courchevel is Big O with the first flop card already face up** when the
 * opening betting round starts, so the page is the Big O page: the wire shape
 * is identical and `communityCards` simply already holds one card during the
 * pre-flop phase. Nothing here needs to know about the difference — the board
 * renders whatever the server sent.
 */
export const CourchevelPage = withTutorial(CourchevelPageContent, 'courchevel', COURCHEVEL_TUTORIAL_STEPS);
/** Inner content of the Courchevel page, wrapped by TutorialProvider. */
function CourchevelPageContent() {
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
    game: 'courchevel',
    exec: courchevelApi.exec,
    phaseKeys: OMAHA_PHASE_KEYS,
    cli: { parseCommand: parseOmahaCommand, formatResponse: formatOmahaState, helpText: OMAHA_HELP },
  });

  // At showdown, highlight the human's winning 5 cards under Courchevel's
  // must-use-exactly-2-hole (of 5) + 3-board rule.
  const showdownBest5 = useMemo(() => {
    const empty = { holeSet: new Set<number>(), boardSet: new Set<number>() };
    if (!isShowdown || !humanPlayer || humanPlayer.folded) return empty;
    const best = omahaBestFive(humanPlayer.cards ?? [], state?.communityCards ?? []);
    if (!best) return empty;
    return { holeSet: new Set(best.holeIdx), boardSet: new Set(best.boardIdx) };
  }, [isShowdown, humanPlayer, state?.communityCards]);

  // Preview the hand the player currently holds under the must-use-exactly-2
  // rule. Courchevel deals five hole cards — ten pairings to weigh by eye — so the
  // preview matters more here than in plain Omaha, which already had it (#4681/#4682).
  const liveBestHandKey = useMemo(
    () => omahaLivePreviewKey(humanPlayer, state?.communityCards ?? [], { isActive, isShowdown }),
    [isActive, isShowdown, humanPlayer, state?.communityCards],
  );

  if (!state)
    return (
      <GameSkeleton
        gameKey="courchevel"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 5, footerHandSize: 5 }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.courchevel')}
      gameThemeBg={gameTheme.courchevel.bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct}
      gamePath="/courchevel"
      gameEndFlag={phase === OmahaPhase.END}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="bo-pot-display">
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
              const cpuPlayerCards = cpuPlayers.map((player) => {
                // At showdown, highlight the exactly-2 hole cards this CPU used
                // in its best hand under the must-use-2-hole + 3-board rule.
                const cpuUsedHoleIdx =
                  isShowdown && !player.folded
                    ? (omahaBestFive(player.cards ?? [], state?.communityCards ?? [])?.holeIdx ?? undefined)
                    : undefined;
                return (
                  <CpuPlayerCard
                    key={player.id}
                    player={player}
                    showCards={isShowdown}
                    faceDownCount={5}
                    showHandName={isShowdown}
                    usedHoleIdx={cpuUsedHoleIdx}
                    extraInfo={
                      player.totalHands > 0 ? (
                        <HudStats
                          namespace="courchevel"
                          vpip={player.vpip}
                          pfr={player.pfr}
                          threeBet={player.threeBet}
                          af={player.af}
                        />
                      ) : undefined
                    }
                  />
                );
              });

              if (isLargeDesktop) {
                return (
                  <PokerTableLayout
                    communityCardsTutorial="bo-community-cards"
                    cpuAreaTutorial="bo-cpu-area"
                    communityCards={communityCardsContent}
                    cpuPlayers={cpuPlayerCards}
                  />
                );
              }

              return (
                <>
                  <div
                    className={isMobile ? 'mb-4 sticky top-0 z-10 bg-game-bg-green-poker pb-1 shadow-sm' : 'mb-4'}
                    data-tutorial="bo-community-cards"
                  >
                    {communityCardsContent}
                  </div>
                  <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="bo-cpu-area">
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
          <GameFooter className={`${gameTheme.courchevel.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="bo-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
                  {t('yourHand')}
                  <span className="ml-3 text-xs">
                    {tc('betting.chips')} {humanPlayer.chips}
                  </span>
                  {humanPlayer.totalHands > 0 && (
                    <HudStats
                      namespace="courchevel"
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
                  className={`mb-1 inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-semibold ${badgeInfoColors}`}
                  data-testid="courchevel-rule-badge"
                  title={t('mandatoryRuleAria')}
                >
                  <span aria-hidden="true">🎯</span>
                  {t('mandatoryRule')}
                </div>
                {liveBestHandKey && (
                  <div className="mb-1" data-testid="courchevel-live-besthand">
                    <span className="text-ds-text-primary text-xs">{t('livePreview')}</span>
                    <span
                      className={`inline-block ml-1.5 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}
                      data-testid="courchevel-live-besthand-name"
                    >
                      {t(`hand.${liveBestHandKey}`)}
                    </span>
                  </div>
                )}
                {/* Courchevel deals 5 hole cards; on mobile keep them on a single
                    scrollable row (never wrap) so all 5 stay visible with
                    full-size tap targets. Desktop keeps the wrapping layout. */}
                <div
                  className={`gap-1.5 mb-2 ${isMobile ? 'flex flex-nowrap overflow-x-auto pb-1' : 'flex flex-wrap'}`}
                  data-testid="courchevel-hole-cards"
                  data-tutorial="bo-combination-rule"
                >
                  {humanPlayer.cards?.length
                    ? humanPlayer.cards.map((card, idx) => {
                        const inBest = showdownBest5.holeSet.has(idx);
                        const showUsage = showdownBest5.holeSet.size > 0;
                        const dim = showUsage && !inBest;
                        return (
                          <div key={`${card.design}-${card.value}`} className="flex shrink-0 flex-col items-center">
                            <div
                              className={`transition-all ${
                                inBest ? '-translate-y-1 ring-2 ring-ds-success motion-safe:animate-pulse' : ''
                              } ${dim ? 'opacity-50' : ''}`}
                              data-best5-hole={inBest || undefined}
                            >
                              <AnimatedCard card={card} width={cardWidth} style={placeholderCardStyle} />
                            </div>
                            {showUsage && (
                              <span
                                className={`mt-0.5 text-[10px] font-semibold ${inBest ? 'text-ds-success' : 'text-ds-text-muted'}`}
                                data-testid={inBest ? 'courchevel-hole-used' : 'courchevel-hole-unused'}
                              >
                                {inBest ? t('cardUsed') : t('cardUnused')}
                              </span>
                            )}
                          </div>
                        );
                      })
                    : !humanPlayer.folded &&
                      Array.from({ length: 5 }).map((_, i) => (
                        <div key={i} className="shrink-0">
                          <AnimatedCardBack width={cardWidth} />
                        </div>
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
              <div data-tutorial="bo-action-buttons">
                <BettingControls
                  inputId="courchevelBetAmount"
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
              isGameEnd={phase === OmahaPhase.SHOWDOWN || phase === OmahaPhase.END}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
              dataTutorial="bo-reset-button"
              className="min-w-[90px]"
            />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
