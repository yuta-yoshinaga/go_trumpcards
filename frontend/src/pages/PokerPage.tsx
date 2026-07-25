import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { pokerApi } from '../api/gameApi';
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
import { RoundResults } from '../components/RoundResults';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { usePokerGame } from '../hooks/usePokerGame';
import { badgeError } from '../styles/badgeStyles';
import { btnSuccess, btnWarning, focusRingAccent } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PokerResponse } from '../types/card';
import { PokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { POKER_HELP, parsePokerCommand } from '../utils/cli/commands/pokerCommands';
import { formatPokerState } from '../utils/cli/formatters/pokerFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName } from '../utils/playerUtils';

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

/** Renders the 5-card Draw Poker game page with betting and card exchange. */
export const PokerPage = withTutorial(PokerPageContent, 'poker', PK_TUTORIAL_STEPS);
/** Inner content of the Poker page, wrapped by TutorialProvider. */
function PokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('poker');
  const phaseNames = usePhaseNames('poker', POKER_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    selected,
    toggleCard,
    clearSelection,
    odds,
    oddsError,
    retryOdds,
    canExchange,
  } = usePokerGame();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('poker', state);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('poker');
  const pokerCliConfig: CliGameConfig<PokerResponse, Parameters<typeof pokerApi.exec>> = useMemo(
    () => ({
      gameName: 'poker',
      parseCommand: parsePokerCommand,
      formatResponse: formatPokerState,
      helpText: POKER_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, pokerCliConfig, state, { addInput, addOutput, addError, clearLog });
  const [betAmount, setBetAmount] = useState(10);
  const [bettingLimit, setBettingLimit] = useState(0);
  const [isLowball, setIsLowball] = useState(false);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const turnStartRef = useRef(0);

  // Sync the raise amount to the current minimum only when that minimum actually
  // changes (a raise, or a new round). Keying on `state` would re-run on every CPU
  // action and clobber the amount the player is typing (#2980).
  const minRaiseValue = state?.minRaise;
  useEffect(() => {
    if (minRaiseValue === undefined) return;
    setBetAmount(minRaiseValue > 0 ? minRaiseValue : 10);
  }, [minRaiseValue]);

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
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void exec('reset', undefined, undefined, { bettingLimit, isLowball, cpuMetaAI });
  }, [exec, hideActionLog, bettingLimit, isLowball, cpuMetaAI]);
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
      if (canExchange && !loading && selected.length > 0) exec('exchange', selected);
    }, [canExchange, loading, exec, selected]),
    onClear: clearSelection,
    enabled: canExchange,
  });

  if (!state)
    return (
      <GameSkeleton
        gameKey="poker"
        layout={{ kind: 'community-poker', opponents: 3, opponentCards: 5, footerHandSize: 5 }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.poker')}
      gameThemeBg={gameTheme.poker.bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct || canExchange}
      gamePath="/poker"
      gameEndFlag={phase === PokerPhase.END}
      winShow={
        phase === PokerPhase.END &&
        state.roundResults.some((r) => state.players[r.playerIdx]?.isHuman && r.wonAmount > 0)
      }
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
          {(state?.jokerCount ?? 0) > 0 && (
            <span>
              {t('joker')} <strong>{state?.jokerCount}</strong>
            </span>
          )}
          {state?.isLowball && (
            <span className="bg-ds-warning text-white px-2 py-0.5 rounded text-xs font-bold">[{t('lowballMode')}]</span>
          )}
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Scrollable: CPU players + logs */}
          <div className={`flex-1 overflow-y-auto pt-4 px-5 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Lowball hand-rank quick reference (only meaningful in 2-7 lowball). */}
            {state?.isLowball && (
              <details className="mb-2" data-testid="pk-lowball-reference">
                <summary className="cursor-pointer select-none text-ds-text-primary text-sm font-bold py-1">
                  {t('lowballRank.title')}
                </summary>
                <ul className="list-disc list-inside text-ds-text-muted text-xs py-1 space-y-0.5">
                  <li className="text-ds-text-primary font-semibold">{t('lowballRank.best')}</li>
                  <li>{t('lowballRank.aceHigh')}</li>
                  <li>{t('lowballRank.straightFlushCount')}</li>
                  <li>{t('lowballRank.goal')}</li>
                </ul>
              </details>
            )}

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
              <div className="bg-black/30 rounded p-2 mb-3 text-ds-text-primary text-xs">
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
                severity={phase === PokerPhase.END ? 'alert' : 'info'}
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
              <div data-tutorial="pk-bet-controls">
                <BettingControls
                  inputId="pokerBetAmount"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={minRaise}
                  maxBetAmount={state?.maxBetAmount}
                  potSize={state?.pot}
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
              <div
                className="bg-black/40 rounded-lg px-4 py-2 mb-2 text-ds-text-primary text-xs"
                data-testid="odds-panel"
              >
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
            {canExchange && oddsError && (
              <div
                role="alert"
                className={`${badgeError} mb-2 flex items-center justify-between gap-2`}
                data-testid="odds-error"
              >
                <span>{t('oddsFetchFailed')}</span>
                <button
                  type="button"
                  onClick={retryOdds}
                  className="underline focus:outline-none focus-visible:ring-2 focus-visible:ring-ds-accent"
                >
                  {t('oddsRetry')}
                </button>
              </div>
            )}

            {/* Exchange controls */}
            {canExchange && (
              <div className="text-center mb-2" data-tutorial="pk-exchange-button">
                <div className="text-game-text-highlight text-xs mb-1" data-testid="pk-exchange-selected">
                  {t('exchangeSelectedCount', { n: selected.length })}
                </div>
                <button
                  type="button"
                  className={`${btnWarning} min-w-[90px]`}
                  disabled={loading || selected.length === 0}
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
                <div className="text-ds-text-muted text-xs mt-1" data-testid="pk-stand-hint">
                  {t('standHint')}
                </div>
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
              <GameResetButton
                isGameEnd={isEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="pk-reset-button"
                className="min-w-[90px]"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
