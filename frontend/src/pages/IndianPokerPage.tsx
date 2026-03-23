import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { indianpokerApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CpuActionLog } from '../components/CpuActionLog';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GameResetDialog } from '../components/GameResetDialog';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { RoundResults } from '../components/RoundResults';
import { HoldemSkeleton } from '../components/skeleton/HoldemSkeleton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { TutorialProvider, useTutorialContext } from '../providers/TutorialProvider';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { IndianPokerPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';

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

/** Indian Poker tutorial configuration. */
const IP_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'indianpoker',
  steps: IP_TUTORIAL_STEPS,
};

/** Tutorial button that starts the Indian Poker tutorial. */
function TutorialButton() {
  const { t } = useTranslation('tutorial');
  const { start } = useTutorialContext();
  return (
    <button
      type="button"
      className={`${btnSecondary} text-xs`}
      onClick={start}
      aria-label={t('tutorialButton')}
      title={t('tutorialButton')}
    >
      ?
    </button>
  );
}

const INDIAN_POKER_PHASE_KEYS: Readonly<Record<number, string>> = {
  [IndianPokerPhase.INIT]: 'init',
  [IndianPokerPhase.ANTE]: 'ante',
  [IndianPokerPhase.BETTING]: 'betting',
  [IndianPokerPhase.SHOWDOWN]: 'showdown',
  [IndianPokerPhase.END]: 'end',
};

/** Renders the Indian Poker game page with opponent cards visible and human card hidden. */
export function IndianPokerPage() {
  const { t: tIp } = useTranslation('indianpoker');
  return (
    <TutorialProvider config={IP_TUTORIAL_CONFIG} translateMessage={tIp}>
      <IndianPokerPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Indian Poker page, wrapped by TutorialProvider. */
function IndianPokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('indianpoker');
  const phaseNames = usePhaseNames('indianpoker', INDIAN_POKER_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi } = useGameApi(indianpokerApi.exec);
  const [betAmount, setBetAmount] = useState(20);
  const [ante] = useState(10);
  const [bettingLimit, setBettingLimit] = useState(2);
  const [cpuMetaAI, setCpuMetaAI] = useState(true);
  const turnStartRef = useRef(0);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

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
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green-poker" aria-busy={loading} aria-live="polite">
      {/* Phase indicator + info bar */}
      <PhaseIndicator phaseName={phaseNames[phase] ?? t('phase.init')} isHumanTurn={canAct}>
        <span>
          {tc('label.pot')} <strong>{state.pot ?? 0}</strong>
        </span>
        <span>
          {t('ante')} <strong>{state.ante ?? 0}</strong>
        </span>
        <TutorialButton />
        <span>
          {tc('label.dealer')} <strong>Player {state.dealerIdx ?? 0}</strong>
        </span>
      </PhaseIndicator>

      {/* Scrollable: opponent cards + CPU players */}
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        {/* CPU players - show cards face-up (opponents can see each other's cards) */}
        <div data-tutorial="ip-cpu-cards">
          {state.players
            ?.filter((p) => !p.isHuman)
            .map((p) => (
              <div key={p.id} className="mb-3">
                <div className="text-white text-sm mb-1">
                  {tc('player.cpu', { id: p.id })} <span className="text-gray-300 text-xs">({p.playStyleName})</span>
                  <span className="ml-2 text-xs">
                    {tc('betting.chips')} {p.chips}
                  </span>
                  {p.currentBet > 0 && (
                    <span className="ml-2 text-xs">
                      {tc('betting.currentBet')} {p.currentBet}
                    </span>
                  )}
                  {p.folded && <span className="ml-2 text-red-300 text-xs">[{tc('status.folded')}]</span>}
                  {p.allIn && <span className="ml-2 text-yellow-300 text-xs">[{tc('status.allIn')}]</span>}
                </div>
                <div className="flex flex-wrap gap-1">
                  {p.card ? (
                    <AnimatedCard card={p.card} width={cardWidth} style={{ border: '3px solid transparent' }} />
                  ) : (
                    <AnimatedCardBack width={cardWidth} />
                  )}
                </div>
              </div>
            ))}
        </div>

        {/* CPU actions log */}
        <CpuActionLog actions={state.cpuActions} />

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
      <GameFooter className="bg-game-bg-green-poker-dark border-white/20 px-5 py-3">
        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2" data-tutorial="ip-player-card">
            <div className="text-white text-lg mb-1">
              {t('yourCard')}
              <span className="ml-3 text-xs">
                {tc('betting.chips')} {humanPlayer.chips}
              </span>
              {humanPlayer.currentBet > 0 && (
                <span className="ml-2 text-xs">
                  {tc('betting.currentBet')} {humanPlayer.currentBet}
                </span>
              )}
              {humanPlayer.folded && <span className="ml-2 text-red-300 text-xs">[{tc('status.folded')}]</span>}
              {humanPlayer.allIn && <span className="ml-2 text-yellow-300 text-xs">[{tc('status.allIn')}]</span>}
            </div>
            <div className="flex flex-wrap gap-1.5 mb-2">
              {isShowdown && humanPlayer.card ? (
                <AnimatedCard card={humanPlayer.card} width={cardWidth} style={{ border: '3px solid transparent' }} />
              ) : !humanPlayer.folded ? (
                <AnimatedCardBack width={cardWidth} />
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

        <ErrorAlert message={error} />

        {/* Betting controls */}
        {canAct && (
          <div data-tutorial="ip-action-buttons">
            <BettingControls
              inputId="indianPokerBetAmount"
              betAmount={betAmount}
              onBetAmountChange={setBetAmount}
              minRaise={minRaise}
              maxBetAmount={state.maxBetAmount}
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
        <div className="text-center flex items-center justify-center gap-3" data-tutorial="ip-reset-button">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                execApi('reset', undefined, { ante, bettingLimit, cpuMetaAI });
              })
            }
          >
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
      <WinCelebration show={phase === IndianPokerPhase.END} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
