import { useEffect, useMemo } from 'react';
import type { tablanetApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
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
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, useTablanetGame } from '../hooks/useTablanetGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TablanetResponse } from '../types/card';
import { TablanetPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTablanetCommand, TABLANET_HELP } from '../utils/cli/commands/tablanetCommands';
import { formatTablanetState } from '../utils/cli/formatters/tablanetFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Tablanet (Tablić) tutorial step definitions. */
const TABLANET_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tablanet-table-cards"]',
    messageKey: 'tutorial.table',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tablanet-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tablanet-actions"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="tablanet-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="tablanet-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Tablanet (Tablić) game page: a 4-player 52-card fishing/capture game. */
export const TablanetPage = withTutorial(TablanetPageContent, 'tablanet', TABLANET_TUTORIAL_STEPS);

/** Inner content of the Tablanet page, wrapped by TutorialWrapper. */
function TablanetPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tablanet');
  const {
    state,
    loading,
    error,
    callApi,
    retry,
    handIndex,
    setHandIndex,
    tableIndices,
    toggleTable,
    configInput,
    handleConfigChange,
    playCard,
    handleNextGame,
    handleResetWithConfig,
  } = useTablanetGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tablanet', state);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    callApi('reset');
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tablanet');
  const cliConfig: CliGameConfig<TablanetResponse, Parameters<typeof tablanetApi.exec>> = useMemo(
    () => ({
      gameName: 'tablanet',
      parseCommand: parseTablanetCommand,
      formatResponse: formatTablanetState,
      helpText: TABLANET_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.tablanet.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const human = state.players.find((p) => p.isHuman) ?? state.players[0] ?? null;
  const isGameEnd = state.phase === TablanetPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn && !isGameEnd;
  const humanWon = isGameEnd && state.winners.includes(0);
  const phaseName = isGameEnd ? t('phase.gameEnd') : t('phase.play');

  // Table indices the currently-selected hand card can capture (backend hint).
  const captureCandidates =
    handIndex !== null && isHumanTurn ? new Set(state.captureOptions[handIndex] ?? []) : new Set<number>();
  const canPlay = isHumanTurn && handIndex !== null;

  const winnerNames = state.winners.map((i) => (state.players[i]?.isHuman ? t('you') : t('cpu', { id: i }))).join(', ');

  const humanStats = human
    ? `${t('cards', { count: human.cardCount })} · ${t('captured', { count: human.capturedCount })} · ${t('tabla', { count: human.tablaCount })}`
    : '';

  return (
    <GamePageShell
      title={tc('nav.tablanet')}
      gameThemeBg={gameTheme.tablanet.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/tablanet"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
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
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            <div className="text-center text-xs text-ds-text-muted" data-tutorial="tablanet-info">
              <span className="mr-3">{t('deal', { n: state.roundNumber })}</span>
              <span>{t('deck', { count: state.remainingDeck })}</span>
            </div>

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="tablanet-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1">
                      {t('cpu', { id: p.id })} — {t('captured', { count: p.capturedCount })} ·{' '}
                      {t('tabla', { count: p.tablaCount })}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.cardCount, 8) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.45} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Table cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="tablanet-table-cards">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('table')}</div>
              <div className="flex justify-center gap-2 min-h-[60px] flex-wrap">
                {state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('tableEmpty')}</span>
                ) : (
                  state.tableCards.map((c, i) => {
                    const isCandidate = captureCandidates.has(i);
                    return (
                      <button
                        key={i}
                        type="button"
                        onClick={() => isHumanTurn && toggleTable(i)}
                        disabled={!isHumanTurn}
                        className={`rounded transition-all ${
                          tableIndices.includes(i)
                            ? 'ring-2 ring-ds-warning -translate-y-1'
                            : isCandidate
                              ? 'ring-2 ring-ds-success motion-safe:animate-pulse'
                              : ''
                        } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                        data-testid={`table-card-${i}`}
                        data-capture-candidate={isCandidate || undefined}
                      >
                        <AnimatedCard card={c} width={cardWidth * 0.9} />
                      </button>
                    );
                  })
                )}
              </div>
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="tablanet-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {t('you')} — {humanStats}
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human?.cards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    onClick={() => isHumanTurn && setHandIndex(handIndex === i ? null : i)}
                    disabled={!isHumanTurn}
                    className={`rounded transition-all ${
                      handIndex === i ? 'ring-2 ring-ds-info -translate-y-2' : ''
                    } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    data-testid={`hand-card-${i}`}
                  >
                    <AnimatedCard card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            {/* Turn prompt */}
            {!isGameEnd && (
              <div className="text-center text-sm text-ds-accent" data-testid="tablanet-prompt">
                {isHumanTurn ? t('turnYours') : t('turnCpu', { id: state.currentTurn })}
              </div>
            )}

            {/* Game-end result */}
            {isGameEnd && (
              <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="tablanet-result">
                <div className="mb-1 text-ds-text-primary">{t('result.title')}</div>
                {state.winners.length > 0 && (
                  <div className="text-ds-success mb-1">{t('result.winner', { names: winnerNames })}</div>
                )}
                {state.players.map((p) => (
                  <div key={p.id}>
                    {t('result.score', {
                      name: p.isHuman ? t('you') : t('cpu', { id: p.id }),
                      score: p.score,
                    })}
                  </div>
                ))}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: String(configInput.cpuDifficulty ?? 1),
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: String(o.value),
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.tablanet.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="tablanet-actions">
              {!isGameEnd && isHumanTurn && (
                <button type="button" className={btnPrimary} onClick={playCard} disabled={loading || !canPlay}>
                  {tableIndices.length > 0 ? t('captureButton') : t('playButton')}
                </button>
              )}
              {isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextGame} disabled={loading}>
                  {t('newGame')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleResetWithConfig}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="tablanet-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
