import { useCallback, useMemo } from 'react';
import type { tressetteApi } from '../api/gameApi';
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
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useTressetteGame } from '../hooks/useTressetteGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TressetteResponse } from '../types/card';
import { TressettePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTressetteCommand, TRESSETTE_HELP } from '../utils/cli/commands/tressetteCommands';
import { formatTressetteState } from '../utils/cli/formatters/tressetteFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tressette tutorial step definitions. */
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

const TRESSETTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TressettePhase.PLAY]: 'play',
  [TressettePhase.TRICK_END]: 'trickEnd',
  [TressettePhase.ROUND_END]: 'roundEnd',
  [TressettePhase.GAME_END]: 'gameEnd',
};

/**
 * Clamps a team's round-thirds count to the [0, 3] range used by the 3-dot
 * indicator. Thirds reset to 0 once they reach 3 (converted to a point), but
 * this guards defensively against any out-of-range value.
 */
function thirdsFilled(thirds: number): number {
  return Math.min(Math.max(thirds, 0), 3);
}

/** Renders the Tressette game page: no-trump must-follow trick play with team scoring. */
export const TressettePage = withTutorial(TressettePageContent, 'tressette', TR_TUTORIAL_STEPS);

/** Inner content of the Tressette page, wrapped by TutorialProvider. */
function TressettePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tressette');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    tressetteConfig,
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
  } = useTressetteGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tressette');
  const tressetteCliConfig: CliGameConfig<TressetteResponse, Parameters<typeof tressetteApi.exec>> = useMemo(
    () => ({
      gameName: 'tressette',
      parseCommand: parseTressetteCommand,
      formatResponse: formatTressetteState,
      helpText: TRESSETTE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, tressetteCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tressette', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const isPlayPhaseForKbd = state?.phase === TressettePhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: handlePlay,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('tressette', TRESSETTE_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void exec('reset', undefined, undefined, {
      cpuDifficulty: tressetteConfig.cpuDifficulty,
      targetPoints: tressetteConfig.targetPoints,
    });
  }, [exec, hideActionLog, tressetteConfig.cpuDifficulty, tressetteConfig.targetPoints]);

  if (!state)
    return <GameSkeleton gameKey="tressette" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === TressettePhase.PLAY;
  const isTrickEnd = state.phase === TressettePhase.TRICK_END;
  const isRoundEnd = state.phase === TressettePhase.ROUND_END;
  const isGameEnd = state.phase === TressettePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  const teamLabels = ['A', 'B'];

  return (
    <GamePageShell
      title={tc('nav.tressette')}
      gameThemeBg={gameTheme.tressette.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/tressette"
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
                    value: tressetteConfig.cpuDifficulty,
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
                    value: tressetteConfig.targetPoints,
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
                        <th scope="col">{t('scoresThirds')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {teamLabels.map((label, idx) => (
                        <tr key={label} className={humanPlayer && humanPlayer.teamId === idx ? 'text-ds-accent' : ''}>
                          <td>{t('teamLabel', { team: label })}</td>
                          <td className="text-center">{state.teamScores[idx] ?? 0}</td>
                          <td className="text-center">
                            <span
                              className="inline-flex gap-0.5 align-middle"
                              title={t('thirdsTooltip', { n: 3 - thirdsFilled(state.teamRoundThirds[idx]) })}
                              data-testid={`tr-thirds-${idx.toString()}`}
                            >
                              {[0, 1, 2].map((d) => (
                                <span
                                  key={`tr-dot-${idx.toString()}-${d.toString()}`}
                                  aria-hidden="true"
                                  className={`inline-block w-2 h-2 rounded-full ${
                                    d < thirdsFilled(state.teamRoundThirds[idx])
                                      ? 'bg-ds-accent'
                                      : 'border border-ds-border-subtle'
                                  }`}
                                />
                              ))}
                              <span className="sr-only">
                                {t('thirdsAria', { filled: thirdsFilled(state.teamRoundThirds[idx]) })}
                              </span>
                            </span>
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
          <GameFooter className={`${gameTheme.tressette.footer} px-4 py-2.5`}>
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

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {hint.cardIndices.map((i) => `[${i}]`).join(', ')} (
                {t(`hintReason.${hint.reason}`)})
              </div>
            )}
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
            <CardNavShortcutsPanel data-testid="tressette-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
