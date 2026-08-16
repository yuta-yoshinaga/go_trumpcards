import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { fiveCardStudApi, sokoApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { BettingControls } from '../components/BettingControls';
import { CpuAccordion } from '../components/CpuAccordion';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuActionToast } from '../components/CpuActionToast';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HudStats } from '../components/HudStats';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { RoundResults } from '../components/RoundResults';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { placeholderCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { FiveCardStudResponse } from '../types/card';
import { FiveCardStudPhase, FiveCardStudRebuyPhaseType } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { FIVECARDSTUD_HELP, parseFiveCardStudCommand } from '../utils/cli/commands/fiveCardStudCommands';
import { formatFiveCardStudState } from '../utils/cli/formatters/fiveCardStudFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Five Card Stud tutorial step definitions. */
const FCS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="fcs-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fcs-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fcs-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fcs-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fcs-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const FCS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [FiveCardStudPhase.SECOND_STREET]: 'secondStreet',
  [FiveCardStudPhase.THIRD_STREET]: 'thirdStreet',
  [FiveCardStudPhase.FOURTH_STREET]: 'fourthStreet',
  [FiveCardStudPhase.FIFTH_STREET]: 'fifthStreet',
  [FiveCardStudPhase.SHOWDOWN]: 'showdown',
  [FiveCardStudPhase.END]: 'end',
  [FiveCardStudPhase.REBUY]: 'rebuy',
};

/**
 * Games served by this page. Soko (Canadian Stud) is Five Card Stud with two
 * extra hand ranks -- the deal, the streets and the betting are the same game,
 * and only the showdown ranking differs, which the server already resolves into
 * `handName`. A second ~560-line copy would be duplication, not a feature.
 */
export type FcsPageGameKey = 'fivecardstud' | 'soko';

/** Renders the Five Card Stud game page with door cards, betting, and showdown. */
export const FiveCardStudPage = withTutorial(
  () => <FiveCardStudPageContent gameKey="fivecardstud" />,
  'fivecardstud',
  FCS_TUTORIAL_STEPS,
);
/** Inner content of the Five Card Stud page, wrapped by TutorialProvider. */
export function FiveCardStudPageContent({ gameKey }: { gameKey: FcsPageGameKey }) {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup(gameKey);
  const phaseNames = usePhaseNames(gameKey, FCS_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();
  // Annotated rather than inferred: both clients are created from the same
  // factory with the same generics, but a union of the two exec signatures
  // widens the response to `{}` and every field read below fails to compile.
  const api: typeof fiveCardStudApi = gameKey === 'soko' ? sokoApi : fiveCardStudApi;
  const { state, loading, error, exec: execApi, retry } = useGameApi(api.exec);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode(gameKey);
  type FcsArgs = Parameters<typeof fiveCardStudApi.exec>;

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint(gameKey, state);
  const cliConfig: CliGameConfig<FiveCardStudResponse, FcsArgs> = useMemo(
    () => ({
      gameName: gameKey,
      parseCommand: parseFiveCardStudCommand,
      formatResponse: (s: FiveCardStudResponse) => formatFiveCardStudState(s, phaseNames),
      helpText: FIVECARDSTUD_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [gameKey, phaseNames, frontendHint],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const [betAmount, setBetAmount] = useState(20);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const turnStartRef = useRef(0);

  useMountReset(execApi);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void execApi('reset', undefined, { cpuMetaAI });
  }, [execApi, hideActionLog, cpuMetaAI]);

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

  const phase = state?.phase ?? FiveCardStudPhase.INIT;
  const isActive = phase >= FiveCardStudPhase.SECOND_STREET && phase <= FiveCardStudPhase.FIFTH_STREET;
  const isShowdown = phase === FiveCardStudPhase.SHOWDOWN || phase === FiveCardStudPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isActive && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;
  const isMuckPhase = phase === FiveCardStudPhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase = phase === FiveCardStudPhase.REBUY && state?.rebuyPhaseType === FiveCardStudRebuyPhaseType.REBUY;
  const isAddonPhase = phase === FiveCardStudPhase.REBUY && state?.rebuyPhaseType === FiveCardStudRebuyPhaseType.ADDON;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const humanRebuyCount = state?.rebuyCounts?.[humanIdx] ?? 0;
  const cpuPlayers = useMemo(() => state?.players?.filter((p) => !p.isHuman) ?? [], [state?.players]);

  // Door cards revealed so far: Second Street shows 1, then +1 each street up to 4 on Fifth Street.
  const doorCount = Math.min(Math.max(phase - FiveCardStudPhase.SECOND_STREET + 1, 1), 4);

  // Celebrate only when the human wins the pot (positive wonAmount at showdown), not on a fold/loss.
  const humanWon =
    phase === FiveCardStudPhase.END &&
    (state?.roundResults?.some((r) => r.playerIdx === humanIdx && r.wonAmount > 0) ?? false);

  const actionBindings = useMemo(
    () => [
      {
        key: 'c',
        action: () => execApi('call', undefined, undefined, getElapsed()),
        enabled: hasOutstandingBet,
        label: 'call',
      },
      {
        key: 'r',
        action: () =>
          hasOutstandingBet
            ? execApi('raise', betAmount, undefined, getElapsed())
            : execApi('bet', betAmount, undefined, getElapsed()),
        label: 'raiseOrBet',
      },
      {
        key: 'k',
        action: () => execApi('check', undefined, undefined, getElapsed()),
        enabled: !hasOutstandingBet,
        label: 'check',
      },
      { key: 'f', action: () => execApi('fold', undefined, undefined, getElapsed()), label: 'fold' },
      { key: 'a', action: () => execApi('allin', undefined, undefined, getElapsed()), label: 'allin' },
    ],
    [execApi, hasOutstandingBet, betAmount, getElapsed],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: canAct && !loading,
  });

  if (!state)
    return (
      <GameSkeleton
        gameKey={gameKey}
        layout={{ kind: 'community-poker', community: 4, opponents: 3, opponentCards: 2, footerHandSize: 2 }}
      />
    );

  return (
    <GamePageShell
      title={tc(`nav.${gameKey}`)}
      gameThemeBg={gameTheme[gameKey].bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct}
      gamePath={`/${gameKey}`}
      gameEndFlag={phase === FiveCardStudPhase.SHOWDOWN || phase === FiveCardStudPhase.END}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="fcs-pot-display">
            {tc('label.pot')} <strong>{state?.pot ?? 0}</strong>
          </span>
          <span>
            {t('anteBringIn')}:{' '}
            <strong>
              {state?.ante ?? 0}/{state?.bringIn ?? 0}
            </strong>
          </span>
          <span>
            {tc('label.dealer')} <strong>{findPlayerName(state.players, state.dealerIdx)}</strong>
          </span>
          {state?.tournamentMode && (
            <span>{t('handNumber', { count: state.handCount, level: state.anteLevelHands })}</span>
          )}
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Scrollable: CPU players */}
          <div className={`flex-1 overflow-y-auto pt-4 px-5 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* CPU players */}
            <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="fcs-cpu-area">
              {cpuPlayers.map((p) => (
                <div key={p.id} className="mb-3 p-2 rounded bg-black/30">
                  <div className="text-ds-text-primary text-sm mb-1">
                    CPU {p.id}
                    <span className="ml-2 text-xs text-ds-text-muted">{p.playStyleName}</span>
                    {p.totalHands > 0 && (
                      <HudStats vpip={p.vpip} pfr={p.pfr} threeBet={p.threeBet} af={p.af} namespace={gameKey} />
                    )}
                    <span className="ml-2 text-xs">
                      {tc('betting.chips')} {p.chips}
                    </span>
                    {p.currentBet > 0 && (
                      <span className="ml-2 text-xs">
                        {tc('betting.currentBet')} {p.currentBet}
                      </span>
                    )}
                    {p.folded && <span className="ml-2 text-ds-error text-xs">[{tc('status.folded')}]</span>}
                    {p.allIn && <span className="ml-2 text-ds-warning text-xs">[{tc('status.allIn')}]</span>}
                    {isShowdown && !p.folded && p.handName && (
                      <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                        {p.handName}
                      </span>
                    )}
                  </div>
                  {/* Door cards (always visible) */}
                  <div className="text-ds-text-muted text-xs mb-0.5">{t('doorCards')}</div>
                  <div className="flex flex-wrap gap-1 mb-1">
                    {p.doorCards?.length
                      ? p.doorCards.map((card, i, arr) =>
                          i === arr.length - 1 ? (
                            <span
                              key={`${card.design}-${card.value}`}
                              className="inline-block rounded ring-2 ring-ds-accent motion-safe:animate-pulse"
                              data-testid={`latest-door-cpu-${p.id}`}
                            >
                              <AnimatedCard card={card} width={cardWidth} style={placeholderCardStyle} />
                            </span>
                          ) : (
                            <AnimatedCard
                              key={`${card.design}-${card.value}`}
                              card={card}
                              width={cardWidth}
                              style={placeholderCardStyle}
                            />
                          ),
                        )
                      : !p.folded &&
                        Array.from({ length: doorCount }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
                  </div>
                  {/* Hole card (face-down unless showdown) */}
                  <div className="text-ds-text-muted text-xs mb-0.5">{t('holeCards')}</div>
                  <div className="flex flex-wrap gap-1">
                    {isShowdown && !p.folded && p.holeCards?.length
                      ? p.holeCards.map((card) => (
                          <AnimatedCard
                            key={`${card.design}-${card.value}`}
                            card={card}
                            width={cardWidth}
                            style={placeholderCardStyle}
                          />
                        ))
                      : !p.folded && <AnimatedCardBack width={cardWidth} />}
                  </div>
                </div>
              ))}
            </CpuAccordion>

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

          {/* Settings */}
          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'checkbox' as const,
                    id: 'cpuMetaAI',
                    label: t('settings.cpuMetaAI'),
                    checked: cpuMetaAI,
                    onToggle: setCpuMetaAI,
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          {/* Sticky footer: player hand + buttons */}
          <GameFooter className={`${gameTheme[gameKey].footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="fcs-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
                  {t('yourHand')}
                  {humanPlayer.totalHands > 0 && (
                    <HudStats
                      vpip={humanPlayer.vpip}
                      pfr={humanPlayer.pfr}
                      threeBet={humanPlayer.threeBet}
                      af={humanPlayer.af}
                      namespace={gameKey}
                    />
                  )}
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
                  {isShowdown && !humanPlayer.folded && humanPlayer.handName && (
                    <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                      {humanPlayer.handName}
                    </span>
                  )}
                </div>
                {/* Door cards */}
                <div className="text-ds-text-muted text-xs mb-0.5">{t('doorCards')}</div>
                <div className="flex flex-wrap gap-1.5 mb-1">
                  {humanPlayer.doorCards?.length
                    ? humanPlayer.doorCards.map((card, i, arr) =>
                        i === arr.length - 1 ? (
                          <span
                            key={`${card.design}-${card.value}`}
                            className="inline-block rounded ring-2 ring-ds-accent motion-safe:animate-pulse"
                            data-testid="latest-door-human"
                          >
                            <AnimatedCard card={card} width={cardWidth} style={placeholderCardStyle} />
                          </span>
                        ) : (
                          <AnimatedCard
                            key={`${card.design}-${card.value}`}
                            card={card}
                            width={cardWidth}
                            style={placeholderCardStyle}
                          />
                        ),
                      )
                    : !humanPlayer.folded &&
                      Array.from({ length: doorCount }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
                </div>
                {/* Hole card */}
                <div className="text-ds-text-muted text-xs mb-0.5">{t('holeCards')}</div>
                <div className="flex flex-wrap gap-1.5 mb-2">
                  {humanPlayer.holeCards?.length
                    ? humanPlayer.holeCards.map((card) => (
                        <AnimatedCard
                          key={`${card.design}-${card.value}`}
                          card={card}
                          width={cardWidth}
                          style={placeholderCardStyle}
                        />
                      ))
                    : !humanPlayer.folded && <AnimatedCardBack width={cardWidth} />}
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

            {/* Muck/Show controls */}
            {isMuckPhase && (
              <div className="mb-2 text-center" data-testid="muck-controls">
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={`${btnPrimary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => execApi('muck')}
                  >
                    {t('muck.muck')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => execApi('show')}
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
                    onClick={() => execApi('rebuy')}
                  >
                    {t('rebuy.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => execApi('skiprebuy')}
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
                    onClick={() => execApi('addon')}
                  >
                    {t('addon.accept')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-w-[90px]`}
                    disabled={loading}
                    onClick={() => execApi('skipaddon')}
                  >
                    {t('addon.skip')}
                  </button>
                </div>
              </div>
            )}

            {/* Betting controls */}
            {canAct && (
              <div data-tutorial="fcs-action-buttons">
                <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

                <BettingControls
                  inputId="fiveCardStudBetAmount"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={minRaise}
                  maxBetAmount={state?.maxBetAmount}
                  potSize={state?.pot}
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

            <GameResetButton
              isGameEnd={phase === FiveCardStudPhase.SHOWDOWN || phase === FiveCardStudPhase.END}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
              dataTutorial="fcs-reset-button"
              className="min-w-[90px]"
            />
            <ActionShortcutsPanel bindings={actionBindings} data-testid="five-card-stud-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
