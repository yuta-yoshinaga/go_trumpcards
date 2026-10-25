import type { buraApi } from '../api/gameApi';
import { CardBackFan } from '../components/CardBackFan';
import { CardBack, CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { useBuraGame } from '../hooks/useBuraGame';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { BuraResponse } from '../types/card';
import { BuraPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BURA_HELP, parseBuraCommand } from '../utils/cli/commands/buraCommands';
import { formatBuraState } from '../utils/cli/formatters/buraFormatter';
import { findPlayerName } from '../utils/playerUtils';

const BURA_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bura-trump"]', messageKey: 'tutorial.trump', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bura-hand"]', messageKey: 'tutorial.multiCard', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="bura-table"]', messageKey: 'tutorial.beating', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bura-points"]', messageKey: 'tutorial.claim', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bura-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

const BURA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BuraPhase.PLAY]: 'play',
  [BuraPhase.GAME_END]: 'end',
};

type BuraGame = ReturnType<typeof useBuraGame>;
type BuraView = ReturnType<typeof useBuraView>;
type BuraContext = GamePageContext<BuraResponse, BuraGame, BuraView>;

function useBuraView() {
  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();
  const phaseNames = usePhaseNames('bura', BURA_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function HeaderExtra({ state, t }: BuraContext) {
  return (
    <span className="text-sm text-ds-text-muted">
      {t('trick')}: {state.trickNumber} / {t('stock')}: {state.stockRemaining}
    </span>
  );
}

function Board({ game, state, t, frontendHint, view }: BuraContext) {
  const { cardWidth, isMobile } = view;
  const ended = state.phase === BuraPhase.GAME_END;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;
  const leadCount = state.currentLead.length;
  const leadPlayerName = leadCount > 0 ? findPlayerName(state.players, state.leadPlayerIdx) : '';
  // The server also ships the exact indices it recommends; the tooltip only
  // carries the reason, so read them off the state response directly.
  const hintedIndices = new Set(state.hint?.cardIndices ?? []);

  return (
    <>
      <div className="text-center mb-3" data-tutorial="bura-trump">
        <div className="text-game-text-muted text-xs mb-1">{t('trump')}</div>
        {state.trumpCard ? (
          <div className="flex justify-center">
            <AnimatedCard card={state.trumpCard} width={cardWidth} draggable={false} />
          </div>
        ) : (
          <div className="text-game-text-muted text-sm">{t('trumpDrawn', { suit: t(`suit.${state.trumpSuit}`) })}</div>
        )}
      </div>

      <div className="flex justify-center gap-6 mb-3 text-sm" data-tutorial="bura-points">
        <span>
          {t('you')}: <span className="font-bold text-ds-warning">{human?.points ?? 0}</span>
        </span>
        <span className="text-game-text-muted">
          {t('target')}: {state.winThreshold}
        </span>
        {opponents.map((o) => (
          <span key={`pts-${o.id.toString()}`}>
            {findPlayerName(state.players, state.players.indexOf(o))}: <span className="font-bold">{o.points}</span>
          </span>
        ))}
      </div>

      <div className="flex justify-center gap-4 mb-4">
        {opponents.map((o) => {
          const opponentName = findPlayerName(state.players, state.players.indexOf(o));
          const revealed = ended && o.cards.length > 0;
          const handCards = revealed ? (
            o.cards.map((card, i) => (
              <CardImage
                key={`opp-${o.id.toString()}-c${i.toString()}`}
                card={card}
                ariaLabel={cardAlt(card)}
                width={cardWidth}
              />
            ))
          ) : isMobile ? (
            Array.from({ length: o.cardCount }, (_, i) => (
              <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
            ))
          ) : (
            <CardBackFan count={o.cardCount} width={Math.round(cardWidth * 0.45)} />
          );
          return (
            <div key={`opp-${o.id.toString()}`} className="text-center">
              <div className="text-game-text-muted text-xs mb-1">
                {t('opponentHand', { name: opponentName, n: o.cardCount })}
              </div>
              {revealed ? (
                <fieldset
                  className="m-0 flex gap-1 justify-center border-0 p-0"
                  aria-label={t('opponentHandRevealedAriaLabel', { name: opponentName })}
                >
                  {handCards}
                </fieldset>
              ) : (
                <div
                  className="flex gap-1 justify-center"
                  role="img"
                  aria-label={t('opponentHandAriaLabel', { name: opponentName, n: o.cardCount })}
                >
                  {handCards}
                </div>
              )}
            </div>
          );
        })}
      </div>

      <div className="text-center mb-4 min-h-24" data-tutorial="bura-table">
        <div className="text-game-text-muted text-xs mb-1">
          {leadCount > 0 ? t('ledCards', { n: leadCount }) : t('noLead')}
        </div>
        {leadCount > 0 && (
          <div className="text-ds-text-primary text-sm mb-1">{t('leadPlayer', { name: leadPlayerName })}</div>
        )}
        <div className="flex gap-1 justify-center">
          {state.currentLead.map((card, i) => (
            <AnimatedCard
              key={`lead-${i.toString()}`}
              card={card}
              ariaLabel={t('leadCardAriaLabel', { name: leadPlayerName, card: cardAlt(card) })}
              width={cardWidth}
              draggable={false}
            />
          ))}
        </div>
      </div>

      <div
        className="text-center lg:sticky lg:bottom-0 lg:z-10 lg:bg-game-bg-casino lg:pt-2 lg:shadow-[0_-8px_16px_rgba(0,0,0,0.35)]"
        data-tutorial="bura-hand"
      >
        <div className="text-game-text-muted text-xs mb-1">{t('yourHand')}</div>
        <div className="flex gap-1 justify-center flex-wrap">
          {(human?.cards ?? []).map((card, i) => {
            const isSelected = game.selected.includes(i);
            return (
              <button
                key={`hand-${i.toString()}`}
                type="button"
                // Kept focusable while it cannot act, so the reason is
                // announced instead of the control vanishing from the
                // tab order.
                aria-disabled={!isHumanTurn}
                aria-pressed={isSelected}
                onClick={() => isHumanTurn && game.toggleCard(i)}
                className={[
                  'rounded transition-transform',
                  isSelected ? '-translate-y-3 ring-2 ring-ds-accent' : '',
                  frontendHint.hintEnabled && hintedIndices.has(i) ? 'ring-2 ring-ds-warning' : '',
                  isHumanTurn ? '' : 'opacity-60',
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
    </>
  );
}

function FooterActions({ game, state, t, busy }: BuraContext) {
  const ended = state.phase === BuraPhase.GAME_END;
  // 役の一覧はサーバが判定順に送ってくる。訳の無いキーが来ても、キー名を
  // そのまま出すより黙って落とす方がまし ── 役の説明として読めないので。
  const declareTitle = [
    t('declareHint'),
    ...state.winningCombinations.map((key) => t(`combo.${key}`, { defaultValue: '' })).filter(Boolean),
  ].join('\n');
  const human = state.players.find((p) => p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;
  const leadCount = state.currentLead.length;

  // A lead may be 1-3 cards of ONE suit; a response has to match the lead's
  // count exactly. Checking here keeps the player from posting a move the
  // server will only reject.
  const selectedCards = game.selected.map((i) => human?.cards[i]).filter((c) => c != null);
  const sameSuit = selectedCards.every((c) => c.design === selectedCards[0]?.design);
  const playReason = !isHumanTurn
    ? ''
    : selectedCards.length === 0
      ? t('playReason.select')
      : leadCount > 0
        ? selectedCards.length < leadCount
          ? t('playReason.responseTooFew', { n: leadCount })
          : selectedCards.length > leadCount
            ? t('playReason.responseTooMany', { n: leadCount })
            : ''
        : !sameSuit
          ? t('playReason.leadSuit')
          : selectedCards.length > 3
            ? t('playReason.leadMax')
            : '';
  const canPlay =
    isHumanTurn &&
    game.selected.length > 0 &&
    (leadCount > 0 ? game.selected.length === leadCount : sameSuit && game.selected.length <= 3);

  return (
    <>
      {!ended && (human?.points ?? 0) < state.winThreshold && (
        <div role="status" className="basis-full rounded bg-ds-warning px-3 py-2 text-sm text-ds-text-on-accent">
          {t('claimWarning')}
        </div>
      )}
      <button
        type="button"
        className={btnPrimary}
        data-hint-action="play"
        onClick={game.handlePlay}
        disabled={busy || !canPlay}
        aria-describedby={playReason ? 'bura-play-reason' : undefined}
      >
        {leadCount > 0 ? t('respond', { n: leadCount }) : t('play')}
      </button>
      <div id="bura-play-reason" role="status" aria-live="polite" className="basis-full text-sm text-ds-text-muted">
        {playReason}
      </div>
      <button
        type="button"
        className={btnSecondary}
        data-hint-action="claim"
        onClick={game.handleClaim}
        disabled={busy || ended}
        title={t('claimWarning')}
      >
        {t('claim', { target: state.winThreshold })}
      </button>
      {/* **何が「役」かはどこにも書かれていなかった。**claim には
          失敗時のリスクが title で出ているのに、declare は押しても
          何も起きない理由が分からないままだった (#5568)。役の一覧は
          サーバから来る ── 画面側で数え直すと、役を足したとき案内だけが
          古くなる。 */}
      <button
        type="button"
        className={btnSecondary}
        data-hint-action="declare"
        onClick={game.handleDeclare}
        disabled={busy || ended}
        title={declareTitle}
      >
        {t('declare')}
      </button>
    </>
  );
}

const buraDefinition: GameDefinition<BuraResponse, BuraGame, Parameters<typeof buraApi.exec>, BuraView> = {
  key: 'bura',
  useGame: useBuraGame,
  useView: useBuraView,
  tutorial: BURA_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 3, tableau: 3 },
  cli: () => ({
    parseCommand: parseBuraCommand,
    formatResponse: formatBuraState,
    helpText: BURA_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === BuraPhase.GAME_END,
  }),
  actionLogDuringPlay: true,
  settingsPlacement: 'top',
  landscapeBanner: true,
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  actionsTutorialId: 'bura-controls',
  resetTutorialId: 'bura-reset-button',
  HeaderExtra,
  Board,
  FooterActions,
};

/** Renders the Bura page: a 36-card trick game won by claiming 31 points. */
export const BuraPage = createGamePage(buraDefinition);
