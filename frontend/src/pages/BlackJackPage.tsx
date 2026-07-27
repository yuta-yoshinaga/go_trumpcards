import { useCallback, useMemo, useState } from 'react';
import type { BlackJackBetOptions, BlackJackConfigInput } from '../api/gameApi';
import { blackjackApi, spanish21Api } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { BjActionPhaseControls } from '../components/blackjack/BjActionPhaseControls';
import { BjBetPhaseControls } from '../components/blackjack/BjBetPhaseControls';
import { BjEarlySurrenderPhaseControls } from '../components/blackjack/BjEarlySurrenderPhaseControls';
import { BjEndPhaseControls } from '../components/blackjack/BjEndPhaseControls';
import { BjInsurancePhaseControls } from '../components/blackjack/BjInsurancePhaseControls';
import {
  BJ_COUNTING_KO,
  BJ_SIDE_BET_PERFECT_PAIRS,
  BJ_SUGGEST_DECLINE_INSURANCE,
  BJ_SUGGEST_DOUBLE,
  BJ_SUGGEST_DOUBLE_STAND,
  BJ_SUGGEST_HIT,
  BJ_SUGGEST_NONE,
  BJ_SUGGEST_SPLIT,
  BJ_SUGGEST_STAND,
  BJ_SUGGEST_SURRENDER,
} from '../components/blackjack/bjConstants';
import { HandStatusBadges } from '../components/blackjack/HandStatusBadges';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { LossFeedback } from '../components/motion/LossFeedback';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnDanger } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BlackJackResponse } from '../types/card';
import { BjPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BLACKJACK_HELP, parseBlackjackCommand } from '../utils/cli/commands/blackjackCommands';
import { formatBlackjackState } from '../utils/cli/formatters/blackjackFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { getBlackjackHint } from '../utils/hints/blackjackHint';

const BJ_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BjPhase.BET]: 'bet',
  [BjPhase.DEAL]: 'deal',
  [BjPhase.INSURANCE]: 'insurance',
  [BjPhase.ACTION]: 'action',
  [BjPhase.END]: 'end',
  [BjPhase.EARLY_SURRENDER]: 'earlySurrender',
};

/**
 * Strips the i18n namespace segment from a backend bonus key so it can be
 * resolved by the page's namespaced `t()`. Backend bonus keys are the fully
 * qualified `<namespace>.<...path>` form (e.g. `spanish21.bonus.777.spade`,
 * see `BlackJackVariant.go`); `t()` here is already scoped to the variant
 * namespace, so only the leading namespace segment is removed — the rest of
 * the dotted path is preserved. Namespace-agnostic (not coupled to the literal
 * `spanish21.` prefix) so a namespace rename cannot silently drop translation.
 */
function bonusBadgeKey(fullKey: string): string {
  // indexOf('.') is -1 when there is no namespace segment, so slice(0) returns
  // the key unchanged — no branch needed.
  return fullKey.slice(fullKey.indexOf('.') + 1);
}

function useSuggestionLabels(t: (key: string) => string): Record<number, string> {
  return {
    [BJ_SUGGEST_HIT]: t('suggest.hit'),
    [BJ_SUGGEST_STAND]: t('suggest.stand'),
    [BJ_SUGGEST_DOUBLE]: t('suggest.double'),
    [BJ_SUGGEST_SPLIT]: t('suggest.split'),
    [BJ_SUGGEST_SURRENDER]: t('suggest.surrender'),
    [BJ_SUGGEST_DECLINE_INSURANCE]: t('suggest.decline'),
    // Basic-strategy "Ds" — double if allowed, otherwise stand. UI surfaces the
    // primary intent ("double") because the stand fallback is an internal state
    // used when the player can no longer double (post-split, low chips).
    [BJ_SUGGEST_DOUBLE_STAND]: t('suggest.double'),
  };
}

/** BlackJack tutorial step definitions. */
const BJ_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bj-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="bj-bet-button"]', messageKey: 'tutorial.betButton', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="bj-dealer-hand"]',
    messageKey: 'tutorial.dealerHand',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bj-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bj-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bj-result-message"]',
    messageKey: 'tutorial.resultMessage',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bj-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/**
 * Spanish 21 tutorial: the BlackJack steps plus two variant-specific stops that
 * explain the 48-card deck / bonus payout table and the bonus-achievement badges.
 */
const SPANISH21_TUTORIAL_STEPS: TutorialStep[] = [
  BJ_TUTORIAL_STEPS[0], // bet controls
  BJ_TUTORIAL_STEPS[1], // bet button
  {
    target: '[data-tutorial="bj-payout-ref"]',
    messageKey: 'tutorial.payoutRef',
    placement: 'bottom',
    advanceOn: 'next',
  },
  BJ_TUTORIAL_STEPS[2], // dealer hand
  BJ_TUTORIAL_STEPS[3], // player hand
  {
    target: '[data-tutorial="bj-bonus-badges"]',
    messageKey: 'tutorial.bonusBadges',
    placement: 'top',
    advanceOn: 'next',
  },
  BJ_TUTORIAL_STEPS[4], // action buttons
  BJ_TUTORIAL_STEPS[5], // result message
  BJ_TUTORIAL_STEPS[6], // reset button
];

/** Variant identifier shared by BlackJack and Spanish 21 (which reuses this page). */
export type BlackJackVariant = 'blackjack' | 'spanish21';

/** Props for {@link BlackJackPage}. */
export interface BlackJackPageProps {
  /** Variant of BlackJack to render. Defaults to 'blackjack'. */
  variant?: BlackJackVariant;
}

/** Renders the BlackJack game page with betting, action, and end phases. */
export function BlackJackPage({ variant = 'blackjack' }: BlackJackPageProps) {
  const steps = variant === 'spanish21' ? SPANISH21_TUTORIAL_STEPS : BJ_TUTORIAL_STEPS;
  return (
    <TutorialWrapper gameName={variant} steps={steps}>
      <BlackJackPageContent variant={variant} />
    </TutorialWrapper>
  );
}

/** Inner content of the BlackJack page, wrapped by TutorialProvider. */
function BlackJackPageContent({ variant = 'blackjack' }: BlackJackPageProps) {
  const apiClient = variant === 'spanish21' ? spanish21Api : blackjackApi;
  const gamePath = variant === 'spanish21' ? '/spanish21' : '/';
  const navTitleKey = variant === 'spanish21' ? 'nav.spanish21' : 'nav.blackjack';
  const themeKey: 'blackjack' | 'spanish21' = variant === 'spanish21' ? 'spanish21' : 'blackjack';
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup(variant);
  const phaseNames = usePhaseNames(variant, BJ_PHASE_KEYS);
  const suggestionLabels = useSuggestionLabels(t);

  const { cardWidth, isMobile } = useCardDimensions();
  const [message, setMessage] = useState('');
  const [betAmount, setBetAmount] = useState(10);
  const [dealerHitsSoft17, setDealerHitsSoft17] = useState(false);
  const [countingEnabled, setCountingEnabled] = useState(false);
  const [cpuPlayerCount, setCpuPlayerCount] = useState(0);
  const [perfectPairsBet, setPerfectPairsBet] = useState(0);
  const [twentyOnePlus3Bet, setTwentyOnePlus3Bet] = useState(0);
  const [handCount, setHandCount] = useState(1);
  const [doubleAfterSplit, setDoubleAfterSplit] = useState(true);
  const [countingSystem, setCountingSystem] = useState(0);
  const [deckPenetration, setDeckPenetration] = useState(75);
  const [surrenderRule, setSurrenderRule] = useState(0);
  const [autoAdvance, setAutoAdvance] = useState(0);

  const onSuccess = useCallback((res: BlackJackResponse) => {
    setMessage(res.message);
    setDealerHitsSoft17(res.dealerHitsSoft17);
    setCountingEnabled(res.countingEnabled);
    setCpuPlayerCount(res.cpuPlayerCount);
    setDoubleAfterSplit(res.doubleAfterSplit);
    setCountingSystem(res.countingSystem);
    setDeckPenetration(res.deckPenetration);
    setSurrenderRule(res.surrenderRule);
  }, []);
  const { state, loading, error, exec, retry } = useGameApi(apiClient.exec, { onSuccess });

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode(variant);
  const bjCliConfig: CliGameConfig<BlackJackResponse, Parameters<typeof apiClient.exec>> = useMemo(
    () => ({
      gameName: variant,
      parseCommand: parseBlackjackCommand,
      formatResponse: formatBlackjackState,
      helpText: BLACKJACK_HELP,
    }),
    [variant],
  );
  const { handleCommand } = useCliGame(exec, bjCliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(exec);

  const phase = state?.phase ?? BjPhase.BET;
  const isRoundInProgress =
    phase === BjPhase.DEAL ||
    phase === BjPhase.EARLY_SURRENDER ||
    phase === BjPhase.INSURANCE ||
    phase === BjPhase.ACTION;
  const hands = state?.hands ?? [];
  const currentHandIdx = state?.currentHandIdx ?? 0;
  const currentHand = hands[currentHandIdx];
  const playerChips = state?.player?.chips ?? 0;
  const hintEnabled = state?.hintEnabled ?? false;
  const suggestedAction = state?.suggestedAction ?? BJ_SUGGEST_NONE;
  // BlackJack uses backend-driven hintEnabled/suggestedAction for hint banner.
  // getBlackjackHint adapts that into HintResult for the reasoning tooltip.
  const bjHintResult = hintEnabled && state ? getBlackjackHint(state) : null;
  const cpuPlayers = state?.cpuPlayers ?? [];
  const sideBetResults = state?.sideBetResults ?? [];

  const showDoubleDown =
    !!currentHand &&
    currentHand.cards.length === 2 &&
    playerChips >= currentHand.bet &&
    ((hands?.length ?? 0) <= 1 || doubleAfterSplit);
  const showSplit = !!currentHand?.canSplit && playerChips >= currentHand.bet;
  const showSurrender = !!currentHand?.canSurrender;

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: () => exec('hit'), label: 'hit' },
      { key: 's', action: () => exec('stand'), label: 'stand' },
      { key: 'd', action: () => exec('doubledown'), enabled: showDoubleDown, label: 'doubledown' },
      { key: 'p', action: () => exec('split'), enabled: showSplit, label: 'split' },
      { key: 'u', action: () => exec('surrender'), enabled: showSurrender, label: 'surrender' },
    ],
    [exec, showDoubleDown, showSplit, showSurrender],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: phase === BjPhase.ACTION && !loading,
  });

  const insuranceBindings = useMemo(
    () => [
      { key: 'i', action: () => exec('insurance'), label: 'insurance' },
      { key: 'n', action: () => exec('declineinsurance'), label: 'declineinsurance' },
    ],
    [exec],
  );

  useActionKeyboardNav({
    bindings: insuranceBindings,
    enabled: phase === BjPhase.INSURANCE && !loading,
  });

  const earlySurrenderBindings = useMemo(
    () => [
      { key: 'u', action: () => exec('earlysurrender'), label: 'earlysurrender' },
      { key: 'n', action: () => exec('declineearlysurrender'), label: 'declineearlysurrender' },
    ],
    [exec],
  );

  useActionKeyboardNav({
    bindings: earlySurrenderBindings,
    enabled: phase === BjPhase.EARLY_SURRENDER && !loading,
  });

  const handleReset = useCallback(() => {
    hideActionLog();
    const config: BlackJackConfigInput = {
      dealerHitsSoft17,
      cpuPlayerCount,
      countingEnabled,
      doubleAfterSplit,
      countingSystem,
      deckPenetration,
      surrenderRule,
    };
    exec('reset', undefined, config);
  }, [
    exec,
    dealerHitsSoft17,
    cpuPlayerCount,
    countingEnabled,
    doubleAfterSplit,
    countingSystem,
    deckPenetration,
    surrenderRule,
    hideActionLog,
  ]);

  // Manual "Next Game" clicks route through the shared reset-confirm dialog so a stray tap
  // does not discard the session. The auto-advance countdown keeps firing handleReset directly
  // (see BjEndPhaseControls), so the automatic next-round flow is unaffected.
  const requestResetConfirm = useCallback(() => requestConfirm(handleReset), [requestConfirm, handleReset]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="blackjack"
        layout={{ kind: 'casino-table', sections: [5], footerStyle: 'hand', footerHandSize: 5 }}
      />
    );

  return (
    <GamePageShell
      title={tc(navTitleKey)}
      gameThemeBg={gameTheme[themeKey].bg}
      phaseName={phaseNames[phase] ?? t('phase.bet')}
      isHumanTurn={
        phase === BjPhase.ACTION || phase === BjPhase.EARLY_SURRENDER || phase === BjPhase.INSURANCE
          ? true
          : phase === BjPhase.END
            ? false
            : undefined
      }
      gamePath={gamePath}
      gameEndFlag={!isRoundInProgress}
      winShow={phase === BjPhase.END}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>
            {t('player')} {state.player.chips} chips
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
      headerEnd={
        <>
          <span>
            {t('deck')} {t('deckUnit', { count: state.deckCount })}
          </span>
          {countingEnabled && (
            <span>
              {t(`countingSystemNames.${countingSystem}`)} RC={state.runningCount}{' '}
              {countingSystem === BJ_COUNTING_KO ? t('trueCountNA') : `TC=${state.trueCount.toFixed(1)}`}
            </span>
          )}
          <span>
            {tc('label.dealer')} {state.dealer.chips} chips
          </span>
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Scrollable: dealer area + CPU players */}
          <div
            data-testid="card-area"
            className={[`overflow-y-auto p-4 lg:px-8 ${lgCardAreaConstraint}`, phase !== BjPhase.BET && 'flex-1']
              .filter(Boolean)
              .join(' ')}
          >
            {phase === BjPhase.BET && (
              <div className="flex flex-col items-center justify-center py-6 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
                <details className="bg-black/30 rounded-lg w-full max-w-sm" data-tutorial="bj-payout-ref">
                  <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                    {t('payoutRef.title')}
                  </summary>
                  <ul className="text-ds-text-muted text-sm space-y-1 px-4 pb-3">
                    {(variant === 'spanish21'
                      ? ([
                          'blackjack',
                          'win',
                          'insurance',
                          'push',
                          'surrender',
                          'bust',
                          'bonusFiveCard21',
                          'bonusSixCard21',
                          'bonusSevenCard21',
                          'bonus678',
                          'bonus777',
                        ] as const)
                      : (['blackjack', 'win', 'insurance', 'push', 'surrender', 'bust'] as const)
                    ).map((key) => (
                      <li key={key}>{t(`payoutRef.${key}`)}</li>
                    ))}
                  </ul>
                </details>
              </div>
            )}
            {phase !== BjPhase.BET && (
              <div data-tutorial="bj-dealer-hand">
                <h2 className="text-ds-text-primary">
                  {t('dealerHand')}
                  {dealerHitsSoft17 ? ' (H17)' : ' (S17)'}
                </h2>
                <p className="text-ds-text-primary">
                  {t('score')} {state.dealer.score ? state.dealer.score : ''}
                </p>
                <div className="flex flex-wrap gap-2">
                  {state.dealer.cards?.map((card, idx) => (
                    <AnimatedCard
                      key={`dealer-${idx}-${card.design}-${card.value}`}
                      card={card}
                      width={cardWidth}
                      dealDelay={idx * 0.2}
                    />
                  ))}
                  {!state.dealer.score && <AnimatedCardBack width={cardWidth} />}
                </div>
              </div>
            )}

            {/* CPU players */}
            {phase !== BjPhase.BET && cpuPlayers.length > 0 && (
              <div className="mt-4">
                {cpuPlayers.map((cpu, cpuIdx) => (
                  <div key={cpuIdx} className="mb-3">
                    <h2 className="text-ds-accent mt-0 mb-1">
                      {tc('player.cpu', { id: cpuIdx + 1 })} ({cpu.chips} chips)
                      {cpu.insuranceBet > 0 && (
                        <span className="text-ds-warning text-sm ml-2">
                          [{t('insurance')} {cpu.insuranceBet}]
                        </span>
                      )}
                    </h2>
                    {cpu.hands.map((hand, handIdx) => (
                      <div key={handIdx} className="mb-1">
                        <div className="text-ds-warning text-sm">
                          {cpu.hands.length > 1 ? `${t('hand', { idx: handIdx + 1 })} ` : ''}
                          {t('score')} {hand.score} / {tc('betting.currentBet')} {hand.bet}
                          <HandStatusBadges
                            busted={hand.busted}
                            doubled={hand.doubled}
                            isBlackJack={hand.isBlackJack}
                            surrendered={hand.surrendered}
                          />
                        </div>
                        <div className="flex flex-wrap gap-1">
                          {hand.cards.map((card, cardIdx) => (
                            <AnimatedCard
                              key={`cpu${cpuIdx}-hand${handIdx}-${cardIdx}-${card.design}-${card.value}`}
                              card={card}
                              width={cardWidth}
                            />
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Sticky footer: player hand + result + buttons */}
          <GameFooter className={`${gameTheme[themeKey].footer} px-4 py-3`}>
            {/* Player hands */}
            {phase !== BjPhase.BET && hands.length > 0 && (
              <div className="mb-2" data-tutorial="bj-player-hand">
                {hands.map((hand, handIndex) => (
                  <div key={`hand-${handIndex}`} className="mb-2">
                    <h2 className="text-ds-text-primary mt-0 mb-0.5">
                      {hands.length > 1 ? t('hand', { idx: handIndex + 1 }) : t('playerHand')}
                      {handIndex === currentHandIdx &&
                        (phase === BjPhase.ACTION || phase === BjPhase.EARLY_SURRENDER) &&
                        ' (*)'}
                      <HandStatusBadges
                        busted={hand.busted}
                        doubled={hand.doubled}
                        isBlackJack={hand.isBlackJack}
                        surrendered={hand.surrendered}
                      />
                    </h2>
                    <p className="text-ds-text-primary mt-0 mb-0.5">
                      {t('score')} {hand.score} / {tc('betting.currentBet')} {hand.bet}
                    </p>
                    <div className="flex flex-wrap gap-1.5">
                      {hand.cards.map((card, cardIdx) => (
                        <AnimatedCard
                          key={`hand-${handIndex}-${cardIdx}-${card.design}-${card.value}`}
                          card={card}
                          width={cardWidth}
                          dealDelay={cardIdx * 0.12}
                        />
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Insurance info */}
            {state.insuranceBet > 0 && (
              <div className="text-ds-warning text-sm mb-1">
                {t('insurance')} {state.insuranceBet}
              </div>
            )}

            {/* Side bet results */}
            {sideBetResults.length > 0 && (
              <div className="mb-2">
                {sideBetResults.map((r) => (
                  <div
                    key={r.betType}
                    className={`text-sm text-center px-3 py-1 rounded mb-1 ${r.payout > 0 ? 'bg-ds-warning/90 text-ds-text-on-accent font-bold' : 'bg-ds-surface-elevated/70 text-ds-text-primary'}`}
                  >
                    {r.betType === BJ_SIDE_BET_PERFECT_PAIRS ? t('sideBet.perfectPairs') : t('sideBet.twentyOnePlus3')}:{' '}
                    {r.payout > 0
                      ? t('sideBet.win', { name: r.resultName, payout: r.payout })
                      : t('sideBet.lose', { name: r.resultName, amount: r.betAmount })}
                  </div>
                ))}
              </div>
            )}

            {/* Variant bonus badges (Spanish 21): 7-7-7 / 6-7-8 / 5+card 21 achievements. */}
            {phase === BjPhase.END && (state?.bonuses?.length ?? 0) > 0 && (
              <div
                className="mb-2 flex flex-wrap justify-center gap-1"
                data-testid="bj-bonus-badges"
                data-tutorial="bj-bonus-badges"
              >
                {state?.bonuses?.map((key, i) => (
                  <span
                    // Multi-hand rounds can repeat the same bonus key, so include the index.
                    key={`${key}-${i}`}
                    data-testid="bj-bonus-badge"
                    className="rounded-full border border-ds-warning bg-ds-surface px-3 py-0.5 text-ds-warning text-sm font-bold motion-safe:animate-pulse-once"
                  >
                    🎉 {t(bonusBadgeKey(key))}
                  </span>
                ))}
              </div>
            )}

            {/* Hint banner */}
            {hintEnabled && suggestedAction !== BJ_SUGGEST_NONE && (
              <div className="mb-2">
                <div className="bg-ds-warning/90 text-ds-text-on-accent text-center text-sm font-bold px-3 py-1 rounded">
                  {t('suggestion')} {suggestionLabels[suggestedAction]}
                </div>
                {bjHintResult && <HintTooltip reason={t(bjHintResult.reason)} confidence={bjHintResult.confidence} />}
              </div>
            )}

            {/* Result message */}
            <div data-tutorial="bj-result-message">
              <GameMessageBox
                message={message}
                messageCode={state?.messageCode}
                messageParams={state?.messageParams}
                severity={phase === BjPhase.END ? 'alert' : 'info'}
              />
            </div>

            {/* Action log */}
            <ActionLogSection
              isEndPhase={phase === BjPhase.END}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />

            <ErrorAlert message={error} onRetry={retry} />

            {/* Phase-based buttons */}
            <div className="text-center">
              {phase === BjPhase.BET && playerChips <= 0 && (
                <button type="button" className={btnDanger} onClick={handleReset} disabled={loading}>
                  {t('outOfChips')}
                </button>
              )}
              {phase === BjPhase.BET && playerChips > 0 && (
                <>
                  <div data-tutorial="bj-bet-controls">
                    <BjBetPhaseControls
                      betAmount={betAmount}
                      onBetAmountChange={setBetAmount}
                      playerChips={playerChips}
                      deckCount={state?.deckCount ?? 1}
                      onDeckCountChange={(v) => exec('setdeckcount', v)}
                      cpuPlayerCount={cpuPlayerCount}
                      onCpuPlayerCountChange={(v) => exec('setcpucount', v)}
                      hintEnabled={hintEnabled}
                      onToggleHint={() => exec('togglehint')}
                      dealerHitsSoft17={dealerHitsSoft17}
                      onToggleSoft17={() => exec('togglesoft17')}
                      countingEnabled={countingEnabled}
                      onToggleCounting={() => exec('togglecounting')}
                      doubleAfterSplit={doubleAfterSplit}
                      onToggleDAS={() => exec('toggledas')}
                      countingSystem={countingSystem}
                      onCountingSystemChange={(v) => exec('setcountingsystem', v)}
                      deckPenetration={deckPenetration}
                      onDeckPenetrationChange={(v) => exec('setpenetration', v)}
                      surrenderRule={surrenderRule}
                      onSurrenderRuleChange={(v) => exec('setsurrenderrule', v)}
                      handCount={handCount}
                      onHandCountChange={setHandCount}
                      loading={loading}
                      onBet={() => {
                        const betOptions: BlackJackBetOptions = {};
                        if (perfectPairsBet > 0) betOptions.perfectPairsBet = perfectPairsBet;
                        if (twentyOnePlus3Bet > 0) betOptions.twentyOnePlus3Bet = twentyOnePlus3Bet;
                        if (handCount > 1) betOptions.handCount = handCount;
                        exec('bet', betAmount, undefined, betOptions);
                      }}
                      perfectPairsBet={perfectPairsBet}
                      onPerfectPairsBetChange={setPerfectPairsBet}
                      twentyOnePlus3Bet={twentyOnePlus3Bet}
                      onTwentyOnePlus3BetChange={setTwentyOnePlus3Bet}
                      autoExpandAdvanced={!isMobile}
                    />
                  </div>
                  <div className="flex items-center justify-center gap-2 mt-2">
                    <label htmlFor="bj-auto-advance" className="text-ds-text-primary text-sm">
                      {t('autoAdvance')}
                    </label>
                    <select
                      id="bj-auto-advance"
                      value={autoAdvance}
                      onChange={(e) => setAutoAdvance(Number(e.target.value))}
                      className="px-3 py-2 rounded text-sm min-h-[44px]"
                    >
                      <option value={0}>OFF</option>
                      <option value={3}>{t('autoAdvanceSec', { sec: 3 })}</option>
                      <option value={5}>{t('autoAdvanceSec', { sec: 5 })}</option>
                      <option value={10}>{t('autoAdvanceSec', { sec: 10 })}</option>
                    </select>
                  </div>
                </>
              )}

              {phase === BjPhase.INSURANCE && (
                <BjInsurancePhaseControls
                  loading={loading}
                  hintEnabled={hintEnabled}
                  suggestedAction={suggestedAction}
                  onInsurance={() => exec('insurance')}
                  onDecline={() => exec('declineinsurance')}
                />
              )}

              {phase === BjPhase.ACTION && (
                <div data-tutorial="bj-action-buttons">
                  <BjActionPhaseControls
                    loading={loading}
                    hintEnabled={hintEnabled}
                    suggestedAction={suggestedAction}
                    showDoubleDown={showDoubleDown}
                    showSplit={showSplit}
                    showSurrender={showSurrender}
                    onHit={() => exec('hit')}
                    onStand={() => exec('stand')}
                    onDoubleDown={() => exec('doubledown')}
                    onSplit={() => exec('split')}
                    onSurrender={() => exec('surrender')}
                  />
                </div>
              )}

              {phase === BjPhase.EARLY_SURRENDER && (
                <BjEarlySurrenderPhaseControls
                  loading={loading}
                  hintEnabled={hintEnabled}
                  suggestedAction={suggestedAction}
                  onSurrender={() => exec('earlysurrender')}
                  onContinue={() => exec('declineearlysurrender')}
                />
              )}

              {phase === BjPhase.END && (
                <div data-tutorial="bj-reset-button">
                  <BjEndPhaseControls
                    loading={loading}
                    onReset={handleReset}
                    onRequestReset={requestResetConfirm}
                    autoAdvanceSeconds={autoAdvance > 0 ? autoAdvance : undefined}
                  />
                </div>
              )}
            </div>
            <ActionShortcutsPanel
              bindings={[...actionBindings, ...insuranceBindings, ...earlySurrenderBindings]}
              data-testid="black-jack-kbd-shortcuts"
            />
          </GameFooter>
        </>
      )}
      <LossFeedback show={hands.some((h) => h.busted)} />
    </GamePageShell>
  );
}
