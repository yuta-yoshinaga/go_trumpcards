import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CpuAccordion } from '../components/CpuAccordion';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuActionToast } from '../components/CpuActionToast';
import { CpuPlayerCard } from '../components/CpuPlayerCard';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { RoundResults } from '../components/RoundResults';
import { PokerSkeleton } from '../components/skeleton/PokerSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { usePokerGame } from '../hooks/usePokerGame';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnOutline, btnSuccess, btnWarning, focusRingBlue } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
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
  const isMobile = useIsMobile();
  const { state, loading, error, exec, selected, toggleCard, clearSelection, odds, canExchange } = usePokerGame();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('poker', state);
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
  const cpuPlayers = useMemo(() => state?.players?.filter((p) => !p.isHuman) ?? [], [state?.players]);

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
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.poker.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.poker')} />
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
      <div className={`flex-1 overflow-y-auto pt-4 px-5 lg:px-8 ${lgCardAreaConstraint}`}>
        {/* CPU players */}
        {(() => {
          const cpuCards = cpuPlayers.map((p) => (
            <CpuPlayerCard
              key={p.id}
              player={p}
              showCards={isEnd}
              faceDownCount={5}
              showHandName={isEnd}
              compactFaceDown={!isEnd}
              extraInfo={
                (phase === PokerPhase.SECOND_BET || isEnd) && p.exchangeCount > 0 && !p.folded ? (
                  <span className="ml-2 text-xs">{t('exchangeCount', { count: p.exchangeCount })}</span>
                ) : undefined
              }
            />
          ));
          return isMobile ? <CpuAccordion playerCount={cpuPlayers.length}>{cpuCards}</CpuAccordion> : cpuCards;
        })()}

        {/* CPU actions: toast on mobile, inline log on desktop */}
        {isMobile ? <CpuActionToast actions={state?.cpuActions} /> : <CpuActionLog actions={state?.cpuActions} />}

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
      <GameFooter className={`${gameTheme.poker.footer} px-5 py-3`}>
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

        {/* Hint display */}
        {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

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

        {/* Settings (collapsible) + Reset */}
        <SettingsPanel
          title={t('settings.title')}
          groups={[
            {
              items: [
                {
                  type: 'select',
                  id: 'pokerBettingLimit',
                  label: tc('betting.bettingLimit'),
                  value: bettingLimit,
                  options: [
                    { value: 0, label: tc('betting.fixed') },
                    { value: 1, label: tc('betting.potLimit') },
                    { value: 2, label: tc('betting.noLimit') },
                  ],
                  onSelect: (v) => setBettingLimit(Number(v)),
                },
                {
                  type: 'checkbox',
                  id: 'pokerLowball',
                  label: t('lowball'),
                  checked: isLowball,
                  onToggle: setIsLowball,
                },
                {
                  type: 'checkbox',
                  id: 'pokerCpuMetaAI',
                  label: t('settings.cpuMetaAI'),
                  checked: cpuMetaAI,
                  onToggle: setCpuMetaAI,
                },
                {
                  type: 'checkbox',
                  id: 'pokerHint',
                  label: tc('hint.toggle', { ns: 'tutorial' }),
                  checked: hintEnabled,
                  onToggle: setHintEnabled,
                },
              ],
            },
          ]}
        />
        <div className="text-center mt-2">
          <button
            type="button"
            className={`${btnOutline} min-w-[90px]`}
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
