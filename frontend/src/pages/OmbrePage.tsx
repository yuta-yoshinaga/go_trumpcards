import { useEffect, useState } from 'react';
import type { ombreApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useOmbreGame } from '../hooks/useOmbreGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { OmbreResponse } from '../types/card';
import { OmbrePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { OMBRE_HELP, parseOmbreCommand } from '../utils/cli/commands/ombreCommands';
import { formatOmbreState } from '../utils/cli/formatters/ombreFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { MATADOR_NAME_KEY, matadorRank } from '../utils/ombreMatadors';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Ombre tutorial step definitions. */
const OMBRE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ombre-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ombre-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ombre-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ombre-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ombre-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const OMBRE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [OmbrePhase.BID]: 'bid',
  [OmbrePhase.PLAY]: 'play',
  [OmbrePhase.TRICK_END]: 'trickEnd',
  [OmbrePhase.ROUND_END]: 'roundEnd',
  [OmbrePhase.GAME_END]: 'gameEnd',
};

/** Bid labels indexed by bid value (0=pass/none, 1=entrar, 2=solo). */
const BID_KEYS: Readonly<Record<number, string>> = { 0: 'bidNone', 1: 'bidEntrar', 2: 'bidSolo' };

/** Trump-suit i18n keys indexed by suit code (1=♠ 2=♣ 3=♥ 4=♦); index 0 = none. */
const SUIT_KEYS: Readonly<Record<number, string>> = {
  0: 'suitNone',
  1: 'suitSpade',
  2: 'suitClub',
  3: 'suitHeart',
  4: 'suitDiamond',
};

/** Outcome i18n keys indexed by outcome value (0=none, 1=Sacar, 2=Puesta, 3=Codille). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeSacar', 'outcomePuesta', 'outcomeCodille'] as const;

/** Selectable trump suits with their playing-card symbols (1=♠ 2=♣ 3=♥ 4=♦). */
const TRUMP_CHOICES = [
  { code: 1, symbol: '♠' },
  { code: 2, symbol: '♣' },
  { code: 3, symbol: '♥' },
  { code: 4, symbol: '♦' },
] as const;

type OmbreGame = ReturnType<typeof useOmbreGame>;
type OmbreView = ReturnType<typeof useOmbreView>;
type OmbreContext = GamePageContext<OmbreResponse, OmbreGame, OmbreView>;

function useOmbreView({ game }: { game: OmbreGame }) {
  // Two-stage bid flow. Stage 1 picks the bid type (entrar/solo/pass); choosing
  // entrar or solo sets `pendingBid` (1 or 2) and advances to stage 2, where a
  // trump suit is picked and the declaration is confirmed. `null` = stage 1.
  const [pendingBid, setPendingBid] = useState<number | null>(null);
  // Trump suit chosen for a pending entrar/solo declaration (null until picked).
  const [selectedTrump, setSelectedTrump] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('ombre', OMBRE_PHASE_KEYS);

  // Stage 1 → pass immediately, or stage into trump selection for entrar/solo.
  const chooseBid = (bid: number) => {
    const highestBid = game.state?.highestBid ?? 0;
    if (bid > 0 && bid <= highestBid) return;
    if (bid === 0) {
      game.handleBid(0);
      return;
    }
    setPendingBid(bid);
    setSelectedTrump(null);
  };

  // Stage 2 → back to bid-type selection, discarding the pending choice.
  const cancelBid = () => {
    setPendingBid(null);
    setSelectedTrump(null);
  };

  // Stage 2 → confirm the pending declaration with the chosen trump suit.
  const confirmBid = () => {
    if (pendingBid === null || selectedTrump === null) return;
    game.handleBid(pendingBid, selectedTrump);
    setPendingBid(null);
    setSelectedTrump(null);
  };

  return {
    cardWidth,
    isMobile,
    phaseNames,
    pendingBid,
    selectedTrump,
    setSelectedTrump,
    chooseBid,
    cancelBid,
    confirmBid,
  };
}

function Board({ state, t, view }: OmbreContext) {
  const { cardWidth, isMobile } = view;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isBidPhase = state.phase === OmbrePhase.BID;
  const isTrickEnd = state.phase === OmbrePhase.TRICK_END;
  const isRoundEnd = state.phase === OmbrePhase.ROUND_END;
  const isGameEnd = state.phase === OmbrePhase.GAME_END || state.gameEndFlag;
  const highestBidder = state.players.find((player) => player.id === state.highestBidderIdx);
  const trumpLabel = state.trumpSuit >= 1 ? t(SUIT_KEYS[state.trumpSuit]) : t('suitNone');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">{t('winningBid', { bid: t(BID_KEYS[state.winningBid]) })}</span>
        <span>{t('trump', { suit: trumpLabel })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
            winnerLabel={
              isTrickEnd
                ? t('previousTrickWinner', { name: findPlayerName(state.players, state.lastTrickWinner) })
                : undefined
            }
            dataTutorial="ombre-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="ombre-info">
          <section aria-label={t('bidHistory')} className="mb-2 p-2 rounded bg-ds-surface text-ds-text-muted text-sm">
            <div className="text-ds-text-primary font-semibold">{t('bidHistory')}</div>
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5">
                {t('bidDeclaration', {
                  name: playerName(p.id, p.isHuman),
                  bid: state.bidActed[p.id]
                    ? t(state.bids[p.id] === 0 ? 'bidPass' : BID_KEYS[state.bids[p.id]])
                    : t('bidWaiting'),
                  trump:
                    state.bidActed[p.id] && state.bids[p.id] > 0
                      ? t('bidTrumpSuffix', { suit: t(SUIT_KEYS[state.bidTrump[p.id]]) })
                      : '',
                  highest: state.highestBidderIdx === p.id ? t('highestBidderSuffix') : '',
                })}
              </div>
            ))}
            {isBidPhase && highestBidder && (
              <div className="text-ds-text-primary">
                {t('currentHighestBid', {
                  name: playerName(highestBidder.id, highestBidder.isHuman),
                  bid: t(BID_KEYS[state.highestBid]),
                })}
              </div>
            )}
          </section>
          {/* Per-player match scores with Ombre badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                <span className={p.isOmbre ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                </span>
                {p.isOmbre && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('ombreBadge')}</span>
                )}
              </div>
            ))}
          </div>

          {/* Players: cards / tricks */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })}
                </div>
              ))}
            </div>
          )}

          {/* Round result: the deal outcome (Sacar / Puesta / Codille) */}
          {(isRoundEnd || isGameEnd) && state.outcome > 0 && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
              {state.ombreIdx >= 0 && (
                <div>{t('roundResult.ombre', { name: playerName(state.ombreIdx, state.ombreIdx === humanIdx) })}</div>
              )}
              <div className="mt-2">
                <div className="text-ds-text-primary">{t('roundResult.scoreChanges')}</div>
                {state.players.map((p) => {
                  const delta = state.playerScoreDeltas[p.id];
                  return (
                    <div key={p.id}>
                      {t('roundResult.scoreChange', {
                        name: playerName(p.id, p.isHuman),
                        delta: formatSignedDelta(delta),
                      })}
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: OmbreContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);

  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="ombre-hint-live" role="status" aria-live="polite">
        {humanPlayer && state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintMessage', {
              cards: state.hint.cardIndices
                .map((i) => cardAlt(humanPlayer.cards[i] as NonNullable<(typeof humanPlayer.cards)[number]>))
                .join(t('listSeparator')),
              reason: t(`hint.${state.hint.reason}`),
            })}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand({ state, game, t, view }: OmbreContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === OmbrePhase.BID;
  const isPlayPhase = state.phase === OmbrePhase.PLAY;
  const canBid = isBidPhase && state.isHumanBidTurn;
  const canPlay = isPlayPhase && state.isHumanTurn;

  // Badge the three matadors (Spadille ♠A / Manille = trump 7 / Basto ♣A) in
  // the human's hand once trump is decided. Ring only — never blocks clicks.
  const matadorBadgeFor = (idx: number): { glyph: string; title: string } | null => {
    const card = humanPlayer?.cards[idx];
    if (!card) return null;
    const rank = matadorRank(card, state.trumpSuit);
    if (rank === null) return null;
    return { glyph: String(rank), title: t(MATADOR_NAME_KEY[rank]) };
  };

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="ombre-prompt-live" role="status" aria-live="polite">
        {canBid && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="ombre-bid-prompt">
            {t('bidPhase')}
          </div>
        )}
      </div>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          dataTutorialPrefix="ombre"
          validIndices={canPlay ? state.playableIndices : undefined}
          highlightIndices={canPlay && isRequestedHint(state) ? state.hint?.cardIndices : undefined}
          restrictedTooltip={t('playButton')}
          cardBadgeFor={matadorBadgeFor}
        />
      )}
    </>
  );
}

function primaryActions({ game, state, t, busy, view }: OmbreContext) {
  const isBidPhase = state.phase === OmbrePhase.BID;
  const isPlayPhase = state.phase === OmbrePhase.PLAY;
  const isTrickEnd = state.phase === OmbrePhase.TRICK_END;
  const isRoundEnd = state.phase === OmbrePhase.ROUND_END;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canPlay = isPlayPhase && state.isHumanTurn;
  const loading = busy;

  if (!canBid && !canPlay && !isTrickEnd && !isRoundEnd) {
    return null;
  }

  const { pendingBid, selectedTrump, setSelectedTrump, chooseBid, cancelBid, confirmBid } = view;

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="ombre-action-buttons">
      {canBid && pendingBid === null && (
        <div className="flex flex-wrap gap-2 items-center" data-testid="ombre-bid-stage1">
          <span className="text-ds-text-muted text-sm">{t('chooseBidType')}:</span>
          {([1, 2] as const).map((bid) => {
            const unavailable = bid <= state.highestBid;
            const reasonId = `ombre-bid-unavailable-${bid}`;
            return (
              <span key={bid}>
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => chooseBid(bid)}
                  disabled={loading}
                  aria-disabled={unavailable || undefined}
                  aria-describedby={unavailable ? reasonId : undefined}
                >
                  {t(BID_KEYS[bid])}
                </button>
                {unavailable && (
                  <span id={reasonId} className="sr-only">
                    {t('bidMustExceedWinningBid')}
                  </span>
                )}
              </span>
            );
          })}
          <button type="button" className={btnSecondary} onClick={() => chooseBid(0)} disabled={loading}>
            {t('bidPass')}
          </button>
        </div>
      )}
      {canBid && pendingBid !== null && (
        <div className="flex flex-wrap gap-2 items-center" data-testid="ombre-bid-stage2">
          <span className="text-ds-text-muted text-sm">{t('chooseTrumpFor', { bid: t(BID_KEYS[pendingBid]) })}:</span>
          {TRUMP_CHOICES.map((c) => (
            <button
              key={c.code}
              type="button"
              className={selectedTrump === c.code ? btnPrimary : btnSecondary}
              onClick={() => setSelectedTrump(c.code)}
              disabled={loading}
              aria-label={t(SUIT_KEYS[c.code])}
              aria-pressed={selectedTrump === c.code}
            >
              {c.symbol}
            </button>
          ))}
          <button
            type="button"
            className={btnSuccess}
            onClick={confirmBid}
            disabled={loading || selectedTrump === null}
            data-testid="ombre-bid-confirm"
          >
            {t('confirmBid')}
          </button>
          <button
            type="button"
            className={btnSecondary}
            onClick={cancelBid}
            disabled={loading}
            data-testid="ombre-bid-back"
          >
            {t('bidBack')}
          </button>
        </div>
      )}
      {canPlay && (
        <>
          <button type="button" className={btnSecondary} onClick={() => void game.exec('hint')} disabled={loading}>
            {t('requestHint')}
          </button>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handlePlay}
            disabled={loading || game.selectedCardIndices.length !== 1}
          >
            {t('playButton')}
          </button>
        </>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={loading}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={loading}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const ombreDefinition: GameDefinition<OmbreResponse, OmbreGame, Parameters<typeof ombreApi.exec>, OmbreView> = {
  key: 'ombre',
  useGame: useOmbreGame,
  useView: useOmbreView,
  tutorial: OMBRE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 9 },
  cli: () => ({
    parseCommand: parseOmbreCommand,
    formatResponse: formatOmbreState,
    helpText: OMBRE_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === OmbrePhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerPlayer === humanIdx,
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn,
  }),
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.ombreConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.ombreConfig.targetRounds,
          options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetRounds', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  resetTutorialId: 'ombre-reset-button',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Ombre (Hombre) game page: a 3-player soloist-vs-coalition 40-card Spanish-deck trick-taker with a bid + trump auction. */
export const OmbrePage = createGamePage(ombreDefinition);
