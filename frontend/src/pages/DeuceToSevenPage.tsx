import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { deuceToSevenApi } from '../api/gameApi';
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
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useDeuceToSevenGame } from '../hooks/useDeuceToSevenGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnSuccess, btnWarning, focusRingAccent } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DeuceToSevenResponse } from '../types/card';
import { DeuceToSevenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { DEUCE_TO_SEVEN_HELP, parseDeuceToSevenCommand } from '../utils/cli/commands/deuceToSevenCommands';
import { formatDeuceToSevenState } from '../utils/cli/formatters/deuceToSevenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { DEUCE_TO_SEVEN_MAX_DRAWS, deuceToSevenBestIndices, isMadePatLow } from '../utils/deuceToSevenUtils';
import { findPlayerName } from '../utils/playerUtils';
import { type PokerHandRank, pokerHandKey } from '../utils/pokerSquaresUtils';

/** 2-7 Triple Draw tutorial step definitions. */
const D7_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="d7-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="d7-exchange-button"]',
    messageKey: 'tutorial.exchangeButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="d7-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="d7-result-message"]',
    messageKey: 'tutorial.resultMessage',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="d7-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the 2-7 Triple Draw (Deuce to Seven lowball) game page. */
export const DeuceToSevenPage = withTutorial(DeuceToSevenPageContent, 'deucetoseven', D7_TUTORIAL_STEPS);
/** Inner content of the 2-7 Triple Draw page, wrapped by TutorialProvider. */
function DeuceToSevenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('deucetoseven');
  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();
  // Translate the poker category via its stable rank so the badge follows the
  // UI locale; fall back to the server's string if the rank is out of range.
  const handLabel = (handRank: number, handName: string): string =>
    t(`hand.${pokerHandKey(handRank as PokerHandRank)}`, { defaultValue: handName });
  const gameHook = useDeuceToSevenGame();
  const { state, loading, error, retry, selected, toggleCard, clearSelection, canExchange } = gameHook;
  const execAction = gameHook.exec;
  const { hint, hintEnabled, setHintEnabled } = useGameHint('deucetoseven', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('deucetoseven');
  const cliConfig: CliGameConfig<DeuceToSevenResponse, Parameters<typeof deuceToSevenApi.exec>> = useMemo(
    () => ({
      gameName: 'deucetoseven',
      parseCommand: parseDeuceToSevenCommand,
      formatResponse: formatDeuceToSevenState,
      helpText: DEUCE_TO_SEVEN_HELP,
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
  // turn transitions INTO the human (see BadugiPage for the rationale).
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

  const phase = state?.phase ?? DeuceToSevenPhase.INIT;
  const isBettingPhase = phase === DeuceToSevenPhase.DEAL || phase === DeuceToSevenPhase.BET;
  const isEnd = phase === DeuceToSevenPhase.END;
  const isHandOver = phase === DeuceToSevenPhase.SHOWDOWN || phase === DeuceToSevenPhase.END;
  const drawIndex = state?.drawIndex ?? 0;
  const phaseLabel =
    phase === DeuceToSevenPhase.DRAW
      ? t('phase.draw', { n: drawIndex, max: DEUCE_TO_SEVEN_MAX_DRAWS })
      : phase === DeuceToSevenPhase.BET
        ? t('phase.bet') +
          (drawIndex > 0 ? ` (${t('drawBadge', { n: drawIndex, max: DEUCE_TO_SEVEN_MAX_DRAWS })})` : '')
        : phase === DeuceToSevenPhase.DEAL
          ? t('phase.deal')
          : phase === DeuceToSevenPhase.SHOWDOWN
            ? t('phase.showdown')
            : phase === DeuceToSevenPhase.END
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
  // Stand-pat protection: warn when the 5 cards are already a made 8-or-better
  // low (no pair / straight / flush). Drawing from this state can only weaken it.
  const humanHasMadeLow = useMemo(
    () => (canExchange ? isMadePatLow(humanPlayer?.cards ?? []) : false),
    [canExchange, humanPlayer?.cards],
  );

  // During the draw, lift the best-low core cards and dim the draw candidates.
  const bestSubset = useMemo(
    () => (canExchange ? new Set(deuceToSevenBestIndices(humanPlayer?.cards ?? [])) : null),
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
        gameKey="deucetoseven"
        layout={{ kind: 'community-poker', opponents: 3, opponentCards: 5, footerHandSize: 5 }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.deucetoseven')}
      gameThemeBg={gameTheme.deucetoseven.bg}
      phaseName={phaseLabel}
      isHumanTurn={canAct || canExchange}
      gamePath="/deucetoseven"
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
            {drawIndex === 0 ? t('preDrawLabel') : t('drawBadge', { n: drawIndex, max: DEUCE_TO_SEVEN_MAX_DRAWS })}
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
                    handName: handLabel(p.handRank, p.handName),
                    cards: p.cards,
                  }}
                  showCards={isEnd}
                  faceDownCount={5}
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
          <GameFooter className={`${gameTheme.deucetoseven.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="d7-player-hand">
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
                      {handLabel(humanPlayer.handRank, humanPlayer.handName)}
                    </span>
                  )}
                </div>
                {canExchange && <div className="text-game-text-highlight text-xs mb-1">{t('exchangeInstruction')}</div>}
                <div className="flex flex-wrap gap-1.5 mb-2">
                  {(humanPlayer.cards ?? []).map((card, i) => {
                    const isSelected = selected.includes(i);
                    let liftOrDim = '';
                    if (bestSubset) {
                      if (bestSubset.has(i)) liftOrDim = '-translate-y-1';
                      else if (!isSelected) liftOrDim = 'opacity-50';
                    }
                    return (
                      <button
                        key={`${card.design}-${card.value}`}
                        type="button"
                        aria-label={`${cardAlt(card)}${isSelected ? ` ${t('cardSelected')}` : ''}`}
                        aria-pressed={isSelected}
                        onClick={() => toggleCard(i)}
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
            <div data-tutorial="d7-result-message">
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
              <div data-tutorial="d7-bet-controls">
                <BettingControls
                  inputId="deuceToSevenBetAmount"
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
              <div className="text-center mb-2" data-tutorial="d7-exchange-button">
                {humanHasMadeLow && (
                  <div
                    role="status"
                    aria-live="polite"
                    data-testid="d7-made-low-banner"
                    className="mb-2 inline-block px-3 py-1 rounded bg-ds-accent/15 border border-ds-accent text-ds-accent text-sm font-bold"
                  >
                    {t('madePatLowBanner')}
                  </div>
                )}
                <button
                  type="button"
                  className={`${btnWarning} min-w-[90px]`}
                  disabled={loading}
                  onClick={() => execAction('exchange', selected)}
                  data-testid="d7-exchange-btn"
                >
                  {t('exchangeLabel')}
                </button>
                <button
                  type="button"
                  className={
                    humanHasMadeLow
                      ? `${btnSuccess} min-w-[90px] ring-2 ring-ds-accent animate-pulse`
                      : `${btnSuccess} min-w-[90px]`
                  }
                  disabled={loading}
                  onClick={() => execAction('stand')}
                  data-testid="d7-stand-btn"
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
                      id: 'deuceToSevenBettingLimit',
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
                      id: 'deuceToSevenCpuMetaAI',
                      label: t('settings.cpuMetaAI'),
                      checked: cpuMetaAI,
                      onToggle: setCpuMetaAI,
                    },
                    {
                      type: 'checkbox',
                      id: 'deuceToSevenHint',
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
                dataTutorial="d7-reset-button"
                className="min-w-[90px]"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
