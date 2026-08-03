import { useEffect, useMemo } from 'react';
import type { tarocchiniApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useTarocchiniGame } from '../hooks/useTarocchiniGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { TAROCCHINI_SURPLUS, type TarocchiniResponse } from '../types/card';
import { TarocchiniPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTarocchiniCommand, TAROCCHINI_HELP } from '../utils/cli/commands/tarocchiniCommands';
import { formatTarocchiniState } from '../utils/cli/formatters/tarocchiniFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tarocchini tutorial step definitions. */
const TAROCCHINI_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tarocchini-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="tarocchini-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tarocchini-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tarocchini-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TAROCCHINI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TarocchiniPhase.SCARTO]: 'scarto',
  [TarocchiniPhase.PLAY]: 'play',
  [TarocchiniPhase.TRICK_END]: 'trickEnd',
  [TarocchiniPhase.ROUND_END]: 'roundEnd',
  [TarocchiniPhase.GAME_END]: 'gameEnd',
};

/** Renders the Tarocchini game page: a Bolognese 4-player 2v2 tarot trick-taker. */
export const TarocchiniPage = withTutorial(TarocchiniPageContent, 'tarocchini', TAROCCHINI_TUTORIAL_STEPS);

/** Inner content of the Tarocchini page, wrapped by TutorialProvider. */
function TarocchiniPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tarocchini');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    tarocchiniConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleScarto,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useTarocchiniGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tarocchini');
  const cliConfig: CliGameConfig<TarocchiniResponse, Parameters<typeof tarocchiniApi.exec>> = useMemo(
    () => ({
      gameName: 'tarocchini',
      parseCommand: parseTarocchiniCommand,
      formatResponse: formatTarocchiniState,
      helpText: TAROCCHINI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tarocchini', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('tarocchini', TAROCCHINI_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="tarocchini" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 15 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;
  const isHumanTurn = state.isHumanTurn;

  const isScartoPhase = state.phase === TarocchiniPhase.SCARTO;
  const isPlayPhase = state.phase === TarocchiniPhase.PLAY;
  const isTrickEnd = state.phase === TarocchiniPhase.TRICK_END;
  const isRoundEnd = state.phase === TarocchiniPhase.ROUND_END;
  const isGameEnd = state.phase === TarocchiniPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const canScarto = isScartoPhase && state.isHumanScarto;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.tarocchini')}
      gameThemeBg={gameTheme.tarocchini.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanTurn || canScarto) && !isGameEnd}
      gamePath="/tarocchini"
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
                    value: tarocchiniConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: tarocchiniConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="tarocchini-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{t('target', { rounds: state.config.targetRounds })}</span>
            </div>

            {/*
              The papi rule is the one thing a player cannot infer from the cards,
              so it gets a permanent line rather than a tooltip: four trumps rank
              equal, and the LATER-played one takes the trick.
            */}
            <div className="text-center mb-2 text-sm font-semibold text-ds-warning" data-testid="tarocchini-papi-note">
              {t('papiNote')}
            </div>

            <div className={lgTwoColGrid}>
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="tarocchini-trick-display"
                />
              </div>

              <div>
                {/* Team scores — this is a team game, not an individual one. */}
                <div
                  className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                  data-testid="tarocchini-team-scores"
                >
                  {state.teamScores.map((score, team) => (
                    <div
                      key={team}
                      className={`py-0.5 ${team === humanTeam ? 'text-ds-text-primary font-semibold' : ''}`}
                    >
                      {t('teamScore', { n: team, score })}
                    </div>
                  ))}
                </div>

                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}
                          {p.isDealer ? ` / ${t('dealerBadge')}` : ''}): {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}
                        {p.isDealer ? ` / ${t('dealerBadge')}` : ''}): {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {state.scartoCount > 0 && (
                  <div className="mb-2 text-ds-text-muted text-sm" data-testid="tarocchini-scarto-done">
                    {t('scartoDone', { count: state.scartoCount })}
                  </div>
                )}

                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {state.players.map((p) => (
                      <div key={p.id}>
                        {t('roundResult.tricks', {
                          name: playerName(p.id, p.isHuman),
                          count: state.roundTricks[p.id] ?? 0,
                        })}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

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
          <GameFooter className={`${gameTheme.tarocchini.footer} px-4 py-2.5`}>
            {canScarto && (
              <div className="mb-1 text-center text-sm text-ds-text-muted" data-testid="tarocchini-scarto-prompt">
                {t('scartoPrompt', { count: TAROCCHINI_SURPLUS })}
              </div>
            )}

            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="tarocchini"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && isRequestedHint(state) && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="tarocchini-action-buttons">
              {canScarto && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleScarto}
                  disabled={loading || selectedCardIndices.length !== TAROCCHINI_SURPLUS}
                >
                  {t('scartoButton')}
                </button>
              )}
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
                dataTutorial="tarocchini-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
