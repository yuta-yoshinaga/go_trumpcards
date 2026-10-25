import type { TFunction } from 'i18next';
import { Fragment, useMemo } from 'react';
import type { canastaApi } from '../api/gameApi';
import { CardFan } from '../components/CardFan';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useCanastaGame } from '../hooks/useCanastaGame';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeInfoColors } from '../styles/badgeStyles';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { CanastaPlayerData, CanastaResponse, Card } from '../types/card';
import { CanastaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { canastaDrawDiscardProblem } from '../utils/canastaDrawDiscard';
import { canastaFamilySelectionPoints } from '../utils/canastaFamilyScore';
import { cardAlt } from '../utils/cardAlt';
import { CANASTA_HELP, parseCanastaCommand } from '../utils/cli/commands/canastaCommands';
import { formatCanastaState } from '../utils/cli/formatters/canastaFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
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

/** Render one player's melds and red threes. */
function CanastaMeldArea({
  player,
  playerIndex,
  cardWidth,
  t,
  tutorialTarget,
  className,
  maxHeight,
}: {
  player: CanastaPlayerData;
  playerIndex: number;
  cardWidth: number;
  t: (key: string) => string;
  tutorialTarget?: string;
  className?: string;
  maxHeight?: number;
}) {
  const hasMeldContent = player.melds.length > 0 || player.red3s.length > 0;
  return (
    <div
      className={`${hasMeldContent ? 'my-2 p-2 rounded bg-black/30 ' : ''}${className ?? ''}`}
      data-tutorial={tutorialTarget}
      style={maxHeight === undefined ? undefined : { maxHeight }}
    >
      <div className="text-ds-text-muted text-sm mb-1">
        {playerName(player.id, player.isHuman)} - {t('melds')}
        {player.hasCanasta && <span className="ml-2 text-ds-warning">★</span>}
      </div>
      {player.melds.map((meld, meldIndex) => {
        const meldWidth = cardWidth * 0.6;
        const canastaLabel = meld.isCanasta ? (meld.isNatural ? t('naturalCanasta') : t('mixedCanasta')) : null;
        return (
          <details key={meldIndex} className="mb-1" data-testid={`ca-meld-${playerIndex}-${meldIndex}`}>
            <summary className="flex items-center gap-2 cursor-pointer list-none marker:hidden" title={t('meldExpand')}>
              <div className="relative flex items-center shrink-0">
                {meld.cards.map((card, cardIndex) => (
                  <div
                    key={`meld-stack-${playerIndex}-${meldIndex}-${cardIndex}`}
                    style={{ marginLeft: cardIndex === 0 ? 0 : -meldWidth * 0.72, zIndex: cardIndex }}
                  >
                    <AnimatedCard card={card} width={meldWidth} silent />
                  </div>
                ))}
              </div>
              <span
                className="text-xs font-bold px-1.5 py-0.5 rounded-full bg-black/50 text-ds-text-primary"
                data-testid={`ca-meld-badge-${playerIndex}-${meldIndex}`}
              >
                {meld.cards.length}
              </span>
              {canastaLabel && <span className="text-xs font-bold text-ds-warning">★ {canastaLabel}</span>}
            </summary>
            <div className="flex flex-wrap gap-1 mt-1 pl-2">
              {meld.cards.map((card, cardIndex) => (
                <AnimatedCard
                  key={`meld-${playerIndex}-${meldIndex}-${cardIndex}`}
                  card={card}
                  width={meldWidth}
                  silent
                />
              ))}
            </div>
          </details>
        );
      })}
      {player.red3s.length > 0 && (
        <div className="flex flex-wrap gap-1 mt-1">
          <span className="text-xs text-ds-error-text self-center mr-1">{t('red3s')}</span>
          {player.red3s.map((card, red3Index) => (
            <AnimatedCard key={`red3-${playerIndex}-${red3Index}`} card={card} width={cardWidth * 0.6} />
          ))}
        </div>
      )}
    </div>
  );
}

function useGame() {
  const game = useCanastaGame();
  return {
    ...game,
    exec: game.gameExec,
  };
}

type CanastaGame = ReturnType<typeof useGame>;
type CanastaView = ReturnType<typeof useCanastaView>;
type CanastaContext = GamePageContext<CanastaResponse, CanastaGame, CanastaView>;

function useCanastaView({ game, t }: { game: CanastaGame; t: TFunction }) {
  const { cardWidth, isMobile } = useCardDimensions();
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const phaseNames = usePhaseNames('canasta', CANASTA_PHASE_KEYS);

  const state = game.state;
  const humanPlayer = state?.players.find((p) => p.isHuman);
  const isDrawPhase = state?.phase === CanastaPhase.DRAW;
  const isMeldPhase = state?.phase === CanastaPhase.MELD;
  // The server rejects a go-out unless the player holds at least one completed
  // canasta (Canasta.PlayerGoOut -> HasCanasta()); mirror that guard client-side.
  const canGoOut = humanPlayer?.hasCanasta === true;

  // Two cards used to be enough to enable the button, whatever they were, while
  // PlayerDrawFromDiscard demands two *natural* cards of the discard top's rank —
  // always, not only while frozen. The pair is now checked before sending (#4825).
  const drawDiscardProblem = useMemo(() => {
    if (!isDrawPhase || !humanPlayer) return null;
    const selected = game.selectedCardIndices
      .map((i) => humanPlayer.cards[i])
      .filter((card): card is Card => card !== undefined);
    return canastaDrawDiscardProblem(selected, state?.discardTop);
  }, [isDrawPhase, humanPlayer, game.selectedCardIndices, state?.discardTop]);

  const drawDiscardReason = useMemo(() => {
    if (!isDrawPhase) return '';
    // Frozen is advice rather than a rejection, so it keeps priority while the
    // selection is still incomplete — at 0 or 1 cards, as it did before.
    if (state?.isFrozen && (drawDiscardProblem === 'selectTwo' || drawDiscardProblem === 'selectOneMore')) {
      return t('drawDiscardReason.frozen');
    }
    return drawDiscardProblem === null ? '' : t(`drawDiscardReason.${drawDiscardProblem}`);
  }, [isDrawPhase, state?.isFrozen, drawDiscardProblem, t]);

  // Meld phase: surface the initial-meld minimum (by score band) and the
  // selected cards' running point total so the player can tell if they qualify.
  // **メルドフェーズ以外では 0 点・未達なしを返す。** null を返すと、この値を
  // 読むたびに「無いかもしれない」分岐が増え、押せるかどうかの判断が追いにくい
  // (#6165、ハンド＆フットと同じ形)。
  const meldPointInfo = useMemo(() => {
    if (!isMeldPhase || !humanPlayer) {
      return { selectedPoints: 0, needInitial: false, minMeld: 0, below: false };
    }
    const selectedCards = game.selectedCardIndices
      .map((i) => humanPlayer.cards[i])
      .filter((c): c is Card => Boolean(c));
    const selectedPoints = canastaFamilySelectionPoints(selectedCards);
    const needInitial = !humanPlayer.hasInitMeld;
    const minMeld = state.minMeld;
    return { selectedPoints, needInitial, minMeld, below: needInitial && selectedPoints < minMeld };
  }, [isMeldPhase, humanPlayer, game.selectedCardIndices, state?.minMeld]);

  return {
    cardWidth,
    isMobile,
    isLargeDesktop,
    phaseNames,
    drawDiscardProblem,
    drawDiscardReason,
    meldPointInfo,
    canGoOut,
  };
}

function LiveRegion({ state, t, view }: CanastaContext) {
  const isMeldPhase = state.phase === CanastaPhase.MELD;
  const isHumanTurn = isMeldPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const { meldPointInfo } = view;

  return (
    <div
      role="status"
      aria-live="polite"
      aria-atomic="true"
      className="sr-only"
      data-testid="ca-meld-points-announcement"
    >
      {isMeldPhase && isHumanTurn
        ? meldPointInfo.needInitial
          ? t('meldPoints.initial', {
              min: meldPointInfo.minMeld,
              points: meldPointInfo.selectedPoints,
            })
          : t('meldPoints.selected', { points: meldPointInfo.selectedPoints })
        : ''}
    </div>
  );
}

function Board({ state, t, tc, frontendHint, view }: CanastaContext) {
  const { cardWidth, isMobile, isLargeDesktop } = view;
  const isRoundEnd = state.phase === CanastaPhase.ROUND_END;
  const isGameEnd = state.phase === CanastaPhase.GAME_END || !!state.gameEndFlag;

  return (
    <>
      <div className="flex flex-wrap justify-center text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span>
          {t('drawPile', { count: state.drawPileCount })} / {t('discardPile', { count: state.discardPileCount })}
        </span>
        {state.isFrozen && (
          <span className="ml-2 shrink-0 whitespace-nowrap text-ds-info font-bold">[{t('frozen')}]</span>
        )}
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          <div
            className={`my-3 p-3 rounded relative ${
              state.isFrozen ? `${badgeInfoColors} ring-2 ring-ds-info` : 'bg-black/40'
            }`}
            data-tutorial="ca-draw-area"
            data-testid="ca-discard-pile"
          >
            <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
              <span className="text-ds-text-muted text-sm">{t('discardTop')}</span>
              {state.isFrozen && (
                <span
                  className="text-ds-info text-xs font-bold"
                  data-testid="ca-frozen-badge"
                  role="img"
                  aria-label={t('frozenIndicator')}
                >
                  {t('frozenIndicator')}
                </span>
              )}
            </div>
            <div className="max-h-48 overflow-y-auto">
              <div className="flex flex-wrap gap-1 items-start">
                {state.discardPile.map((card, idx) => (
                  <div
                    key={`ca-discard-${idx}`}
                    data-discard-top={idx === state.discardPile.length - 1 ? 'true' : undefined}
                    className={
                      idx === state.discardPile.length - 1
                        ? 'outline outline-2 outline-ds-accent outline-offset-2 rounded'
                        : undefined
                    }
                  >
                    {idx === state.discardPile.length - 1 && <span className="sr-only">{t('discardPileTop')}</span>}
                    <AnimatedCard card={card} width={cardWidth} />
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Player melds */}
          {state.players.map((p, pi) => {
            if (isLargeDesktop && p.isHuman) return null;
            const hasMeldContent = p.melds.length > 0 || p.red3s.length > 0;
            if (!hasMeldContent && !p.isHuman) return null;
            return (
              <CanastaMeldArea
                key={pi}
                player={p}
                playerIndex={pi}
                cardWidth={cardWidth}
                t={t}
                tutorialTarget={pi === 0 ? 'ca-meld-area' : undefined}
              />
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
                  <Fragment key={p.id}>
                    <tr className={p.isHuman ? 'text-ds-accent' : ''}>
                      <td>{playerName(p.id, p.isHuman)}</td>
                      <td className="text-center">{p.roundScore}</td>
                      <td className="text-center">{p.cumulativeScore}</td>
                    </tr>
                    {(isRoundEnd || isGameEnd) && (
                      <tr>
                        <td colSpan={3} className="pb-2">
                          <dl className="grid grid-cols-2 gap-x-2 text-xs">
                            <dt>{t('score.meldCards')}</dt>
                            <dd className="text-right">{p.scoreBreakdown.meldCards}</dd>
                            <dt>{t('score.canastaBonus')}</dt>
                            <dd className="text-right">{p.scoreBreakdown.canastaBonus}</dd>
                            <dt>{t('score.red3Bonus')}</dt>
                            <dd className="text-right">{p.scoreBreakdown.red3Bonus}</dd>
                            <dt>{t('score.goOutBonus')}</dt>
                            <dd className="text-right">{p.scoreBreakdown.goOutBonus}</dd>
                            <dt>{t('score.handPenalty')}</dt>
                            <dd className="text-right">−{p.scoreBreakdown.handPenalty}</dd>
                          </dl>
                        </td>
                      </tr>
                    )}
                  </Fragment>
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

      {isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}

      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand({ game, state, t, view }: CanastaContext) {
  const { cardWidth, isMobile, isLargeDesktop } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);

  return (
    <div className={isLargeDesktop ? 'flex items-start gap-4' : undefined}>
      {isLargeDesktop && humanPlayer && (
        <CanastaMeldArea
          player={humanPlayer}
          playerIndex={0}
          cardWidth={cardWidth}
          t={t}
          tutorialTarget="ca-meld-area"
          className="shrink-0 max-w-[30%] overflow-y-auto"
          maxHeight={cardWidth * 1.4}
        />
      )}
      <div className={isLargeDesktop ? 'min-w-0 flex-1' : undefined}>
        {humanPlayer &&
          (isMobile ? (
            <MobileHandGrid
              cards={humanPlayer.cards}
              selectedIndices={game.selectedCardIndices}
              onToggle={game.toggleCard}
              cardWidth={cardWidth}
              dataTutorial="ca-player-hand"
              stackRows
            />
          ) : isLargeDesktop ? (
            <CardFan itemWidth={cardWidth + 6} data-testid="ca-player-hand" dataTutorial="ca-player-hand">
              {humanPlayer.cards.map((card, idx) => (
                <button
                  type="button"
                  key={`${card.design}-${card.value}-${idx}`}
                  onClick={() => game.toggleCard(idx)}
                  aria-label={cardAlt(card)}
                  aria-pressed={game.selectedCardIndices.includes(idx)}
                  data-testid={`hand-card-${idx}`}
                  className={`transition-transform ${focusRingCard}`}
                  style={{
                    background: 'none',
                    padding: 0,
                    borderRadius: 8,
                    ...selectedCardStyle(game.selectedCardIndices.includes(idx)),
                    boxSizing: 'border-box',
                  }}
                >
                  <AnimatedCard card={card} width={cardWidth} />
                </button>
              ))}
            </CardFan>
          ) : (
            <div className="flex flex-wrap gap-1 mb-2" data-tutorial="ca-player-hand">
              {humanPlayer.cards.map((card, idx) => (
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
                    boxSizing: 'border-box',
                  }}
                >
                  <AnimatedCard card={card} width={cardWidth} />
                </button>
              ))}
            </div>
          ))}
      </div>
    </div>
  );
}

function primaryActions({ game, state, t, busy, view }: CanastaContext) {
  const loading = busy;
  const isDrawPhase = state.phase === CanastaPhase.DRAW;
  const isMeldPhase = state.phase === CanastaPhase.MELD;
  const isDiscardPhase = state.phase === CanastaPhase.DISCARD;
  const isRoundEnd = state.phase === CanastaPhase.ROUND_END;
  const isHumanTurn =
    (isDrawPhase || isMeldPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
  const { drawDiscardProblem, drawDiscardReason, meldPointInfo, canGoOut } = view;

  return (
    <div className="flex gap-2 items-center flex-wrap" data-tutorial="ca-actions">
      {isDrawPhase && isHumanTurn && (
        <div className="flex gap-2 flex-col">
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
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleMeldSelected}
            // **最低点未達なら押させない** (#6165)。警告テキストは出て
            // いたのにボタンはそれを見ておらず、サーバのバリデーションで
            // 弾かれて初めて気づく形だった。
            disabled={loading || game.selectedCardIndices.length < 3 || meldPointInfo.below}
            aria-describedby={meldPointInfo.below ? 'ca-meld-points' : undefined}
          >
            {t('meldButton')}
          </button>
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
            className={btnSuccess}
            onClick={game.handleGoOut}
            disabled={loading || !canGoOut}
            title={canGoOut ? undefined : t('goOutReason')}
            aria-describedby={canGoOut ? undefined : 'ca-go-out-reason'}
            data-testid="ca-go-out-button"
          >
            {t('goOutButton')}
          </button>
          {!canGoOut && (
            <div id="ca-go-out-reason" data-testid="ca-go-out-reason" className="w-full text-xs text-ds-text-muted">
              {t('goOutReason')}
            </div>
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

const canastaDefinition: GameDefinition<
  CanastaResponse,
  CanastaGame,
  Parameters<typeof canastaApi.exec>,
  CanastaView
> = {
  key: 'canasta',
  useGame,
  useView: useCanastaView,
  tutorial: CA_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 11 },
  cli: ({ frontendHint }) => ({
    parseCommand: parseCanastaCommand,
    formatResponse: formatCanastaState,
    helpText: CANASTA_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === CanastaPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isDrawPhase = state.phase === CanastaPhase.DRAW;
    const isMeldPhase = state.phase === CanastaPhase.MELD;
    const isDiscardPhase = state.phase === CanastaPhase.DISCARD;
    const isHumanTurn =
      (isDrawPhase || isMeldPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
    return { isHumanTurn };
  },
  reset: ({ game }) => {
    void game.exec('reset', undefined, {
      cpuDifficulty: game.canastaConfig.cpuDifficulty,
      pointLimit: game.canastaConfig.pointLimit,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.canastaConfig.cpuDifficulty,
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
          value: game.canastaConfig.pointLimit,
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
  actionsClassName: '',
  message: ({ state, view }) =>
    view.isMobile ? undefined : (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ),
  cardNav: ({ game, state, view }) => {
    const isDiscardPhase = state.phase === CanastaPhase.DISCARD;
    const isMeldPhase = state.phase === CanastaPhase.MELD;
    const humanPlayer = state.players.find((p) => p.isHuman);
    const humanCardCount = humanPlayer?.cards?.length ?? 0;
    const isHumanTurn =
      (state.phase === CanastaPhase.DRAW || isMeldPhase || isDiscardPhase) &&
      state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      cardCount: humanCardCount,
      onToggle: game.toggleCard,
      // **キーボードの確定もボタンと同じ門番を通す** (#6165)。Enter が
      // handleMeldSelected を直に呼ぶと、ボタンを無効化してもキーだけ通り、
      // サーバのバリデーションで弾かれる形が残る。
      onConfirm: () => {
        if (isDiscardPhase) game.handleDiscard();
        else if (isMeldPhase && !view.meldPointInfo.below) game.handleMeldSelected();
      },
      onClear: game.clearSelection,
      enabled: !!isHumanTurn,
    };
  },
  shortcutsTestId: 'canasta-kbd-shortcuts',
  LiveRegion,
  Board,
  FooterHand,
};

/** Canasta game page. */
export const CanastaPage = createGamePage(canastaDefinition);
