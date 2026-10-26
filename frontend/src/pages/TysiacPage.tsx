import { useEffect } from 'react';
import type { tysiacApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useTysiacGame } from '../hooks/useTysiacGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { TysiacResponse } from '../types/card';
import { TysiacPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { parseTysiacCommand, TYSIAC_HELP } from '../utils/cli/commands/tysiacCommands';
import { formatTysiacState } from '../utils/cli/formatters/tysiacFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/**
 * Share of the target at which a player counts as "close to winning".
 *
 * Paired with `tysiacNearWinRatio` in `internal/adapter/presenter/TysiacCuiPresenter.go`
 * so the CUI highlights the same rows; `scripts/check-near-win-threshold.mjs` fails the
 * build if the two drift (#6483).
 */
const NEAR_WIN_RATIO = 0.8;

/** Bid increment applied by a raise (mirrors backend `TysiacBidStep`). */
const TYSIAC_BID_STEP = 10;

/** Card design string → suit number (1=♠ 2=♣ 3=♥ 4=♦), to align with trumpSuit. */
const DESIGN_TO_SUIT: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/** Marriage points by suit number (1=♠ 40, 2=♣ 60, 3=♥ 100, 4=♦ 80). */
const MARRIAGE_POINTS: Readonly<Record<number, number>> = { 1: 40, 2: 60, 3: 100, 4: 80 };

/** Tysiąc tutorial step definitions. */
const TYSIAC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tysiac-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tysiac-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tysiac-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="tysiac-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="tysiac-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TYSIAC_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TysiacPhase.BID]: 'bid',
  [TysiacPhase.TALON]: 'talon',
  [TysiacPhase.PLAY]: 'play',
  [TysiacPhase.TRICK_END]: 'trickEnd',
  [TysiacPhase.ROUND_END]: 'roundEnd',
  [TysiacPhase.GAME_END]: 'gameEnd',
};

type TysiacGame = ReturnType<typeof useTysiacGame>;
type TysiacView = ReturnType<typeof useTysiacView>;
type TysiacContext = GamePageContext<TysiacResponse, TysiacGame, TysiacView>;

function useTysiacView({ game }: { game: TysiacGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('tysiac', TYSIAC_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, view }: TysiacContext) {
  const { cardWidth, isMobile } = view;
  const isTrickEnd = state.phase === TysiacPhase.TRICK_END;
  const isRoundEnd = state.phase === TysiacPhase.ROUND_END;
  const isGameEnd = state.phase === TysiacPhase.GAME_END || state.gameEndFlag;
  const trumpSymbol = suitSymbolAt(state.trumpSuit, '-');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
        <span className="mr-4">{t('currentBid', { points: state.currentBid })}</span>
        <span className="mr-4">{t('contract', { points: state.contract })}</span>
        <span>{t('target', { points: state.config.targetPoints })}</span>
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
            dataTutorial="tysiac-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="tysiac-info">
          {/* Per-player match scores with Declarer badge + progress toward the target */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => {
              const target = state.config.targetPoints;
              const progressValue = Math.max(0, Math.min(target, p.score));
              const pct = (progressValue / target) * 100;
              // Warn once a player is within striking distance of the target.
              const isNearWin = p.score / target > NEAR_WIN_RATIO;
              // Declarer forecast: score if this round's contract is met, marked on the bar.
              const forecastPct =
                p.isDeclarer && state.contract > 0
                  ? Math.max(0, Math.min(100, ((p.score + state.contract) / target) * 100))
                  : null;
              const barLabel = t('progressLabel', { score: p.score, target });
              const progressText =
                p.score >= target ? t('targetReached') : t('pointsRemaining', { points: target - p.score });
              return (
                <div key={p.id} className="py-0.5">
                  <div className="flex items-center gap-2">
                    <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                      {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                    </span>
                    {p.isDeclarer && (
                      <span
                        className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}
                        data-testid={`tysiac-declarer-${p.id.toString()}`}
                      >
                        {t('declarerBadge')}
                      </span>
                    )}
                    {p.id === state.dealerIdx && (
                      <span
                        className="px-1.5 py-0.5 rounded text-xs bg-ds-surface-elevated text-ds-text-primary"
                        data-testid={`tysiac-dealer-${p.id.toString()}`}
                      >
                        <span aria-hidden="true">◆ </span>
                        {t('dealerBadge')}
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-ds-text-muted">{progressText}</p>
                  <div
                    role="progressbar"
                    aria-label={barLabel}
                    aria-valuemin={0}
                    aria-valuemax={target}
                    aria-valuenow={progressValue}
                    data-testid={`tysiac-progress-${p.id}`}
                    className="relative mt-0.5 h-1.5 w-full rounded-sm bg-white/15 overflow-hidden"
                  >
                    <div
                      className={`h-full rounded-sm ${isNearWin ? 'bg-ds-warning' : 'bg-ds-accent'}`}
                      style={{ width: `${pct}%` }}
                    />
                    {forecastPct !== null && (
                      <div
                        data-testid={`tysiac-forecast-${p.id}`}
                        aria-hidden="true"
                        className="absolute top-0 h-full border-l border-dashed border-ds-text-primary"
                        style={{ left: `${forecastPct}%` }}
                      />
                    )}
                  </div>
                </div>
              );
            })}
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

          {/* Round result: per-player card points + marriage */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              {state.players.map((p) => (
                <div key={p.id}>
                  <div>
                    {t('roundResult.cardPoints', {
                      name: playerName(p.id, p.isHuman),
                      points: state.roundCardPoints[p.id] ?? 0,
                    })}
                  </div>
                  {(state.roundMarriage[p.id] ?? 0) > 0 && (
                    <div>
                      {t('roundResult.marriage', {
                        name: playerName(p.id, p.isHuman),
                        points: state.roundMarriage[p.id] ?? 0,
                      })}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ state, game, t, view }: TysiacContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isBidPhase = state.phase === TysiacPhase.BID;
  const isTalonPhase = state.phase === TysiacPhase.TALON;
  const isPlayPhase = state.phase === TysiacPhase.PLAY;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const canDiscard = isTalonPhase && state.declarerIdx === humanIdx;
  const canPlay = isPlayPhase && isHumanTurn;

  // Suits where the human holds both K (13) and Q (12) — a marriage that sets
  // trump when led and scores 40/60/80/100 by suit (♠/♣/♦/♥). Surfaced as a
  // banner during play so the bonus is visible while it can still be earned.
  // 結婚の宣言はリードするとき (場に札が無いとき) だけ。
  const canLeadMarriage = isPlayPhase && state.currentTrick.length === 0;
  const marriages = canLeadMarriage
    ? [1, 2, 3, 4]
        .filter((suit) => {
          const cards = humanPlayer?.cards ?? [];
          const hasK = cards.some((c) => DESIGN_TO_SUIT[c.design] === suit && c.value === 13);
          const hasQ = cards.some((c) => DESIGN_TO_SUIT[c.design] === suit && c.value === 12);
          return hasK && hasQ;
        })
        .map((suit) => ({ symbol: suitSymbolAt(suit, '-'), points: MARRIAGE_POINTS[suit] ?? 0 }))
    : [];

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="tysiac-prompt-live" role="status" aria-live="polite">
        {isPlayPhase && (
          <span className="sr-only" data-testid="tysiac-marriage-announcement">
            {marriages.length > 0
              ? t('marriageAvailable', {
                  list: marriages.map((m) => `${m.symbol} K-Q (+${m.points})`).join(t('listSeparator')),
                })
              : t('marriageUnavailable')}
          </span>
        )}
        {isBidPhase && (
          <div className="mb-1 text-center" data-testid="tysiac-bid-prompt">
            <div className="text-sm text-ds-text-primary font-semibold" data-testid="tysiac-current-bid">
              {t('currentBid', { points: state.currentBid })}
            </div>
          </div>
        )}
        {canDiscard && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="tysiac-talon-prompt">
            {t('talonPhase', {
              name: state.players[state.talonRecipientIdx]?.isHuman
                ? t('you')
                : t('cpu', { id: state.players[state.talonRecipientIdx]?.id }),
            })}
          </div>
        )}
      </div>
      {marriages.length > 0 && (
        <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="tysiac-marriage">
          {t('marriageAvailable', {
            list: marriages.map((m) => `${m.symbol} K-Q (+${m.points})`).join(t('listSeparator')),
          })}
        </div>
      )}
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={cardWidth}
          isMobile={isMobile}
          dataTutorialPrefix="tysiac"
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
        />
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: TysiacContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="tysiac-hint-live" role="status" aria-live="polite">
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

function FooterActions({ state, game, t, busy }: TysiacContext) {
  const isHumanTurn = state.isHumanTurn;
  const isBidPhase = state.phase === TysiacPhase.BID;
  const isTalonPhase = state.phase === TysiacPhase.TALON;
  const isPlayPhase = state.phase === TysiacPhase.PLAY;
  const isTrickEnd = state.phase === TysiacPhase.TRICK_END;
  const isRoundEnd = state.phase === TysiacPhase.ROUND_END;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const canBid = isBidPhase && state.currentPlayerIdx === humanIdx;
  const canDiscard = isTalonPhase && state.declarerIdx === humanIdx;
  const canPlay = isPlayPhase && isHumanTurn;

  return (
    <>
      {canBid && (
        <>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => game.handleBid(true)}
            disabled={busy}
            data-testid="tysiac-bid-raise"
          >
            {t('bidRaiseTo', { points: state.currentBid + TYSIAC_BID_STEP })}
          </button>
          <button type="button" className={btnSecondary} onClick={() => game.handleBid(false)} disabled={busy}>
            {t('bidPass')}
          </button>
        </>
      )}
      {canDiscard && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handleDiscard}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('giveCard')}
        </button>
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
    </>
  );
}

const tysiacDefinition: GameDefinition<TysiacResponse, TysiacGame, Parameters<typeof tysiacApi.exec>, TysiacView> = {
  key: 'tysiac',
  useGame: useTysiacGame,
  useView: useTysiacView,
  tutorial: TYSIAC_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 7 },
  cli: () => ({
    parseCommand: parseTysiacCommand,
    formatResponse: formatTysiacState,
    helpText: TYSIAC_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === TysiacPhase.GAME_END || state.gameEndFlag;
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
          value: game.tysiacConfig.cpuDifficulty,
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
          value: game.tysiacConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'tysiac-action-buttons',
  resetTutorialId: 'tysiac-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Tysiąc (Thousand) game page: a Polish 3-player 24-card trump trick-taker with bidding and a talon exchange. */
export const TysiacPage = createGamePage(tysiacDefinition);
