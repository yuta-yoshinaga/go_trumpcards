import { useCallback, useMemo } from 'react';
import type { ginrummyApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useGinRummyGame } from '../hooks/useGinRummyGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, meldCardStyle, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GinRummyResponse } from '../types/card';
import { GinRummyPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { GINRUMMY_HELP, parseGinrummyCommand } from '../utils/cli/commands/ginrummyCommands';
import { formatGinrummyState } from '../utils/cli/formatters/ginrummyFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import {
  bestDeadwoodValue,
  bestMeldSplit,
  GIN_RUMMY_KNOCK_THRESHOLD,
  ginRummyMeldLabel,
  ginRummyScoreBreakdown,
} from '../utils/ginRummyDeadwood';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const GINRUMMY_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GinRummyPhase.DRAW]: 'draw',
  [GinRummyPhase.DISCARD]: 'discard',
  [GinRummyPhase.LAYOFF]: 'layoff',
  [GinRummyPhase.ROUND_END]: 'roundEnd',
  [GinRummyPhase.GAME_END]: 'gameEnd',
};

/** Gin Rummy tutorial step definitions. */
const GR_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="gr-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="gr-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-knock-button"]',
    messageKey: 'tutorial.knockButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Gin Rummy game page with draw, discard, knock, and layoff phases. */
export const GinRummyPage = withTutorial(GinRummyPageContent, 'ginrummy', GR_TUTORIAL_STEPS);
/** Inner content of the Gin Rummy page, wrapped by TutorialProvider. */
function GinRummyPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ginrummy');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    ginRummyConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleKnock,
    handleLayoff,
    handleSkipLayoff,
    handleNextRound,
  } = useGinRummyGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('ginrummy', state);
  const { cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('ginrummy');
  const cliConfig: CliGameConfig<GinRummyResponse, Parameters<typeof ginrummyApi.exec>> = useMemo(
    () => ({
      gameName: 'ginrummy',
      parseCommand: parseGinrummyCommand,
      formatResponse: formatGinrummyState,
      helpText: GINRUMMY_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDiscardPhaseForKbd = state?.phase === GinRummyPhase.DISCARD;
  const isLayoffPhaseForKbd = state?.phase === GinRummyPhase.LAYOFF;
  const isHumanTurnForKbd =
    (isDiscardPhaseForKbd || isLayoffPhaseForKbd) && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isDiscardPhaseForKbd) {
      handleDiscard();
    } else if (isLayoffPhaseForKbd) {
      handleLayoff();
    }
  }, [isDiscardPhaseForKbd, isLayoffPhaseForKbd, handleDiscard, handleLayoff]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('ginrummy', GINRUMMY_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: ginRummyConfig.cpuDifficulty,
      pointLimit: ginRummyConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, ginRummyConfig.cpuDifficulty, ginRummyConfig.pointLimit]);

  // Knock requires the *post-discard* hand to be ≤ 10. When the user has a
  // card selected we project that discard; otherwise show the best-case
  // deadwood across all possible single discards (a knockability hint).
  const liveDeadwood = useMemo<number | null>(() => {
    if (!state) return null;
    if (state.phase !== GinRummyPhase.DISCARD) return null;
    const human = state.players.find((p) => p.isHuman);
    if (!human) return null;
    const cards = human.cards;
    if (cards.length === 0) return null;
    if (selectedCardIndices.length === 1) {
      const drop = selectedCardIndices[0];
      return bestDeadwoodValue(cards.filter((_, i) => i !== drop));
    }
    let best = Number.POSITIVE_INFINITY;
    for (let i = 0; i < cards.length; i++) {
      const dv = bestDeadwoodValue(cards.filter((_, j) => j !== i));
      if (dv < best) best = dv;
      if (best === 0) break;
    }
    return Number.isFinite(best) ? best : null;
  }, [state, selectedCardIndices]);
  const canKnockNow = liveDeadwood != null && liveDeadwood <= GIN_RUMMY_KNOCK_THRESHOLD;

  // During DISCARD, color-code the human's hand by best meld split so the
  // player can see which cards form melds vs. which are deadwood. Shares the
  // same search as the deadwood indicator, so the two always agree. Empty set
  // outside DISCARD leaves every card in its neutral (uncolored) state.
  const meldedIndices = useMemo<ReadonlySet<number>>(() => {
    if (!state || state.phase !== GinRummyPhase.DISCARD) return new Set();
    const human = state.players.find((p) => p.isHuman);
    if (!human || human.cards.length === 0) return new Set();
    return bestMeldSplit(human.cards).meldedIndices;
  }, [state]);

  // At round/game end, derive the additive score breakdown (outcome, each
  // player's deadwood, and the deadwood-difference + bonus components) from the
  // exposed round result, mirroring the domain's scoreRound. `null` for a drawn
  // round (stock exhausted) so the panel stays hidden.
  const scoreBreakdown = useMemo(() => {
    if (!state) return null;
    const atRoundEnd = state.phase === GinRummyPhase.ROUND_END || state.phase === GinRummyPhase.GAME_END;
    if (!atRoundEnd && !state.gameEndFlag) return null;
    return ginRummyScoreBreakdown(state.players, state.knockerIdx, state.knockerDeadwood, state.isGin);
  }, [state]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="ginrummy"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 10 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === GinRummyPhase.DRAW;
  const isDiscardPhase = state.phase === GinRummyPhase.DISCARD;
  const isLayoffPhase = state.phase === GinRummyPhase.LAYOFF;
  const isRoundEnd = state.phase === GinRummyPhase.ROUND_END;
  const isGameEnd = state.phase === GinRummyPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn =
    (isDrawPhase || isDiscardPhase || isLayoffPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.ginrummy')}
      gameThemeBg={gameTheme.ginrummy.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/ginrummy"
      gameEndFlag={isGameEnd}
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
                    value: ginRummyConfig.cpuDifficulty,
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
                    value: ginRummyConfig.pointLimit,
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
              <span>{t('drawPile', { count: state.drawPileCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Discard pile top */}
                {state.discardTop && (
                  <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
                    <AnimatedCard card={state.discardTop} width={cardWidth} />
                    <div className="text-ds-text-muted text-sm">
                      <div>{t('discardTop')}</div>
                    </div>
                  </div>
                )}

                {/* Knocker melds */}
                {state.knockerMelds.length > 0 && (
                  <div className="my-3 p-2 rounded bg-black/30">
                    <div className="text-ds-text-muted text-sm mb-1">{t('knockerMelds')}</div>
                    {state.knockerMelds.map((meld, meldIdx) => (
                      <div key={`meld-${meldIdx}`} className="mb-1">
                        <span
                          data-testid={`gr-meld-badge-${meldIdx}`}
                          className="inline-block rounded border border-ds-secondary px-1.5 py-0.5 text-ds-text-primary text-xs mb-0.5"
                        >
                          {ginRummyMeldLabel(meld.cards)}
                        </span>
                        <div className="flex flex-wrap gap-1">
                          {meld.cards.map((card, cardIdx) => (
                            <AnimatedCard
                              key={`meld-${meldIdx}-${card.design}-${card.value}-${cardIdx}`}
                              card={card}
                              width={cardWidth * 0.7}
                            />
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU player */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })}
                      </div>
                      {/* Show CPU cards during layoff/round end/game end */}
                      {(isLayoffPhase || isRoundEnd || isGameEnd) && p.cards.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {p.cards.map((card, idx) => (
                            <AnimatedCard
                              key={`cpu-${card.design}-${card.value}-${idx}`}
                              card={card}
                              width={cardWidth * 0.8}
                            />
                          ))}
                        </div>
                      )}
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="gr-score-table">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresPlayer')}
                        </th>
                        <th scope="col">{t('scoresRound')}</th>
                        <th scope="col">{t('scoresTotal')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.roundScore}</td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {scoreBreakdown &&
              (() => {
                const winner = state.players.find((p) => p.id === scoreBreakdown.winnerId);
                const winnerLabel = winner ? playerName(winner.id, winner.isHuman) : '';
                const basePart =
                  scoreBreakdown.outcome === 'gin'
                    ? t('breakdown.opponentDeadwoodPart', { value: scoreBreakdown.base })
                    : t('breakdown.differencePart', { value: scoreBreakdown.base });
                const bonusPart =
                  scoreBreakdown.bonus > 0
                    ? scoreBreakdown.outcome === 'gin'
                      ? t('breakdown.ginBonusPart', { value: scoreBreakdown.bonus })
                      : t('breakdown.undercutBonusPart', { value: scoreBreakdown.bonus })
                    : null;
                const formula = `${basePart}${bonusPart ? ` + ${bonusPart}` : ''} = ${scoreBreakdown.total}`;
                return (
                  <div className="my-3 p-3 rounded bg-black/30" data-testid="ginrummy-score-breakdown">
                    <div className="text-ds-text-muted text-sm mb-1">{t('breakdown.title')}</div>
                    <div className="flex items-center gap-2 mb-1 flex-wrap">
                      <span
                        data-testid="ginrummy-breakdown-outcome"
                        className="inline-block rounded border border-ds-secondary px-1.5 py-0.5 text-ds-text-primary text-xs"
                      >
                        {t(`breakdown.outcome.${scoreBreakdown.outcome}`)}
                      </span>
                      <span className="text-ds-accent text-sm font-bold">
                        {t('breakdown.winner', { name: winnerLabel, total: scoreBreakdown.total })}
                      </span>
                    </div>
                    <div className="text-ds-text-muted text-xs mb-1">
                      {t('breakdown.deadwood', {
                        knocker: scoreBreakdown.knockerDeadwood,
                        opponent: scoreBreakdown.opponentDeadwood,
                      })}
                    </div>
                    <div className="text-ds-text-primary text-sm" data-testid="ginrummy-breakdown-formula">
                      {formula}
                    </div>
                  </div>
                );
              })()}

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.ginrummy.footer} px-4 py-2.5`}>
            {liveDeadwood != null && (
              <div
                data-testid="ginrummy-deadwood-indicator"
                className={`text-xs font-bold mb-1 ${canKnockNow ? 'text-ds-success' : 'text-ds-text-muted'}`}
              >
                {t('deadwoodLabel', { score: liveDeadwood, threshold: GIN_RUMMY_KNOCK_THRESHOLD })}
              </div>
            )}
            {isDiscardPhase && (
              <div
                data-testid="ginrummy-meld-legend"
                className="flex items-center gap-3 text-xs text-ds-text-muted mb-1"
              >
                <span className="flex items-center gap-1">
                  <span
                    aria-hidden="true"
                    className="inline-block h-3 w-3 rounded-sm"
                    style={{ outline: '2px solid var(--color-ds-success)', outlineOffset: '-2px' }}
                  />
                  {t('meldLegend')}
                </span>
                <span className="flex items-center gap-1">
                  <span
                    aria-hidden="true"
                    className="inline-block h-3 w-3 rounded-sm"
                    style={{ outline: '2px dashed rgba(148, 163, 184, 0.6)', outlineOffset: '-2px' }}
                  />
                  {t('deadwoodLegend')}
                </span>
              </div>
            )}
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="gr-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    data-testid={`gr-hand-card-${idx}`}
                    data-meld={isDiscardPhase ? (meldedIndices.has(idx) ? 'meld' : 'deadwood') : undefined}
                    className={`transition-transform ${focusRingCard}`}
                    style={{
                      background: 'none',
                      padding: 0,
                      borderRadius: 8,
                      ...(isDiscardPhase ? meldCardStyle(meldedIndices.has(idx)) : undefined),
                      ...selectedCardStyle(selectedCardIndices.includes(idx)),
                      boxSizing: 'border-box',
                    }}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                ))}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center flex-wrap">
              {isDrawPhase && isHumanTurn && (
                <div className="flex gap-2" data-tutorial="gr-draw-area">
                  <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
                    {t('drawStockButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDrawDiscard}
                    disabled={loading || !state.discardTop}
                  >
                    {t('drawDiscardButton')}
                  </button>
                </div>
              )}
              {isDiscardPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDiscard}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="gr-discard-button"
                  >
                    {t('discardButton')}
                  </button>
                  <button
                    type="button"
                    className={`${btnPrimary} ${canKnockNow ? 'motion-safe:animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={handleKnock}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="gr-knock-button"
                    data-testid="ginrummy-knock-button"
                  >
                    {t('knockButton')}
                  </button>
                </>
              )}
              {isLayoffPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleLayoff}
                    disabled={loading || selectedCardIndices.length === 0}
                  >
                    {t('layoffButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleSkipLayoff} disabled={loading}>
                    {t('skipLayoffButton')}
                  </button>
                </>
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
                dataTutorial="gr-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
