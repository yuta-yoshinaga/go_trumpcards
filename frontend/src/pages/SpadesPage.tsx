import { useCallback, useMemo, useState } from 'react';
import type { spadesApi } from '../api/gameApi';
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
import {
  BAG_PENALTY_THRESHOLD_OPTIONS,
  CPU_DIFFICULTY_OPTIONS,
  NIL_BONUS_OPTIONS,
  POINT_LIMIT_OPTIONS,
  useSpadesGame,
} from '../hooks/useSpadesGame';
import { badgeError, badgeInfo, badgeSuccess, badgeWarning } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SpadesResponse } from '../types/card';
import { SpadesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSpadesCommand, SPADES_HELP } from '../utils/cli/commands/spadesCommands';
import { formatSpadesState } from '../utils/cli/formatters/spadesFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { spadesBagWarning, spadesBidProgress } from '../utils/spadesBid';

/** Spades tutorial step definitions (exported so tests can assert the rules are covered). */
export const SP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sp-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  // **標準スペードと違う点はスコア表の直後に言う。** この実装は4人が個別に
  // 得点を競うカットスロート方式で、ScoreRound にチームの概念が無い。パートナーが
  // 表示されないことに戸惑うのは、標準ルールを知っているプレイヤーほど強い (#5498)。
  {
    target: '[data-tutorial="sp-score-table"]',
    messageKey: 'tutorial.cutthroat',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-bags-info"]',
    messageKey: 'tutorial.bagsInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SPADES_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SpadesPhase.BID]: 'bid',
  [SpadesPhase.PLAY]: 'play',
  [SpadesPhase.TRICK_END]: 'trickEnd',
  [SpadesPhase.ROUND_END]: 'roundEnd',
  [SpadesPhase.GAME_END]: 'gameEnd',
};

/** Renders the Spades game page with bidding, trick play, and scoring. */
export const SpadesPage = withTutorial(SpadesPageContent, 'spades', SP_TUTORIAL_STEPS);
/** Inner content of the Spades page, wrapped by TutorialProvider. */
function SpadesPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('spades');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    spadesConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useSpadesGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('spades', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const [bidValue, setBidValue] = useState(1);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('spades');
  const cliConfig: CliGameConfig<SpadesResponse, Parameters<typeof spadesApi.exec>> = useMemo(
    () => ({
      gameName: 'spades',
      parseCommand: parseSpadesCommand,
      formatResponse: formatSpadesState,
      helpText: SPADES_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === SpadesPhase.PLAY;
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

  const phaseNames = usePhaseNames('spades', SPADES_PHASE_KEYS);

  const runAction = exec;
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void runAction('reset', undefined, undefined, {
      cpuDifficulty: spadesConfig.cpuDifficulty,
      pointLimit: spadesConfig.pointLimit,
      nilBonus: spadesConfig.nilBonus,
      bagPenaltyThreshold: spadesConfig.bagPenaltyThreshold,
    });
  }, [
    runAction,
    hideActionLog,
    spadesConfig.cpuDifficulty,
    spadesConfig.pointLimit,
    spadesConfig.nilBonus,
    spadesConfig.bagPenaltyThreshold,
  ]);

  if (!state)
    return <GameSkeleton gameKey="spades" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === SpadesPhase.BID;
  const isPlayPhase = state.phase === SpadesPhase.PLAY;
  const isTrickEnd = state.phase === SpadesPhase.TRICK_END;
  const isRoundEnd = state.phase === SpadesPhase.ROUND_END;
  const isGameEnd = state.phase === SpadesPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  // Bid-contract progress for the human player, shown during the play phase.
  const bidProgress =
    (isPlayPhase || isTrickEnd) && humanPlayer && humanPlayer.bid >= 0
      ? spadesBidProgress(humanPlayer.bid, humanPlayer.trickCount)
      : null;
  // Bag-penalty proximity warning for the human player.
  const bagThreshold = state.config.bagPenaltyThreshold;
  const humanBagWarning = humanPlayer ? spadesBagWarning(humanPlayer.bags, bagThreshold) : null;
  // Warning color for a score-table bags cell based on that player's own bags.
  const bagCellClass = (bags: number): string => {
    const w = spadesBagWarning(bags, bagThreshold);
    if (!w) return '';
    return w.level === 'danger' ? 'text-ds-warning font-bold' : 'text-ds-warning';
  };

  return (
    <GamePageShell
      title={tc('nav.spades')}
      gameThemeBg={gameTheme.spades.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn}
      gamePath="/spades"
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
                    value: spadesConfig.cpuDifficulty,
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
                    value: spadesConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  {
                    type: 'select',
                    id: 'nilBonus',
                    label: t('settings.nilBonus'),
                    value: spadesConfig.nilBonus,
                    options: NIL_BONUS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('nilBonus', v),
                    tooltip: t('settings.nextGameNote'),
                    testId: 'sp-setting-nil-bonus',
                  },
                  {
                    type: 'select',
                    id: 'bagPenaltyThreshold',
                    label: t('settings.bagPenaltyThreshold'),
                    value: spadesConfig.bagPenaltyThreshold,
                    options: BAG_PENALTY_THRESHOLD_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('bagPenaltyThreshold', v),
                    tooltip: t('settings.nextGameNote'),
                    testId: 'sp-setting-bag-threshold',
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
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
              <span>{state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}</span>
            </div>

            {(bidProgress || humanBagWarning) && (
              <div className="text-center mb-2 flex flex-wrap gap-2 justify-center items-center">
                {bidProgress && (
                  <span
                    data-testid="sp-bid-progress"
                    className={
                      bidProgress.kind === 'nilFail'
                        ? badgeWarning
                        : bidProgress.kind === 'made'
                          ? badgeSuccess
                          : badgeInfo
                    }
                  >
                    {bidProgress.kind === 'remaining'
                      ? t('bidProgress.remaining', { n: bidProgress.remaining })
                      : bidProgress.kind === 'made'
                        ? t('bidProgress.made', { bags: bidProgress.bags })
                        : bidProgress.kind === 'nilFail'
                          ? t('bidProgress.nilFail')
                          : t('bidProgress.nilOk')}
                  </span>
                )}
                {humanBagWarning && (
                  <span
                    data-testid="sp-bag-warning"
                    className={
                      humanBagWarning.level === 'danger' ? `${badgeError} motion-safe:animate-pulse` : badgeWarning
                    }
                  >
                    {t('bagWarning', { bags: humanBagWarning.bags, threshold: humanBagWarning.threshold })}
                  </span>
                )}
              </div>
            )}

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Bid phase instruction */}
                {isHumanBidTurn && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="sp-bid-controls">
                    {t('bidPhase')}
                  </div>
                )}

                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="sp-trick-display"
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
                            {t('roundScore', { score: p.roundScore })} |{' '}
                            {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')} | {t('bags', { count: p.bags })}
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
                          {t('roundScore', { score: p.roundScore })} |{' '}
                          {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')} | {t('bags', { count: p.bags })}
                        </div>
                      </div>
                    ))
                )}

                {/* Score table */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30 relative"
                    data-tutorial="sp-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[360px] mt-1">
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoresPlayer')}
                            </th>
                            <th scope="col">{t('scoresBid')}</th>
                            <th scope="col">{t('scoresTricks')}</th>
                            <th scope="col">{t('scoresBags')}</th>
                            <th scope="col">{t('scoresRound')}</th>
                            <th scope="col">{t('scoresTotal')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {state.players.map((p) => (
                            <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                              <td>{playerName(p.id, p.isHuman)}</td>
                              <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                              <td className="text-center">{p.trickCount}</td>
                              <td
                                className={`text-center ${bagCellClass(p.bags)}`}
                                data-testid={`sp-bags-cell-${p.id}`}
                              >
                                {p.bags}
                              </td>
                              <td className="text-center">{p.roundScore}</td>
                              <td className="text-center">{p.cumulativeScore}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    <ScrollFadeHint />
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="sp-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[360px]">
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoresPlayer')}
                            </th>
                            <th scope="col">{t('scoresBid')}</th>
                            <th scope="col">{t('scoresTricks')}</th>
                            <th scope="col">{t('scoresBags')}</th>
                            <th scope="col">{t('scoresRound')}</th>
                            <th scope="col">{t('scoresTotal')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {state.players.map((p) => (
                            <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                              <td>{playerName(p.id, p.isHuman)}</td>
                              <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                              <td className="text-center">{p.trickCount}</td>
                              <td
                                className={`text-center ${bagCellClass(p.bags)}`}
                                data-testid={`sp-bags-cell-${p.id}`}
                              >
                                {p.bags}
                              </td>
                              <td className="text-center">{p.roundScore}</td>
                              <td className="text-center">{p.cumulativeScore}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
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
            <div data-tutorial="sp-bags-info">
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
          <GameFooter className={`${gameTheme.spades.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="sp"
                validIndices={isHumanTurn ? state.validPlayIndices : undefined}
                restrictedTooltip={t('restrictedCard')}
              />
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {hint.bid != null
                  ? `${t('hintBid')}: ${hint.bid} (${t(`hintReason.${hint.reason}`)})`
                  : `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center" data-tutorial="sp-play-button">
              {(isHumanBidTurn || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isHumanBidTurn && (
                <>
                  <fieldset className="m-0 border-0 p-0" aria-label={t('bidPhase')}>
                    <div className="flex flex-wrap gap-1.5 justify-center">
                      <button
                        type="button"
                        className={
                          bidValue === 0
                            ? `${btnSecondary} min-w-[44px] ring-2 ring-ds-warning`
                            : `${btnSecondary} min-w-[44px]`
                        }
                        aria-pressed={bidValue === 0}
                        onClick={() => setBidValue(0)}
                        disabled={loading}
                      >
                        {t('nilButton')}
                      </button>
                      {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => (
                        <button
                          key={v}
                          type="button"
                          className={
                            bidValue === v
                              ? `${btnSecondary} min-w-[44px] ring-2 ring-ds-warning`
                              : `${btnSecondary} min-w-[44px]`
                          }
                          aria-pressed={bidValue === v}
                          onClick={() => setBidValue(v)}
                          disabled={loading}
                        >
                          {v}
                        </button>
                      ))}
                    </div>
                  </fieldset>
                  <button type="button" className={btnPrimary} onClick={() => handleBid(bidValue)} disabled={loading}>
                    {t('bidButton')}
                  </button>
                </>
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
                dataTutorial="sp-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="spades-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
