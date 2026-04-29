import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { indianpokerApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuActionToast } from '../components/CpuActionToast';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetButton } from '../components/GameResetButton';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { RoundResults } from '../components/RoundResults';
import { HoldemSkeleton } from '../components/skeleton/HoldemSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { IndianPokerResponse } from '../types/card';
import { IndianPokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { INDIANPOKER_HELP, parseIndianpokerCommand } from '../utils/cli/commands/indianpokerCommands';
import { formatIndianpokerState } from '../utils/cli/formatters/indianpokerFormatter';
import type { CliGameConfig } from '../utils/cli/types';

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
export function IndianPokerPage() {
  return (
    <TutorialWrapper gameName="indianpoker" steps={IP_TUTORIAL_STEPS}>
      <IndianPokerPageContent />
    </TutorialWrapper>
  );
}

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
  const [ante] = useState(10);
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
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

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
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isBetting && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;

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

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: canAct && !loading,
  });

  if (!state) return <HoldemSkeleton />;

  // Build results with handName for RoundResults component
  const roundResultsForDisplay = state.roundResults?.map((r) => ({
    playerIdx: r.playerIdx,
    handName: r.card ? `${r.card.design} ${r.card.value}` : '',
    wonAmount: r.wonAmount,
  }));

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.indianpoker.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.indianpoker')} />
      {/* Phase indicator + info bar */}
      <PhaseIndicator phaseName={phaseNames[phase] ?? t('phase.init')} isHumanTurn={canAct}>
        <span>
          {tc('label.pot')} <strong>{state.pot ?? 0}</strong>
        </span>
        <span>
          {t('ante')} <strong>{state.ante ?? 0}</strong>
        </span>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/indianpoker" />
        <span>
          {tc('label.dealer')} <strong>Player {state.dealerIdx ?? 0}</strong>
        </span>
      </PhaseIndicator>

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
                        <AnimatedCard
                          card={p.card}
                          width={cardWidth}
                          style={{ border: '3px solid transparent' }}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ) : (
                        <AnimatedCardBack width={cardWidth} onFlipComplete={() => playSound('cardFlip')} />
                      )}
                    </div>
                  </div>
                ))}
            </div>

            {/* CPU actions: toast on mobile, inline log on desktop */}
            {isMobile ? <CpuActionToast actions={state.cpuActions} /> : <CpuActionLog actions={state.cpuActions} />}

            {/* Round results */}
            {isShowdown && <RoundResults results={roundResultsForDisplay} players={state.players ?? []} />}

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
                    <AnimatedCard
                      card={humanPlayer.card}
                      width={cardWidth}
                      style={{ border: '3px solid transparent' }}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                  ) : !humanPlayer.folded ? (
                    <AnimatedCardBack width={cardWidth} onFlipComplete={() => playSound('cardFlip')} />
                  ) : null}
                </div>
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
          </GameFooter>
        </>
      )}
      <WinCelebration show={phase === IndianPokerPhase.END} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
