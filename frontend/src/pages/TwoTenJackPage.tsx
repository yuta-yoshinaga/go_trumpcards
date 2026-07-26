import { useCallback, useMemo } from 'react';
import type { twoTenJackApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardNavShortcutsPanel } from '../components/CardNavShortcutsPanel';
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
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useTwoTenJackGame } from '../hooks/useTwoTenJackGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TwoTenJackResponse } from '../types/card';
import { TwoTenJackPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTwoTenJackCommand, TWOTENJACK_HELP } from '../utils/cli/commands/twoTenJackCommands';
import { formatTwoTenJackState } from '../utils/cli/formatters/twoTenJackFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tutorial steps for Two Ten Jack. Walks the player through declare, play, and scoring elements. */
const TTJ_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tt-declare-controls"]',
    messageKey: 'tutorial.declareControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TWOTENJACK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TwoTenJackPhase.DECLARE]: 'declare',
  [TwoTenJackPhase.PLAY]: 'play',
  [TwoTenJackPhase.TRICK_END]: 'trickEnd',
  [TwoTenJackPhase.ROUND_END]: 'roundEnd',
  [TwoTenJackPhase.GAME_END]: 'gameEnd',
};

/** Suit options shown in the declare phase picker. Values match the domain CardDesign constants. */
const SUIT_OPTIONS: Readonly<{ value: number; symbol: string; key: string }[]> = [
  { value: 1, symbol: '\u2660', key: 'spade' },
  { value: 2, symbol: '\u2663', key: 'club' },
  { value: 3, symbol: '\u2665', key: 'heart' },
  { value: 4, symbol: '\u2666', key: 'diamond' },
];

/** Returns the human-readable suit symbol for a declared trump suit value. */
function trumpSymbol(trumpSuit: number): string {
  return SUIT_OPTIONS.find((s) => s.value === trumpSuit)?.symbol ?? '-';
}

/** Renders the Two Ten Jack game page: declare trump, trick play, and team scoring. */
export const TwoTenJackPage = withTutorial(TwoTenJackPageContent, 'twotenjack', TTJ_TUTORIAL_STEPS);
/** Inner content of the Two Ten Jack page, wrapped by TutorialProvider. */
function TwoTenJackPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('twotenjack');
  const {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
    handleHint,
    exec: dispatch,
    retry,
    twoTenJackConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDeclare,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useTwoTenJackGame();
  const { cardWidth, isMobile } = useCardDimensions();

  // CLI mode (stub: parseCommand returns an error, help is empty).
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('twotenjack');
  const cliConfig: CliGameConfig<TwoTenJackResponse, Parameters<typeof twoTenJackApi.exec>> = useMemo(
    () => ({
      gameName: 'twotenjack',
      parseCommand: parseTwoTenJackCommand,
      formatResponse: formatTwoTenJackState,
      helpText: TWOTENJACK_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('twotenjack', state);

  const isPlayPhaseForKbd = state?.phase === TwoTenJackPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('twotenjack', TWOTENJACK_PHASE_KEYS);
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset', undefined, undefined, {
      cpuDifficulty: twoTenJackConfig.cpuDifficulty,
      pointLimit: twoTenJackConfig.pointLimit,
    });
  }, [dispatch, hideActionLog, twoTenJackConfig.cpuDifficulty, twoTenJackConfig.pointLimit]);

  if (!state)
    return <GameSkeleton gameKey="twotenjack" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDeclarePhase = state.phase === TwoTenJackPhase.DECLARE;
  const isPlayPhase = state.phase === TwoTenJackPhase.PLAY;
  const isTrickEnd = state.phase === TwoTenJackPhase.TRICK_END;
  const isRoundEnd = state.phase === TwoTenJackPhase.ROUND_END;
  const isGameEnd = state.phase === TwoTenJackPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanDeclarer = isDeclarePhase && state.players[state.declarerIdx]?.isHuman === true;

  // The human sits at seat 0, so team 0 is the human team.
  const humanWon = isGameEnd && state.winnerTeam === 0;

  // Team totals (team 0 = seats 0,2 ; team 1 = seats 1,3).
  const team0Total = (state.players[0]?.cumulativeScore ?? 0) + (state.players[2]?.cumulativeScore ?? 0);
  const team1Total = (state.players[1]?.cumulativeScore ?? 0) + (state.players[3]?.cumulativeScore ?? 0);
  const team0Captured = (state.players[0]?.capturedPoints ?? 0) + (state.players[2]?.capturedPoints ?? 0);
  const team1Captured = (state.players[1]?.capturedPoints ?? 0) + (state.players[3]?.capturedPoints ?? 0);

  return (
    <GamePageShell
      title={tc('nav.twotenjack')}
      gameThemeBg={gameTheme.twotenjack.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanDeclarer || isHumanTurn}
      gamePath="/twotenjack"
      gameEndFlag={!!state?.gameEndFlag}
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: twoTenJackConfig.cpuDifficulty,
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
                    value: twoTenJackConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>
                {state.trumpSuit > 0 ? `${t('trump')}: ${trumpSymbol(state.trumpSuit)}` : t('trumpUndeclared')}
              </span>
            </div>

            <div className={lgTwoColGrid}>
              <div>
                {isHumanDeclarer && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="tt-declare-controls">
                    {t('declarePhase')}
                  </div>
                )}

                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="tt-trick-display"
                />
              </div>

              <div>
                {/* CPU players */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                      {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
                    </summary>
                    <div className="mt-1">
                      {state.players
                        .filter((p) => !p.isHuman)
                        .map((p) => (
                          <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                            {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                            {t('tricks', { count: p.trickCount })} | {t('capturedPoints', { count: p.capturedPoints })}
                          </div>
                        ))}
                    </div>
                  </details>
                ) : (
                  state.players
                    .filter((p) => !p.isHuman)
                    .map((p) => (
                      <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                        <div className="text-ds-text-muted text-sm">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('cumulativeScore', { score: p.cumulativeScore })} | {t('tricks', { count: p.trickCount })}{' '}
                          | {t('capturedPoints', { count: p.capturedPoints })}
                        </div>
                      </div>
                    ))
                )}

                {/* Score table */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30 relative"
                    data-tutorial="tt-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[320px] mt-1">
                        <caption className="sr-only">{t('scoresCaption')}</caption>
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoresTeam')}
                            </th>
                            <th scope="col">{t('scoresCaptured')}</th>
                            <th scope="col">{t('scoresTotal')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr className="text-ds-accent">
                            <td>{t('team0')}</td>
                            <td className="text-center">{team0Captured}</td>
                            <td className="text-center">{team0Total}</td>
                          </tr>
                          <tr>
                            <td>{t('team1')}</td>
                            <td className="text-center">{team1Captured}</td>
                            <td className="text-center">{team1Total}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                    <ScrollFadeHint />
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="tt-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[320px]">
                        <caption className="sr-only">{t('scoresCaption')}</caption>
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoresTeam')}
                            </th>
                            <th scope="col">{t('scoresCaptured')}</th>
                            <th scope="col">{t('scoresTotal')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr className="text-ds-accent">
                            <td>{t('team0')}</td>
                            <td className="text-center">{team0Captured}</td>
                            <td className="text-center">{team0Total}</td>
                          </tr>
                          <tr>
                            <td>{t('team1')}</td>
                            <td className="text-center">{team1Captured}</td>
                            <td className="text-center">{team1Total}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}
                <RoundScoreAnnouncement
                  active={isRoundEnd || isGameEnd}
                  entries={[
                    { name: t('team0'), roundScore: team0Captured, cumulativeScore: team0Total },
                    { name: t('team1'), roundScore: team1Captured, cumulativeScore: team1Total },
                  ]}
                />
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.twotenjack.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="tt"
                highlightIndices={isHumanTurn && hint?.cardIndex !== undefined ? [hint.cardIndex] : undefined}
              />
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2" data-testid="tt-hint">
                {hint.trumpSuit !== undefined
                  ? `${t('hintDeclare')}: ${trumpSymbol(hint.trumpSuit)} (${t(`hint.${hint.reason}`)})`
                  : (() => {
                      const card = hint.cardIndex !== undefined ? humanPlayer?.cards[hint.cardIndex] : undefined;
                      const name = card ? cardAlt(card) : '-';
                      return `${t('hintPlay')}: ${name} [${hint.cardIndex ?? '-'}] (${t(`hint.${hint.reason}`)})`;
                    })()}
              </div>
            )}

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="tt-play-button">
              {(isHumanDeclarer || isHumanTurn) && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleHint}
                  disabled={loading || hintLoading}
                  data-testid="tt-hint-button"
                >
                  {tc('button.hint')}
                </button>
              )}
              {isHumanDeclarer && (
                <span data-testid="tt-declare-prompt" className="text-ds-warning text-sm font-medium w-full sm:w-auto">
                  {t('declarePrompt')}
                </span>
              )}
              {isDeclarePhase && !isHumanDeclarer && (
                <span
                  data-testid="tt-cpu-declaring"
                  role="status"
                  className="flex items-center gap-2 text-ds-info text-sm font-medium w-full sm:w-auto"
                >
                  <span
                    aria-hidden="true"
                    className="inline-block w-3 h-3 rounded-full border-2 border-ds-info border-t-transparent motion-safe:animate-spin"
                  />
                  {t('cpuDeclaring')}
                </span>
              )}
              {isHumanDeclarer &&
                SUIT_OPTIONS.map((suit) => (
                  <button
                    key={suit.key}
                    type="button"
                    className={btnPrimary}
                    aria-label={t(`suit.${suit.key}`)}
                    onClick={() => handleDeclare(suit.value)}
                    disabled={loading}
                  >
                    {suit.symbol}
                  </button>
                ))}
              {isHumanTurn && (
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
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="tt-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="two-ten-jack-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
