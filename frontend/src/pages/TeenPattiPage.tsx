import { useEffect, useState } from 'react';
import type { teenPattiApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { ANTE_OPTIONS, STARTING_CHIPS_OPTIONS, useTeenPattiGame } from '../hooks/useTeenPattiGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { TeenPattiResponse } from '../types/card';
import { TeenPattiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTeenPattiCommand, TEEN_PATTI_HELP } from '../utils/cli/commands/teenPattiCommands';
import { formatTeenPattiState } from '../utils/cli/formatters/teenPattiFormatter';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Teen Patti tutorial step definitions. */
const TEEN_PATTI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="teenpatti-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-action-buttons"]',
    messageKey: 'tutorial.bet',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-sideshow"]',
    messageKey: 'tutorial.sideShow',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TEEN_PATTI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TeenPattiPhase.BETTING]: 'betting',
  [TeenPattiPhase.SIDE_SHOW]: 'sideShow',
  [TeenPattiPhase.SHOWDOWN]: 'showdown',
  [TeenPattiPhase.ROUND_END]: 'roundEnd',
  [TeenPattiPhase.GAME_END]: 'gameEnd',
};

// No card selection in Teen Patti — bets are placed, not cards. Provide a stable no-op.
const noopToggle = () => {};

type TeenPattiGame = ReturnType<typeof useTeenPattiGame>;

function useTeenPattiView({ game }: { game: TeenPattiGame }) {
  // Raise stake amount (local UI state).
  const [raiseStake, setRaiseStake] = useState(2);

  // **サーバーが送る範囲に追従させる。**CPU がレイズすると minRaise が上がるが、
  // ローカルの raiseStake は据え置きなので、下限を割った額を送信できてしまう
  // (上限を塞いだのと同じ穴が下側に残っていた)。範囲が動くたびに畳み込む。
  // state は読み込み前 null になり得るが、フックは条件付きで呼べない。
  const minRaise = game.state?.minRaise ?? 0;
  const maxRaise = game.state?.maxRaise ?? 0;
  useEffect(() => {
    if (minRaise <= 0) return; // まだ状態が来ていない
    setRaiseStake((a) => Math.min(Math.max(a, minRaise), Math.max(minRaise, maxRaise)));
  }, [minRaise, maxRaise]);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('teenpatti', TEEN_PATTI_PHASE_KEYS);

  return {
    raiseStake,
    setRaiseStake,
    cardWidth,
    isMobile,
    phaseNames,
  };
}

type TeenPattiView = ReturnType<typeof useTeenPattiView>;
type TeenPattiContext = GamePageContext<TeenPattiResponse, TeenPattiGame, TeenPattiView>;

function Board({ game, state, t, view, busy }: TeenPattiContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isSideShowPhase = state.phase === TeenPattiPhase.SIDE_SHOW;
  const isRoundEnd = state.phase === TeenPattiPhase.ROUND_END;
  const isGameEnd = state.phase === TeenPattiPhase.GAME_END || state.gameEndFlag;
  // The human is being asked to accept/decline a pending Side Show.
  const isHumanSideShowTarget = isSideShowPhase && state.sideShowTarget === humanIdx;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handName = (key?: string): string => (key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="teenpatti-info">
        <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('pot', { amount: state.pot })}</span>
        <span>{t('stake', { amount: state.stake })}</span>
      </div>

      {/* Players */}
      <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="teenpatti-players">
        <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
        {state.players.map((p) => {
          const badge = p.out
            ? t('badge.out')
            : p.folded
              ? t('badge.folded')
              : p.seen
                ? t('badge.seen')
                : t('badge.blind');
          return (
            <div
              key={p.id}
              className={`text-sm py-0.5 ${p.id === state.currentPlayerIdx ? 'text-ds-warning' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
            >
              {playerLabel(p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
              {t('roundBet', { amount: p.roundBet })} · [{badge}]{p.handName ? ` · ${handName(p.handName)}` : ''}
              {p.id === state.currentPlayerIdx && <span className="ml-2 font-semibold">{t('turnIndicator')}</span>}
              {p.id === state.dealerIdx && (
                <span className={`ml-2 px-1.5 py-0.5 rounded text-xs whitespace-nowrap shrink-0 ${badgeWarningColors}`}>
                  {t('dealerBadge')}
                </span>
              )}
            </div>
          );
        })}
      </div>

      {/* Pending Side Show request + response — unified, emphasized panel */}
      {isSideShowPhase && state.sideShowRequester >= 0 && state.sideShowTarget >= 0 && (
        <div
          className="mb-2 p-3 rounded bg-black/30 ring-2 ring-ds-warning motion-safe:animate-pulse"
          data-tutorial="teenpatti-sideshow"
          data-testid="teenpatti-sideshow-panel"
        >
          <div className="text-ds-warning text-sm font-semibold mb-1">{t('sideShowTitle')}</div>
          <div className="text-ds-text-primary text-sm">
            {t('sideShowPending', {
              requester: playerLabel(state.sideShowRequester, state.sideShowRequester === humanIdx),
              target: playerLabel(state.sideShowTarget, state.sideShowTarget === humanIdx),
            })}
          </div>
          {isHumanSideShowTarget && (
            <div className="flex flex-wrap gap-2 mt-2">
              <button
                type="button"
                className={btnSuccess}
                onClick={() => game.handleRespondSideShow(true)}
                disabled={busy}
                title={t('consequence.accept')}
                aria-describedby="teenpatti-accept-desc"
              >
                {t('acceptButton')}
              </button>
              <button
                type="button"
                className={btnDanger}
                onClick={() => game.handleRespondSideShow(false)}
                disabled={busy}
                title={t('consequence.refuse')}
                aria-describedby="teenpatti-refuse-desc"
              >
                {t('declineButton')}
              </button>
              <span id="teenpatti-accept-desc" className="sr-only">
                {t('consequence.accept')}
              </span>
              <span id="teenpatti-refuse-desc" className="sr-only">
                {t('consequence.refuse')}
              </span>
              <div
                className="sm:hidden basis-full text-xs text-ds-text-muted mt-1 space-y-0.5"
                data-testid="teenpatti-sideshow-consequences"
                aria-hidden="true"
              >
                <div>
                  {t('acceptButton')}: {t('consequence.accept')}
                </div>
                <div>
                  {t('declineButton')}: {t('consequence.refuse')}
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Resolved Side Show comparison — both hands + who won (human-involved only) */}
      {state.lastSideShow && (
        <div className="mb-2 p-3 rounded bg-black/30 ring-1 ring-ds-warning/60" data-testid="teenpatti-sideshow-result">
          <div className="text-ds-warning text-sm font-semibold mb-1">{t('sideShowResult.title')}</div>
          <div className="text-ds-text-primary text-sm mb-2">
            {t('sideShowResult.outcome', {
              winner: playerLabel(state.lastSideShow.winnerIdx, state.lastSideShow.winnerIdx === humanIdx),
              loser: playerLabel(state.lastSideShow.loserIdx, state.lastSideShow.loserIdx === humanIdx),
              winnerHand: handName(
                (state.lastSideShow.winnerIdx === state.lastSideShow.requester.playerIdx
                  ? state.lastSideShow.requester
                  : state.lastSideShow.target
                ).handName,
              ),
              loserHand: handName(
                (state.lastSideShow.loserIdx === state.lastSideShow.requester.playerIdx
                  ? state.lastSideShow.requester
                  : state.lastSideShow.target
                ).handName,
              ),
            })}
          </div>
          <div className="flex flex-wrap gap-4">
            {[state.lastSideShow.requester, state.lastSideShow.target].map((h) => (
              <div key={h.playerIdx}>
                <div className="text-ds-text-muted text-xs mb-0.5">
                  {playerLabel(h.playerIdx, h.playerIdx === humanIdx)}
                  {h.handName ? ` — ${handName(h.handName)}` : ''}
                </div>
                <div className="flex gap-1">
                  {h.cards.map((c, i) => (
                    <CardImage key={`ss-${h.playerIdx}-${i}`} card={c} width={view.cardWidth} />
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Revealed hands at showdown */}
      {state.isShowdown && (
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
                    <CardImage key={`${p.id}-${i}`} card={c} width={view.cardWidth} />
                  ))}
                </div>
              </div>
            ))}
        </div>
      )}

      {/* Deal result */}
      {(isRoundEnd || isGameEnd) && state.roundWinnerIdx >= 0 && (
        <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
          <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
          <div>
            {t('roundResult.winner', {
              name: playerLabel(state.roundWinnerIdx, state.roundWinnerIdx === humanIdx),
              pot: state.pot,
            })}
          </div>
        </div>
      )}

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ state, t, view }: TeenPattiContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  return humanPlayer && (humanPlayer.seen || state.isShowdown) && humanPlayer.cards.length > 0 ? (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={[]}
      toggleCard={noopToggle}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="teenpatti"
    />
  ) : (
    <div className="mb-2" data-tutorial="teenpatti-player-hand">
      <div className="text-ds-text-muted text-sm">
        {t('handLabel')} — {t('badge.blind')}
      </div>
      {humanPlayer && !humanPlayer.seen && !state.isShowdown && humanPlayer.cardCount > 0 && (
        <fieldset aria-label={t('blindHandDescription')} className="border-0 p-0 m-0 min-w-0 flex gap-2 mt-2">
          {Array.from({ length: humanPlayer.cardCount }, (_, i) => (
            <CardBack key={i} width={view.cardWidth} />
          ))}
        </fieldset>
      )}
    </div>
  );
}

function FooterStatus({ t, frontendHint }: TeenPattiContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

function FooterActions({ game, state, t, busy, view }: TeenPattiContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBettingPhase = state.phase === TeenPattiPhase.BETTING;
  const isRoundEnd = state.phase === TeenPattiPhase.ROUND_END;
  const isGameEnd = state.phase === TeenPattiPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const isHumanBetTurn = isBettingPhase && isHumanTurn;
  const humanCallCost = state.humanCallCost;
  const { raiseStake, setRaiseStake } = view;

  return (
    <>
      {isHumanBetTurn && (
        <>
          <div className="w-full text-ds-text-muted text-sm">
            {t(humanPlayer?.seen ? 'callCost.seen' : 'callCost.blind', { amount: humanCallCost })}
          </div>
          {!humanPlayer?.seen && (
            <button type="button" className={btnSecondary} onClick={game.handleSee} disabled={busy}>
              {t('seeButton')}
            </button>
          )}
          <button type="button" className={btnPrimary} onClick={game.handleBet} disabled={busy}>
            {t('betButton', { amount: humanCallCost })}
          </button>
          <div className="flex items-center gap-1">
            <button
              type="button"
              className={btnSecondary}
              onClick={() => setRaiseStake((a) => Math.max(state.minRaise, a - 1))}
              disabled={busy || raiseStake <= state.minRaise}
              aria-label={t('raiseDecrease')}
            >
              −
            </button>
            <span className="text-ds-text-primary text-sm min-w-[4rem] text-center" aria-live="polite">
              {t('raisePrompt')} {raiseStake}
            </span>
            <button
              type="button"
              className={btnSecondary}
              // **上限はサーバーが送る値でクランプする (#4729)。**以前は
              // 上限が無く、払えない額を送信できてサーバーエラーで初めて
              // 気づいた。
              onClick={() => setRaiseStake((a) => Math.min(state.maxRaise, a + 1))}
              disabled={busy || raiseStake >= state.maxRaise}
              aria-label={t('raiseIncrease')}
            >
              ＋
            </button>
            <button
              type="button"
              className={btnWarning}
              onClick={() => game.handleRaise(raiseStake)}
              disabled={busy || !state.canRaise}
              title={state.canRaise ? undefined : t('raiseUnavailable')}
              data-testid="tp-raise-button"
            >
              {t('raiseButton', { amount: raiseStake })}
            </button>
            {/* **範囲を出さないと、+/- を連打してボタンが無効になる位置を
                探るしかない (#5660)。** 姉妹の Three Card Brag は
                tcb-raise-range で常時出しており、CUI も promptRaiseRange で
                min/max を出している。 */}
            <span className="text-ds-text-muted text-xs ml-1 whitespace-nowrap" data-testid="tp-raise-range">
              {state.canRaise ? t('raiseRange', { min: state.minRaise, max: state.maxRaise }) : t('raiseUnavailable')}
            </span>
          </div>
          <button type="button" className={btnDanger} onClick={game.handleFold} disabled={busy}>
            {t('foldButton')}
          </button>
          {state.canShow && (
            <button type="button" className={btnPrimary} onClick={game.handleShow} disabled={busy}>
              {t('showButton')}
            </button>
          )}
          {state.canRequestSideShow && (
            <button type="button" className={btnSecondary} onClick={game.handleSideShow} disabled={busy}>
              {t('sideShowButton')}
            </button>
          )}
        </>
      )}

      {isRoundEnd && !isGameEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const teenPattiDefinition: GameDefinition<
  TeenPattiResponse,
  TeenPattiGame,
  Parameters<typeof teenPattiApi.exec>,
  TeenPattiView
> = {
  key: 'teenpatti',
  useGame: useTeenPattiGame,
  useView: useTeenPattiView,
  tutorial: TEEN_PATTI_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 3 },
  cli: () => ({
    parseCommand: parseTeenPattiCommand,
    formatResponse: formatTeenPattiState,
    helpText: TEEN_PATTI_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === TeenPattiPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.matchWinnerIdx === humanIdx,
    };
  },
  shell: ({ state }) => {
    const isGameEnd = state.phase === TeenPattiPhase.GAME_END || state.gameEndFlag;
    return {
      isHumanTurn: state.isHumanTurn && !isGameEnd,
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
          id: 'ante',
          label: t('settings.ante'),
          value: game.teenPattiConfig.ante,
          options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('ante', v),
        },
        {
          type: 'select',
          id: 'startingChips',
          label: t('settings.startingChips'),
          value: game.teenPattiConfig.startingChips,
          options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('startingChips', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'teenpatti-action-buttons',
  resetTutorialId: 'teenpatti-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Teen Patti game page: a 4-player Indian chips/pot vying game with Side Show. */
export const TeenPattiPage = createGamePage(teenPattiDefinition);
