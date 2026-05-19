import { useCallback, useMemo, useState } from 'react';
import type { tarneebApi } from '../api/gameApi';
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
  TARNEEB_CPU_DIFFICULTY_OPTIONS,
  TARNEEB_MIN_BID_OPTIONS,
  TARNEEB_POINT_LIMIT_OPTIONS,
  useTarneebGame,
} from '../hooks/useTarneebGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TarneebResponse } from '../types/card';
import { TarneebPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTarneebCommand, TARNEEB_HELP } from '../utils/cli/commands/tarneebCommands';
import { formatTarneebState } from '../utils/cli/formatters/tarneebFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Tutorial step definitions for the Tarneeb page. */
const TARNEEB_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tn-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-trump-controls"]',
    messageKey: 'tutorial.trumpControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Tarneeb phase enum → i18n key. */
const TARNEEB_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TarneebPhase.BID]: 'bid',
  [TarneebPhase.TRUMP_DECLARATION]: 'trump',
  [TarneebPhase.PLAY]: 'play',
  [TarneebPhase.TRICK_END]: 'trickEnd',
  [TarneebPhase.ROUND_END]: 'roundEnd',
  [TarneebPhase.GAME_END]: 'gameEnd',
};

/** Map suit code (1-4) → display label. */
const TRUMP_LABELS: Record<number, string> = {
  1: '♠',
  2: '♣',
  3: '♥',
  4: '♦',
};

/** Render the Tarneeb game page (partnership trick-taking with chosen trump). */
export const TarneebPage = withTutorial(TarneebPageContent, 'tarneeb', TARNEEB_TUTORIAL_STEPS);

/** Inner Tarneeb content. */
function TarneebPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tarneeb');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    tarneebConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handleDeclareTrump,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useTarneebGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tarneeb', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const [bidValue, setBidValue] = useState(7);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tarneeb');
  const cliConfig: CliGameConfig<TarneebResponse, Parameters<typeof tarneebApi.exec>> = useMemo(
    () => ({
      gameName: 'tarneeb',
      parseCommand: parseTarneebCommand,
      formatResponse: formatTarneebState,
      helpText: TARNEEB_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === TarneebPhase.PLAY;
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

  const phaseNames = usePhaseNames('tarneeb', TARNEEB_PHASE_KEYS);
  const runAction = exec;

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void runAction('reset', undefined, undefined, {
      cpuDifficulty: tarneebConfig.cpuDifficulty,
      pointLimit: tarneebConfig.pointLimit,
      minBid: tarneebConfig.minBid,
    });
  }, [runAction, hideActionLog, tarneebConfig.cpuDifficulty, tarneebConfig.pointLimit, tarneebConfig.minBid]);

  if (!state)
    return <GameSkeleton gameKey="tarneeb" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === TarneebPhase.BID;
  const isTrumpPhase = state.phase === TarneebPhase.TRUMP_DECLARATION;
  const isPlayPhase = state.phase === TarneebPhase.PLAY;
  const isTrickEnd = state.phase === TarneebPhase.TRICK_END;
  const isRoundEnd = state.phase === TarneebPhase.ROUND_END;
  const isGameEnd = state.phase === TarneebPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanTrumpTurn =
    isTrumpPhase && state.bidWinnerIdx >= 0 && state.players[state.bidWinnerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.tarneeb')}
      gameThemeBg={gameTheme.tarneeb.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTrumpTurn || isHumanTurn}
      gamePath="/tarneeb"
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
                    value: tarneebConfig.cpuDifficulty,
                    options: TARNEEB_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'pointLimit',
                    label: t('settings.pointLimit'),
                    value: tarneebConfig.pointLimit,
                    options: TARNEEB_POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  {
                    type: 'select',
                    id: 'minBid',
                    label: t('settings.minBid'),
                    value: tarneebConfig.minBid,
                    options: TARNEEB_MIN_BID_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('minBid', v),
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
              <span className="mr-4">
                {t('trump')}: {TRUMP_LABELS[state.trumpSuit] ?? t('trumpUndeclared')}
              </span>
              {state.highestBid > 0 && (
                <span>
                  {t('highestBid')}: {state.highestBid}
                </span>
              )}
            </div>

            <div className={lgTwoColGrid}>
              <div>
                {isHumanBidTurn && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="tn-bid-controls">
                    {t('bidPhase')}
                  </div>
                )}
                {isHumanTrumpTurn && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="tn-trump-controls">
                    {t('trumpPhase')}
                  </div>
                )}

                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="tn-trick-display"
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
                          <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                            {playerName(p.id, p.isHuman)} (T{p.team}): {t('cards', { count: p.cardCount })} |{' '}
                            {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
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
                          {playerName(p.id, p.isHuman)} (T{p.team}): {t('cards', { count: p.cardCount })} |{' '}
                          {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
                        </div>
                      </div>
                    ))
                )}

                <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="tn-score-table">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <div className="overflow-x-auto -mx-2 px-2">
                    <table className="w-full text-sm text-ds-text-muted min-w-[280px]">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('scoresTeam')}
                          </th>
                          <th scope="col">{t('scoresTotal')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.teamScores.map((score, i) => (
                          <tr key={i} className={humanPlayer && humanPlayer.team === i ? 'text-ds-accent' : ''}>
                            <td>{t('team', { n: i })}</td>
                            <td className="text-center">{score}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  <ScrollFadeHint />
                </div>
              </div>
            </div>

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

          <GameFooter className={`${gameTheme.tarneeb.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="tn"
              />
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {hint.bid != null
                  ? `${t('hintBid')}: ${hint.bid} (${t(`hintReason.${hint.reason}`)})`
                  : hint.trumpSuit != null
                    ? `${t('hintTrump')}: ${TRUMP_LABELS[hint.trumpSuit] ?? '?'} (${t(`hintReason.${hint.reason}`)})`
                    : `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="tn-play-button">
              {(isHumanBidTurn || isHumanTrumpTurn || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isHumanBidTurn && (
                <>
                  {/*
                    Valid Tarneeb bids are 0 (pass) or 7-13. The pass case is exposed via a
                    dedicated button below, so the input range is restricted to 7-13. Out-of-range
                    values are clamped on Bid-button click to avoid noisy backend errors.
                  */}
                  <input
                    type="number"
                    min={state.config.minBid}
                    max={13}
                    value={bidValue}
                    onChange={(e) => setBidValue(Number(e.target.value))}
                    className="w-16 px-2 py-1 rounded bg-white/20 text-ds-text-primary text-center"
                    aria-label="bid-input"
                  />
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => {
                      const clamped = Math.min(13, Math.max(state.config.minBid, bidValue));
                      handleBid(clamped);
                    }}
                    disabled={loading || bidValue < state.config.minBid || bidValue > 13}
                  >
                    {t('bidButton')}
                  </button>
                  <button type="button" className={btnSuccess} onClick={() => handleBid(0)} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}
              {isHumanTrumpTurn && (
                <div className="flex gap-1">
                  {[1, 2, 3, 4].map((suit) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleDeclareTrump(suit)}
                      disabled={loading}
                      aria-label={`trump-${suit}`}
                    >
                      {TRUMP_LABELS[suit]}
                    </button>
                  ))}
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
                dataTutorial="tn-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
