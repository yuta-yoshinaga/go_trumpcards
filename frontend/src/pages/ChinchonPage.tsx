import { useCallback, useMemo } from 'react';
import type { chinchonApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import {
  CPU_DIFFICULTY_OPTIONS,
  ELIMINATION_LIMIT_OPTIONS,
  KNOCK_THRESHOLD_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  useChinchonGame,
} from '../hooks/useChinchonGame';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, meldCardStyle, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ChinchonResponse } from '../types/card';
import { ChinchonPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import {
  bestChinchonMeldSplit,
  type ChinchonDeadwoodBreakdown,
  chinchonDeadwoodBreakdown,
  chinchonMeldLabel,
} from '../utils/chinchonDeadwood';
import { CHINCHON_HELP, parseChinchonCommand } from '../utils/cli/commands/chinchonCommands';
import { formatChinchonState } from '../utils/cli/formatters/chinchonFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const CHINCHON_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ChinchonPhase.DRAW]: 'draw',
  [ChinchonPhase.DISCARD]: 'discard',
  [ChinchonPhase.LAYOFF]: 'layoff',
  [ChinchonPhase.ROUND_END]: 'roundEnd',
  [ChinchonPhase.GAME_END]: 'gameEnd',
};

/** Chinchón tutorial step definitions. */
const CH_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ch-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ch-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ch-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ch-knock-button"]',
    messageKey: 'tutorial.knockButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ch-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ch-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Chinchón game page with draw, discard, knock, and layoff phases. */
export const ChinchonPage = withTutorial(ChinchonPageContent, 'chinchon', CH_TUTORIAL_STEPS);
/** Inner content of the Chinchón page, wrapped by TutorialProvider. */
function ChinchonPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('chinchon');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    chinchonConfig,
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
  } = useChinchonGame();
  const { cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('chinchon');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('chinchon', state);
  const cliConfig: CliGameConfig<ChinchonResponse, Parameters<typeof chinchonApi.exec>> = useMemo(
    () => ({
      gameName: 'chinchon',
      parseCommand: parseChinchonCommand,
      formatResponse: formatChinchonState,
      helpText: CHINCHON_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDiscardPhaseForKbd = state?.phase === ChinchonPhase.DISCARD;
  const isLayoffPhaseForKbd = state?.phase === ChinchonPhase.LAYOFF;
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

  const phaseNames = usePhaseNames('chinchon', CHINCHON_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: chinchonConfig.cpuDifficulty,
      playerCount: chinchonConfig.playerCount,
      knockThreshold: chinchonConfig.knockThreshold,
      eliminationLimit: chinchonConfig.eliminationLimit,
    });
  }, [
    gameExec,
    hideActionLog,
    chinchonConfig.cpuDifficulty,
    chinchonConfig.playerCount,
    chinchonConfig.knockThreshold,
    chinchonConfig.eliminationLimit,
  ]);

  // Knock requires the *post-discard* hand deadwood to be ≤ the threshold. When
  // a card is selected we project that discard; otherwise show the best-case
  // deadwood across all single discards (a knockability hint).
  const knockThreshold = state?.config.knockThreshold ?? chinchonConfig.knockThreshold;
  // Breakdown of the deadwood that would remain after the chosen discard (or the
  // best discard when 0/2+ cards are selected), used for both the score and the
  // per-card "5 + 3 + 2 = 10" hint.
  const liveDeadwoodBreakdown = useMemo<ChinchonDeadwoodBreakdown | null>(() => {
    if (!state || state.phase !== ChinchonPhase.DISCARD) return null;
    const human = state.players.find((p) => p.isHuman);
    if (!human || human.cards.length === 0) return null;
    const cards = human.cards;
    if (selectedCardIndices.length === 1) {
      return chinchonDeadwoodBreakdown(cards.filter((_, i) => i !== selectedCardIndices[0]));
    }
    let best: ChinchonDeadwoodBreakdown | null = null;
    for (let i = 0; i < cards.length; i++) {
      const bd = chinchonDeadwoodBreakdown(cards.filter((_, j) => j !== i));
      if (best === null || bd.total < best.total) best = bd;
      if (best.total === 0) break;
    }
    return best;
  }, [state, selectedCardIndices]);
  const liveDeadwood = liveDeadwoodBreakdown?.total ?? null;
  const canKnockNow = liveDeadwood != null && liveDeadwood <= knockThreshold;

  // During DISCARD, color-code the human's hand by the best meld split so the
  // player can see which cards form melds vs. deadwood. When a single discard
  // is selected the split is projected onto the post-discard hand (the chosen
  // card is excluded), so the coloring follows the selection and stays in sync
  // with the deadwood indicator. Returned indices reference the *rendered*
  // hand. Empty set outside DISCARD, so no coloring is applied.
  const meldedIndices = useMemo<ReadonlySet<number>>(() => {
    if (!state || state.phase !== ChinchonPhase.DISCARD) return new Set();
    const human = state.players.find((p) => p.isHuman);
    if (!human || human.cards.length === 0) return new Set();
    const cards = human.cards;
    if (selectedCardIndices.length !== 1) return bestChinchonMeldSplit(cards).meldedIndices;
    const skip = selectedCardIndices[0];
    const subMelded = bestChinchonMeldSplit(cards.filter((_, i) => i !== skip)).meldedIndices;
    // Map post-discard sub-hand indices back to the rendered hand indices.
    const result = new Set<number>();
    let k = 0;
    for (let i = 0; i < cards.length; i++) {
      if (i === skip) continue;
      if (subMelded.has(k)) result.add(i);
      k++;
    }
    return result;
  }, [state, selectedCardIndices]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="chinchon"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 7 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === ChinchonPhase.DRAW;
  const isDiscardPhase = state.phase === ChinchonPhase.DISCARD;
  const isLayoffPhase = state.phase === ChinchonPhase.LAYOFF;
  const isRoundEnd = state.phase === ChinchonPhase.ROUND_END;
  const isGameEnd = state.phase === ChinchonPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn =
    (isDrawPhase || isDiscardPhase || isLayoffPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.chinchon')}
      gameThemeBg={gameTheme.chinchon.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/chinchon"
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
                    value: chinchonConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: chinchonConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'eliminationLimit',
                    label: t('settings.eliminationLimit'),
                    value: chinchonConfig.eliminationLimit,
                    options: ELIMINATION_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('eliminationLimit', v),
                  },
                  {
                    type: 'select',
                    id: 'knockThreshold',
                    label: t('settings.knockThreshold'),
                    value: chinchonConfig.knockThreshold,
                    options: KNOCK_THRESHOLD_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('knockThreshold', v),
                    testId: 'chinchon-knock-threshold-select',
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
                <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

                {/* Discard pile top */}
                {state.discardTop && (
                  <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3" data-tutorial="ch-draw-area">
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
                          data-testid={`ch-meld-badge-${meldIdx}`}
                          className="inline-block rounded border border-ds-secondary px-1.5 py-0.5 text-ds-text-primary text-xs mb-0.5"
                        >
                          {chinchonMeldLabel(meld.cards)}
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
                {/* CPU players */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}
                        {p.eliminated ? ` (${t('eliminated')})` : ''}: {t('cards', { count: p.cardCount })} |{' '}
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
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="ch-score-table">
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
                        <tr
                          key={p.id}
                          className={`${p.isHuman ? 'text-ds-accent' : ''} ${p.eliminated ? 'opacity-50 line-through' : ''}`}
                        >
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

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.chinchon.footer} px-4 py-2.5`}>
            {liveDeadwood != null && (
              <div
                data-testid="chinchon-deadwood-indicator"
                className={`text-xs font-bold mb-1 ${canKnockNow ? 'text-ds-success' : 'text-ds-text-muted'}`}
              >
                {t('deadwoodLabel', { score: liveDeadwood, threshold: knockThreshold })}
              </div>
            )}
            {liveDeadwoodBreakdown != null && liveDeadwoodBreakdown.cards.length > 0 && (
              <div data-testid="chinchon-deadwood-breakdown" className="text-[10px] text-ds-text-muted mb-1">
                {t('deadwoodBreakdown', {
                  breakdown: `${liveDeadwoodBreakdown.values.join(' + ')} = ${liveDeadwoodBreakdown.total}`,
                })}
              </div>
            )}
            {isDiscardPhase && (
              <div
                data-testid="chinchon-meld-legend"
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
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="ch-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    data-testid={`chinchon-hand-card-${idx}`}
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

            <div className="flex gap-2 items-center flex-wrap">
              {isDrawPhase && isHumanTurn && (
                <div className="flex gap-2">
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
                    data-tutorial="ch-discard-button"
                  >
                    {t('discardButton')}
                  </button>
                  <button
                    type="button"
                    className={`${btnPrimary} ${canKnockNow ? 'motion-safe:animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={handleKnock}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="ch-knock-button"
                    data-testid="chinchon-knock-button"
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
                dataTutorial="ch-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="chinchon-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
