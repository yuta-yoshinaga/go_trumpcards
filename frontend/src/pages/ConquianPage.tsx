import type { TFunction } from 'i18next';
import type { conquianApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { TARGET_WINS_OPTIONS, useConquianGame } from '../hooks/useConquianGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { ConquianPlayerData, ConquianResponse } from '../types/card';
import { ConquianPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CONQUIAN_HELP, parseConquianCommand } from '../utils/cli/commands/conquianCommands';
import { formatConquianState } from '../utils/cli/formatters/conquianFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Total cards a player must lay in table melds to win a Conquian round. */
const CONQUIAN_MELD_TARGET = 11;

const CONQUIAN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ConquianPhase.DRAW]: 'draw',
  [ConquianPhase.MELD]: 'meld',
  [ConquianPhase.ROUND_END]: 'roundEnd',
  [ConquianPhase.GAME_END]: 'gameEnd',
};

/** Conquian tutorial step definitions. */
const CQ_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="cq-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="cq-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cq-meld-button"]',
    messageKey: 'tutorial.meldButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cq-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cq-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cq-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

function useGame() {
  const game = useConquianGame();
  return {
    ...game,
    exec: game.gameExec,
  };
}

type ConquianGame = ReturnType<typeof useGame>;
type ConquianView = ReturnType<typeof useConquianView>;
type ConquianContext = GamePageContext<ConquianResponse, ConquianGame, ConquianView>;

function useConquianView({ game }: { game: ConquianGame }) {
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('conquian', CONQUIAN_PHASE_KEYS);

  // Card keyboard navigation is wired here rather than through the factory's `cardNav`, because
  // `cardNav` also renders a card-shortcuts panel that the hand-wired page never had.
  const state = game.state;
  const isMeldPhase = state?.phase === ConquianPhase.MELD;
  const isHumanTurn =
    (state?.phase === ConquianPhase.DRAW || isMeldPhase) && state?.players[state.currentPlayerIdx]?.isHuman === true;
  useCardKeyboardNav({
    cardCount: state?.players.find((p) => p.isHuman)?.cards?.length ?? 0,
    onToggle: game.toggleCard,
    onConfirm: () => {
      if (isMeldPhase) game.handleMeldSelected();
    },
    onClear: game.clearSelection,
    enabled: isHumanTurn && !game.loading,
  });

  return { cardWidth, phaseNames };
}

// Meld-progress indicator: count the cards laid across a player's table melds
// toward the 11-card win condition, highlighting the final stretch (<=2 left).
function renderMeldProgress(t: TFunction, p: ConquianPlayerData, testId: string) {
  const melded = p.melds.reduce((sum, m) => sum + m.cards.length, 0);
  const remaining = CONQUIAN_MELD_TARGET - melded;
  const isClose = remaining > 0 && remaining <= 2;
  const pct = Math.min(100, (melded / CONQUIAN_MELD_TARGET) * 100);
  const label = t('meldProgress', { count: melded, total: CONQUIAN_MELD_TARGET });
  return (
    <div data-testid={testId} className="mt-1">
      <div className="flex items-center justify-between text-xs mb-0.5">
        <span className="text-ds-text-muted">{label}</span>
        {isClose && <span className="text-ds-warning font-bold">{t('meldRemaining', { count: remaining })}</span>}
      </div>
      <div
        role="progressbar"
        aria-label={label}
        aria-valuemin={0}
        aria-valuemax={CONQUIAN_MELD_TARGET}
        aria-valuenow={melded}
        className="h-1.5 w-full rounded-sm bg-white/15 overflow-hidden"
      >
        <div
          className={`h-full rounded-sm ${isClose ? 'bg-ds-warning' : 'bg-ds-accent'}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}

function Board({ state, game, t, view, frontendHint }: ConquianContext) {
  const isDrawPhase = state.phase === ConquianPhase.DRAW;
  const isMeldPhase = state.phase === ConquianPhase.MELD;
  const isRoundEnd = state.phase === ConquianPhase.ROUND_END;
  const isGameEnd = state.phase === ConquianPhase.GAME_END || !!state.gameEndFlag;
  const isHumanTurn = (isDrawPhase || isMeldPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span>{t('drawPile', { count: state.drawPileCount })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />

          {/* Discard pile top */}
          {state.discardTop && (
            <div
              className="my-3 p-3 rounded bg-black/40 flex items-center gap-3"
              data-tutorial="cq-draw-area"
              data-testid="cq-discard-pile"
            >
              <AnimatedCard card={state.discardTop} width={view.cardWidth} />
              <div className="text-ds-text-muted text-sm">{t('discardTop')}</div>
            </div>
          )}

          {/* Per-player table melds */}
          {state.players.map((p) =>
            p.melds.length === 0 ? null : (
              <div key={`melds-${p.id}`} className="my-2 p-2 rounded bg-black/30">
                <div className="text-ds-text-muted text-sm mb-1">
                  {playerName(p.id, p.isHuman)} - {t('tableMelds')}
                </div>
                {p.melds.map((m, mi) => {
                  // 自分のメルドだけがレイオフ先。1 枚選んでいるときに押すと、
                  // そのメルドへ足す (延長先が複数ある場面で選べるようにする — #4837)。
                  // **押せる先だけを押せるように。**選んだ札が実際に足せる
                  // メルドだけをレイオフ先にする。バックエンドは足せない
                  // 指定を黙って別のメルドへ落とすので、見た目と結果が
                  // 食い違ってしまう (#4837)。
                  const selectedIdx = game.selectedCardIndices.length === 1 ? game.selectedCardIndices[0] : -1;
                  const canLayoff =
                    p.isHuman &&
                    isMeldPhase &&
                    isHumanTurn &&
                    selectedIdx >= 0 &&
                    (state.layoffTargets?.[selectedIdx]?.includes(mi) ?? false);
                  return (
                    <button
                      key={`meld-${p.id}-${mi}`}
                      type="button"
                      disabled={!canLayoff}
                      aria-label={
                        canLayoff
                          ? t('layoffToMeld', {
                              n: mi + 1,
                              cards: m.cards.map(cardAlt).join(t('listSeparator')),
                            })
                          : undefined
                      }
                      data-layoff-target={canLayoff ? mi : undefined}
                      onClick={() => canLayoff && game.handleMeldSelected(mi)}
                      className={[
                        'flex flex-wrap gap-1 mb-1 rounded p-0.5',
                        canLayoff ? 'ring-1 ring-ds-accent cursor-pointer' : 'cursor-default',
                      ].join(' ')}
                    >
                      {m.cards.map((card, ci) => (
                        <AnimatedCard key={`meld-${p.id}-${mi}-${ci}`} card={card} width={view.cardWidth * 0.6} />
                      ))}
                    </button>
                  );
                })}
              </div>
            ),
          )}
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* CPU player */}
          {state.players
            .filter((p) => !p.isHuman)
            .map((p) => (
              <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                <div className="text-ds-text-muted text-sm">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('wins', { count: p.wins })}
                </div>
                {renderMeldProgress(t, p, `conquian-meld-progress-cpu-${p.id}`)}
                {/* Reveal CPU cards at round/game end */}
                {(isRoundEnd || isGameEnd) && p.cards.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {p.cards.map((card, idx) => (
                      <AnimatedCard
                        key={`cpu-${card.design}-${card.value}-${idx}`}
                        card={card}
                        width={view.cardWidth * 0.8}
                      />
                    ))}
                  </div>
                )}
              </div>
            ))}

          {/* Wins table */}
          <div className="my-3 p-2 rounded bg-black/30" data-tutorial="cq-score-table">
            <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
            <table className="w-full text-sm text-ds-text-muted">
              <thead>
                <tr>
                  <th scope="col" className="text-left">
                    {t('scoresPlayer')}
                  </th>
                  <th scope="col">{t('scoresWins')}</th>
                </tr>
              </thead>
              <tbody>
                {state.players.map((p) => (
                  <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                    <td>{playerName(p.id, p.isHuman)}</td>
                    <td className="text-center">{p.wins}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* `roundWinnerIdx` は毎レスポンスに乗っているのに一度も読まれておらず、
          誰が制したかは累計勝利数の表を前後で見比べるしかなかった。
          -1 は山札切れの引き分け。
          **文言は必ず「このラウンド」に限定する。**`endMatchOnDraw` は山札切れで
          引き分けた局でも累積勝利数から `matchWinnerIdx` を決めるので、
          `targetWins > 1` なら `roundWinnerIdx === -1` かつ `winnerIdx >= 0` に
          なりうる。ここが「引き分けです」とだけ言うと、真下の GameMessageBox が
          告げるマッチ勝者と真っ向から食い違う (レビュー指摘)。 */}
      {(isRoundEnd || isGameEnd) && (
        <div className="text-center text-ds-text-primary text-sm mb-1" data-testid="conquian-round-winner">
          {state.roundWinnerIdx < 0
            ? t('roundWinnerDraw')
            : t('roundWinner', { name: findPlayerName(state.players, state.roundWinnerIdx) })}
        </div>
      )}

      {state.roundHistory.length > 0 && (
        <section className="my-2 text-sm" aria-label={t('roundHistory')} data-testid="conquian-round-history">
          <h3 className="text-ds-text-muted mb-1">{t('roundHistory')}</h3>
          <ol className="text-ds-text-primary">
            {state.roundHistory.map((winnerIdx, i) => (
              <li key={i}>
                {t('roundHistoryEntry', {
                  round: i + 1,
                  result:
                    winnerIdx < 0
                      ? t('roundWinnerDraw')
                      : t('roundWinner', { name: findPlayerName(state.players, winnerIdx) }),
                })}
              </li>
            ))}
          </ol>
        </section>
      )}

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ state, game, t, view }: ConquianContext) {
  const isDrawPhase = state.phase === ConquianPhase.DRAW;
  const isMeldPhase = state.phase === ConquianPhase.MELD;
  const isHumanTurn = (isDrawPhase || isMeldPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
  const humanPlayer = state.players.find((p) => p.isHuman);

  return (
    <>
      <div
        className="text-xs font-bold mb-1 text-ds-info"
        data-testid="conquian-forced-use"
        role="status"
        aria-live="polite"
      >
        {isMeldPhase && isHumanTurn && state.tookDiscard && t('forcedUse')}
      </div>
      {humanPlayer && (
        <div className="mb-2 max-w-xs">{renderMeldProgress(t, humanPlayer, 'conquian-meld-progress')}</div>
      )}
      {humanPlayer && (
        <div className="flex flex-wrap gap-1 mb-2" data-tutorial="cq-player-hand">
          {humanPlayer.cards.map((card, idx) => (
            <button
              type="button"
              key={`${card.design}-${card.value}-${idx}`}
              onClick={() => game.toggleCard(idx)}
              aria-label={cardAlt(card)}
              aria-pressed={game.selectedCardIndices.includes(idx)}
              className={`transition-transform ${focusRingCard}`}
              style={{
                background: 'none',
                padding: 0,
                borderRadius: 8,
                ...selectedCardStyle(game.selectedCardIndices.includes(idx)),
                boxSizing: 'border-box',
              }}
            >
              <AnimatedCard card={card} width={view.cardWidth} />
            </button>
          ))}
        </div>
      )}
    </>
  );
}

function primaryActions({ game, state, t, busy }: ConquianContext) {
  const isDrawPhase = state.phase === ConquianPhase.DRAW;
  const isMeldPhase = state.phase === ConquianPhase.MELD;
  const isRoundEnd = state.phase === ConquianPhase.ROUND_END;
  const isHumanTurn = (isDrawPhase || isMeldPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  if (!((isDrawPhase && isHumanTurn) || (isMeldPhase && isHumanTurn) || isRoundEnd)) {
    return null;
  }

  const selectedCardCanLayoff =
    game.selectedCardIndices.length === 1 &&
    state.layoffTargets.some((targets, cardIdx) => cardIdx === game.selectedCardIndices[0] && targets.length > 0);

  return (
    <div className="flex gap-2 items-center flex-wrap">
      {isDrawPhase && isHumanTurn && (
        <div className="flex gap-2">
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleDrawStock}
            disabled={busy}
            aria-keyshortcuts="s"
          >
            {t('drawStockButton')}
            <KbdBadge label={t('kbd.stock')} />
          </button>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleDrawDiscard}
            disabled={busy || !state.discardTop}
            aria-keyshortcuts="d"
          >
            {t('drawDiscardButton')}
            <KbdBadge label={t('kbd.discard')} />
          </button>
        </div>
      )}
      {isMeldPhase && isHumanTurn && (
        <>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => game.handleMeldSelected()}
            disabled={busy || game.selectedCardIndices.length < 3}
            data-tutorial="cq-meld-button"
            data-testid="conquian-meld-button"
          >
            {t('meldButton')}
          </button>
          <button
            type="button"
            className={`${btnPrimary} ${selectedCardCanLayoff && !busy ? '' : 'opacity-70 cursor-not-allowed saturate-50'}`}
            onClick={() => {
              if (!busy && selectedCardCanLayoff) game.handleMeldSelected();
            }}
            aria-disabled={!selectedCardCanLayoff || busy}
            data-testid="conquian-layoff-button"
          >
            {t('layoffButton')}
          </button>
          <button
            type="button"
            className={`${btnPrimary} ${busy || game.selectedCardIndices.length !== 1 || state.tookDiscard ? 'opacity-70 cursor-not-allowed saturate-50' : ''}`}
            onClick={() => {
              if (!busy && game.selectedCardIndices.length === 1 && !state.tookDiscard) game.handleDiscard();
            }}
            aria-disabled={busy || game.selectedCardIndices.length !== 1 || state.tookDiscard || undefined}
            data-tutorial="cq-discard-button"
          >
            {t('discardButton')}
          </button>
        </>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const conquianDefinition: GameDefinition<
  ConquianResponse,
  ConquianGame,
  Parameters<typeof conquianApi.exec>,
  ConquianView
> = {
  key: 'conquian',
  useGame,
  useView: useConquianView,
  tutorial: CQ_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 10 },
  cli: ({ frontendHint }) => ({
    parseCommand: parseConquianCommand,
    formatResponse: formatConquianState,
    helpText: CONQUIAN_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === ConquianPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isDrawPhase = state.phase === ConquianPhase.DRAW;
    const isMeldPhase = state.phase === ConquianPhase.MELD;
    const isHumanTurn = (isDrawPhase || isMeldPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
    return { isHumanTurn };
  },
  reset: ({ game }) => {
    void game.gameExec('reset', undefined, {
      targetWins: game.conquianConfig.targetWins,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'targetWins',
          label: t('settings.targetWins'),
          value: game.conquianConfig.targetWins,
          options: TARGET_WINS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetWins', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  controlsInFooter: true,
  resetTutorialId: 'cq-reset-button',
  actionsClassName: '',
  // DRAW phase: s draws from the stock, d takes the discard top. Number keys
  // and Enter/Escape are handled by useCardKeyboardNav (in useConquianView); these letter keys don't
  // collide with it.
  shortcuts: ({ game, state }) => {
    const isDrawPhase = state.phase === ConquianPhase.DRAW;
    const isHumanTurn = isDrawPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
    return [
      { key: 's', action: game.handleDrawStock, enabled: isHumanTurn },
      { key: 'd', action: game.handleDrawDiscard, enabled: isHumanTurn && !!state.discardTop },
    ];
  },

  Board,
  FooterHand,
};

/** Renders the Conquian game page with draw and meld phases and per-player table melds. */
export const ConquianPage = createGamePage(conquianDefinition);
