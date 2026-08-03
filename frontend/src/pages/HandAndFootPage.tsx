import { useCallback, useMemo } from 'react';
import type { handandfootApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useHandAndFootGame } from '../hooks/useHandAndFootGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { HandAndFootResponse } from '../types/card';
import { HandAndFootPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { canastaMinMeld, canastaSelectionPoints } from '../utils/canastaScore';
import { cardAlt } from '../utils/cardAlt';
import { HANDANDFOOT_HELP, parseHandAndFootCommand } from '../utils/cli/commands/handandfootCommands';
import { formatHandAndFootState } from '../utils/cli/formatters/handandfootFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const HANDANDFOOT_PHASE_KEYS: Readonly<Record<number, string>> = {
  [HandAndFootPhase.DRAW]: 'draw',
  [HandAndFootPhase.MELD]: 'meld',
  [HandAndFootPhase.DISCARD]: 'discard',
  [HandAndFootPhase.ROUND_END]: 'roundEnd',
  [HandAndFootPhase.GAME_END]: 'gameEnd',
};

/** Hand and Foot tutorial step definitions. */
const HF_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="hf-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="hf-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="hf-meld-area"]', messageKey: 'tutorial.meldArea', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="hf-actions"]', messageKey: 'tutorial.actionButtons', placement: 'top', advanceOn: 'next' },
];

/** Hand and Foot game page. */
export const HandAndFootPage = withTutorial(HandAndFootPageContent, 'handandfoot', HF_TUTORIAL_STEPS);
/** Inner content of the Hand and Foot page. */
function HandAndFootPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('handandfoot');
  const {
    state,
    loading,
    error,
    retry,
    gameExec,
    handAndFootConfig,
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
  } = useHandAndFootGame();

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('handandfoot', HANDANDFOOT_PHASE_KEYS);

  const humanPlayer = state?.players.find((p) => p.isHuman);
  const humanCardCount = humanPlayer?.cards?.length ?? 0;
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('handandfoot');
  const cliConfig: CliGameConfig<HandAndFootResponse, Parameters<typeof handandfootApi.exec>> = useMemo(
    () => ({
      gameName: 'handandfoot',
      parseCommand: parseHandAndFootCommand,
      formatResponse: formatHandAndFootState,
      helpText: HANDANDFOOT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDrawPhase = state?.phase === HandAndFootPhase.DRAW;
  const isMeldPhase = state?.phase === HandAndFootPhase.MELD;
  const isDiscardPhase = state?.phase === HandAndFootPhase.DISCARD;
  const isRoundEnd = state?.phase === HandAndFootPhase.ROUND_END;
  const isGameEnd = state?.phase === HandAndFootPhase.GAME_END || !!state?.gameEndFlag;

  const drawDiscardReason = useMemo(() => {
    if (!isDrawPhase) return '';
    const n = selectedCardIndices.length;
    if (n > 2) return t('drawDiscardReason.tooMany');
    if (n === 2) return '';
    if (state?.isFrozen) return t('drawDiscardReason.frozen');
    if (n === 1) return t('drawDiscardReason.selectOneMore');
    return t('drawDiscardReason.selectTwo');
  }, [isDrawPhase, selectedCardIndices.length, state?.isFrozen, t]);

  // Whether the human may currently "go out", plus the first unmet requirement.
  // The web build always uses the default go-out rule (>=1 red/natural and >=1
  // black/mixed team canasta, plus the player having entered their foot); see
  // DefaultHandAndFootConfig — the web config never exposes the rc/bc thresholds.
  const goOutGuidance = useMemo(() => {
    if (!isDiscardPhase || !humanPlayer) return null;
    const team = state?.teams.find((tm) => tm.team === humanPlayer.team);
    const canastas = team?.melds.filter((m) => m.isCanasta) ?? [];
    const redCanastas = canastas.filter((m) => m.isNatural).length;
    const blackCanastas = canastas.filter((m) => !m.isNatural).length;
    if (!humanPlayer.inFoot) return { canGoOut: false, reasonKey: 'goOutReason.needFoot' };
    if (redCanastas < 1) return { canGoOut: false, reasonKey: 'goOutReason.needRedCanasta' };
    if (blackCanastas < 1) return { canGoOut: false, reasonKey: 'goOutReason.needBlackCanasta' };
    return { canGoOut: true, reasonKey: 'goOutReason.ready' };
  }, [isDiscardPhase, humanPlayer, state?.teams]);

  // Meld phase: show the selected cards' running point total and, until the
  // team has opened (no melds yet), the initial-meld minimum so the player can
  // tell if the selection qualifies. Point values + minimum bands mirror the
  // shared Canasta-family scoring (the backend uses CanastaCardValue); the web
  // build does not enforce the minimum, so this is a display-only readout.
  const meldPointInfo = useMemo(() => {
    if (!isMeldPhase || !humanPlayer) return null;
    const selectedCards = selectedCardIndices
      .map((i) => humanPlayer.cards[i])
      .filter((c): c is NonNullable<typeof c> => !!c);
    const selectedPoints = canastaSelectionPoints(selectedCards);
    const team = state?.teams.find((tm) => tm.team === humanPlayer.team);
    const needInitial = (team?.melds.length ?? 0) === 0;
    const minMeld = canastaMinMeld(humanPlayer.cumulativeScore);
    const met = selectedPoints >= minMeld;
    return { selectedPoints, needInitial, minMeld, met, below: needInitial && !met };
  }, [isMeldPhase, humanPlayer, selectedCardIndices, state?.teams]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: handAndFootConfig.cpuDifficulty,
      pointLimit: handAndFootConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, handAndFootConfig.cpuDifficulty, handAndFootConfig.pointLimit]);
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

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('handandfoot', state);

  if (!state) {
    return (
      <GameSkeleton
        gameKey="handandfoot"
        layout={{ kind: 'trick-taking', opponents: 3, centerCard: true, trickArea: true, footerHandSize: 11 }}
      />
    );
  }

  return (
    <GamePageShell
      title={tc('nav.handandfoot')}
      gameThemeBg={gameTheme.handandfoot.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/handandfoot"
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
                    value: handAndFootConfig.cpuDifficulty,
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
                    value: handAndFootConfig.pointLimit,
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
                <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

                {/* Discard pile top */}
                {state.discardTop && (
                  <div
                    className={`my-3 p-3 rounded flex items-center gap-3 relative ${
                      state.isFrozen ? 'bg-ds-info/20 ring-2 ring-ds-info' : 'bg-black/40'
                    }`}
                    data-tutorial="hf-draw-area"
                    data-testid="hf-discard-pile"
                  >
                    <AnimatedCard card={state.discardTop} width={cardWidth} />
                    <div className="text-ds-text-muted text-sm">{t('discardTop')}</div>
                    {state.isFrozen && (
                      <span
                        className="absolute top-1 right-2 text-ds-info text-xs font-bold"
                        data-testid="hf-frozen-badge"
                        role="img"
                        aria-label={t('frozenIndicator')}
                      >
                        {t('frozenIndicator')}
                      </span>
                    )}
                  </div>
                )}

                {/* Team melds */}
                {state.teams.map((team, ti) => {
                  if (team.melds.length === 0 && team.red3s.length === 0) return null;
                  return (
                    <div
                      key={team.team}
                      className="my-2 p-2 rounded bg-black/30"
                      data-tutorial={ti === 0 ? 'hf-meld-area' : undefined}
                    >
                      <div className="text-ds-text-muted text-sm mb-1">
                        {t('team', { n: team.team })} - {t('melds')}
                      </div>
                      {team.melds.map((m, mi) => (
                        <div key={mi} className="flex flex-wrap gap-1 mb-1">
                          <span className="text-xs text-ds-text-muted self-center mr-1">
                            {m.isCanasta
                              ? m.isNatural
                                ? t('naturalCanasta')
                                : t('mixedCanasta')
                              : `(${m.cards.length})`}
                          </span>
                          {m.cards.map((card, ci) => (
                            <AnimatedCard key={`meld-${ti}-${mi}-${ci}`} card={card} width={cardWidth * 0.6} />
                          ))}
                        </div>
                      ))}
                      {team.red3s.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          <span className="text-xs text-ds-error self-center mr-1">{t('red3s')}</span>
                          {team.red3s.map((card, ri) => (
                            <AnimatedCard key={`red3-${ti}-${ri}`} card={card} width={cardWidth * 0.6} />
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
                        <th scope="col">{t('teamLabel')}</th>
                        <th scope="col">{t('foot')}</th>
                        <th scope="col">{t('score.round')}</th>
                        <th scope="col">{t('score.cumulative')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.team}</td>
                          <td className="text-center">{p.inFoot ? t('inFoot') : p.footCount}</td>
                          <td className="text-center">{p.roundScore}</td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* CPU hands (shown at round/game end) */}
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

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.handandfoot.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="hf-player-hand">
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

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="hf-actions">
              {isDrawPhase && isHumanTurn && (
                <div className="flex gap-2 flex-col">
                  <div className="flex gap-2">
                    <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
                      {t('drawStockButton')}
                    </button>
                    <button
                      type="button"
                      className={`${btnPrimary} ${
                        state.isFrozen && selectedCardIndices.length === 2 && !loading
                          ? 'motion-safe:animate-pulse ring-2 ring-ds-info'
                          : ''
                      }`}
                      data-frozen={state.isFrozen && selectedCardIndices.length === 2 && !loading ? 'true' : undefined}
                      onClick={handleDrawDiscard}
                      disabled={loading || selectedCardIndices.length !== 2}
                      title={drawDiscardReason || undefined}
                      aria-describedby={drawDiscardReason ? 'hf-draw-discard-reason' : undefined}
                    >
                      {t('drawDiscardButton')}
                    </button>
                  </div>
                  {drawDiscardReason && (
                    <div
                      id="hf-draw-discard-reason"
                      data-testid="hf-draw-discard-reason"
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
                      id="hf-meld-points"
                      data-testid="hf-meld-points"
                      className={`w-full text-xs ${
                        meldPointInfo.below
                          ? 'text-ds-warning'
                          : meldPointInfo.needInitial
                            ? 'text-ds-success'
                            : 'text-ds-text-muted'
                      }`}
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
                    aria-describedby={meldPointInfo?.below ? 'hf-meld-points' : undefined}
                  >
                    {t('meldButton')}
                  </button>
                  <button type="button" className={btnOutline} onClick={handleSkipMeld} disabled={loading}>
                    {t('skipMeldButton')}
                  </button>
                </>
              )}
              {isDiscardPhase && isHumanTurn && (
                <div className="flex gap-2 flex-col">
                  <div className="flex gap-2">
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
                      disabled={loading || goOutGuidance?.canGoOut !== true}
                      title={goOutGuidance && !goOutGuidance.canGoOut ? t(goOutGuidance.reasonKey) : undefined}
                      aria-describedby={goOutGuidance ? 'hf-go-out-guidance' : undefined}
                    >
                      {t('goOutButton')}
                    </button>
                  </div>
                  {goOutGuidance && (
                    <div
                      id="hf-go-out-guidance"
                      data-testid="hf-go-out-guidance"
                      className={`text-xs ${goOutGuidance.canGoOut ? 'text-ds-success' : 'text-ds-text-muted'}`}
                    >
                      {t(goOutGuidance.reasonKey)}
                    </div>
                  )}
                </div>
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
            <CardNavShortcutsPanel data-testid="hand-and-foot-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
