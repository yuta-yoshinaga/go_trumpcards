import { useCallback, useMemo } from 'react';
import type { heartsApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { HeartsSkeleton } from '../components/skeleton/HeartsSkeleton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useHeartsGame } from '../hooks/useHeartsGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { HeartsResponse } from '../types/card';
import { HeartsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { HEARTS_HELP, parseHeartsCommand } from '../utils/cli/commands/heartsCommands';
import { formatHeartsState } from '../utils/cli/formatters/heartsFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Hearts tutorial step definitions. */
const HT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ht-pass-area"]',
    messageKey: 'tutorial.passArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-penalty-info"]',
    messageKey: 'tutorial.penaltyInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const HEARTS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [HeartsPhase.PASS]: 'pass',
  [HeartsPhase.PLAY]: 'play',
  [HeartsPhase.TRICK_END]: 'trickEnd',
  [HeartsPhase.ROUND_END]: 'roundEnd',
  [HeartsPhase.GAME_END]: 'gameEnd',
};

const passDirectionKeys = ['left', 'right', 'across', 'none'] as const;

/** Renders the Hearts game page with card passing, trick play, and scoring. */
export function HeartsPage() {
  return (
    <TutorialWrapper gameName="hearts" steps={HT_TUTORIAL_STEPS}>
      <HeartsPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Hearts page, wrapped by TutorialProvider. */
function HeartsPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('hearts');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    heartsConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleToggle,
    handlePass,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useHeartsGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('hearts');
  const heartsCliConfig: CliGameConfig<HeartsResponse, Parameters<typeof heartsApi.exec>> = useMemo(
    () => ({
      gameName: 'hearts',
      parseCommand: parseHeartsCommand,
      formatResponse: formatHeartsState,
      helpText: HEARTS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, heartsCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('hearts', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const isPassPhaseForKbd = state?.phase === HeartsPhase.PASS;
  const isPlayPhaseForKbd = state?.phase === HeartsPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isPassPhaseForKbd) {
      handlePass();
    } else {
      handlePlay();
    }
  }, [isPassPhaseForKbd, handlePass, handlePlay]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: (isPassPhaseForKbd || !!isHumanTurnForKbd) && !loading,
  });

  const phaseNames = usePhaseNames('hearts', HEARTS_PHASE_KEYS);

  if (!state) return <HeartsSkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPassPhase = state.phase === HeartsPhase.PASS;
  const isPlayPhase = state.phase === HeartsPhase.PLAY;
  const isTrickEnd = state.phase === HeartsPhase.TRICK_END;
  const isRoundEnd = state.phase === HeartsPhase.ROUND_END;
  const isGameEnd = state.phase === HeartsPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.hearts')}
      gameThemeBg={gameTheme.hearts.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isPassPhase || isHumanTurn}
      gamePath="/hearts"
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
                    value: heartsConfig.cpuDifficulty,
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
                    value: heartsConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'omnibusJD',
                    label: t('settings.omnibusJD'),
                    checked: heartsConfig.omnibusJD,
                    onToggle: (v) => handleToggle('omnibusJD', v),
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
            <div className="text-white text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{state.heartsBroken ? t('heartsBroken') : t('heartsNotBroken')}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Pass direction (pass phase) */}
                {isPassPhase && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="ht-pass-area">
                    {t(`passDirection.${passDirectionKeys[state.passDirection]}`)}
                  </div>
                )}

                {/* Current trick */}
                {state.currentTrick.length > 0 && (
                  <div className="my-3 p-3 rounded bg-black/40" data-tutorial="ht-trick-display">
                    <div className="text-white/70 text-sm mb-1">{t('currentTrick')}</div>
                    <div className="flex gap-2">
                      {state.currentTrick.map((trickCard) => (
                        <div key={`trick-${trickCard.playerIdx}`} className="text-center">
                          <AnimatedCard
                            card={trickCard.card}
                            width={cardWidth}
                            onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                          />
                          <div className="text-game-text-muted text-xs mt-1">
                            {playerName(
                              state.players[trickCard.playerIdx]?.id ?? trickCard.playerIdx,
                              state.players[trickCard.playerIdx]?.isHuman ?? false,
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-white/70 text-sm">
                      {tc('game.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
                    </summary>
                    <div className="mt-1">
                      {state.players
                        .filter((p) => !p.isHuman)
                        .map((p) => (
                          <div key={p.id} className="text-white/70 text-sm py-0.5">
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
                        <div className="text-white/70 text-sm">
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
                    data-tutorial="ht-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-white/70 text-sm">{t('scores')}</summary>
                    <table className="w-full text-sm text-white/70 mt-1">
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
                  <div className="my-3 p-2 rounded bg-black/30" data-tutorial="ht-score-table">
                    <div className="text-white/70 text-sm mb-1">{t('scores')}</div>
                    <table className="w-full text-sm text-white/70">
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
            <div data-tutorial="ht-penalty-info">
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
          <GameFooter className={`${gameTheme.hearts.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="ht"
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

            <div className="flex gap-2 items-center" data-tutorial="ht-play-button">
              {(isPassPhase || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isPassPhase && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePass}
                  disabled={loading || selectedCardIndices.length !== 3}
                >
                  {t('passButton')}
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
              <button
                type="button"
                className={btnOutline}
                data-tutorial="ht-reset-button"
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    return exec('reset', undefined, undefined, {
                      cpuDifficulty: heartsConfig.cpuDifficulty,
                      pointLimit: heartsConfig.pointLimit,
                      omnibusJD: heartsConfig.omnibusJD,
                    });
                  })
                }
                disabled={loading}
              >
                {tc('button.reset')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
