import { useCallback, useMemo, useState } from 'react';
import type { callBreakApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BidProgressBar } from '../components/BidProgressBar';
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
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import {
  CALLBREAK_CPU_DIFFICULTY_OPTIONS,
  CALLBREAK_MAX_ROUNDS_OPTIONS,
  useCallBreakGame,
} from '../hooks/useCallBreakGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import i18n from '../i18n';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CallBreakResponse } from '../types/card';
import { CallBreakPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CALLBREAK_HELP, parseCallBreakCommand } from '../utils/cli/commands/callbreakCommands';
import { formatCallBreakState } from '../utils/cli/formatters/callbreakFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/**
 * Render a Call Break int×10 score (e.g. 41) as "X.Y" for the UI, using the
 * active locale's decimal separator. Mirrors `FormatCallBreakScore` in the Go
 * backend for ja/en (both use "."), while a locale that formats decimals with a
 * comma (e.g. "de-DE") renders "X,Y" instead of a hardcoded period. The `locale`
 * parameter defaults to the active i18n language and is injectable for testing.
 */
export function fmtScore(internal: number, locale: string = i18n.language): string {
  return new Intl.NumberFormat(locale, {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
    useGrouping: false,
  }).format(internal / 10);
}

/** Call Break tutorial step definitions. */
const CB_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cb-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-bags-info"]',
    messageKey: 'tutorial.bagsInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const CALLBREAK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CallBreakPhase.BID]: 'bid',
  [CallBreakPhase.PLAY]: 'play',
  [CallBreakPhase.TRICK_END]: 'trickEnd',
  [CallBreakPhase.ROUND_END]: 'roundEnd',
  [CallBreakPhase.GAME_END]: 'gameEnd',
};

/** Renders the Call Break game page with bidding, trick play, and decimal scoring. */
export const CallBreakPage = withTutorial(CallBreakPageContent, 'callbreak', CB_TUTORIAL_STEPS);
/** Inner content of the Call Break page, wrapped by TutorialProvider. */
function CallBreakPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('callbreak');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    callBreakConfig,
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
  } = useCallBreakGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('callbreak', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const [bidValue, setBidValue] = useState(1);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('callbreak');
  const cliConfig: CliGameConfig<CallBreakResponse, Parameters<typeof callBreakApi.exec>> = useMemo(
    () => ({
      gameName: 'callbreak',
      parseCommand: parseCallBreakCommand,
      formatResponse: formatCallBreakState,
      helpText: CALLBREAK_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === CallBreakPhase.PLAY;
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

  const phaseNames = usePhaseNames('callbreak', CALLBREAK_PHASE_KEYS);

  const runAction = exec;
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void runAction('reset', undefined, undefined, {
      cpuDifficulty: callBreakConfig.cpuDifficulty,
      maxRounds: callBreakConfig.maxRounds,
    });
  }, [runAction, hideActionLog, callBreakConfig.cpuDifficulty, callBreakConfig.maxRounds]);

  if (!state)
    return <GameSkeleton gameKey="callbreak" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === CallBreakPhase.BID;
  const isPlayPhase = state.phase === CallBreakPhase.PLAY;
  const isTrickEnd = state.phase === CallBreakPhase.TRICK_END;
  const isRoundEnd = state.phase === CallBreakPhase.ROUND_END;
  const isGameEnd = state.phase === CallBreakPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.callbreak')}
      gameThemeBg={gameTheme.callbreak.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn}
      gamePath="/callbreak"
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
                    value: callBreakConfig.cpuDifficulty,
                    options: CALLBREAK_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'maxRounds',
                    label: t('settings.maxRounds'),
                    value: callBreakConfig.maxRounds,
                    options: CALLBREAK_MAX_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('maxRounds', v),
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
              <span className="mr-4">{t('round', { n: state.roundNumber, max: state.config.maxRounds })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}</span>
            </div>

            <div className={lgTwoColGrid}>
              <div>
                {isHumanBidTurn && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="cb-bid-controls">
                    {t('bidPhase')}
                  </div>
                )}

                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="cb-trick-display"
                />
              </div>

              <div>
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                      {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
                    </summary>
                    <div className="mt-1">
                      {state.players
                        .filter((p) => !p.isHuman)
                        .map((p) => (
                          <div key={p.id} className="py-0.5 text-ds-text-muted text-sm">
                            <div>
                              {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                              {t('cumulativeScore', { score: fmtScore(p.cumulativeScore) })} |{' '}
                              {t('roundScore', { score: fmtScore(p.roundScore) })} |{' '}
                              {p.bid >= 0
                                ? `${t('bid', { n: p.bid })} / ${t('tricksWon', { n: p.trickCount })}`
                                : t('bidNone')}
                            </div>
                            <BidProgressBar
                              bid={p.bid}
                              tricksWon={p.trickCount}
                              ariaLabel={t('bidProgressAria', { name: playerName(p.id, p.isHuman) })}
                              testId={`bid-progress-${p.id.toString()}`}
                            />
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
                          {t('cumulativeScore', { score: fmtScore(p.cumulativeScore) })} |{' '}
                          {t('roundScore', { score: fmtScore(p.roundScore) })} |{' '}
                          {p.bid >= 0
                            ? `${t('bid', { n: p.bid })} / ${t('tricksWon', { n: p.trickCount })}`
                            : t('bidNone')}
                        </div>
                        <BidProgressBar
                          bid={p.bid}
                          tricksWon={p.trickCount}
                          ariaLabel={t('bidProgressAria', { name: playerName(p.id, p.isHuman) })}
                          testId={`bid-progress-${p.id.toString()}`}
                        />
                      </div>
                    ))
                )}

                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30 relative"
                    data-tutorial="cb-score-table"
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
                              <td className="text-center">{fmtScore(p.roundScore)}</td>
                              <td className="text-center">{fmtScore(p.cumulativeScore)}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    <ScrollFadeHint />
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="cb-score-table">
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
                              <td className="text-center">{fmtScore(p.roundScore)}</td>
                              <td className="text-center">{fmtScore(p.cumulativeScore)}</td>
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

            <div data-tutorial="cb-bags-info">
              <div
                className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                role="status"
                aria-label={t('bagsAria')}
                data-testid="cb-bags-counter"
              >
                <span className="mr-2">{t('bags')}:</span>
                {state.players.map((p) => {
                  const bags = p.bid >= 0 ? Math.max(0, p.trickCount - p.bid) : 0;
                  return (
                    <span key={p.id} className="mr-3" data-testid={`cb-bags-${p.id.toString()}`}>
                      {t('bagsValue', { name: playerName(p.id, p.isHuman), n: bags })}
                    </span>
                  );
                })}
              </div>
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

          <GameFooter className={`${gameTheme.callbreak.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <>
                <div className="mb-1 text-ds-text-muted text-xs" role="status" data-testid="cb-spades-break-footer">
                  {state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}
                </div>
                {humanPlayer.bid >= 0 && (
                  <div className="mb-1 text-ds-text-muted text-xs">
                    <div>
                      {t('bid', { n: humanPlayer.bid })} / {t('tricksWon', { n: humanPlayer.trickCount })}
                    </div>
                    <BidProgressBar
                      bid={humanPlayer.bid}
                      tricksWon={humanPlayer.trickCount}
                      ariaLabel={t('bidProgressAria', { name: playerName(humanPlayer.id, true) })}
                      testId={`bid-progress-${humanPlayer.id.toString()}`}
                    />
                  </div>
                )}
                <PlayerHandSection
                  humanPlayer={humanPlayer}
                  selectedCardIndices={selectedCardIndices}
                  toggleCard={toggleCard}
                  cardWidth={cardWidth}
                  isMobile={isMobile}
                  dataTutorialPrefix="cb"
                  validIndices={isHumanTurn ? state.validPlayIndices : undefined}
                  restrictedTooltip={t('mustTrumpSpadeTooltip')}
                />
              </>
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {hint.bid != null
                  ? `${t('hintBid')}: ${hint.bid} (${t(`hintReason.${hint.reason}`)})`
                  : `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center" data-tutorial="cb-play-button">
              {(isHumanBidTurn || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isHumanBidTurn && (
                <div className="flex flex-col items-center gap-2">
                  <fieldset
                    className="grid max-w-[16rem] grid-cols-7 gap-1 border-0 p-0"
                    aria-label={t('bidSelectLabel')}
                  >
                    {Array.from({ length: 13 }, (_, i) => i + 1).map((n) => (
                      <button
                        key={n}
                        type="button"
                        onClick={() => setBidValue(n)}
                        disabled={loading}
                        aria-pressed={bidValue === n}
                        data-testid={`bid-option-${n}`}
                        className={`h-9 w-9 rounded-lg font-medium text-sm transition-all ${
                          bidValue === n
                            ? 'bg-ds-accent text-white ring-2 ring-ds-accent'
                            : 'bg-white/20 text-ds-text-primary hover:bg-white/30'
                        }`}
                      >
                        {n}
                      </button>
                    ))}
                  </fieldset>
                  {/* Announce the current bid selection to screen readers, since the
                      grid's aria-pressed alone isn't read back as a running value. */}
                  <span className="sr-only" role="status" aria-live="polite" data-testid="cb-bid-selected">
                    {t('bidSelected', { n: bidValue })}
                  </span>
                  <button type="button" className={btnPrimary} onClick={() => handleBid(bidValue)} disabled={loading}>
                    {t('bidButton')}
                  </button>
                </div>
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
                dataTutorial="cb-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
