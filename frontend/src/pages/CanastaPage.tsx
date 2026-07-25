import { useCallback, useMemo } from 'react';
import type { canastaApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useCanastaGame } from '../hooks/useCanastaGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CanastaResponse, Card } from '../types/card';
import { CanastaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { canastaMinMeld, canastaSelectionPoints } from '../utils/canastaScore';
import { cardAlt } from '../utils/cardAlt';
import { CANASTA_HELP, parseCanastaCommand } from '../utils/cli/commands/canastaCommands';
import { formatCanastaState } from '../utils/cli/formatters/canastaFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const CANASTA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CanastaPhase.DRAW]: 'draw',
  [CanastaPhase.MELD]: 'meld',
  [CanastaPhase.DISCARD]: 'discard',
  [CanastaPhase.ROUND_END]: 'roundEnd',
  [CanastaPhase.GAME_END]: 'gameEnd',
};

/** Canasta tutorial step definitions. */
const CA_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ca-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ca-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ca-meld-area"]', messageKey: 'tutorial.meldArea', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ca-actions"]', messageKey: 'tutorial.actionButtons', placement: 'top', advanceOn: 'next' },
];

/** Canasta game page. */
export const CanastaPage = withTutorial(CanastaPageContent, 'canasta', CA_TUTORIAL_STEPS);
/** Inner content of the Canasta page. */
function CanastaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('canasta');
  const {
    state,
    loading,
    error,
    retry,
    gameExec,
    canastaConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleDrawStock,
    handleDrawDiscard,
    handleMeldSelected,
    handleSkipMeld,
    handleDiscard,
    handleGoOut,
    handleNextRound,
  } = useCanastaGame();

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const phaseNames = usePhaseNames('canasta', CANASTA_PHASE_KEYS);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('canasta', state);

  const humanPlayer = state?.players.find((p) => p.isHuman);
  const humanCardCount = humanPlayer?.cards?.length ?? 0;
  // The server rejects a go-out unless the player holds at least one completed
  // canasta (Canasta.PlayerGoOut -> HasCanasta()); mirror that guard client-side.
  const canGoOut = humanPlayer?.hasCanasta === true;
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('canasta');
  const cliConfig: CliGameConfig<CanastaResponse, Parameters<typeof canastaApi.exec>> = useMemo(
    () => ({
      gameName: 'canasta',
      parseCommand: parseCanastaCommand,
      formatResponse: formatCanastaState,
      helpText: CANASTA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDrawPhase = state?.phase === CanastaPhase.DRAW;
  const isMeldPhase = state?.phase === CanastaPhase.MELD;
  const isDiscardPhase = state?.phase === CanastaPhase.DISCARD;
  const isRoundEnd = state?.phase === CanastaPhase.ROUND_END;
  const isGameEnd = state?.phase === CanastaPhase.GAME_END || !!state?.gameEndFlag;

  const drawDiscardReason = useMemo(() => {
    if (!isDrawPhase) return '';
    const n = selectedCardIndices.length;
    if (n > 2) return t('drawDiscardReason.tooMany');
    if (n === 2) return '';
    // Frozen takes priority while the player is still picking — the wildcard restriction
    // is the load-bearing rule players forget; surface it whether they've picked 0 or 1 cards.
    if (state?.isFrozen) return t('drawDiscardReason.frozen');
    if (n === 1) return t('drawDiscardReason.selectOneMore');
    return t('drawDiscardReason.selectTwo');
  }, [isDrawPhase, selectedCardIndices.length, state?.isFrozen, t]);

  // Meld phase: surface the initial-meld minimum (by score band) and the
  // selected cards' running point total so the player can tell if they qualify.
  const meldPointInfo = useMemo(() => {
    if (!isMeldPhase || !humanPlayer) return null;
    const selectedCards = selectedCardIndices.map((i) => humanPlayer.cards[i]).filter((c): c is Card => Boolean(c));
    const selectedPoints = canastaSelectionPoints(selectedCards);
    const needInitial = !humanPlayer.hasInitMeld;
    const minMeld = canastaMinMeld(humanPlayer.cumulativeScore);
    return { selectedPoints, needInitial, minMeld, below: needInitial && selectedPoints < minMeld };
  }, [isMeldPhase, humanPlayer, selectedCardIndices]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: canastaConfig.cpuDifficulty,
      pointLimit: canastaConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, canastaConfig.cpuDifficulty, canastaConfig.pointLimit]);
  const isHumanTurn =
    (isDrawPhase || isMeldPhase || isDiscardPhase) && state?.players[state.currentPlayerIdx]?.isHuman === true;

  const kbdConfirmAction = useCallback(() => {
    if (isDiscardPhase) handleDiscard();
    else if (isMeldPhase) handleMeldSelected();
  }, [isDiscardPhase, isMeldPhase, handleDiscard, handleMeldSelected]);

  useCardKeyboardNav({
    cardCount: humanCardCount,
    onToggle: toggleCard,
    onConfirm: kbdConfirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurn && !loading,
  });

  if (!state) {
    return (
      <GameSkeleton
        gameKey="canasta"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 11 }}
      />
    );
  }

  return (
    <GamePageShell
      title={tc('nav.canasta')}
      gameThemeBg={gameTheme.canasta.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/canasta"
      gameEndFlag={isGameEnd}
      onCelebrate={() => playSound('winFanfare')}
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
                    value: canastaConfig.cpuDifficulty,
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
                    value: canastaConfig.pointLimit,
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
              <span>
                {t('drawPile', { count: state.drawPileCount })} / {t('discardPile', { count: state.discardPileCount })}
              </span>
              {state.isFrozen && <span className="ml-2 text-ds-info font-bold">[{t('frozen')}]</span>}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Discard pile top */}
                {state.discardTop && (
                  <div
                    className={`my-3 p-3 rounded flex items-center gap-3 relative ${
                      state.isFrozen ? 'bg-ds-info/20 ring-2 ring-ds-info' : 'bg-black/40'
                    }`}
                    data-tutorial="ca-draw-area"
                    data-testid="ca-discard-pile"
                  >
                    <AnimatedCard card={state.discardTop} width={cardWidth} />
                    <div className="text-ds-text-muted text-sm">{t('discardTop')}</div>
                    {state.isFrozen && (
                      <span
                        className="absolute top-1 right-2 text-ds-info text-xs font-bold"
                        data-testid="ca-frozen-badge"
                        role="img"
                        aria-label={t('frozenIndicator')}
                      >
                        {t('frozenIndicator')}
                      </span>
                    )}
                  </div>
                )}

                {/* Player melds */}
                {state.players.map((p, pi) => {
                  if (p.melds.length === 0 && p.red3s.length === 0) return null;
                  return (
                    <div
                      key={pi}
                      className="my-2 p-2 rounded bg-black/30"
                      data-tutorial={pi === 0 ? 'ca-meld-area' : undefined}
                    >
                      <div className="text-ds-text-muted text-sm mb-1">
                        {playerName(p.id, p.isHuman)} - {t('melds')}
                        {p.hasCanasta && <span className="ml-2 text-ds-warning">★</span>}
                      </div>
                      {p.melds.map((m, mi) => {
                        // Same-rank cards are collapsed into an overlapping stack so a meld
                        // occupies ~one card of height; the <details> expands to the full spread.
                        const meldW = cardWidth * 0.6;
                        const canastaLabel = m.isCanasta
                          ? m.isNatural
                            ? t('naturalCanasta')
                            : t('mixedCanasta')
                          : null;
                        return (
                          <details key={mi} className="mb-1" data-testid={`ca-meld-${pi}-${mi}`}>
                            <summary
                              className="flex items-center gap-2 cursor-pointer list-none marker:hidden"
                              title={t('meldExpand')}
                            >
                              <div className="relative flex items-center shrink-0">
                                {m.cards.map((card, ci) => (
                                  <div
                                    key={`meld-stack-${pi}-${mi}-${ci}`}
                                    style={{ marginLeft: ci === 0 ? 0 : -meldW * 0.72, zIndex: ci }}
                                  >
                                    <AnimatedCard card={card} width={meldW} silent />
                                  </div>
                                ))}
                              </div>
                              <span
                                className="text-xs font-bold px-1.5 py-0.5 rounded-full bg-black/50 text-ds-text-primary"
                                data-testid={`ca-meld-badge-${pi}-${mi}`}
                              >
                                {m.cards.length}
                              </span>
                              {canastaLabel && (
                                <span className="text-xs font-bold text-ds-warning">★ {canastaLabel}</span>
                              )}
                            </summary>
                            <div className="flex flex-wrap gap-1 mt-1 pl-2">
                              {m.cards.map((card, ci) => (
                                <AnimatedCard key={`meld-${pi}-${mi}-${ci}`} card={card} width={meldW} silent />
                              ))}
                            </div>
                          </details>
                        );
                      })}
                      {p.red3s.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          <span className="text-xs text-ds-error self-center mr-1">{t('red3s')}</span>
                          {p.red3s.map((card, ri) => (
                            <AnimatedCard key={`red3-${pi}-${ri}`} card={card} width={cardWidth * 0.6} />
                          ))}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30">
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {tc('label.player')}
                        </th>
                        <th scope="col">{t('score.round')}</th>
                        <th scope="col">{t('score.cumulative')}</th>
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

                {/* CPU hand (shown at round/game end) */}
                {(isRoundEnd || isGameEnd) &&
                  state.players
                    .filter((p) => !p.isHuman)
                    .map((p) => (
                      <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                        <div className="text-ds-text-muted text-sm">
                          {playerName(p.id, p.isHuman)}: {p.cardCount} cards
                        </div>
                        {p.cards.length > 0 && (
                          <div className="flex flex-wrap gap-1 mt-1">
                            {p.cards.map((card, idx) => (
                              <AnimatedCard
                                key={`cpu-${card.design}-${card.value}-${idx}`}
                                card={card}
                                width={cardWidth * 0.7}
                              />
                            ))}
                          </div>
                        )}
                      </div>
                    ))}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.canasta.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="ca-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    className={`transition-transform ${focusRingCard}`}
                    style={{
                      background: 'none',
                      padding: 0,
                      borderRadius: 8,
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

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="ca-actions">
              {isDrawPhase && isHumanTurn && (
                <div className="flex gap-2 flex-col">
                  <div className="flex gap-2">
                    <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
                      {t('drawStockButton')}
                    </button>
                    <button
                      type="button"
                      className={btnPrimary}
                      onClick={handleDrawDiscard}
                      disabled={loading || selectedCardIndices.length !== 2}
                      title={drawDiscardReason || undefined}
                      aria-describedby={drawDiscardReason ? 'ca-draw-discard-reason' : undefined}
                    >
                      {t('drawDiscardButton')}
                    </button>
                  </div>
                  {drawDiscardReason && (
                    <div
                      id="ca-draw-discard-reason"
                      data-testid="ca-draw-discard-reason"
                      className="text-xs text-ds-text-muted"
                    >
                      {drawDiscardReason}
                    </div>
                  )}
                </div>
              )}
              {isMeldPhase && isHumanTurn && (
                <>
                  {meldPointInfo && (
                    <div
                      id="ca-meld-points"
                      data-testid="ca-meld-points"
                      className={`w-full text-xs ${meldPointInfo.below ? 'text-ds-warning' : 'text-ds-text-muted'}`}
                    >
                      {meldPointInfo.needInitial
                        ? t('meldPoints.initial', {
                            min: meldPointInfo.minMeld,
                            points: meldPointInfo.selectedPoints,
                          })
                        : t('meldPoints.selected', { points: meldPointInfo.selectedPoints })}
                    </div>
                  )}
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleMeldSelected}
                    disabled={loading || selectedCardIndices.length < 3}
                    aria-describedby={meldPointInfo?.below ? 'ca-meld-points' : undefined}
                  >
                    {t('meldButton')}
                  </button>
                  <button type="button" className={btnOutline} onClick={handleSkipMeld} disabled={loading}>
                    {t('skipMeldButton')}
                  </button>
                </>
              )}
              {isDiscardPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDiscard}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('discardButton')}
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleGoOut}
                    disabled={loading || !canGoOut}
                    title={canGoOut ? undefined : t('goOutReason')}
                    aria-describedby={canGoOut ? undefined : 'ca-go-out-reason'}
                    data-testid="ca-go-out-button"
                  >
                    {t('goOutButton')}
                  </button>
                  {!canGoOut && (
                    <div
                      id="ca-go-out-reason"
                      data-testid="ca-go-out-reason"
                      className="w-full text-xs text-ds-text-muted"
                    >
                      {t('goOutReason')}
                    </div>
                  )}
                </>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
