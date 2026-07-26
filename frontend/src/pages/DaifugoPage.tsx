import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { daifugoApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardNavShortcutsPanel } from '../components/CardNavShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ReplaySpeedSettingsPanel } from '../components/common/ReplaySpeedSettingsPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { DaifugoCpuArea } from '../components/daifugo/DaifugoCpuArea';
import { DaifugoCpuCompact } from '../components/daifugo/DaifugoCpuCompact';
import { DaifugoExchangeLog } from '../components/daifugo/DaifugoExchangeLog';
import { DaifugoHumanArea } from '../components/daifugo/DaifugoHumanArea';
import { DaifugoRulesBadges } from '../components/daifugo/DaifugoRulesBadges';
import { DaifugoSettingsPanel } from '../components/daifugo/DaifugoSettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCardSwipeSelection } from '../hooks/useCardSwipeSelection';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useDaifugoGame } from '../hooks/useDaifugoGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DaifugoAction, DaifugoResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardLabel } from '../utils/cardUtils';
import { DAIFUGO_HELP, parseDaifugoCommand } from '../utils/cli/commands/daifugoCommands';
import { formatDaifugoState } from '../utils/cli/formatters/daifugoFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Daifugo tutorial step definitions. */
const DF_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="df-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="df-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="df-play-pass"]',
    messageKey: 'tutorial.playPass',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="df-rules-badges"]',
    messageKey: 'tutorial.rulesBadges',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="df-sort-buttons"]',
    messageKey: 'tutorial.sortButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="df-special-actions"]',
    messageKey: 'tutorial.specialActions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="df-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Daifugo game page with card play, revolution, and rule settings. */
export const DaifugoPage = withTutorial(DaifugoPageContent, 'daifugo', DF_TUTORIAL_STEPS);
/** Inner content of the Daifugo page, wrapped by TutorialProvider. */
function DaifugoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('daifugo');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    selectedIndices,
    toggleCardSelection,
    clearSelection,
    configInput,
    handleDragCard,
    handleDrop,
    handleConfigChange,
  } = useDaifugoGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('daifugo');
  const daifugoCliConfig: CliGameConfig<DaifugoResponse, Parameters<typeof daifugoApi.exec>> = useMemo(
    () => ({
      gameName: 'daifugo',
      parseCommand: parseDaifugoCommand,
      formatResponse: formatDaifugoState,
      helpText: DAIFUGO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, daifugoCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('daifugo', state);

  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();

  // Transient toast announcing a newly-triggered rank inversion (revolution /
  // eleven-back), on top of the background-color change which can be missed.
  const [inversionToast, setInversionToast] = useState<'revolution' | 'elevenBack' | null>(null);
  const prevInversion = useRef({ revolution: false, elevenBack: false });
  const inversionToastTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => () => clearTimeout(inversionToastTimer.current ?? undefined), []);
  useEffect(() => {
    const rev = state?.revolutionActive ?? false;
    const eb = state?.elevenBackActive ?? false;
    const prev = prevInversion.current;
    const newlyTriggered =
      (rev && !prev.revolution && 'revolution') || (eb && !prev.elevenBack && 'elevenBack') || null;
    prevInversion.current = { revolution: rev, elevenBack: eb };
    if (!newlyTriggered) return;
    setInversionToast(newlyTriggered);
    clearTimeout(inversionToastTimer.current ?? undefined);
    inversionToastTimer.current = setTimeout(() => setInversionToast(null), 3000);
  }, [state?.revolutionActive, state?.elevenBackActive]);

  const isHumanTurnForKbd = !!state && !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const kbdConfirm = useCallback(() => {
    if (!loading && isHumanTurnForKbd && selectedIndices.length > 0) {
      exec(
        'play',
        [...selectedIndices].sort((a, b) => a - b),
      );
    }
  }, [loading, isHumanTurnForKbd, selectedIndices, exec]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCardSelection,
    onConfirm: kbdConfirm,
    onClear: clearSelection,
    enabled: isHumanTurnForKbd && !loading,
  });

  const { onPointerDown: handleDaifugoSwipeStart } = useCardSwipeSelection({
    selected: selectedIndices,
    toggle: toggleCardSelection,
    enabled: isHumanTurnForKbd && !loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void exec('reset', [], configInput);
  }, [exec, configInput, hideActionLog]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="daifugo"
        layout={{
          kind: 'trick-taking',
          titleBar: false,
          opponents: 3,
          opponentStyle: 'hand',
          opponentHandSize: 4,
          trickArea: true,
          footerHandSize: 5,
          footerButton: 'wide',
        }}
      />
    );

  const pendingAction = state.pendingAction ?? 'none';
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  // When following a non-empty table, a play must match the table combo's card
  // count — a count mismatch is illegal regardless of rank or revolution/eleven-back
  // inversion, so we can flag it client-side. Strength is still validated server-side.
  const tableCount = state.tableCards?.length ?? 0;
  const countMismatch =
    pendingAction === 'none' && tableCount > 0 && selectedIndices.length > 0 && selectedIndices.length !== tableCount;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const humanPlayer = state.players.find((p) => p.isHuman);

  // When the strength order is inverted (revolution or eleven-back), recolor the page
  // background so the player sees the rule change at a glance instead of reading the badge.
  const inversionActive = state.revolutionActive || state.elevenBackActive;
  const bgClass = inversionActive ? 'bg-game-bg-revolution' : gameTheme.daifugo.bg;
  const footerClass = inversionActive ? 'bg-game-bg-revolution-dark border-white/20' : gameTheme.daifugo.footer;

  let playButtonLabel = t('playButton');
  let pendingBanner: string | null = null;
  if (pendingAction === 'sevenPass') {
    playButtonLabel = t('passButton');
    const targetName = findPlayerName(state.players, state.pendingActionTarget);
    pendingBanner = t('sevenPassBanner', { target: targetName });
  } else if (pendingAction === 'tenDiscard') {
    playButtonLabel = t('discardButton');
    pendingBanner = t('tenDiscardBanner');
  } else if (pendingAction === 'queenBomber') {
    pendingBanner = t('queenBomberBanner');
  }

  const actionDescription = (players: { id: number; isHuman: boolean }[], action: DaifugoAction): string => {
    if (!action.playedCards || action.playedCards.length === 0) {
      return t('actionPassed', { name: findPlayerName(players, action.playerIdx) });
    }
    const cards = action.playedCards.map(cardLabel).join(', ');
    return t('actionPlayed', { name: findPlayerName(players, action.playerIdx), cards });
  };

  const sortModes = [
    { mode: 0, label: t('sort.strength') },
    { mode: 1, label: t('sort.suit') },
    { mode: 2, label: t('sort.number') },
  ] as const;

  return (
    <GamePageShell
      title={tc('nav.daifugo')}
      gameThemeBg={`${bgClass} motion-safe:transition-colors motion-safe:duration-500`}
      phaseName={state.gameEndFlag ? t('phase.end') : t('phase.play')}
      isHumanTurn={isHumanTurn}
      gamePath="/daifugo"
      gameEndFlag={!!state.gameEndFlag}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {inversionToast && (
            <div
              data-testid="df-inversion-toast"
              role="status"
              aria-live="assertive"
              className="-translate-x-1/2 motion-safe:animate-pulse-once fixed top-16 left-1/2 z-30 rounded-full border border-ds-warning bg-ds-surface px-4 py-1.5 text-ds-warning text-sm font-bold shadow-lg"
            >
              {t(inversionToast === 'revolution' ? 'toast.revolution' : 'toast.elevenBack')}
            </div>
          )}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <DaifugoSettingsPanel config={configInput} onChange={handleConfigChange} />
            <ReplaySpeedSettingsPanel />
            <SettingsPanel
              title=""
              groups={[
                {
                  items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
                },
              ]}
            />
            {isMobile ? (
              <div className="flex gap-1.5 mb-2 overflow-x-auto py-3">
                {cpuPlayers.map((player) => (
                  <DaifugoCpuCompact key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
                ))}
              </div>
            ) : (
              <div className="flex gap-2.5 flex-wrap mb-2.5">
                {cpuPlayers.map((player) => (
                  <DaifugoCpuArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
                ))}
              </div>
            )}

            {/* biome-ignore lint/a11y/noStaticElementInteractions: drag-and-drop target; keyboard play uses select+button */}
            <div
              className="bg-black/30 rounded-[10px] p-2.5 my-2"
              data-tutorial="df-table-cards"
              onDragOver={(e) => e.preventDefault()}
              onDrop={handleDrop}
            >
              <div className="text-ds-text-primary font-bold mb-1.5">
                {t('tableCards')}
                {tableCount > 0 && state.lastPlayPlayerIdx >= 0 && (
                  <span className="ml-2 text-xs font-normal text-ds-text-muted" data-testid="daifugo-last-player">
                    {t('lastPlayedBy', { name: findPlayerName(state.players, state.lastPlayPlayerIdx) })}
                  </span>
                )}
              </div>
              <div className="flex flex-wrap gap-1">
                {!state.tableCards || state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted">{t('tableEmpty')}</span>
                ) : (
                  state.tableCards.map((card) => (
                    <AnimatedCard key={`${card.design}-${card.value}`} card={card} width={cardWidth} />
                  ))
                )}
              </div>
            </div>

            {pendingBanner && (
              <div
                className={`${badgeWarningColors} rounded-[10px] text-center py-2 px-4 text-sm font-bold my-2`}
                data-tutorial="df-special-actions"
              >
                {pendingBanner}
                {pendingAction === 'queenBomber' && isHumanTurn && (
                  <div className="flex flex-wrap justify-center gap-1 mt-2">
                    {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => (
                      <button
                        key={v}
                        type="button"
                        className={`${btnPrimary} min-w-[36px] text-sm`}
                        disabled={loading}
                        onClick={() => exec('play', [v])}
                      >
                        {v === 1 ? 'A' : v === 11 ? 'J' : v === 12 ? 'Q' : v === 13 ? 'K' : String(v)}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}

            <div data-tutorial="df-rules-badges">
              <DaifugoRulesBadges state={state} />
            </div>

            {state.exchangeActions && state.exchangeActions.length > 0 && (
              <DaifugoExchangeLog players={state.players} actions={state.exchangeActions} />
            )}

            {state.humanAction && (
              <div className="bg-black/40 rounded-lg text-ds-success/80 py-2 px-3.5 my-2 text-xs">
                {actionDescription(state.players, state.humanAction)}
              </div>
            )}

            {state.cpuActions && state.cpuActions.length > 0 && (
              <div className="bg-black/40 rounded-lg text-ds-text-primary py-2 px-3.5 my-2 whitespace-pre-line text-xs">
                {[tc('label.cpuActions'), ...state.cpuActions.map((a) => actionDescription(state.players, a))].join(
                  '\n',
                )}
              </div>
            )}

            <GameMessageBox
              message={
                state.gameEndFlag
                  ? `${t('resultPrefix')} ${state.players
                      .filter((p) => p.rank > 0)
                      .sort((a, b) => a.rank - b.rank)
                      .map((p) => t('resultEntry', { name: playerName(p.id, p.isHuman), rank: t(`rank.${p.rank}`) }))
                      .join(' ')}`
                  : state.message
              }
              messageCode={state.gameEndFlag ? undefined : state.messageCode}
              messageParams={state.gameEndFlag ? undefined : state.messageParams}
            />

            {/* Action log */}
            <ActionLogSection
              isEndPhase={state.gameEndFlag}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${footerClass} px-4 py-2.5 motion-safe:transition-colors motion-safe:duration-500`}>
            <fieldset className="text-center mb-1 border-0 p-0 m-0" data-tutorial="df-sort-buttons">
              <legend className="sr-only">{t('sort.label')}</legend>
              {sortModes.map(({ mode, label }) => (
                <button
                  key={mode}
                  type="button"
                  className={state.sortMode === mode ? `${btnPrimary} min-w-[70px]` : `${btnSecondary} min-w-[70px]`}
                  disabled={loading}
                  aria-pressed={state.sortMode === mode}
                  data-testid={`df-sort-${mode}`}
                  onClick={() => exec('sort', undefined, undefined, mode)}
                >
                  {label}
                </button>
              ))}
            </fieldset>

            {humanPlayer && (
              <div className="mb-2" data-tutorial="df-player-hand">
                <DaifugoHumanArea
                  player={humanPlayer}
                  selectedIndices={selectedIndices}
                  onToggle={toggleCardSelection}
                  isCurrentTurn={isHumanTurn}
                  onDragCard={handleDragCard}
                  onSwipeStart={handleDaifugoSwipeStart}
                />
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="text-center" data-tutorial="df-play-pass">
              {countMismatch && (
                <p className="mb-1.5 text-xs text-ds-error" role="alert" data-testid="daifugo-count-warning">
                  {t('countMismatch', { count: tableCount })}
                </p>
              )}
              <GameResetButton
                isGameEnd={!!state.gameEndFlag}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="df-reset-button"
                className="min-w-[90px]"
              />
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px]`}
                disabled={loading || !isHumanTurn || state.gameEndFlag || pendingAction !== 'none'}
                onClick={() => exec('play', [])}
              >
                {tc('button.pass')}
              </button>
              <button
                type="button"
                className={`${btnSuccess} min-w-[120px]`}
                disabled={
                  loading ||
                  !isHumanTurn ||
                  state.gameEndFlag ||
                  selectedIndices.length === 0 ||
                  pendingAction === 'queenBomber' ||
                  countMismatch
                }
                onClick={() =>
                  exec(
                    'play',
                    [...selectedIndices].sort((a, b) => a - b),
                  )
                }
              >
                {playButtonLabel}
              </button>
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px]`}
                disabled={loading || selectedIndices.length === 0}
                onClick={clearSelection}
                data-testid="daifugo-clear-selection"
              >
                {t('clearSelectionButton')}
              </button>
            </div>
            <CardNavShortcutsPanel data-testid="daifugo-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
