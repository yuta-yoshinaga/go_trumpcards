import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { sambaApi } from '../api/gameApi';
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
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useSambaGame } from '../hooks/useSambaGame';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, SambaResponse } from '../types/card';
import { SambaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSambaCommand, SAMBA_HELP } from '../utils/cli/commands/sambaCommands';
import { formatSambaState } from '../utils/cli/formatters/sambaFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { sambaMinMeld, sambaSelectionPoints } from '../utils/sambaScore';
import { hintCheckboxItem } from '../utils/settingsItems';

/**
 * Cards required to complete a canasta (7-card set) or samba (7-card sequence).
 * Matches the Go domain rule (`SambaMeld.IsCanasta`/`IsSamba`: `len(Cards) >= 7`).
 */
const SAMBA_CANASTA_SIZE = 7;

const SAMBA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SambaPhase.DRAW]: 'draw',
  [SambaPhase.MELD]: 'meld',
  [SambaPhase.DISCARD]: 'discard',
  [SambaPhase.ROUND_END]: 'roundEnd',
  [SambaPhase.GAME_END]: 'gameEnd',
};

/** Samba tutorial step definitions. */
const SA_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sa-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sa-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sa-meld-area"]', messageKey: 'tutorial.meldArea', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sa-actions"]', messageKey: 'tutorial.actionButtons', placement: 'top', advanceOn: 'next' },
];

/** Samba game page. */
export const SambaPage = withTutorial(SambaPageContent, 'samba', SA_TUTORIAL_STEPS);
/** Inner content of the Samba page. */
function SambaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('samba');
  const {
    state,
    loading,
    error,
    retry,
    gameExec,
    sambaConfig,
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
  } = useSambaGame();

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('samba', SAMBA_PHASE_KEYS);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('samba', state);

  const humanPlayer = state?.players.find((p) => p.isHuman);
  const humanCardCount = humanPlayer?.cards?.length ?? 0;
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('samba');
  const cliConfig: CliGameConfig<SambaResponse, Parameters<typeof sambaApi.exec>> = useMemo(
    () => ({
      gameName: 'samba',
      parseCommand: parseSambaCommand,
      formatResponse: formatSambaState,
      helpText: SAMBA_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDrawPhase = state?.phase === SambaPhase.DRAW;
  const isMeldPhase = state?.phase === SambaPhase.MELD;
  const isDiscardPhase = state?.phase === SambaPhase.DISCARD;
  const isRoundEnd = state?.phase === SambaPhase.ROUND_END;
  const isGameEnd = state?.phase === SambaPhase.GAME_END || !!state?.gameEndFlag;

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

  // Meld phase: surface the initial-meld minimum (by team score band) and the
  // selected cards' running point total so the player can tell if they qualify.
  const meldPointInfo = useMemo(() => {
    if (!isMeldPhase || !humanPlayer) return null;
    const selectedCards = selectedCardIndices.map((i) => humanPlayer.cards[i]).filter((c): c is Card => Boolean(c));
    const selectedPoints = sambaSelectionPoints(selectedCards);
    const needInitial = !humanPlayer.hasInitMeld;
    const minMeld = sambaMinMeld(humanPlayer.cumulativeScore);
    return { selectedPoints, needInitial, minMeld, below: needInitial && selectedPoints < minMeld };
  }, [isMeldPhase, humanPlayer, selectedCardIndices]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: sambaConfig.cpuDifficulty,
      pointLimit: sambaConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, sambaConfig.cpuDifficulty, sambaConfig.pointLimit]);
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

  // Announce freeze/thaw transitions: freezing changes the draw rules (you now
  // need two matching cards to take the discard pile), so it must reach SR users
  // beyond the colour/badge cue.
  const [frozenMsg, setFrozenMsg] = useState('');
  const prevFrozenRef = useRef<boolean | null>(null);
  useEffect(() => {
    const frozen = state?.isFrozen;
    if (frozen == null) return;
    const prev = prevFrozenRef.current;
    prevFrozenRef.current = frozen;
    if (prev != null && prev !== frozen) {
      setFrozenMsg(frozen ? t('frozenAnnounceOn') : t('frozenAnnounceOff'));
      const id = setTimeout(() => setFrozenMsg(''), 3000);
      return () => clearTimeout(id);
    }
  }, [state?.isFrozen, t]);

  if (!state) {
    return (
      <GameSkeleton
        gameKey="samba"
        layout={{ kind: 'trick-taking', opponents: 3, centerCard: true, trickArea: true, footerHandSize: 13 }}
      />
    );
  }

  return (
    <GamePageShell
      title={tc('nav.samba')}
      gameThemeBg={gameTheme.samba.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/samba"
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
                    value: sambaConfig.cpuDifficulty,
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
                    value: sambaConfig.pointLimit,
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
              {/* Self-contained live region announcing freeze/thaw transitions. */}
              <span className="sr-only" role="status" aria-live="polite" data-testid="sa-frozen-announce">
                {frozenMsg}
              </span>
            </div>
            <div className="text-ds-text-muted text-center mb-2 text-sm" data-testid="sa-team-scores">
              {t('teamScores', { a: state.teamScores[0] ?? 0, b: state.teamScores[1] ?? 0 })}
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
                    data-tutorial="sa-draw-area"
                    data-testid="sa-discard-pile"
                  >
                    <AnimatedCard card={state.discardTop} width={cardWidth} />
                    <div className="text-ds-text-muted text-sm">{t('discardTop')}</div>
                    {state.isFrozen && (
                      <span
                        className="absolute top-1 right-2 text-ds-info text-xs font-bold"
                        data-testid="sa-frozen-badge"
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
                      data-tutorial={pi === 0 ? 'sa-meld-area' : undefined}
                    >
                      <div className="text-ds-text-muted text-sm mb-1">
                        {playerName(p.id, p.isHuman)} ({t('teamLabel', { n: p.team })}) - {t('melds')}
                        {p.hasCanasta && <span className="ml-2 text-ds-warning">★</span>}
                        {p.hasSamba && <span className="ml-1 text-ds-accent">▲</span>}
                      </div>
                      {p.melds.map((m, mi) => {
                        // Progress toward the 7-card canasta (set) / samba (sequence)
                        // milestone, derived purely from `cards.length` and `kind`.
                        const remaining = SAMBA_CANASTA_SIZE - m.cards.length;
                        const toSamba = m.kind === 1;
                        const complete = remaining <= 0;
                        return (
                          <div key={mi} className="flex flex-wrap gap-1 mb-1">
                            <span className="text-xs text-ds-text-muted self-center mr-1">
                              {m.kind === 1
                                ? m.isSamba
                                  ? t('samba')
                                  : t('sequence')
                                : m.isCanasta
                                  ? m.isNatural
                                    ? t('naturalCanasta')
                                    : t('mixedCanasta')
                                  : `(${m.cards.length})`}
                            </span>
                            <span
                              data-testid={`sa-meld-progress-${pi}-${mi}`}
                              className={`text-xs self-center mr-1 ${
                                complete ? 'text-ds-success font-bold motion-safe:animate-pulse' : 'text-ds-info'
                              }`}
                            >
                              {complete
                                ? t(toSamba ? 'meldProgress.sambaComplete' : 'meldProgress.canastaComplete')
                                : t(toSamba ? 'meldProgress.toSamba' : 'meldProgress.toCanasta', { n: remaining })}
                            </span>
                            {m.cards.map((card, ci) => (
                              <AnimatedCard key={`meld-${pi}-${mi}-${ci}`} card={card} width={cardWidth * 0.6} />
                            ))}
                          </div>
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
                        <th scope="col">{t('score.team')}</th>
                        <th scope="col">{t('score.round')}</th>
                        <th scope="col">{t('score.cumulative')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.team}</td>
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
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })}
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

          <GameFooter className={`${gameTheme.samba.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="sa-player-hand">
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

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="sa-actions">
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
                      aria-describedby={drawDiscardReason ? 'sa-draw-discard-reason' : undefined}
                    >
                      {t('drawDiscardButton')}
                    </button>
                  </div>
                  {drawDiscardReason && (
                    <div
                      id="sa-draw-discard-reason"
                      data-testid="sa-draw-discard-reason"
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
                      id="sa-meld-points"
                      data-testid="sa-meld-points"
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
                    aria-describedby={meldPointInfo?.below ? 'sa-meld-points' : undefined}
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
                  <button type="button" className={btnSuccess} onClick={handleGoOut} disabled={loading}>
                    {t('goOutButton')}
                  </button>
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
            <CardNavShortcutsPanel data-testid="samba-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
