import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { biribaApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useBiribaGame } from '../hooks/useBiribaGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, hintRingStyle, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BiribaResponse, Card } from '../types/card';
import { BiribaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { type BiribaSortMode, loadBiribaSortMode, saveBiribaSortMode, sortedBiribaHand } from '../utils/biribaSort';
import { canastaDrawDiscardProblem } from '../utils/canastaDrawDiscard';
import { cardAlt } from '../utils/cardAlt';
import { BIRIBA_HELP, parseBiribaCommand } from '../utils/cli/commands/biribaCommands';
import { formatBiribaState } from '../utils/cli/formatters/biribaFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const BIRIBA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BiribaPhase.DRAW]: 'draw',
  [BiribaPhase.MELD]: 'meld',
  [BiribaPhase.DISCARD]: 'discard',
  [BiribaPhase.ROUND_END]: 'roundEnd',
  [BiribaPhase.GAME_END]: 'gameEnd',
};

/** Hand sort options for the Biriba footer. */
const BIRIBA_SORT_MODES: { mode: BiribaSortMode; labelKey: string }[] = [
  { mode: 'original', labelKey: 'sort.original' },
  { mode: 'rank', labelKey: 'sort.rank' },
  { mode: 'suit', labelKey: 'sort.suit' },
];

/** Biriba tutorial step definitions. */
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

/** Biriba game page. */
export const BiribaPage = withTutorial(BiribaPageContent, 'biriba', CA_TUTORIAL_STEPS);
/** Inner content of the Biriba page. */
function BiribaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('biriba');
  const {
    state,
    loading,
    error,
    retry,
    gameExec,
    biribaConfig,
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
  } = useBiribaGame();

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const phaseNames = usePhaseNames('biriba', BIRIBA_PHASE_KEYS);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('biriba', state);

  const humanPlayer = state?.players.find((p) => p.isHuman);
  // ヒントが無効なとき・サーバがヒントを返さない場面 (CPU の手番など) では空。
  // **useGameHint が無効時に null を返す**ので、ここで再度フラグを見ない
  // (見ると、条件が二重になって片方が死ぬ)。
  const hintedCards = useMemo(() => new Set(frontendHint?.targetIndices ?? []), [frontendHint?.targetIndices]);
  const humanCardCount = humanPlayer?.cards?.length ?? 0;
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('biriba');
  const cliConfig: CliGameConfig<BiribaResponse, Parameters<typeof biribaApi.exec>> = useMemo(
    () => ({
      gameName: 'biriba',
      parseCommand: parseBiribaCommand,
      formatResponse: formatBiribaState,
      helpText: BIRIBA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDrawPhase = state?.phase === BiribaPhase.DRAW;
  const isMeldPhase = state?.phase === BiribaPhase.MELD;
  const isDiscardPhase = state?.phase === BiribaPhase.DISCARD;
  const isRoundEnd = state?.phase === BiribaPhase.ROUND_END;
  const isGameEnd = state?.phase === BiribaPhase.GAME_END || !!state?.gameEndFlag;

  // Biriba uses the Canasta discard-pile mechanism: two natural cards matching
  // the top card are required to take the pile.
  const drawDiscardProblem = useMemo(() => {
    if (!isDrawPhase || !humanPlayer) return null;
    const selected = selectedCardIndices
      .map((i) => humanPlayer.cards[i])
      .filter((card): card is Card => card !== undefined);
    return canastaDrawDiscardProblem(selected, state?.discardTop);
  }, [isDrawPhase, humanPlayer, selectedCardIndices, state?.discardTop]);

  const drawDiscardReason = useMemo(() => {
    if (!isDrawPhase) return '';
    if (state?.isFrozen && (drawDiscardProblem === 'selectTwo' || drawDiscardProblem === 'selectOneMore')) {
      return t('drawDiscardReason.frozen');
    }
    return drawDiscardProblem === null ? '' : t(`drawDiscardReason.${drawDiscardProblem}`);
  }, [isDrawPhase, state?.isFrozen, drawDiscardProblem, t]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: biribaConfig.cpuDifficulty,
      pointLimit: biribaConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, biribaConfig.cpuDifficulty, biribaConfig.pointLimit]);
  const isHumanTurn =
    (isDrawPhase || isMeldPhase || isDiscardPhase) && state?.players[state.currentPlayerIdx]?.isHuman === true;

  // Transient feedback when a player grabs the pozzetto (a pivotal Biriba moment)
  // and a one-shot pulse on round-score cells that just changed.
  // Display-only hand sort: reorders the rendered hand while every click maps
  // back to the card's original (server-dealt) index, so selection / meld /
  // discard stay index-correct. Persisted across sessions in localStorage.
  const [sortMode, setSortMode] = useState<BiribaSortMode>(loadBiribaSortMode);
  const handleSortMode = useCallback((mode: BiribaSortMode) => {
    setSortMode(mode);
    saveBiribaSortMode(mode);
  }, []);

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

  // useGameApi intentionally does not fetch on mount. Start Biriba explicitly
  // so the page can leave its skeleton state.
  useEffect(() => {
    void gameExec('reset');
  }, [gameExec]);

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
        gameKey="biriba"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 11 }}
      />
    );
  }

  return (
    <GamePageShell
      title={tc('nav.biriba')}
      gameThemeBg={gameTheme.biriba.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/biriba"
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
                    value: biribaConfig.cpuDifficulty,
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
                    value: biribaConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
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

                {/* Full discard pile viewer: in Biriba the whole pile is taken at once,
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
                        {p.hasBiriba && <span className="ml-2 text-ds-warning">★</span>}
                        {p.tookPozzetto && <span className="ml-2 text-ds-info text-xs">[{t('tookPozzetto')}]</span>}
                      </div>
                      {p.melds.map((m, mi) => (
                        <div key={mi} className="flex flex-wrap gap-1 mb-1">
                          <span className="text-xs text-ds-text-muted self-center mr-1">
                            {m.isBiriba ? (m.isNatural ? t('naturalBiriba') : t('mixedBiriba')) : `(${m.cards.length})`}
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
                          {playerName(p.id, p.isHuman)}: {t('handCards', { count: p.cardCount })}
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

          <GameFooter className={`${gameTheme.biriba.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.cards.length > 0 && (
              <fieldset className="flex flex-wrap justify-center gap-1.5 mb-2 border-0 p-0 m-0">
                <legend className="sr-only">{t('sort.label')}</legend>
                {BIRIBA_SORT_MODES.map(({ mode, labelKey }) => (
                  <button
                    key={mode}
                    type="button"
                    onClick={() => handleSortMode(mode)}
                    className={sortMode === mode ? `${btnPrimary} min-w-[64px]` : `${btnSecondary} min-w-[64px]`}
                    aria-pressed={sortMode === mode}
                    data-testid={`bu-sort-${mode}`}
                  >
                    {t(labelKey)}
                  </button>
                ))}
              </fieldset>
            )}
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="ca-player-hand">
                {sortedBiribaHand(humanPlayer.cards, sortMode).map(({ card, index: idx }) => (
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
                      // **どの札を指しているのかを盤面でも言う** (#5990)。
                      // ヒントは「どの札か」まで持っているのに、文言だけ出して
                      // 探させていた。
                      ...(hintedCards.has(idx) ? hintRingStyle() : {}),
                      boxSizing: 'border-box',
                    }}
                    data-testid={`bu-hand-card-${idx}`}
                    data-hint-card={hintedCards.has(idx) ? 'true' : undefined}
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
                      disabled={loading || drawDiscardProblem !== null}
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
            <CardNavShortcutsPanel data-testid="biriba-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
