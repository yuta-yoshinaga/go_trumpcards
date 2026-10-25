import { useEffect, useRef, useState } from 'react';
import type { twentyNineApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useTwentyNineGame } from '../hooks/useTwentyNineGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { TwentyNineResponse } from '../types/card';
import { TwentyNinePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbol, suitSymbolAt } from '../utils/cardAlt';
import { parseTwentyNineCommand, TWENTY_NINE_HELP } from '../utils/cli/commands/twentyNineCommands';
import { formatTwentyNineState } from '../utils/cli/formatters/twentyNineFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { twentyNineContractProgress } from '../utils/twentyNineContract';

/** Bid button options (Pass / 16 / 20 / 24 / 28). */
const BIDS: { value: number; key: string }[] = [
  { value: 0, key: 'bid.pass' },
  { value: 16, key: 'bid.sixteen' },
  { value: 20, key: 'bid.twenty' },
  { value: 24, key: 'bid.twentyfour' },
  { value: 28, key: 'bid.twentyeight' },
];

/** Twenty-Nine (29) tutorial step definitions. */
const TWENTY_NINE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="twentynine-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="twentynine-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="twentynine-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="twentynine-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TWENTY_NINE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TwentyNinePhase.BID]: 'bid',
  [TwentyNinePhase.PLAY]: 'play',
  [TwentyNinePhase.TRICK_END]: 'trickEnd',
  [TwentyNinePhase.ROUND_END]: 'roundEnd',
  [TwentyNinePhase.GAME_END]: 'gameEnd',
};

type TwentyNineGame = ReturnType<typeof useTwentyNineGame>;
type TwentyNineContext = GamePageContext<TwentyNineResponse, TwentyNineGame, TwentyNineView>;

function useTwentyNineView({ game }: { game: TwentyNineGame }) {
  const { state } = game;

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('twentynine', TWENTY_NINE_PHASE_KEYS);

  // Flash a transient banner the moment the hidden trump is revealed mid-play,
  // so the reveal is not missed while a CPU is playing (mirrors SpoilFive's potDelta).
  const [showTrumpBanner, setShowTrumpBanner] = useState(false);
  const prevTrumpRevealedRef = useRef<boolean | null>(null);
  useEffect(() => {
    const revealed = state?.trumpRevealed;
    if (revealed == null) return;
    const prev = prevTrumpRevealedRef.current;
    prevTrumpRevealedRef.current = revealed;
    if (prev === false && revealed === true) {
      setShowTrumpBanner(true);
      const id = setTimeout(() => setShowTrumpBanner(false), 3000);
      return () => clearTimeout(id);
    }
  }, [state?.trumpRevealed]);

  const humanIdx = state?.players.findIndex((p) => p.isHuman) ?? -1;
  const humanTeam = humanIdx >= 0 ? humanIdx % 2 : -1;

  return { cardWidth, isMobile, phaseNames, showTrumpBanner, humanTeam, humanIdx };
}

type TwentyNineView = ReturnType<typeof useTwentyNineView>;

function AboveContent({ game, state, t }: TwentyNineContext) {
  if (
    game.twentyNineConfig.cpuDifficulty === state.config.cpuDifficulty &&
    game.twentyNineConfig.targetPoints === state.config.targetPoints
  ) {
    return null;
  }
  return <p className="px-4 text-sm text-ds-warning">{t('settings.pendingReset')}</p>;
}

function Board({ state, t, view }: TwentyNineContext) {
  const { isMobile, cardWidth, showTrumpBanner, humanTeam, humanIdx } = view;

  const isBidPhase = state.phase === TwentyNinePhase.BID;
  const isPlayPhase = state.phase === TwentyNinePhase.PLAY;
  const isRoundEnd = state.phase === TwentyNinePhase.ROUND_END;
  const isGameEnd = state.phase === TwentyNinePhase.GAME_END || state.gameEndFlag;
  const contractProgress = twentyNineContractProgress(state.declarerIdx, state.contract, state.roundTeamPoints);
  const contractStatusColor =
    contractProgress?.status === 'made'
      ? 'text-ds-success'
      : contractProgress?.status === 'failed'
        ? 'text-ds-error-text'
        : 'text-ds-warning';

  // The trump suit is hidden until trumpRevealed flips true mid-play.
  const revealedTrumpSymbol = state.trumpSuit === 0 ? t('noTrump') : suitSymbolAt(state.trumpSuit, '');
  const isHumanDeclarer = state.declarerIdx >= 0 && humanIdx === state.declarerIdx;
  const trumpSymbol = !state.trumpRevealed && !isHumanDeclarer ? t('hiddenTrump') : revealedTrumpSymbol;
  const trumpDisplay =
    !state.trumpRevealed && isHumanDeclarer ? t('trumpVisibleToYou', { suit: trumpSymbol }) : trumpSymbol;
  const leadCard = state.currentTrick[0]?.card;

  const bidHistory = state.players.map((player, seat) => {
    let committed = !isBidPhase;
    if (isBidPhase) {
      for (let step = 1; step < state.players.length; step++) {
        if ((state.dealerIdx + step) % state.players.length === state.currentPlayerIdx) break;
        if ((state.dealerIdx + step) % state.players.length === seat) committed = true;
      }
    }
    return { player, seat, committed, bid: state.bids[seat] };
  });

  const contractLabel = state.contract === 0 ? t('contractUndecided') : String(state.contract);

  /** Renders the per-player rows for one team (seats with seat%2 === team). */
  function renderTeamPlayers(team: number, cpuOnly = false) {
    return state.players
      .filter((p) => p.id % 2 === team && (!cpuOnly || !p.isHuman))
      .map((p) => (
        <div key={p.id} className="py-0.5 flex items-center gap-2 text-ds-text-muted text-sm">
          <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
          </span>
          {p.isDeclarer && (
            <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('declarerBadge')}</span>
          )}
        </div>
      ));
  }

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="twentynine-info">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">{t('trump', { suit: trumpDisplay })}</span>
        <span>{t('target', { points: state.config.targetPoints })}</span>
      </div>

      <details className="mb-3 text-sm">
        <summary className="cursor-pointer select-none text-ds-text-primary hover:text-ds-accent">
          {t('strengthReference.title')}
        </summary>
        <div className="mt-2 rounded-lg bg-ds-surface p-3 text-ds-text-primary">
          <div className="grid gap-3 sm:grid-cols-2">
            <div>
              <div className="mb-1 font-semibold">{t('strengthReference.trump')}</div>
              <div>{t('strengthReference.order')}</div>
            </div>
            <div>
              <div className="mb-1 font-semibold">{t('strengthReference.plain')}</div>
              <div>{t('strengthReference.order')}</div>
            </div>
          </div>
          <div className="mt-2 text-ds-text-muted">{t('strengthReference.points')}</div>
        </div>
      </details>

      {showTrumpBanner && (
        <div
          role="status"
          data-testid="tn-trump-reveal-banner"
          className={`mx-auto mb-2 w-fit rounded-lg px-4 py-1.5 text-center font-semibold motion-safe:animate-pulse ${badgeWarningColors}`}
        >
          {t('trumpRevealBanner', { suit: revealedTrumpSymbol })}
        </div>
      )}

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
          <div data-testid="tn29-lead-suit" className="mb-2 text-center text-ds-text-primary text-sm">
            {isPlayPhase && leadCard && (
              <span className="inline-flex items-center gap-2 rounded-lg bg-ds-surface px-3 py-1.5">
                {t('leadSuit', {
                  symbol: suitSymbol(leadCard.design),
                  suit: t(`suits.${leadCard.design.toLowerCase()}`),
                })}
              </span>
            )}
          </div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            cardDetailFor={(_card, trickCard) => {
              const index = state.currentTrick.findIndex((card) => card.playerIdx === trickCard.playerIdx);
              return t('trickCardPoints', { points: state.currentTrickPoints[index] ?? 0 });
            }}
            dataTutorial="twentynine-trick-display"
          />
          {state.isLastTrick && <p className="mt-1 text-center text-sm text-ds-warning">{t('lastTrickBonus')}</p>}
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* Team match scores */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            <div>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] ?? 0 })}</div>
            <div>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] ?? 0 })}</div>
            <div className="mt-1">
              {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
            </div>
          </div>

          <details
            open={!isMobile || undefined}
            className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
            data-testid="tn29-bid-history"
          >
            <summary className="cursor-pointer select-none text-ds-text-primary text-xs">
              {t('bidHistory.title')}
            </summary>
            {bidHistory.map(({ player, seat, committed, bid }) => (
              <div key={seat} className="py-0.5">
                {t('bidHistory.player', {
                  seat,
                  player: playerName(player.id, player.isHuman),
                  team: seat % 2 === 0 ? t('team.a') : t('team.b'),
                })}
                : {committed ? (bid === 0 ? t('bid.pass') : bid) : t('bidHistory.pending')}
              </div>
            ))}
          </details>

          {/* Players grouped by team, with the declarer badge */}
          {state.players
            .filter((p) => p.isHuman)
            .map((p) => (
              <section
                key={p.id}
                className="mb-2 px-2 py-1 rounded bg-black/30 flex items-center gap-2 text-ds-text-muted text-sm"
                data-testid="tn29-human-seat"
                data-tutorial="twentynine-human-seat"
                aria-label={t('yourTeam')}
              >
                <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })}
                </span>
                {p.isDeclarer && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('declarerBadge')}</span>
                )}
              </section>
            ))}
          <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
            <div className="flex flex-wrap sm:block">
              {([0, 1] as const).map(
                (team) =>
                  state.players.some((p) => !p.isHuman && p.id % 2 === team) && (
                    <div key={team} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-primary text-xs mb-1">{team === 0 ? t('team.a') : t('team.b')}</div>
                      {renderTeamPlayers(team, true)}
                    </div>
                  ),
              )}
            </div>
          </CpuAccordion>

          {/* Live round card points during bidding/play (matches the CUI's roundPoints
              readout); the round-result block below takes over once the round ends. */}
          {!(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="tn29-round-points">
              <div className="mb-1 text-ds-text-primary">{t('roundPointsTitle')}</div>
              <div>{t('roundResult.teamA', { points: state.roundTeamPoints[0] })}</div>
              <div>{t('roundResult.teamB', { points: state.roundTeamPoints[1] })}</div>
              {/* 契約に届くかは目標点・現在点・場に残る点を突き合わせないと
                  分からない。姉妹ゲームの FortyFives は #4724 で出している (#5644)。 */}
              {contractProgress && (
                <div className="mt-1 text-ds-text-primary" data-testid="tn29-contract-progress">
                  {t('contractProgress', {
                    team: contractProgress.declarerTeam === 0 ? t('team.a') : t('team.b'),
                    got: contractProgress.points,
                    contract: contractProgress.contract,
                  })}
                  {' — '}
                  <span className={contractStatusColor}>
                    {t(`contractStatus.${contractProgress.status}`, {
                      remaining: contractProgress.remaining,
                    })}
                  </span>
                </div>
              )}
            </div>
          )}

          {/* Round result: card points per team */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.teamA', { points: state.roundTeamPoints[0] ?? 0 })}</div>
              <div>{t('roundResult.teamB', { points: state.roundTeamPoints[1] ?? 0 })}</div>
            </div>
          )}
        </div>
      </div>

      {isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}
    </>
  );
}

function FooterHand({ state, game, t, view }: TwentyNineContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;
  const canPlay = state.phase === TwentyNinePhase.PLAY && state.isHumanTurn;
  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="twentynine"
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
    />
  );
}

function FooterStatus({ state, t, frontendHint }: TwentyNineContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="twentynine-hint-live" role="status" aria-live="polite">
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

function FooterActions({ state, game, t, busy }: TwentyNineContext) {
  const loading = busy;
  const isBidPhase = state.phase === TwentyNinePhase.BID;
  const isPlayPhase = state.phase === TwentyNinePhase.PLAY;
  const isTrickEnd = state.phase === TwentyNinePhase.TRICK_END;
  const isRoundEnd = state.phase === TwentyNinePhase.ROUND_END;
  const canPlay = isPlayPhase && state.isHumanTurn;
  const highestBid = Math.max(0, ...state.bids);
  const highestBidder = highestBid > 0 ? state.players[state.bids.indexOf(highestBid)] : undefined;
  const highestBidderName = highestBidder ? playerName(highestBidder.id, highestBidder.isHuman) : '';

  return (
    <>
      {isBidPhase && state.isHumanBidTurn && (
        <>
          <span className="text-xs text-ds-text-muted self-center mr-1">{t('bidPrompt')}</span>
          <span className="text-xs text-ds-text-muted self-center mr-1" data-testid="tn29-highest-bid">
            {highestBid > 0 ? t('bidHighest', { bid: highestBid, player: highestBidderName }) : t('bidNone')}
          </span>
          {BIDS.map((b) => {
            // Pass (0) is always allowed; a non-pass bid must beat the current highest.
            const tooLow = b.value !== 0 && b.value <= highestBid;
            const disabled = loading || tooLow;
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
          disabled={loading || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
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
    </>
  );
}

const twentyNineDefinition: GameDefinition<
  TwentyNineResponse,
  TwentyNineGame,
  Parameters<typeof twentyNineApi.exec>,
  TwentyNineView
> = {
  key: 'twentynine',
  useGame: useTwentyNineGame,
  useView: useTwentyNineView,
  tutorial: TWENTY_NINE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 8 },
  cli: () => ({
    parseCommand: parseTwentyNineCommand,
    formatResponse: formatTwentyNineState,
    helpText: TWENTY_NINE_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === TwentyNinePhase.GAME_END || state.gameEndFlag;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerTeam === view.humanTeam,
    };
  },
  shell: ({ state }) => ({
    isHumanTurn:
      (state.isHumanTurn || state.isHumanBidTurn) && !(state.phase === TwentyNinePhase.GAME_END || state.gameEndFlag),
  }),
  reset: ({ game }) => {
    game.reset();
  },
  settingsPlacement: 'top',
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.twentyNineConfig.cpuDifficulty,
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
          value: game.twentyNineConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) =>
    !view.isMobile && (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ),
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'twentynine-action-buttons',
  resetTutorialId: 'twentynine-reset-button',
  AboveContent,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Twenty-Nine (29) game page: an Indian/Bangladeshi 4-player (2 vs 2) hidden-trump bidding trick-taker. */
export const TwentyNinePage = createGamePage(twentyNineDefinition);
