import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuPlayerCard } from '../components/CpuPlayerCard';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GameResetDialog } from '../components/GameResetDialog';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { RoundResults } from '../components/RoundResults';
import { PokerSkeleton } from '../components/skeleton/PokerSkeleton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { usePokerGame } from '../hooks/usePokerGame';
import { TutorialProvider, useTutorialContext } from '../providers/TutorialProvider';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning, focusRingBlue } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { PokerPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

const POKER_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PokerPhase.INIT]: 'init',
  [PokerPhase.DEAL]: 'deal',
  [PokerPhase.EXCHANGE]: 'exchange',
  [PokerPhase.SECOND_BET]: 'secondBet',
  [PokerPhase.END]: 'end',
};

/** Poker tutorial step definitions. */
const PK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pk-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pk-exchange-button"]',
    messageKey: 'tutorial.exchangeButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pk-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pk-result-message"]',
    messageKey: 'tutorial.resultMessage',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pk-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Poker tutorial configuration. */
const PK_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'poker',
  steps: PK_TUTORIAL_STEPS,
};

/** Tutorial button that starts the Poker tutorial. */
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

/** Renders the 5-card Draw Poker game page with betting and card exchange. */
export function PokerPage() {
  const { t: tPk } = useTranslation('poker');
  return (
    <TutorialProvider config={PK_TUTORIAL_CONFIG} translateMessage={tPk}>
      <PokerPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Poker page, wrapped by TutorialProvider. */
function PokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('poker');
  const phaseNames = usePhaseNames('poker', POKER_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec, selected, toggleCard, clearSelection, odds, canExchange } = usePokerGame();
  const [betAmount, setBetAmount] = useState(10);
  const [bettingLimit, setBettingLimit] = useState(0);
  const [isLowball, setIsLowball] = useState(false);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const turnStartRef = useRef(0);

  useEffect(() => {
    if (state?.minRaise && state.minRaise > 0) {
      setBetAmount(state.minRaise);
    } else if (state) {
      setBetAmount(10);
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

  const phase = state?.phase ?? PokerPhase.INIT;
  const isBettingPhase = phase === PokerPhase.DEAL || phase === PokerPhase.SECOND_BET;
  const isEnd = phase === PokerPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isBettingPhase && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 10;
  const cardCount = humanPlayer?.cards?.length ?? 0;

  useCardKeyboardNav({
    cardCount,
    onToggle: toggleCard,
    onConfirm: useCallback(() => {
      if (canExchange && !loading) exec('exchange', selected);
    }, [canExchange, loading, exec, selected]),
    onClear: clearSelection,
    enabled: canExchange,
  });

  if (!state) return <PokerSkeleton />;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green-poker" aria-busy={loading} aria-live="polite">
      {/* Phase indicator + info bar */}
      <PhaseIndicator phaseName={phaseNames[phase] ?? t('phase.init')} isHumanTurn={canAct || canExchange}>
        <span>
          {tc('label.pot')} <strong>{state?.pot ?? 0}</strong>
        </span>
        <TutorialButton />
        <span>
          {tc('label.dealer')} <strong>Player {state?.dealerIdx ?? 0}</strong>
        </span>
        {(state?.jokerCount ?? 0) > 0 && (
          <span>
            {t('joker')} <strong>{state?.jokerCount}</strong>
          </span>
        )}
        {state?.isLowball && (
          <span className="bg-yellow-600 text-white px-2 py-0.5 rounded text-xs font-bold">[{t('lowballMode')}]</span>
        )}
      </PhaseIndicator>

      {/* Scrollable: CPU players + logs */}
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        {/* CPU players */}
        {state?.players
          ?.filter((p) => !p.isHuman)
          .map((p) => (
            <CpuPlayerCard
              key={p.id}
              player={p}
              showCards={isEnd}
              faceDownCount={5}
              showHandName={isEnd}
              extraInfo={
                (phase === PokerPhase.SECOND_BET || isEnd) && p.exchangeCount > 0 && !p.folded ? (
                  <span className="ml-2 text-xs">{t('exchangeCount', { count: p.exchangeCount })}</span>
                ) : undefined
              }
            />
          ))}

        {/* CPU actions log */}
        <CpuActionLog actions={state?.cpuActions} />

        {/* CPU exchanges log */}
        {state?.cpuExchanges && state.cpuExchanges.length > 0 && (
          <div className="bg-black/30 rounded p-2 mb-3 text-white text-xs">
            <div className="font-bold mb-1">{t('cpuExchange')}</div>
            {state.cpuExchanges.map((ex, i) => (
              <div key={`${i}-${ex.playerIdx}`}>
                {t('cpuExchangeEntry', { idx: ex.playerIdx, count: ex.exchangeCount })}
              </div>
            ))}
          </div>
        )}

        {/* Round results */}
        {isEnd && <RoundResults results={state?.roundResults} players={state?.players ?? []} />}
      </div>

      {/* Sticky footer: player hand + buttons */}
      <GameFooter className="bg-game-bg-green-poker-dark border-white/20 px-5 py-3">
        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2" data-tutorial="pk-player-hand">
            <div className="text-white text-lg mb-1">
              {t('yourHand')}
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
              {isEnd && !humanPlayer.folded && humanPlayer.handName && (
                <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                  {humanPlayer.handName}
                </span>
              )}
            </div>
            {canExchange && <div className="text-game-text-highlight text-xs mb-1">{t('exchangeInstruction')}</div>}
            <div className="flex flex-wrap gap-1.5 mb-2">
              {humanPlayer.cards?.map((card, i) => {
                const isSelected = selected.includes(i);
                return (
                  <button
                    key={`${card.design}-${card.value}`}
                    type="button"
                    aria-label={`${cardAlt(card)}${isSelected ? ` ${t('cardSelected')}` : ''}`}
                    aria-pressed={isSelected}
                    onClick={() => toggleCard(i)}
                    className={`${focusRingBlue} rounded`}
                    style={{
                      background: 'none',
                      padding: 0,
                      cursor: canExchange ? 'pointer' : 'default',
                      borderRadius: 8,
                      ...selectedCardStyle(isSelected),
                      boxSizing: 'border-box',
                    }}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                );
              })}
            </div>
          </div>
        )}

        {/* Message */}
        <div data-tutorial="pk-result-message">
          <GameMessageBox
            message={state?.message}
            messageCode={state?.messageCode}
            messageParams={state?.messageParams}
            alwaysVisible
          />
        </div>

        {/* Action log */}
        <ActionLogSection
          isEndPhase={!!state?.gameEndFlag}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />

        <ErrorAlert message={error} />

        {/* Betting controls */}
        {canAct && (
          <div data-tutorial="pk-bet-controls">
            <BettingControls
              inputId="pokerBetAmount"
              betAmount={betAmount}
              onBetAmountChange={setBetAmount}
              minRaise={minRaise}
              maxBetAmount={state?.maxBetAmount}
              hasOutstandingBet={hasOutstandingBet}
              loading={loading}
              onCall={() => exec('call', undefined, undefined, undefined, getElapsed())}
              onRaise={() => exec('raise', undefined, betAmount, undefined, getElapsed())}
              onBet={() => exec('bet', undefined, betAmount, undefined, getElapsed())}
              onCheck={() => exec('check', undefined, undefined, undefined, getElapsed())}
              onFold={() => exec('fold', undefined, undefined, undefined, getElapsed())}
              onAllIn={() => exec('allin', undefined, undefined, undefined, getElapsed())}
            />
          </div>
        )}

        {/* Draw odds panel */}
        {canExchange && odds?.some((o) => o.probability > 0) && (
          <div className="bg-black/40 rounded-lg px-4 py-2 mb-2 text-white text-xs" data-testid="odds-panel">
            <div className="font-bold mb-1">{t('drawOdds')}</div>
            {odds
              .filter((o) => o.probability > 0)
              .map((o) => (
                <div key={o.handRank} className="flex justify-between">
                  <span>{o.handName}</span>
                  <span>{(o.probability * 100).toFixed(1)}%</span>
                </div>
              ))}
          </div>
        )}

        {/* Exchange controls */}
        {canExchange && (
          <div className="text-center mb-2" data-tutorial="pk-exchange-button">
            <button
              type="button"
              className={`${btnWarning} min-w-[90px]`}
              disabled={loading}
              onClick={() => exec('exchange', selected)}
            >
              {t('exchangeLabel')}
            </button>
            <button
              type="button"
              className={`${btnSuccess} min-w-[90px]`}
              disabled={loading}
              onClick={() => exec('stand')}
            >
              {t('standLabel')}
            </button>
          </div>
        )}

        {/* Settings + Reset */}
        <div className="text-center flex items-center justify-center gap-3">
          <label className="text-white text-sm flex items-center gap-1">
            {tc('betting.bettingLimit')}
            <select
              value={bettingLimit}
              onChange={(e) => setBettingLimit(Number(e.target.value))}
              className="px-2 py-1 text-sm rounded bg-white/90 text-gray-900"
            >
              <option value={0}>{tc('betting.fixed')}</option>
              <option value={1}>{tc('betting.potLimit')}</option>
              <option value={2}>{tc('betting.noLimit')}</option>
            </select>
          </label>
          <label className="text-white text-sm flex items-center gap-1">
            <input type="checkbox" checked={isLowball} onChange={(e) => setIsLowball(e.target.checked)} />
            {t('lowball')}
          </label>
          <label className="text-white text-sm flex items-center gap-1">
            <input type="checkbox" checked={cpuMetaAI} onChange={(e) => setCpuMetaAI(e.target.checked)} />
            {t('settings.cpuMetaAI')}
          </label>
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            data-tutorial="pk-reset-button"
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                exec('reset', undefined, undefined, { bettingLimit, isLowball, cpuMetaAI });
              })
            }
          >
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
      <WinCelebration show={phase === PokerPhase.END} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
