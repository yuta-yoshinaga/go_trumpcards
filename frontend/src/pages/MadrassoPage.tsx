import { useCallback, useMemo } from 'react';
import type { madrassoApi } from '../api/gameApi';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useMadrassoGame } from '../hooks/useMadrassoGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MadrassoResponse } from '../types/card';
import { MadrassoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MADRASSO_HELP, parseMadrassoCommand } from '../utils/cli/commands/madrassoCommands';
import { formatMadrassoState } from '../utils/cli/formatters/madrassoFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Madrasso tutorial step definitions. */
const TR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tr-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-point-info"]',
    messageKey: 'tutorial.pointInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MADRASSO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MadrassoPhase.PLAY]: 'play',
  [MadrassoPhase.TRICK_END]: 'trickEnd',
  [MadrassoPhase.ROUND_END]: 'roundEnd',
  [MadrassoPhase.GAME_END]: 'gameEnd',
};

/**
 * Clamps a team's round score for display.
 *
 * **Madrasso scores in whole points, not thirds.** The clone source (Tressette)
 * counts 1/3 points and shows a 3-dot indicator that fills and resets; here the
 * round total is simply the card points taken so far, out of
 * {@link MADRASSO_ROUND_POINTS}.
 */
function roundPointsShown(points: number): number {
  return Math.min(Math.max(points, 0), MADRASSO_ROUND_POINTS);
}

/** Total card points contested in one deal (sync: MadrassoRoundPoints in Go). */
const MADRASSO_ROUND_POINTS = 121;

/**
 * Renders the Madrasso game page: must-follow trick play with a trump decided by
 * the deal, integer card points, and team scoring.
 */
export const MadrassoPage = withTutorial(MadrassoPageContent, 'madrasso', TR_TUTORIAL_STEPS);

/** Inner content of the Madrasso page, wrapped by TutorialProvider. */
function MadrassoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('madrasso');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    madrassoConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useMadrassoGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('madrasso');
  const madrassoCliConfig: CliGameConfig<MadrassoResponse, Parameters<typeof madrassoApi.exec>> = useMemo(
    () => ({
      gameName: 'madrasso',
      parseCommand: parseMadrassoCommand,
      formatResponse: formatMadrassoState,
      helpText: MADRASSO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, madrassoCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('madrasso', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const isPlayPhaseForKbd = state?.phase === MadrassoPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: handlePlay,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('madrasso', MADRASSO_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void exec('reset', undefined, undefined, {
      cpuDifficulty: madrassoConfig.cpuDifficulty,
      targetPoints: madrassoConfig.targetPoints,
    });
  }, [exec, hideActionLog, madrassoConfig.cpuDifficulty, madrassoConfig.targetPoints]);

  if (!state)
    return <GameSkeleton gameKey="madrasso" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === MadrassoPhase.PLAY;
  const isTrickEnd = state.phase === MadrassoPhase.TRICK_END;
  const isRoundEnd = state.phase === MadrassoPhase.ROUND_END;
  const isGameEnd = state.phase === MadrassoPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  const teamLabels = ['A', 'B'];

  return (
    <GamePageShell
      title={tc('nav.madrasso')}
      gameThemeBg={gameTheme.madrasso.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/madrasso"
      gameEndFlag={!!state?.gameEndFlag}
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
                    value: madrassoConfig.cpuDifficulty,
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
                    value: madrassoConfig.targetPoints,
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
                  dataTutorial="tr-trick-display"
                />

                {/* Previous trick reviewer: lets the player recount the just-completed trick */}
                <details className="mb-2 p-2 rounded bg-black/30" data-testid="tr-previous-trick">
                  <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                    {t('previousTrick')}
                  </summary>
                  <div className="mt-1">
                    {state.lastTrick.length > 0 ? (
                      <TrickDisplay
                        currentTrick={state.lastTrick}
                        players={state.players}
                        cardWidth={Math.round(cardWidth * 0.7)}
                        label={
                          state.lastTrickWinner >= 0
                            ? t('previousTrickWinner', {
                                name: playerName(
                                  state.lastTrickWinner,
                                  state.players[state.lastTrickWinner]?.isHuman === true,
                                ),
                              })
                            : t('previousTrick')
                        }
                        winnerIdx={state.lastTrickWinner >= 0 ? state.lastTrickWinner : undefined}
                      />
                    ) : (
                      <div className="text-ds-text-muted text-sm">{t('previousTrickEmpty')}</div>
                    )}
                  </div>
                </details>
              </div>

              {/* Right: info sidebar */}
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
                            {playerName(p.id, p.isHuman)} [{t('teamLabel', { team: teamLabels[p.teamId] })}]:{' '}
                            {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
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
                          {playerName(p.id, p.isHuman)} [{t('teamLabel', { team: teamLabels[p.teamId] })}]:{' '}
                          {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                        </div>
                      </div>
                    ))
                )}

                {/* Team score table */}
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="tr-score-table">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresTeam')}
                        </th>
                        <th scope="col">{t('scoresPoints')}</th>
                        <th scope="col">{t('scoresRoundPoints')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {teamLabels.map((label, idx) => (
                        <tr key={label} className={humanPlayer && humanPlayer.teamId === idx ? 'text-ds-accent' : ''}>
                          <td>{t('teamLabel', { team: label })}</td>
                          <td className="text-center">{state.teamScores[idx] ?? 0}</td>
                          <td className="text-center" data-testid={`tr-round-points-${idx.toString()}`}>
                            {t('roundPointsOf', {
                              points: roundPointsShown(state.teamRoundPoints[idx] ?? 0),
                              total: MADRASSO_ROUND_POINTS,
                            })}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* Card-point legend (static reference) */}
                <details className="my-3 p-2 rounded bg-black/30" data-testid="tr-point-legend">
                  <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                    {t('pointLegend.title')}
                  </summary>
                  <div className="mt-1 text-ds-text-muted text-xs">
                    <table className="w-full">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left font-normal">
                            {t('pointLegend.cardCol')}
                          </th>
                          <th scope="col" className="text-right font-normal">
                            {t('pointLegend.pointCol')}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr>
                          <td>{t('pointLegend.ace')}</td>
                          <td className="text-right">{t('pointLegend.aceValue')}</td>
                        </tr>
                        <tr>
                          <td>{t('pointLegend.figures')}</td>
                          <td className="text-right">{t('pointLegend.figuresValue')}</td>
                        </tr>
                        <tr>
                          <td>{t('pointLegend.others')}</td>
                          <td className="text-right">{t('pointLegend.othersValue')}</td>
                        </tr>
                        <tr>
                          <td>{t('pointLegend.lastTrick')}</td>
                          <td className="text-right">{t('pointLegend.lastTrickValue')}</td>
                        </tr>
                      </tbody>
                    </table>
                    <div className="mt-1">{t('pointLegend.note')}</div>
                  </div>
                </details>
              </div>
            </div>

            {/* Message */}
            <div data-tutorial="tr-point-info">
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />
            </div>

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.madrasso.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="tr"
                validIndices={isHumanTurn ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-testid="madrasso-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {hint.cardIndices.map((i) => `[${i}]`).join(', ')} (
                  {t(`hintReason.${hint.reason}`)})
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center" data-tutorial="tr-play-button">
              {isHumanTurn && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
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
                dataTutorial="tr-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="madrasso-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
