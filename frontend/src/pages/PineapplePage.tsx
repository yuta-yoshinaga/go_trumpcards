import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { crazyPineappleApi, irishPokerApi, pineappleApi } from '../api/gameApi';
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
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
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
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { placeholderCardStyle, selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PineappleResponse } from '../types/card';
import { HoldemRebuyPhaseType, PineapplePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { PINEAPPLE_HELP, parsePineappleCommand } from '../utils/cli/commands/pineappleCommands';
import { formatPineappleState } from '../utils/cli/formatters/pineappleFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName } from '../utils/playerUtils';

/** Pineapple Poker tutorial step definitions. */
const PN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pn-community-cards"]',
    messageKey: 'tutorial.communityCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-discard-controls"]',
    messageKey: 'tutorial.discardPhase',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-learning-mode"]',
    messageKey: 'tutorial.learningMode',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PINEAPPLE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PineapplePhase.PRE_FLOP]: 'preFlop',
  [PineapplePhase.FLOP]: 'flop',
  [PineapplePhase.TURN]: 'turn',
  [PineapplePhase.RIVER]: 'river',
  [PineapplePhase.SHOWDOWN]: 'showdown',
  [PineapplePhase.END]: 'end',
  [PineapplePhase.REBUY]: 'rebuy',
  [PineapplePhase.DISCARD]: 'discard',
};

/** Variant of the Pineapple page: standard "pineapple" or the Crazy Pineapple
 * variant (discard happens after the flop betting round instead of before). */
export type PineappleVariant = 'pineapple' | 'crazypineapple' | 'irishpoker';

/** Renders the Pineapple Poker game page with community cards, discard phase, betting, and showdown. */
export function PineapplePage({ variant = 'pineapple' }: { variant?: PineappleVariant } = {}) {
  return (
    <TutorialWrapper gameName={variant} steps={PN_TUTORIAL_STEPS}>
      <PineapplePageContent variant={variant} />
    </TutorialWrapper>
  );
}

/** Inner content of the Pineapple Poker page, wrapped by TutorialProvider. */
function PineapplePageContent({ variant }: { variant: PineappleVariant }) {
  const apiClient =
    variant === 'irishpoker' ? irishPokerApi : variant === 'crazypineapple' ? crazyPineappleApi : pineappleApi;
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup(variant);
  const phaseNames = usePhaseNames(variant, PINEAPPLE_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const isMobile = useIsMobile();
  const { state, loading, error, exec: apiExec, retry } = useGameApi(apiClient.exec);
  const [betAmount, setBetAmount] = useState(20);
  const [learningMode, setLearningMode] = useState(false);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const { hint, hintEnabled, setHintEnabled } = useGameHint(variant, state);
  const [selectedDiscard, setSelectedDiscard] = useState<number | null>(null);
  const turnStartRef = useRef(0);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode(variant);
  const cliConfig: CliGameConfig<PineappleResponse, Parameters<typeof apiClient.exec>> = useMemo(
    () => ({
      gameName: variant,
      parseCommand: parsePineappleCommand,
      formatResponse: formatPineappleState,
      helpText: PINEAPPLE_HELP,
    }),
    [variant],
  );
  const { handleCommand } = useCliGame(apiExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(apiExec);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void apiExec('reset', undefined, { cpuMetaAI });
  }, [apiExec, hideActionLog, cpuMetaAI]);

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

  // Reset discard selection when leaving discard phase
  useEffect(() => {
    if (!state?.isDiscardPhase) {
      setSelectedDiscard(null);
    }
  }, [state?.isDiscardPhase]);

  const getElapsed = useCallback(() => {
    if (!cpuMetaAI || turnStartRef.current === 0) return 0;
    const elapsed = Date.now() - turnStartRef.current;
    turnStartRef.current = 0;
    return elapsed;
  }, [cpuMetaAI]);

  const phase = state?.phase ?? PineapplePhase.INIT;
  const isActive = phase >= PineapplePhase.PRE_FLOP && phase <= PineapplePhase.RIVER;
  const isShowdown = phase === PineapplePhase.SHOWDOWN || phase === PineapplePhase.END;
  const isDiscardPhase = state?.isDiscardPhase ?? false;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isActive && !humanFolded && !humanAllIn && !isDiscardPhase && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;
  const isMuckPhase = phase === PineapplePhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase = phase === PineapplePhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.REBUY;
  const isAddonPhase = phase === PineapplePhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.ADDON;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const humanRebuyCount = state?.rebuyCounts?.[humanIdx] ?? 0;
  const cpuPlayers = useMemo(() => state?.players?.filter((p) => !p.isHuman) ?? [], [state?.players]);
  const humanDiscardDone = state?.discardDone?.[humanIdx] ?? false;
  const canDiscard = isDiscardPhase && !humanDiscardDone;

  const actionBindings = useMemo(
    () => [
      { key: 'c', action: () => apiExec('call', undefined, undefined, getElapsed()), enabled: hasOutstandingBet },
      {
        key: 'r',
        action: () =>
          hasOutstandingBet
            ? apiExec('raise', betAmount, undefined, getElapsed())
            : apiExec('bet', betAmount, undefined, getElapsed()),
      },
      { key: 'k', action: () => apiExec('check', undefined, undefined, getElapsed()), enabled: !hasOutstandingBet },
      { key: 'f', action: () => apiExec('fold', undefined, undefined, getElapsed()) },
      { key: 'a', action: () => apiExec('allin', undefined, undefined, getElapsed()) },
    ],
    [apiExec, hasOutstandingBet, betAmount, getElapsed],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: canAct && !loading,
  });

  if (!state)
    return (
      <GameSkeleton
        gameKey="pineapple"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 2, footerHandSize: 2 }}
      />
    );

  return (
    <GamePageShell
      title={tc(`nav.${variant}`)}
      gameThemeBg={gameTheme[variant].bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct || canDiscard}
      gamePath={`/${variant}`}
      gameEndFlag={!!state?.gameEndFlag}
      winShow={phase === PineapplePhase.END}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="pn-pot-display">
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
                  <div className="text-ds-text-primary text-lg mb-1.5">{t('communityCards')}</div>
                  <div className="flex flex-wrap gap-2">
                    {state?.communityCards?.length
                      ? state.communityCards.map((card) => (
                          <AnimatedCard
                            key={`${card.design}-${card.value}`}
                            card={card}
                            width={cardWidth}
                            style={placeholderCardStyle}
                          />
                        ))
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
                      <HudStats namespace={variant} vpip={p.vpip} pfr={p.pfr} threeBet={p.threeBet} af={p.af} />
                    ) : undefined
                  }
                />
              ));

              if (!isMobile) {
                return (
                  <PokerTableLayout
                    communityCardsTutorial="pn-community-cards"
                    cpuAreaTutorial="pn-cpu-area"
                    communityCards={communityCardsContent}
                    cpuPlayers={cpuPlayerCards}
                  />
                );
              }

              return (
                <>
                  <div
                    className="sticky top-0 z-10 bg-game-bg-green-poker pb-1 shadow-sm"
                    data-tutorial="pn-community-cards"
                  >
                    {communityCardsContent}
                  </div>
                  <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="pn-cpu-area">
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
          <GameFooter className={`${gameTheme[variant].footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="pn-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
                  {t('yourHand')}
                  <span className="ml-3 text-xs">
                    {tc('betting.chips')} {humanPlayer.chips}
                  </span>
                  {humanPlayer.totalHands > 0 && (
                    <HudStats
                      namespace={variant}
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
                    <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                      {humanPlayer.handName}
                    </span>
                  )}
                </div>
                <div className="flex flex-wrap gap-1.5 mb-2">
                  {humanPlayer.cards?.length
                    ? humanPlayer.cards.map((card, idx) => (
                        <button
                          key={`${card.design}-${card.value}`}
                          type="button"
                          onClick={() => canDiscard && setSelectedDiscard(idx)}
                          aria-pressed={canDiscard ? selectedDiscard === idx : undefined}
                          className={canDiscard ? 'cursor-pointer' : 'cursor-default'}
                          disabled={!canDiscard}
                          style={selectedCardStyle(canDiscard && selectedDiscard === idx)}
                        >
                          <AnimatedCard card={card} width={cardWidth} />
                        </button>
                      ))
                    : !humanPlayer.folded &&
                      Array.from({ length: state?.initialDealCount ?? 3 }).map((_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth} />
                      ))}
                </div>
              </div>
            )}

            {/* Message */}
            <GameMessageBox
              message={state?.message}
              messageCode={state?.messageCode}
              messageParams={state?.messageParams}
              alwaysVisible
            />

            <ErrorAlert message={error} onRetry={retry} />

            {/* Discard controls */}
            {canDiscard && (
              <div className="mb-2 text-center" data-testid="discard-controls" data-tutorial="pn-discard-controls">
                <p className="text-ds-text-primary mb-2">{t('discard.select')}</p>
                <button
                  type="button"
                  className={`${btnPrimary} min-w-[90px]`}
                  disabled={loading || selectedDiscard === null}
                  onClick={() => {
                    if (selectedDiscard !== null) {
                      apiExec('discard', undefined, { cardIdx: selectedDiscard });
                      setSelectedDiscard(null);
                    }
                  }}
                >
                  {t('discard.prompt')}
                </button>
              </div>
            )}

            {/* Muck/Show controls */}
            {isMuckPhase && (
              <div className="mb-2 text-center" data-testid="muck-controls">
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={`${btnPrimary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('muck')}
                  >
                    {t('muck.muck')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('show')}
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
                    onClick={() => apiExec('rebuy')}
                  >
                    {t('rebuy.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('skiprebuy')}
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
                    onClick={() => apiExec('addon')}
                  >
                    {t('addon.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => apiExec('skipaddon')}
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
              <div data-tutorial="pn-action-buttons">
                <BettingControls
                  inputId="pineappleBetAmount"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={minRaise}
                  maxBetAmount={state?.maxBetAmount}
                  potSize={state?.pot}
                  hasOutstandingBet={hasOutstandingBet}
                  loading={loading}
                  onCall={() => apiExec('call', undefined, undefined, getElapsed())}
                  onRaise={() => apiExec('raise', betAmount, undefined, getElapsed())}
                  onBet={() => apiExec('bet', betAmount, undefined, getElapsed())}
                  onCheck={() => apiExec('check', undefined, undefined, getElapsed())}
                  onFold={() => apiExec('fold', undefined, undefined, getElapsed())}
                  onAllIn={() => apiExec('allin', undefined, undefined, getElapsed())}
                />
              </div>
            )}

            {/* Settings + Reset */}
            <details className="mb-1" data-tutorial="pn-learning-mode" open={learningMode || undefined}>
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
              </div>
            </details>
            <GameResetButton
              isGameEnd={phase === PineapplePhase.SHOWDOWN || phase === PineapplePhase.END}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
              dataTutorial="pn-reset-button"
              className="min-w-[90px]"
            />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
