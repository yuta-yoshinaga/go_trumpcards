import type { TFunction } from 'i18next';
import { useEffect, useRef, useState } from 'react';
import type { toepenApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useToepenGame } from '../hooks/useToepenGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { ToepenResponse } from '../types/card';
import { ToepenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseToepenCommand, TOEPEN_HELP } from '../utils/cli/commands/toepenCommands';
import { formatToepenState } from '../utils/cli/formatters/toepenFormatter';

const TOEPEN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="toepen-ranking"]',
    messageKey: 'tutorial.ranking',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="toepen-hand"]', messageKey: 'tutorial.follow', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="toepen-lives"]',
    messageKey: 'tutorial.lastTrick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="toepen-controls"]', messageKey: 'tutorial.toep', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="toepen-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

interface ToepenView {
  lifeAnnouncement: string;
  cardWidth: number;
  phaseNames: {
    end: string;
    handEnd: string;
    play: string;
  };
}

type ToepenGame = ReturnType<typeof useToepenGame>;
type ToepenContext = GamePageContext<ToepenResponse, ToepenGame, ToepenView>;

function useToepenView({ game, t }: { game: ToepenGame; t: TFunction }): ToepenView {
  const { state } = game;
  const previousLives = useRef<Map<number, { lives: number; eliminated: boolean }> | null>(null);
  const [lifeAnnouncement, setLifeAnnouncement] = useState('');
  useEffect(() => {
    if (!state) return;
    const previous = previousLives.current;
    const changes = previous
      ? state.players.flatMap((player) => {
          const old = previous.get(player.id);
          if (!old) return [];
          if (!old.eliminated && player.eliminated)
            return [t('lifeEliminated', { name: player.isHuman ? t('you') : t('cpu', { n: player.id }) })];
          if (old.lives === player.lives) return [];
          const name = player.isHuman ? t('you') : t('cpu', { n: player.id });
          return [t(player.lives < old.lives ? 'lifeLost' : 'lifeGained', { name, lives: player.lives })];
        })
      : [];
    previousLives.current = new Map(
      state.players.map((player) => [player.id, { lives: player.lives, eliminated: player.eliminated }]),
    );
    setLifeAnnouncement(changes.join(t('listSeparator')));
  }, [state, t]);

  const { cardWidth } = useCardDimensions();
  const phaseNames = {
    end: t('phase.end'),
    handEnd: t('phase.handEnd'),
    play: t('phase.play'),
  };

  return { lifeAnnouncement, cardWidth, phaseNames };
}

function LiveRegion({ view }: ToepenContext) {
  return (
    <div role="status" aria-live="polite" aria-atomic="true" className="sr-only">
      {view.lifeAnnouncement}
    </div>
  );
}

function HeaderExtra({ state, t }: ToepenContext) {
  return (
    <span className="text-sm text-ds-text-muted">
      {t('hand', { n: state.handNumber })} / {t('stake')}: {state.stake}
    </span>
  );
}

function Board({ game, state, t, view, frontendHint }: ToepenContext) {
  const { cardWidth } = view;
  const ended = state.phase === ToepenPhase.GAME_END;
  const handOver = state.phase === ToepenPhase.HAND_END;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && !handOver && state.phase === ToepenPhase.PLAY && state.currentPlayerIdx === 0;

  // Legality comes from the server, which already applies the follow-suit
  // obligation. Deriving it here would be a second implementation of one rule.
  const playable = new Set(state.validPlayIndices);

  return (
    <>
      {/* Shown permanently: this ranking is inverted from every other
          game here and is the single easiest thing to misplay. */}
      <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="toepen-ranking">
        {t('ranking')}
      </div>

      <div className="flex justify-center gap-4 mb-3 text-sm flex-wrap" data-tutorial="toepen-lives">
        {state.players.map((p) => (
          <span key={`lv-${p.id.toString()}`} className={p.eliminated ? 'opacity-50 line-through' : ''}>
            {p.isHuman ? t('you') : t('cpu', { n: p.id })}:{' '}
            <span className="font-bold">
              {p.lives}/{state.maxLives}
            </span>
            {p.folded && !p.eliminated && <span className="text-ds-text-muted ml-1">{t('folded')}</span>}
          </span>
        ))}
      </div>

      <div className="flex justify-center gap-4 mb-3">
        {opponents.map((o) => {
          const revealed = ended && o.cards.length > 0;
          return (
            <div key={`opp-${o.id.toString()}`} className="text-center">
              <div className="text-game-text-muted text-xs mb-1">{t('cpuHand', { n: o.id, count: o.cardCount })}</div>
              {revealed ? (
                <fieldset
                  className="m-0 flex gap-1 justify-center border-0 p-0"
                  aria-label={t('cpuHandRevealedAriaLabel', { n: o.id })}
                >
                  {o.cards.map((card, i) => (
                    <CardImage
                      key={`opp-${o.id.toString()}-c${i.toString()}`}
                      card={card}
                      ariaLabel={cardAlt(card)}
                      width={cardWidth}
                    />
                  ))}
                </fieldset>
              ) : (
                <div
                  className="flex gap-1 justify-center"
                  role="img"
                  aria-label={t('cpuHandAriaLabel', { n: o.id, count: o.cardCount })}
                >
                  {Array.from({ length: o.cardCount }, (_, i) => (
                    <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </div>

      <div className="text-center mb-4 min-h-24">
        <div className="text-game-text-muted text-xs mb-1">
          {state.currentTrick.length > 0 ? t('trick', { n: state.trickNumber + 1 }) : t('noTrick')}
        </div>
        <div className="flex gap-2 justify-center" data-testid="toepen-current-trick">
          {state.currentTrick.map((tc2, i) => {
            const player = state.players[tc2.playerIdx];
            if (!player) return null;
            const playerName = player.isHuman ? t('you') : t('cpu', { n: player.id });
            const isWinningCard = state.currentTrickWinnerIdx === tc2.playerIdx;
            return (
              <div
                key={`trick-${i.toString()}`}
                className={`flex flex-col items-center gap-1 ${isWinningCard ? 'rounded ring-2 ring-ds-warning p-1' : ''}`}
                data-testid="toepen-trick-card"
                data-winning={isWinningCard}
              >
                <div
                  role="img"
                  aria-label={isWinningCard ? `${cardAlt(tc2.card)} (${t('currentTrickWinner')})` : cardAlt(tc2.card)}
                >
                  <AnimatedCard card={tc2.card} width={cardWidth} draggable={false} />
                </div>
                <span className="text-xs text-ds-text-primary">{playerName}</span>
                {player.folded && <span className="text-xs text-ds-text-muted">{t('folded')}</span>}
              </div>
            );
          })}
        </div>
      </div>

      <div className="text-center" data-tutorial="toepen-hand">
        <div className="text-game-text-muted text-xs mb-1">{t('yourHand')}</div>
        <div className="flex gap-1 justify-center flex-wrap">
          {(human?.cards ?? []).map((card, i) => {
            const canPlay = isHumanTurn && playable.has(i);
            // Blocked only by the follow-suit duty; off-turn cards are not
            // "wrong", they are just not yours to play yet.
            const blockedBySuit = isHumanTurn && !playable.has(i);
            return (
              <button
                key={`hand-${i.toString()}`}
                type="button"
                data-hint-action="play"
                data-testid={`toepen-hand-${i.toString()}`}
                // Kept focusable while it cannot act so the reason is
                // announced rather than the control leaving the tab order.
                aria-disabled={!canPlay}
                title={blockedBySuit ? t('followSuitTooltip') : undefined}
                aria-label={blockedBySuit ? `${cardAlt(card)} (${t('followSuitAria')})` : cardAlt(card)}
                onClick={() => canPlay && game.handlePlay(i)}
                className={[
                  'rounded transition-transform',
                  canPlay ? 'hover:-translate-y-2' : 'opacity-50',
                  frontendHint.hintEnabled && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                ].join(' ')}
              >
                <AnimatedCard card={card} width={cardWidth} draggable={false} />
              </button>
            );
          })}
        </div>
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {handOver && state.lastTrickWinner >= 0 && (
        <div className="mt-3 text-center text-ds-warning font-bold" data-testid="toepen-last-trick-winner">
          {t('lastTrickWinner', {
            // 席名は既存の出し方（137行）に合わせて **id** を使う。
            // 配列インデックスを渡すと id とズレたときに別の席を名乗る。
            name: state.players[state.lastTrickWinner]?.isHuman
              ? t('you')
              : t('cpu', { n: state.players[state.lastTrickWinner]?.id }),
          })}
        </div>
      )}
    </>
  );
}

function FooterActions({ game, state, t, busy }: ToepenContext) {
  const ended = state.phase === ToepenPhase.GAME_END;
  const handOver = state.phase === ToepenPhase.HAND_END;
  const answering = state.phase === ToepenPhase.RESPOND && state.pendingRespondent === 0;
  const human = state.players.find((p) => p.isHuman);
  const canToep = !ended && !handOver && state.phase === ToepenPhase.PLAY && !human?.folded;

  return (
    <>
      {answering && (
        <>
          {/* **誰の判断に応答しているのか。** knockerIdx はサーバから届いて
              いるのに読んでおらず、賭け点しか出していなかった。相手が複数
              いると誰に応答するのか分からない (#5570)。CUI の respondLine は
              最初から名前を出している。

              応答者が toep 宣言者本人になることはない -- domain の
              nextRespondent が knockerIdx に当たった時点で -1 を返す。
              つまりこのブロックが出ているとき knockerIdx は必ず CPU 席で、
              「あなた」の分岐は書いても到達しない。 */}
          <span className="text-sm text-ds-text-muted" data-testid="toepen-toeped-by">
            {state.knockerIdx >= 0
              ? t('toepedBy', { name: t('cpu', { n: state.knockerIdx }), stake: state.stake })
              : t('toepedAt', { stake: state.stake })}
          </span>
          <button
            type="button"
            className={btnPrimary}
            data-hint-action="stay"
            onClick={() => game.handleRespond(true)}
            disabled={busy}
          >
            {t('stay')}
          </button>
          <button
            type="button"
            className={btnSecondary}
            data-hint-action="fold"
            onClick={() => game.handleRespond(false)}
            disabled={busy}
            // Folding costs the stake BEFORE the raise, so the number
            // shown here is deliberately one less than `stake`.
            title={t('foldCost', { cost: state.stake - 1 })}
          >
            {t('fold', { cost: state.stake - 1 })}
          </button>
        </>
      )}

      {state.canRedeal && (
        <>
          <span className="text-sm text-ds-text-muted" data-testid="toepen-redeal-hint">
            {t('redealHint')}
          </span>
          <button
            type="button"
            className={btnSecondary}
            onClick={game.handleRedeal}
            disabled={busy}
            title={t('redealHint')}
          >
            {t('redeal')}
          </button>
        </>
      )}

      {canToep && (
        <button type="button" className={btnSecondary} onClick={game.handleToep} disabled={busy}>
          {t('toep')}
        </button>
      )}

      {handOver && !ended && (
        <button type="button" className={btnPrimary} onClick={game.handleNextHand} disabled={busy}>
          {t('nextHand')}
        </button>
      )}
    </>
  );
}

const toepenDefinition: GameDefinition<ToepenResponse, ToepenGame, Parameters<typeof toepenApi.exec>, ToepenView> = {
  key: 'toepen',
  useGame: useToepenGame,
  useView: useToepenView,
  tutorial: TOEPEN_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 4, tableau: 4 },
  cli: () => ({
    parseCommand: parseToepenCommand,
    formatResponse: formatToepenState,
    helpText: TOEPEN_HELP,
  }),
  phase: ({ state, view }) => {
    const ended = state.phase === ToepenPhase.GAME_END;
    const handOver = state.phase === ToepenPhase.HAND_END;
    return {
      name: ended ? view.phaseNames.end : handOver ? view.phaseNames.handEnd : view.phaseNames.play,
      ended,
    };
  },
  actionLogEndPhase: ({ state }) => state.phase === ToepenPhase.GAME_END || state.phase === ToepenPhase.HAND_END,
  settingsPlacement: 'top',
  settingsInCliMode: true,
  landscapeBanner: true,
  landscapeBannerInCliMode: true,
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  actionsTutorialId: 'toepen-controls',
  resetTutorialId: 'toepen-reset-button',
  LiveRegion,
  HeaderExtra,
  Board,
  FooterActions,
};

/** Renders the Toepen page: 32 cards, inverted ranking, only the last trick pays. */
export const ToepenPage = createGamePage(toepenDefinition);
