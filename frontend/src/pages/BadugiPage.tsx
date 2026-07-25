import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { badugiApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CpuAccordion } from '../components/CpuAccordion';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuActionToast } from '../components/CpuActionToast';
import { CpuPlayerCard } from '../components/CpuPlayerCard';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useBadugiGame } from '../hooks/useBadugiGame';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnSuccess, btnWarning, focusRingAccent } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BadugiResponse } from '../types/card';
import { BadugiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { badugiBestSubsetIndices } from '../utils/badugiBestSubset';
import { isCompleteBadugiHand } from '../utils/badugiUtils';
import { cardAlt } from '../utils/cardAlt';
import { BADUGI_HELP, parseBadugiCommand } from '../utils/cli/commands/badugiCommands';
import { formatBadugiState } from '../utils/cli/formatters/badugiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName } from '../utils/playerUtils';

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
export const BadugiPage = withTutorial(BadugiPageContent, 'badugi', BG_TUTORIAL_STEPS);
/** Inner content of the Badugi page, wrapped by TutorialProvider. */
function BadugiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('badugi');
  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();
  const gameHook = useBadugiGame();
  const { state, loading, error, retry, selected, toggleCard, clearSelection, canExchange } = gameHook;
  const execAction = gameHook.exec;
  const { hint, hintEnabled, setHintEnabled } = useGameHint('badugi', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('badugi');
  const cliConfig: CliGameConfig<BadugiResponse, Parameters<typeof badugiApi.exec>> = useMemo(
    () => ({
      gameName: 'badugi',
      parseCommand: parseBadugiCommand,
      formatResponse: formatBadugiState,
      helpText: BADUGI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execAction, cliConfig, state, { addInput, addOutput, addError, clearLog });

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

  // Track the previous turn so the deliberation clock only starts when the
  // turn transitions INTO the human. Without this guard, every state update
  // during the human turn would overwrite turnStartRef and any re-render
  // during a pending action would zero out the elapsed time.
  const prevTurnRef = useRef<number | null>(null);
  useEffect(() => {
    const humanId = state?.players?.find((p) => p.isHuman)?.id;
    if (state && state.currentTurn === humanId && prevTurnRef.current !== humanId) {
      turnStartRef.current = Date.now();
    }
    prevTurnRef.current = state?.currentTurn ?? null;
  }, [state]);

  const getElapsed = useCallback(() => {
    if (!cpuMetaAI || turnStartRef.current === 0) return 0;
    return Date.now() - turnStartRef.current;
  }, [cpuMetaAI]);

  const phase = state?.phase ?? BadugiPhase.INIT;
  const isBettingPhase = phase === BadugiPhase.DEAL || phase === BadugiPhase.BET;
  const isEnd = phase === BadugiPhase.END;
  const isHandOver = phase === BadugiPhase.SHOWDOWN || phase === BadugiPhase.END;
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
  // Stand-pat protection: warn the player when their 4 cards are already a complete Badugi
  // (4 distinct ranks + 4 distinct suits). Exchanging from this state can only weaken the hand.
  // Memoised so the Set allocations only run when the hand actually changes.
  const humanHasCompleteBadugi = useMemo(
    () => (canExchange ? isCompleteBadugiHand(humanPlayer?.cards ?? []) : false),
    [canExchange, humanPlayer?.cards],
  );

  // Set of card indices belonging to the current best Badugi subset. Only computed during the draw
  // phase — outside of that window the lift / dim visuals would distract more than help, since the
  // player can't act on "dead weight" until the next exchange opens.
  const subsetIndices = useMemo(
    () => (canExchange ? new Set(badugiBestSubsetIndices(humanPlayer?.cards ?? [])) : null),
    [canExchange, humanPlayer?.cards],
  );

  useCardKeyboardNav({
    cardCount,
    onToggle: toggleCard,
    onConfirm: useCallback(() => {
      if (canExchange && !loading) execAction('exchange', selected);
    }, [canExchange, loading, execAction, selected]),
    onClear: clearSelection,
    enabled: canExchange,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    execAction('reset', undefined, undefined, { bettingLimit, cpuMetaAI });
  }, [execAction, hideActionLog, bettingLimit, cpuMetaAI]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="badugi"
        layout={{ kind: 'community-poker', opponents: 3, opponentCards: 5, footerHandSize: 5 }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.badugi')}
      gameThemeBg={gameTheme.badugi.bg}
      phaseName={phaseLabel}
      isHumanTurn={canAct || canExchange}
      gamePath="/badugi"
      gameEndFlag={isEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>
            {tc('label.pot')} <strong>{state?.pot ?? 0}</strong>
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
      headerEnd={
        <>
          <span>
            {tc('label.dealer')} <strong>{findPlayerName(state.players, state.dealerIdx)}</strong>
          </span>
          <span className="text-xs bg-black/20 text-ds-text-primary px-2 py-0.5 rounded">
            {drawIndex === 0 ? t('preDrawLabel') : t('drawBadge', { n: drawIndex })}
          </span>
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
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
              <div className="bg-black/30 rounded p-2 mb-3 text-ds-text-primary text-xs">
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
          <GameFooter className={`${gameTheme.badugi.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="bg-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
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
                  {(humanPlayer.cards ?? []).map((card, i) => {
                    const isSelected = selected.includes(i);
                    // Only annotate cards during the draw phase; outside of it we don't want to
                    // distract the player with a "dead weight" hint they can't act on.
                    const inSubset = subsetIndices?.has(i) ?? false;
                    const showSubsetHint = subsetIndices !== null;
                    let liftOrDim = '';
                    if (showSubsetHint) {
                      if (inSubset) liftOrDim = '-translate-y-1';
                      else if (!isSelected) liftOrDim = 'opacity-50';
                    }
                    // During the draw phase, mirror the visual lift/dim cue into the
                    // aria-label so screen readers learn which cards form the current
                    // best hand and which are exchange candidates.
                    const subsetHint = showSubsetHint ? ` ${inSubset ? t('subsetBest') : t('subsetExchange')}` : '';
                    return (
                      <button
                        key={`${card.design}-${card.value}`}
                        type="button"
                        aria-label={`${cardAlt(card)}${isSelected ? ` ${t('cardSelected')}` : ''}${subsetHint}`}
                        aria-pressed={isSelected}
                        onClick={() => toggleCard(i)}
                        data-badugi-subset={showSubsetHint && inSubset ? 'true' : undefined}
                        className={`${focusRingAccent} rounded transition-transform ${liftOrDim}`}
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
                {humanHasCompleteBadugi && (
                  <div
                    role="status"
                    aria-live="polite"
                    data-testid="bg-complete-badugi-banner"
                    className="mb-2 inline-block px-3 py-1 rounded bg-ds-accent/15 border border-ds-accent text-ds-accent text-sm font-bold"
                  >
                    {t('completeBadugiBanner')}
                  </div>
                )}
                <button
                  type="button"
                  className={`${btnWarning} min-w-[90px]`}
                  disabled={loading}
                  onClick={() => execAction('exchange', selected)}
                  data-testid="bg-exchange-btn"
                >
                  {t('exchangeLabel')}
                </button>
                <button
                  type="button"
                  className={`${btnSuccess} min-w-[90px]${
                    humanHasCompleteBadugi ? ' ring-2 ring-ds-accent animate-pulse' : ''
                  }`}
                  disabled={loading}
                  onClick={() => execAction('stand')}
                  data-testid="bg-stand-btn"
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
                      tooltip: t('settings.bettingLimitHelp'),
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
                      tooltip: t('settings.cpuMetaAIHelp'),
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
              <GameResetButton
                isGameEnd={isHandOver}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="bg-reset-button"
                className="min-w-[90px]"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
