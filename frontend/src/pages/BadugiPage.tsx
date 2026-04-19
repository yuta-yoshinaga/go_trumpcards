import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
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
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { PokerSkeleton } from '../components/skeleton/PokerSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useBadugiGame } from '../hooks/useBadugiGame';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnSuccess, btnWarning, focusRingAccent } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { BadugiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

/** Badugi tutorial step definitions. */
const BG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bg-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bg-exchange-button"]',
    messageKey: 'tutorial.exchangeButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bg-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bg-result-message"]',
    messageKey: 'tutorial.resultMessage',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bg-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Badugi (4-card draw lowball) game page. */
export function BadugiPage() {
  return (
    <TutorialWrapper gameName="badugi" steps={BG_TUTORIAL_STEPS}>
      <BadugiPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Badugi page, wrapped by TutorialProvider. */
function BadugiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('badugi');
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const isMobile = useIsMobile();
  const gameHook = useBadugiGame();
  const { state, loading, error, retry, selected, toggleCard, clearSelection, canExchange } = gameHook;
  const execAction = gameHook.exec;
  const { hint, hintEnabled, setHintEnabled } = useGameHint('badugi', state);

  const [betAmount, setBetAmount] = useState(10);
  const [bettingLimit, setBettingLimit] = useState(0);
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

  const phase = state?.phase ?? BadugiPhase.INIT;
  const isBettingPhase = phase === BadugiPhase.DEAL || phase === BadugiPhase.BET;
  const isEnd = phase === BadugiPhase.END;
  const drawIndex = state?.drawIndex ?? 0;
  const phaseLabel =
    phase === BadugiPhase.DRAW
      ? t('phase.draw', { n: drawIndex })
      : phase === BadugiPhase.BET
        ? t('phase.bet') + (drawIndex > 0 ? ` (${t('drawBadge', { n: drawIndex })})` : '')
        : phase === BadugiPhase.DEAL
          ? t('phase.deal')
          : phase === BadugiPhase.SHOWDOWN
            ? t('phase.showdown')
            : phase === BadugiPhase.END
              ? t('phase.end')
              : t('phase.init');
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
      if (canExchange && !loading) execAction('exchange', selected);
    }, [canExchange, loading, execAction, selected]),
    onClear: clearSelection,
    enabled: canExchange,
  });

  if (!state) return <PokerSkeleton />;

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.poker.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.badugi')} />
      {/* Phase indicator + info bar */}
      <PhaseIndicator phaseName={phaseLabel} isHumanTurn={canAct || canExchange}>
        <span>
          {tc('label.pot')} <strong>{state?.pot ?? 0}</strong>
        </span>
        <TutorialButton />
        <ManualButton gamePath="/badugi" />
        <span>
          {tc('label.dealer')} <strong>Player {state?.dealerIdx ?? 0}</strong>
        </span>
        <span className="text-xs bg-black/20 text-white px-2 py-0.5 rounded">
          {drawIndex === 0 ? t('preDrawLabel') : t('drawBadge', { n: drawIndex })}
        </span>
      </PhaseIndicator>

      {
        <>
          {/* Scrollable: CPU players + logs */}
          <div className={`flex-1 overflow-y-auto pt-4 px-5 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* CPU players */}
            {(() => {
              const cpuCards = cpuPlayers.map((p) => (
                <CpuPlayerCard
                  key={p.id}
                  player={{
                    id: p.id,
                    playStyleName: p.playStyleName,
                    chips: p.chips,
                    currentBet: p.currentBet,
                    folded: p.folded,
                    allIn: p.allIn,
                    handName: p.handName,
                    cards: p.cards,
                  }}
                  showCards={isEnd}
                  faceDownCount={4}
                  showHandName={isEnd}
                  compactFaceDown={!isEnd}
                  extraInfo={
                    p.drawCount > 0 && !p.folded ? (
                      <span className="ml-2 text-xs">{t('exchangeCount', { count: p.drawCount })}</span>
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
                    {t('cpuExchangeEntry', { idx: ex.playerIdx, draw: ex.drawIndex, count: ex.exchangeCount })}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Sticky footer: player hand + buttons */}
          <GameFooter className={`${gameTheme.poker.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="bg-player-hand">
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
                  {humanPlayer.folded && <span className="ml-2 text-ds-error text-xs">[{tc('status.folded')}]</span>}
                  {humanPlayer.allIn && <span className="ml-2 text-ds-warning text-xs">[{tc('status.allIn')}]</span>}
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
                        className={`${focusRingAccent} rounded`}
                        style={{
                          background: 'none',
                          padding: 0,
                          cursor: canExchange ? 'pointer' : 'default',
                          borderRadius: 8,
                          ...selectedCardStyle(isSelected),
                          boxSizing: 'border-box',
                        }}
                      >
                        <AnimatedCard
                          card={card}
                          width={cardWidth}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Message */}
            <div data-tutorial="bg-result-message">
              <GameMessageBox
                message={state?.message}
                messageCode={state?.messageCode}
                messageParams={state?.messageParams}
                alwaysVisible
                severity={isEnd ? 'alert' : 'info'}
              />
            </div>

            {/* Action log */}
            <ActionLogSection
              isEndPhase={!!state?.gameEndFlag}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />

            <ErrorAlert message={error} onRetry={retry} />

            {/* Hint display */}
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {/* Betting controls */}
            {canAct && (
              <div data-tutorial="bg-bet-controls">
                <BettingControls
                  inputId="badugiBetAmount"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={minRaise}
                  maxBetAmount={state?.maxBetAmount}
                  potSize={state?.pot}
                  hasOutstandingBet={hasOutstandingBet}
                  loading={loading}
                  onCall={() => execAction('call', undefined, undefined, undefined, getElapsed())}
                  onRaise={() => execAction('raise', undefined, betAmount, undefined, getElapsed())}
                  onBet={() => execAction('bet', undefined, betAmount, undefined, getElapsed())}
                  onCheck={() => execAction('check', undefined, undefined, undefined, getElapsed())}
                  onFold={() => execAction('fold', undefined, undefined, undefined, getElapsed())}
                  onAllIn={() => execAction('allin', undefined, undefined, undefined, getElapsed())}
                />
              </div>
            )}

            {/* Exchange controls */}
            {canExchange && (
              <div className="text-center mb-2" data-tutorial="bg-exchange-button">
                <button
                  type="button"
                  className={`${btnWarning} min-w-[90px]`}
                  disabled={loading}
                  onClick={() => execAction('exchange', selected)}
                >
                  {t('exchangeLabel')}
                </button>
                <button
                  type="button"
                  className={`${btnSuccess} min-w-[90px]`}
                  disabled={loading}
                  onClick={() => execAction('stand')}
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
                      id: 'badugiBettingLimit',
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
                      id: 'badugiCpuMetaAI',
                      label: t('settings.cpuMetaAI'),
                      checked: cpuMetaAI,
                      onToggle: setCpuMetaAI,
                    },
                    {
                      type: 'checkbox',
                      id: 'badugiHint',
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
                data-tutorial="bg-reset-button"
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    execAction('reset', undefined, undefined, { bettingLimit, cpuMetaAI });
                  })
                }
              >
                {tc('button.reset')}
              </button>
            </div>
          </GameFooter>
        </>
      }
      <WinCelebration show={isEnd} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
