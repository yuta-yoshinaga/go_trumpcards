import type { TFunction } from 'i18next';
import { useEffect } from 'react';
import type { cinchApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useCinchGame } from '../hooks/useCinchGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { CinchResponse } from '../types/card';
import { CinchPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { estimateCinchBidStrength } from '../utils/cinchBidStrength';
import { CINCH_HELP, parseCinchCommand } from '../utils/cli/commands/cinchCommands';
import { formatCinchState } from '../utils/cli/formatters/cinchFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** i18n suit-name key by suit number (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_KEYS: Readonly<Record<number, string>> = { 1: 'spade', 2: 'club', 3: 'heart', 4: 'diamond' };

/** Hearts and diamonds render red; spades and clubs stay the default text color. */
const isRedSuit = (suit: number): boolean => suit === 3 || suit === 4;

/** Selectable trump suits named by the bid winner. */
const TRUMP_SUITS = [1, 2, 3, 4] as const;

/** Maximum Cinch bid (all 14 points). */
const MAX_BID = 14;

/** Cinch tutorial step definitions. */
const CINCH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cinch-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cinch-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cinch-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="cinch-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="cinch-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const CINCH_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CinchPhase.BID]: 'bid',
  [CinchPhase.NAME_TRUMP]: 'nameTrump',
  [CinchPhase.PLAY]: 'play',
  [CinchPhase.TRICK_END]: 'trickEnd',
  [CinchPhase.ROUND_END]: 'roundEnd',
  [CinchPhase.GAME_END]: 'gameEnd',
};

type CinchGame = ReturnType<typeof useCinchGame>;
type CinchView = ReturnType<typeof useCinchView>;
type CinchContext = GamePageContext<CinchResponse, CinchGame, CinchView>;

function useCinchView({ game }: { game: CinchGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('cinch', CINCH_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

/** Localized suit name for a suit number (e.g. 3 -> "ハート"). */
const suitLabel = (t: TFunction, suit: number): string => (SUIT_KEYS[suit] ? t(`suit.${SUIT_KEYS[suit]}`) : '');

/** Colored suit symbol: hearts/diamonds red, spades/clubs default. */
const renderSuitSymbol = (suit: number) => (
  <span className={isRedSuit(suit) ? 'text-ds-error-text' : undefined}>{suitSymbolAt(suit, '-')}</span>
);

function Board({ game, state, t, busy, view }: CinchContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === CinchPhase.BID;
  const isRoundEnd = state.phase === CinchPhase.ROUND_END;
  const isGameEnd = state.phase === CinchPhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.bidPlayerIdx === humanIdx && isHumanTurn;

  // Rough, hand-only strength guide shown alongside the bid buttons so a player
  // can gauge how high to bid. Purely advisory — not a capture guarantee.
  const bidStrength = canBid && humanPlayer ? estimateCinchBidStrength(humanPlayer.cards) : null;

  // Legal bids: pass (0) plus any raise strictly above the current highest bid, up to 14.
  const minRaise = Math.max(1, state.currentBid + 1);
  const bidChoices: number[] = [];
  for (let b = minRaise; b <= MAX_BID; b++) bidChoices.push(b);

  return (
    <>
      {canBid && (
        <div data-tutorial="cinch-action-buttons">
          <div className="flex flex-wrap gap-2" data-testid="cinch-bid-buttons">
            <button type="button" className={btnSecondary} onClick={() => game.handleBid(0)} disabled={busy}>
              {t('bidPass')}
            </button>
            {bidChoices.map((b) => (
              <button key={b} type="button" className={btnPrimary} onClick={() => game.handleBid(b)} disabled={busy}>
                {b}
              </button>
            ))}
          </div>
          {/* Shown right under the bid buttons: it is what the bid is based on. */}
          {bidStrength && (
            <div
              className="mb-2 mx-auto max-w-xl p-2 rounded bg-black/30 text-ds-text-muted text-xs text-center"
              data-testid="cinch-bid-strength"
            >
              <div className="text-ds-text-primary" data-testid="cinch-bid-strength-range">
                {t('bidStrength.range', { min: bidStrength.minPoints, max: bidStrength.maxPoints })}
              </div>
              <div data-testid="cinch-bid-strength-best">
                {t('bidStrength.best', {
                  symbol: suitSymbolAt(bidStrength.bestSuit, '-'),
                  suit: suitLabel(t, bidStrength.bestSuit),
                  points: bidStrength.pointsBySuit[bidStrength.bestSuit],
                })}
              </div>
              <div
                className="mt-1 grid grid-cols-2 gap-x-3 text-left sm:grid-cols-4"
                data-testid="cinch-bid-strength-suits"
              >
                {TRUMP_SUITS.map((suit) => (
                  <div key={suit}>
                    <span className="text-ds-text-primary">
                      {suitSymbolAt(suit, '-')} {suitLabel(t, suit)}
                    </span>
                    <div>
                      {t('bidStrength.suitEstimate', {
                        points: bidStrength.pointsBySuit[suit],
                        trumps: bidStrength.trumpCountBySuit[suit],
                        control: bidStrength.controlBySuit[suit],
                      })}
                    </div>
                  </div>
                ))}
              </div>
              <div className="mt-1">{t('bidStrength.controlMethod')}</div>
              <div className="mt-1">
                <span className="text-ds-text-primary">{t('bidStrength.legendTitle')}:</span> {t('bidStrength.legend')}
              </div>
              <div className="mt-1 italic">{t('bidStrength.note')}</div>
            </div>
          )}
        </div>
      )}
      <div className="flex flex-wrap justify-center gap-x-4 gap-y-1 text-ds-text-primary text-center mb-2">
        <span className="whitespace-nowrap">{t('deal', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber, total: state.totalTricks })}</span>
        <span className="whitespace-nowrap">{t('highBid', { bid: state.currentBid })}</span>
        <span className="whitespace-nowrap" data-testid="cinch-trump-header">
          {t('trumpLabel')}:{' '}
          {state.trumpSuit >= 1 ? (
            <>
              {renderSuitSymbol(state.trumpSuit)} {suitLabel(t, state.trumpSuit)}
            </>
          ) : (
            '-'
          )}
        </span>
        <span className="whitespace-nowrap">{t('target', { points: state.config.pointLimit })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={view.cardWidth}
            label={t('currentTrick')}
            dataTutorial="cinch-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="cinch-info">
          {/* Per-player match scores with a Bidder badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                <span className={p.id === state.bidWinnerIdx ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.totalScore })}
                </span>
                {p.id === state.bidWinnerIdx && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('bidderBadge')}</span>
                )}
              </div>
            ))}
          </div>

          {isBidPhase && (
            <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="cinch-bid-status">
              {state.players.map((p) => (
                <div key={p.id} className="py-0.5">
                  {playerName(p.id, p.isHuman)}:{' '}
                  {p.bid < 0 ? t('bidStatus.unbid') : p.bid === 0 ? t('bidStatus.pass') : p.bid}
                  {p.id === state.dealerIdx && (
                    <span className={`ml-2 px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                      {t('dealerBadge')}
                    </span>
                  )}
                </div>
              ))}
              {state.dealerIdx === humanIdx && (
                <div className="mt-1 text-ds-text-primary" data-testid="cinch-dealer-bid-notice">
                  {t('forcedBidNotice')}
                </div>
              )}
            </div>
          )}

          {/* Players: cards / tricks */}
          {view.isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })} | {t('dealPoints', { points: p.dealPoints })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })} | {t('dealPoints', { points: p.dealPoints })}
                </div>
              ))}
            </div>
          )}

          {/* Deal result: per-player gained points */}
          {(isRoundEnd || isGameEnd) && state.lastDealDetail && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="cinch-deal-result">
              <div className="mb-1 text-ds-text-primary">{t('dealResult.title')}</div>
              {state.lastDealDetail.bidderIdx >= 0 && (
                <div
                  className={`mb-1 ${state.lastDealDetail.setBack ? 'text-ds-error-text font-semibold' : 'text-ds-text-primary'}`}
                  data-testid="cinch-bidder-detail"
                >
                  {t(state.lastDealDetail.setBack ? 'dealResult.bidderSet' : 'dealResult.bidderMade', {
                    name: playerName(state.lastDealDetail.bidderIdx, state.lastDealDetail.bidderIdx === humanIdx),
                    bid: state.lastDealDetail.bid,
                    captured: state.lastDealDetail.points[state.lastDealDetail.bidderIdx] ?? 0,
                    delta: formatSignedDelta(state.lastDealDetail.gained[state.lastDealDetail.bidderIdx] ?? 0),
                  })}
                </div>
              )}
              {state.players.map((p) => {
                const isSetBackRow = p.id === state.lastDealDetail?.bidderIdx && state.lastDealDetail?.setBack;
                return (
                  <div
                    key={p.id}
                    className={isSetBackRow ? 'text-ds-error-text font-semibold' : undefined}
                    data-testid={isSetBackRow ? 'cinch-setback-row' : undefined}
                  >
                    {t('dealResult.gained', {
                      name: playerName(p.id, p.isHuman),
                      points: state.lastDealDetail?.gained[p.id] ?? 0,
                    })}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterLead({ state, t }: CinchContext) {
  return (
    <div data-testid="cinch-trump-selection-live" role="status" aria-live="polite">
      {state.trumpSuit >= 1 && (
        <div className="mb-1 text-center text-sm text-ds-accent font-semibold">
          {t('trumpConfirmed', { suit: suitLabel(t, state.trumpSuit) })}
        </div>
      )}
    </div>
  );
}

function FooterStatus({ state, t, frontendHint }: CinchContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="cinch-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndices &&
              state.hint.cardIndices.length > 0 &&
              ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand({ game, state, t, view }: CinchContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === CinchPhase.BID;
  const isNameTrumpPhase = state.phase === CinchPhase.NAME_TRUMP;
  const isPlayPhase = state.phase === CinchPhase.PLAY;

  const canBid = isBidPhase && state.bidPlayerIdx === humanIdx && isHumanTurn;
  const canNameTrump = isNameTrumpPhase && state.bidWinnerIdx === humanIdx;
  const canPlay = isPlayPhase && isHumanTurn;

  return (
    <>
      {isBidPhase && !canBid && (
        <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cinch-bid-cpu">
          {t('bidCpu', { id: state.bidPlayerIdx })}
        </div>
      )}
      {isNameTrumpPhase && !canNameTrump && (
        <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cinch-trump-cpu">
          {t('trumpCpu', { id: state.bidWinnerIdx })}
        </div>
      )}
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="cinch-prompt-live" role="status" aria-live="polite">
        {canBid && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cinch-bid-prompt">
            {t('bidPrompt')}
          </div>
        )}
        {canNameTrump && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cinch-trump-prompt">
            {t('trumpPrompt')}
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
          dataTutorialPrefix="cinch"
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
        />
      )}
    </>
  );
}

function primaryActions({ game, state, t, busy }: CinchContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isNameTrumpPhase = state.phase === CinchPhase.NAME_TRUMP;
  const isPlayPhase = state.phase === CinchPhase.PLAY;
  const isRoundEnd = state.phase === CinchPhase.ROUND_END;

  const canNameTrump = isNameTrumpPhase && state.bidWinnerIdx === humanIdx;
  const canPlay = isPlayPhase && isHumanTurn;

  if (!canNameTrump && !canPlay && !isRoundEnd) return null;

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="cinch-action-buttons">
      {canNameTrump && (
        <div className="flex flex-wrap gap-2" data-testid="cinch-trump-buttons">
          {TRUMP_SUITS.map((suit) => (
            <button
              key={suit}
              type="button"
              className={btnPrimary}
              onClick={() => game.handleNameTrump(suit)}
              disabled={busy}
              aria-label={suitLabel(t, suit)}
            >
              {renderSuitSymbol(suit)}
            </button>
          ))}
        </div>
      )}
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextDeal} disabled={busy}>
          {t('nextDeal')}
        </button>
      )}
    </div>
  );
}

const cinchDefinition: GameDefinition<CinchResponse, CinchGame, Parameters<typeof cinchApi.exec>, CinchView> = {
  key: 'cinch',
  useGame: useCinchGame,
  useView: useCinchView,
  tutorial: CINCH_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 9 },
  cli: () => ({
    parseCommand: parseCinchCommand,
    formatResponse: formatCinchState,
    helpText: CINCH_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === CinchPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerIdx === humanIdx,
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
          value: game.cinchConfig.cpuDifficulty,
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
          value: game.cinchConfig.pointLimit,
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
  resetTutorialId: 'cinch-reset-button',
  // The hand-wired page rendered a bare reset button; a class-less div lays out the same, unlike the default flex row.
  actionsClassName: '',
  Board,
  FooterHand,
  FooterLead,
  FooterStatus,
};

/** Renders the Cinch (Double Pedro) game page: a 4-player 52-card All-Fours/Pitch-family bidding trick-taker. */
export const CinchPage = createGamePage(cinchDefinition);
