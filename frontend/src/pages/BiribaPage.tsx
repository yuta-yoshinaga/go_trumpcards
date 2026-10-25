import type { TFunction } from 'i18next';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { biribaApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useBiribaGame } from '../hooks/useBiribaGame';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { badgeInfoColors } from '../styles/badgeStyles';
import { btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, hintRingStyle, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { BiribaResponse, Card } from '../types/card';
import { BiribaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { evaluateBiribaMeld } from '../utils/biribaMeld';
import { type BiribaSortMode, loadBiribaSortMode, saveBiribaSortMode } from '../utils/biribaSort';
import { canastaDrawDiscardProblem } from '../utils/canastaDrawDiscard';
import { cardAlt } from '../utils/cardAlt';
import { BIRIBA_HELP, parseBiribaCommand } from '../utils/cli/commands/biribaCommands';
import { formatBiribaState } from '../utils/cli/formatters/biribaFormatter';
import { sortedHandForDisplay } from '../utils/handDisplaySort';
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

function useGame() {
  const game = useBiribaGame();
  return {
    ...game,
    exec: game.gameExec,
  };
}

type BiribaGame = ReturnType<typeof useGame>;
type BiribaView = ReturnType<typeof useBiribaView>;
type BiribaContext = GamePageContext<BiribaResponse, BiribaGame, BiribaView>;

function useBiribaView({ game, tc }: { game: BiribaGame; t: TFunction; tc: TFunction }) {
  const { state } = game;
  const { cardWidth, isMobile } = useCardDimensions();
  const isMobileViewport = useIsMobile();
  const { playSound } = useSound();
  const phaseNames = usePhaseNames('biriba', BIRIBA_PHASE_KEYS);

  // useGameApi intentionally does not fetch on mount. Start Biriba explicitly
  // so the page can leave its skeleton state.
  useMountReset(game.gameExec);

  const [discardPileViewerOpen, setDiscardPileViewerOpen] = useState(!isMobileViewport);
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

  return {
    cardWidth,
    isMobile,
    isMobileViewport,
    phaseNames,
    discardPileViewerOpen,
    setDiscardPileViewerOpen,
    sortMode,
    handleSortMode,
    pozzettoBanner,
    pulsingScoreIds,
  };
}

function Board({ state, t, tc, view, frontendHint }: BiribaContext) {
  const { cardWidth, isMobile, discardPileViewerOpen, setDiscardPileViewerOpen, pozzettoBanner, pulsingScoreIds } =
    view;
  const isRoundEnd = state.phase === BiribaPhase.ROUND_END;
  const isGameEnd = state.phase === BiribaPhase.GAME_END || !!state.gameEndFlag;

  return (
    <>
      {pozzettoBanner && (
        <div
          role="status"
          data-testid="bu-pozzetto-banner"
          className="mb-2 text-center text-ds-info font-bold motion-safe:animate-pulse"
        >
          {t('pozzettoTakenBanner', { player: pozzettoBanner })}
        </div>
      )}
      <div className="flex flex-wrap justify-center text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span>
          {t('drawPile', { count: state.drawPileCount })} / {t('discardPile', { count: state.discardPileCount })}
        </span>
        <span className="ml-4" data-testid="bu-pozzetto-count">
          {t('pozzetto', { count: state.pozzettoCount })}
        </span>
        {state.isFrozen && (
          <span className="ml-2 shrink-0 whitespace-nowrap text-ds-info font-bold">[{t('frozen')}]</span>
        )}
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          {/* Discard pile top */}
          {state.discardTop && (
            <div
              className={`my-3 p-3 rounded flex items-center gap-3 relative ${
                state.isFrozen ? `${badgeInfoColors} ring-2 ring-ds-info` : 'bg-black/40'
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
            open={discardPileViewerOpen || undefined}
            onToggle={(event) => setDiscardPileViewerOpen(event.currentTarget.open)}
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
                    <span className="text-xs text-ds-error-text self-center mr-1">{t('red3s')}</span>
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
          <details className="my-3 p-2 rounded bg-black/30" open={!view.isMobileViewport || undefined}>
            <summary className="cursor-pointer select-none text-sm text-ds-text-muted">{t('score.round')}</summary>
            <div className="mt-2">
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
                        {isRoundEnd && (
                          <ul className="mt-2 space-y-1 text-left text-xs text-ds-text-muted">
                            <li>{t('score.breakdown')}</li>
                            <li>
                              {t('score.meldCards')}: {p.scoreBreakdown.meldCards || t('score.noPoints')}
                            </li>
                            <li>
                              {t('score.biribaBonus')}: {p.scoreBreakdown.canastaBonus || t('score.noPoints')}
                            </li>
                            <li>
                              {t('score.red3Bonus')}: {p.scoreBreakdown.red3Bonus || t('score.noPoints')}
                            </li>
                            <li>
                              {t('score.goingOutBonus')}: {p.scoreBreakdown.goOutBonus || t('score.noPoints')}
                            </li>
                            <li>
                              {t('score.handPenalty')}:{' '}
                              {p.scoreBreakdown.handPenalty ? `−${p.scoreBreakdown.handPenalty}` : t('score.noPoints')}
                            </li>
                          </ul>
                        )}
                      </td>
                      <td className="text-center">{p.cumulativeScore}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </details>

          {/* CPU hand (shown at round/game end) */}
          {(isRoundEnd || isGameEnd) && (
            <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
              <div className="flex flex-wrap gap-2">
                {state.players
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
            </CpuAccordion>
          )}
        </div>
      </div>

      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />

      {isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}
    </>
  );
}

function FooterLead({ state, t, view }: BiribaContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer || humanPlayer.cards.length === 0) return null;

  return (
    <fieldset className="flex flex-wrap justify-center gap-1.5 mb-2 border-0 p-0 m-0">
      <legend className="sr-only">{t('sort.label')}</legend>
      {BIRIBA_SORT_MODES.map(({ mode, labelKey }) => (
        <button
          key={mode}
          type="button"
          onClick={() => view.handleSortMode(mode)}
          className={view.sortMode === mode ? `${btnPrimary} min-w-[64px]` : `${btnSecondary} min-w-[64px]`}
          aria-pressed={view.sortMode === mode}
          data-testid={`bu-sort-${mode}`}
        >
          {t(labelKey)}
        </button>
      ))}
    </fieldset>
  );
}

function FooterHand({ game, state, t, view, frontendHint }: BiribaContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === BiribaPhase.DRAW;
  const isMeldPhase = state.phase === BiribaPhase.MELD;
  const isDiscardPhase = state.phase === BiribaPhase.DISCARD;
  const isHumanTurn =
    (isDrawPhase || isMeldPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  // ヒントが無効なとき・サーバがヒントを返さない場面 (CPU の手番など) では空。
  // **useGameHint が無効時に null を返す**ので、ここで再度フラグを見ない
  // (見ると、条件が二重になって片方が死ぬ)。
  const hintedCards = useMemo(
    () => new Set(frontendHint.hint?.targetIndices ?? []),
    [frontendHint.hint?.targetIndices],
  );

  const goOutMissingRequirements = [
    ...(humanPlayer?.tookPozzetto ? [] : [t('goOutRequirementPozzetto')]),
    ...(humanPlayer?.hasBiriba ? [] : [t('goOutRequirementBiriba')]),
  ];
  const goOutReason = goOutMissingRequirements.length
    ? t('goOutReason', { requirements: goOutMissingRequirements.join(t('listSeparator')) })
    : '';

  return (
    <>
      {humanPlayer && (
        <div className="flex flex-wrap gap-1 mb-2" data-tutorial="ca-player-hand">
          {sortedHandForDisplay(humanPlayer.cards, view.sortMode).map(({ card, index: idx }) => (
            <button
              type="button"
              key={`${card.design}-${card.value}-${idx}`}
              onClick={() => game.toggleCard(idx)}
              aria-label={cardAlt(card)}
              aria-pressed={game.selectedCardIndices.includes(idx)}
              className={`transition-transform ${focusRingCard}`}
              style={{
                background: 'none',
                padding: 0,
                borderRadius: 8,
                ...selectedCardStyle(game.selectedCardIndices.includes(idx)),
                // **どの札を指しているのかを盤面でも言う** (#5990)。
                // ヒントは「どの札か」まで持っているのに、文言だけ出して
                // 探させていた。
                ...(hintedCards.has(idx) ? hintRingStyle() : {}),
                boxSizing: 'border-box',
              }}
              data-testid={`bu-hand-card-${idx}`}
              data-hint-card={hintedCards.has(idx) ? 'true' : undefined}
            >
              <AnimatedCard card={card} width={view.cardWidth} />
            </button>
          ))}
        </div>
      )}
      <div id="biriba-go-out-reason" data-testid="biriba-go-out-reason" role="status" aria-live="polite">
        {isDiscardPhase && isHumanTurn && goOutReason && <span className="sr-only">{goOutReason}</span>}
      </div>
    </>
  );
}

function primaryActions({ game, state, t, busy }: BiribaContext) {
  const loading = busy;
  const isDrawPhase = state.phase === BiribaPhase.DRAW;
  const isMeldPhase = state.phase === BiribaPhase.MELD;
  const isDiscardPhase = state.phase === BiribaPhase.DISCARD;
  const isRoundEnd = state.phase === BiribaPhase.ROUND_END;
  const isHumanTurn =
    (isDrawPhase || isMeldPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const selectedMeldCards = humanPlayer?.cards.filter((_, index) => game.selectedCardIndices.includes(index)) ?? [];
  const meldStatus = humanPlayer
    ? evaluateBiribaMeld(selectedMeldCards, humanPlayer.melds, {
        hasInitMeld: humanPlayer.hasInitMeld,
        minMeld: state.minMeld,
        drewFromDiscard: state.drewFromDiscard,
        includesDrawnCard: game.selectedCardIndices.includes(state.drawnCardIndex),
      })
    : { ok: false as const, reason: 'selectCards' as const };
  const meldReason = meldStatus.reason ? t(`meldReason.${meldStatus.reason}`, { min: state.minMeld }) : '';

  // Biriba uses the Canasta discard-pile mechanism: two natural cards matching
  // the top card are required to take the pile.
  const drawDiscardProblem = (() => {
    if (!isDrawPhase || !humanPlayer) return null;
    const selected = game.selectedCardIndices
      .map((i) => humanPlayer.cards[i])
      .filter((card): card is Card => card !== undefined);
    return canastaDrawDiscardProblem(selected, state.discardTop);
  })();

  const drawDiscardReason = (() => {
    if (!isDrawPhase) return '';
    if (state.isFrozen && (drawDiscardProblem === 'selectTwo' || drawDiscardProblem === 'selectOneMore')) {
      return t('drawDiscardReason.frozen');
    }
    return drawDiscardProblem === null ? '' : t(`drawDiscardReason.${drawDiscardProblem}`);
  })();

  const goOutMissingRequirements = [
    ...(humanPlayer?.tookPozzetto ? [] : [t('goOutRequirementPozzetto')]),
    ...(humanPlayer?.hasBiriba ? [] : [t('goOutRequirementBiriba')]),
  ];
  const goOutReason = goOutMissingRequirements.length
    ? t('goOutReason', { requirements: goOutMissingRequirements.join(t('listSeparator')) })
    : '';

  return (
    <div className="flex gap-2 items-center flex-wrap" data-tutorial="ca-actions">
      {isDrawPhase && isHumanTurn && (
        <div className="flex gap-2 flex-col">
          {state.isFrozen && (
            <div role="status" data-testid="ca-draw-freeze-guide" className="text-xs text-ds-warning">
              {t('drawFreezeGuide')}
            </div>
          )}
          <div className="flex gap-2">
            <button type="button" className={btnPrimary} onClick={game.handleDrawStock} disabled={loading}>
              {t('drawStockButton')}
            </button>
            <button
              type="button"
              className={btnPrimary}
              onClick={game.handleDrawDiscard}
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
          <div role="status" aria-live="polite" data-testid="biriba-meld-status" className="text-xs text-ds-text-muted">
            {meldStatus.ok ? t('meldStatus.valid') : meldReason}
          </div>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => {
              if (!loading && meldStatus.ok) game.handleMeldSelected();
            }}
            aria-disabled={!meldStatus.ok}
            aria-describedby={meldReason ? 'biriba-meld-reason' : undefined}
          >
            {t('meldButton')}
          </button>
          {meldReason && (
            <span id="biriba-meld-reason" className="sr-only">
              {meldReason}
            </span>
          )}
          <button type="button" className={btnOutline} onClick={game.handleSkipMeld} disabled={loading}>
            {t('skipMeldButton')}
          </button>
        </>
      )}
      {isDiscardPhase && isHumanTurn && (
        <>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleDiscard}
            disabled={loading || game.selectedCardIndices.length !== 1}
          >
            {t('discardButton')}
          </button>
          <button
            type="button"
            className={`${btnSuccess} ${goOutReason || loading ? 'aria-disabled:opacity-70 aria-disabled:cursor-not-allowed aria-disabled:saturate-50' : ''}`}
            onClick={() => {
              if (!loading && !goOutReason) game.handleGoOut();
            }}
            aria-disabled={!!goOutReason || loading || undefined}
            aria-describedby={goOutReason ? 'biriba-go-out-reason-visible' : undefined}
          >
            {t('goOutButton')}
          </button>
          {goOutReason && (
            <span
              id="biriba-go-out-reason-visible"
              data-testid="biriba-go-out-reason-visible"
              className="text-xs text-ds-text-muted"
              aria-hidden="true"
            >
              {goOutReason}
            </span>
          )}
        </>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={loading}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const biribaDefinition: GameDefinition<BiribaResponse, BiribaGame, Parameters<typeof biribaApi.exec>, BiribaView> = {
  key: 'biriba',
  useGame,
  useView: useBiribaView,
  tutorial: CA_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 11 },
  cli: () => ({
    parseCommand: parseBiribaCommand,
    formatResponse: formatBiribaState,
    helpText: BIRIBA_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === BiribaPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isDrawPhase = state.phase === BiribaPhase.DRAW;
    const isMeldPhase = state.phase === BiribaPhase.MELD;
    const isDiscardPhase = state.phase === BiribaPhase.DISCARD;
    const isHumanTurn =
      (isDrawPhase || isMeldPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      isHumanTurn,
    };
  },
  reset: ({ game }) => {
    void game.gameExec('reset', undefined, {
      cpuDifficulty: game.biribaConfig.cpuDifficulty,
      pointLimit: game.biribaConfig.pointLimit,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.biribaConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'pointLimit',
          label: t('settings.pointLimit'),
          value: game.biribaConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('pointLimit', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  message: ({ state, view }) =>
    view.isMobile ? undefined : (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ),
  cardNav: ({ game, state }) => {
    const isDrawPhase = state.phase === BiribaPhase.DRAW;
    const isMeldPhase = state.phase === BiribaPhase.MELD;
    const isDiscardPhase = state.phase === BiribaPhase.DISCARD;
    const isHumanTurn =
      (isDrawPhase || isMeldPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
    const humanPlayer = state.players.find((p) => p.isHuman);

    return {
      cardCount: humanPlayer?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: () => {
        if (isDiscardPhase) game.handleDiscard();
        else if (isMeldPhase) game.handleMeldSelected();
      },
      onClear: game.clearSelection,
      enabled: isHumanTurn,
    };
  },
  shortcutsTestId: 'biriba-kbd-shortcuts',
  actionsClassName: '',
  Board,
  FooterLead,
  FooterHand,
};

/** Biriba game page. */
export const BiribaPage = createGamePage(biribaDefinition);
