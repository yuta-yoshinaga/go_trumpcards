import { useEffect, useMemo } from 'react';
import type { suecaApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_GAME_POINTS_OPTIONS, useSuecaGame } from '../hooks/useSuecaGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SuecaResponse } from '../types/card';
import { SuecaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSuecaCommand, SUECA_HELP } from '../utils/cli/commands/suecaCommands';
import { formatSuecaState } from '../utils/cli/formatters/suecaFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Card points needed to win a Sueca round (majority of the 120 total). */
const SUECA_WIN_POINTS = 61;

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;
/** Suit id → `suitName.*` i18n key (1=♠ .. 4=♦). */
const SUIT_KEYS = ['', 'spade', 'club', 'heart', 'diamond'] as const;

/** Sueca tutorial step definitions. */
const SUECA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sueca-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sueca-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sueca-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sueca-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sueca-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SUECA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SuecaPhase.PLAY]: 'play',
  [SuecaPhase.TRICK_END]: 'trickEnd',
  [SuecaPhase.ROUND_END]: 'roundEnd',
  [SuecaPhase.GAME_END]: 'gameEnd',
};

/** Renders the Sueca game page: a Portuguese 4-player (2 vs 2) trump trick-taker. */
export const SuecaPage = withTutorial(SuecaPageContent, 'sueca', SUECA_TUTORIAL_STEPS);

/** Inner content of the Sueca page, wrapped by TutorialProvider. */
function SuecaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sueca');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    suecaConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useSuecaGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sueca');
  const suecaCliConfig: CliGameConfig<SuecaResponse, Parameters<typeof suecaApi.exec>> = useMemo(
    () => ({
      gameName: 'sueca',
      parseCommand: parseSuecaCommand,
      formatResponse: formatSuecaState,
      helpText: SUECA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, suecaCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sueca', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('sueca', SUECA_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="sueca" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 6 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;

  const isPlayPhase = state.phase === SuecaPhase.PLAY;
  const isTrickEnd = state.phase === SuecaPhase.TRICK_END;
  const isRoundEnd = state.phase === SuecaPhase.ROUND_END;
  const isGameEnd = state.phase === SuecaPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const humanTeam = humanIdx % 2;
  const trumpSymbol = SUIT_SYMBOLS[state.trumpSuit] ?? '?';
  // Spoken trump: the suit name (not the ♠♣♥♦ glyph, which SRs read poorly).
  const trumpSuitName = SUIT_KEYS[state.trumpSuit] ? t(`suitName.${SUIT_KEYS[state.trumpSuit]}`) : trumpSymbol;
  const trumpAriaLabel = t('trump', { suit: trumpSuitName });

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.sueca')}
      gameThemeBg={gameTheme.sueca.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/sueca"
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
                    value: suecaConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetGamePoints',
                    label: t('settings.targetGamePoints'),
                    value: suecaConfig.targetGamePoints,
                    options: TARGET_GAME_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetGamePoints', v),
                  },
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

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span role="img" aria-label={trumpAriaLabel}>
                <span aria-hidden="true">{t('trump', { suit: trumpSymbol })}</span>
              </span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="sueca-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="sueca-info">
                {/* Team game points */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {/* Trump kept in the always-visible sidebar so it isn't lost off-screen on mobile. */}
                  <div
                    className="mb-1 text-ds-text-primary font-semibold"
                    data-testid="sueca-sidebar-trump"
                    role="img"
                    aria-label={trumpAriaLabel}
                  >
                    <span aria-hidden="true">{t('trump', { suit: trumpSymbol })}</span>
                  </div>
                  <div>{t('teamScore', { team: t('team.a'), score: state.teamGamePoints[0] ?? 0 })}</div>
                  <div>{t('teamScore', { team: t('team.b'), score: state.teamGamePoints[1] ?? 0 })}</div>
                  <div className="mt-1">
                    {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
                  </div>
                </div>

                {/* Players: cards / tricks */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Live captured card points during play; the round-result block
                    below takes over once the round ends. Progress is shown against
                    the 61-point round-win line. */}
                {!(isRoundEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    role="status"
                    aria-live="polite"
                    data-testid="sueca-round-points"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('livePoints.title')}</div>
                    <div>
                      {t('livePoints.team', {
                        team: t('team.a'),
                        points: state.roundCardPoints[0] ?? 0,
                        target: SUECA_WIN_POINTS,
                      })}
                    </div>
                    <div>
                      {t('livePoints.team', {
                        team: t('team.b'),
                        points: state.roundCardPoints[1] ?? 0,
                        target: SUECA_WIN_POINTS,
                      })}
                    </div>
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

            {/* At TrickEnd leadPlayerIdx is the trick winner, so their team shows before Next Trick. */}
            {isTrickEnd && (
              <div
                className="my-2 p-2 rounded bg-ds-accent/15 text-center text-sm font-semibold text-ds-accent"
                role="status"
                aria-live="polite"
                data-testid="sueca-trick-winner"
              >
                {t('trickWinner', { team: state.leadPlayerIdx % 2 === 0 ? t('team.a') : t('team.b') })}
              </div>
            )}

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
          <GameFooter className={`${gameTheme.sueca.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="sueca"
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
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="sueca-action-buttons">
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
                dataTutorial="sueca-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
