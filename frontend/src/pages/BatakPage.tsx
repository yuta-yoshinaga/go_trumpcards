import { useCallback, useEffect, useMemo, useState } from 'react';
import type { batakApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BidProgressBar } from '../components/BidProgressBar';
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
import { BATAK_CPU_DIFFICULTY_OPTIONS, BATAK_MAX_ROUNDS_OPTIONS, useBatakGame } from '../hooks/useBatakGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BatakResponse } from '../types/card';
import { BatakPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BATAK_HELP, parseBatakCommand } from '../utils/cli/commands/batakCommands';
import { formatBatakState } from '../utils/cli/formatters/batakFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Batak tutorial step definitions. */
const BATAK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="batak-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BATAK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BatakPhase.BID]: 'bid',
  [BatakPhase.PLAY]: 'play',
  [BatakPhase.TRICK_END]: 'trickEnd',
  [BatakPhase.ROUND_END]: 'roundEnd',
  [BatakPhase.GAME_END]: 'gameEnd',
};

/** Renders the Batak game page with bidding, trick play, and scoring. */
export const BatakPage = withTutorial(BatakPageContent, 'batak', BATAK_TUTORIAL_STEPS);
/** Inner content of the Batak page, wrapped by TutorialProvider. */
function BatakPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('batak');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    batakConfig,
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
  } = useBatakGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('batak', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const minLegalBid = state?.minLegalBid ?? 0;
  const [bidValue, setBidValue] = useState(5);

  useEffect(() => {
    if (state?.minLegalBid && state.minLegalBid > 0) {
      setBidValue((prev) => (prev < state.minLegalBid || prev > 13 ? state.minLegalBid : prev));
    }
  }, [state?.minLegalBid]);

  const effectiveBidValue = minLegalBid > 0 ? Math.max(minLegalBid, Math.min(13, bidValue)) : 0;

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('batak');
  const cliConfig: CliGameConfig<BatakResponse, Parameters<typeof batakApi.exec>> = useMemo(
    () => ({
      gameName: 'batak',
      parseCommand: parseBatakCommand,
      formatResponse: formatBatakState,
      helpText: BATAK_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === BatakPhase.PLAY;
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

  const phaseNames = usePhaseNames('batak', BATAK_PHASE_KEYS);

  const runAction = exec;
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void runAction('reset', undefined, undefined, {
      cpuDifficulty: batakConfig.cpuDifficulty,
      maxRounds: batakConfig.maxRounds,
    });
  }, [runAction, hideActionLog, batakConfig.cpuDifficulty, batakConfig.maxRounds]);

  if (!state)
    return <GameSkeleton gameKey="batak" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === BatakPhase.BID;
  const isPlayPhase = state.phase === BatakPhase.PLAY;
  const isTrickEnd = state.phase === BatakPhase.TRICK_END;
  const isRoundEnd = state.phase === BatakPhase.ROUND_END;
  const isGameEnd = state.phase === BatakPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.batak')}
      gameThemeBg={gameTheme.batak.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn}
      gamePath="/batak"
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
                    value: batakConfig.cpuDifficulty,
                    options: BATAK_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'maxRounds',
                    label: t('settings.maxRounds'),
                    value: batakConfig.maxRounds,
                    options: BATAK_MAX_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('maxRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber, max: state.config.maxRounds })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}</span>
              {state.declarerIdx >= 0 && (
                <span data-testid="batak-declarer">
                  {t('declarer', { name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman) })}
                </span>
              )}
            </div>

            <div className={lgTwoColGrid}>
              <div>
                {isHumanBidTurn && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="batak-bid-controls">
                    {t('bidPhase')}
                  </div>
                )}

                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="batak-trick-display"
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
                              {playerName(p.id, p.isHuman)}
                              {p.id === state.declarerIdx ? ` (${t('declarerBadge')})` : ''}:{' '}
                              {t('cards', { count: p.cardCount })} |{' '}
                              {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                              {t('roundScore', { score: p.roundScore })} |{' '}
                              {p.bid < 0
                                ? t('bidNone')
                                : `${p.bid === 0 ? t('bidPass') : t('bid', { n: p.bid })} / ${t('tricksWon', { n: p.trickCount })}`}
                            </div>
                            {p.id === state.declarerIdx && p.bid > 0 && (
                              <BidProgressBar
                                bid={p.bid}
                                tricksWon={p.trickCount}
                                ariaLabel={t('bidProgressAria', { name: playerName(p.id, p.isHuman) })}
                                testId={`bid-progress-${p.id.toString()}`}
                              />
                            )}
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
                          {playerName(p.id, p.isHuman)}
                          {p.id === state.declarerIdx ? ` (${t('declarerBadge')})` : ''}:{' '}
                          {t('cards', { count: p.cardCount })} | {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                          {t('roundScore', { score: p.roundScore })} |{' '}
                          {p.bid < 0
                            ? t('bidNone')
                            : `${p.bid === 0 ? t('bidPass') : t('bid', { n: p.bid })} / ${t('tricksWon', { n: p.trickCount })}`}
                        </div>
                        {p.id === state.declarerIdx && p.bid > 0 && (
                          <BidProgressBar
                            bid={p.bid}
                            tricksWon={p.trickCount}
                            ariaLabel={t('bidProgressAria', { name: playerName(p.id, p.isHuman) })}
                            testId={`bid-progress-${p.id.toString()}`}
                          />
                        )}
                      </div>
                    ))
                )}

                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30 relative"
                    data-tutorial="batak-score-table"
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
                              <td>
                                {playerName(p.id, p.isHuman)}
                                {p.id === state.declarerIdx ? ` (${t('declarerBadge')})` : ''}
                              </td>
                              <td className="text-center">{p.bid < 0 ? '-' : p.bid === 0 ? t('bidPass') : p.bid}</td>
                              <td className="text-center">{p.trickCount}</td>
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
                  <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="batak-score-table">
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
                              <td>
                                {playerName(p.id, p.isHuman)}
                                {p.id === state.declarerIdx ? ` (${t('declarerBadge')})` : ''}
                              </td>
                              <td className="text-center">{p.bid < 0 ? '-' : p.bid === 0 ? t('bidPass') : p.bid}</td>
                              <td className="text-center">{p.trickCount}</td>
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

            <div>
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

          <GameFooter className={`${gameTheme.batak.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <>
                <div className="mb-1 text-ds-text-muted text-xs" role="status" data-testid="batak-spades-break-footer">
                  {state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}
                </div>
                {humanPlayer.bid >= 0 && (
                  <div className="mb-1 text-ds-text-muted text-xs">
                    <div>
                      {humanPlayer.id === state.declarerIdx ? `(${t('declarerBadge')}) ` : ''}
                      {humanPlayer.bid === 0 ? t('bidPass') : t('bid', { n: humanPlayer.bid })} /{' '}
                      {t('tricksWon', { n: humanPlayer.trickCount })}
                    </div>
                    {humanPlayer.id === state.declarerIdx && humanPlayer.bid > 0 && (
                      <BidProgressBar
                        bid={humanPlayer.bid}
                        tricksWon={humanPlayer.trickCount}
                        ariaLabel={t('bidProgressAria', { name: playerName(humanPlayer.id, true) })}
                        testId={`bid-progress-${humanPlayer.id.toString()}`}
                      />
                    )}
                  </div>
                )}
                <PlayerHandSection
                  humanPlayer={humanPlayer}
                  selectedCardIndices={selectedCardIndices}
                  toggleCard={toggleCard}
                  cardWidth={cardWidth}
                  isMobile={isMobile}
                  dataTutorialPrefix="batak"
                  validIndices={isHumanTurn ? state.validPlayIndices : undefined}
                  restrictedTooltip={t('mustTrumpSpadeTooltip')}
                />
              </>
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
                付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
                読み上げられないことがある (#5955, #6663)。 */}
            <div data-testid="batak-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {hint.bid != null
                    ? `${t('hintBid')}: ${hint.bid === 0 ? t('bidPass') : hint.bid} (${t(`hintReason.${hint.reason}`)})`
                    : (() => {
                        const card = hint.cardIndex !== undefined ? humanPlayer?.cards[hint.cardIndex] : undefined;
                        const name = card ? cardAlt(card) : '-';
                        return `${t('hintPlay')}: ${name} [${hint.cardIndex ?? '-'}] (${t(`hintReason.${hint.reason}`)})`;
                      })()}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center" data-tutorial="batak-play-button">
              {(isHumanBidTurn || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isHumanBidTurn && (
                <div className="flex flex-col items-center gap-2">
                  {minLegalBid > 0 && (
                    <>
                      <fieldset
                        className="grid max-w-[16rem] grid-cols-7 gap-1 border-0 p-0"
                        aria-label={t('bidSelectLabel')}
                      >
                        {Array.from({ length: 13 - minLegalBid + 1 }, (_, i) => minLegalBid + i).map((n) => (
                          <button
                            key={n}
                            type="button"
                            onClick={() => setBidValue(n)}
                            disabled={loading}
                            aria-pressed={effectiveBidValue === n}
                            data-testid={`bid-option-${n}`}
                            className={`h-9 w-9 rounded-lg font-medium text-sm transition-all ${
                              effectiveBidValue === n
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
                      <span className="sr-only" role="status" aria-live="polite" data-testid="batak-bid-selected">
                        {t('bidSelected', { n: effectiveBidValue })}
                      </span>
                    </>
                  )}
                  <div className="flex gap-2">
                    {minLegalBid > 0 && (
                      <button
                        type="button"
                        className={btnPrimary}
                        onClick={() => handleBid(effectiveBidValue)}
                        disabled={loading}
                      >
                        {t('bidButton')}
                      </button>
                    )}
                    <button
                      type="button"
                      className={btnSecondary}
                      onClick={() => handleBid(0)}
                      disabled={loading}
                      data-testid="bid-pass"
                    >
                      {t('passButton')}
                    </button>
                  </div>
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
                dataTutorial="batak-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="batak-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
