import type { TFunction } from 'i18next';
import { useEffect, useMemo } from 'react';
import type { gutsApi } from '../api/gameApi';
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
  useGutsGame,
} from '../hooks/useGutsGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { GutsResponse } from '../types/card';
import { GutsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { GUTS_HELP, parseGutsCommand } from '../utils/cli/commands/gutsCommands';
import { formatGutsState } from '../utils/cli/formatters/gutsFormatter';
import * as gutsGuideUtils from '../utils/gutsGuideUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Guts tutorial step definitions. */
const GUTS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="guts-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guts-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guts-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guts-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GUTS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GutsPhase.DECLARE]: 'declare',
  [GutsPhase.RESULT]: 'result',
};

type GutsGame = ReturnType<typeof useGutsGame>;
type GutsView = ReturnType<typeof useGutsView>;
type GutsContext = GamePageContext<GutsResponse, GutsGame, GutsView>;

function useGutsView({ game }: { game: GutsGame }) {
  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('guts', GUTS_PHASE_KEYS);

  const state = game.state;
  const humanPlayer = state?.players.find((p) => p.isHuman);
  const humanIdx = state?.players.findIndex((p) => p.isHuman) ?? -1;
  const humanCards = humanPlayer?.cards;
  const humanChips = humanPlayer?.chips;
  const pot = state?.pot;
  const isDeclarePhase = state?.phase === GutsPhase.DECLARE;
  const isGameEnd = state?.gameEndFlag ?? false;
  const cpuSeatOutKey = JSON.stringify(state?.players.filter((p) => !p.isHuman).map((p) => [p.id, p.out]) ?? []);

  // Rough hand-name + win-chance readout shown while the human must call In/Out.
  const declareGuide = useMemo(() => {
    if (
      !isDeclarePhase ||
      isGameEnd ||
      !humanCards ||
      humanCards.length === 0 ||
      pot === undefined ||
      humanChips === undefined
    )
      return null;
    const cpuPlayers = JSON.parse(cpuSeatOutKey) as [number, boolean][];
    const opponents = cpuPlayers.map(([seat, out]) => ({ seat, out }));
    return {
      ...gutsGuideUtils.evaluateGutsGuide(humanCards, opponents, humanIdx),
      maxPayment: Math.min(pot, humanChips),
    };
  }, [isDeclarePhase, isGameEnd, humanCards, cpuSeatOutKey, humanIdx, pot, humanChips]);

  return { cardWidth, phaseNames, declareGuide };
}

function playerLabel(t: TFunction, id: number, isHuman: boolean): string {
  return isHuman ? t('you') : t('cpu', { id });
}

function handName(t: TFunction, key?: string): string {
  return key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '';
}

function playerBadge(t: TFunction, p: GutsResponse['players'][number]): string {
  return p.out
    ? t('badge.out')
    : p.isWinner
      ? t('badge.winner')
      : p.isMatcher
        ? t('badge.matcher')
        : p.in
          ? t('badge.in')
          : t('badge.waiting');
}

function HeaderExtra({ state, t }: GutsContext) {
  return <span>{t('chips', { amount: state.chips })}</span>;
}

function Board({ state, t, view }: GutsContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isResultPhase = state.phase === GutsPhase.RESULT;
  const declareGuide = view.declareGuide;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="guts-info">
        <span className="mr-4">{t('round', { n: state.roundNumber, total: state.config.targetRounds })}</span>
        <span className="mr-4">{t('pot', { amount: state.pot })}</span>
        <span>{t('ante', { amount: state.ante })}</span>
      </div>

      {declareGuide && (
        <div
          className="mb-2 p-2 rounded bg-ds-surface text-sm"
          data-testid="guts-declare-guide"
          role="status"
          aria-live="polite"
        >
          <div className="text-ds-text-primary font-semibold">
            {t('guide.handLabel')}: <span className="text-base">{handName(t, declareGuide.handKey)}</span>
          </div>
          <div className="text-ds-text-muted">
            {t('guide.tierLabel')}: <span data-testid="guts-guide-tier">{t(`guide.tier.${declareGuide.tier}`)}</span>
          </div>
          <div className="text-ds-text-primary" data-testid="guts-guide-probability">
            {t('guide.probability', { percent: declareGuide.winChance })}
          </div>
          <div className="text-ds-text-muted text-xs">{t('guide.assumptions')}</div>
          <div className="text-ds-error-text text-xs" data-testid="guts-guide-risk">
            {t('guide.matchRisk', { amount: declareGuide.maxPayment })}
          </div>
        </div>
      )}

      {/* Players */}
      <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="guts-players">
        <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
        <ul className="list-none p-0 m-0" data-testid="guts-player-list">
          {state.players.map((p) => (
            <li
              key={p.id}
              className={`text-sm py-0.5 ${p.isWinner ? 'text-ds-success' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
              aria-label={t(p.handName ? 'playerSummaryWithHand' : 'playerSummary', {
                name: playerLabel(t, p.id, p.isHuman),
                chips: p.chips,
                bet: p.roundBet,
                status: playerBadge(t, p),
                hand: p.handName ? handName(t, p.handName) : '',
              })}
            >
              {playerLabel(t, p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
              {t('roundBet', { amount: p.roundBet })} · [{playerBadge(t, p)}]
              {p.handName ? ` · ${handName(t, p.handName)}` : ''}
            </li>
          ))}
        </ul>
      </div>

      {/* Revealed hands at result */}
      {isResultPhase && (
        <div className="mb-2 p-2 rounded bg-black/30">
          {state.players
            .filter((p) => !p.isHuman && p.cards.length > 0)
            .map((p) => (
              <div key={p.id} className="mb-1">
                <div className="text-ds-text-muted text-xs mb-0.5">
                  {playerLabel(t, p.id, p.isHuman)}
                  {p.handName ? ` — ${handName(t, p.handName)}` : ''}
                </div>
                <div className="flex gap-1">
                  {p.cards.map((c, i) => (
                    <CardImage key={`${p.id}-${i}`} card={c} width={view.cardWidth} />
                  ))}
                </div>
              </div>
            ))}
        </div>
      )}

      {/* Nobody stayed: the pot carries. The CUI has said so since the game
          shipped (guts.result.carry); the Web page rendered no panel at all
          for winnerIdx < 0, which is exactly when carryPot matters (#4847). */}
      {isResultPhase && state.winnerIdx < 0 && (
        <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="guts-carry-result">
          <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
          <div className="text-ds-warning font-semibold">
            {t('roundResult.carry', { pot: state.carryPot, count: state.carryCount })}
          </div>
        </div>
      )}

      {/* Round result */}
      {isResultPhase && state.winnerIdx >= 0 && (
        <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
          <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
          <div>
            {t('roundResult.winner', {
              name: playerLabel(t, state.winnerIdx, state.winnerIdx === humanIdx),
              pot: state.pot,
            })}
          </div>
          {state.matchers.map((idx) => {
            const matcher = state.players[idx];
            if (!matcher) return null;
            const amount = Math.max(0, matcher.roundBet - state.ante);
            return (
              <div
                key={`matcher-${idx}`}
                className="text-ds-error-text font-semibold"
                data-testid="guts-matcher-payment"
              >
                {t('roundResult.matchPayment', {
                  name: playerLabel(t, idx, idx === humanIdx),
                  amount,
                })}
              </div>
            );
          })}
        </div>
      )}

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ state, t, view }: GutsContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);

  return humanPlayer && humanPlayer.cards.length > 0 ? (
    <div className="mb-2" data-tutorial="guts-hand">
      <div className="text-ds-text-muted text-xs mb-0.5">
        {t('handLabel')}
        {humanPlayer.handName ? ` — ${handName(t, humanPlayer.handName)}` : ''}
      </div>
      <div className="flex gap-1">
        {humanPlayer.cards.map((c, i) => (
          <CardImage key={`human-${i}`} card={c} width={view.cardWidth} />
        ))}
      </div>
    </div>
  ) : (
    <div className="text-ds-text-muted text-sm mb-2" data-tutorial="guts-hand">
      {t('handLabel')}
    </div>
  );
}

function FooterStatus({ frontendHint, t }: GutsContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

function primaryActions({ game, state, t, busy }: GutsContext) {
  const isDeclarePhase = state.phase === GutsPhase.DECLARE;
  const isResultPhase = state.phase === GutsPhase.RESULT;
  const isGameEnd = state.gameEndFlag ?? false;
  if ((!isDeclarePhase && !isResultPhase) || isGameEnd) return null;

  const humanPlayer = state.players.find((p) => p.isHuman);

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="guts-action-buttons">
      {isDeclarePhase && (
        <>
          {humanPlayer && (
            <div
              className="w-full rounded border border-ds-border bg-ds-surface px-3 py-2 text-sm text-ds-text-primary sm:w-auto"
              data-testid="guts-declare-money"
            >
              <span className="mr-3">{t('declareMoney.chips', { amount: humanPlayer.chips })}</span>
              <span className="mr-3">{t('declareMoney.bet', { amount: humanPlayer.roundBet })}</span>
              <span>{t('pot', { amount: state.pot })}</span>
            </div>
          )}
          <button type="button" className={btnPrimary} onClick={game.handleIn} disabled={busy}>
            {t('inButton')}
          </button>
          <button type="button" className={btnDanger} onClick={game.handleOut} disabled={busy}>
            {t('outButton')}
          </button>
        </>
      )}
      {isResultPhase && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const gutsDefinition: GameDefinition<GutsResponse, GutsGame, Parameters<typeof gutsApi.exec>, GutsView> = {
  key: 'guts',
  useGame: useGutsGame,
  useView: useGutsView,
  tutorial: GUTS_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 2 },
  cli: () => ({
    parseCommand: parseGutsCommand,
    formatResponse: formatGutsState,
    helpText: GUTS_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.gameEndFlag ?? false,
    won:
      state.phase === GutsPhase.RESULT &&
      ((state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false)) || state.result > 0),
  }),
  shell: ({ state }) => ({
    isHumanTurn: state.phase === GutsPhase.DECLARE && !state.gameEndFlag,
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
          value: game.gutsConfig.playerCount,
          options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('playerCount', v),
        },
        {
          type: 'select',
          id: 'ante',
          label: t('settings.ante'),
          value: game.gutsConfig.ante,
          options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('ante', v),
        },
        {
          type: 'select',
          id: 'startingChips',
          label: t('settings.startingChips'),
          value: game.gutsConfig.startingChips,
          options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('startingChips', v),
        },
        {
          type: 'select',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.gutsConfig.targetRounds,
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
  resetTutorialId: 'guts-reset-button',
  actionsClassName: 'flex gap-2 items-center justify-center',
  HeaderExtra,
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Guts game page: a fast multi-player pot-vying gambling game. */
export const GutsPage = createGamePage(gutsDefinition);
