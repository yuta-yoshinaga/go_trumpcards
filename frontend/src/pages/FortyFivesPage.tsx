import type { TFunction } from 'i18next';
import { useEffect } from 'react';
import type { fortyFivesApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useFortyFivesGame } from '../hooks/useFortyFivesGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeInfoColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { FortyFivesResponse } from '../types/card';
import { FortyFivesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { FORTY_FIVES_HELP, parseFortyFivesCommand } from '../utils/cli/commands/fortyFivesCommands';
import { formatFortyFivesState } from '../utils/cli/formatters/fortyFivesFormatter';
import { fortyFivesTopTrumpIndices } from '../utils/fortyFivesTopTrump';
import { formatFortyFivesTrumpOrder } from '../utils/fortyFivesTrump';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Bid button options (Pass / 15 / 20 / 25 Jink). */
const BIDS: { value: number; key: string }[] = [
  { value: 0, key: 'bid.pass' },
  { value: 15, key: 'bid.fifteen' },
  { value: 20, key: 'bid.twenty' },
  { value: 25, key: 'bid.twentyfive' },
];

/** Auction Forty-Fives tutorial step definitions. */
const FORTY_FIVES_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="fortyfives-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="fortyfives-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fortyfives-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fortyfives-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const FORTY_FIVES_PHASE_KEYS: Readonly<Record<number, string>> = {
  [FortyFivesPhase.BID]: 'bid',
  [FortyFivesPhase.PLAY]: 'play',
  [FortyFivesPhase.TRICK_END]: 'trickEnd',
  [FortyFivesPhase.ROUND_END]: 'roundEnd',
  [FortyFivesPhase.GAME_END]: 'gameEnd',
};

type FortyFivesGame = ReturnType<typeof useFortyFivesGame>;
type FortyFivesView = ReturnType<typeof useFortyFivesView>;
type FortyFivesContext = GamePageContext<FortyFivesResponse, FortyFivesGame, FortyFivesView>;

function useFortyFivesView({ game }: { game: FortyFivesGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('fortyfives', FORTY_FIVES_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

// Map a bid value to its localized name (e.g. 25 -> "25 (Jink)"), matching the
// CUI's fortyFivesBidName; fall back to the raw number for any unknown value.
function bidName(t: TFunction, value: number): string {
  const bid = BIDS.find((b) => b.value === value);
  return bid ? t(bid.key) : String(value);
}

function FortyFivesMessageBox({ state }: Pick<FortyFivesContext, 'state'>) {
  return (
    <GameMessageBox
      message={state.message}
      messageCode={state.messageCode}
      messageParams={state.messageParams}
      alwaysVisible
      bareWhenEmpty
      testId="forty-fives-game-message"
    />
  );
}

function Board({ state, t, view }: FortyFivesContext) {
  const { isMobile, cardWidth } = view;

  const isBidPhase = state.phase === FortyFivesPhase.BID;
  const isPlayPhase = state.phase === FortyFivesPhase.PLAY;
  const isRoundEnd = state.phase === FortyFivesPhase.ROUND_END;
  const isGameEnd = state.phase === FortyFivesPhase.GAME_END || state.gameEndFlag;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const humanTeam = humanIdx >= 0 ? humanIdx % 2 : -1;

  const trumpSymbol = state.trumpSuit === 0 ? t('noTrump') : suitSymbolAt(state.trumpSuit, '?');

  const contractLabel = state.contract === 0 ? t('contractUndecided') : bidName(t, state.contract);

  // Live round-points readout. Each trick awards 5 points to the winning team and a
  // round has 5 tricks, so points accrue up to 25. Tricks resolved so far is the total
  // points awarded divided by 5, and the points still up for grabs follow from that.
  const POINTS_PER_TRICK = 5;
  const TOTAL_TRICKS = 5;
  const roundPointsA = state.roundTeamPoints[0] ?? 0;
  const roundPointsB = state.roundTeamPoints[1] ?? 0;
  const tricksResolved = (roundPointsA + roundPointsB) / POINTS_PER_TRICK;
  const remainingPoints = Math.max(0, (TOTAL_TRICKS - tricksResolved) * POINTS_PER_TRICK);

  // Contract progress for the declaring team (known once bidding resolves). The team makes
  // the bid when its accrued points reach the contract; it can no longer make it once even
  // every remaining trick would not close the gap.
  const declarerTeam = state.declarerIdx >= 0 ? state.declarerIdx % 2 : -1;
  const declarerPoints = declarerTeam >= 0 ? (state.roundTeamPoints[declarerTeam] ?? 0) : 0;
  const contractStatus: 'made' | 'failed' | 'needMore' =
    declarerPoints >= state.contract
      ? 'made'
      : declarerPoints + remainingPoints < state.contract
        ? 'failed'
        : 'needMore';
  const contractRemaining = Math.max(0, state.contract - declarerPoints);
  const contractStatusColor =
    contractStatus === 'made'
      ? 'text-ds-success'
      : contractStatus === 'failed'
        ? 'text-ds-error-text'
        : 'text-ds-warning';

  /** Renders the per-player rows for one team (seats with seat%2 === team). */
  const renderTeamPlayers = (team: number) => {
    return state.players
      .filter((p) => p.id % 2 === team)
      .map((p) => (
        <div key={p.id} className="py-0.5 flex items-center gap-2 text-ds-text-muted text-sm">
          <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
          </span>
          {p.isDeclarer && (
            <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('declarerBadge')}</span>
          )}
          {(isBidPhase || isPlayPhase) && p.id === state.currentPlayerIdx && (
            <span className={`px-1.5 py-0.5 rounded text-xs ${badgeInfoColors}`}>{t('currentTurn')}</span>
          )}
        </div>
      ));
  };

  return (
    <>
      <div
        className="flex flex-wrap justify-center gap-x-4 gap-y-1 text-ds-text-primary mb-2"
        data-tutorial="fortyfives-info"
      >
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span className="whitespace-nowrap">{t('trump', { suit: trumpSymbol })}</span>
        <span className="whitespace-nowrap">{t('target', { points: state.config.targetPoints })}</span>
      </div>

      <div className="text-ds-text-muted text-center mb-2 text-sm">
        {state.declarerIdx >= 0
          ? t('contractLine', {
              name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false),
              team: state.declarerIdx % 2 === 0 ? t('team.a') : t('team.b'),
              contract: contractLabel,
            })
          : t('contractUndecided')}
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="fortyfives-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div>
          {state.trumpSuit > 0 && (
            <section className="mb-2 p-2 rounded bg-black/30" data-testid="ff-trump-legend">
              <h2 className="text-ds-text-muted text-sm">{t('trumpLegend.title')}</h2>
              <div className="mt-1 text-ds-text-muted text-xs">
                <div className="mb-1">{t('trumpLegend.caption')}</div>
                <div className="flex flex-wrap items-center gap-x-1 gap-y-0.5">
                  {formatFortyFivesTrumpOrder(trumpSymbol).map((symbol, i, order) => (
                    <span key={symbol} className="inline-flex items-center gap-1">
                      <span className={i < 3 ? 'font-mono text-ds-warning' : 'font-mono text-ds-text-primary'}>
                        {symbol}
                      </span>
                      {i < order.length - 1 && <span aria-hidden="true">&gt;</span>}
                    </span>
                  ))}
                </div>
              </div>
            </section>
          )}

          {/* Team match scores */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            <div>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] ?? 0 })}</div>
            <div>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] ?? 0 })}</div>
            <div className="mt-1">
              {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
            </div>
          </div>

          {/* Players grouped by team, with the declarer badge */}
          {([0, 1] as const).map((team) => (
            <div key={team} className="mb-2 p-2 rounded bg-black/30">
              <div className="text-ds-text-primary text-xs mb-1">{team === 0 ? t('team.a') : t('team.b')}</div>
              {isMobile ? (
                <details>
                  <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                  <div className="mt-1">{renderTeamPlayers(team)}</div>
                </details>
              ) : (
                renderTeamPlayers(team)
              )}
            </div>
          ))}

          {/* Live round points during bidding/play (matches the CUI readout); the
              round-result block below takes over once the round ends. */}
          {!(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="ff-live-points">
              <div className="mb-1 text-ds-text-primary">{t('livePoints.title')}</div>
              <div>{t('roundResult.teamA', { points: roundPointsA })}</div>
              <div>{t('roundResult.teamB', { points: roundPointsB })}</div>
              {declarerTeam >= 0 && state.contract > 0 && (
                <div className="mt-1 text-ds-text-primary" data-testid="ff-contract-progress">
                  {t('livePoints.contract', {
                    team: declarerTeam === 0 ? t('team.a') : t('team.b'),
                    got: declarerPoints,
                    contract: state.contract,
                  })}
                  {' — '}
                  <span className={contractStatusColor}>
                    {t(`livePoints.status.${contractStatus}`, { remaining: contractRemaining })}
                  </span>
                </div>
              )}
            </div>
          )}

          {/* Round result: points per team */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="ff-round-result">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.teamA', { points: state.roundTeamPoints[0] ?? 0 })}</div>
              <div>{t('roundResult.teamB', { points: state.roundTeamPoints[1] ?? 0 })}</div>
              {declarerTeam >= 0 && state.contract > 0 && (
                <div className="mt-1 text-ds-text-primary" data-testid="ff-contract-result">
                  {t('roundResult.contract', {
                    team: declarerTeam === 0 ? t('team.a') : t('team.b'),
                    got: declarerPoints,
                    contract: state.contract,
                    status: t(`roundResult.status.${contractStatus === 'made' ? 'made' : 'failed'}`),
                  })}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {isMobile && <FortyFivesMessageBox state={state} />}
    </>
  );
}

function FooterHand({ game, state, t, view }: FortyFivesContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const isPlayPhase = state.phase === FortyFivesPhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn;
  // 固定の最上位切り札 (切り札の5・切り札のJ・♥A) は手札を見ても分からない。
  // ♥A は切り札スート外でも切り札扱いで、持っているとフォロー義務が外れる (#5643)。
  const topTrumpIndices = fortyFivesTopTrumpIndices(humanPlayer.cards ?? [], state.trumpSuit);

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="fortyfives"
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
      trumpIndices={topTrumpIndices}
      trumpTitle={t('topTrumpTitle')}
    />
  );
}

function FooterStatus({ state, t, frontendHint }: FortyFivesContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="fortyfives-hint-live" role="status" aria-live="polite">
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

      <div className="text-xs text-ds-text-muted mb-2" data-testid="ff-bid-history">
        <span className="font-semibold mr-1">{t('bidHistoryTitle')}:</span>
        <span>
          {state.players
            .map(
              (p, i) =>
                `${playerName(p.id, p.isHuman)}=${state.bidDone[i] ? bidName(t, state.bids[i]) : t('bidNotYet')}`,
            )
            .join(t('listSeparator'))}
        </span>
      </div>
    </>
  );
}

function primaryActions({ game, state, t, busy }: FortyFivesContext) {
  const isBidPhase = state.phase === FortyFivesPhase.BID;
  const isPlayPhase = state.phase === FortyFivesPhase.PLAY;
  const isTrickEnd = state.phase === FortyFivesPhase.TRICK_END;
  const isRoundEnd = state.phase === FortyFivesPhase.ROUND_END;
  const isHumanTurn = state.isHumanTurn;
  const isHumanBidTurn = state.isHumanBidTurn;

  const canPlay = isPlayPhase && isHumanTurn;

  if (!((isBidPhase && isHumanBidTurn) || canPlay || isTrickEnd || isRoundEnd)) {
    return null;
  }

  // The current highest (non-pass) bid; a new non-pass bid must beat it.
  const highestBid = Math.max(0, ...state.bids);
  // Look up the holder via the players array (playerName expects a player id, not an index).
  const highestBidder = highestBid > 0 ? state.players[state.bids.indexOf(highestBid)] : undefined;
  const highestBidderName = highestBidder ? playerName(highestBidder.id, highestBidder.isHuman) : '';

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="fortyfives-action-buttons">
      {isBidPhase && isHumanBidTurn && (
        <>
          <span className="text-xs text-ds-text-muted self-center mr-1">{t('bidPrompt')}</span>
          <span className="text-xs text-ds-text-muted self-center mr-1" data-testid="ff-highest-bid">
            {highestBid > 0
              ? t('bidHighest', { bid: bidName(t, highestBid), player: highestBidderName })
              : t('bidNone')}
          </span>
          {BIDS.map((b) => {
            // Pass (0) is always allowed; a non-pass bid must beat the current highest.
            const tooLow = b.value !== 0 && b.value <= highestBid;
            const disabled = busy || tooLow;
            const reason = tooLow ? t('bidDisabledReason', { currentBid: highestBid }) : undefined;
            // The title lives on the wrapping span: browsers suppress native tooltips on
            // disabled buttons, so hovering the span still surfaces the reason.
            return (
              <span key={b.value} title={reason} data-testid={`bid-wrap-${b.value}`}>
                <button
                  type="button"
                  className="px-3 py-2 rounded-lg bg-ds-info text-ds-text-on-accent text-sm disabled:opacity-40"
                  onClick={() => game.handleBid(b.value)}
                  disabled={disabled}
                  aria-disabled={disabled}
                  aria-label={reason ? `${t(b.key)} — ${reason}` : undefined}
                  data-testid={`bid-${b.value}`}
                >
                  {t(b.key)}
                </button>
              </span>
            );
          })}
        </>
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
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const fortyFivesDefinition: GameDefinition<
  FortyFivesResponse,
  FortyFivesGame,
  Parameters<typeof fortyFivesApi.exec>,
  FortyFivesView
> = {
  key: 'fortyfives',
  useGame: useFortyFivesGame,
  useView: useFortyFivesView,
  tutorial: FORTY_FIVES_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseFortyFivesCommand,
    formatResponse: formatFortyFivesState,
    helpText: FORTY_FIVES_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === FortyFivesPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const humanTeam = humanIdx >= 0 ? humanIdx % 2 : -1;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerTeam === humanTeam,
    };
  },
  shell: ({ state }) => {
    const isGameEnd = state.phase === FortyFivesPhase.GAME_END || state.gameEndFlag;
    return {
      isHumanTurn: (state.isHumanTurn || state.isHumanBidTurn) && !isGameEnd,
    };
  },
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
          value: game.fortyFivesConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetPoints',
          label: t('settings.targetPoints'),
          value: game.fortyFivesConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  message: ({ state, view }) => (!view.isMobile ? <FortyFivesMessageBox state={state} /> : null),
  resetTutorialId: 'fortyfives-reset-button',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Auction Forty-Fives game page: an Irish/Canadian 4-player (2 vs 2) bidding trick-taker. */
export const FortyFivesPage = createGamePage(fortyFivesDefinition);
