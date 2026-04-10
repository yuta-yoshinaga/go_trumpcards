import { useCallback, useMemo } from 'react';
import type { daifugoApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { DaifugoCpuArea } from '../components/daifugo/DaifugoCpuArea';
import { DaifugoExchangeLog } from '../components/daifugo/DaifugoExchangeLog';
import { DaifugoHumanArea } from '../components/daifugo/DaifugoHumanArea';
import { DaifugoRulesBadges } from '../components/daifugo/DaifugoRulesBadges';
import { DaifugoSettingsPanel } from '../components/daifugo/DaifugoSettingsPanel';
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
import { DaifugoSkeleton } from '../components/skeleton/DaifugoSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCardSwipeSelection } from '../hooks/useCardSwipeSelection';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useDaifugoGame } from '../hooks/useDaifugoGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DaifugoAction, DaifugoResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardLabel } from '../utils/cardUtils';
import { DAIFUGO_HELP, parseDaifugoCommand } from '../utils/cli/commands/daifugoCommands';
import { formatDaifugoState } from '../utils/cli/formatters/daifugoFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName, playerName } from '../utils/playerUtils';

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
export function DaifugoPage() {
  return (
    <TutorialWrapper gameName="daifugo" steps={DF_TUTORIAL_STEPS}>
      <DaifugoPageContent />
    </TutorialWrapper>
  );
}

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
  const { playSound } = useSound();

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

  if (!state) return <DaifugoSkeleton />;

  const pendingAction = state.pendingAction ?? 'none';
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const humanPlayer = state.players.find((p) => p.isHuman);

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
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.daifugo.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.daifugo')} />
      <PhaseIndicator phaseName={state.gameEndFlag ? t('phase.end') : t('phase.play')} isHumanTurn={isHumanTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/daifugo" />
      </PhaseIndicator>
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <DaifugoSettingsPanel config={configInput} onChange={handleConfigChange} />
            <SettingsPanel
              title=""
              groups={[
                {
                  items: [
                    {
                      type: 'checkbox',
                      id: 'frontendHint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: frontendHintEnabled,
                      onToggle: setFrontendHintEnabled,
                    },
                  ],
                },
              ]}
            />
            <div className="flex gap-2.5 flex-wrap mb-2.5">
              {cpuPlayers.map((player) => (
                <DaifugoCpuArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
              ))}
            </div>

            {/* biome-ignore lint/a11y/noStaticElementInteractions: drag-and-drop target; keyboard play uses select+button */}
            <div
              className="bg-black/30 rounded-[10px] p-2.5 my-2"
              data-tutorial="df-table-cards"
              onDragOver={(e) => e.preventDefault()}
              onDrop={handleDrop}
            >
              <div className="text-white font-bold mb-1.5">{t('tableCards')}</div>
              <div className="flex flex-wrap gap-1">
                {!state.tableCards || state.tableCards.length === 0 ? (
                  <span className="text-gray-400">{t('tableEmpty')}</span>
                ) : (
                  state.tableCards.map((card) => (
                    <AnimatedCard
                      key={`${card.design}-${card.value}`}
                      card={card}
                      width={cardWidth}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                  ))
                )}
              </div>
            </div>

            {pendingBanner && (
              <div
                className="bg-yellow-700/80 rounded-[10px] text-white text-center py-2 px-4 text-sm font-bold my-2"
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
              <div className="bg-black/40 rounded-lg text-green-200 py-2 px-3.5 my-2 text-xs">
                {actionDescription(state.players, state.humanAction)}
              </div>
            )}

            {state.cpuActions && state.cpuActions.length > 0 && (
              <div className="bg-black/40 rounded-lg text-white py-2 px-3.5 my-2 whitespace-pre-line text-xs">
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

          <GameFooter className={`${gameTheme.daifugo.footer} px-4 py-2.5`}>
            <div className="text-center mb-1" data-tutorial="df-sort-buttons">
              {sortModes.map(({ mode, label }) => (
                <button
                  key={mode}
                  type="button"
                  className={state.sortMode === mode ? `${btnPrimary} min-w-[70px]` : `${btnSecondary} min-w-[70px]`}
                  disabled={loading}
                  onClick={() => exec('sort', undefined, undefined, mode)}
                >
                  {label}
                </button>
              ))}
            </div>

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

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="text-center" data-tutorial="df-play-pass">
              <button
                type="button"
                className={`${btnOutline} min-w-[90px]`}
                data-tutorial="df-reset-button"
                disabled={loading}
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    exec('reset', [], configInput);
                  })
                }
              >
                {tc('button.reset')}
              </button>
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
                  pendingAction === 'queenBomber'
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
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration show={!!state?.gameEndFlag} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
