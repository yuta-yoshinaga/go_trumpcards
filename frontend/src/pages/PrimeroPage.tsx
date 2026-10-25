import { useEffect } from 'react';
import type { primeroApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  ANTE_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  usePrimeroGame,
} from '../hooks/usePrimeroGame';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { PrimeroResponse } from '../types/card';
import { PrimeroPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { PRIMERO_HELP, parsePrimeroCommand } from '../utils/cli/commands/primeroCommands';
import { formatPrimeroState } from '../utils/cli/formatters/primeroFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { raiseAvailability, raiseCost } from '../utils/raiseAvailability';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Primero tutorial step definitions. */
const PRIMERO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="primero-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="primero-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="primero-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="primero-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PRIMERO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PrimeroPhase.BETTING]: 'betting',
  [PrimeroPhase.RESULT]: 'result',
};

/** Primero hand categories ordered strongest-first (mirrors the Go domain ranking). */
const PRIMERO_HAND_RANKING = ['fluxus', 'supremus', 'primero', 'numerus'] as const;

type PrimeroGame = ReturnType<typeof usePrimeroGame>;
type PrimeroView = ReturnType<typeof usePrimeroView>;
type PrimeroContext = GamePageContext<PrimeroResponse, PrimeroGame, PrimeroView>;

function usePrimeroView({ game }: { game: PrimeroGame }) {
  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('primero', PRIMERO_PHASE_KEYS);

  return { cardWidth, phaseNames };
}

function playerLabel(t: PrimeroContext['t'], id: number, isHuman: boolean): string {
  return isHuman ? t('you') : t('cpu', { id });
}

function handName(t: PrimeroContext['t'], key?: string): string {
  return key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '';
}

function playerBadge(t: PrimeroContext['t'], p: PrimeroResponse['players'][number]): string {
  return p.out ? t('badge.out') : p.isWinner ? t('badge.winner') : p.folded ? t('badge.folded') : t('badge.active');
}

/** A colour-independent glyph for each player state (WCAG 1.4.1). */
function playerBadgeIcon(p: PrimeroResponse['players'][number]): string {
  return p.out ? '—' : p.isWinner ? '👑' : p.folded ? '×' : '●';
}

function HeaderExtra({ state, t }: PrimeroContext) {
  return <span>{t('chips', { amount: state.chips })}</span>;
}

function Board({ state, t, view }: PrimeroContext) {
  const { cardWidth } = view;
  const isResultPhase = state.phase === PrimeroPhase.RESULT;
  const isGameEnd = state.gameEndFlag;
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  return (
    <>
      <div
        className="text-ds-text-primary text-center mb-2 flex flex-wrap justify-center gap-x-4"
        data-tutorial="primero-info"
      >
        <span className="whitespace-nowrap">
          {t('round', { n: state.roundNumber, total: state.config.targetRounds })}
        </span>
        <span className="whitespace-nowrap">{t('pot', { amount: state.pot })}</span>
        <span className="whitespace-nowrap">{t('ante', { amount: state.ante })}</span>
        <span className="whitespace-nowrap">{t('currentBet', { amount: state.currentBet })}</span>
      </div>

      {/* Players */}
      <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="primero-players">
        <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
        <ul className="list-none">
          {state.players.map((p) => {
            const isCurrentTurn = !isGameEnd && state.currentPlayerIdx === p.id;
            return (
              <li
                key={p.id}
                data-testid={`primero-player-${p.id}`}
                aria-label={t('playerRowLabel', { name: playerLabel(t, p.id, p.isHuman), status: playerBadge(t, p) })}
                className={`text-sm py-0.5 ${p.isWinner ? 'text-ds-success' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''} ${
                  isCurrentTurn ? 'border-l-2 border-ds-accent pl-1' : ''
                }`}
              >
                {playerLabel(t, p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
                {t('roundBet', { amount: p.roundBet })} · <span aria-hidden="true">{playerBadgeIcon(p)} </span>[
                <span className="whitespace-nowrap">{playerBadge(t, p)}</span>]
                {p.handName && <span className="whitespace-nowrap"> · {handName(t, p.handName)}</span>}
                {isResultPhase && (
                  <>
                    {' · '}
                    {t('roundResult.payout', { amount: p.roundPayout })}
                    {' · '}
                    {t('roundResult.net', { amount: formatSignedDelta(p.netChange) })}
                  </>
                )}
              </li>
            );
          })}
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
                    <CardImage key={`${p.id}-${i}`} card={c} width={cardWidth} />
                  ))}
                </div>
              </div>
            ))}
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
        </div>
      )}

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ state, t, view }: PrimeroContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  return humanPlayer && humanPlayer.cards.length > 0 ? (
    <div className="mb-2" data-tutorial="primero-hand">
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
    <div className="text-ds-text-muted text-sm mb-2" data-tutorial="primero-hand">
      {t('handLabel')}
    </div>
  );
}

function FooterLead({ state, t }: PrimeroContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  return (
    /* Hand-ranking legend (static reference), highlighting the current hand */
    <details className="mb-2 p-2 rounded bg-black/30" data-testid="primero-hand-legend">
      <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('handLegend.title')}</summary>
      <div className="mt-1 text-ds-text-muted text-xs">
        <table className="w-full">
          <thead>
            <tr>
              <th scope="col" className="text-left font-normal w-6">
                {t('handLegend.rankCol')}
              </th>
              <th scope="col" className="text-left font-normal">
                {t('handLegend.handCol')}
              </th>
              <th scope="col" className="text-left font-normal">
                {t('handLegend.descCol')}
              </th>
            </tr>
          </thead>
          <tbody>
            {PRIMERO_HAND_RANKING.map((key, i) => {
              const isCurrent = humanPlayer?.handName?.toLowerCase() === key;
              return (
                <tr
                  key={key}
                  data-testid={`primero-hand-legend-row-${key}`}
                  aria-current={isCurrent ? 'true' : undefined}
                  className={isCurrent ? 'text-ds-accent font-semibold' : ''}
                >
                  <td className="align-top">{i + 1}</td>
                  <td className="align-top pr-2 whitespace-nowrap">
                    {t(`hand.${key}`)}
                    {isCurrent ? <span className="ml-1">{`← ${t('handLegend.current')}`}</span> : null}
                  </td>
                  <td className="align-top">{t(`handLegend.${key}`)}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
        <div className="mt-1">{t('handLegend.pointsNote')}</div>
      </div>
    </details>
  );
}

function FooterStatus({ frontendHint, t }: PrimeroContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

function FooterActions({ game, state, t, busy }: PrimeroContext) {
  const isBettingPhase = state.phase === PrimeroPhase.BETTING;
  const isResultPhase = state.phase === PrimeroPhase.RESULT;
  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const humanPlayer = state.players.find((p) => p.isHuman);

  // **レイズが消えた理由は 2 つある。**上限と枚数不足を区別しないと、
  // 「レイズ 1/3回」がボタンの無い画面で「まだできる」と読めてしまう。
  const raiseInput = {
    raiseCount: state.raiseCount,
    maxRaises: state.maxRaises,
    chips: humanPlayer?.chips ?? 0,
    currentBet: state.currentBet,
    roundBet: humanPlayer?.roundBet ?? 0,
    ante: state.ante,
  };
  const raiseBlock = raiseAvailability(raiseInput);
  const raiseNeeded = raiseCost(raiseInput);

  return (
    <>
      {isBettingPhase && isHumanTurn && !isGameEnd && humanPlayer && (
        <>
          <div className="w-full text-ds-text-primary text-sm" data-testid="primero-betting-controls">
            <span className="mr-4">{t('yourRoundBet', { amount: humanPlayer.roundBet })}</span>
            <span>{t('callAmount', { amount: Math.max(0, state.currentBet - humanPlayer.roundBet) })}</span>
            {state.canRaise && <span className="ml-4">{t('raiseAmount', { amount: raiseNeeded })}</span>}
          </div>
          <button type="button" className={btnPrimary} onClick={game.handleCall} disabled={busy}>
            {t('callButton')}
          </button>
          {/* **レイズが消えた理由を書く。**回数上限とチップ不足を
              区別できないと、突然選択肢を奪われたように見える (#4925)。 */}
          <span
            className="text-ds-text-muted text-xs"
            data-testid="primero-raise-count"
            role="status"
            aria-live="polite"
          >
            {raiseBlock === 'cap'
              ? t('raiseCapReached', { max: state.maxRaises })
              : raiseBlock === 'chips'
                ? t('raiseNoChips', { cost: raiseNeeded })
                : t('raiseCount', { count: state.raiseCount, max: state.maxRaises })}
          </span>
          {state.canRaise && (
            <button type="button" className={btnSuccess} onClick={game.handleRaise} disabled={busy}>
              {t('raiseButton')}
            </button>
          )}
          <button type="button" className={btnDanger} onClick={game.handleFold} disabled={busy}>
            {t('foldButton')}
          </button>
        </>
      )}

      {isResultPhase && !isGameEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const primeroDefinition: GameDefinition<
  PrimeroResponse,
  PrimeroGame,
  Parameters<typeof primeroApi.exec>,
  PrimeroView
> = {
  key: 'primero',
  useGame: usePrimeroGame,
  useView: usePrimeroView,
  tutorial: PRIMERO_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 4 },
  cli: () => ({
    parseCommand: parsePrimeroCommand,
    formatResponse: formatPrimeroState,
    helpText: PRIMERO_HELP,
  }),
  phase: ({ state, view }) => {
    const isResultPhase = state.phase === PrimeroPhase.RESULT;
    const humanWonMatch = state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false);
    return {
      name: view.phaseNames[state.phase],
      ended: state.gameEndFlag,
      won: isResultPhase && (humanWonMatch || state.result > 0),
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.phase === PrimeroPhase.BETTING && state.isHumanTurn && !state.gameEndFlag,
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
          value: game.primeroConfig.playerCount,
          options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('playerCount', v),
        },
        {
          type: 'select',
          id: 'ante',
          label: t('settings.ante'),
          value: game.primeroConfig.ante,
          options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('ante', v),
        },
        {
          type: 'select',
          id: 'startingChips',
          label: t('settings.startingChips'),
          value: game.primeroConfig.startingChips,
          options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('startingChips', v),
        },
        {
          type: 'select',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.primeroConfig.targetRounds,
          options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetRounds', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  resetTutorialId: 'primero-reset-button',
  actionsTutorialId: 'primero-action-buttons',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  HeaderExtra,
  Board,
  FooterHand,
  FooterLead,
  FooterStatus,
  FooterActions,
};

/** Renders the Primero game page: a Renaissance 4-card pot-vying game. */
export const PrimeroPage = createGamePage(primeroDefinition);
