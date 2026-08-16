import { motion } from 'framer-motion';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { indianpokerApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { BettingControls } from '../components/BettingControls';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuActionToast } from '../components/CpuActionToast';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
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
import { useReducedMotion } from '../hooks/useReducedMotion';
import { useSound } from '../providers/SoundProvider';
import { placeholderCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { flipSpring } from '../styles/motionPresets';
import type { IndianPokerResponse } from '../types/card';
import { IndianPokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { INDIANPOKER_HELP, parseIndianpokerCommand } from '../utils/cli/commands/indianpokerCommands';
import { formatIndianpokerState } from '../utils/cli/formatters/indianpokerFormatter';
import { hintCliText, isHintCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { computeIndianPokerEquity } from '../utils/indianPokerEquity';
import { findPlayerName } from '../utils/playerUtils';

/** Indian Poker tutorial step definitions. */
const IP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ip-cpu-cards"]',
    messageKey: 'tutorial.cpuCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ip-player-card"]',
    messageKey: 'tutorial.playerCard',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ip-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ip-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const INDIAN_POKER_PHASE_KEYS: Readonly<Record<number, string>> = {
  [IndianPokerPhase.INIT]: 'init',
  [IndianPokerPhase.ANTE]: 'ante',
  [IndianPokerPhase.BETTING]: 'betting',
  [IndianPokerPhase.SHOWDOWN]: 'showdown',
  [IndianPokerPhase.END]: 'end',
};

/** Renders the Indian Poker game page with opponent cards visible and human card hidden. */
export const IndianPokerPage = withTutorial(IndianPokerPageContent, 'indianpoker', IP_TUTORIAL_STEPS);
/** Inner content of the Indian Poker page, wrapped by TutorialProvider. */
function IndianPokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('indianpoker');
  const phaseNames = usePhaseNames('indianpoker', INDIAN_POKER_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const isMobile = useIsMobile();
  const { state, loading, error, exec: execApi, retry } = useGameApi(indianpokerApi.exec);
  const [betAmount, setBetAmount] = useState(20);
  const [ante, setAnte] = useState(10);
  const [bettingLimit, setBettingLimit] = useState(2);
  const [cpuMetaAI, setCpuMetaAI] = useState(true);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('indianpoker', state);
  const turnStartRef = useRef(0);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('indianpoker');
  const cliConfig: CliGameConfig<IndianPokerResponse, Parameters<typeof indianpokerApi.exec>> = useMemo(
    () => ({
      gameName: 'indianpoker',
      parseCommand: parseIndianpokerCommand,
      formatResponse: formatIndianpokerState,
      helpText: INDIANPOKER_HELP,
      localCommand: (input: string) => (isHintCommand(input) ? hintCliText(hint) : null),
    }),
    [hint],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void execApi('reset', undefined, { ante, bettingLimit, cpuMetaAI });
  }, [execApi, hideActionLog, ante, bettingLimit, cpuMetaAI]);

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

  const phase = state?.phase ?? IndianPokerPhase.INIT;
  const isBetting = phase === IndianPokerPhase.BETTING;
  const isShowdown = phase === IndianPokerPhase.SHOWDOWN || phase === IndianPokerPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);

  // Staged showdown reveal (#3068): 0 = concealed, 1 = own card flipped face-up,
  // 2 = round results shown. In Indian Poker you never see your own card during play,
  // so at showdown we flip it first and hold back the results panel for 600ms to avoid
  // spoiling the outcome before the card is read. Keyed by hand + phase so each new
  // showdown restarts the sequence. prefers-reduced-motion jumps straight to step 2.
  const reduced = useReducedMotion();
  const revealSignature = isShowdown ? `${state?.handCount ?? 0}:${phase}` : 'hidden';
  const [revealStep, setRevealStep] = useState(0);
  // biome-ignore lint/correctness/useExhaustiveDependencies: revealSignature already encodes handCount and phase; listing raw state would re-fire the effect mid-hand.
  useEffect(() => {
    if (!isShowdown) {
      setRevealStep(0);
      return;
    }
    if (reduced) {
      setRevealStep(2);
      return;
    }
    setRevealStep(1);
    playSound('cardPlace');
    const timer = setTimeout(() => setRevealStep(2), 600);
    return () => clearTimeout(timer);
  }, [isShowdown, revealSignature, reduced, playSound]);

  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const equity = useMemo(() => {
    if (!state || isShowdown) return null;
    const opponentRanks = state.players.filter((p) => !p.isHuman && !p.folded && p.cardRank > 0).map((p) => p.cardRank);
    if (opponentRanks.length === 0) return null;
    return computeIndianPokerEquity(opponentRanks);
  }, [state, isShowdown]);
  const equityPct = equity === null ? null : Math.round(equity * 100);
  const canAct = isBetting && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;

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
        gameKey="indianpoker"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 2, footerHandSize: 2 }}
      />
    );

  // Build results with handName for RoundResults component
  const roundResultsForDisplay = state.roundResults?.map((r) => ({
    playerIdx: r.playerIdx,
    handName: r.card ? `${r.card.design} ${r.card.value}` : '',
    wonAmount: r.wonAmount,
  }));

  return (
    <GamePageShell
      title={tc('nav.indianpoker')}
      gameThemeBg={gameTheme.indianpoker.bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct}
      gamePath="/indianpoker"
      gameEndFlag={!!state.gameEndFlag}
      winShow={phase === IndianPokerPhase.END}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>
            {tc('label.pot')} <strong>{state.pot ?? 0}</strong>
          </span>
          <span>
            {t('ante')} <strong>{state.ante ?? 0}</strong>
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
      headerEnd={
        <span>
          {tc('label.dealer')} <strong>{findPlayerName(state.players, state.dealerIdx)}</strong>
        </span>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Scrollable: opponent cards + CPU players */}
          <div className={`flex-1 overflow-y-auto pt-4 px-5 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* CPU players - show cards face-up (opponents can see each other's cards) */}
            <div data-tutorial="ip-cpu-cards" className={isMobile ? 'grid grid-cols-3 gap-2 mb-3' : ''}>
              {state.players
                ?.filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className={isMobile ? 'text-center' : 'mb-3'}>
                    <div className={`text-ds-text-primary text-sm mb-1 ${isMobile ? 'truncate' : ''}`}>
                      {tc('player.cpu', { id: p.id })}
                      {!isMobile && <span className="text-ds-text-muted text-xs"> ({p.playStyleName})</span>}
                      <span className={`text-xs ${isMobile ? 'block' : 'ml-2'}`}>
                        {tc('betting.chips')} {p.chips}
                      </span>
                      {p.currentBet > 0 && (
                        <span className={`text-xs ${isMobile ? 'block' : 'ml-2'}`}>
                          {tc('betting.currentBet')} {p.currentBet}
                        </span>
                      )}
                      {p.folded && <span className="ml-1 text-ds-error text-xs">[{tc('status.folded')}]</span>}
                      {p.allIn && <span className="ml-1 text-ds-warning text-xs">[{tc('status.allIn')}]</span>}
                    </div>
                    <div className={isMobile ? 'flex justify-center' : 'flex flex-wrap gap-1'}>
                      {p.card ? (
                        <AnimatedCard card={p.card} width={cardWidth} style={placeholderCardStyle} />
                      ) : (
                        <AnimatedCardBack width={cardWidth} />
                      )}
                    </div>
                  </div>
                ))}
            </div>

            {/* CPU actions: toast on mobile, inline log on desktop */}
            {isMobile ? <CpuActionToast actions={state.cpuActions} /> : <CpuActionLog actions={state.cpuActions} />}

            {/* Round results — held back until the own-card reveal completes (#3068) */}
            {isShowdown && revealStep >= 2 && (
              <RoundResults results={roundResultsForDisplay} players={state.players ?? []} />
            )}

            {/* Action log */}
            <ActionLogSection
              isEndPhase={!!state.gameEndFlag}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Sticky footer: player card + buttons */}
          <GameFooter className={`${gameTheme.indianpoker.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="ip-player-card">
                <div className="text-ds-text-primary text-lg mb-1">
                  {t('yourCard')}
                  <span className="ml-3 text-xs">
                    {tc('betting.chips')} {humanPlayer.chips}
                  </span>
                  {humanPlayer.currentBet > 0 && (
                    <span className="ml-2 text-xs">
                      {tc('betting.currentBet')} {humanPlayer.currentBet}
                    </span>
                  )}
                  {humanPlayer.folded && <span className="ml-2 text-ds-error text-xs">[{tc('status.folded')}]</span>}
                  {humanPlayer.allIn && <span className="ml-2 text-ds-warning text-xs">[{tc('status.allIn')}]</span>}
                </div>
                <div className="flex flex-wrap gap-1.5 mb-2">
                  {isShowdown && humanPlayer.card ? (
                    <motion.div
                      key={revealSignature}
                      className="[transform-style:preserve-3d]"
                      initial={reduced ? false : { rotateY: 180, opacity: 0.5 }}
                      animate={{ rotateY: 0, opacity: 1 }}
                      transition={flipSpring}
                      data-testid="indianpoker-own-reveal"
                    >
                      <AnimatedCard card={humanPlayer.card} width={cardWidth} style={placeholderCardStyle} silent />
                    </motion.div>
                  ) : !humanPlayer.folded ? (
                    <AnimatedCardBack width={cardWidth} />
                  ) : null}
                </div>
                {hintEnabled && equityPct !== null && !humanFolded && (
                  <div
                    className="mt-1 text-xs text-ds-text-primary"
                    data-testid="indianpoker-equity-meter"
                    aria-live="polite"
                  >
                    <div className="mb-0.5 flex items-center justify-between">
                      <span>{t('equity.label')}</span>
                      <span className="font-semibold tabular-nums">{equityPct}%</span>
                    </div>
                    <div
                      className="h-1.5 w-full overflow-hidden rounded bg-black/30"
                      role="progressbar"
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-valuenow={equityPct}
                      aria-label={t('equity.label')}
                    >
                      <div
                        className={`h-full transition-all ${equityPct >= 60 ? 'bg-ds-success' : equityPct >= 35 ? 'bg-ds-warning' : 'bg-ds-error'}`}
                        style={{ width: `${equityPct}%` }}
                      />
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
              alwaysVisible
            />

            <ErrorAlert message={error} onRetry={retry} />

            {/* Hint */}
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {/* Betting controls */}
            {canAct && (
              <div data-tutorial="ip-action-buttons">
                <BettingControls
                  inputId="indianPokerBetAmount"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={minRaise}
                  maxBetAmount={state.maxBetAmount}
                  potSize={state.pot}
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

            {/* Settings */}
            <SettingsPanel
              title={t('settings.ante')}
              groups={[
                {
                  items: [
                    {
                      type: 'select' as const,
                      id: 'indianpoker-ante',
                      label: t('settings.anteAmount'),
                      value: ante,
                      options: [
                        { value: 5, label: '5' },
                        { value: 10, label: '10' },
                        { value: 20, label: '20' },
                        { value: 50, label: '50' },
                      ],
                      onSelect: (v: string) => setAnte(Number(v)),
                    },
                    {
                      type: 'select' as const,
                      id: 'indianpoker-betting-limit',
                      label: t('settings.bettingLimit'),
                      value: bettingLimit,
                      options: [
                        { value: 0, label: tc('betting.fixed') },
                        { value: 1, label: tc('betting.potLimit') },
                        { value: 2, label: tc('betting.noLimit') },
                      ],
                      onSelect: (v: string) => setBettingLimit(Number(v)),
                    },
                    {
                      type: 'checkbox' as const,
                      id: 'indianpoker-hint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: hintEnabled,
                      onToggle: setHintEnabled,
                    },
                    {
                      type: 'checkbox' as const,
                      id: 'indianpoker-meta-ai',
                      label: t('settings.metaAI'),
                      checked: cpuMetaAI,
                      onToggle: setCpuMetaAI,
                    },
                  ],
                },
              ]}
            />

            {/* Reset */}
            <div className="text-center flex items-center justify-center gap-3">
              <GameResetButton
                isGameEnd={phase === IndianPokerPhase.SHOWDOWN || phase === IndianPokerPhase.END}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="ip-reset-button"
                className="min-w-[90px]"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="indian-poker-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
