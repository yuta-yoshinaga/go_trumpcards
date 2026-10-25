import type { TFunction } from 'i18next';
import { useCallback, useEffect, useMemo, useState } from 'react';
import type { binokelApi } from '../api/gameApi';
import { CardFan } from '../components/CardFan';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useBinokelGame } from '../hooks/useBinokelGame';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import type { BinokelMeldData, BinokelResponse } from '../types/card';
import { BinokelPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BINOKEL_HELP, parseBinokelCommand } from '../utils/cli/commands/binokelCommands';
import { formatBinokelState } from '../utils/cli/formatters/binokelFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Phase name keys for Binokel. */
const BINOKEL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BinokelPhase.BID]: 'bid',
  [BinokelPhase.DABB]: 'dabb',
  [BinokelPhase.TRUMP]: 'trump',
  [BinokelPhase.MELD]: 'meld',
  [BinokelPhase.PLAY]: 'play',
  [BinokelPhase.TRICK_END]: 'trickEnd',
  [BinokelPhase.ROUND_END]: 'roundEnd',
  [BinokelPhase.GAME_END]: 'gameEnd',
};

/** Suit labels for display. */
const SUIT_LABELS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Binokel tutorial step definitions. */
const BN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bn-game-info"]',
    messageKey: 'tutorial.gameInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bn-player-info"]',
    messageKey: 'tutorial.playerInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bn-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bn-meld-area"]',
    messageKey: 'tutorial.meldArea',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bn-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bn-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bn-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type BinokelGame = ReturnType<typeof useBinokelGame>;
type BinokelView = ReturnType<typeof useBinokelView>;
type BinokelContext = GamePageContext<BinokelResponse, BinokelGame, BinokelView>;

function useBinokelView({ game, t }: { game: BinokelGame; t: TFunction }) {
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('binokel', BINOKEL_PHASE_KEYS);

  // Selected card indices for Dabb discard phase (3 cards required).
  const [selectedCardIndices, setSelectedCardIndices] = useState<number[]>([]);

  // Bidding minimum is 150 or 10 higher than current highest bid.
  const state = game.state;
  const highestBid = state?.highestBid;
  const minLegalBid = useMemo(() => {
    if (highestBid === undefined) return 150;
    return highestBid >= 150 ? highestBid + 10 : 150;
  }, [highestBid]);

  const [bidAmount, setBidAmount] = useState(150);

  // Auto-update bid amount to at least the minimum legal bid when highest bid changes
  useEffect(() => {
    if (state?.highestBid !== undefined) {
      const min = state.highestBid >= 150 ? state.highestBid + 10 : 150;
      setBidAmount((prev) => (prev < min ? min : prev));
    }
  }, [state?.highestBid]);

  // Index into the human player's meld list. When set, every hand card whose
  // (design, value) appears in that meld is ringed so the player can read the
  // meld back to the cards. Reset on phase exit.
  const [highlightedMeldIdx, setHighlightedMeldIdx] = useState<number | null>(null);

  // Drop the meld highlight whenever the meld phase exits
  useEffect(() => {
    if (state?.phase !== BinokelPhase.MELD) {
      setHighlightedMeldIdx(null);
    }
  }, [state?.phase]);

  // Clear discard selection when leaving DABB phase
  useEffect(() => {
    if (state?.phase !== BinokelPhase.DABB) {
      setSelectedCardIndices([]);
    }
  }, [state?.phase]);

  const toggleCardSelection = useCallback((idx: number) => {
    setSelectedCardIndices((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  }, []);

  // Memoised highlight keys for meld viewing
  const highlightedCardKeys = useMemo(() => {
    if (!state || highlightedMeldIdx === null) return new Set<string>();
    const humanIdxLocal = state.players.findIndex((p) => p.isHuman);
    const meld = humanIdxLocal >= 0 ? state.playerMelds[humanIdxLocal]?.[highlightedMeldIdx] : null;
    if (!meld) return new Set<string>();
    return new Set<string>(meld.cards.map((c) => `${c.design}-${c.value}`));
  }, [state, highlightedMeldIdx]);

  // Precompute the localized meld-list tooltip per card key
  const cardKeyToMeldTitle = useMemo(() => {
    const out = new Map<string, string>();
    const players = state?.players;
    const playerMelds = state?.playerMelds;
    if (!players || !playerMelds) return out;
    const humanIdxLocal = players.findIndex((p) => p.isHuman);
    const myMelds = humanIdxLocal >= 0 ? playerMelds[humanIdxLocal] : null;
    if (!myMelds) return out;
    const cardKeyToTypes = new Map<string, number[]>();
    for (const m of myMelds) {
      for (const c of m.cards) {
        const key = `${c.design}-${c.value}`;
        const existing = cardKeyToTypes.get(key);
        if (existing) {
          if (!existing.includes(m.type)) existing.push(m.type);
        } else {
          cardKeyToTypes.set(key, [m.type]);
        }
      }
    }
    for (const [key, types] of cardKeyToTypes) {
      out.set(key, types.map((mt) => t(`meldTypes.${mt}`)).join(', '));
    }
    return out;
  }, [state?.players, state?.playerMelds, t]);

  // Generate discrete bid options starting strictly from minLegalBid (never below)
  const bidOptions = useMemo(() => {
    const options: number[] = [];
    const max = Math.max(500, minLegalBid + 100);
    for (let b = minLegalBid; b <= max; b += 10) {
      options.push(b);
    }
    return options;
  }, [minLegalBid]);

  return {
    isLargeDesktop,
    cardWidth,
    isMobile,
    phaseNames,
    selectedCardIndices,
    setSelectedCardIndices,
    toggleCardSelection,
    bidAmount,
    setBidAmount,
    minLegalBid,
    bidOptions,
    highlightedMeldIdx,
    setHighlightedMeldIdx,
    highlightedCardKeys,
    cardKeyToMeldTitle,
  };
}

function BinokelMessageBox({ state }: Pick<BinokelContext, 'state'>) {
  return <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />;
}

function Board({ state, t, busy, view, frontendHint }: BinokelContext) {
  const loading = busy;
  const phase = state.phase;
  const humanPlayer = state.players?.find((p) => p.isHuman);
  const humanIdx = humanPlayer ? state.players.indexOf(humanPlayer) : -1;
  const isBidTurn = phase === BinokelPhase.BID && state.players?.[state.bidPlayerIdx]?.isHuman;
  const isHumanDeclarer = state.highestBidder >= 0 && state.players?.[state.highestBidder]?.isHuman === true;
  const isDabbPhase = phase === BinokelPhase.DABB;
  // 早見表はビッド〜メルド確認の間だけ。プレイ中は盤面を見る場面なので畳む。
  const meldTable =
    phase === BinokelPhase.BID ||
    phase === BinokelPhase.DABB ||
    phase === BinokelPhase.TRUMP ||
    phase === BinokelPhase.MELD
      ? state.meldTable
      : undefined;

  return (
    <>
      {isBidTurn && (
        <div className="mb-2 flex flex-col items-start gap-2">
          <fieldset
            className="grid grid-cols-4 sm:grid-cols-6 max-h-36 overflow-y-auto gap-1 border-0 p-1 bg-black/20 rounded"
            aria-label={t('bidSelectLabel')}
          >
            {view.bidOptions.map((n) => (
              <button
                key={n}
                type="button"
                onClick={() => view.setBidAmount(n)}
                disabled={loading}
                aria-pressed={view.bidAmount === n}
                data-testid={`bid-option-${n}`}
                className={`h-9 px-2 rounded-lg font-medium text-sm transition-all ${
                  view.bidAmount === n
                    ? 'bg-ds-accent text-ds-text-on-accent ring-2 ring-ds-accent'
                    : 'bg-white/20 text-ds-text-primary hover:bg-white/30'
                }`}
              >
                {n}
              </button>
            ))}
          </fieldset>
          <span className="sr-only" role="status" aria-live="polite" data-testid="bn-bid-selected">
            {t('bidSelected', { n: view.bidAmount })}
          </span>
        </div>
      )}
      {/* Game Info */}
      <div className="text-ds-text-primary text-center mb-2 text-sm" data-tutorial="bn-game-info">
        <span className="mr-4">
          {t('round')}: {state.roundNumber} / {t('trick')}: {state.trickNumber}
        </span>
        <span className="mr-4">
          {t('trumpSuit')}: {state.trumpSuit > 0 ? SUIT_LABELS[state.trumpSuit] : '-'}
        </span>
        <span className="mr-4">
          {t('declarer')}:{' '}
          {state.highestBidder >= 0
            ? playerName(state.highestBidder, state.players[state.highestBidder]?.isHuman)
            : '-'}
        </span>
        {state.highestBid > 0 && (
          <span>
            {t('highestBid')}: {state.highestBid}
          </span>
        )}
      </div>

      <div
        className={view.isLargeDesktop ? 'flex flex-wrap items-start gap-2' : ''}
        data-testid="bn-player-and-dabb-row"
      >
        {/* Players Info: 3 individual seats */}
        <div className={`mb-3 ${view.isLargeDesktop ? 'flex-1 min-w-0' : ''}`} data-tutorial="bn-player-info">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
            {state.players
              ?.filter((p) => p.isHuman)
              .map((p) => {
                const onTurn =
                  phase === BinokelPhase.BID
                    ? p.id === state.bidPlayerIdx
                    : (phase === BinokelPhase.DABB ||
                        phase === BinokelPhase.TRUMP ||
                        phase === BinokelPhase.MELD ||
                        phase === BinokelPhase.PLAY) &&
                      p.id === state.currentPlayerIdx;
                const isDeclarer = state.highestBidder === p.id;
                const bidDisplay = p.hasPassed
                  ? t('bidStatus.passed')
                  : p.bid > 0
                    ? t('bidStatus.bid', { amount: p.bid })
                    : t('bidStatus.waiting');
                const scoreVal = state.scores ? state.scores[p.id] : p.score;
                return (
                  <div
                    key={p.id}
                    data-on-turn={onTurn || undefined}
                    className={`rounded p-2 text-sm ${
                      p.isHuman ? 'bg-ds-accent/20 text-ds-accent' : 'bg-black/30 text-ds-text-muted'
                    } ${onTurn ? 'ring-2 ring-ds-warning' : ''}`}
                  >
                    <div className="font-bold flex items-center justify-between">
                      <span className="min-w-0 truncate" title={playerName(p.id, p.isHuman)}>
                        {playerName(p.id, p.isHuman)}
                      </span>
                      {isDeclarer && (
                        <span className="shrink-0 whitespace-nowrap text-xs px-1.5 py-0.5 rounded bg-ds-accent text-black font-semibold">
                          {t('declarer')}
                        </span>
                      )}
                    </div>
                    <div className="flex flex-wrap gap-x-2 gap-y-1 text-xs">
                      <span className="whitespace-nowrap">
                        {t('scores')}: {scoreVal}
                      </span>
                      <span className="whitespace-nowrap">
                        {t('bid')}: {bidDisplay}
                      </span>
                      <span className="whitespace-nowrap">
                        {t('meldScore')}: {p.meldScore}
                      </span>
                      <span className="whitespace-nowrap">
                        {t('trickCount')}: {p.trickCount} ({t('points', { points: p.trickPoints })})
                      </span>
                    </div>
                  </div>
                );
              })}
            <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length} className="mb-0 sm:col-span-2">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                {state.players
                  ?.filter((p) => !p.isHuman)
                  .map((p) => {
                    const onTurn =
                      phase === BinokelPhase.BID
                        ? p.id === state.bidPlayerIdx
                        : (phase === BinokelPhase.DABB ||
                            phase === BinokelPhase.TRUMP ||
                            phase === BinokelPhase.MELD ||
                            phase === BinokelPhase.PLAY) &&
                          p.id === state.currentPlayerIdx;
                    const isDeclarer = state.highestBidder === p.id;
                    const bidDisplay = p.hasPassed
                      ? t('bidStatus.passed')
                      : p.bid > 0
                        ? t('bidStatus.bid', { amount: p.bid })
                        : t('bidStatus.waiting');
                    const scoreVal = state.scores ? state.scores[p.id] : p.score;
                    return (
                      <div
                        key={p.id}
                        data-on-turn={onTurn || undefined}
                        className={`rounded p-2 text-sm bg-black/30 text-ds-text-muted ${onTurn ? 'ring-2 ring-ds-warning' : ''}`}
                      >
                        <div className="font-bold flex items-center justify-between">
                          <span className="min-w-0 truncate" title={playerName(p.id, p.isHuman)}>
                            {playerName(p.id, p.isHuman)}
                          </span>
                          {isDeclarer && (
                            <span className="shrink-0 whitespace-nowrap text-xs px-1.5 py-0.5 rounded bg-ds-accent text-black font-semibold">
                              {t('declarer')}
                            </span>
                          )}
                        </div>
                        <div className="flex flex-wrap gap-x-2 gap-y-1 text-xs">
                          <span className="whitespace-nowrap">
                            {t('scores')}: {scoreVal}
                          </span>
                          <span className="whitespace-nowrap">
                            {t('bid')}: {bidDisplay}
                          </span>
                          <span className="whitespace-nowrap">
                            {t('meldScore')}: {p.meldScore}
                          </span>
                          <span className="whitespace-nowrap">
                            {t('trickCount')}: {p.trickCount} ({t('points', { points: p.trickPoints })})
                          </span>
                        </div>
                      </div>
                    );
                  })}
              </div>
            </CpuAccordion>
          </div>
        </div>

        {/* Dabb Cards Display */}
        {state.dabb && state.dabb.length > 0 && (
          <div
            className={`${view.isLargeDesktop ? 'mb-1 py-1' : 'mb-3 p-2'} ${view.isLargeDesktop ? 'flex-none' : ''} rounded bg-black/40 text-center`}
            data-testid="bn-dabb-display"
          >
            <div className="text-ds-text-muted text-sm mb-1">{t('dabbCardsLabel')}:</div>
            <div className="flex gap-2 justify-center">
              {state.dabb.map((c, i) => (
                <AnimatedCard key={`dabb-${i}`} card={c} width={view.cardWidth * 0.8} />
              ))}
            </div>
          </div>
        )}
      </div>

      {/* CPU Dabb waiting notice */}
      {isDabbPhase && !isHumanDeclarer && state.highestBidder >= 0 && (
        <div className="mb-3 p-3 rounded bg-black/30 text-center text-sm text-ds-text-muted" data-testid="dabb-waiting">
          {t('dabbWaiting', {
            name: playerName(state.highestBidder, state.players[state.highestBidder]?.isHuman),
          })}
        </div>
      )}

      {/* Current Trick */}
      {state.currentTrick?.length > 0 && (
        <div className="mb-3 p-2 rounded bg-black/40" data-tutorial="bn-trick-display">
          <div className="text-ds-text-muted text-sm mb-1">{t('table')}:</div>
          <div className="flex gap-2 justify-center">
            {state.currentTrick.map((tc, i) => {
              const isHuman = state.players[tc.playerIdx]?.isHuman === true;
              return (
                <div key={i} className="text-center">
                  <AnimatedCard card={tc.card} width={view.cardWidth * 0.8} />
                  <div className={`text-xs mt-1 ${isHuman ? 'text-ds-accent font-semibold' : 'text-ds-text-muted'}`}>
                    {playerName(tc.playerIdx, isHuman)}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Meld reference: server-provided domain values */}
      {meldTable && (
        <details data-testid="bn-meld-table" className="mb-3 rounded bg-black/30" open={!view.isMobile || undefined}>
          <summary className="cursor-pointer select-none px-3 py-2 text-ds-text-primary font-bold text-sm">
            {t('meldTableTitle')}
          </summary>
          <ul className="px-3 pb-2 text-ds-text-muted text-sm grid grid-cols-2 sm:grid-cols-3 gap-x-3">
            {meldTable.map((e) => (
              <li key={e.type}>
                {t(`meldTypes.${e.type}`)} {e.points}
              </li>
            ))}
          </ul>
        </details>
      )}

      {/* Melds */}
      {(phase === BinokelPhase.MELD || phase === BinokelPhase.PLAY || phase === BinokelPhase.ROUND_END) &&
        state.playerMelds && (
          // biome-ignore lint/a11y/noStaticElementInteractions: panel acts as a hover-out reset for the per-badge highlight; not interactive on its own.
          <div
            className="mb-3 p-2 rounded bg-ds-accent/15"
            data-tutorial="bn-meld-area"
            onMouseLeave={() => view.setHighlightedMeldIdx(null)}
          >
            <div className="text-ds-text-primary font-bold mb-1">{t('meldScore')}:</div>
            {state.playerMelds.map((melds: BinokelMeldData[], pIdx: number) =>
              melds.length > 0 ? (
                <div key={pIdx} className="text-ds-text-muted text-sm mb-1">
                  <span className="font-semibold">{playerName(pIdx, state.players[pIdx]?.isHuman)}: </span>
                  {melds.map((m: BinokelMeldData, mIdx: number) => {
                    const isHumanMeld = pIdx === humanIdx && phase === BinokelPhase.MELD;
                    const isActive = isHumanMeld && view.highlightedMeldIdx === mIdx;
                    if (!isHumanMeld) {
                      return (
                        <span key={mIdx} className="mr-2">
                          {t(`meldTypes.${m.type}`)} ({m.points})
                        </span>
                      );
                    }
                    return (
                      <button
                        key={mIdx}
                        type="button"
                        onClick={() => view.setHighlightedMeldIdx((prev) => (prev === mIdx ? null : mIdx))}
                        onMouseEnter={() => view.setHighlightedMeldIdx(mIdx)}
                        onFocus={() => view.setHighlightedMeldIdx(mIdx)}
                        data-testid={`bn-meld-badge-${mIdx}`}
                        data-active={isActive ? 'true' : undefined}
                        className={`mr-2 inline-block px-2 py-0.5 rounded text-xs ${
                          isActive ? 'bg-ds-accent text-black font-bold' : 'bg-black/20 hover:bg-ds-accent/30'
                        }`}
                      >
                        {t(`meldTypes.${m.type}`)} ({m.points})
                      </button>
                    );
                  })}
                </div>
              ) : null,
            )}
            {phase === BinokelPhase.MELD && view.highlightedCardKeys.size > 0 && (
              <div className="text-ds-text-muted text-xs mt-1">{t('meldHighlightHint')}</div>
            )}
          </div>
        )}

      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      {view.isMobile && <BinokelMessageBox state={state} />}
    </>
  );
}

function FooterStatus({ game, state, t, busy, view }: BinokelContext) {
  const { hint } = game;
  const isBidTurn = state.phase === BinokelPhase.BID && state.players?.[state.bidPlayerIdx]?.isHuman;
  const humanPlayer = state.players?.find((p) => p.isHuman);

  return (
    <div data-testid="binokel-hint-live" role="status" aria-live="polite">
      {hint?.reason && (
        <div className="text-ds-warning text-sm mb-2 flex items-center gap-2 flex-wrap" data-testid="bn-server-hint">
          <span>
            {hint.pass
              ? t('serverHintPass')
              : hint.bidAmount !== undefined
                ? t('serverHintBid', { n: hint.bidAmount })
                : hint.suit !== undefined
                  ? t('serverHintTrump', { suit: SUIT_LABELS[hint.suit] ?? '-' })
                  : hint.cardIndex !== undefined
                    ? t('serverHintPlay', {
                        card: humanPlayer?.cards[hint.cardIndex] ? cardAlt(humanPlayer.cards[hint.cardIndex]) : '-',
                        idx: hint.cardIndex,
                      })
                    : ''}{' '}
            ({t(`hintReason.${hint.reason}`)})
          </span>
          {isBidTurn && hint.bidAmount !== undefined && (
            <button
              type="button"
              className={btnOutline}
              data-testid="bn-hint-apply-bid"
              onClick={() => hint.bidAmount !== undefined && view.setBidAmount(hint.bidAmount)}
              disabled={busy}
            >
              {t('applyBid')}
            </button>
          )}
        </div>
      )}
    </div>
  );
}

function primaryActions({ game, state, t, tc, busy, view }: BinokelContext) {
  const phase = state.phase;
  const isBidTurn = phase === BinokelPhase.BID && state.players?.[state.bidPlayerIdx]?.isHuman;
  const isHumanDeclarer = state.highestBidder >= 0 && state.players?.[state.highestBidder]?.isHuman === true;
  const isDabbTurn = phase === BinokelPhase.DABB && isHumanDeclarer;
  const isTrumpTurn = phase === BinokelPhase.TRUMP && state.players?.[state.currentPlayerIdx]?.isHuman;
  const isPlayTurn = phase === BinokelPhase.PLAY && state.players?.[state.currentPlayerIdx]?.isHuman;

  return (
    <div className="flex gap-2 items-center flex-wrap" data-tutorial="bn-action-buttons">
      {/* Server hint button */}
      {(isBidTurn || isDabbTurn || isTrumpTurn || isPlayTurn) && (
        <button
          type="button"
          className={btnSuccess}
          data-testid="bn-hint-button"
          onClick={game.handleHint}
          disabled={busy || game.hintLoading}
        >
          {tc('button.hint')}
        </button>
      )}

      {/* Bid: discrete buttons from minLegalBid upwards */}
      {isBidTurn && (
        <div className="flex flex-col items-center gap-2">
          <div className="flex gap-2">
            <button
              type="button"
              className={btnPrimary}
              onClick={() => game.handleBid(view.bidAmount)}
              disabled={busy || view.bidAmount < view.minLegalBid}
            >
              {t('bid')}
            </button>
            <button
              type="button"
              className={btnOutline}
              onClick={game.handlePass}
              disabled={busy}
              data-testid="bid-pass"
            >
              {t('pass')}
            </button>
          </div>
        </div>
      )}

      {/* Dabb discard: discard 3 cards */}
      {isDabbTurn && (
        <button
          type="button"
          className={btnPrimary}
          data-testid="discard-dabb-button"
          onClick={() => game.handleDiscard(view.selectedCardIndices)}
          disabled={busy || view.selectedCardIndices.length !== 3}
        >
          {t('discardDabb', { count: view.selectedCardIndices.length })}
        </button>
      )}

      {/* Trump */}
      {isTrumpTurn &&
        [1, 2, 3, 4].map((suit) => (
          <button
            key={suit}
            type="button"
            className={btnPrimary}
            aria-label={t('trumpButtonLabel', { suit: t(`suits.${suit}`) })}
            onClick={() => game.handleCallTrump(suit)}
            disabled={busy}
          >
            {SUIT_LABELS[suit]}
          </button>
        ))}

      {/* Meld confirm */}
      {phase === BinokelPhase.MELD && (
        <button type="button" className={btnSuccess} onClick={game.handleConfirmMelds} disabled={busy}>
          {t('confirmMelds')}
        </button>
      )}

      {/* Trick End */}
      {phase === BinokelPhase.TRICK_END && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}

      {/* Round End */}
      {phase === BinokelPhase.ROUND_END && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

function FooterHand({ game, state, t, busy, view }: BinokelContext) {
  const phase = state.phase;
  const humanPlayer = state.players?.find((p) => p.isHuman);
  const isHumanDeclarer = state.highestBidder >= 0 && state.players?.[state.highestBidder]?.isHuman === true;
  const isDabbTurn = phase === BinokelPhase.DABB && isHumanDeclarer;
  const isPlayTurn = phase === BinokelPhase.PLAY && state.players?.[state.currentPlayerIdx]?.isHuman;

  return (
    <>
      {/* Hand */}
      {humanPlayer && humanPlayer.cards.length > 0 && (
        <div className="mb-2" data-tutorial="bn-player-hand">
          <CardFan itemWidth={view.cardWidth} data-testid="bn-player-hand">
            {humanPlayer.cards.map((card, idx) => {
              const isValid = state.validPlayIndices?.includes(idx);
              const cardKey = `${card.design}-${card.value}`;
              const inHighlightedMeld = view.highlightedCardKeys.has(cardKey);
              const meldTitle = view.cardKeyToMeldTitle.get(cardKey);
              const isInMeld = meldTitle !== undefined;
              const dimmedByMeldFocus = view.highlightedMeldIdx !== null && !inHighlightedMeld;
              const isSelectedForDiscard = view.selectedCardIndices.includes(idx);

              const handleClick = () => {
                if (isDabbTurn) {
                  view.toggleCardSelection(idx);
                } else if (isPlayTurn && isValid) {
                  game.handlePlay(idx);
                }
              };

              const isClickable = isDabbTurn || (isPlayTurn && isValid);
              const isDisabled = busy || !isClickable;

              return (
                <button
                  type="button"
                  key={`${card.design}-${card.value}-${idx}`}
                  onClick={handleClick}
                  disabled={isDisabled}
                  aria-label={cardAlt(card)}
                  aria-pressed={isDabbTurn ? isSelectedForDiscard : undefined}
                  data-meld-highlighted={inHighlightedMeld ? 'true' : undefined}
                  data-in-meld={isInMeld ? 'true' : undefined}
                  data-selected-discard={isSelectedForDiscard ? 'true' : undefined}
                  title={meldTitle}
                  className={`relative transition-all${
                    isSelectedForDiscard
                      ? ' -translate-y-3 ring-4 ring-ds-warning'
                      : inHighlightedMeld
                        ? ' -translate-y-2 ring-4 ring-ds-accent'
                        : ''
                  }${dimmedByMeldFocus ? ' opacity-40' : ''}`}
                  style={{
                    background: 'none',
                    padding: 0,
                    borderRadius: 8,
                    opacity: dimmedByMeldFocus ? undefined : isPlayTurn && !isValid ? 0.5 : 1,
                    boxSizing: 'border-box',
                  }}
                >
                  <AnimatedCard card={card} width={view.cardWidth} />
                  {isInMeld && (
                    <span
                      aria-hidden="true"
                      className="absolute top-0.5 right-0.5 px-1 rounded-full bg-ds-accent text-ds-text-on-accent text-[10px] font-bold shadow-sm pointer-events-none"
                      data-testid={`bn-meld-card-badge-${idx}`}
                    >
                      M
                    </span>
                  )}
                </button>
              );
            })}
          </CardFan>
        </div>
      )}

      <span className="sr-only" role="status" aria-live="polite" data-testid="bn-dabb-selection-live">
        {isDabbTurn ? t('dabbSelectionCount', { count: view.selectedCardIndices.length }) : ''}
      </span>
    </>
  );
}

const binokelDefinition: GameDefinition<
  BinokelResponse,
  BinokelGame,
  Parameters<typeof binokelApi.exec>,
  BinokelView
> = {
  key: 'binokel',
  useGame: useBinokelGame,
  useView: useBinokelView,
  tutorial: BN_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 15 },
  cli: () => ({
    parseCommand: parseBinokelCommand,
    formatResponse: formatBinokelState,
    helpText: BINOKEL_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === BinokelPhase.GAME_END || !!state.gameEndFlag,
    won: state.phase === BinokelPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const phase = state.phase;
    const isBidTurn = phase === BinokelPhase.BID && state.players?.[state.bidPlayerIdx]?.isHuman;
    const isHumanDeclarer = state.highestBidder >= 0 && state.players?.[state.highestBidder]?.isHuman === true;
    const isDabbTurn = phase === BinokelPhase.DABB && isHumanDeclarer;
    const isTrumpTurn = phase === BinokelPhase.TRUMP && state.players?.[state.currentPlayerIdx]?.isHuman;
    const isPlayTurn = phase === BinokelPhase.PLAY && state.players?.[state.currentPlayerIdx]?.isHuman;
    return {
      isHumanTurn: isBidTurn || isDabbTurn || isTrumpTurn || isPlayTurn,
    };
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select' as const,
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.binokelConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({ value: o.value, label: t(o.labelKey) })),
          onSelect: (v: string) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select' as const,
          id: 'pointLimit',
          label: t('settings.pointLimit'),
          value: game.binokelConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v: string) => game.handleConfigChange('pointLimit', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  primaryActions,
  message: (ctx) => (!ctx.view.isMobile ? <BinokelMessageBox state={ctx.state} /> : null),
  resetTutorialId: 'bn-reset-button',
  actionsClassName: 'flex justify-end',
  Board,
  FooterHand,
  FooterStatus,
};

/** Binokel game page. */
export const BinokelPage = createGamePage(binokelDefinition);
