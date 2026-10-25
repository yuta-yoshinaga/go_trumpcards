import type { TFunction } from 'i18next';
import { useCallback, useMemo, useState } from 'react';
import type { napoleonApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { PartnerRevealFlash } from '../components/PartnerRevealFlash';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { TrickDisplay } from '../components/TrickDisplay';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import {
  CPU_DIFFICULTY_OPTIONS,
  MIN_BID_OPTIONS,
  POINT_LIMIT_OPTIONS,
  useNapoleonGame,
} from '../hooks/useNapoleonGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeInfo, badgeSuccess } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { scoreTableClass } from '../styles/scoreTableStyles';
import type { NapoleonResponse } from '../types/card';
import { NapoleonPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { NAPOLEON_HELP, parseNapoleonCommand } from '../utils/cli/commands/napoleonCommands';
import { formatNapoleonState } from '../utils/cli/formatters/napoleonFormatter';
import { type AdjutantCardOption, buildAdjutantCardRows, isAdjutantCardInHand } from '../utils/napoleonAdjutant';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Napoleon tutorial step definitions. */
const NP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="np-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-trump-declaration"]',
    messageKey: 'tutorial.trumpDeclaration',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-adjutant-info"]',
    messageKey: 'tutorial.adjutantInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-kitty-cards"]',
    messageKey: 'tutorial.kittyCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const NAPOLEON_PHASE_KEYS: Readonly<Record<number, string>> = {
  [NapoleonPhase.BID]: 'bid',
  [NapoleonPhase.TRUMP_DECLARATION]: 'trumpDeclaration',
  [NapoleonPhase.KITTY_EXCHANGE]: 'kittyExchange',
  [NapoleonPhase.PLAY]: 'play',
  [NapoleonPhase.TRICK_END]: 'trickEnd',
  [NapoleonPhase.ROUND_END]: 'roundEnd',
  [NapoleonPhase.GAME_END]: 'gameEnd',
};

const SUIT_KEYS: Record<number, string> = { 1: 'spade', 2: 'club', 3: 'heart', 4: 'diamond' };
const DESIGN_ORDER: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };
const MAX_BID = 17;

function useGame() {
  const game = useNapoleonGame();
  return { ...game, exec: game.apiExec };
}

type NapoleonGame = ReturnType<typeof useGame>;
type NapoleonView = ReturnType<typeof useNapoleonView>;
type NapoleonContext = GamePageContext<NapoleonResponse, NapoleonGame, NapoleonView>;

function useNapoleonView({ game }: { game: NapoleonGame }) {
  const state = game.state;
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const { cardWidth, isMobile } = useCardDimensions();

  const [bidValue, setBidValue] = useState(12);
  const [trumpSuitValue, setTrumpSuitValue] = useState(1);
  // Adjutant designation: the card tapped in the visual picker, carrying the
  // numeric (suit, value) the `trump` action submits (null until chosen).
  const [adjSelection, setAdjSelection] = useState<AdjutantCardOption | null>(null);
  const adjutantRows = useMemo(() => buildAdjutantCardRows(), []);

  const isPlayPhaseForKbd = state?.phase === NapoleonPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;
  const playableIndicesForSelection =
    state?.phase === NapoleonPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman
      ? state.validPlayIndices
      : undefined;
  const hasRestrictedPlayForSelection =
    playableIndicesForSelection != null && playableIndicesForSelection.length < humanCardCountForKbd;
  const canSelectPlayCard = useCallback(
    (idx: number) => !hasRestrictedPlayForSelection || playableIndicesForSelection.includes(idx),
    [hasRestrictedPlayForSelection, playableIndicesForSelection],
  );
  const togglePlayableCard = useCallback(
    (idx: number) => {
      if (canSelectPlayCard(idx)) game.toggleCard(idx);
    },
    [canSelectPlayCard, game.toggleCard],
  );

  const confirmAction = useCallback(() => {
    if (game.selectedCardIndices.length === 1 && canSelectPlayCard(game.selectedCardIndices[0])) game.handlePlay();
  }, [canSelectPlayCard, game.handlePlay, game.selectedCardIndices]);

  const phaseNames = usePhaseNames('napoleon', NAPOLEON_PHASE_KEYS);

  return {
    isLargeDesktop,
    cardWidth,
    isMobile,
    bidValue,
    setBidValue,
    trumpSuitValue,
    setTrumpSuitValue,
    adjSelection,
    setAdjSelection,
    adjutantRows,
    phaseNames,
    humanCardCountForKbd,
    isHumanTurnForKbd,
    hasRestrictedPlayForSelection,
    canSelectPlayCard,
    togglePlayableCard,
    confirmAction,
  };
}

function roleBadge(
  t: TFunction,
  p: { isNapoleon: boolean; isAdjutant: boolean; adjutantRevealed: boolean },
  adjutantRevealed: boolean,
) {
  if (p.isNapoleon) return ` [${t('role.napoleon')}]`;
  if (p.isAdjutant && (p.adjutantRevealed || adjutantRevealed)) return ` [${t('role.adjutant')}]`;
  return '';
}

// GUI-only, as on the hand-wired page: AboveContent unmounts in CLI mode, so the flash neither covers the
// terminal nor misses its reset when the GUI comes back.
function PartnerReveal({ state, t }: NapoleonContext) {
  return (
    <PartnerRevealFlash
      revealed={state.adjutantRevealed}
      partnerName={
        state.adjutantIdx >= 0 && state.players[state.adjutantIdx]
          ? playerName(state.adjutantIdx, state.players[state.adjutantIdx].isHuman)
          : ''
      }
      headline={t('role.adjutantRevealed')}
    />
  );
}

function NapoleonMessageBox({ state }: Pick<NapoleonContext, 'state'>) {
  return <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />;
}

function Board({ game, state, t, tc, view }: NapoleonContext) {
  const { isLargeDesktop, cardWidth, isMobile, adjSelection, setAdjSelection, adjutantRows } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === NapoleonPhase.BID;
  const isTrumpDeclaration = state.phase === NapoleonPhase.TRUMP_DECLARATION;
  const isKittyExchange = state.phase === NapoleonPhase.KITTY_EXCHANGE;
  const isPlayPhase = state.phase === NapoleonPhase.PLAY;
  const isTrickEnd = state.phase === NapoleonPhase.TRICK_END;
  const isRoundEnd = state.phase === NapoleonPhase.ROUND_END;
  const isGameEnd = state.phase === NapoleonPhase.GAME_END || state.gameEndFlag;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanNapoleon = isTrumpDeclaration && state.players[state.napoleonIdx]?.isHuman === true;
  const isHumanExchange = isKittyExchange && state.players[state.napoleonIdx]?.isHuman === true;

  // Napoleon-side face-card progress toward the bid (target). The adjutant's
  // haul is only counted once revealed, to avoid leaking their identity.
  const napoleonFaceProgress = (() => {
    if (!isPlayPhase || state.napoleonIdx < 0) return null;
    const napoleon = state.players[state.napoleonIdx];
    if (!napoleon) return null;
    const adjutant =
      state.adjutantRevealed && state.adjutantIdx >= 0 && state.adjutantIdx !== state.napoleonIdx
        ? state.players[state.adjutantIdx]
        : undefined;
    const collected = napoleon.pictureCards + (adjutant?.pictureCards ?? 0);
    return { collected, bid: state.highestBid, achieved: collected >= state.highestBid };
  })();

  const trumpLabel = state.trumpSuit > 0 ? t(`suitName.${SUIT_KEYS[state.trumpSuit]}`) : '';

  return (
    <>
      {/* Round/Trick info */}
      <div className="text-ds-text-primary flex flex-wrap justify-center gap-x-4 gap-y-1 text-center mb-2">
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        {trumpLabel && (
          <span>
            {t('trumpSuit')}: {trumpLabel}
          </span>
        )}
      </div>

      {napoleonFaceProgress && (
        <div className="text-center mb-2">
          <span data-testid="np-face-progress" className={napoleonFaceProgress.achieved ? badgeSuccess : badgeInfo}>
            {t('faceProgress', {
              collected: napoleonFaceProgress.collected,
              bid: napoleonFaceProgress.bid,
            })}
          </span>
        </div>
      )}

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          {/* Adjutant card info */}
          {state.adjutantCard && (
            <div className="text-ds-text-muted text-center text-sm mb-2" data-tutorial="np-adjutant-info">
              {t('adjutantCard')}: <AnimatedCard card={state.adjutantCard} width={cardWidth * 0.6} />
            </div>
          )}

          {/* Highest bid info */}
          {state.highestBid > 0 && (
            <div className="text-ds-text-muted text-center text-sm mb-2">
              <span>{t('highestBid', { bid: state.highestBid })}</span>{' '}
              <span>
                {t('highestBidder')}: {playerName(state.highestBidder, state.players[state.highestBidder].isHuman)}
              </span>
            </div>
          )}

          {/* Bid phase instruction */}
          {isHumanBidTurn && (
            <div className="text-ds-warning text-center mb-2" data-tutorial="np-bid-controls">
              {t('bidPhase', { min: game.napoleonConfig.minBid })}
            </div>
          )}

          {/* Trump declaration instruction */}
          {isHumanNapoleon && (
            <div className="text-ds-warning text-center mb-2" data-tutorial="np-trump-declaration">
              {t('trumpDeclarationPhase')}
            </div>
          )}

          {/* Adjutant card picker: tap a card face to designate the adjutant */}
          {isHumanNapoleon && (
            <div className="my-2 p-2 rounded bg-black/40" data-testid="np-adjutant-picker">
              <div className="text-ds-text-muted text-sm mb-1">{t('adjutantPickerLabel')}</div>
              <div className="text-ds-info text-xs mb-2">{t('adjutantInHandNote')}</div>
              <div className="overflow-x-auto -mx-1 px-1">
                <div className="flex flex-col gap-1 min-w-max">
                  {adjutantRows.map((row, rowIdx) => (
                    <div key={`adj-row-${SUIT_KEYS[row[0].suit] ?? 'joker'}-${rowIdx}`} className="flex gap-1">
                      {row.map((opt) => {
                        const isSelected = adjSelection?.suit === opt.suit && adjSelection?.value === opt.value;
                        const inHand = isAdjutantCardInHand(opt, humanPlayer?.cards ?? []);
                        return (
                          <button
                            type="button"
                            key={`adj-${opt.card.design}-${opt.value}`}
                            data-testid={`np-adjutant-option-${opt.suit}-${opt.value}`}
                            onClick={() => setAdjSelection(opt)}
                            aria-label={cardAlt(opt.card)}
                            aria-pressed={isSelected}
                            title={inHand ? t('adjutantInHandNote') : undefined}
                            className={`transition-transform ${focusRingCard}`}
                            style={{
                              background: 'none',
                              padding: 0,
                              borderRadius: 8,
                              opacity: inHand && !isSelected ? 0.4 : 1,
                              ...selectedCardStyle(isSelected),
                              boxSizing: 'border-box',
                            }}
                          >
                            <AnimatedCard card={opt.card} width={44} silent />
                          </button>
                        );
                      })}
                    </div>
                  ))}
                </div>
              </div>
              {adjSelection && (
                <div className="text-ds-text-primary text-sm mt-2" data-testid="np-adjutant-selected">
                  {t('adjutantSelected', { card: cardAlt(adjSelection.card) })}
                </div>
              )}
            </div>
          )}

          {/* Kitty exchange instruction */}
          {isHumanExchange && <div className="text-ds-warning text-center mb-2">{t('kittyExchangePhase')}</div>}

          {/* Kitty cards (during exchange phase) */}
          {isKittyExchange && state.kitty.length > 0 && (
            <div className="my-2 p-2 rounded bg-black/40" data-tutorial="np-kitty-cards">
              <div className="text-ds-text-muted text-sm mb-1">{t('kittyLabel')}</div>
              {isHumanExchange && <div className="text-ds-info text-xs mb-1">{t('kittyAcquireLabel')}</div>}
              <div className="flex gap-2">
                {state.kitty.map((card, idx) => (
                  <AnimatedCard key={`kitty-${card.design}-${card.value}-${idx}`} card={card} width={cardWidth} />
                ))}
              </div>
            </div>
          )}

          {/* Current trick */}
          <TrickDisplay
            currentTrick={state.currentTrick}
            winnerIdx={isTrickEnd ? state.leadPlayerIdx : undefined}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="np-trick-display"
            cardBadgeFor={(card) => {
              if (!isPlayPhase || DESIGN_ORDER[card.design] !== state.trumpSuit) return null;
              return { glyph: '★', title: t('trumpCardBadge', { suit: trumpLabel }) };
            }}
          />
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* CPU players */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
              </summary>
              <div className="mt-1">
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      {playerName(p.id, p.isHuman)}
                      {roleBadge(t, p, state.adjutantRevealed)}: {t('cards', { count: p.cardCount })} |{' '}
                      {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })}{' '}
                      | {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')} |{' '}
                      {t('pictureCards', { count: p.pictureCards })}
                    </div>
                  ))}
              </div>
            </details>
          ) : (
            state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                  <div className="flex flex-wrap gap-x-1 text-ds-text-muted text-sm">
                    <span className="whitespace-nowrap">
                      {playerName(p.id, p.isHuman)}
                      {roleBadge(t, p, state.adjutantRevealed)}: {t('cards', { count: p.cardCount })}
                    </span>
                    <span className="whitespace-nowrap">{t('cumulativeScore', { score: p.cumulativeScore })}</span>
                    <span className="whitespace-nowrap">{t('roundScore', { score: p.roundScore })}</span>
                    <span className="whitespace-nowrap">{p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}</span>
                    <span className="whitespace-nowrap">{t('pictureCards', { count: p.pictureCards })}</span>
                  </div>
                </div>
              ))
          )}

          {/* **あと何点で決着するのかが対局中どこにも出ていなかった。**
              目標点数は開始前の設定でしか見えず、思い出すには Settings を
              開き直すしかなかった (#5504)。設定値はレスポンスに載っているので
              そのまま出す。 */}
          <div className="text-ds-text-muted text-xs mt-2" data-testid="np-point-limit">
            {t('pointLimitLine', { limit: state.config.pointLimit })}
          </div>

          {/* Score table */}
          {isMobile ? (
            <details
              className="my-3 p-2 rounded bg-black/30 relative"
              data-tutorial="np-score-table"
              open={isRoundEnd || isGameEnd || undefined}
            >
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
              <div className="overflow-x-auto -mx-2 px-2">
                <table
                  className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[420px]', 'mt-1')}
                >
                  <thead>
                    <tr>
                      <th scope="col" className="text-left whitespace-nowrap text-[10px]">
                        {t('scoresPlayer')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresRole')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresBid')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresPictureCards')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTricks')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresRound')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTotal')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {state.players.map((p) => (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <td>{playerName(p.id, p.isHuman)}</td>
                        <td className="text-center">
                          {p.isNapoleon
                            ? t('role.napoleon')
                            : p.isAdjutant && (p.adjutantRevealed || state.adjutantRevealed)
                              ? t('role.adjutant')
                              : '-'}
                        </td>
                        <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                        <td className="text-center">{p.pictureCards}</td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className="text-center">{p.roundScore}</td>
                        <td className="text-center">{p.cumulativeScore}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <ScrollFadeHint />
            </details>
          ) : (
            <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="np-score-table">
              <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
              <div className="overflow-x-auto -mx-2 px-2">
                <table className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[420px]')}>
                  <thead>
                    <tr>
                      <th scope="col" className="text-left whitespace-nowrap text-[10px]">
                        {t('scoresPlayer')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresRole')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresBid')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresPictureCards')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTricks')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresRound')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTotal')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {state.players.map((p) => (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <td>{playerName(p.id, p.isHuman)}</td>
                        <td className="text-center">
                          {p.isNapoleon
                            ? t('role.napoleon')
                            : p.isAdjutant && (p.adjutantRevealed || state.adjutantRevealed)
                              ? t('role.adjutant')
                              : '-'}
                        </td>
                        <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                        <td className="text-center">{p.pictureCards}</td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className="text-center">{p.roundScore}</td>
                        <td className="text-center">{p.cumulativeScore}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
          <RoundScoreAnnouncement
            active={isRoundEnd || isGameEnd}
            entries={state.players.map((p) => ({
              name: playerName(p.id, p.isHuman),
              roundScore: p.roundScore,
              cumulativeScore: p.cumulativeScore,
            }))}
          />
        </div>
      </div>

      {/* Message */}
      {isMobile && <NapoleonMessageBox state={state} />}
    </>
  );
}

function FooterHand({ game, state, t, view }: NapoleonContext) {
  const { cardWidth, isMobile, hasRestrictedPlayForSelection, canSelectPlayCard, togglePlayableCard } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isKittyExchange = state.phase === NapoleonPhase.KITTY_EXCHANGE;
  const isHumanExchange = isKittyExchange && state.players[state.napoleonIdx]?.isHuman === true;
  const isPlayPhase = state.phase === NapoleonPhase.PLAY;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const validPlayIndices =
    state.phase === NapoleonPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman ? state.validPlayIndices : [];

  return (
    <>
      {isHumanExchange && (
        <div className="text-ds-info text-xs text-center mb-1" data-testid="np-hand-discard-label">
          {t('handDiscardLabel')}
        </div>
      )}
      {/* Human cards */}
      {humanPlayer &&
        (isMobile ? (
          <MobileHandGrid
            cards={humanPlayer.cards}
            selectedIndices={game.selectedCardIndices}
            onToggle={togglePlayableCard}
            cardWidth={cardWidth}
            dataTutorial="np-player-hand"
            // Disable illegal cards only when restricted; highlight legal cards on the human turn.
            validIndices={hasRestrictedPlayForSelection ? validPlayIndices : undefined}
            legalIndices={isHumanTurn ? validPlayIndices : undefined}
          />
        ) : (
          <div className="flex flex-wrap gap-1 mb-2" data-tutorial="np-player-hand">
            {humanPlayer.cards.map((card, idx) => (
              <button
                type="button"
                key={`${card.design}-${card.value}-${idx}`}
                onClick={() => togglePlayableCard(idx)}
                aria-label={cardAlt(card)}
                aria-pressed={game.selectedCardIndices.includes(idx)}
                aria-disabled={(hasRestrictedPlayForSelection && !canSelectPlayCard(idx)) || undefined}
                className={`transition-transform ${focusRingCard} ${isHumanTurn && canSelectPlayCard(idx) ? 'rounded-lg ring-2 ring-ds-success' : ''} ${hasRestrictedPlayForSelection && !canSelectPlayCard(idx) ? 'opacity-50 cursor-not-allowed' : ''}`}
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
    </>
  );
}

function FooterStatus({ game, t, frontendHint }: NapoleonContext) {
  const hint = game.hint;
  return (
    <>
      {/*
        領域と中身を同じコミットで DOM に入れると、変化として扱われず読み上げ
        られないことがある (#5955)。だから role/aria-live はヒントの有無に
        かかわらず常設のラッパー側に置く。
      */}
      <div data-testid="np-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {hint.bid != null
              ? `${t('hintBid')}: ${hint.bid} (${t(`hintReason.${hint.reason}`)})`
              : hint.trumpSuit != null
                ? `${t('hintTrump')}: ${t(`suitName.${SUIT_KEYS[hint.trumpSuit]}`)} (${t(`hintReason.${hint.reason}`)})`
                : hint.discardIndex != null
                  ? `${t('hintDiscard')}: [${hint.discardIndex}] (${t(`hintReason.${hint.reason}`)})`
                  : `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function primaryActions({ game, state, t, tc, busy, view }: NapoleonContext) {
  const { bidValue, setBidValue, trumpSuitValue, setTrumpSuitValue, adjSelection, confirmAction } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === NapoleonPhase.BID;
  const isTrumpDeclaration = state.phase === NapoleonPhase.TRUMP_DECLARATION;
  const isKittyExchange = state.phase === NapoleonPhase.KITTY_EXCHANGE;
  const isPlayPhase = state.phase === NapoleonPhase.PLAY;
  const isTrickEnd = state.phase === NapoleonPhase.TRICK_END;
  const isRoundEnd = state.phase === NapoleonPhase.ROUND_END;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const minLegalBid = Math.max(game.napoleonConfig.minBid, state.highestBid + 1);
  const effectiveBidValue = Math.max(bidValue, minLegalBid);
  const isBidValueLegal = effectiveBidValue <= MAX_BID;
  const isHumanNapoleon = isTrumpDeclaration && state.players[state.napoleonIdx]?.isHuman === true;
  const isHumanExchange = isKittyExchange && state.players[state.napoleonIdx]?.isHuman === true;

  if (!isHumanBidTurn && !isHumanTurn && !isHumanNapoleon && !isHumanExchange && !isTrickEnd && !isRoundEnd) {
    return null;
  }

  return (
    <div className="flex gap-2 items-center flex-wrap">
      {(isHumanBidTurn || isHumanTurn || isHumanNapoleon || isHumanExchange) && (
        <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
          {tc('button.hint')}
        </button>
      )}

      {/* Bid controls */}
      {isHumanBidTurn && (
        <>
          <input
            type="number"
            min={minLegalBid}
            max={MAX_BID}
            value={effectiveBidValue}
            onChange={(e) => setBidValue(Number(e.target.value))}
            className="w-16 px-2 py-1 rounded bg-white/20 text-ds-text-primary text-center"
            aria-label={t('bidInputLabel')}
          />
          <button
            type="button"
            className={`${btnPrimary}${!isBidValueLegal ? ' opacity-50 cursor-not-allowed active:scale-100 hover:shadow-none' : ''}`}
            onClick={() => {
              if (isBidValueLegal) game.handleBid(effectiveBidValue);
            }}
            disabled={busy}
            aria-disabled={!isBidValueLegal || undefined}
          >
            {t('bidButton')}
          </button>
          <button type="button" className={btnPrimary} onClick={game.handlePass} disabled={busy}>
            {t('passButton')}
          </button>
        </>
      )}

      {/* Trump declaration controls (adjutant card is chosen via the visual picker above) */}
      {isHumanNapoleon && (
        <>
          <select
            value={trumpSuitValue}
            onChange={(e) => setTrumpSuitValue(Number(e.target.value))}
            className="px-2 py-1 rounded bg-white/20 text-ds-text-primary"
            aria-label={t('trumpSuitLabel')}
          >
            {[1, 2, 3, 4].map((s) => (
              <option key={s} value={s}>
                {t(`suitName.${SUIT_KEYS[s]}`)}
              </option>
            ))}
          </select>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => {
              if (adjSelection) game.handleTrumpDeclaration(trumpSuitValue, adjSelection.suit, adjSelection.value);
            }}
            disabled={busy || !adjSelection}
          >
            {t('declareButton')}
          </button>
        </>
      )}

      {/* Kitty exchange controls */}
      {isHumanExchange && (
        <button
          type="button"
          className={btnPrimary}
          onClick={() => {
            if (game.selectedCardIndices.length === 1) game.handleExchange(game.selectedCardIndices[0]);
          }}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {game.selectedCardIndices.length === 1 && humanPlayer?.cards[game.selectedCardIndices[0]]
            ? t('exchangeButtonNamed', { card: cardAlt(humanPlayer.cards[game.selectedCardIndices[0]]) })
            : t('exchangeButton')}
        </button>
      )}

      {/* Play controls */}
      {isHumanTurn && (
        <button
          type="button"
          className={btnPrimary}
          onClick={confirmAction}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}

      {/* Trick end */}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}

      {/* Round end */}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const napoleonDefinition: GameDefinition<
  NapoleonResponse,
  NapoleonGame,
  Parameters<typeof napoleonApi.exec>,
  NapoleonView
> = {
  key: 'napoleon',
  useGame,
  useView: useNapoleonView,
  tutorial: NP_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 4, trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseNapoleonCommand,
    formatResponse: formatNapoleonState,
    helpText: NAPOLEON_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === NapoleonPhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isBidPhase = state.phase === NapoleonPhase.BID;
    const isTrumpDeclaration = state.phase === NapoleonPhase.TRUMP_DECLARATION;
    const isKittyExchange = state.phase === NapoleonPhase.KITTY_EXCHANGE;
    const isPlayPhase = state.phase === NapoleonPhase.PLAY;
    const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
    const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
    const isHumanNapoleon = isTrumpDeclaration && state.players[state.napoleonIdx]?.isHuman === true;
    const isHumanExchange = isKittyExchange && state.players[state.napoleonIdx]?.isHuman === true;
    return {
      isHumanTurn: isHumanBidTurn || isHumanTurn || isHumanNapoleon || isHumanExchange,
    };
  },
  reset: ({ game }) => {
    void game.apiExec('reset', undefined, undefined, undefined, undefined, undefined, undefined, {
      cpuDifficulty: game.napoleonConfig.cpuDifficulty,
      pointLimit: game.napoleonConfig.pointLimit,
      minBid: game.napoleonConfig.minBid,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.napoleonConfig.cpuDifficulty,
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
          value: game.napoleonConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('pointLimit', v),
        },
        {
          type: 'select',
          id: 'minBid',
          label: t('settings.minBid'),
          value: game.napoleonConfig.minBid,
          options: MIN_BID_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('minBid', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  message: (ctx) => !ctx.view.isMobile && <NapoleonMessageBox state={ctx.state} />,
  cardNav: ({ game, view }) => ({
    cardCount: view.humanCardCountForKbd,
    onToggle: view.togglePlayableCard,
    onConfirm: view.confirmAction,
    onClear: game.clearSelection,
    enabled: !!view.isHumanTurnForKbd,
  }),
  shortcutsTestId: 'napoleon-kbd-shortcuts',
  resetTutorialId: 'np-reset-button',
  actionsClassName: 'flex items-center',
  controlsInFooter: true,
  AboveContent: PartnerReveal,
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Napoleon game page with bidding, trump declaration, kitty exchange, trick play, and scoring. */
export const NapoleonPage = createGamePage(napoleonDefinition);
