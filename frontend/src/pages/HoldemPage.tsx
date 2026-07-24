import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { type HoldemConfigInput, holdemApi } from '../api/gameApi';
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
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { badgeSuccessColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { placeholderCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { HoldemResponse } from '../types/card';
import { HoldemPhase, HoldemRebuyPhaseType } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { HOLDEM_HELP, parseHoldemCommand } from '../utils/cli/commands/holdemCommands';
import { formatHoldemState } from '../utils/cli/formatters/holdemFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { holdemBestFive } from '../utils/holdemBestFive';
import { findPlayerName } from '../utils/playerUtils';

/** Texas Hold'em tutorial step definitions. */
const HE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="he-community-cards"]',
    messageKey: 'tutorial.communityCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="he-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="he-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="he-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="he-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="he-learning-mode"]',
    messageKey: 'tutorial.learningMode',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="he-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const HOLDEM_PHASE_KEYS: Readonly<Record<number, string>> = {
  [HoldemPhase.PRE_FLOP]: 'preFlop',
  [HoldemPhase.FLOP]: 'flop',
  [HoldemPhase.TURN]: 'turn',
  [HoldemPhase.RIVER]: 'river',
  [HoldemPhase.SHOWDOWN]: 'showdown',
  [HoldemPhase.END]: 'end',
  [HoldemPhase.REBUY]: 'rebuy',
};

/** Renders the Texas Hold'em game page with community cards, betting, and showdown. */
export const HoldemPage = withTutorial(HoldemPageContent, 'holdem', HE_TUTORIAL_STEPS);
/** Inner content of the Texas Hold'em page, wrapped by TutorialProvider. */
function HoldemPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('holdem');
  const phaseNames = usePhaseNames('holdem', HOLDEM_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const isMobile = useIsMobile();
  const { state, loading, error, exec, retry } = useGameApi(holdemApi.exec);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('holdem');
  const cliConfig: CliGameConfig<HoldemResponse, Parameters<typeof holdemApi.exec>> = useMemo(
    () => ({
      gameName: 'holdem',
      parseCommand: parseHoldemCommand,
      formatResponse: formatHoldemState,
      helpText: HOLDEM_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const [betAmount, setBetAmount] = useState(20);
  const [learningMode, setLearningMode] = useState(false);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const [tournamentMode, setTournamentMode] = useState(false);
  const [rebuyEnabled, setRebuyEnabled] = useState(false);
  const [addonEnabled, setAddonEnabled] = useState(false);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('holdem', state);
  const turnStartRef = useRef(0);

  useMountReset(exec);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    const config: HoldemConfigInput = { cpuMetaAI };
    if (tournamentMode) {
      config.tournamentMode = true;
      config.rebuyEnabled = rebuyEnabled;
      config.addonEnabled = addonEnabled;
    }
    void exec('reset', undefined, config);
  }, [exec, hideActionLog, cpuMetaAI, tournamentMode, rebuyEnabled, addonEnabled]);

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
  // Best-5 highlight at showdown: enumerate hole + community → mark which of
  // the 7 visible cards contribute to the winning 5-card hand. Indices 0..1
  // map to the hole cards, 2..6 to the community board.
  const showdownBest5 = useMemo(() => {
    if (!isShowdown || !humanPlayer || humanPlayer.folded) {
      return { holeSet: new Set<number>(), boardSet: new Set<number>() };
    }
    const hole = humanPlayer?.cards ?? [];
    const board = state?.communityCards ?? [];
    if (hole.length !== 2 || board.length < 5) {
      return { holeSet: new Set<number>(), boardSet: new Set<number>() };
    }
    const combined = [...hole, ...board.slice(0, 5)];
    const picked = holdemBestFive(combined) ?? [];
    const holeSet = new Set<number>();
    const boardSet = new Set<number>();
    for (const i of picked) {
      if (i < hole.length) holeSet.add(i);
      else boardSet.add(i - hole.length);
    }
    return { holeSet, boardSet };
  }, [isShowdown, humanPlayer, state?.communityCards]);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isActive && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;
  const isMuckPhase = phase === HoldemPhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase = phase === HoldemPhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.REBUY;
  const isAddonPhase = phase === HoldemPhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.ADDON;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const humanRebuyCount = state?.rebuyCounts?.[humanIdx] ?? 0;
  const cpuPlayers = useMemo(() => state?.players?.filter((p) => !p.isHuman) ?? [], [state?.players]);

  const actionBindings = useMemo(
    () => [
      { key: 'c', action: () => exec('call', undefined, undefined, getElapsed()), enabled: hasOutstandingBet },
      {
        key: 'r',
        action: () =>
          hasOutstandingBet
            ? exec('raise', betAmount, undefined, getElapsed())
            : exec('bet', betAmount, undefined, getElapsed()),
      },
      { key: 'k', action: () => exec('check', undefined, undefined, getElapsed()), enabled: !hasOutstandingBet },
      { key: 'f', action: () => exec('fold', undefined, undefined, getElapsed()) },
      { key: 'a', action: () => exec('allin', undefined, undefined, getElapsed()) },
    ],
    [exec, hasOutstandingBet, betAmount, getElapsed],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: canAct && !loading,
  });

  if (!state)
    return (
      <GameSkeleton
        gameKey="holdem"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 2, footerHandSize: 2 }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.holdem')}
      gameThemeBg={gameTheme.holdem.bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct}
      gamePath="/holdem"
      gameEndFlag={phase === HoldemPhase.END}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="he-pot-display">
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
                  <div className="text-ds-text-primary text-lg mb-1.5">
                    {t('communityCards')}
                    {isShowdown && humanPlayer && !humanPlayer.folded && humanPlayer.handName && (
                      <span
                        className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${badgeSuccessColors}`}
                        data-testid="board-winning-hand"
                      >
                        {t('winningHand', { hand: humanPlayer.handName })}
                      </span>
                    )}
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
                      <HudStats vpip={p.vpip} pfr={p.pfr} threeBet={p.threeBet} af={p.af} />
                    ) : undefined
                  }
                />
              ));

              if (!isMobile) {
                return (
                  <PokerTableLayout
                    communityCardsTutorial="he-community-cards"
                    cpuAreaTutorial="he-cpu-area"
                    communityCards={communityCardsContent}
                    cpuPlayers={cpuPlayerCards}
                  />
                );
              }

              return (
                <>
                  <div
                    className="sticky top-0 z-10 bg-game-bg-green-poker pb-1 shadow-sm"
                    data-tutorial="he-community-cards"
                  >
                    {communityCardsContent}
                  </div>
                  <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="he-cpu-area">
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
          <GameFooter className={`${gameTheme.holdem.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="he-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
                  {t('yourHand')}
                  <span className="ml-3 text-xs">
                    {tc('betting.chips')} {humanPlayer.chips}
                  </span>
                  {humanPlayer.totalHands > 0 && (
                    <HudStats
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
                    <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${badgeSuccessColors}`}>
                      {humanPlayer.handName}
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
              severity={phase === HoldemPhase.END || phase === HoldemPhase.SHOWDOWN ? 'alert' : 'info'}
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
                    onClick={() => exec('muck')}
                  >
                    {t('muck.muck')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => exec('show')}
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
                    onClick={() => exec('rebuy')}
                  >
                    {t('rebuy.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => exec('skiprebuy')}
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
                    onClick={() => exec('addon')}
                  >
                    {t('addon.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => exec('skipaddon')}
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
              <div data-tutorial="he-action-buttons">
                <BettingControls
                  inputId="holdemBetAmount"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={minRaise}
                  maxBetAmount={state?.maxBetAmount}
                  potSize={state?.pot}
                  hasOutstandingBet={hasOutstandingBet}
                  loading={loading}
                  onCall={() => exec('call', undefined, undefined, getElapsed())}
                  onRaise={() => exec('raise', betAmount, undefined, getElapsed())}
                  onBet={() => exec('bet', betAmount, undefined, getElapsed())}
                  onCheck={() => exec('check', undefined, undefined, getElapsed())}
                  onFold={() => exec('fold', undefined, undefined, getElapsed())}
                  onAllIn={() => exec('allin', undefined, undefined, getElapsed())}
                />
              </div>
            )}

            {/* Settings + Reset */}
            <details className="mb-1" data-tutorial="he-learning-mode" open={learningMode || undefined}>
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
                <div className="flex flex-col gap-2 border-t border-ds-border pt-2" data-testid="tournament-settings">
                  <label className="text-ds-text-primary text-sm flex items-center gap-1">
                    <input
                      type="checkbox"
                      data-testid="tournament-mode-toggle"
                      checked={tournamentMode}
                      onChange={(e) => setTournamentMode(e.target.checked)}
                    />
                    {t('settings.tournamentMode')}
                  </label>
                  {tournamentMode && (
                    <div className="flex flex-col gap-2 pl-5">
                      <label className="text-ds-text-primary text-sm flex items-center gap-1">
                        <input
                          type="checkbox"
                          data-testid="tournament-rebuy-toggle"
                          checked={rebuyEnabled}
                          onChange={(e) => setRebuyEnabled(e.target.checked)}
                        />
                        {t('settings.rebuyEnabled')}
                      </label>
                      <label className="text-ds-text-primary text-sm flex items-center gap-1">
                        <input
                          type="checkbox"
                          data-testid="tournament-addon-toggle"
                          checked={addonEnabled}
                          onChange={(e) => setAddonEnabled(e.target.checked)}
                        />
                        {t('settings.addonEnabled')}
                      </label>
                    </div>
                  )}
                  <p className="text-ds-text-secondary text-xs">{t('settings.tournamentNote')}</p>
                </div>
              </div>
            </details>
            <GameResetButton
              isGameEnd={phase === HoldemPhase.SHOWDOWN || phase === HoldemPhase.END}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
              dataTutorial="he-reset-button"
              className="min-w-[90px]"
            />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
