import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { burracoApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useBurracoGame } from '../hooks/useBurracoGame';
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
import type { BurracoResponse } from '../types/card';
import { BurracoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BURRACO_HELP, parseBurracoCommand } from '../utils/cli/commands/burracoCommands';
import { formatBurracoState } from '../utils/cli/formatters/burracoFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

const BURRACO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BurracoPhase.DRAW]: 'draw',
  [BurracoPhase.MELD]: 'meld',
  [BurracoPhase.DISCARD]: 'discard',
  [BurracoPhase.ROUND_END]: 'roundEnd',
  [BurracoPhase.GAME_END]: 'gameEnd',
};

/** Burraco tutorial step definitions. */
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

/** Burraco game page. */
export const BurracoPage = withTutorial(BurracoPageContent, 'burraco', CA_TUTORIAL_STEPS);
/** Inner content of the Burraco page. */
function BurracoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('burraco');
  const {
    state,
    loading,
    error,
    retry,
    gameExec,
    burracoConfig,
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
  } = useBurracoGame();

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const phaseNames = usePhaseNames('burraco', BURRACO_PHASE_KEYS);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('burraco', state);

  const humanPlayer = state?.players.find((p) => p.isHuman);
  const humanCardCount = humanPlayer?.cards?.length ?? 0;
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('burraco');
  const cliConfig: CliGameConfig<BurracoResponse, Parameters<typeof burracoApi.exec>> = useMemo(
    () => ({
      gameName: 'burraco',
      parseCommand: parseBurracoCommand,
      formatResponse: formatBurracoState,
      helpText: BURRACO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDrawPhase = state?.phase === BurracoPhase.DRAW;
  const isMeldPhase = state?.phase === BurracoPhase.MELD;
  const isDiscardPhase = state?.phase === BurracoPhase.DISCARD;
  const isRoundEnd = state?.phase === BurracoPhase.ROUND_END;
  const isGameEnd = state?.phase === BurracoPhase.GAME_END || !!state?.gameEndFlag;

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

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: burracoConfig.cpuDifficulty,
      pointLimit: burracoConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, burracoConfig.cpuDifficulty, burracoConfig.pointLimit]);
  const isHumanTurn =
    (isDrawPhase || isMeldPhase || isDiscardPhase) && state?.players[state.currentPlayerIdx]?.isHuman === true;

  // Transient feedback when a player grabs the pozzetto (a pivotal Burraco moment)
  // and a one-shot pulse on round-score cells that just changed.
  const [pozzettoBanner, setPozzettoBanner] = useState<string | null>(null);
  const [pulsingScoreIds, setPulsingScoreIds] = useState<Set<number>>(new Set());
  const prevPozzettoRef = useRef<boolean[]>([]);
  const prevScoresRef = useRef<number[]>([]);
  const bannerTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pulseTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(
    () => () => {
      clearTimeout(bannerTimerRef.current ?? undefined);
      clearTimeout(pulseTimerRef.current ?? undefined);
    },
    [],
  );

  useEffect(() => {
    if (!state) return;
    const prevPozzetto = prevPozzettoRef.current;
    const prevScores = prevScoresRef.current;
    const taker = state.players.find((p, i) => p.tookPozzetto && prevPozzetto[i] === false);
    const changedScoreIds = state.players
      .filter((p, i) => prevScores[i] !== undefined && prevScores[i] !== p.roundScore)
      .map((p) => p.id);
    prevPozzettoRef.current = state.players.map((p) => p.tookPozzetto);
    prevScoresRef.current = state.players.map((p) => p.roundScore);

    if (taker) {
      setPozzettoBanner(taker.isHuman ? tc('player.you') : tc('player.cpu', { id: taker.id }));
      playSound('chipClick');
      clearTimeout(bannerTimerRef.current ?? undefined);
      bannerTimerRef.current = setTimeout(() => setPozzettoBanner(null), 2000);
    }
    if (changedScoreIds.length > 0) {
      setPulsingScoreIds(new Set(changedScoreIds));
      clearTimeout(pulseTimerRef.current ?? undefined);
      pulseTimerRef.current = setTimeout(() => setPulsingScoreIds(new Set()), 1000);
    }
  }, [state, tc, playSound]);

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
        gameKey="burraco"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 11 }}
      />
    );
  }

  return (
    <GamePageShell
      title={tc('nav.burraco')}
      gameThemeBg={gameTheme.burraco.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/burraco"
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
                    value: burracoConfig.cpuDifficulty,
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
                    value: burracoConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
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
            {pozzettoBanner && (
              <div
                role="status"
                data-testid="bu-pozzetto-banner"
                className="mb-2 text-center text-ds-info font-bold motion-safe:animate-pulse"
              >
                {t('pozzettoTakenBanner', { player: pozzettoBanner })}
              </div>
            )}
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span>
                {t('drawPile', { count: state.drawPileCount })} / {t('discardPile', { count: state.discardPileCount })}
              </span>
              <span className="ml-4" data-testid="bu-pozzetto-count">
                {t('pozzetto', { count: state.pozzettoCount })}
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

                {/* Full discard pile viewer: in Burraco the whole pile is taken at once,
                    so its contents are decision-critical public information. */}
                <details
                  className="my-3 rounded bg-black/30 p-2"
                  data-testid="ca-discard-pile-viewer"
                  open={isDrawPhase}
                >
                  <summary className="cursor-pointer select-none text-sm text-ds-text-muted">
                    {t('discardPileViewer', { count: state.discardPile.length })}
                  </summary>
                  {state.discardPile.length === 0 ? (
                    <div className="mt-2 text-xs text-ds-text-muted">{t('discardPileEmpty')}</div>
                  ) : (
                    <>
                      <div className="mt-1 text-xs text-ds-text-muted">{t('discardPileOrderHint')}</div>
                      <div className="mt-2 flex flex-wrap gap-1" data-testid="ca-discard-pile-cards">
                        {state.discardPile.map((card, di) => (
                          <AnimatedCard
                            key={`discard-${card.design}-${card.value}-${di}`}
                            card={card}
                            width={cardWidth * 0.6}
                          />
                        ))}
                      </div>
                    </>
                  )}
                </details>

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
                        {p.hasBurraco && <span className="ml-2 text-ds-warning">★</span>}
                        {p.tookPozzetto && <span className="ml-2 text-ds-info text-xs">[{t('tookPozzetto')}]</span>}
                      </div>
                      {p.melds.map((m, mi) => (
                        <div key={mi} className="flex flex-wrap gap-1 mb-1">
                          <span className="text-xs text-ds-text-muted self-center mr-1">
                            {m.isBurraco
                              ? m.isNatural
                                ? t('naturalBurraco')
                                : t('mixedBurraco')
                              : `(${m.cards.length})`}
                          </span>
                          {m.cards.map((card, ci) => (
                            <AnimatedCard key={`meld-${pi}-${mi}-${ci}`} card={card} width={cardWidth * 0.6} />
                          ))}
                        </div>
                      ))}
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
                          <td
                            className={`text-center ${pulsingScoreIds.has(p.id) ? 'motion-safe:animate-pulse text-ds-info' : ''}`}
                            data-testid={`bu-round-score-${p.id.toString()}`}
                          >
                            {p.roundScore}
                          </td>
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

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.burraco.footer} px-4 py-2.5`}>
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
                  {state.isFrozen && (
                    <div role="status" data-testid="ca-draw-freeze-guide" className="text-xs text-ds-warning">
                      {t('drawFreezeGuide')}
                    </div>
                  )}
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
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleMeldSelected}
                    disabled={loading || selectedCardIndices.length < 3}
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
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
