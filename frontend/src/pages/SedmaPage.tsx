import { useEffect, useMemo } from 'react';
import type { sedmaApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useSedmaGame } from '../hooks/useSedmaGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SedmaResponse } from '../types/card';
import { SedmaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSedmaCommand, SEDMA_HELP } from '../utils/cli/commands/sedmaCommands';
import { formatSedmaState } from '../utils/cli/formatters/sedmaFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Sedma tutorial step definitions. */
const SEDMA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sedma-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sedma-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sedma-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sedma-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sedma-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SEDMA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SedmaPhase.PLAY]: 'play',
  [SedmaPhase.TRICK_END]: 'trickEnd',
  [SedmaPhase.ROUND_END]: 'roundEnd',
  [SedmaPhase.GAME_END]: 'gameEnd',
};

/** Renders the Sedma game page: a Czech/Slovak 4-player (2 vs 2) no-trump capture trick-taker. */
export const SedmaPage = withTutorial(SedmaPageContent, 'sedma', SEDMA_TUTORIAL_STEPS);

/** Inner content of the Sedma page, wrapped by TutorialProvider. */
function SedmaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sedma');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    sedmaConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useSedmaGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sedma');
  const sedmaCliConfig: CliGameConfig<SedmaResponse, Parameters<typeof sedmaApi.exec>> = useMemo(
    () => ({
      gameName: 'sedma',
      parseCommand: parseSedmaCommand,
      formatResponse: formatSedmaState,
      helpText: SEDMA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, sedmaCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sedma', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('sedma', SEDMA_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="sedma" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isPlayPhase = state.phase === SedmaPhase.PLAY;
  const isTrickEnd = state.phase === SedmaPhase.TRICK_END;
  const isRoundEnd = state.phase === SedmaPhase.ROUND_END;
  const isGameEnd = state.phase === SedmaPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const humanTeam = humanIdx % 2;

  // One info-sidebar row per player, colour-coded by team (A = even ids = blue,
  // B = odd ids = red), matching the trick display's ally/opponent relationship.
  // Shared by the mobile (<details>) and desktop layouts.
  const renderPlayerRow = (p: (typeof state.players)[number]) => {
    const team = p.id % 2;
    const teamName = team === 0 ? t('team.a') : t('team.b');
    return (
      <div
        key={p.id}
        data-testid={`sedma-player-${p.id}`}
        data-team={team}
        className={`text-ds-text-muted text-sm py-0.5 border-l-2 pl-1 ${
          team === 0 ? 'border-ds-info' : 'border-ds-error'
        }`}
      >
        {/* Colour-independent team marker (WCAG 1.4.1): a visible A/B badge plus an
            sr-only full team name, so team membership isn't conveyed by colour alone. */}
        <span
          className={`inline-block mr-1 px-1 rounded text-xs font-bold text-white ${
            team === 0 ? 'bg-ds-info' : 'bg-ds-error'
          }`}
          aria-hidden="true"
        >
          {team === 0 ? 'A' : 'B'}
        </span>
        <span className="sr-only">{teamName}: </span>
        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
      </div>
    );
  };

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.sedma')}
      gameThemeBg={gameTheme.sedma.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/sedma"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerTeam === humanTeam}
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: sedmaConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetPoints',
                    label: t('settings.targetPoints'),
                    value: sedmaConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span>{t('trick', { n: state.trickNumber })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="sedma-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="sedma-info">
                {/* Team match scores */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] ?? 0 })}</div>
                  <div>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] ?? 0 })}</div>
                  <div className="mt-1">
                    {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
                  </div>
                </div>

                {/* Players: cards / tricks, colour-coded by 2-vs-2 team. Shared renderer for
                    both layouts so the team-colour logic lives in one place. */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">{state.players.map(renderPlayerRow)}</div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">{state.players.map(renderPlayerRow)}</div>
                )}

                {/* Live captured card points during play (A and 10 = 10 pts each, +10
                    last-trick bonus); the round-result block below takes over at round end. */}
                {!(isRoundEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="sedma-round-points"
                    role="status"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roundPointsTitle')}</div>
                    <div>{t('roundResult.teamA', { points: state.roundCardPoints[0] ?? 0 })}</div>
                    <div>{t('roundResult.teamB', { points: state.roundCardPoints[1] ?? 0 })}</div>
                  </div>
                )}

                {/* Round result */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.teamA', { points: state.roundCardPoints[0] ?? 0 })}</div>
                    <div>{t('roundResult.teamB', { points: state.roundCardPoints[1] ?? 0 })}</div>
                  </div>
                )}
              </div>
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.sedma.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="sedma"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="sedma-action-buttons">
              {canPlay && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('playButton')}
                </button>
              )}
              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sedma-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
