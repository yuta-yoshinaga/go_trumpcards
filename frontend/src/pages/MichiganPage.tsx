import type { TFunction } from 'i18next';
import { useEffect, useState } from 'react';
import type { michiganApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { useCardDimensions } from '../hooks/useCardDimensions';
import {
  ANTE_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useMichiganGame,
} from '../hooks/useMichiganGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { MichiganResponse } from '../types/card';
import { MichiganPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MICHIGAN_HELP, parseMichiganCommand } from '../utils/cli/commands/michiganCommands';
import { formatMichiganState } from '../utils/cli/formatters/michiganFormatter';
import { michiganBoodleGuides } from '../utils/michiganBoodleGuide';
import { michiganNextPlayable } from '../utils/michiganPlayable';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Number of center boodle cards (always 4: A♥, K♣, Q♦, J♠). */
const MICHIGAN_BOODLE_COUNT = 4;

/** Distributes `budget` chips as evenly as possible across `n` boodles (remainder to the first slots). */
function evenSplit(budget: number, n: number): number[] {
  const base = Math.floor(budget / n);
  const rem = budget - base * n;
  return Array.from({ length: n }, (_, i) => base + (i < rem ? 1 : 0));
}

/** Michigan tutorial step definitions. */
const MICHIGAN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="michigan-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="michigan-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="michigan-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="michigan-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MICHIGAN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MichiganPhase.BET]: 'bet',
  [MichiganPhase.PLAY]: 'play',
  [MichiganPhase.RESULT]: 'result',
};

type MichiganGame = ReturnType<typeof useMichiganGame>;
type MichiganView = ReturnType<typeof useMichiganView>;
type MichiganContext = GamePageContext<MichiganResponse, MichiganGame, MichiganView>;

function useMichiganView({ game }: { game: MichiganGame }) {
  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const state = game.state;

  // Local boodle-bet distribution; re-seeded (even split) whenever a new bet turn begins.
  const [bets, setBets] = useState<number[]>(() => evenSplit(state?.betBudget ?? 0, MICHIGAN_BOODLE_COUNT));
  // biome-ignore lint/correctness/useExhaustiveDependencies: re-seed only when the budget / round / placed flag changes.
  useEffect(() => {
    setBets(evenSplit(state?.betBudget ?? 0, MICHIGAN_BOODLE_COUNT));
  }, [state?.betBudget, state?.roundNumber, state?.humanBetPlaced]);

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('michigan', MICHIGAN_PHASE_KEYS);

  const adjustBet = (index: number, delta: number) => {
    setBets((prev) => {
      const next = [...prev];
      const value = (next[index] ?? 0) + delta;
      if (value < 0) return prev;
      // Derive the remaining budget from prev (not the render-closure value) so
      // rapid successive clicks can't overspend via a stale betRemaining.
      const remaining = (state?.betBudget ?? 0) - prev.reduce((a, b) => a + b, 0);
      if (delta > 0 && remaining <= 0) return prev;
      next[index] = value;
      return next;
    });
  };

  const betSum = bets.reduce((a, b) => a + b, 0);
  const betRemaining = (state?.betBudget ?? 0) - betSum;
  const canPlaceBets = betRemaining === 0;

  return {
    bets,
    adjustBet,
    betRemaining,
    canPlaceBets,
    cardWidth,
    phaseNames,
  };
}

function playerLabel(t: TFunction, id: number, isHuman: boolean): string {
  return isHuman ? t('you') : t('cpu', { id });
}

function playerBadge(t: TFunction, p: MichiganResponse['players'][number]): string {
  return p.isWinner ? t('badge.wentOut') : p.isCurrent ? t('badge.toPlay') : t('badge.waiting');
}

function HeaderExtra({ state, t }: MichiganContext) {
  return <span>{t('chips', { amount: state.chips })}</span>;
}

function Board({ game, state, t, view, busy }: MichiganContext) {
  const isPlayPhase = state.phase === MichiganPhase.PLAY;
  const isResultPhase = state.phase === MichiganPhase.RESULT;
  const isGameEnd = state.gameEndFlag;

  const sequenceAnnouncement =
    state.needNewSequence || state.seqSuit === 0
      ? t('sequence.new')
      : t('sequence.active', { suit: state.seqSuitName, value: state.seqHighValue });

  const roundResultAnnouncement = isResultPhase
    ? [
        ...(state.winnerIdx >= 0
          ? [
              t('roundResult.winner', {
                name: playerLabel(
                  t,
                  state.winnerIdx,
                  (state.players[state.winnerIdx] as MichiganResponse['players'][number]).isHuman,
                ),
              }),
            ]
          : []),
        state.roundBoodleWins.length === 0
          ? t('roundResult.noBoodleWins')
          : state.roundBoodleWins
              .map((win) =>
                t('roundResult.boodleWin', {
                  card: cardAlt(win.card),
                  name: playerLabel(
                    t,
                    win.seat,
                    (state.players[win.seat] as MichiganResponse['players'][number]).isHuman,
                  ),
                  amount: win.amount,
                }),
              )
              .join(t('listSeparator')),
      ].join(t('listSeparator'))
    : '';

  return (
    <>
      {isResultPhase && !isGameEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="michigan-info">
        <span className="mr-4">{t('round', { n: state.roundNumber, total: state.config.targetRounds })}</span>
        <span className="mr-4">{t('ante', { amount: state.ante })}</span>
        <span>{t('deadHand', { count: state.deadHandCount })}</span>
        {/* 数字だけでは謎の数でしかない。用語もその意味もどこにも
            説明が無かった (#5700)。 */}
        <details className="mt-1 inline-block text-left" data-testid="michigan-deadhand-help">
          <summary className="cursor-pointer select-none text-ds-text-muted text-xs">{t('deadHandWhatTitle')}</summary>
          <p className="mt-1 max-w-xs text-ds-text-muted text-xs">{t('deadHandWhat')}</p>
        </details>
      </div>

      {/* Boodles (center betting cards) */}
      <div className="mb-2 p-2 rounded bg-black/30">
        <div className="mb-1 text-ds-text-primary text-sm">{t('boodlesTitle')}</div>
        <div className="flex flex-wrap justify-center gap-3">
          {state.boodles.map((b, i) => (
            <div key={`boodle-${i}`} className="flex flex-col items-center">
              <CardImage card={b.card} width={view.cardWidth} />
              <div className="text-ds-text-muted text-xs mt-0.5">{t('boodleChips', { amount: b.chips })}</div>
              <div className={`text-xs ${b.claimedBy >= 0 ? 'text-ds-success' : 'text-ds-text-muted'}`}>
                {b.claimedBy >= 0
                  ? t('boodleClaimed', {
                      name: playerLabel(t, b.claimedBy, state.players[b.claimedBy]?.isHuman ?? false),
                    })
                  : t('boodleOpen')}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Current sequence */}
      <div data-testid="michigan-sequence-live" className="sr-only" role="status" aria-live="polite" aria-atomic="true">
        {isPlayPhase ? sequenceAnnouncement : ''}
      </div>
      <div
        data-testid="michigan-round-result-live"
        className="sr-only"
        role="status"
        aria-live="polite"
        aria-atomic="true"
      >
        {roundResultAnnouncement}
      </div>
      <div className="mb-2 text-center text-ds-text-muted text-sm" aria-hidden="true">
        {sequenceAnnouncement}
      </div>

      {/* Players */}
      <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="michigan-players">
        <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
        {state.players.map((p) => (
          <div
            key={p.id}
            className={`text-sm py-0.5 ${p.isWinner ? 'text-ds-success' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
          >
            {playerLabel(t, p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
            {t('handCount', { count: p.cardCount })} · {t('roundBet', { amount: p.roundBet })} · [{playerBadge(t, p)}]
          </div>
        ))}
      </div>

      {/* Revealed hands at result */}
      {isResultPhase && (
        <div className="mb-2 p-2 rounded bg-black/30">
          {state.players
            .filter((p) => !p.isHuman && p.cards.length > 0)
            .map((p) => (
              <div key={p.id} className="mb-1">
                <div className="text-ds-text-muted text-xs mb-0.5">{playerLabel(t, p.id, p.isHuman)}</div>
                <div className="flex flex-wrap gap-1">
                  {p.cards.map((c, i) => (
                    <CardImage key={`${p.id}-${i}`} card={c} width={view.cardWidth} />
                  ))}
                </div>
              </div>
            ))}
        </div>
      )}

      {/* Round result */}
      {isResultPhase && (
        <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
          <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
          {state.winnerIdx >= 0 && (
            <div>
              {t('roundResult.winner', {
                name: playerLabel(t, state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false),
              })}
            </div>
          )}
          <div className="mt-1">
            {state.roundBoodleWins.length === 0
              ? t('roundResult.noBoodleWins')
              : state.roundBoodleWins.map((win, i) => (
                  <div key={`${win.card.design}-${win.card.value}-${i}`}>
                    {t('roundResult.boodleWin', {
                      card: cardAlt(win.card),
                      name: playerLabel(t, win.seat, state.players[win.seat]?.isHuman ?? false),
                      amount: win.amount,
                    })}
                  </div>
                ))}
          </div>
        </div>
      )}

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterStatus({ frontendHint, t }: MichiganContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

function FooterHand({ game, state, t, view, busy }: MichiganContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBetPhase = state.phase === MichiganPhase.BET;
  const isPlayPhase = state.phase === MichiganPhase.PLAY;
  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;

  const showBetControls = isBetPhase && isHumanTurn && !state.humanBetPlaced && !isGameEnd;
  const canPlay = (i: number) => isPlayPhase && isHumanTurn && state.playableIndices.includes(i);
  const nextPlayable = michiganNextPlayable(state);
  const showPlayHints = isPlayPhase && isHumanTurn && !isGameEnd;

  return (
    <>
      {showBetControls && (
        <div className="text-ds-text-muted text-sm mb-2">{t('betPrompt', { budget: state.betBudget })}</div>
      )}
      {humanPlayer && humanPlayer.cards.length > 0 ? (
        <div className="mb-2" data-tutorial="michigan-hand">
          <div className="text-ds-text-muted text-xs mb-0.5">{t('handLabel')}</div>
          {showPlayHints && (
            <div className="text-ds-success text-sm mb-1" data-testid="michigan-next-hint">
              {nextPlayable.isNewSequence
                ? t('nextHint.new')
                : t('nextHint.next', { suit: nextPlayable.suitName, value: nextPlayable.nextValue })}
            </div>
          )}
          <div className="flex flex-wrap gap-1">
            {humanPlayer.cards.map((c, i) => (
              <button
                key={`human-${i}`}
                type="button"
                onClick={() => canPlay(i) && game.handlePlay(i)}
                disabled={!canPlay(i) || busy}
                className={`rounded transition-all ${
                  canPlay(i)
                    ? 'cursor-pointer hover:-translate-y-2 ring-2 ring-ds-success'
                    : 'cursor-default opacity-60'
                }`}
                data-testid={`hand-card-${i}`}
                data-playable={canPlay(i) ? 'true' : 'false'}
                aria-label={t('playCardAria', { card: cardAlt(c) })}
              >
                <CardImage card={c} width={view.cardWidth} />
                {canPlay(i) && (
                  <span
                    className="block text-center text-ds-success text-xs mt-0.5"
                    data-testid={`playable-badge-${i}`}
                  >
                    {nextPlayable.isNewSequence ? t('nextHint.leadBadge') : t('nextHint.nextBadge')}
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>
      ) : (
        <div className="text-ds-text-muted text-sm mb-2" data-tutorial="michigan-hand">
          {t('handLabel')}
        </div>
      )}

      <div className="flex flex-col gap-2" data-tutorial="michigan-action-buttons">
        {isPlayPhase && isHumanTurn && !isGameEnd && (
          <div className="text-ds-text-muted text-sm">{t('playPrompt')}</div>
        )}
      </div>
    </>
  );
}

function primaryActions({ game, state, t, view, busy }: MichiganContext) {
  const isBetPhase = state.phase === MichiganPhase.BET;
  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const showBetControls = isBetPhase && isHumanTurn && !state.humanBetPlaced && !isGameEnd;

  if (!showBetControls) return null;

  const humanPlayer = state.players.find((p) => p.isHuman);
  // Per-boodle betting guidance: which boodles the human can recover (holds the
  // matching card) and which are already claimed, so bets can be biased sensibly.
  const boodleGuides = michiganBoodleGuides(state.boodles, humanPlayer?.cards ?? []);

  return (
    <div className="flex flex-col gap-2" data-tutorial="michigan-action-buttons">
      <div className="flex items-center justify-between gap-2 text-ds-text-primary text-sm">
        <span>{t('betRemaining', { amount: view.betRemaining })}</span>
        <button
          type="button"
          className={btnPrimary}
          onClick={() => game.handleBet(view.bets)}
          disabled={busy || !view.canPlaceBets}
        >
          {t('placeBets')}
        </button>
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-1 items-start" data-testid="michigan-bet-grid">
        {state.boodles.map((b, i) => (
          <div
            key={`bet-${i}`}
            className="flex flex-wrap items-center justify-center gap-1"
            data-testid={`bet-boodle-${i}`}
          >
            <div className="flex items-center gap-1">
              <CardImage card={b.card} width={24} />
              <span className="text-ds-text-primary text-xs">{cardAlt(b.card)}</span>
            </div>
            <div className="flex flex-col items-center gap-0.5">
              {boodleGuides[i]?.collectible && (
                <span className="text-ds-success text-xs" data-testid={`bet-collectible-${i}`}>
                  {t('betCollectible')}
                </span>
              )}
              {boodleGuides[i]?.claimed && (
                <span className="text-ds-warning text-xs" data-testid={`bet-claimed-warning-${i}`}>
                  {t('betClaimedWarning')}
                </span>
              )}
            </div>
            <div className="flex items-center gap-1">
              <button
                type="button"
                className={`${btnDanger} h-11`}
                onClick={() => view.adjustBet(i, -1)}
                disabled={busy || (view.bets[i] ?? 0) <= 0}
                aria-label={t('betMinusAria', { card: cardAlt(b.card), amount: view.bets[i] })}
              >
                −
              </button>
              <span className="w-6 text-center text-ds-text-primary">{view.bets[i] ?? 0}</span>
              <button
                type="button"
                className={`${btnSuccess} h-11`}
                onClick={() => view.adjustBet(i, 1)}
                disabled={busy || view.betRemaining <= 0}
                aria-label={t('betPlusAria', { card: cardAlt(b.card), amount: view.bets[i] })}
              >
                +
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

const michiganDefinition: GameDefinition<
  MichiganResponse,
  MichiganGame,
  Parameters<typeof michiganApi.exec>,
  MichiganView
> = {
  key: 'michigan',
  useGame: useMichiganGame,
  useView: useMichiganView,
  tutorial: MICHIGAN_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseMichiganCommand,
    formatResponse: formatMichiganState,
    helpText: MICHIGAN_HELP,
  }),
  phase: ({ state, view }) => {
    const humanWonMatch = state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false);
    return {
      name: view.phaseNames[state.phase],
      ended: state.gameEndFlag,
      won: state.phase === MichiganPhase.RESULT && (humanWonMatch || state.result > 0),
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn && state.phase !== MichiganPhase.RESULT && !state.gameEndFlag,
  }),
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'playerCount',
          label: t('settings.playerCount'),
          value: game.michiganConfig.playerCount,
          options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('playerCount', v),
        },
        {
          type: 'select',
          id: 'ante',
          label: t('settings.ante'),
          value: game.michiganConfig.ante,
          options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('ante', v),
        },
        {
          type: 'select',
          id: 'startingChips',
          label: t('settings.startingChips'),
          value: game.michiganConfig.startingChips,
          options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('startingChips', v),
        },
        {
          type: 'select',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.michiganConfig.targetRounds,
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
  resetTutorialId: 'michigan-reset-button',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  HeaderExtra,
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Michigan (Newmarket) game page: a "stops" chip-betting family game. */
export const MichiganPage = createGamePage(michiganDefinition);
