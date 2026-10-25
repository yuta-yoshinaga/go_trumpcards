import type { TFunction } from 'i18next';
import { useCallback, useEffect, useRef, useState } from 'react';
import type { tongitsApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { TongitsOnDealCelebration } from '../components/TongitsOnDealCelebration';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useTongitsGame } from '../hooks/useTongitsGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, playableCardStyle, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { TongitsResponse } from '../types/card';
import { TongitsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTongitsCommand, TONGITS_HELP } from '../utils/cli/commands/tongitsCommands';
import { formatTongitsState } from '../utils/cli/formatters/tongitsFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { tongitsMeldIndices } from '../utils/tongitsMeldIndices';

/** Builds a stable public meld row key that is unique across players. */
export function publicMeldRowKey(playerId: number, meldIdx: number): string {
  return `opp-meld-${playerId}-${meldIdx}`;
}

const TONGITS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TongitsPhase.DRAW]: 'draw',
  [TongitsPhase.DISCARD]: 'discard',
  [TongitsPhase.ROUND_END]: 'roundEnd',
  [TongitsPhase.GAME_END]: 'gameEnd',
};

/** Tongits tutorial step definitions. */
const TONGITS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tongits-draw-area"]',
    messageKey: 'tutorial.drawArea',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tongits-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tongits-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tongits-challenge-button"]',
    messageKey: 'tutorial.challengeButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tongits-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tongits-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type TongitsGame = ReturnType<typeof useTongitsGame>;
type TongitsView = ReturnType<typeof useTongitsView>;
type TongitsContext = GamePageContext<TongitsResponse, TongitsGame, TongitsView>;

function useTongitsView({ game, t }: { game: TongitsGame; t: TFunction; tc: TFunction }) {
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('tongits', TONGITS_PHASE_KEYS);

  const [drawPileAnnouncement, setDrawPileAnnouncement] = useState('');
  const pendingStockDrawRef = useRef(false);
  const previousDrawPileCountRef = useRef<number | null>(null);

  const state = game.state;
  useEffect(() => {
    if (!state) return;
    if (previousDrawPileCountRef.current !== null && pendingStockDrawRef.current) {
      if (state.drawPileCount !== previousDrawPileCountRef.current) {
        setDrawPileAnnouncement(t('drawPileAnnouncement', { count: state.drawPileCount }));
      }
      pendingStockDrawRef.current = false;
    }
    previousDrawPileCountRef.current = state.drawPileCount;
  }, [state, t]);

  const announceStockDraw = useCallback(() => {
    pendingStockDrawRef.current = true;
    game.handleDrawStock();
  }, [game.handleDrawStock]);

  return {
    cardWidth,
    isMobile,
    phaseNames,
    drawPileAnnouncement,
    announceStockDraw,
  };
}

function LiveRegion({ state, t }: TongitsContext) {
  const isRoundEnd = state.phase === TongitsPhase.ROUND_END;
  const isGameEnd = state.phase === TongitsPhase.GAME_END || state.gameEndFlag;
  const winner =
    state.winnerIdx >= 0 && state.winnerIdx < state.players.length ? state.players[state.winnerIdx] : undefined;

  return (
    <>
      <div data-testid="tongits-round-score-announcement">
        <RoundScoreAnnouncement
          active={isRoundEnd || isGameEnd}
          entries={state.players.map((p) => ({
            name: playerName(p.id, p.isHuman),
            roundScore: p.roundScore,
            cumulativeScore: p.cumulativeScore,
          }))}
          separator={t('listSeparator')}
        />
      </div>
      <TongitsOnDealCelebration
        show={!!state.isTongits && (isRoundEnd || isGameEnd) && winner !== undefined}
        winnerCards={winner?.cards ?? []}
        winnerName={winner ? playerName(winner.id, winner.isHuman) : undefined}
      />
    </>
  );
}

function Board({ game, state, t, view, busy }: TongitsContext) {
  const isDrawPhase = state.phase === TongitsPhase.DRAW;
  const isRoundEnd = state.phase === TongitsPhase.ROUND_END;
  const isGameEnd = state.phase === TongitsPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn =
    (isDrawPhase || state.phase === TongitsPhase.DISCARD) && state.players[state.currentPlayerIdx]?.isHuman === true;
  const messageBox = (
    <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
  );

  return (
    <>
      <div
        className="sr-only"
        role="status"
        aria-live="polite"
        aria-atomic="true"
        data-testid="tongits-draw-pile-announcement"
      >
        {view.drawPileAnnouncement && <span>{view.drawPileAnnouncement}</span>}
      </div>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span>{t('drawPile', { count: state.drawPileCount })}</span>
      </div>

      <div className={lgTwoColGrid}>
        <div>
          {state.discardTop &&
            (() => {
              // The discard pile top card doubles as a draw target: on the human's
              // draw turn, clicking it draws from the discard pile (same action as the
              // footer "draw from discard" button). Outside that phase it stays inert
              // but focusable-disabled so it never traps keyboard/tap focus.
              const canDrawDiscard = isDrawPhase && isHumanTurn;
              return (
                <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
                  <button
                    type="button"
                    onClick={game.handleDrawDiscard}
                    disabled={!canDrawDiscard || busy}
                    aria-label={t('drawDiscardCardLabel', { card: cardAlt(state.discardTop) })}
                    data-testid="tongits-discard-pile"
                    className={`transition-transform ${focusRingCard} ${
                      canDrawDiscard
                        ? 'cursor-pointer ring-2 ring-ds-info motion-safe:hover:scale-105'
                        : 'cursor-default'
                    }`}
                    style={{ background: 'none', padding: 0, border: 'none' }}
                  >
                    <AnimatedCard card={state.discardTop} width={view.cardWidth} />
                  </button>
                  <div className="text-ds-text-muted text-sm">
                    <div>{t('discardTop')}</div>
                  </div>
                </div>
              );
            })()}

          {state.players.flatMap((p) => p.melds.map((meld, meldIdx) => ({ p, meld, meldIdx }))).length > 0 && (
            <div className="my-3 p-2 rounded bg-black/30" data-testid="tongits-melds">
              <div className="text-ds-text-muted text-sm mb-1">{t('tableMelds')}</div>
              {state.players
                .flatMap((p) => p.melds.map((meld, meldIdx) => ({ p, meld, meldIdx })))
                .map(({ p, meld, meldIdx }) => (
                  <div key={publicMeldRowKey(p.id, meldIdx)} className="flex flex-wrap gap-1 mb-1">
                    <span className="text-ds-text-muted text-xs w-full">{playerName(p.id, p.isHuman)}</span>
                    {meld.cards.map((card, cardIdx) => (
                      <AnimatedCard
                        key={`opp-meld-${meldIdx}-${card.design}-${card.value}-${cardIdx}`}
                        card={card}
                        width={view.cardWidth * 0.7}
                      />
                    ))}
                  </div>
                ))}
            </div>
          )}
        </div>

        <div>
          {state.players
            .filter((p) => !p.isHuman)
            .map((p) => (
              <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                <div className="text-ds-text-muted text-sm">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })}
                </div>
                {(isRoundEnd || isGameEnd) && p.cards.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {p.cards.map((card, idx) => (
                      <AnimatedCard
                        key={`cpu-${card.design}-${card.value}-${idx}`}
                        card={card}
                        width={view.cardWidth * 0.8}
                      />
                    ))}
                  </div>
                )}
              </div>
            ))}

          <div className="my-3 p-2 rounded bg-black/30" data-tutorial="tongits-score-table">
            {(isRoundEnd || isGameEnd) && (
              <div data-testid="tongits-round-result" className="mb-2">
                <div>
                  {t('roundEndReasonLabel')}: {t(`roundEndReason.${state.roundEndReason}`)}
                </div>
                <div>
                  {t('roundWinnerLabel')}:{' '}
                  {state.roundWinner < 0
                    ? t('roundDraw')
                    : playerName(state.roundWinner, state.players[state.roundWinner]?.isHuman ?? false)}
                </div>
              </div>
            )}
            <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
            <table className="w-full text-sm text-ds-text-muted">
              <caption className="sr-only">{t('scoresCaption')}</caption>
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

      {view.isMobile && messageBox}
    </>
  );
}

function FooterStatus({ frontendHint, t }: TongitsContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

function FooterHand({ game, state, view }: TongitsContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const isDiscardPhase = state.phase === TongitsPhase.DISCARD;
  const isHumanTurn =
    (state.phase === TongitsPhase.DRAW || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  // When exactly three cards are selected, highlight the selected meld candidates.
  const meldHighlight = (() => {
    if (!humanPlayer || !isDiscardPhase || !isHumanTurn || game.selectedCardIndices.length < 3) {
      return new Set<number>();
    }
    const selected = humanPlayer.cards
      .map((card, idx) => ({ card, idx }))
      .filter((x) => game.selectedCardIndices.includes(x.idx));
    const meldedPositions = tongitsMeldIndices(selected.map((x) => x.card));
    const result = new Set<number>();
    for (const pos of meldedPositions) result.add(selected[pos].idx);
    return result;
  })();

  if (view.isMobile) {
    return (
      <MobileHandGrid
        cards={humanPlayer.cards}
        selectedIndices={game.selectedCardIndices}
        onToggle={game.toggleCard}
        cardWidth={view.cardWidth}
        dataTutorial="tongits-player-hand"
        stackRows
      />
    );
  }

  return (
    <div className="flex flex-wrap gap-1 mb-2" data-tutorial="tongits-player-hand">
      {humanPlayer.cards.map((card, idx) => {
        const isCardSelected = game.selectedCardIndices.includes(idx);
        const isMeldCard = !isCardSelected && meldHighlight.has(idx);
        return (
          <button
            type="button"
            key={`${card.design}-${card.value}-${idx}`}
            onClick={() => game.toggleCard(idx)}
            aria-label={cardAlt(card)}
            aria-pressed={isCardSelected}
            data-meld={isMeldCard ? 'true' : undefined}
            data-testid={`tongits-hand-${idx.toString()}`}
            className={`transition-transform ${focusRingCard}`}
            style={{
              background: 'none',
              padding: 0,
              borderRadius: 8,
              // Selection wins; otherwise green-highlight cards that already form a meld.
              ...(isCardSelected
                ? selectedCardStyle(true)
                : isMeldCard
                  ? playableCardStyle(true)
                  : selectedCardStyle(false)),
              boxSizing: 'border-box',
            }}
          >
            <AnimatedCard card={card} width={view.cardWidth} />
          </button>
        );
      })}
    </div>
  );
}

function primaryActions({ game, state, t, busy, view }: TongitsContext) {
  const isDrawPhase = state.phase === TongitsPhase.DRAW;
  const isDiscardPhase = state.phase === TongitsPhase.DISCARD;
  const isRoundEnd = state.phase === TongitsPhase.ROUND_END;
  const isHumanTurn = (isDrawPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
  // The hand-wired page always rendered this row, even empty on CPU turns, so the footer keeps its height.

  const canChallenge = state.remainingPoints >= 0;

  return (
    <div className="flex gap-2 items-center flex-wrap">
      {isDrawPhase && isHumanTurn && (
        <div className="flex gap-2" data-tutorial="tongits-draw-area">
          <button type="button" className={btnPrimary} onClick={view.announceStockDraw} disabled={busy}>
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
            data-tutorial="tongits-discard-button"
          >
            {t('discardButton')}
          </button>
          {state.remainingPoints >= 0 && (
            <span
              className={`self-center text-xs ${canChallenge ? 'text-ds-success' : 'text-ds-text-muted'}`}
              role="status"
              aria-live="polite"
              data-testid="tongits-remaining-points"
              data-challengeable={canChallenge ? 'true' : undefined}
            >
              {t('remainingPoints', { value: state.remainingPoints })}{' '}
              {canChallenge ? t('challengeable') : t('notChallengeable')}
            </span>
          )}
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleMeld}
            disabled={busy || game.selectedCardIndices.length < 3}
          >
            {t('meldButton')}
          </button>
          {state.players
            .flatMap((p) => p.melds.map((_, meldIdx) => ({ p, meldIdx })))
            .map(({ p, meldIdx }) => (
              <button
                key={`sapaw-${p.id}-${meldIdx}`}
                type="button"
                className={btnPrimary}
                onClick={() => game.handleSapaw(p.id, meldIdx)}
                disabled={busy || game.selectedCardIndices.length !== 1}
              >
                {t('sapawButton', { player: playerName(p.id, p.isHuman), meld: meldIdx + 1 })}
              </button>
            ))}
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleChallenge}
            disabled={busy || !canChallenge}
            data-tutorial="tongits-challenge-button"
            data-challengeable={canChallenge ? 'true' : undefined}
          >
            {t('challengeButton')}
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

const tongitsDefinition: GameDefinition<
  TongitsResponse,
  TongitsGame,
  Parameters<typeof tongitsApi.exec>,
  TongitsView
> = {
  key: 'tongits',
  useGame: useTongitsGame,
  useView: useTongitsView,
  tutorial: TONGITS_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 1, centerCard: true, footerHandSize: 5 },
  cli: ({ frontendHint }) => ({
    parseCommand: parseTongitsCommand,
    formatResponse: formatTongitsState,
    helpText: TONGITS_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase] ?? '',
    ended: state.phase === TongitsPhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ state }) => ({
    isHumanTurn:
      (state.phase === TongitsPhase.DRAW || state.phase === TongitsPhase.DISCARD) &&
      state.players[state.currentPlayerIdx]?.isHuman === true,
  }),
  reset: ({ game }) => {
    void game.exec('reset', undefined, {
      cpuDifficulty: game.tongitsConfig.cpuDifficulty,
      pointLimit: game.tongitsConfig.pointLimit,
    });
  },
  resetTutorialId: 'tongits-reset-button',
  actionsClassName: 'flex gap-2 items-center flex-wrap',
  settingsPlacement: 'top',
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.tongitsConfig.cpuDifficulty,
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
          value: game.tongitsConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('pointLimit', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
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
    const isDiscardPhaseForKbd = state.phase === TongitsPhase.DISCARD;
    const isHumanTurnForKbd = isDiscardPhaseForKbd && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: () => {
        if (state.phase === TongitsPhase.DISCARD) {
          game.handleDiscard();
        }
      },
      onClear: game.clearSelection,
      enabled: isHumanTurnForKbd,
    };
  },
  shortcutsTestId: 'tongits-kbd-shortcuts',
  LiveRegion,
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Tongits game page with draw, meld, sapaw, discard, and challenge actions. */
export const TongitsPage = createGamePage(tongitsDefinition);
