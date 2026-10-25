import type { TFunction } from 'i18next';
import { useCallback, useState } from 'react';
import { karnoffelApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { KarnoffelResponse } from '../types/card';
import { KarnoffelPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { KARNOFFEL_HELP, parseKarnoffelCommand } from '../utils/cli/commands/karnoffelCommands';
import { formatKarnoffelState } from '../utils/cli/formatters/karnoffelFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import { karnoffelRankKey } from '../utils/karnoffelRanks';

/** Karnöffel tutorial step definitions. */
const KARNOFFEL_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="karnoffel-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="karnoffel-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="karnoffel-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="karnoffel-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="karnoffel-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KARNOFFEL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KarnoffelPhase.PLAY]: 'play',
  [KarnoffelPhase.HAND_END]: 'handEnd',
  [KarnoffelPhase.GAME_END]: 'gameEnd',
};

type KarnoffelGame = ReturnType<typeof useKarnoffelGame>;
type KarnoffelView = ReturnType<typeof useKarnoffelView>;
type KarnoffelContext = GamePageContext<KarnoffelResponse, KarnoffelGame, KarnoffelView>;

// 人間は席 0 = チーム 0。勝敗はチームで判定する。
function isGameEnded(state: KarnoffelResponse): boolean {
  return state.phase === KarnoffelPhase.GAME_END || state.gameEndFlag;
}

function humanWon(state: KarnoffelResponse): boolean {
  return isGameEnded(state) && state.winnerTeam === 0;
}

function useKarnoffelGame() {
  const { state, loading, error, exec, retry } = useGameApi(karnoffelApi.exec);

  const [selected, setSelected] = useState<number | null>(null);
  const [targetHands, setTargetHands] = useState(3);

  useMountReset(exec);

  const handlePlay = useCallback(() => {
    if (selected === null) return;
    void exec('play', { cardIndex: selected });
    setSelected(null);
  }, [exec, selected]);

  const handleReset = useCallback(() => {
    setSelected(null);
    void exec('reset', { config: { targetHands } });
  }, [exec, targetHands]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    selected,
    setSelected,
    targetHands,
    setTargetHands,
    handlePlay,
    handleReset,
  };
}

function useKarnoffelView() {
  const { cardWidth, isMobile } = useCardDimensions();
  const breakpoint = useBreakpoint();
  const phaseNames = usePhaseNames('karnoffel', KARNOFFEL_PHASE_KEYS);
  return { cardWidth, isMobile, breakpoint, phaseNames };
}

const playerLabel = (t: TFunction, id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
const suitLabel = (t: TFunction, s: number): string => t(`suitName.${s}`);
const rankForPosition = (baseRankKey: string | null, isLead: boolean) =>
  baseRankKey === 'devil' ? (isLead ? 'devilLead' : 'devilFollow') : baseRankKey;

function Board({ state, game, t, tc, view }: KarnoffelContext) {
  const { cardWidth, isMobile, breakpoint } = view;
  const isHandEnd = state.phase === KarnoffelPhase.HAND_END;
  const isGameEnd = isGameEnded(state);
  const result = state.lastResult;

  const renderPlayer = (p: (typeof state.players)[number]) => (
    <div
      key={`player-${p.id}`}
      data-testid="karnoffel-player"
      className={`grid grid-cols-[1fr_auto] items-center gap-x-1 rounded bg-black/30 p-1 text-xs ${
        p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
      } ${p.isHuman ? 'font-semibold' : ''}`}
    >
      <span className="col-span-2 flex min-w-0 items-center gap-1 truncate">
        <span className="truncate">{playerLabel(t, p.id, p.isHuman)}</span>
        <span>({t('team', { n: p.team })})</span>
        {p.isDealer && <span className="text-ds-accent">[{t('dealer')}]</span>}
      </span>
      <span className="col-span-2 row-start-2 flex items-center gap-1">
        <span className="flex items-center gap-1">
          {t('upCard')}:
          {p.upCard ? (
            (() => {
              const upKey = karnoffelRankKey(p.upCard, state.chosenSuit);
              return (
                <span className="relative inline-block">
                  <CardImage card={p.upCard} width={cardWidth} />
                  {upKey && (
                    <span
                      data-testid={`karnoffel-up-rank-${p.id}`}
                      className={`absolute left-0 right-0 bottom-0 rounded-b px-0.5 text-[9px] font-bold text-center truncate ${badgeWarningColors}`}
                    >
                      {t(`rankBadge.${upKey}`)}
                    </span>
                  )}
                </span>
              );
            })()
          ) : (
            <span>-</span>
          )}
        </span>
        <span className="ml-auto">{t('tricksWon', { n: p.tricksWon })}</span>
        {!p.isHuman && p.cards.length === 0 && (
          <span className="whitespace-nowrap">{t('hiddenHand', { count: p.cardCount })}</span>
        )}
      </span>
    </div>
  );

  const renderTrick = () =>
    state.trick.length > 0 && (
      <div
        className={`mb-2 flex ${breakpoint === 'largeDesktop' ? 'flex-1' : ''} items-center gap-2`}
        data-testid="karnoffel-trick"
      >
        <span className="text-ds-text-muted text-sm">{t('trick')}</span>
        {state.trick.map((c, i) => {
          const seatIdx = (state.trickLeaderIdx + i) % state.players.length;
          const isHuman = state.players.some((p) => p.id === seatIdx && p.isHuman);
          const name = t('seatPlayer', { idx: seatIdx, name: playerLabel(t, seatIdx, isHuman) });
          const baseRankKey = karnoffelRankKey(c, state.chosenSuit);
          const rankKey = rankForPosition(baseRankKey, i === 0);
          const accessibleName = rankKey
            ? `${name}${t('listSeparator')}${cardAlt(c)}${t('listSeparator')}${t(`rankBadge.${rankKey}`)}`
            : `${name}${t('listSeparator')}${cardAlt(c)}`;
          return (
            <div key={`trick-${c.design}-${c.value}-${i}`} className="flex flex-col items-center">
              <span className="mb-1 text-ds-text-muted text-xs" aria-hidden="true">
                {name}
              </span>
              <span className="relative inline-block">
                <CardImage card={c} width={cardWidth} ariaLabel={accessibleName} />
                {rankKey && (
                  <span
                    aria-hidden="true"
                    data-testid={`karnoffel-trick-rank-${i}`}
                    className={`absolute left-0 right-0 bottom-0 rounded-b px-0.5 text-[9px] font-bold text-center truncate ${badgeWarningColors}`}
                  >
                    {t(`rankBadge.${rankKey}`)}
                  </span>
                )}
              </span>
            </div>
          );
        })}
      </div>
    );

  const cpuAccordion = (
    <CpuAccordion playerCount={state.players.filter((player) => !player.isHuman).length}>
      {/* CPUの表向き札は全員ぶん見える。切札の根拠がここにある。 */}
      <div className="grid grid-cols-2 gap-1 sm:grid-cols-4">
        {state.players.filter((player) => !player.isHuman).map(renderPlayer)}
      </div>
    </CpuAccordion>
  );

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="karnoffel-info">
        <span className="mr-4">{t('hand', { n: state.handNumber })}</span>
        <span className="mr-4" data-testid="karnoffel-chosen">
          {t('chosenSuit')}: {suitLabel(t, state.chosenSuit)}
        </span>
      </div>

      <details className="mb-2 text-center text-ds-text-primary text-sm">
        <summary className="cursor-pointer select-none" data-testid="karnoffel-settings-summary">
          {tc('settings.title')}
        </summary>
        <div className="flex justify-center">
          <label className="flex min-h-[44px] items-center gap-2 text-sm text-ds-text-primary">
            <span>{t('targetHandsLabel')}</span>
            <select
              aria-label={t('targetHandsLabel')}
              className="min-h-[44px] rounded border border-ds-border bg-ds-surface px-3 text-ds-text-primary"
              value={game.targetHands}
              onChange={(event) => game.setTargetHands(Number(event.target.value))}
            >
              {Array.from({ length: 10 }, (_, index) => index + 1).map((count) => (
                <option key={count} value={count}>
                  {count}
                </option>
              ))}
            </select>
          </label>
        </div>
      </details>

      {/* The lowest face-up card decides the suit — not a turn-up. */}
      <details className="mb-2 text-center text-ds-text-muted text-xs">
        <summary className="cursor-pointer select-none" data-testid="karnoffel-chosen-note-summary">
          {t('chosenNoteSummary')}
        </summary>
        <div data-testid="karnoffel-chosen-note">{t('chosenNote')}</div>
      </details>

      {/* The irregular ranking, which is the whole point of the game. */}
      <details className="mb-2 p-2 rounded bg-black/20 text-xs">
        <summary className="cursor-pointer select-none text-ds-text-primary" data-testid="karnoffel-ladder-summary">
          {t('ladderTitle')}
        </summary>
        <div data-testid="karnoffel-ladder">
          <div className="text-ds-text-primary">{t('ladderLine')}</div>
          <div className="mt-1 text-ds-text-muted">{t('ladderNote')}</div>
        </div>
      </details>

      {/* Score sheet */}
      <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="karnoffel-scores">
        <div>
          {t('team', { n: 0 })}: {t('handsWon', { n: state.handsWon[0] })} ({t('tricksWon', { n: state.teamTricks[0] })}
          ) / {t('team', { n: 1 })}: {t('handsWon', { n: state.handsWon[1] })} (
          {t('tricksWon', { n: state.teamTricks[1] })})
        </div>
        <div className="text-xs text-ds-text-muted">
          {t('targetNote', { n: state.targetHands, need: state.tricksToWin })}
        </div>
      </div>

      {/* Players */}
      <div className="mb-2" data-testid="karnoffel-players" data-tutorial="karnoffel-players">
        <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
        {breakpoint === 'largeDesktop' ? (
          <>
            <div className="flex flex-wrap items-start gap-4" data-testid="karnoffel-seat-trick-row">
              <div className="grid grid-cols-1 gap-1" data-testid="karnoffel-human-player">
                {state.players.filter((player) => player.isHuman).map(renderPlayer)}
              </div>
              {renderTrick()}
            </div>
            {cpuAccordion}
          </>
        ) : (
          <>
            <div className="grid grid-cols-2 gap-1 sm:grid-cols-4" data-testid="karnoffel-human-player">
              {state.players.filter((player) => player.isHuman).map(renderPlayer)}
            </div>
            {cpuAccordion}
            {renderTrick()}
          </>
        )}
      </div>

      {/* Hand result */}
      {(isHandEnd || isGameEnd) && result && (
        <div
          className={`mb-2 p-2 rounded text-sm ${badgeWarningColors}`}
          data-testid="karnoffel-result"
          role="status"
          aria-live="polite"
        >
          <div className="mb-1">{t('resultTitle')}</div>
          <div>
            {result.winnerTeam >= 0
              ? t('handWonLine', { team: result.winnerTeam, t0: result.tricks[0], t1: result.tricks[1] })
              : t('handDrawnLine', { t0: result.tricks[0], t1: result.tricks[1] })}
          </div>
        </div>
      )}
      {isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}
    </>
  );
}

function FooterHand({ state, game, t, busy, view }: KarnoffelContext) {
  const { cardWidth } = view;
  const human = state.players.find((p) => p.isHuman);
  // Hoisted so a seat-less state exercises the empty case rather than leaving it
  // as an unreachable fallback inside the hand render.
  const humanCards = human?.cards ?? [];
  const isPlay = state.phase === KarnoffelPhase.PLAY;
  const canPlay = (i: number) => state.validPlays.includes(i);

  return (
    <div className="mb-2" data-tutorial="karnoffel-hand">
      <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
      <div className="flex flex-wrap gap-1">
        {humanCards.map((c, i) => {
          // The ladder text names the titled cards, but which card in hand
          // holds a title depends on the suit chosen this deal (#4773).
          const rankKey = rankForPosition(karnoffelRankKey(c, state.chosenSuit), state.trick.length === 0);
          const cardName = rankKey ? `${cardAlt(c)} (${t(`rankBadge.${rankKey}`)})` : cardAlt(c);
          const accessibleName = isPlay
            ? `${cardName}, ${t(canPlay(i) ? 'playableThisTurn' : 'notPlayableThisTurn')}`
            : cardName;
          return (
            <button
              key={`hand-${c.design}-${c.value}-${i}`}
              type="button"
              data-hint-action="play"
              onClick={() => game.setSelected(i)}
              disabled={busy || (isPlay && !canPlay(i))}
              aria-pressed={game.selected === i}
              aria-label={accessibleName}
              className={`relative rounded ${game.selected === i ? 'ring-2 ring-ds-accent' : ''} ${
                isPlay && !canPlay(i) ? 'opacity-40' : ''
              }`}
            >
              <CardImage card={c} width={cardWidth} />
              {rankKey && (
                <span
                  aria-hidden="true"
                  data-testid={`karnoffel-rank-${rankKey}`}
                  className={`absolute left-0 right-0 bottom-0 rounded-b px-0.5 text-[9px] font-bold text-center truncate ${badgeWarningColors}`}
                >
                  {t(`rankBadge.${rankKey}`)}
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}

function FooterStatus({ t, tc, frontendHint }: KarnoffelContext) {
  return (
    <>
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

function FooterActions({ state, game, t, busy }: KarnoffelContext) {
  const isPlay = state.phase === KarnoffelPhase.PLAY;
  const isHandEnd = state.phase === KarnoffelPhase.HAND_END;
  const isGameEnd = isGameEnded(state);
  const didHumanWin = humanWon(state);
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  return (
    <>
      {isHumanPlay && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handlePlay}
          disabled={busy || game.selected === null}
        >
          {t('playButton')}
        </button>
      )}

      {isHandEnd && !isGameEnd && (
        <button type="button" className={btnSuccess} onClick={() => void game.exec('next')} disabled={busy}>
          {t('nextHand')}
        </button>
      )}

      {isGameEnd && (
        <span className="text-ds-text-primary text-sm font-semibold mr-1">{didHumanWin ? t('win') : t('lose')}</span>
      )}
    </>
  );
}

const karnoffelDefinition: GameDefinition<
  KarnoffelResponse,
  KarnoffelGame,
  Parameters<typeof karnoffelApi.exec>,
  KarnoffelView
> = {
  key: 'karnoffel',
  useGame: useKarnoffelGame,
  useView: useKarnoffelView,
  tutorial: KARNOFFEL_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: ({ frontendHint }) => ({
    parseCommand: parseKarnoffelCommand,
    formatResponse: formatKarnoffelState,
    helpText: KARNOFFEL_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => {
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnded(state),
      won: humanWon(state),
    };
  },
  shell: ({ state }) => {
    const isPlay = state.phase === KarnoffelPhase.PLAY;
    const isGameEnd = isGameEnded(state);
    const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;
    return {
      isHumanTurn: isHumanPlay,
    };
  },
  settingsPlacement: 'none',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) =>
    !view.isMobile ? (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ) : null,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'karnoffel-actions',
  resetTutorialId: 'karnoffel-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Karnöffel game page: the oldest named card game, with its irregular ranking. */
export const KarnoffelPage = createGamePage(karnoffelDefinition);
