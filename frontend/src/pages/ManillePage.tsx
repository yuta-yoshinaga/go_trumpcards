import { useEffect, useMemo } from 'react';
import type { manilleApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useManilleGame } from '../hooks/useManilleGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ManilleResponse } from '../types/card';
import { ManillePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MANILLE_HELP, parseManilleCommand } from '../utils/cli/commands/manilleCommands';
import { formatManilleState } from '../utils/cli/formatters/manilleFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Manille tutorial step definitions. */
const MANILLE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="manille-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="manille-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="manille-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="manille-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="manille-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MANILLE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ManillePhase.PLAY]: 'play',
  [ManillePhase.TRICK_END]: 'trickEnd',
  [ManillePhase.ROUND_END]: 'roundEnd',
  [ManillePhase.GAME_END]: 'gameEnd',
};

/** Renders the Manille game page: a French/Belgian 4-player (2 vs 2) trump trick-taker. */
export const ManillePage = withTutorial(ManillePageContent, 'manille', MANILLE_TUTORIAL_STEPS);

/** Inner content of the Manille page, wrapped by TutorialProvider. */
function ManillePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('manille');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    manilleConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useManilleGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('manille');
  const manilleCliConfig: CliGameConfig<ManilleResponse, Parameters<typeof manilleApi.exec>> = useMemo(
    () => ({
      gameName: 'manille',
      parseCommand: parseManilleCommand,
      formatResponse: formatManilleState,
      helpText: MANILLE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, manilleCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('manille', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('manille', MANILLE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="manille" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isPlayPhase = state.phase === ManillePhase.PLAY;
  const isTrickEnd = state.phase === ManillePhase.TRICK_END;
  const isRoundEnd = state.phase === ManillePhase.ROUND_END;
  const isGameEnd = state.phase === ManillePhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const humanTeam = humanIdx % 2;
  const trumpSymbol = SUIT_SYMBOLS[state.trumpSuit] ?? '?';

  // One info-sidebar row per player; same-team members are emphasised. Shared by the
  // mobile (<details>) and desktop layouts so the highlight logic lives in one place.
  const renderPlayerRow = (p: (typeof state.players)[number]) => {
    const isMyTeam = p.id % 2 === humanTeam;
    return (
      <div
        key={p.id}
        data-testid={`manille-player-${p.id}`}
        data-own-team={isMyTeam ? 'true' : undefined}
        className={`text-sm py-0.5 ${
          isMyTeam ? 'font-bold text-ds-text-primary border-l-2 border-ds-accent pl-1' : 'text-ds-text-muted'
        }`}
      >
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
      title={tc('nav.manille')}
      gameThemeBg={gameTheme.manille.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/manille"
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
                    value: manilleConfig.cpuDifficulty,
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
                    value: manilleConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
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
              <span>{t('trump', { suit: trumpSymbol })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="manille-trick-display"
                />
                {/* At TrickEnd leadPlayerIdx is the trick winner (they lead next), so show who took it. */}
                {isTrickEnd &&
                  (() => {
                    const winnerIdx = state.leadPlayerIdx;
                    const isMyTeam = winnerIdx % 2 === humanTeam;
                    return (
                      <div
                        className={`my-2 p-2 rounded text-center text-sm font-semibold ${
                          isMyTeam ? 'bg-ds-accent/15 text-ds-accent' : 'bg-black/30 text-ds-text-muted'
                        }`}
                        role="status"
                        aria-live="polite"
                        data-testid="manille-trick-winner"
                      >
                        {t('trickWinner', {
                          name: playerName(winnerIdx, state.players[winnerIdx]?.isHuman ?? false),
                          team: winnerIdx % 2 === 0 ? t('team.a') : t('team.b'),
                        })}
                      </div>
                    );
                  })()}
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="manille-info">
                {/* Team match scores */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] ?? 0 })}</div>
                  <div>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] ?? 0 })}</div>
                  <div className="mt-1">
                    {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
                  </div>
                </div>

                {/* Players: cards / tricks. Same-team members (p.id % 2 === humanTeam) are
                    emphasised so the 2-vs-2 partnership is visible. Shared by both layouts. */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">{state.players.map(renderPlayerRow)}</div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">{state.players.map(renderPlayerRow)}</div>
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
          <GameFooter className={`${gameTheme.manille.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="manille"
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="manille-action-buttons">
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
                dataTutorial="manille-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
