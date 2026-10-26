import { useState } from 'react';
import { killeApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { KbdBadge } from '../components/KbdBadge';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { KilleEvent, KillePlayer, KilleResponse } from '../types/card';
import { KillePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KILLE_HELP, parseKilleCommand } from '../utils/cli/commands/killeCommands';
import { formatKilleState } from '../utils/cli/formatters/killeFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';

/** Stake options for the Kille settings panel. */
const STAKE_OPTIONS = [1, 2, 5, 10];

/** Buy-backs allowed per player (sync: `KilleMaxReentries` in internal/domain/KillePlayer.go). */
const KILLE_MAX_REENTRIES = 3;

/**
 * The pack in descending strength, for the on-page reference.
 *
 * Number cards are listed individually so their strength relative to effect
 * cards is clear at a glance.
 */
const KILLE_LADDER = [
  { key: 'harlequin', color: 'text-ds-accent' },
  { key: 'cuckoo', color: 'text-ds-info' },
  { key: 'hussar', color: 'text-ds-error-text' },
  { key: 'pig', color: 'text-ds-warning' },
  { key: 'cavalier', color: 'text-ds-success' },
  { key: 'inn', color: 'text-ds-success' },
  ...Array.from({ length: 12 }, (_, index) => ({ key: String(12 - index), color: 'text-ds-text-muted' })),
  { key: 'wreath', color: 'text-ds-text-muted' },
  { key: 'flowerpot', color: 'text-ds-text-muted' },
  { key: 'mask', color: 'text-ds-text-muted' },
];

const KILLE_EFFECTS = [
  { glyph: '🐦', nameKey: 'cards.cuckoo', descriptionKey: 'effect.cuckoo' },
  { glyph: '⚔', nameKey: 'cards.hussar', descriptionKey: 'effect.hussar' },
  { glyph: '🐖', nameKey: 'cards.pig', descriptionKey: 'effect.pig' },
  { glyph: '🐎', nameKey: 'cards.cavalier', descriptionKey: 'effect.cavalier' },
  { glyph: '🏠', nameKey: 'cards.inn', descriptionKey: 'effect.inn' },
] as const;

/** Kille tutorial step definitions. */
const KILLE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="kille-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kille-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kille-card"]',
    messageKey: 'tutorial.card',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kille-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kille-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KILLE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KillePhase.EXCHANGE]: 'exchange',
  [KillePhase.SHOWDOWN]: 'showdown',
  [KillePhase.GAME_END]: 'gameEnd',
};

type KilleGame = ReturnType<typeof useKilleGame>;
type KilleView = ReturnType<typeof useKilleView>;
type KilleContext = GamePageContext<KilleResponse, KilleGame, KilleView>;

function useKilleGame() {
  const { state, loading, error, exec, retry } = useGameApi(killeApi.exec);

  const [stake, setStake] = useState(1);

  useMountReset(exec);

  const handleStakeChange = (value: string) => {
    const next = Number(value);
    setStake(next);
    void exec('reset', { config: { stake: next } });
  };

  const handleReset = () => {
    void exec('reset', { config: { stake } });
  };

  return {
    state,
    loading,
    error,
    exec,
    retry,
    stake,
    handleStakeChange,
    handleReset,
  };
}

function useKilleView() {
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('kille', KILLE_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function KilleMessageBox({ state }: Pick<KilleContext, 'state'>) {
  return <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />;
}

function Board({ state, t, view }: KilleContext) {
  const { cardWidth, isMobile } = view;
  const isShowdown = state.phase === KillePhase.SHOWDOWN;
  const isGameEnd = state.phase === KillePhase.GAME_END || state.gameEndFlag;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  /** Why a seat went out. The Hussar and the Pig fire regardless of strength. */
  const outReason = (p: KillePlayer): string => {
    if (p.knockedBy === 'hussar') return t('outHussar');
    if (p.knockedBy === 'pig') return t('outPig');
    return t('outLowest');
  };

  const eventLine = (e: KilleEvent): string =>
    t(`event.${e.kind}`, {
      actor: playerLabel(e.actor, e.actor === 0),
      target: e.target >= 0 ? playerLabel(e.target, e.target === 0) : t('stock', { count: state.stockCount }),
      defaultValue: '',
    });

  return (
    <>
      <div
        className="flex flex-wrap justify-center gap-x-4 text-ds-text-primary text-center mb-2"
        data-tutorial="kille-info"
      >
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">{t('pot', { count: state.pot })}</span>
        <span className="whitespace-nowrap">
          {t('dealer')}: {playerLabel(state.dealerIdx, state.dealerIdx === 0)}
        </span>
        <span className="whitespace-nowrap" data-testid="kille-stock">
          {t('stock', { count: state.stockCount })}
        </span>
      </div>

      <details
        className="mb-2 text-center text-ds-text-muted text-xs"
        data-testid="kille-rules"
        open={!isMobile || undefined}
      >
        <summary className="cursor-pointer select-none text-ds-text-primary">{t('rulesLabel')}</summary>
        <div className="pt-1" data-testid="kille-rules-note">
          {t('rulesNote')}
        </div>
      </details>

      {/* Players */}
      <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="kille-players">
        <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
        {state.players
          .filter((p) => p.isHuman)
          .map((p) => (
            <div
              key={`player-${p.id}`}
              data-testid="kille-player"
              className={`text-sm py-0.5 flex flex-wrap items-center gap-2 ${
                p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
              } font-semibold ${p.isFinished ? 'opacity-50' : ''}`}
            >
              <span>{playerLabel(p.id, p.isHuman)}</span>
              {p.id === state.dealerIdx && <span className="text-ds-accent">[{t('dealer')}]</span>}
              <span>{t('chips', { n: p.chips })}</span>
              <span>{t('reentriesRemaining', { count: KILLE_MAX_REENTRIES - p.reentries })}</span>
              {p.isSatisfied && !p.isOut && <span className="text-ds-success">[{t('satisfied')}]</span>}
              {p.isOut && <span className="text-ds-error-text">[{outReason(p)}]</span>}
              {p.isFinished && <span>({t('eliminated')})</span>}
            </div>
          ))}
        <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
          <div className="flex flex-wrap gap-x-3">
            {state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="kille-player"
                  className={`text-sm py-0.5 flex flex-wrap items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isFinished ? 'opacity-50' : ''}`}
                >
                  <span>{playerLabel(p.id, false)}</span>
                  {p.id === state.dealerIdx && <span className="text-ds-accent">[{t('dealer')}]</span>}
                  <span>{t('chips', { n: p.chips })}</span>
                  <span>{t('reentriesRemaining', { count: KILLE_MAX_REENTRIES - p.reentries })}</span>
                  {p.isSatisfied && !p.isOut && <span className="text-ds-success">[{t('satisfied')}]</span>}
                  {p.isOut && <span className="text-ds-error-text">[{outReason(p)}]</span>}
                  {p.isFinished && <span>({t('eliminated')})</span>}
                  {p.card && <CardImage card={p.card} width={cardWidth} />}
                </div>
              ))}
          </div>
        </CpuAccordion>
      </div>

      {/* Exchanges this round. The region is mounted unconditionally and only
          its contents vary: a live region created in the same paint as its
          first text is not reliably announced. */}
      <div
        className={state.events.length > 0 ? 'mb-2 p-2 rounded bg-black/20 text-sm' : 'sr-only'}
        data-testid="kille-events"
        role="status"
        aria-live="polite"
      >
        {state.events.length > 0 && (
          <>
            <div className="mb-1 text-ds-text-primary">{t('eventsTitle')}</div>
            {state.events.map((e, i) => (
              <div key={`event-${e.kind}-${e.actor}-${e.target}-${i}`} className="text-ds-text-muted">
                {eventLine(e)}
              </div>
            ))}
          </>
        )}
      </div>

      {/* Showdown */}
      <div
        className={
          (isShowdown || isGameEnd) && state.loserIdxs.length > 0
            ? `mb-2 p-2 rounded text-sm ${badgeWarningColors}`
            : 'sr-only'
        }
        data-testid="kille-showdown"
        role="status"
        aria-live="polite"
      >
        {(isShowdown || isGameEnd) && (state.loserIdxs.length > 0 || (isGameEnd && state.winnerIdx >= 0)) && (
          <>
            <div className="mb-1">{t('showdownTitle')}</div>
            {isGameEnd && state.winnerIdx >= 0 ? (
              <div>{t('gameWinner', { name: playerLabel(state.winnerIdx, state.winnerIdx === 0) })}</div>
            ) : (
              <div>
                {t('showdownSummary', {
                  names: state.loserIdxs.map((idx) => playerLabel(idx, idx === 0)).join(t('listSeparator')),
                })}
              </div>
            )}
            {state.loserIdxs.map((idx) => (
              <div key={`loser-${idx}`}>
                {t('loserLine', {
                  name: playerLabel(idx, idx === 0),
                  reason: outReason(state.players[idx]),
                })}
              </div>
            ))}
          </>
        )}
      </div>

      {/* The pack, for reference: a single suit means denomination is everything. */}
      <details
        className="mb-2 p-2 rounded bg-black/20 text-xs"
        data-testid="kille-ladder"
        open={!isMobile || undefined}
      >
        <summary className="cursor-pointer select-none text-ds-text-primary">{t('deckLabel')}</summary>
        <div className="pt-1">
          <div className="mb-1 text-ds-text-primary">{t('deckTitle')}</div>
          <div className="flex flex-wrap gap-x-2 gap-y-1">
            {KILLE_LADDER.map((r) => (
              <span key={r.key} className={r.color}>
                {t(`cards.${r.key}`)}
              </span>
            ))}
          </div>
          <div className="mt-1 text-ds-text-muted">{t('deckNote')}</div>
          <div className="mt-2 grid gap-1" data-testid="kille-effect-legend">
            <div className="text-ds-text-primary">{t('effectLegendTitle')}</div>
            {KILLE_EFFECTS.map((effect) => (
              <div key={effect.nameKey}>
                <span aria-hidden="true">{effect.glyph}</span>{' '}
                <span className="font-semibold">{t(effect.nameKey)}</span>: {t(effect.descriptionKey)}
              </div>
            ))}
          </div>
        </div>
      </details>

      {isMobile && <KilleMessageBox state={state} />}
    </>
  );
}

function FooterHand({ state, t, view }: KilleContext) {
  const { cardWidth } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);

  return (
    <div className="mb-2" data-tutorial="kille-card">
      <div className="text-ds-text-muted text-xs mb-1">{t('yourCard')}</div>
      {humanPlayer?.card ? (
        <CardImage card={humanPlayer.card} width={cardWidth} />
      ) : (
        <div className="text-ds-text-muted text-sm">{t('out')}</div>
      )}
    </div>
  );
}

function FooterStatus({ state, t, tc, frontendHint }: KilleContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isShowdown = state.phase === KillePhase.SHOWDOWN;
  const isGameEnd = state.phase === KillePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn =
    state.phase === KillePhase.EXCHANGE && state.currentPlayerIdx === 0 && !isGameEnd && !humanPlayer?.isOut;
  const humanIsDealer = humanPlayer ? state.dealerIdx === humanPlayer.id : false;
  const humanIsOutForGood = isShowdown && !!humanPlayer?.isOut && !humanPlayer?.canReenter;

  return (
    <>
      {isHumanTurn && (
        <div className="text-ds-text-muted text-xs mb-2" data-testid="kille-turn-notice">
          {humanIsDealer ? t('dealerNotice') : t('turnNotice')}
        </div>
      )}
      {humanIsOutForGood && (
        <div className="text-ds-text-muted text-xs mb-2" data-testid="kille-reenter-exhausted">
          {t('reenterExhausted')}
        </div>
      )}

      <label className="flex items-center gap-1 text-ds-text-primary text-xs w-full justify-center cursor-pointer min-h-[44px]">
        <input
          type="checkbox"
          checked={frontendHint.hintEnabled}
          onChange={(e) => frontendHint.setHintEnabled(e.target.checked)}
        />
        {tc('hint.toggle', { ns: 'tutorial' })}
      </label>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: KilleContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isShowdown = state.phase === KillePhase.SHOWDOWN;
  const isGameEnd = state.phase === KillePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn =
    state.phase === KillePhase.EXCHANGE && state.currentPlayerIdx === 0 && !isGameEnd && !humanPlayer?.isOut;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  // The dealer swaps with the stock rather than a neighbour, so its button says
  // something different and it can never be challenged.
  const humanIsDealer = humanPlayer ? state.dealerIdx === humanPlayer.id : false;
  const humanCanReenter = isShowdown && !isGameEnd && !!humanPlayer?.isOut && !!humanPlayer?.canReenter;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  return (
    <>
      {isHumanTurn && (
        <>
          <button type="button" className={btnPrimary} onClick={() => void game.exec('exchange')} disabled={busy}>
            {humanIsDealer ? t('exchangeDealerButton') : t('exchangeButton')}
            <KbdBadge label="E" />
          </button>
          <button type="button" className={btnSuccess} onClick={() => void game.exec('satisfied')} disabled={busy}>
            {t('satisfiedButton')}
            <KbdBadge label="S" />
          </button>
        </>
      )}

      {humanCanReenter && (
        <button
          type="button"
          className={btnWarning}
          onClick={() => void game.exec('reenter')}
          disabled={busy}
          data-testid="kille-reenter-button"
        >
          {t('reenterButton', { cost: humanPlayer?.reentryCost ?? 0 })}
        </button>
      )}

      {isShowdown && !isGameEnd && (
        <button type="button" className={btnSuccess} onClick={() => void game.exec('nextround')} disabled={busy}>
          {t('nextRound')}
          <KbdBadge label="N" />
        </button>
      )}

      {isGameEnd && (
        <span className="text-ds-text-primary text-sm font-semibold mr-1">
          {humanWon ? t('win') : t('lose', { name: playerLabel(state.winnerIdx, state.winnerIdx === 0) })}
        </span>
      )}
    </>
  );
}

const killeDefinition: GameDefinition<KilleResponse, KilleGame, Parameters<typeof killeApi.exec>, KilleView> = {
  key: 'kille',
  useGame: useKilleGame,
  useView: useKilleView,
  tutorial: KILLE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 1 },
  cli: ({ frontendHint }) => ({
    parseCommand: parseKilleCommand,
    formatResponse: formatKilleState,
    helpText: KILLE_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === KillePhase.GAME_END || state.gameEndFlag;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerIdx === 0,
    };
  },
  shell: ({ state }) => {
    const humanPlayer = state.players.find((p) => p.isHuman);
    const isGameEnd = state.phase === KillePhase.GAME_END || state.gameEndFlag;
    return {
      isHumanTurn:
        state.phase === KillePhase.EXCHANGE && state.currentPlayerIdx === 0 && !isGameEnd && !humanPlayer?.isOut,
    };
  },
  settings: ({ game, t }) => [
    {
      items: [
        {
          type: 'select',
          id: 'stake',
          label: t('settings.stake'),
          value: game.stake,
          options: STAKE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: game.handleStakeChange,
        },
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  shortcuts: ({ game, state }) => {
    const kbIsHumanTurn =
      state.phase === KillePhase.EXCHANGE &&
      state.currentPlayerIdx === 0 &&
      !state.gameEndFlag &&
      !state.players[0].isOut;
    const kbIsShowdown = state.phase === KillePhase.SHOWDOWN && !state.gameEndFlag;
    return [
      { key: 'e', action: () => void game.exec('exchange'), enabled: !!kbIsHumanTurn },
      { key: 's', action: () => void game.exec('satisfied'), enabled: !!kbIsHumanTurn },
      { key: 'n', action: () => void game.exec('nextround'), enabled: !!kbIsShowdown },
    ];
  },
  message: ({ state, view }) => (!view.isMobile ? <KilleMessageBox state={state} /> : null),
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'kille-actions',
  resetTutorialId: 'kille-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Kille game page: the Swedish Cuckoo game on its own 42-card pack. */
export const KillePage = createGamePage(killeDefinition);
