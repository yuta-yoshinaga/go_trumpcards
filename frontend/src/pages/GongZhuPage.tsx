import { useCallback, useMemo } from 'react';
import type { gongzhuApi } from '../api/gameApi';
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
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useGongZhuGame } from '../hooks/useGongZhuGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GongZhuResponse } from '../types/card';
import { GongZhuPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { GONGZHU_HELP, parseGongZhuCommand } from '../utils/cli/commands/gongzhuCommands';
import { formatGongZhuState } from '../utils/cli/formatters/gongzhuFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Gong Zhu tutorial step definitions. */
const GZ_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="gz-expose-area"]',
    messageKey: 'tutorial.exposeArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gz-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gz-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gz-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gz-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gz-point-info"]',
    messageKey: 'tutorial.pointInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gz-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GONGZHU_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GongZhuPhase.EXPOSE]: 'expose',
  [GongZhuPhase.PLAY]: 'play',
  [GongZhuPhase.TRICK_END]: 'trickEnd',
  [GongZhuPhase.ROUND_END]: 'roundEnd',
  [GongZhuPhase.GAME_END]: 'gameEnd',
};

/** Renders the Gong Zhu game page with an exposure phase, trick play, and scoring. */
export const GongZhuPage = withTutorial(GongZhuPageContent, 'gongzhu', GZ_TUTORIAL_STEPS);
/** Inner content of the Gong Zhu page, wrapped by TutorialProvider. */
function GongZhuPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('gongzhu');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    gongzhuConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleExpose,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useGongZhuGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('gongzhu');
  const gongzhuCliConfig: CliGameConfig<GongZhuResponse, Parameters<typeof gongzhuApi.exec>> = useMemo(
    () => ({
      gameName: 'gongzhu',
      parseCommand: parseGongZhuCommand,
      formatResponse: formatGongZhuState,
      helpText: GONGZHU_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, gongzhuCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('gongzhu', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const isExposePhaseForKbd = state?.phase === GongZhuPhase.EXPOSE;
  const isPlayPhaseForKbd = state?.phase === GongZhuPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isExposePhaseForKbd) {
      handleExpose();
    } else {
      handlePlay();
    }
  }, [isExposePhaseForKbd, handleExpose, handlePlay]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: (isExposePhaseForKbd || !!isHumanTurnForKbd) && !loading,
  });

  const phaseNames = usePhaseNames('gongzhu', GONGZHU_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void exec('reset', undefined, undefined, {
      cpuDifficulty: gongzhuConfig.cpuDifficulty,
      pointLimit: gongzhuConfig.pointLimit,
    });
  }, [exec, hideActionLog, gongzhuConfig.cpuDifficulty, gongzhuConfig.pointLimit]);

  const exposureSummary = useMemo(() => {
    if (!state) return '';
    const parts: string[] = [];
    if (state.exposed.pig) parts.push(t('card.spadeQueen'));
    if (state.exposed.sheep) parts.push(t('card.diamondJack'));
    if (state.exposed.ace) parts.push(t('card.heartAce'));
    if (state.exposed.doubler) parts.push(t('card.clubTen'));
    return parts.length > 0 ? parts.join(', ') : t('exposedNone');
  }, [state, t]);

  if (!state)
    return <GameSkeleton gameKey="gongzhu" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isExposePhase = state.phase === GongZhuPhase.EXPOSE;
  const isPlayPhase = state.phase === GongZhuPhase.PLAY;
  const isTrickEnd = state.phase === GongZhuPhase.TRICK_END;
  const isRoundEnd = state.phase === GongZhuPhase.ROUND_END;
  const isGameEnd = state.phase === GongZhuPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.gongzhu')}
      gameThemeBg={gameTheme.gongzhu.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isExposePhase || isHumanTurn}
      gamePath="/gongzhu"
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
          {/* Settings */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: gongzhuConfig.cpuDifficulty,
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
                    value: gongzhuConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
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

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Round/Trick info */}
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span data-testid="exposure-summary" role="img" aria-label={t('exposed', { cards: exposureSummary })}>
                {t('exposed', { cards: exposureSummary })}
              </span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Exposure hint (expose phase) */}
                {isExposePhase && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="gz-expose-area" role="status">
                    {state.exposableIndices.length > 0
                      ? t('exposableHint', {
                          // Include the card name after each index so SR users and
                          // beginners know which card [0]/[3]… refers to.
                          indices: state.exposableIndices
                            .map((i) => {
                              const c = humanPlayer?.cards[i];
                              return c ? `[${i}] ${cardAlt(c)}` : `[${i}]`;
                            })
                            .join(', '),
                        })
                      : t('exposableNone')}
                  </div>
                )}

                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="gz-trick-display"
                />
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
                            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                            {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                            {t('roundScore', { score: p.roundScore })}
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
                          {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                          {t('roundScore', { score: p.roundScore })}
                        </div>
                      </div>
                    ))
                )}

                {/* Score table */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30"
                    data-tutorial="gz-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
                    <table className="w-full text-sm text-ds-text-muted mt-1">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('scoresPlayer')}
                          </th>
                          <th scope="col">{t('scoresRound')}</th>
                          <th scope="col">{t('scoresTotal')}</th>
                          <th scope="col">{t('scoresTricks')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.players.map((p) => (
                          <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                            <td>{playerName(p.id, p.isHuman)}</td>
                            <td className="text-center">{p.roundScore}</td>
                            <td className="text-center">{p.cumulativeScore}</td>
                            <td className="text-center">{p.trickCount}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30" data-tutorial="gz-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                    <table className="w-full text-sm text-ds-text-muted">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('scoresPlayer')}
                          </th>
                          <th scope="col">{t('scoresRound')}</th>
                          <th scope="col">{t('scoresTotal')}</th>
                          <th scope="col">{t('scoresTricks')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.players.map((p) => (
                          <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                            <td>{playerName(p.id, p.isHuman)}</td>
                            <td className="text-center">{p.roundScore}</td>
                            <td className="text-center">{p.cumulativeScore}</td>
                            <td className="text-center">{p.trickCount}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
                <RoundScoreAnnouncement
                  active={isRoundEnd || isGameEnd}
                  entries={state.players.map((p) => ({
                    name: playerName(p.id, p.isHuman),
                    roundScore: p.roundScore,
                    cumulativeScore: p.cumulativeScore,
                  }))}
                />
              </div>
            </div>

            {/* Message */}
            <div data-tutorial="gz-point-info">
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />
            </div>

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.gongzhu.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="gz"
                highlightIndices={
                  isExposePhase && state.exposableIndices.length > 0 ? state.exposableIndices : undefined
                }
              />
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {hint.cardIndices.map((i) => `[${i}]`).join(', ')} (
                {t(`hintReason.${hint.reason}`)})
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center" data-tutorial="gz-play-button">
              {(isExposePhase || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isExposePhase && (
                <button type="button" className={btnPrimary} onClick={handleExpose} disabled={loading}>
                  {selectedCardIndices.length === 0 ? t('skipExpose') : t('exposeButton')}
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
                dataTutorial="gz-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
