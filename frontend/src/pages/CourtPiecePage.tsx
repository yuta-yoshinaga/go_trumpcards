import { useEffect, useMemo } from 'react';
import type { courtPieceApi } from '../api/gameApi';
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
import {
  CPU_DIFFICULTY_OPTIONS,
  POINT_LIMIT_OPTIONS,
  TRUMP_SUIT_OPTIONS,
  useCourtPieceGame,
} from '../hooks/useCourtPieceGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CourtPieceResponse } from '../types/card';
import { CourtPiecePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { COURT_PIECE_HELP, parseCourtPieceCommand } from '../utils/cli/commands/courtPieceCommands';
import { formatCourtPieceState } from '../utils/cli/formatters/courtPieceFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { courtPieceLegalPlayIndices } from '../utils/courtPieceLegal';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 = undeclared). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Tricks a team must take within a 13-trick round to win it (Sar); mirrors CourtPieceTricksToWin in internal/domain/CourtPiece.go. */
const COURT_PIECE_TRICKS_TO_WIN = 7;

/** Court Piece (Rang) tutorial step definitions. */
const COURT_PIECE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="courtpiece-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="courtpiece-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="courtpiece-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="courtpiece-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const COURT_PIECE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CourtPiecePhase.TRUMP_DECLARATION]: 'trumpDeclaration',
  [CourtPiecePhase.PLAY]: 'play',
  [CourtPiecePhase.TRICK_END]: 'trickEnd',
  [CourtPiecePhase.ROUND_END]: 'roundEnd',
  [CourtPiecePhase.GAME_END]: 'gameEnd',
};

/** Renders the Court Piece (Rang) game page: a South-Asian 4-player (2 vs 2) called-trump trick-taker. */
export const CourtPiecePage = withTutorial(CourtPiecePageContent, 'courtpiece', COURT_PIECE_TUTORIAL_STEPS);

/** Inner content of the Court Piece (Rang) page, wrapped by TutorialProvider. */
function CourtPiecePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('courtpiece');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    courtPieceConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleDeclareTrump,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useCourtPieceGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('courtpiece');
  const cliConfig: CliGameConfig<CourtPieceResponse, Parameters<typeof courtPieceApi.exec>> = useMemo(
    () => ({
      gameName: 'courtpiece',
      parseCommand: parseCourtPieceCommand,
      formatResponse: formatCourtPieceState,
      helpText: COURT_PIECE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('courtpiece', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('courtpiece', COURT_PIECE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="courtpiece" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const humanTeam = humanIdx % 2;

  const isTrumpPhase = state.phase === CourtPiecePhase.TRUMP_DECLARATION;
  const isPlayPhase = state.phase === CourtPiecePhase.PLAY;
  const isTrickEnd = state.phase === CourtPiecePhase.TRICK_END;
  const isRoundEnd = state.phase === CourtPiecePhase.ROUND_END;
  const isGameEnd = state.phase === CourtPiecePhase.GAME_END || state.gameEndFlag;

  // The web contract carries no explicit turn flags, so derive them from the
  // current seat: it is the human's turn whenever currentPlayerIdx is the human.
  const isHumanCurrent = state.currentPlayerIdx === humanIdx;
  const isHumanTrumpTurn = isTrumpPhase && state.callerIdx === humanIdx;
  const isHumanTurn = isPlayPhase && isHumanCurrent;
  const canPlay = isHumanTurn;

  // On the human's play turn, mirror the server's follow-suit rule
  // (internal/domain/CourtPiece.go validatePlay) so legal cards get a success
  // ring. This is an additive hint only — illegal cards stay clickable and the
  // backend remains the source of truth for rejecting an illegal play.
  const legalPlayIndices =
    canPlay && humanPlayer ? courtPieceLegalPlayIndices(humanPlayer.cards, state.currentTrick) : undefined;

  const trumpSymbol = state.trumpSuit === 0 ? t('noTrump') : (SUIT_SYMBOLS[state.trumpSuit] ?? '?');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.courtpiece')}
      gameThemeBg={gameTheme.courtpiece.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanTurn || isHumanTrumpTurn) && !isGameEnd}
      gamePath="/courtpiece"
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
                    value: courtPieceConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'pointLimit',
                    label: t('settings.pointLimit'),
                    value: courtPieceConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="courtpiece-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span>{t('target', { points: state.config.pointLimit })}</span>
            </div>

            <div className="text-ds-text-muted text-center mb-2 text-sm">
              {state.callerIdx >= 0
                ? t('callerLine', {
                    name: playerName(state.callerIdx, state.players[state.callerIdx]?.isHuman ?? false),
                    team: state.callerIdx % 2 === 0 ? t('team.a') : t('team.b'),
                  })
                : t('callerUndecided')}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="courtpiece-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* Team match scores */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] ?? 0 })}</div>
                  <div>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] ?? 0 })}</div>
                  <div className="mt-1">
                    {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
                  </div>
                  {/* Live running trick tally toward the 7-trick round target (during play). */}
                  {(isPlayPhase || isTrickEnd) && (
                    <div className="mt-1 pt-1 border-t border-white/10" data-testid="cp-live-tricks">
                      <span className="text-ds-text-primary">{t('liveTricks.title')}: </span>
                      {([0, 1] as const).map((team, i) => {
                        const tricks = teamTricks(team);
                        const reached = tricks >= COURT_PIECE_TRICKS_TO_WIN;
                        return (
                          <span key={team}>
                            {i > 0 && ' · '}
                            <span
                              className={reached ? 'text-ds-accent font-semibold' : ''}
                              data-testid={`cp-live-tricks-team-${team}`}
                            >
                              {t('liveTricks.team', {
                                team: team === 0 ? t('team.a') : t('team.b'),
                                tricks,
                                target: COURT_PIECE_TRICKS_TO_WIN,
                              })}
                            </span>
                          </span>
                        );
                      })}
                    </div>
                  )}
                </div>

                {/* Players grouped by team, with the caller badge */}
                {([0, 1] as const).map((team) => (
                  <div key={team} className="mb-2 p-2 rounded bg-black/30">
                    <div className="text-ds-text-primary text-xs mb-1">{team === 0 ? t('team.a') : t('team.b')}</div>
                    {isMobile ? (
                      <details>
                        <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                          {t('players')}
                        </summary>
                        <div className="mt-1">{renderTeamPlayers(team)}</div>
                      </details>
                    ) : (
                      renderTeamPlayers(team)
                    )}
                  </div>
                ))}

                {/* Round result: tricks per team */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.teamA', { tricks: teamTricks(0) })}</div>
                    <div>{t('roundResult.teamB', { tricks: teamTricks(1) })}</div>
                    {state.lastRoundCourt && <div className="mt-1 text-ds-warning">{t('roundResult.court')}</div>}
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
          <GameFooter className={`${gameTheme.courtpiece.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="courtpiece"
                legalIndices={legalPlayIndices}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndex != null && ` ([${state.hint.cardIndex}])`}
                {state.hint.trumpSuit != null && ` (${SUIT_SYMBOLS[state.hint.trumpSuit] ?? '?'})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="courtpiece-action-buttons">
              {isTrumpPhase && isHumanTrumpTurn && (
                <>
                  <span className="text-xs font-bold text-ds-warning self-center mr-1" data-testid="cp-trump-status">
                    {t('trumpUndeclared')}
                  </span>
                  <span className="text-xs text-ds-text-muted self-center mr-1">{t('trumpPrompt')}</span>
                  {TRUMP_SUIT_OPTIONS.map((s) => (
                    <button
                      key={s.value}
                      type="button"
                      className="px-3 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                      onClick={() => handleDeclareTrump(s.value)}
                      disabled={loading}
                      aria-label={t('declareTrumpAria', { suit: t(s.key) })}
                      data-testid={`trump-${s.value}`}
                    >
                      {t(s.key)}
                    </button>
                  ))}
                </>
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
                dataTutorial="courtpiece-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );

  /** Sums the tricks won by all seats on one team (team 0 = seats 0&2, team 1 = seats 1&3). */
  function teamTricks(team: number): number {
    if (!state) return 0;
    return state.players.filter((p) => p.team === team).reduce((sum, p) => sum + p.trickCount, 0);
  }

  /** Renders the per-player rows for one team (seats with that team index). */
  function renderTeamPlayers(team: number) {
    if (!state) return null;
    return state.players
      .filter((p) => p.team === team)
      .map((p) => (
        <div key={p.id} className="py-0.5 flex items-center gap-2 text-ds-text-muted text-sm">
          <span className={p.id === state.callerIdx ? 'text-ds-warning font-semibold' : ''}>
            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
          </span>
          {p.id === state.callerIdx && (
            <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">{t('callerBadge')}</span>
          )}
        </div>
      ));
  }
}
