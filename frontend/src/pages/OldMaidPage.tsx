import { useCallback, useMemo, useState } from 'react';
import type { oldmaidApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ReplaySpeedSettingsPanel } from '../components/common/ReplaySpeedSettingsPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { OldMaidDiscardedArea } from '../components/oldmaid/OldMaidDiscardedArea';
import { OldMaidDrawHistory } from '../components/oldmaid/OldMaidDrawHistory';
import { OldMaidPlayerArea } from '../components/oldmaid/OldMaidPlayerArea';
import { OldMaidSettingsDialog } from '../components/oldmaid/OldMaidSettingsDialog';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { OldMaidMode, useOldMaidGame } from '../hooks/useOldMaidGame';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CpuAction, OldMaidResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardLabel } from '../utils/cardUtils';
import { OLDMAID_HELP, parseOldmaidCommand } from '../utils/cli/commands/oldmaidCommands';
import { formatOldmaidState } from '../utils/cli/formatters/oldmaidFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Old Maid tutorial step definitions. */
const OM_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="om-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="om-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="om-draw-button"]',
    messageKey: 'tutorial.drawButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="om-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Old Maid game page with settings dialog, player areas, and draw history. */
export const OldMaidPage = withTutorial(OldMaidPageContent, 'oldmaid', OM_TUTORIAL_STEPS);
/** Inner content of the Old Maid page, wrapped by TutorialProvider. */
function OldMaidPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('oldmaid');
  const { playSound } = useSound();
  const {
    displayState,
    setupMode,
    setupStrategy,
    setupMemoryAI,
    setupHesitation,
    setupMetaAI,
    gameSettings,
    suspectPins,
    setSuspectPins,
    shakeKey,
    revealedCard,
    loading,
    error,
    retry,
    gameExec,
    handleStart,
    handleReset,
    handleReorder,
    setSetupMode,
    setSetupStrategy,
    setSetupMemoryAI,
    setSetupHesitation,
    setSetupMetaAI,
  } = useOldMaidGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('oldmaid', displayState);

  const [settingsOpen, setSettingsOpen] = useState(false);

  const syncSetupFromSettings = () => {
    if (gameSettings) {
      setSetupMode(gameSettings.mode);
      setSetupStrategy(gameSettings.cpuPlacementStrategy);
      setSetupMemoryAI(gameSettings.cpuMemoryAI);
      setSetupHesitation(gameSettings.cpuHesitationEnabled);
      setSetupMetaAI(gameSettings.cpuMetaAI);
    }
  };

  const { cardWidth, isMobile } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('oldmaid');
  const cliConfig: CliGameConfig<OldMaidResponse, Parameters<typeof oldmaidApi.exec>> = useMemo(
    () => ({
      gameName: 'oldmaid',
      parseCommand: parseOldmaidCommand,
      formatResponse: formatOldmaidState,
      helpText: OLDMAID_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, displayState, { addInput, addOutput, addError, clearLog });

  const isHumanTurnForKbd =
    !!displayState && !displayState.gameEndFlag && !!displayState.players[displayState.currentTurn]?.isHuman;

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: () => gameExec('draw') },
      { key: 's', action: () => gameExec('shuffle') },
    ],
    [gameExec],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: isHumanTurnForKbd && !loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  if (!displayState)
    return (
      <GameSkeleton
        gameKey="oldmaid"
        layout={{
          kind: 'trick-taking',
          titleBar: false,
          opponents: 3,
          opponentStyle: 'hand',
          opponentHandSize: 3,
          footerHandSize: 5,
          footerButton: 'wide',
        }}
      />
    );

  const state = displayState;
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const humanPlayer = state.players.find((p) => p.isHuman);

  // When the human just drew a card into their own hand, locate where it landed
  // so the footer hand can briefly ring-highlight that position (issue #2984).
  // Returns -1 when the card was immediately discarded as a pair (not in hand),
  // when the drawer was not the human, or at game end.
  const drawnCard = state.lastDrawCard;
  const humanDrawnCardIdx =
    humanPlayer && drawnCard && !state.gameEndFlag && state.hasDrawn && state.lastDrawPlayerIdx === humanPlayer.id
      ? (humanPlayer.cards ?? []).findIndex((c) => c.design === drawnCard.design && c.value === drawnCard.value)
      : -1;

  const statusLines: string[] = [];
  if (!state.gameEndFlag && state.hasDrawn) {
    const from = findPlayerName(state.players, state.lastDrawPlayerIdx);
    const target = findPlayerName(state.players, state.lastDrawFromIdx);
    let msg = state.lastDrawCard
      ? t('drewCardWithLabel', { from, target, card: cardLabel(state.lastDrawCard) })
      : t('drewCard', { from, target });
    if (state.lastDiscardedPairs > 0) msg += t('discardedPairs', { count: state.lastDiscardedPairs });
    statusLines.push(msg);
  }
  if (isHumanTurn) {
    statusLines.push(t('yourTurn', { target: findPlayerName(state.players, state.nextDrawTargetIdx) }));
  }

  return (
    <GamePageShell
      title={tc('nav.oldmaid')}
      gameThemeBg={`${gameTheme.oldmaid.bg}${shakeKey > 0 ? ' animate-shake' : ''}`}
      phaseName={state.gameEndFlag ? t('phase.end') : t('phase.play')}
      isHumanTurn={isHumanTurn}
      gamePath="/oldmaid"
      gameEndFlag={!!state.gameEndFlag}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      outerKey={shakeKey}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={t('button.settings')}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />
          <ReplaySpeedSettingsPanel />
          {/* Scrollable: CPU rows + discard + status + logs + result */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Mode badge */}
            {state.mode === OldMaidMode.JijiNuki && (
              <div className="text-center mb-1">
                <span className="inline-block rounded-md bg-ds-error px-2.5 py-0.5 text-sm font-bold text-white">
                  {t('badge.jijiNuki')}
                </span>
              </div>
            )}

            {/* CPU row */}
            <div className="flex gap-2 flex-wrap mb-2 justify-center" data-tutorial="om-cpu-area">
              {cpuPlayers.map((player) => {
                // Surface a transient bubble on the CPU that just drew so the
                // player can visually track who acted without reading the log.
                // See issue #1490.
                const isDrawer =
                  !state.gameEndFlag && state.hasDrawn && state.lastDrawPlayerIdx === player.id && !player.isHuman;
                const bubble = isDrawer
                  ? {
                      message:
                        state.lastDiscardedPairs > 0
                          ? t('bubble.drewAndPaired', {
                              from: findPlayerName(state.players, state.lastDrawFromIdx),
                            })
                          : t('bubble.drewFrom', {
                              from: findPlayerName(state.players, state.lastDrawFromIdx),
                            }),
                      // Prepend the monotonic drawHistory length so identical
                      // back-to-back draws (same player, same target, same
                      // card bouncing back, e.g. Joker) still re-trigger the
                      // animation. Go Fish gets this for free via turnNumber.
                      triggerKey: `${state.drawHistory?.length ?? 0}-${state.lastDrawPlayerIdx}-${state.lastDrawFromIdx}-${state.lastDrawCard?.design ?? 'x'}-${state.lastDrawCard?.value ?? 0}`,
                    }
                  : undefined;
                return (
                  <OldMaidPlayerArea
                    key={player.id}
                    player={player}
                    isTarget={state.nextDrawTargetIdx === player.id}
                    isHumanTurn={isHumanTurn}
                    gameEndFlag={state.gameEndFlag}
                    loading={loading}
                    highlightedCardIdx={state.nextDrawTargetIdx === player.id ? state.cpuHighlightedCardIdx : -1}
                    isSuspect={suspectPins.has(player.id)}
                    compactNonTarget={isMobile}
                    onToggleSuspect={() =>
                      setSuspectPins((prev) => {
                        const next = new Set(prev);
                        if (next.has(player.id)) {
                          next.delete(player.id);
                        } else {
                          next.add(player.id);
                        }
                        return next;
                      })
                    }
                    onDraw={(drawIdx) => gameExec('draw', drawIdx)}
                    bubble={bubble}
                  />
                );
              })}
            </div>

            {/* Discarded Area */}
            <OldMaidDiscardedArea cards={state.lastDiscardedCards} />

            {/* Card reveal area */}
            {state.lastDrawCard && !state.gameEndFlag && (
              <div className="flex justify-center my-2" data-testid="card-reveal-area">
                {revealedCard ? (
                  <div className="animate-flipIn">
                    <AnimatedCard card={revealedCard} width={cardWidth} />
                  </div>
                ) : (
                  <AnimatedCardBack width={cardWidth} />
                )}
              </div>
            )}

            {/* Status */}
            {statusLines.length > 0 && (
              <div className="bg-black/50 rounded-lg text-ds-text-primary py-2 px-3 my-2 whitespace-pre-line text-sm">
                {statusLines.join('\n')}
              </div>
            )}

            {/* CPU log */}
            {state.cpuActions && state.cpuActions.length > 0 && (
              <div className="bg-black/40 rounded-lg text-game-text-muted py-1.5 px-2.5 my-1.5 whitespace-pre-line text-xs max-h-[120px] overflow-y-auto">
                {[
                  tc('label.cpuActions'),
                  ...state.cpuActions.map((action: CpuAction) => {
                    const from = findPlayerName(state.players, action.drawPlayerIdx);
                    const target = findPlayerName(state.players, action.drawFromIdx);
                    let msg = t('drewCard', { from, target });
                    // CPU drawn card is intentionally hidden to preserve game fairness
                    if (action.discardedPairs > 0) msg += t('discardedPairs', { count: action.discardedPairs });
                    return msg;
                  }),
                ].join('\n')}
              </div>
            )}

            {/* Draw History Timeline */}
            <OldMaidDrawHistory entries={state.drawHistory ?? []} players={state.players} suspectPins={suspectPins} />

            {/* Result */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* JijiNuki: show removed card at game end */}
            {state.gameEndFlag && state.removedCard && (
              <div className="text-center my-2 text-ds-text-primary text-sm">
                {t('removedCard', { card: cardLabel(state.removedCard) })}
              </div>
            )}

            {/* Action log */}
            <ActionLogSection
              isEndPhase={state.gameEndFlag}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Sticky footer: human player hand + buttons */}
          <GameFooter className={`${gameTheme.oldmaid.footer} px-4 py-2.5`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="om-player-hand">
                <OldMaidPlayerArea
                  player={humanPlayer}
                  isTarget={false}
                  isHumanTurn={isHumanTurn}
                  gameEndFlag={state.gameEndFlag}
                  loading={loading}
                  highlightedCardIdx={-1}
                  drawnCardIdx={humanDrawnCardIdx}
                  onDraw={(drawIdx) => gameExec('draw', drawIdx)}
                  onReorder={handleReorder}
                />
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {/* Buttons */}
            <div className="text-center">
              <button
                type="button"
                className={`${btnSecondary} min-w-[80px]`}
                disabled={loading}
                onClick={() => {
                  syncSetupFromSettings();
                  setSettingsOpen(true);
                }}
              >
                {t('button.settings')}
              </button>
              <GameResetButton
                isGameEnd={!!state.gameEndFlag}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="om-reset-button"
                className="min-w-[80px]"
              />
              <span data-tutorial="om-draw-button">
                <button
                  type="button"
                  className={`${btnPrimary} min-w-[110px]`}
                  disabled={loading || !isHumanTurn || state.gameEndFlag}
                  onClick={() => gameExec('draw')}
                  // Lowercase to mean the unmodified key; only advertised while the
                  // keyboard bindings are actually active (the human's turn).
                  aria-keyshortcuts={isHumanTurn ? 'd' : undefined}
                >
                  {t('button.drawRandom')}
                </button>
              </span>
              <button
                type="button"
                className={`${btnSecondary} min-w-[110px]`}
                disabled={loading || state.gameEndFlag}
                onClick={() => gameExec('shuffle')}
                aria-keyshortcuts={isHumanTurn ? 's' : undefined}
              >
                {t('button.shuffle')}
              </button>
            </div>

            {/* Keyboard shortcut hints, shown on the human's turn (matches the d/s
                bindings in useActionKeyboardNav above). */}
            {isHumanTurn && (
              <p className="text-center text-game-text-muted text-xs mt-2" data-testid="oldmaid-key-hints">
                {t('keyHints')}
              </p>
            )}
          </GameFooter>
        </>
      )}
      <OldMaidSettingsDialog
        open={settingsOpen}
        mode={setupMode}
        cpuPlacementStrategy={setupStrategy}
        cpuMemoryAI={setupMemoryAI}
        cpuHesitationEnabled={setupHesitation}
        cpuMetaAI={setupMetaAI}
        onModeChange={setSetupMode}
        onStrategyChange={setSetupStrategy}
        onMemoryAIChange={setSetupMemoryAI}
        onHesitationChange={setSetupHesitation}
        onMetaAIChange={setSetupMetaAI}
        onApply={() => {
          setSettingsOpen(false);
          handleStart();
        }}
        onClose={() => {
          syncSetupFromSettings();
          setSettingsOpen(false);
        }}
      />
    </GamePageShell>
  );
}
