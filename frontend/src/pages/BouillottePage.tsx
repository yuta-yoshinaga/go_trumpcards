import type { TFunction } from 'i18next';
import { useEffect } from 'react';
import type { bouillotteApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import {
  ANTE_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useBouillotteGame,
} from '../hooks/useBouillotteGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { BouillotteResponse } from '../types/card';
import { BouillottePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { computeBouillottePotOdds } from '../utils/bouillottePotOdds';
import { BOUILLOTTE_HELP, parseBouillotteCommand } from '../utils/cli/commands/bouillotteCommands';
import { formatBouillotteState } from '../utils/cli/formatters/bouillotteFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { raiseAvailability, raiseCost } from '../utils/raiseAvailability';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Bouillotte tutorial step definitions. */
const BOUILLOTTE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bouillotte-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bouillotte-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bouillotte-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bouillotte-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BOUILLOTTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BouillottePhase.BETTING]: 'betting',
  [BouillottePhase.RESULT]: 'result',
};

type BouillotteGame = ReturnType<typeof useBouillotteGame>;
type BouillotteView = ReturnType<typeof useBouillotteView>;
type BouillotteContext = GamePageContext<BouillotteResponse, BouillotteGame, BouillotteView>;

function useBouillotteView({ game, t }: { game: BouillotteGame; t: TFunction }) {
  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('bouillotte', BOUILLOTTE_PHASE_KEYS);

  const state = game.state;
  const humanPlayer = state?.players.find((p) => p.isHuman);
  // **レイズが消えた理由は 2 つある。**上限と枚数不足を区別しないと、
  // 「レイズ 1/3回」がボタンの無い画面で「まだできる」と読めてしまう。
  const raiseInput = state
    ? {
        raiseCount: state.raiseCount,
        maxRaises: state.maxRaises,
        chips: humanPlayer?.chips ?? 0,
        currentBet: state.currentBet,
        roundBet: humanPlayer?.roundBet ?? 0,
        ante: state.ante,
      }
    : null;
  const raiseBlock = raiseInput ? raiseAvailability(raiseInput) : null;
  const raiseNeeded = raiseInput ? raiseCost(raiseInput) : 0;
  const raiseStatusText =
    !raiseInput || !state
      ? ''
      : raiseBlock === 'cap'
        ? t('raiseCapReached', { max: state.maxRaises })
        : raiseBlock === 'chips'
          ? t('raiseNoChips', { cost: raiseNeeded })
          : t('raiseCount', { count: state.raiseCount, max: state.maxRaises });

  return { cardWidth, isMobile, phaseNames, raiseStatusText, raiseNeeded };
}

function LiveRegion({ view }: BouillotteContext) {
  return (
    <div className="sr-only" role="status" aria-live="polite" data-testid="bouillotte-raise-announcement">
      {view.raiseStatusText}
    </div>
  );
}

function HeaderExtra({ state, t }: BouillotteContext) {
  return <span>{t('chips', { amount: state.chips })}</span>;
}

function Board({ state, t, view }: BouillotteContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  // Pot odds and chip costs facing the human at the Call/Raise/Fold decision.
  const humanRoundBet = humanPlayer?.roundBet ?? 0;
  const humanPayout = humanPlayer?.roundPayout ?? 0;
  const humanNet = humanPlayer?.netChange ?? 0;
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  // Which of the human's cards share the retourne's rank, and any combo it completes.
  // Computed server-side (#6494): reading state.retourneMatch avoids re-implementing the
  // matching logic on the client.
  const retourneMatch = state.retourneMatch ?? { matchingIndices: [], noteKey: '' };
  const matchingSet = new Set(retourneMatch.matchingIndices);

  const isResultPhase = state.phase === BouillottePhase.RESULT;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handName = (key?: string): string => (key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '');

  const playerBadge = (p: BouillotteResponse['players'][number]): string =>
    p.out ? t('badge.out') : p.isWinner ? t('badge.winner') : p.folded ? t('badge.folded') : t('badge.active');

  const messageBox = (
    <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
  );

  return (
    <>
      <div
        className="text-ds-text-primary text-center mb-2 flex flex-wrap justify-center gap-x-4"
        data-tutorial="bouillotte-info"
      >
        <span className="whitespace-nowrap">
          {t('round', { n: state.roundNumber, total: state.config.targetRounds })}
        </span>
        <span className="whitespace-nowrap">{t('pot', { amount: state.pot })}</span>
        <span className="whitespace-nowrap">{t('ante', { amount: state.ante })}</span>
        <span className="whitespace-nowrap">{t('currentBet', { amount: state.currentBet })}</span>
      </div>

      {/* Retourne (shared turned-up card) */}
      {state.retourne && (
        <div className="mb-2 flex flex-col items-center">
          <div className="text-ds-text-muted text-xs mb-0.5">{t('retourneLabel')}</div>
          <span
            data-testid="retourne-card"
            className={matchingSet.size > 0 ? 'inline-block rounded-md ring-2 ring-ds-accent' : 'inline-block'}
          >
            <CardImage card={state.retourne} width={cardWidth} />
          </span>
        </div>
      )}

      {/* Players */}
      <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="bouillotte-players">
        <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
        {state.players.map((p) => (
          <div
            key={p.id}
            className={`text-sm py-0.5 ${p.isWinner ? 'text-ds-success' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
          >
            {playerLabel(p.id, p.isHuman)} — {t('chips', { amount: p.chips })} · {t('roundBet', { amount: p.roundBet })}{' '}
            · [{playerBadge(p)}]{p.handName && <span className="whitespace-nowrap"> · {handName(p.handName)}</span>}
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
                <div className="text-ds-text-muted text-xs mb-0.5">
                  {playerLabel(p.id, p.isHuman)}
                  {p.handName ? ` — ${handName(p.handName)}` : ''}
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
              name: playerLabel(state.winnerIdx, state.winnerIdx === humanIdx),
              pot: state.pot,
            })}
          </div>
          <div data-testid="bouillotte-human-round-result">
            {t('roundResult.personal', {
              bet: humanRoundBet,
              payout: humanPayout,
              net: formatSignedDelta(humanNet),
            })}
          </div>
        </div>
      )}

      {isMobile && messageBox}
    </>
  );
}

function FooterHand({ state, t, view }: BouillotteContext) {
  const { cardWidth } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);

  // Which of the human's cards share the retourne's rank, and any combo it completes.
  // Computed server-side (#6494): reading state.retourneMatch avoids re-implementing the
  // matching logic on the client.
  const retourneMatch = state.retourneMatch ?? { matchingIndices: [], noteKey: '' };
  const matchingSet = new Set(retourneMatch.matchingIndices);
  const handName = (key?: string): string => (key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '');

  return humanPlayer && humanPlayer.cards.length > 0 ? (
    <div className="mb-2" data-tutorial="bouillotte-hand">
      <div className="text-ds-text-muted text-xs mb-0.5">
        {t('handLabel')}
        {humanPlayer.handName ? ` — ${handName(humanPlayer.handName)}` : ''}
      </div>
      <div className="flex gap-1">
        {humanPlayer.cards.map((c, i) => (
          <span
            key={`human-${i}`}
            data-testid={`hand-card-${i}`}
            className={matchingSet.has(i) ? 'inline-block rounded-md ring-2 ring-ds-accent' : 'inline-block'}
          >
            <CardImage card={c} width={cardWidth} />
          </span>
        ))}
      </div>
      {retourneMatch.noteKey && (
        <div className="mt-1 text-xs font-semibold text-ds-accent" data-testid="retourne-note">
          {t(`retourneNote.${retourneMatch.noteKey}`)}
        </div>
      )}
    </div>
  ) : (
    <div className="text-ds-text-muted text-sm mb-2" data-tutorial="bouillotte-hand">
      {t('handLabel')}
    </div>
  );
}

function FooterStatus({ state, t, frontendHint }: BouillotteContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanRoundBet = humanPlayer?.roundBet ?? 0;
  const potOdds = computeBouillottePotOdds(state.pot, state.currentBet, humanRoundBet);
  const isBettingPhase = state.phase === BouillottePhase.BETTING;
  const isHumanTurn = state.isHumanTurn;
  const isGameEnd = state.gameEndFlag;

  return (
    <>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />

      {isBettingPhase && isHumanTurn && !isGameEnd && (
        <div className="mb-2 text-ds-text-muted text-sm" data-testid="bouillotte-pot-odds" aria-live="polite">
          {potOdds.isFree
            ? t('potOdds.free')
            : t('potOdds.value', {
                call: potOdds.callAmount,
                pot: state.pot,
                percentage: potOdds.percentage,
                ratioPot: potOdds.ratioPot,
                ratioCall: potOdds.ratioCall,
              })}
        </div>
      )}
    </>
  );
}

function FooterActions({ game, state, t, view, busy }: BouillotteContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanRoundBet = humanPlayer?.roundBet ?? 0;
  const potOdds = computeBouillottePotOdds(state.pot, state.currentBet, humanRoundBet);
  const isBettingPhase = state.phase === BouillottePhase.BETTING;
  const isResultPhase = state.phase === BouillottePhase.RESULT;
  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;

  return (
    <>
      {isBettingPhase && isHumanTurn && !isGameEnd && (
        <>
          <button type="button" className={btnPrimary} onClick={game.handleCall} disabled={busy}>
            {potOdds.isFree ? t('callButton') : t('callButtonAmount', { amount: potOdds.callAmount })}
          </button>
          {/* **レイズが消えた理由を書く。**回数上限とチップ不足を
              区別できないと、突然選択肢を奪われたように見える (#4924)。 */}
          <span className="text-ds-text-muted text-xs" data-testid="bouillotte-raise-count">
            {view.raiseStatusText}
          </span>
          {state.canRaise && (
            <button type="button" className={btnSuccess} onClick={game.handleRaise} disabled={busy}>
              {t('raiseButtonAmount', { amount: view.raiseNeeded })}
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

const bouillotteDefinition: GameDefinition<
  BouillotteResponse,
  BouillotteGame,
  Parameters<typeof bouillotteApi.exec>,
  BouillotteView
> = {
  key: 'bouillotte',
  useGame: useBouillotteGame,
  useView: useBouillotteView,
  tutorial: BOUILLOTTE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 3 },
  cli: () => ({
    parseCommand: parseBouillotteCommand,
    formatResponse: formatBouillotteState,
    helpText: BOUILLOTTE_HELP,
  }),
  phase: ({ state, view }) => {
    const isResultPhase = state.phase === BouillottePhase.RESULT;
    const humanWonMatch = state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false);
    return {
      name: view.phaseNames[state.phase],
      ended: state.gameEndFlag,
      won: isResultPhase && (humanWonMatch || state.result > 0),
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.phase === BouillottePhase.BETTING && state.isHumanTurn && !state.gameEndFlag,
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
          value: game.bouillotteConfig.playerCount,
          options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('playerCount', v),
        },
        {
          type: 'select',
          id: 'ante',
          label: t('settings.ante'),
          value: game.bouillotteConfig.ante,
          options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('ante', v),
        },
        {
          type: 'select',
          id: 'startingChips',
          label: t('settings.startingChips'),
          value: game.bouillotteConfig.startingChips,
          options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('startingChips', v),
        },
        {
          type: 'select',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.bouillotteConfig.targetRounds,
          options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetRounds', v),
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
    ) : null,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'bouillotte-action-buttons',
  resetTutorialId: 'bouillotte-reset-button',
  LiveRegion,
  HeaderExtra,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Bouillotte game page: an 18th-century French 3-card pot-vying game. */
export const BouillottePage = createGamePage(bouillotteDefinition);
