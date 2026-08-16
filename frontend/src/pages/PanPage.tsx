import { useCallback, useMemo } from 'react';
import type { panApi } from '../api/gameApi';
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
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, PLAYER_COUNT_OPTIONS, TARGET_ROUNDS_OPTIONS, usePanGame } from '../hooks/usePanGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PanPlayer, PanResponse } from '../types/card';
import { PanPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PAN_HELP, parsePanCommand } from '../utils/cli/commands/panCommands';
import { formatPanState } from '../utils/cli/formatters/panFormatter';
import { hintCliText, isHintCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { isPanValleMeld, panLayoffIndices, panMeldCandidates } from '../utils/panMeldCandidates';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const PAN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PanPhase.DRAW]: 'draw',
  [PanPhase.PLAY]: 'play',
  [PanPhase.ROUND_END]: 'roundEnd',
  [PanPhase.GAME_END]: 'gameEnd',
};

/** Panguingue (Pan) tutorial step definitions. */
const PAN_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="pan-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="pan-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="pan-meld-area"]', messageKey: 'tutorial.meldArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="pan-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pan-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pan-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Panguingue (Pan) game page with draw, play, and layoff phases. */
export const PanPage = withTutorial(PanPageContent, 'pan', PAN_TUTORIAL_STEPS);
/** Inner content of the Panguingue page, wrapped by TutorialProvider. */
function PanPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pan');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    panConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    selectCards,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleMeld,
    handleLayoff,
    handleDiscard,
    handleNextRound,
  } = usePanGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('pan', state);
  const { cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pan');
  const cliConfig: CliGameConfig<PanResponse, Parameters<typeof panApi.exec>> = useMemo(
    () => ({
      gameName: 'pan',
      parseCommand: parsePanCommand,
      formatResponse: formatPanState,
      helpText: PAN_HELP,
      localCommand: (input: string) => (isHintCommand(input) ? hintCliText(frontendHint) : null),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === PanPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isPlayPhaseForKbd) handleDiscard();
  }, [isPlayPhaseForKbd, handleDiscard]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('pan', PAN_PHASE_KEYS);

  // Pre-computed meld-candidate hints for the human's hand (only on their play
  // turn). `candidates` are minimal legal melds to tap-and-select; the two index
  // sets drive the additive ring marks on individual hand cards. Mirrors the
  // domain's meld/layoff rules exactly (see panMeldCandidates.ts).
  const {
    candidates: meldCandidates,
    candidateIndices,
    layoffIndices,
  } = useMemo(() => {
    const empty = {
      candidates: [] as ReturnType<typeof panMeldCandidates>,
      candidateIndices: new Set<number>(),
      layoffIndices: new Set<number>(),
    };
    if (!state || !frontendHintEnabled) return empty;
    const human = state.players.find((p) => p.isHuman);
    const isPlay = state.phase === PanPhase.PLAY;
    const isHuman = state.players[state.currentPlayerIdx]?.isHuman === true;
    if (!human || !isPlay || !isHuman) return empty;
    const candidates = panMeldCandidates(human.cards);
    const candidateIndices = new Set<number>();
    for (const cand of candidates) for (const i of cand.indices) candidateIndices.add(i);
    const tableMelds = state.players.flatMap((p) => p.laidMelds.map((m) => m.cards));
    const layoffIndices = panLayoffIndices(human.cards, tableMelds);
    return { candidates, candidateIndices, layoffIndices };
  }, [state, frontendHintEnabled]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      playerCount: panConfig.playerCount,
      cpuDifficulty: panConfig.cpuDifficulty,
      targetRounds: panConfig.targetRounds,
    });
  }, [gameExec, hideActionLog, panConfig.playerCount, panConfig.cpuDifficulty, panConfig.targetRounds]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="pan"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 10 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === PanPhase.DRAW;
  const isPlayPhase = state.phase === PanPhase.PLAY;
  const isRoundEnd = state.phase === PanPhase.ROUND_END;
  const isGameEnd = state.phase === PanPhase.GAME_END || state.gameEndFlag;
  const revealCpu = isRoundEnd || isGameEnd;
  const isHumanTurn = (isDrawPhase || isPlayPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
  const canLayoff = isPlayPhase && isHumanTurn && selectedCardIndices.length === 1 && !loading;

  // Meld-progress indicator toward the win condition (winMeldCount cards laid on
  // the table), highlighting the final stretch (<=2 cards left) to go out.
  const renderMeldProgress = (p: PanPlayer) => {
    const melded = p.meldedCount;
    const total = state.winMeldCount;
    const remaining = total - melded;
    const isClose = remaining > 0 && remaining <= 2;
    const pct = total > 0 ? Math.min(100, (melded / total) * 100) : 0;
    const label = t('meldProgress', { count: melded, total });
    return (
      <div data-testid="pan-meld-progress" className="mx-auto mb-2 max-w-xs">
        <div className="flex items-center justify-between text-xs mb-0.5">
          <span className="text-ds-text-muted">{label}</span>
          {isClose && <span className="text-ds-warning font-bold">{t('meldRemaining', { count: remaining })}</span>}
        </div>
        <div
          role="progressbar"
          aria-label={label}
          aria-valuemin={0}
          aria-valuemax={total}
          aria-valuenow={melded}
          className="h-1.5 w-full rounded-sm bg-white/15 overflow-hidden"
        >
          <div
            className={`h-full rounded-sm ${isClose ? 'bg-ds-warning' : 'bg-ds-accent'}`}
            style={{ width: `${pct}%` }}
          />
        </div>
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.pan')}
      gameThemeBg={gameTheme.pan.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/pan"
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
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: panConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: panConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: panConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber, total: state.targetRounds })}</span>
              <span className="mr-4">{t('drawPile', { count: state.drawPileCount })}</span>
              <span>{t('winMeld', { count: state.winMeldCount })}</span>
            </div>

            {humanPlayer && renderMeldProgress(humanPlayer)}

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

                {/* Table melds (all players) with layoff targets */}
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="pan-meld-area">
                  <div className="text-ds-text-muted text-sm mb-1">{t('meldsTitle')}</div>
                  {state.players.every((p) => p.laidMelds.length === 0) ? (
                    <div className="text-ds-text-muted text-sm">{t('noMelds')}</div>
                  ) : (
                    state.players
                      .filter((p) => p.laidMelds.length > 0)
                      .map((p) => (
                        <div key={`melds-${p.id}`} className="mb-2">
                          <div className="text-ds-text-muted text-xs mb-1">
                            {t('meldOwner', { name: playerName(p.id, p.isHuman) })}
                          </div>
                          {p.laidMelds.map((meld, meldIdx) => (
                            <div
                              key={`meld-${p.id}-${meldIdx}-${meld.cards.map((c) => `${c.design}${c.value}`).join('')}`}
                              className="flex flex-wrap items-center gap-1 mb-1"
                            >
                              {meld.cards.map((card, idx) => (
                                <AnimatedCard
                                  key={`meldcard-${p.id}-${meldIdx}-${card.design}-${card.value}-${idx}`}
                                  card={card}
                                  width={cardWidth * 0.7}
                                />
                              ))}
                              {/* バジェ (3/5/7 のセット) は全員にチップを配る。どのメルドが
                                  その原因なのかが盤面から読めなかった (#4853)。 */}
                              {isPanValleMeld(meld.cards) && (
                                <span
                                  className={`text-xs font-bold px-1.5 py-0.5 rounded ${badgeWarningColors}`}
                                  data-testid={`pan-valle-${p.id}-${meldIdx}`}
                                >
                                  {t('valleBadge')}
                                </span>
                              )}
                              {canLayoff && (
                                <button
                                  type="button"
                                  className={btnPrimary}
                                  onClick={() => handleLayoff(p.id, meldIdx)}
                                  data-testid={`pan-layoff-${p.id}-${meldIdx}`}
                                >
                                  {t('layoffButton')}
                                </button>
                              )}
                            </div>
                          ))}
                        </div>
                      ))
                  )}
                </div>
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('chips', { count: p.chips })} | {t('melded', { count: p.meldedCount })}
                      </div>
                      {revealCpu && p.cards.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {p.cards.map((card, idx) => (
                            <AnimatedCard
                              key={`cpu-${p.id}-${card.design}-${card.value}-${idx}`}
                              card={card}
                              width={cardWidth * 0.8}
                            />
                          ))}
                        </div>
                      )}
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="pan-score-table">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresPlayer')}
                        </th>
                        <th scope="col">{t('scoresChips')}</th>
                        <th scope="col">{t('scoresRound')}</th>
                        <th scope="col">{t('scoresTotal')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.chips}</td>
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

          <GameFooter className={`${gameTheme.pan.footer} px-4 py-2.5`}>
            {humanPlayer && meldCandidates.length > 0 && (
              <div className="mb-2 p-2 rounded bg-black/30" data-testid="pan-meld-candidates">
                <div className="text-ds-text-muted text-xs mb-1">{t('candidates.title')}</div>
                <div className="flex flex-wrap gap-2">
                  {meldCandidates.map((cand) => {
                    const cards = cand.indices.map((i) => humanPlayer.cards[i]);
                    const kindLabel = t(cand.kind === 'set' ? 'candidates.set' : 'candidates.run');
                    return (
                      <button
                        type="button"
                        key={`cand-${cand.kind}-${cand.indices.join('-')}`}
                        onClick={() => selectCards(cand.indices)}
                        className={`${btnPrimary} flex items-center gap-1`}
                        aria-label={`${kindLabel}: ${cards.map(cardAlt).join(' ')}`}
                        data-testid={`pan-candidate-${cand.indices.join('-')}`}
                      >
                        <span className="text-xs font-bold">{kindLabel}</span>
                        {isPanValleMeld(cards) && (
                          <span
                            className={`text-xs font-bold px-1.5 py-0.5 rounded ${badgeWarningColors}`}
                            data-testid={`pan-candidate-valle-${cand.indices.join('-')}`}
                          >
                            {t('valleBadge')}
                          </span>
                        )}
                        {cards.map((card, ci) => (
                          <AnimatedCard
                            key={`cand-card-${card.design}-${card.value}-${ci}`}
                            card={card}
                            width={cardWidth * 0.55}
                          />
                        ))}
                      </button>
                    );
                  })}
                </div>
                <div className="text-ds-text-muted text-xs mt-1">{t('candidates.hint')}</div>
              </div>
            )}
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="pan-player-hand">
                {humanPlayer.cards.map((card, idx) => {
                  const isLayoff = layoffIndices.has(idx);
                  const isCandidate = candidateIndices.has(idx);
                  // Additive outline ring (stacks on the selection border without
                  // clobbering it): warning = can lay off a single card onto a
                  // table meld; success = part of a legal meld candidate.
                  const hintRing = isLayoff
                    ? { outline: '2px solid var(--color-ds-warning)', outlineOffset: '1px' }
                    : isCandidate
                      ? { outline: '2px solid var(--color-ds-success)', outlineOffset: '1px' }
                      : {};
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => toggleCard(idx)}
                      aria-label={cardAlt(card)}
                      aria-pressed={selectedCardIndices.includes(idx)}
                      data-meld-candidate={isCandidate ? 'true' : undefined}
                      data-layoff-target={isLayoff ? 'true' : undefined}
                      className={`transition-transform ${focusRingCard}`}
                      style={{
                        background: 'none',
                        padding: 0,
                        borderRadius: 8,
                        ...selectedCardStyle(selectedCardIndices.includes(idx)),
                        ...hintRing,
                        boxSizing: 'border-box',
                      }}
                    >
                      <AnimatedCard card={card} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center flex-wrap">
              {isDrawPhase && isHumanTurn && (
                <div className="flex gap-2" data-tutorial="pan-draw-area">
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
              {isPlayPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleMeld}
                    disabled={loading || selectedCardIndices.length < 3}
                    data-testid="pan-meld-button"
                  >
                    {t('meldButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDiscard}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="pan-discard-button"
                    data-testid="pan-discard-button"
                  >
                    {t('discardButton')}
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
                dataTutorial="pan-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="pan-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
