import { useMemo } from 'react';
import type { ginrummyApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useGinRummyGame } from '../hooks/useGinRummyGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, meldCardStyle, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { GinRummyResponse } from '../types/card';
import { GinRummyCpu, GinRummyPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { GINRUMMY_HELP, parseGinrummyCommand } from '../utils/cli/commands/ginrummyCommands';
import { formatGinrummyState } from '../utils/cli/formatters/ginrummyFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import {
  bestDeadwoodValue,
  bestMeldSplit,
  GIN_RUMMY_KNOCK_THRESHOLD,
  ginRummyMeldLabel,
  ginRummyScoreBreakdown,
} from '../utils/ginRummyDeadwood';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const GINRUMMY_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GinRummyPhase.DRAW]: 'draw',
  [GinRummyPhase.DISCARD]: 'discard',
  [GinRummyPhase.LAYOFF]: 'layoff',
  [GinRummyPhase.ROUND_END]: 'roundEnd',
  [GinRummyPhase.GAME_END]: 'gameEnd',
};

/** Gin Rummy tutorial step definitions. */
const GR_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="gr-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="gr-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-knock-button"]',
    messageKey: 'tutorial.knockButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Interpolation values for the CPU policy descriptions, kept next to the constants. */
const GIN_RUMMY_POLICY_VALUES = {
  knockNormal: GinRummyCpu.KNOCK_DEADWOOD_NORMAL,
  knockHard: GinRummyCpu.KNOCK_DEADWOOD_HARD,
  gainNormal: GinRummyCpu.DISCARD_GAIN_NORMAL,
  easyOneIn: GinRummyCpu.EASY_PICK_ONE_IN,
  knockLimit: GinRummyCpu.KNOCK_THRESHOLD,
};

type GinRummyGame = ReturnType<typeof useGinRummyGame>;
type GinRummyView = ReturnType<typeof useGinRummyView>;
type GinRummyContext = GamePageContext<GinRummyResponse, GinRummyGame, GinRummyView>;

function useGinRummyView({ game }: { game: GinRummyGame }) {
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('ginrummy', GINRUMMY_PHASE_KEYS);
  const state = game.state;
  const selectedCardIndices = game.selectedCardIndices;

  // Knock requires the *post-discard* hand to be ≤ 10. When the user has a
  // card selected we project that discard; otherwise show the best-case
  // deadwood across all possible single discards (a knockability hint).
  const liveDeadwood = useMemo<number | null>(() => {
    if (!state) return null;
    if (state.phase !== GinRummyPhase.DISCARD) return null;
    const human = state.players.find((p) => p.isHuman);
    if (!human) return null;
    const cards = human.cards;
    if (cards.length === 0) return null;
    if (selectedCardIndices.length === 1) {
      const drop = selectedCardIndices[0];
      return bestDeadwoodValue(cards.filter((_, i) => i !== drop));
    }
    let best = Number.POSITIVE_INFINITY;
    for (let i = 0; i < cards.length; i++) {
      const dv = bestDeadwoodValue(cards.filter((_, j) => j !== i));
      if (dv < best) best = dv;
      if (best === 0) break;
    }
    return Number.isFinite(best) ? best : null;
  }, [state, selectedCardIndices]);
  const canKnockNow = liveDeadwood != null && liveDeadwood <= GIN_RUMMY_KNOCK_THRESHOLD;
  const canKnockSelectedCard = selectedCardIndices.length === 1 && canKnockNow;

  // During DISCARD, color-code the human's hand by best meld split so the
  // player can see which cards form melds vs. which are deadwood. Shares the
  // same search as the deadwood indicator, so the two always agree. Empty set
  // outside DISCARD leaves every card in its neutral (uncolored) state.
  const meldedIndices = useMemo<ReadonlySet<number>>(() => {
    if (!state || state.phase !== GinRummyPhase.DISCARD) return new Set();
    const human = state.players.find((p) => p.isHuman);
    if (!human || human.cards.length === 0) return new Set();
    return bestMeldSplit(human.cards).meldedIndices;
  }, [state]);

  // レイオフできる手札。判定はサーバー (ドメインの canLayoff と同じ) が返す
  // layoffTargets をそのまま使う — ここで書き直すと結果が食い違う (#4823)。
  const layoffableIndices = useMemo<ReadonlySet<number>>(() => {
    const targets = state?.layoffTargets ?? [];
    return new Set(targets.flatMap((list, idx) => (list.length > 0 ? [idx] : [])));
  }, [state?.layoffTargets]);

  // At round/game end, derive the additive score breakdown (outcome, each
  // player's deadwood, and the deadwood-difference + bonus components) from the
  // exposed round result, mirroring the domain's scoreRound. `null` for a drawn
  // round (stock exhausted) so the panel stays hidden.
  const scoreBreakdown = useMemo(() => {
    if (!state) return null;
    const atRoundEnd = state.phase === GinRummyPhase.ROUND_END || state.phase === GinRummyPhase.GAME_END;
    if (!atRoundEnd && !state.gameEndFlag) return null;
    return ginRummyScoreBreakdown(state.players, state.knockerIdx, state.knockerDeadwood, state.isGin);
  }, [state]);

  return {
    cardWidth,
    phaseNames,
    liveDeadwood,
    canKnockNow,
    canKnockSelectedCard,
    meldedIndices,
    layoffableIndices,
    scoreBreakdown,
  };
}

function Board({ state, t, view }: GinRummyContext) {
  const { cardWidth, scoreBreakdown } = view;
  const isLayoffPhase = state.phase === GinRummyPhase.LAYOFF;
  const isRoundEnd = state.phase === GinRummyPhase.ROUND_END;
  const isGameEnd = state.phase === GinRummyPhase.GAME_END || state.gameEndFlag;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span>{t('drawPile', { count: state.drawPileCount })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          {/* Discard pile top */}
          {state.discardTop && (
            <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
              <AnimatedCard card={state.discardTop} width={cardWidth} />
              <div className="text-ds-text-muted text-sm">
                <div>{t('discardTop')}</div>
                <div>{t('discardTopCard', { card: cardAlt(state.discardTop) })}</div>
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
                    data-testid={`gr-meld-badge-${meldIdx}`}
                    className="inline-block rounded border border-ds-border-subtle px-1.5 py-0.5 text-ds-text-primary text-xs mb-0.5"
                  >
                    {ginRummyMeldLabel(meld.cards)}
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
          {/* CPU player */}
          {state.players
            .filter((p) => !p.isHuman)
            .map((p) => (
              <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                <div className="text-ds-text-muted text-sm">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })}
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
          <div className="my-3 p-2 rounded bg-black/30" data-tutorial="gr-score-table">
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
                  <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
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

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {scoreBreakdown &&
        (() => {
          const winner = state.players.find((p) => p.id === scoreBreakdown.winnerId);
          const winnerLabel = winner ? playerName(winner.id, winner.isHuman) : '';
          const basePart =
            scoreBreakdown.outcome === 'gin'
              ? t('breakdown.opponentDeadwoodPart', { value: scoreBreakdown.base })
              : t('breakdown.differencePart', { value: scoreBreakdown.base });
          const bonusPart =
            scoreBreakdown.bonus > 0
              ? scoreBreakdown.outcome === 'gin'
                ? t('breakdown.ginBonusPart', { value: scoreBreakdown.bonus })
                : t('breakdown.undercutBonusPart', { value: scoreBreakdown.bonus })
              : null;
          const formula = `${basePart}${bonusPart ? ` + ${bonusPart}` : ''} = ${scoreBreakdown.total}`;
          return (
            <div className="my-3 p-3 rounded bg-black/30" data-testid="ginrummy-score-breakdown">
              <div className="text-ds-text-muted text-sm mb-1">{t('breakdown.title')}</div>
              <div className="flex items-center gap-2 mb-1 flex-wrap">
                <span
                  data-testid="ginrummy-breakdown-outcome"
                  className="inline-block rounded border border-ds-border-subtle px-1.5 py-0.5 text-ds-text-primary text-xs"
                >
                  {t(`breakdown.outcome.${scoreBreakdown.outcome}`)}
                </span>
                <span className="text-ds-accent text-sm font-bold">
                  {t('breakdown.winner', { name: winnerLabel, total: scoreBreakdown.total })}
                </span>
              </div>
              <div className="text-ds-text-muted text-xs mb-1">
                {t('breakdown.deadwood', {
                  knocker: scoreBreakdown.knockerDeadwood,
                  opponent: scoreBreakdown.opponentDeadwood,
                })}
              </div>
              <div className="text-ds-text-primary text-sm" data-testid="ginrummy-breakdown-formula">
                {formula}
              </div>
            </div>
          );
        })()}
    </>
  );
}

function primaryActions({ game, state, t, busy, view }: GinRummyContext) {
  const isDrawPhase = state.phase === GinRummyPhase.DRAW;
  const isDiscardPhase = state.phase === GinRummyPhase.DISCARD;
  const isLayoffPhase = state.phase === GinRummyPhase.LAYOFF;
  const isRoundEnd = state.phase === GinRummyPhase.ROUND_END;
  const isHumanTurn =
    (isDrawPhase || isDiscardPhase || isLayoffPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  if (
    !((isDrawPhase && isHumanTurn) || (isDiscardPhase && isHumanTurn) || (isLayoffPhase && isHumanTurn) || isRoundEnd)
  ) {
    return null;
  }

  return (
    <div className="flex gap-2 items-center flex-wrap">
      {isDrawPhase && isHumanTurn && (
        <div className="flex gap-2" data-tutorial="gr-draw-area">
          <button type="button" className={btnPrimary} onClick={game.handleDrawStock} disabled={busy}>
            {t('drawStockButton')}
          </button>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleDrawDiscard}
            disabled={busy || !state.discardTop}
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
            onClick={game.handleDiscard}
            disabled={busy || game.selectedCardIndices.length !== 1}
            data-tutorial="gr-discard-button"
          >
            {t('discardButton')}
          </button>
          <button
            type="button"
            className={`${btnPrimary} aria-disabled:cursor-not-allowed aria-disabled:opacity-50 ${view.canKnockSelectedCard ? 'motion-safe:animate-pulse ring-2 ring-ds-success' : ''}`}
            onClick={() => {
              if (!busy && view.canKnockSelectedCard) game.handleKnock();
            }}
            disabled={busy}
            aria-disabled={!view.canKnockSelectedCard || undefined}
            aria-describedby={!view.canKnockSelectedCard ? 'ginrummy-knock-disabled-reason' : undefined}
            data-tutorial="gr-knock-button"
            data-testid="ginrummy-knock-button"
          >
            {t('knockButton')}
          </button>
          {!view.canKnockSelectedCard && (
            <span id="ginrummy-knock-disabled-reason" className="sr-only">
              {t('knockDisabledReason')}
            </span>
          )}
        </>
      )}
      {isLayoffPhase && isHumanTurn && (
        <>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleLayoff}
            disabled={busy || game.selectedCardIndices.length === 0}
          >
            {t('layoffButton')}
          </button>
          <button type="button" className={btnPrimary} onClick={game.handleSkipLayoff} disabled={busy}>
            {t('skipLayoffButton')}
          </button>
        </>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

function FooterHand({ game, state, t, view }: GinRummyContext) {
  const isDiscardPhase = state.phase === GinRummyPhase.DISCARD;
  const isLayoffPhase = state.phase === GinRummyPhase.LAYOFF;
  const humanPlayer = state.players.find((p) => p.isHuman);

  return (
    <>
      {view.liveDeadwood != null && (
        <div
          data-testid="ginrummy-deadwood-indicator"
          className={`text-xs font-bold mb-1 ${view.canKnockNow ? 'text-ds-success' : 'text-ds-text-muted'}`}
        >
          {t('deadwoodLabel', { score: view.liveDeadwood, threshold: GIN_RUMMY_KNOCK_THRESHOLD })}
        </div>
      )}
      {isDiscardPhase && (
        <div data-testid="ginrummy-meld-legend" className="flex items-center gap-3 text-xs text-ds-text-muted mb-1">
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
        <div className="flex flex-wrap gap-1 mb-2" data-tutorial="gr-player-hand">
          {isLayoffPhase && (
            <span id="ginrummy-layoff-unavailable" className="sr-only">
              {t('notLayoffableCard')}
            </span>
          )}
          {humanPlayer.cards.map((card, idx) => (
            <button
              type="button"
              key={`${card.design}-${card.value}-${idx}`}
              onClick={() => {
                if (isLayoffPhase && !view.layoffableIndices.has(idx)) return;
                game.toggleCard(idx);
              }}
              aria-disabled={isLayoffPhase && !view.layoffableIndices.has(idx) ? 'true' : undefined}
              aria-describedby={
                isLayoffPhase && !view.layoffableIndices.has(idx) ? 'ginrummy-layoff-unavailable' : undefined
              }
              aria-label={
                isDiscardPhase
                  ? t('cardStatus', {
                      card: cardAlt(card),
                      status: t(view.meldedIndices.has(idx) ? 'meldLegend' : 'deadwoodLegend'),
                    })
                  : isLayoffPhase
                    ? t('cardStatus', {
                        card: cardAlt(card),
                        status: t(view.layoffableIndices.has(idx) ? 'layoffableCard' : 'notLayoffableCard'),
                      })
                    : cardAlt(card)
              }
              aria-pressed={game.selectedCardIndices.includes(idx)}
              data-testid={`gr-hand-card-${idx}`}
              data-meld={isDiscardPhase ? (view.meldedIndices.has(idx) ? 'meld' : 'deadwood') : undefined}
              // **レイオフフェーズの主題そのもの。**ディスカードでは
              // メルド/デッドウッドを見せているのに、レイオフには補助が
              // 無く、押してサーバーの応答で初めて成否が分かった (#4823)。
              data-layoff={isLayoffPhase ? (view.layoffableIndices.has(idx) ? 'yes' : 'no') : undefined}
              className={`transition-transform ${focusRingCard} ${isLayoffPhase && !view.layoffableIndices.has(idx) ? 'opacity-50 cursor-not-allowed' : ''}`}
              style={{
                background: 'none',
                padding: 0,
                borderRadius: 8,
                ...(isDiscardPhase ? meldCardStyle(view.meldedIndices.has(idx)) : undefined),
                ...(isLayoffPhase ? meldCardStyle(view.layoffableIndices.has(idx)) : undefined),
                ...selectedCardStyle(game.selectedCardIndices.includes(idx)),
                boxSizing: 'border-box',
              }}
            >
              <AnimatedCard card={card} width={view.cardWidth} />
            </button>
          ))}
        </div>
      )}
    </>
  );
}

function FooterStatus({ frontendHint, t }: GinRummyContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

const ginRummyDefinition: GameDefinition<
  GinRummyResponse,
  GinRummyGame,
  Parameters<typeof ginrummyApi.exec>,
  GinRummyView
> = {
  key: 'ginrummy',
  useGame: useGinRummyGame,
  useView: useGinRummyView,
  tutorial: GR_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 10 },
  cli: ({ frontendHint }) => ({
    parseCommand: parseGinrummyCommand,
    formatResponse: formatGinrummyState,
    helpText: GINRUMMY_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === GinRummyPhase.GAME_END || Boolean(state.gameEndFlag),
  }),
  shell: ({ state }) => {
    const isDrawPhase = state.phase === GinRummyPhase.DRAW;
    const isDiscardPhase = state.phase === GinRummyPhase.DISCARD;
    const isLayoffPhase = state.phase === GinRummyPhase.LAYOFF;
    const isHumanTurn =
      (isDrawPhase || isDiscardPhase || isLayoffPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      isHumanTurn,
    };
  },
  reset: ({ game }) => {
    void game.exec('reset', undefined, {
      cpuDifficulty: game.ginRummyConfig.cpuDifficulty,
      pointLimit: game.ginRummyConfig.pointLimit,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.ginRummyConfig.cpuDifficulty,
          // **難易度で変わるのはノックと拾いの基準。** Easy/Normal/Hard の
          // ラベルしか出ておらず、実際の判断基準はプレイヤーから見えな
          // かった (#5500)。数値は GinRummyCpu から取る -- domain 定数と
          // 突き合わせる Go のテストがあるので、片方だけ変えると落ちる。
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            // 括弧はロケール側に持たせる。JSX で全角括弧を書くと英語表示にも
            // 混ざる (Memory の difficultyOption と同じ形にそろえた)。
            label: t('settings.difficultyWithPolicy', {
              level: t(`settings.${o.label.toLowerCase()}`),
              policy: t(`settings.policy${o.label}`, GIN_RUMMY_POLICY_VALUES),
            }),
          })),
          tooltip: t('settings.policyTooltip', GIN_RUMMY_POLICY_VALUES),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'pointLimit',
          label: t('settings.pointLimit'),
          value: game.ginRummyConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('pointLimit', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  cardNav: ({ game, state }) => {
    const isDiscardPhase = state.phase === GinRummyPhase.DISCARD;
    const isLayoffPhase = state.phase === GinRummyPhase.LAYOFF;
    const isHumanTurn = (isDiscardPhase || isLayoffPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
    const humanCardCount = state.players.find((p) => p.isHuman)?.cards?.length ?? 0;
    return {
      cardCount: humanCardCount,
      onToggle: game.toggleCard,
      onConfirm: () => {
        if (isDiscardPhase) {
          game.handleDiscard();
        } else if (isLayoffPhase) {
          game.handleLayoff();
        }
      },
      onClear: game.clearSelection,
      enabled: isHumanTurn,
    };
  },
  shortcutsTestId: 'gin-rummy-kbd-shortcuts',
  controlsInFooter: true,
  resetTutorialId: 'gr-reset-button',
  actionsClassName: '',
  primaryActions,
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Gin Rummy game page with draw, discard, knock, and layoff phases. */
export const GinRummyPage = createGamePage(ginRummyDefinition);
