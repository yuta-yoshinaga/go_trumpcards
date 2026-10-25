import { useEffect } from 'react';
import type { sutdaApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { useCardDimensions } from '../hooks/useCardDimensions';
import {
  SUTDA_CHIP_OPTIONS,
  SUTDA_CPU_DIFFICULTY_OPTIONS,
  SUTDA_SEAT_OPTIONS,
  useSutdaGame,
} from '../hooks/useSutdaGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { SutdaResponse } from '../types/card';
import { SutdaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import {
  parseSutdaCommand,
  SUTDA_HAND_RANKING_KEYS,
  SUTDA_HELP,
  sutdaLocalCommand,
} from '../utils/cli/commands/sutdaCommands';
import { formatSutdaState } from '../utils/cli/formatters/sutdaFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Sutda tutorial step definitions. */
const SUTDA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sutda-table"]',
    messageKey: 'tutorial.table',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sutda-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sutda-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sutda-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type SutdaGame = ReturnType<typeof useSutdaGame>;
type SutdaView = ReturnType<typeof useSutdaView>;
type SutdaContext = GamePageContext<SutdaResponse, SutdaGame, SutdaView>;

function useSutdaView({ game, t }: { game: SutdaGame; t: SutdaContext['t'] }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();

  // **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしているので
  // ここは直接引く。
  const phaseNames: Readonly<Record<string, string>> = {
    [SutdaPhase.BET]: t('phase.bet'),
    [SutdaPhase.SHOWDOWN]: t('phase.showdown'),
    [SutdaPhase.GAME_END]: t('phase.gameEnd'),
  };

  return { cardWidth, isMobile, phaseNames };
}

function AboveContent({ game, state, t }: SutdaContext) {
  const { sutdaConfig } = game;
  return (
    <>
      {(sutdaConfig.cpuDifficulty !== state.config.cpuDifficulty ||
        sutdaConfig.seats !== state.config.seats ||
        sutdaConfig.startChips !== state.config.startChips) && (
        <p className="px-4 pt-1 text-sm text-ds-warning">{t('settings.pendingReset')}</p>
      )}

      <details className="px-4 pt-2" data-testid="sutda-hand-ranking">
        <summary className="text-ds-text-primary text-sm cursor-pointer select-none hover:text-ds-accent">
          {t('handRanking.title')}
        </summary>
        <div className="glass-panel rounded-lg p-3 mt-1 text-xs text-ds-text-primary">
          <p className="mb-2 text-ds-text-muted">{t('handRanking.description')}</p>
          <ol className="grid grid-cols-2 gap-x-4 gap-y-1 list-decimal list-inside">
            {SUTDA_HAND_RANKING_KEYS.map((key) => (
              <li key={key}>{t(`handName.${key}`)}</li>
            ))}
          </ol>
        </div>
      </details>
    </>
  );
}

function Board({ state, t, view }: SutdaContext) {
  const { cardWidth, isMobile } = view;
  const isShowdown = state.phase === SutdaPhase.SHOWDOWN;
  const isGameEnd = state.phase === SutdaPhase.GAME_END || state.gameEndFlag;
  const lastResult = state.lastResult;
  const showdownWinners = lastResult?.winners.map((winner, index) => {
    const amount = lastResult.shares[index];
    const handName = lastResult.handNames[winner];
    const name = playerName(winner, winner === 0);
    if (amount === undefined) return handName ? t('winnerHand', { name, hand: t(`handName.${handName}`) }) : name;
    return handName
      ? t('winnerShareHand', { name, amount, hand: t(`handName.${handName}`) })
      : t('winnerShare', { name, amount });
  });
  const resultHandNumber = state.handNumber - (isShowdown || isGameEnd ? 0 : 1);

  const messageBox = (
    <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
  );

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('hand', { n: state.handNumber })}</span>
        <span data-testid="sutda-pot">{t('pot', { pot: state.pot, bet: state.currentBet })}</span>
      </div>

      <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="sutda-table" data-testid="sutda-table">
        {state.players.map((p) => (
          <section
            key={p.id}
            aria-labelledby={`sutda-player-${p.id}`}
            className="py-2 border-b border-white/5 last:border-0"
          >
            <div className="text-ds-text-muted text-sm flex items-center gap-2 mb-1">
              <span id={`sutda-player-${p.id}`} className={p.folded ? 'opacity-50 line-through' : ''}>
                {playerName(p.id, p.isHuman)}
              </span>
              {p.isDealer && (
                <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
              )}
              {p.folded && (
                <span className="text-xs text-ds-text-muted" data-testid={`sutda-folded-${p.id}`}>
                  {t('folded')}
                </span>
              )}
            </div>
            <div className="grid grid-cols-2 gap-2">
              <div className="rounded px-2 py-1 bg-ds-surface">
                <div className="text-xs text-ds-text-muted">{t('remainingChips')}</div>
                <div className="font-semibold tabular-nums text-ds-text-primary" data-testid={`sutda-chips-${p.id}`}>
                  {p.chips}
                </div>
              </div>
              <div className="rounded px-2 py-1 bg-ds-surface">
                <div className="text-xs text-ds-text-muted">{t('currentContribution')}</div>
                <div className="font-semibold tabular-nums text-ds-text-primary" data-testid={`sutda-bet-${p.id}`}>
                  {p.bet}
                </div>
              </div>
            </div>
            {/* **伏せているうちは自分のぶんだけ。** 相手の役が見えると
                賭ける意味が無くなる。 */}
            {p.cards.length > 0 && (
              <div className="flex items-center gap-1 mt-1" data-testid={`sutda-cards-${p.id}`}>
                {p.cards.map((c, i) => (
                  <CardImage key={`${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                ))}
                <span className="ml-2 text-ds-text-primary text-sm">{t(`handName.${p.handName}`)}</span>
              </div>
            )}
          </section>
        ))}
      </div>

      {state.lastResult && (
        <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="sutda-result">
          <p>{t('previousHand', { n: resultHandNumber })}</p>
          {t('showdownResult', {
            names: showdownWinners?.join(t('listSeparator')),
            pot: state.lastResult.pot,
          })}
        </div>
      )}

      {isGameEnd && (
        <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="sutda-winner">
          {t('winner', { name: playerName(state.winnerIdx, state.winnerIdx === 0) })}
        </div>
      )}

      {isMobile && messageBox}
    </>
  );
}

function FooterHand({ state, t, view }: SutdaContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBetPhase = state.phase === SutdaPhase.BET;
  const canAct = isBetPhase && state.isHumanTurn;
  return (
    <>
      {/* **自分の役は常に見える。** 伏せているのは相手の札であって、自分の
          2 枚が何の役かは配られた時点で分かっている。 */}
      {humanPlayer && (
        <div className="mb-2 flex items-center gap-2" data-tutorial="sutda-hand" data-testid="sutda-hand">
          {humanPlayer.cards.map((c, i) => (
            <CardImage key={`${c.design}-${c.value}-${i}`} card={c} width={view.cardWidth} />
          ))}
          <span className="text-ds-text-primary">{t(`handName.${state.humanHandName}`)}</span>
        </div>
      )}

      {canAct && (
        <div className="mb-1 text-sm text-ds-text-muted" data-testid="sutda-to-call">
          {state.callAmount > 0 ? t('toCall', { n: state.callAmount }) : t('canCheck')}
        </div>
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: SutdaContext) {
  return (
    <>
      {/* ライブ領域は常設 (#5955)。 */}
      <div data-testid="sutda-hint-live" role="status" aria-live="polite">
        {isRequestedHint(state) && state.hintAction && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`action.${state.hintAction}`)} ({t(`hint.${state.hintReason}`)})
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: SutdaContext) {
  const isBetPhase = state.phase === SutdaPhase.BET;
  const isShowdown = state.phase === SutdaPhase.SHOWDOWN;
  const canAct = isBetPhase && state.isHumanTurn;
  return (
    <>
      {canAct && (
        <button type="button" className={btnPrimary} onClick={game.handleCall} disabled={busy} data-testid="sutda-call">
          {state.callAmount > 0 ? t('callButton', { n: state.callAmount }) : t('checkButton')}
        </button>
      )}
      {/* **上限とチップの両方を見た結果が canRaise。** 押せない条件を
          画面側で組み直すと、サーバの判断と食い違う。 */}
      {canAct && state.canRaise && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleRaise}
          disabled={busy}
          data-testid="sutda-raise"
        >
          {t('raiseButton', { n: state.betUnit, left: state.maxRaises - state.raiseCount })}
        </button>
      )}
      {canAct && (
        <button type="button" className={btnDanger} onClick={game.handleFold} disabled={busy} data-testid="sutda-fold">
          {t('foldButton')}
        </button>
      )}
      {isShowdown && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleNextHand}
          disabled={busy}
          data-testid="sutda-next-hand"
        >
          {t('nextHand')}
        </button>
      )}
    </>
  );
}

const sutdaDefinition: GameDefinition<SutdaResponse, SutdaGame, Parameters<typeof sutdaApi.exec>, SutdaView> = {
  key: 'sutda',
  useGame: useSutdaGame,
  useView: useSutdaView,
  tutorial: SUTDA_TUTORIAL_STEPS,
  skeleton: { kind: 'casino-table', sections: [2, 2, 2], footerStyle: 'hand', footerHandSize: 2 },
  cli: () => ({
    parseCommand: parseSutdaCommand,
    formatResponse: formatSutdaState,
    helpText: SUTDA_HELP,
    localCommand: sutdaLocalCommand,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase] ?? view.phaseNames[SutdaPhase.BET] ?? '',
    ended: state.phase === SutdaPhase.GAME_END || state.gameEndFlag,
    won: (state.phase === SutdaPhase.GAME_END || state.gameEndFlag) && state.winnerIdx === 0,
  }),
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
          value: game.sutdaConfig.cpuDifficulty,
          options: SUTDA_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'seats',
          label: t('settings.seats'),
          value: game.sutdaConfig.seats,
          options: SUTDA_SEAT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('seats', v),
        },
        {
          type: 'select',
          id: 'startChips',
          label: t('settings.startChips'),
          value: game.sutdaConfig.startChips,
          options: SUTDA_CHIP_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('startChips', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) =>
    !view.isMobile ? (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ) : undefined,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'sutda-action-buttons',
  resetTutorialId: 'sutda-reset-button',
  AboveContent,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/**
 * Renders the Sutda page: the Korean two-card betting game on a 20-card
 * hanafuda pack.
 */
export const SutdaPage = createGamePage(sutdaDefinition);
