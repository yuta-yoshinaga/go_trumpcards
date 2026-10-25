import type { TFunction } from 'i18next';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { burracoApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useBurracoGame } from '../hooks/useBurracoGame';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { badgeInfoColors } from '../styles/badgeStyles';
import { btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, hintRingStyle, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { BurracoResponse, Card } from '../types/card';
import { BurracoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { burracoMeldSelectionStatus } from '../utils/burracoMeld';
import { type BurracoSortMode, loadBurracoSortMode, saveBurracoSortMode } from '../utils/burracoSort';
import { canastaDrawDiscardProblem } from '../utils/canastaDrawDiscard';
import { cardAlt } from '../utils/cardAlt';
import { BURRACO_HELP, parseBurracoCommand } from '../utils/cli/commands/burracoCommands';
import { formatBurracoState } from '../utils/cli/formatters/burracoFormatter';
import { sortedHandForDisplay } from '../utils/handDisplaySort';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const BURRACO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BurracoPhase.DRAW]: 'draw',
  [BurracoPhase.MELD]: 'meld',
  [BurracoPhase.DISCARD]: 'discard',
  [BurracoPhase.ROUND_END]: 'roundEnd',
  [BurracoPhase.GAME_END]: 'gameEnd',
};

/** Hand sort options for the Burraco footer. */
const BURRACO_SORT_MODES: { mode: BurracoSortMode; labelKey: string }[] = [
  { mode: 'original', labelKey: 'sort.original' },
  { mode: 'rank', labelKey: 'sort.rank' },
  { mode: 'suit', labelKey: 'sort.suit' },
];

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

function useGame() {
  const game = useBurracoGame();
  return {
    ...game,
    exec: game.gameExec,
  };
}

type BurracoGame = ReturnType<typeof useGame>;
type BurracoView = ReturnType<typeof useBurracoView>;
type BurracoContext = GamePageContext<BurracoResponse, BurracoGame, BurracoView>;

function useBurracoView({ game, t, tc }: { game: BurracoGame; t: TFunction; tc: TFunction }) {
  const isMobile = useIsMobile();
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const phaseNames = usePhaseNames('burraco', BURRACO_PHASE_KEYS);

  const state = game.state;
  const humanPlayer = state?.players.find((p) => p.isHuman);
  const humanSeat = state?.players.find((player) => player.isHuman)?.id ?? 0;

  const isDrawPhase = state?.phase === BurracoPhase.DRAW;

  const canGoOut = humanPlayer?.tookPozzetto === true && humanPlayer.hasBurraco;
  const goOutReason = canGoOut
    ? ''
    : t('goOutReason', {
        requirements: [
          humanPlayer?.tookPozzetto ? null : t('goOutRequirementPozzetto'),
          humanPlayer?.hasBurraco ? null : t('goOutRequirementBurraco'),
        ]
          .filter((requirement): requirement is string => requirement !== null)
          .join(t('listSeparator')),
      });

  // Burraco runs on the Canasta domain (`BurracoGame = CanastaGame`), so the same
  // pair rules apply and the same count-only check was letting rejected
  // selections look ready.
  const drawDiscardProblem = useMemo(() => {
    if (!isDrawPhase || !humanPlayer) return null;
    const selected = game.selectedCardIndices
      .map((i) => humanPlayer.cards[i])
      .filter((card): card is Card => card !== undefined);
    return canastaDrawDiscardProblem(selected, state?.discardTop);
  }, [isDrawPhase, humanPlayer, game.selectedCardIndices, state?.discardTop]);

  const drawDiscardReason = useMemo(() => {
    if (!isDrawPhase) return '';
    if (state?.isFrozen && (drawDiscardProblem === 'selectTwo' || drawDiscardProblem === 'selectOneMore')) {
      return t('drawDiscardReason.frozen');
    }
    return drawDiscardProblem === null ? '' : t(`drawDiscardReason.${drawDiscardProblem}`);
  }, [isDrawPhase, state?.isFrozen, drawDiscardProblem, t]);

  const meldSelectionStatus = useMemo(() => {
    if (!humanPlayer) return 'select' as const;
    const selected = game.selectedCardIndices
      .map((i) => humanPlayer.cards[i])
      .filter((card): card is Card => card !== undefined);
    return burracoMeldSelectionStatus(selected, humanPlayer.melds);
  }, [game.selectedCardIndices, humanPlayer]);

  const meldSelectionReason =
    meldSelectionStatus === 'select'
      ? t('meldSelectionReason.select')
      : meldSelectionStatus === 'valid'
        ? t(humanPlayer?.hasInitMeld ? 'meldSelectionReason.valid' : 'meldSelectionReason.validInitial')
        : t('meldSelectionReason.invalid');

  // Transient feedback when a player grabs the pozzetto (a pivotal Burraco moment)
  // and a one-shot pulse on round-score cells that just changed.
  // Display-only hand sort: reorders the rendered hand while every click maps
  // back to the card's original (server-dealt) index, so selection / meld /
  // discard stay index-correct. Persisted across sessions in localStorage.
  const [sortMode, setSortMode] = useState<BurracoSortMode>(loadBurracoSortMode);
  const handleSortMode = useCallback((mode: BurracoSortMode) => {
    setSortMode(mode);
    saveBurracoSortMode(mode);
  }, []);

  const [pozzettoBanner, setPozzettoBanner] = useState<string | null>(null);
  const [scoreAnnouncement, setScoreAnnouncement] = useState('');
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
    const changedScores = state.players.filter((p, i) => prevScores[i] !== undefined && prevScores[i] !== p.roundScore);
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
      setScoreAnnouncement(
        changedScores
          .map((p) =>
            t('score.update', {
              player: playerName(p.id, p.isHuman),
              score: p.roundScore,
              cumulativeScore: p.cumulativeScore,
            }),
          )
          .join(t('listSeparator')),
      );
      clearTimeout(pulseTimerRef.current ?? undefined);
      pulseTimerRef.current = setTimeout(() => setPulsingScoreIds(new Set()), 1000);
    }
  }, [state, tc, playSound, t]);

  return {
    isMobile,
    cardWidth,
    phaseNames,
    humanSeat,
    canGoOut,
    goOutReason,
    drawDiscardProblem,
    drawDiscardReason,
    meldSelectionReason,
    sortMode,
    handleSortMode,
    pozzettoBanner,
    scoreAnnouncement,
    pulsingScoreIds,
  };
}

function LiveRegion({ view }: BurracoContext) {
  return (
    <div role="status" aria-live="polite" className="sr-only" data-testid="bu-score-announcement">
      {view.scoreAnnouncement}
    </div>
  );
}

function Board({ state, t, tc, view, frontendHint }: BurracoContext) {
  const { isMobile, cardWidth, humanSeat, pozzettoBanner, pulsingScoreIds } = view;
  const isRoundEnd = state.phase === BurracoPhase.ROUND_END;
  const isGameEnd = state.phase === BurracoPhase.GAME_END || !!state.gameEndFlag;

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
      <div className="mb-2 flex flex-wrap items-center justify-center gap-2 text-[10px] text-ds-text-primary sm:gap-4 sm:text-base">
        <span>{t('round', { n: state.roundNumber })}</span>
        <span>
          {t('drawPile', { count: state.drawPileCount })} / {t('discardPile', { count: state.discardPileCount })}
        </span>
        <span
          data-testid="bu-pozzetto-count"
          title={t('pozzettoDescription')}
          aria-describedby="bu-pozzetto-description"
        >
          {t('pozzetto', { count: state.pozzettoCount })}
        </span>
        {state.isFrozen && <span className="shrink-0 whitespace-nowrap text-ds-info font-bold">[{t('frozen')}]</span>}
      </div>
      <details className="mb-2 text-xs text-ds-text-muted" data-testid="bu-rules">
        <summary className="min-h-11 cursor-pointer py-2">{t('rulesLabel')}</summary>
        <div id="bu-pozzetto-description" data-testid="bu-pozzetto-description">
          {t('pozzettoDescription')}
        </div>
      </details>

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
              <AnimatedCard card={state.discardTop} width={isMobile ? 36 : cardWidth} />
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
          <details className="my-3 rounded bg-black/30 p-2" data-testid="ca-discard-pile-viewer">
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
                      width={isMobile ? 36 : cardWidth * 0.6}
                    />
                  ))}
                </div>
              </>
            )}
          </details>

          {/* Player melds */}
          {state.players.map((p, pi) => {
            if (p.melds.length === 0 && p.red3s.length === 0) return null;
            // Burraco は 2 人戦 (NewDefaultBurraco: 人間 1 + CPU 1) なので席の偶奇で自分側が決まる。
            const isOwnTeam = Math.abs(p.id - humanSeat) % 2 === 0;
            const meldCount = p.melds.length + p.red3s.length;
            const meldContent = (
              <div
                key={pi}
                className="my-2 p-2 rounded bg-ds-surface"
                data-tutorial={pi === 0 ? 'ca-meld-area' : undefined}
              >
                <div className="text-ds-text-muted text-sm mb-1">
                  {playerName(p.id, p.isHuman)} - {t('melds')}
                  {p.hasBurraco && <span className="ml-2 text-ds-warning">★</span>}
                  {p.tookPozzetto && <span className="ml-2 text-ds-info text-xs">[{t('tookPozzetto')}]</span>}
                </div>
                {p.melds.map((m, mi) => (
                  <div key={mi} className="mb-1 flex flex-nowrap gap-0 sm:gap-1" data-testid="bu-meld-cards">
                    <span className="text-xs text-ds-text-muted self-center mr-1">
                      {m.isBurraco ? (m.isNatural ? t('naturalBurraco') : t('mixedBurraco')) : `(${m.cards.length})`}
                    </span>
                    {m.cards.map((card, ci) => (
                      <AnimatedCard
                        key={`meld-${pi}-${mi}-${ci}`}
                        card={card}
                        width={isMobile ? 36 : cardWidth * 0.6}
                        wrapperClassName={isMobile && ci > 0 ? '-ml-[22px]' : undefined}
                      />
                    ))}
                  </div>
                ))}
                {p.red3s.length > 0 && (
                  <div className="mt-1 flex flex-wrap gap-0 sm:gap-1">
                    <span className="text-xs text-ds-error-text self-center mr-1">{t('red3s')}</span>
                    {p.red3s.map((card, ri) => (
                      <AnimatedCard
                        key={`red3-${pi}-${ri}`}
                        card={card}
                        width={isMobile ? 36 : cardWidth * 0.6}
                        wrapperClassName={isMobile && ri > 0 ? '-ml-[22px]' : undefined}
                      />
                    ))}
                  </div>
                )}
              </div>
            );
            return isOwnTeam ? (
              meldContent
            ) : (
              <details key={pi} className="my-2 rounded bg-ds-surface p-2" open={!isMobile || undefined}>
                <summary className="cursor-pointer select-none text-sm text-ds-text-muted">
                  {playerName(p.id, p.isHuman)} - {t('melds')} ({meldCount})
                </summary>
                {meldContent}
              </details>
            );
          })}
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* Score table */}
          <details className="my-3 rounded bg-black/30 p-2" data-testid="bu-score-table" open={!isMobile || undefined}>
            <summary className="min-h-11 cursor-pointer py-2 text-sm text-ds-text-muted">{t('score.title')}</summary>
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

function FooterLead({ state, t, view }: BurracoContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer || humanPlayer.cards.length === 0) return null;

  return (
    <fieldset className="flex flex-wrap justify-center gap-1.5 mb-2 border-0 p-0 m-0">
      <legend className="sr-only">{t('sort.label')}</legend>
      {BURRACO_SORT_MODES.map(({ mode, labelKey }) => (
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

function FooterHand({ state, game, view, frontendHint }: BurracoContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  // ヒントが無効なとき・サーバがヒントを返さない場面 (CPU の手番など) では空。
  // **useGameHint が無効時に null を返す**ので、ここで再度フラグを見ない
  // (見ると、条件が二重になって片方が死ぬ)。
  const hintedCards = useMemo(
    () => new Set(frontendHint.hint?.targetIndices ?? []),
    [frontendHint.hint?.targetIndices],
  );

  if (!humanPlayer) return null;

  return (
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
  );
}

function primaryActions({ game, state, t, busy, view }: BurracoContext) {
  const loading = busy;
  const isDrawPhase = state.phase === BurracoPhase.DRAW;
  const isMeldPhase = state.phase === BurracoPhase.MELD;
  const isDiscardPhase = state.phase === BurracoPhase.DISCARD;
  const isRoundEnd = state.phase === BurracoPhase.ROUND_END;
  const isHumanTurn =
    (isDrawPhase || isMeldPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
  const { drawDiscardProblem, drawDiscardReason, meldSelectionReason, canGoOut, goOutReason } = view;

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
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleMeldSelected}
            disabled={loading || game.selectedCardIndices.length < 3}
            aria-describedby="bu-meld-selection-reason"
          >
            {t('meldButton')}
          </button>
          <button type="button" className={btnOutline} onClick={game.handleSkipMeld} disabled={loading}>
            {t('skipMeldButton')}
          </button>
          <span
            id="bu-meld-selection-reason"
            data-testid="bu-meld-selection-reason"
            className="text-xs text-ds-text-muted"
          >
            {meldSelectionReason}
          </span>
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
            className={`${btnSuccess} ${!canGoOut ? 'aria-disabled:opacity-70 aria-disabled:cursor-not-allowed' : ''}`}
            onClick={() => {
              if (!loading && canGoOut) game.handleGoOut();
            }}
            disabled={loading}
            aria-disabled={!canGoOut || undefined}
            aria-describedby={!canGoOut ? 'bu-go-out-reason' : undefined}
          >
            {t('goOutButton')}
          </button>
          {goOutReason && (
            <span id="bu-go-out-reason" data-testid="bu-go-out-reason" className="text-xs text-ds-text-muted">
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

const burracoDefinition: GameDefinition<
  BurracoResponse,
  BurracoGame,
  Parameters<typeof burracoApi.exec>,
  BurracoView
> = {
  key: 'burraco',
  useGame,
  useView: useBurracoView,
  tutorial: CA_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 11 },
  cli: () => ({
    parseCommand: parseBurracoCommand,
    formatResponse: formatBurracoState,
    helpText: BURRACO_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === BurracoPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isDrawPhase = state.phase === BurracoPhase.DRAW;
    const isMeldPhase = state.phase === BurracoPhase.MELD;
    const isDiscardPhase = state.phase === BurracoPhase.DISCARD;
    const isHumanTurn =
      (isDrawPhase || isMeldPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      isHumanTurn,
    };
  },
  reset: ({ game }) => {
    void game.exec('reset', undefined, {
      cpuDifficulty: game.burracoConfig.cpuDifficulty,
      pointLimit: game.burracoConfig.pointLimit,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.burracoConfig.cpuDifficulty,
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
          value: game.burracoConfig.pointLimit,
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
  message: (ctx) =>
    !ctx.view.isMobile && (
      <GameMessageBox
        message={ctx.state.message}
        messageCode={ctx.state.messageCode}
        messageParams={ctx.state.messageParams}
      />
    ),
  cardNav: ({ game, state }) => {
    const isDrawPhase = state.phase === BurracoPhase.DRAW;
    const isMeldPhase = state.phase === BurracoPhase.MELD;
    const isDiscardPhase = state.phase === BurracoPhase.DISCARD;
    const isHumanTurn =
      (isDrawPhase || isMeldPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
    const humanPlayer = state.players.find((p) => p.isHuman);

    return {
      cardCount: humanPlayer?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: () => {
        if (state.phase === BurracoPhase.DISCARD) game.handleDiscard();
        else if (state.phase === BurracoPhase.MELD) game.handleMeldSelected();
      },
      onClear: game.clearSelection,
      enabled: isHumanTurn,
    };
  },
  shortcutsTestId: 'burraco-kbd-shortcuts',
  actionsClassName: '',
  LiveRegion,
  Board,
  FooterLead,
  FooterHand,
};

/** Burraco game page. */
export const BurracoPage = createGamePage(burracoDefinition);
