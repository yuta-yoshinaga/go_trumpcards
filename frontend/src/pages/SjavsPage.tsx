import type { sjavsApi } from '../api/gameApi';
import { CardBackFan } from '../components/CardBackFan';
import { CardBack } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSjavsGame } from '../hooks/useSjavsGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { SjavsResponse } from '../types/card';
import { SjavsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSjavsCommand, SJAVS_HELP } from '../utils/cli/commands/sjavsCommands';
import { formatSjavsState } from '../utils/cli/formatters/sjavsFormatter';
import { playerName } from '../utils/playerUtils';

function getUnplayableReason(bidding: boolean, handOver: boolean, ended: boolean): string {
  if (bidding) return 'cardReason.bidding';
  if (handOver) return 'cardReason.handEnd';
  if (ended) return 'cardReason.gameEnd';
  return 'cardReason.follow';
}

const SJ_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sj-rule"]', messageKey: 'tutorial.trumps', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sj-bid"]', messageKey: 'tutorial.bid', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sj-hand"]', messageKey: 'tutorial.follow', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sj-score"]', messageKey: 'tutorial.rubber', placement: 'bottom', advanceOn: 'next' },
];

const SUIT_GLYPHS = ['♠', '♣', '♥', '♦'];

const SJAVS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SjavsPhase.BID]: 'bid',
  [SjavsPhase.PLAY]: 'play',
  [SjavsPhase.HAND_END]: 'handEnd',
  [SjavsPhase.GAME_END]: 'end',
};

type SjavsGame = ReturnType<typeof useSjavsGame>;
type SjavsView = ReturnType<typeof useSjavsView>;
type SjavsContext = GamePageContext<SjavsResponse, SjavsGame, SjavsView>;

function useSjavsView() {
  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();
  const phaseNames = usePhaseNames('sjavs', SJAVS_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function HeaderExtra({ state, t }: SjavsContext) {
  return (
    <span className="text-sm text-ds-text-muted" data-tutorial="sj-score">
      {t('trump')}: {state.trumpSuit >= 0 ? SUIT_GLYPHS[state.trumpSuit] : t('trumpUndecided')}
      {state.trumpCount > 0 && ` (${t('trumpCount', { n: state.trumpCount })})`}
      {' / '}
      {t('remaining')}: {t('team', { n: 0 })} {state.remaining[0]} · {t('team', { n: 1 })} {state.remaining[1]}
    </span>
  );
}

function Board({ game, state, t, frontendHint, view }: SjavsContext) {
  const { cardWidth, isMobile } = view;
  const ended = state.phase === SjavsPhase.GAME_END;
  const bidding = state.phase === SjavsPhase.BID;
  const handOver = state.phase === SjavsPhase.HAND_END;
  const handResult = state.handResult;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  // Playability comes from the server, which owns the following rule -- and in
  // this game "trump" is not the same thing as "a card of the trump suit".
  const playable = new Set(state.validIndices);
  const trumps = new Set(state.trumpIndices);
  const unplayableReason = getUnplayableReason(bidding, handOver, ended);
  const reasonApplies = bidding || handOver || ended || isHumanTurn;
  const hasUnplayableCard = (human?.cards ?? []).some((_, i) => !isHumanTurn || bidding || !playable.has(i));
  const unplayableReasonId = `sjavs-${unplayableReason.slice('cardReason.'.length)}-reason`;

  return (
    <>
      {/* Permanent, not tutorial-only: believing that only the trump suit
          is trump is the standard mistake in this game. */}
      <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="sj-rule">
        {t('ruleLine')}
      </div>

      <div className="text-center text-xs text-ds-text-muted mb-3">
        {t('handPoints', { a: state.teamPoints[0], b: state.teamPoints[1] })}
      </div>

      <div className="flex justify-center gap-4 mb-3 flex-wrap">
        {opponents.map((o) => (
          <div key={`opp-${o.id.toString()}`} className="text-center">
            <div className="text-game-text-muted text-xs mb-1">
              {t('opponentHand', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
              {' · '}
              {t('team', { n: o.team })}
              {o.bid > 0 && ` · ${t('bidMark', { n: o.bid })}`}
            </div>
            <div
              className="flex gap-1 justify-center flex-wrap"
              role="img"
              aria-label={t('opponentHandAriaLabel', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
            >
              {isMobile ? (
                Array.from({ length: o.cardCount }, (_, i) => (
                  <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                ))
              ) : (
                <CardBackFan count={o.cardCount} width={Math.round(cardWidth * 0.45)} />
              )}
            </div>
          </div>
        ))}
      </div>

      {!bidding && (
        <div className="text-center mb-4">
          {state.trick.length === 0 ? (
            <>
              <div className="text-game-text-muted text-xs mb-1">{t('trick')}</div>
              <div className="flex gap-1 justify-center flex-wrap">
                <span className="text-game-text-muted text-xs">—</span>
              </div>
            </>
          ) : (
            <TrickDisplay
              currentTrick={state.trick}
              players={state.players}
              cardWidth={cardWidth}
              label={t('trick')}
              cardAriaLabelFor={(player, playedCard) =>
                t('trickCardAriaLabel', {
                  player: playerName(player.id, player.isHuman),
                  card: cardAlt(playedCard),
                })
              }
            />
          )}
        </div>
      )}

      {handOver && state.handResult && (
        <div className="text-center text-sm mb-3" data-testid="sjavs-hand-result">
          {state.handResult.scoringTeam < 0
            ? t('handTie')
            : `${t('handScore', {
                team: state.handResult.scoringTeam,
                amount: state.handResult.amount,
              })}${state.handResult.vol ? t('vol') : ''}`}
        </div>
      )}

      <RoundScoreAnnouncement
        active={handOver && Boolean(handResult)}
        entries={[0, 1].map((team) => ({
          name: t('team', { n: team }),
          roundScore: handResult?.scoringTeam === team ? handResult.amount : 0,
          cumulativeScore: state.rubberPoints - state.remaining[team],
        }))}
        separator={t('listSeparator')}
      />

      <div className="text-center" data-tutorial="sj-hand">
        <div className="text-game-text-muted text-xs mb-1">
          {t('yourHand')}
          {' · '}
          {t('team', { n: human?.team ?? 0 })}
          {(human?.bid ?? 0) > 0 && ` · ${t('bidMark', { n: human?.bid ?? 0 })}`}
        </div>
        <div className="flex gap-1 justify-center flex-wrap">
          {(human?.cards ?? []).map((card, i) => {
            const canPlay = isHumanTurn && !bidding && playable.has(i);
            const isTrump = trumps.has(i);
            return (
              <button
                key={`hand-${i.toString()}`}
                type="button"
                data-hint-action="play"
                // Kept focusable while it cannot act so the reason is
                // announced rather than the control leaving the tab order.
                aria-disabled={!canPlay}
                aria-describedby={!canPlay && reasonApplies ? unplayableReasonId : undefined}
                onClick={() => canPlay && game.handlePlay(i)}
                className={[
                  'rounded transition-transform',
                  canPlay ? 'hover:-translate-y-2' : 'opacity-60',
                  frontendHint.hintEnabled && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                ].join(' ')}
              >
                <AnimatedCard card={card} width={cardWidth} draggable={false} />
                {/* **常時切札の 6 枚 (♣Q ♠Q ♣J ♠J ♥J ♦J) はスートを見ても
                    分からない。**規則文は出ているのに、手札のどれがそれかは
                    暗記に頼らせていた (#5575)。添字はサーバが
                    SjavsIsTrump から作るので、6 枚の一覧を持たない。 */}
                {isTrump && (
                  <span
                    className="block text-[10px] text-ds-warning leading-none"
                    data-testid={`sjavs-trump-${i.toString()}`}
                  >
                    {t('trumpMark')}
                  </span>
                )}
              </button>
            );
          })}
        </div>
        {reasonApplies && hasUnplayableCard ? (
          <span id={unplayableReasonId} className="sr-only">
            {t(unplayableReason)}
          </span>
        ) : null}
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {ended && state.doubleVictory && (
        <div className="text-center text-ds-warning text-sm font-semibold mb-3" data-testid="sjavs-double-victory">
          {t('doubleVictory')}
        </div>
      )}
    </>
  );
}

function FooterStatus({ game, state, t, frontendHint }: SjavsContext) {
  const bidding = state.phase === SjavsPhase.BID;
  const ended = state.phase === SjavsPhase.GAME_END;
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  if (!bidding || !isHumanTurn) return null;

  // Every length the human can legally bid. Below minBid you must pass, and you
  // can never bid more than you hold.
  const bidChoices: number[] = [];
  for (let n = state.minBid; n <= state.myLongest; n++) {
    bidChoices.push(n);
  }

  // CUI は `sjavs.hintBid` で推奨するビッド長を明示しているのに、Web はビッド
  // ボタンを素で並べていて、シャウス最大の判断点に手掛かりが無かった (#4883)。
  // 0 は「パスを推奨」(sjavs.hint.pass) なので、パスのボタンに印が付く。
  // 判定が `frontendHint.hintEnabled` だけなのは意図的 — このゲームの hint は
  // ゲーム追加時 (#4403) から Output に乗る常時表示で、`isRequestedHint` を
  // 通すと恒久的に false になる。check-hint-gate の ALLOWED に登録済み。
  const hintBidLength = frontendHint.hintEnabled ? state.hint?.bidLength : undefined;

  return (
    <div className="mb-2" data-tutorial="sj-bid">
      <div className="text-xs text-ds-text-muted mb-1">
        {t('bidPrompt', { min: state.minBid, longest: state.myLongest })}
      </div>
      <div className="flex gap-2 flex-wrap items-center">
        {bidChoices.map((n) => (
          <button
            key={`bid-${n.toString()}`}
            type="button"
            data-hint-action="bid"
            className={[btnPrimary, 'min-h-11', hintBidLength === n ? 'ring-2 ring-ds-warning' : ''].join(' ')}
            data-hint-bid={hintBidLength === n ? 'true' : undefined}
            onClick={() => game.handleBid(n)}
          >
            {t('bid', { n })}
          </button>
        ))}
        {/* Pass is always available -- and it is a bid of zero, not
            an omitted parameter. */}
        <button
          type="button"
          data-hint-action="bid"
          className={[btnSecondary, 'min-h-11', hintBidLength === 0 ? 'ring-2 ring-ds-warning' : ''].join(' ')}
          data-hint-bid={hintBidLength === 0 ? 'true' : undefined}
          onClick={() => game.handleBid(0)}
        >
          {t('pass')}
        </button>
      </div>
    </div>
  );
}

function FooterActions({ game, state, t }: SjavsContext) {
  const handOver = state.phase === SjavsPhase.HAND_END;
  const ended = state.phase === SjavsPhase.GAME_END;
  if (!handOver || ended) return null;
  return (
    <button type="button" className={`${btnPrimary} min-h-11`} onClick={game.handleNextHand}>
      {t('nextHand')}
    </button>
  );
}

const sjavsDefinition: GameDefinition<SjavsResponse, SjavsGame, Parameters<typeof sjavsApi.exec>, SjavsView> = {
  key: 'sjavs',
  useGame: useSjavsGame,
  useView: useSjavsView,
  tutorial: SJ_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 3, tableau: 4 },
  cli: () => ({
    parseCommand: parseSjavsCommand,
    formatResponse: formatSjavsState,
    helpText: SJAVS_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === SjavsPhase.GAME_END,
  }),
  actionLogDuringPlay: true,
  settingsPlacement: 'top',
  settingsInCliMode: true,
  landscapeBanner: true,
  landscapeBannerInCliMode: true,
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  resetTutorialId: 'sj-reset-button',
  HeaderExtra,
  Board,
  FooterStatus,
  FooterActions,
};

/** Renders the Sjavs page: six permanent trumps, a length bid, and a rubber counted down from 24. */
export const SjavsPage = createGamePage(sjavsDefinition);
