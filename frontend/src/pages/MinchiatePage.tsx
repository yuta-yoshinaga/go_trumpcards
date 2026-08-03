import { useEffect, useMemo } from 'react';
import type { minchiateApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useMinchiateGame } from '../hooks/useMinchiateGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { MINCHIATE_SURPLUS, type MinchiateResponse } from '../types/card';
import { MinchiatePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MINCHIATE_HELP, parseMinchiateCommand } from '../utils/cli/commands/minchiateCommands';
import { formatMinchiateState } from '../utils/cli/formatters/minchiateFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Minchiate tutorial step definitions. */
const MINCHIATE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="minchiate-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="minchiate-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="minchiate-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="minchiate-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MINCHIATE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MinchiatePhase.SCARTO]: 'scarto',
  [MinchiatePhase.PLAY]: 'play',
  [MinchiatePhase.TRICK_END]: 'trickEnd',
  [MinchiatePhase.ROUND_END]: 'roundEnd',
  [MinchiatePhase.GAME_END]: 'gameEnd',
};

/** Renders the Minchiate game page: a Bolognese 4-player 2v2 tarot trick-taker. */
export const MinchiatePage = withTutorial(MinchiatePageContent, 'minchiate', MINCHIATE_TUTORIAL_STEPS);

/** Inner content of the Minchiate page, wrapped by TutorialProvider. */
function MinchiatePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('minchiate');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    minchiateConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleScarto,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useMinchiateGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('minchiate');
  const cliConfig: CliGameConfig<MinchiateResponse, Parameters<typeof minchiateApi.exec>> = useMemo(
    () => ({
      gameName: 'minchiate',
      parseCommand: parseMinchiateCommand,
      formatResponse: formatMinchiateState,
      helpText: MINCHIATE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('minchiate', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('minchiate', MINCHIATE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="minchiate" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 21 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;
  const isHumanTurn = state.isHumanTurn;

  const isScartoPhase = state.phase === MinchiatePhase.SCARTO;
  const isPlayPhase = state.phase === MinchiatePhase.PLAY;
  const isTrickEnd = state.phase === MinchiatePhase.TRICK_END;
  const isRoundEnd = state.phase === MinchiatePhase.ROUND_END;
  const isGameEnd = state.phase === MinchiatePhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const canScarto = isScartoPhase && state.isHumanScarto;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.minchiate')}
      gameThemeBg={gameTheme.minchiate.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanTurn || canScarto) && !isGameEnd}
      gamePath="/minchiate"
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
                    value: minchiateConfig.cpuDifficulty,
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
                    value: minchiateConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="minchiate-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{t('target', { rounds: state.config.targetRounds })}</span>
            </div>

            {/*
              40 trumps rather than the usual 21 is the thing a player cannot
              eyeball from the hand, and "how many still outrank me" is the whole
              read — so it gets a permanent line rather than a tooltip.
            */}
            <div
              className="text-center mb-2 text-sm font-semibold text-ds-warning"
              data-testid="minchiate-trump-count-note"
            >
              {t('trumpCountNote')}
            </div>

            <div className={lgTwoColGrid}>
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="minchiate-trick-display"
                />
              </div>

              <div>
                {/* Team scores — this is a team game, not an individual one. */}
                <div
                  className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                  data-testid="minchiate-team-scores"
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
                  <div className="mb-2 text-ds-text-muted text-sm" data-testid="minchiate-scarto-done">
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
          <GameFooter className={`${gameTheme.minchiate.footer} px-4 py-2.5`}>
            {canScarto && (
              <div className="mb-1 text-center text-sm text-ds-text-muted" data-testid="minchiate-scarto-prompt">
                {t('scartoPrompt', { count: MINCHIATE_SURPLUS })}
              </div>
            )}

            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="minchiate"
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="minchiate-action-buttons">
              {canScarto && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleScarto}
                  disabled={loading || selectedCardIndices.length !== MINCHIATE_SURPLUS}
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
                dataTutorial="minchiate-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
