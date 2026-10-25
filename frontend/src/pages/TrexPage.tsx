import { Fragment } from 'react';
import type { trexApi } from '../api/gameApi';
import { CardBack } from '../components/CardImage';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useTrexGame } from '../hooks/useTrexGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { TrexResponse } from '../types/card';
import { TrexPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTrexCommand, TREX_HELP } from '../utils/cli/commands/trexCommands';
import { formatTrexState } from '../utils/cli/formatters/trexFormatter';
import { trexCardPenalty } from '../utils/trexPenaltyCards';

const TX_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tx-rule"]', messageKey: 'tutorial.contracts', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tx-header"]', messageKey: 'tutorial.kingdom', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tx-table"]', messageKey: 'tutorial.dominoes', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tx-seats"]', messageKey: 'tutorial.score', placement: 'bottom', advanceOn: 'next' },
];

const SUIT_GLYPHS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

const TREX_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TrexPhase.CHOOSE]: 'choose',
  [TrexPhase.PLAY]: 'play',
  [TrexPhase.DEAL_END]: 'dealEnd',
  [TrexPhase.GAME_END]: 'end',
};

type TrexGame = ReturnType<typeof useTrexGame>;
type TrexView = ReturnType<typeof useTrexView>;
type TrexContext = GamePageContext<TrexResponse, TrexGame, TrexView>;

function useTrexView() {
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('trex', TREX_PHASE_KEYS);
  return { cardWidth, phaseNames };
}

function getContractName(t: TrexContext['t'], n: number): string {
  return t(`contract${n}`, { defaultValue: t('contractNone') });
}

function HeaderExtra({ state, t }: TrexContext) {
  return (
    <span className="text-sm text-ds-text-muted" data-tutorial="tx-header">
      {t('deal')}: {state.dealNo}/{state.totalDeals}
      {' / '}
      {t('king')}: {t('seat', { n: state.kingIdx })}
      {' / '}
      {t('contract')}: {getContractName(t, state.contract)}
    </span>
  );
}

function Board({ game, state, t, view, frontendHint }: TrexContext) {
  const { cardWidth } = view;
  const { hint: frontendHintVal, hintEnabled: frontendHintEnabled } = frontendHint;
  const ended = state.phase === TrexPhase.GAME_END;
  const choosing = state.phase === TrexPhase.CHOOSE;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  // Playability comes from the server: each contract follows a different rule,
  // and the dominoes are not a trick game at all.
  const playable = new Set(state.validIndices);

  return (
    <>
      {/* Permanent, not tutorial-only: "once per kingdom" and "the
          dominoes start from the JACK" are the two things a player gets
          wrong. */}
      <details className="mb-2 p-2 rounded bg-ds-surface" data-testid="tx-rule-details">
        <summary className="cursor-pointer select-none text-center text-xs text-ds-warning font-medium">
          {t('rulesLabel')}
        </summary>
        <p className="mt-2 text-center text-xs text-ds-warning" data-tutorial="tx-rule">
          {t('ruleLine')}
        </p>
      </details>

      <CpuAccordion playerCount={opponents.length} dataTutorial="tx-seats" openOnDesktop={false}>
        <div className="flex flex-wrap sm:flex-nowrap justify-center gap-4 mb-3">
          {opponents.map((o) => (
            <div key={`opp-${o.id.toString()}`} className="text-center">
              <div className="text-game-text-muted text-xs mb-1">
                {t('opponentHand', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                {' · '}
                {t('score', { n: o.score })}
                {o.id === state.kingIdx && ` · ${t('king')}`}
              </div>
              <div
                className="flex gap-1 justify-center flex-wrap"
                role="img"
                aria-label={t('opponentHandAriaLabel', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
              >
                {Array.from({ length: o.cardCount }, (_, i) => (
                  <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                ))}
              </div>
            </div>
          ))}
        </div>
      </CpuAccordion>

      <div className="text-center mb-4" data-tutorial="tx-table">
        {state.isTrix ? (
          <>
            <div className="text-game-text-muted text-xs mb-1">{t('runs')}</div>
            <div className="flex gap-3 justify-center flex-wrap text-sm">
              {state.runs.map((run) => (
                <span key={`run-${run.suit.toString()}`} data-testid="trex-run">
                  {SUIT_GLYPHS[run.suit] ?? '?'}{' '}
                  {run.started ? `${run.low.toString()}–${run.high.toString()}` : t('runEmpty')}
                </span>
              ))}
            </div>
          </>
        ) : (
          <>
            <div className="text-game-text-muted text-xs mb-1">{t('trick')}</div>
            <div className="flex gap-1 justify-center flex-wrap">
              {state.trick.length === 0 ? (
                <span className="text-game-text-muted text-xs">—</span>
              ) : (
                state.trick.map((tc2) => {
                  // **失点対象は契約ごとに違う。**5 種が 1 王国内で切り替わる
                  // ので、どれが危険かを記憶と暗算で追わせることになる (#4911)。
                  const penalty = trexCardPenalty(tc2.card, state.contract);
                  return (
                    <div
                      key={`trick-${tc2.playerIdx.toString()}`}
                      className={penalty !== 0 ? 'rounded ring-2 ring-ds-error' : ''}
                      data-testid={penalty !== 0 ? 'trex-penalty-card' : undefined}
                      title={penalty !== 0 ? t('penaltyCardWithPoints', { points: penalty }) : undefined}
                    >
                      <AnimatedCard
                        card={tc2.card}
                        width={cardWidth}
                        draggable={false}
                        ariaLabel={
                          penalty !== 0
                            ? t('trickCardPenaltyAria', { card: cardAlt(tc2.card), points: penalty })
                            : undefined
                        }
                      />
                    </div>
                  );
                })
              )}
            </div>
          </>
        )}
      </div>

      <div className="text-center" data-tutorial="tx-hand">
        <div className="text-game-text-muted text-xs mb-1">
          {t('yourHand')}
          {' · '}
          {t('score', { n: human?.score ?? 0 })}
          {' · '}
          {t('dealScore', { n: human?.dealScore ?? 0 })}
        </div>
        <div className="flex gap-1 justify-center flex-wrap">
          {(human?.cards ?? []).map((card, i) => {
            const canPlay = isHumanTurn && !choosing && playable.has(i);
            const unavailableReason = choosing
              ? t('handDisabled.contract')
              : !isHumanTurn
                ? t('handDisabled.turn')
                : t('handDisabled.illegal');
            const penalty = trexCardPenalty(card, state.contract);
            const describedBy = [
              !canPlay ? `trex-hand-disabled-${i.toString()}` : undefined,
              penalty !== 0 ? `trex-hand-penalty-${i.toString()}` : undefined,
            ]
              .filter(Boolean)
              .join(' ');
            return (
              <Fragment key={`hand-${i.toString()}`}>
                <button
                  type="button"
                  data-hint-action="play"
                  data-testid={penalty !== 0 ? 'trex-hand-penalty-card' : undefined}
                  title={penalty !== 0 ? t('penaltyCardWithPoints', { points: penalty }) : undefined}
                  // Kept focusable while it cannot act so the reason is
                  // announced rather than the control leaving the tab order.
                  aria-disabled={!canPlay}
                  aria-describedby={describedBy || undefined}
                  onClick={() => canPlay && game.handlePlay(i)}
                  className={[
                    'rounded transition-transform',
                    canPlay ? 'hover:-translate-y-2' : 'opacity-60',
                    frontendHintEnabled && state.hint?.cardIndex === i
                      ? 'ring-2 ring-ds-warning'
                      : penalty !== 0
                        ? 'ring-2 ring-ds-error'
                        : '',
                  ].join(' ')}
                >
                  <AnimatedCard card={card} width={cardWidth} draggable={false} />
                </button>
                {!canPlay && (
                  <span id={`trex-hand-disabled-${i.toString()}`} className="sr-only">
                    {unavailableReason}
                  </span>
                )}
                {penalty !== 0 && (
                  <span id={`trex-hand-penalty-${i.toString()}`} className="sr-only">
                    {t('penaltyCardWithPoints', { points: penalty })}
                  </span>
                )}
              </Fragment>
            );
          })}
        </div>
        <FrontendHintTooltip hint={frontendHintVal} enabled={frontendHintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterStatus({ game, state, t }: TrexContext) {
  const choosing = state.phase === TrexPhase.CHOOSE;
  const humanIsKing = state.kingIdx === 0;

  return (
    <>
      {choosing && humanIsKing && (
        <div className="mb-2" data-tutorial="tx-choose">
          <div className="text-xs text-ds-text-muted mb-1">{t('choosePrompt')}</div>
          <div className="flex gap-2 flex-wrap items-center">
            {/* Only what this king has not played. A contract is played
                once per kingdom, so offering a spent one is a dead end. */}
            {state.availableContracts.map((n) => (
              <div key={`contract-${n.toString()}`} className="flex flex-col items-start gap-1">
                <button
                  type="button"
                  data-hint-action="choose"
                  aria-describedby={`trex-contract-desc-${n.toString()}`}
                  className={`${btnPrimary} min-h-11`}
                  onClick={() => game.handleChoose(n)}
                >
                  {getContractName(t, n)}
                </button>
                <span id={`trex-contract-desc-${n.toString()}`} className="text-xs text-ds-text-muted">
                  {t(`contractDescription${n}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
      {choosing && !humanIsKing && (
        <div className="mb-2 text-xs text-ds-text-muted">{t('waitingForKing', { n: state.kingIdx })}</div>
      )}
    </>
  );
}

function FooterActions({ game, state, t }: TrexContext) {
  const ended = state.phase === TrexPhase.GAME_END;
  const dealOver = state.phase === TrexPhase.DEAL_END;

  return (
    <>
      {/* Enabled only in the dominoes with no legal play. Trick
          contracts have no pass at all. */}
      {state.isTrix && (
        <button
          type="button"
          data-hint-action="pass"
          aria-disabled={!state.canPass}
          onClick={() => state.canPass && game.handlePass()}
          className={[
            'px-4 py-2 rounded font-bold min-h-11',
            state.canPass ? 'bg-ds-accent text-ds-text-on-accent' : 'bg-ds-surface-elevated text-ds-text-muted',
          ].join(' ')}
        >
          {t('pass')}
        </button>
      )}
      {state.canPass && <span className="text-xs text-ds-text-muted">{t('passHint')}</span>}
      {dealOver && !ended && (
        <button type="button" className={`${btnSecondary} min-h-11`} onClick={game.handleNextDeal}>
          {t('nextDeal')}
        </button>
      )}
    </>
  );
}

const trexDefinition: GameDefinition<TrexResponse, TrexGame, Parameters<typeof trexApi.exec>, TrexView> = {
  key: 'trex',
  useGame: useTrexGame,
  useView: useTrexView,
  tutorial: TX_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 3, tableau: 4 },
  landscapeBanner: true,
  landscapeBannerInCliMode: true,
  settingsPlacement: 'top',
  settingsInCliMode: true,
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  actionLogDuringPlay: true,
  cli: () => ({
    parseCommand: parseTrexCommand,
    formatResponse: formatTrexState,
    helpText: TREX_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase] ?? '',
    ended: state.phase === TrexPhase.GAME_END,
  }),
  resetTutorialId: 'tx-reset-button',
  HeaderExtra,
  Board,
  FooterStatus,
  FooterActions,
};

/** Renders the Trex page: the king picks one of five contracts, twenty deals in all. */
export const TrexPage = createGamePage(trexDefinition);
