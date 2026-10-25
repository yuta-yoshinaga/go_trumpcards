import type { TFunction } from 'i18next';
import { useEffect } from 'react';
import type { scartoApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, SCARTO_DISCARD_COUNT, useScartoGame } from '../hooks/useScartoGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { ScartoResponse } from '../types/card';
import { ScartoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseScartoCommand, SCARTO_HELP } from '../utils/cli/commands/scartoCommands';
import { formatScartoState } from '../utils/cli/formatters/scartoFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { scartoUndiscardableReason } from '../utils/scartoDiscard';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Scarto tutorial step definitions. */
const SCARTO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="scarto-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scarto-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scarto-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="scarto-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="scarto-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SCARTO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ScartoPhase.SCARTO]: 'scarto',
  [ScartoPhase.PLAY]: 'play',
  [ScartoPhase.TRICK_END]: 'trickEnd',
  [ScartoPhase.ROUND_END]: 'roundEnd',
  [ScartoPhase.GAME_END]: 'gameEnd',
};

/** Outcome i18n keys indexed by outcome value (0=none/average, 1=above average, 2=below average). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** Formats a card-point number with up to one decimal place, dropping a trailing `.0`. */
function formatPoints(n: number): string {
  return Number.isInteger(n) ? String(n) : n.toFixed(1);
}

/** Formats an average difference with the page's one-decimal precision. */
function formatSigned(n: number): string {
  const value = formatPoints(n);
  return n > 0 ? `+${value}` : value;
}

type ScartoGame = ReturnType<typeof useScartoGame>;
type ScartoView = ReturnType<typeof useScartoView>;
type ScartoContext = GamePageContext<ScartoResponse, ScartoGame, ScartoView>;

function useScartoView({ game }: { game: ScartoGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('scarto', SCARTO_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

/** Renders the average, formula, and each player's earned-points breakdown. */
function ScartoBreakdown({
  players,
  average,
  formulaTestId,
  t,
}: {
  players: ScartoResponse['players'];
  average: number;
  formulaTestId?: string;
  t: TFunction;
}) {
  return (
    <>
      <div>{t('roundResult.average', { avg: formatPoints(average) })}</div>
      <div data-testid={formulaTestId}>{t('roundResult.formula', { n: players.length })}</div>
      {players.map((p) => (
        <div key={p.id}>
          {t('roundResult.earnedLine', {
            name: playerName(p.id, p.isHuman),
            points: p.cardPoints,
            diff: formatSigned(p.cardPoints - average),
            scaled: formatSigned((p.cardPoints - average) * players.length),
            outcome: t(OUTCOME_KEYS[p.cardPoints > average ? 1 : p.cardPoints < average ? 2 : 0]),
          })}
        </div>
      ))}
    </>
  );
}

function Board({ state, t, view }: ScartoContext) {
  const isScartoPhase = state.phase === ScartoPhase.SCARTO;
  const isPlayPhase = state.phase === ScartoPhase.PLAY;
  const isTrickEnd = state.phase === ScartoPhase.TRICK_END;
  const isRoundEnd = state.phase === ScartoPhase.ROUND_END;
  const isGameEnd = state.phase === ScartoPhase.GAME_END || state.gameEndFlag;

  const humanPlayer = state.players.find((p) => p.isHuman);

  // Scarto settles each deal against the mean: dealScore_i = N × (cardPoints_i − average),
  // so surfacing the table average and each seat's captured card-points explains the ± delta.
  const totalCardPoints = state.players.reduce((sum, p) => sum + p.cardPoints, 0);
  const averageCardPoints = state.players.length > 0 ? totalCardPoints / state.players.length : 0;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span>{t('trick', { n: state.trickNumber })}</span>
      </div>

      {/* **エクスキューズだけが勝敗の外にいる。**出した札が誰の手にも渡らず
          自分の得点山に戻るという規則は実装されているのに説明が無く、
          cardPoints の内訳が読み解けなかった (#6514)。Minchiate の mattoNote /
          Tarocchini の papiNote と同じ常設の 1 行。 */}
      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={view.cardWidth}
            label={t('currentTrick')}
            winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
            winnerLabel={
              isTrickEnd
                ? t('previousTrickWinner', { name: findPlayerName(state.players, state.lastTrickWinner) })
                : undefined
            }
            dataTutorial="scarto-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="scarto-info">
          {/* Per-player scores with a dealer badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm flex flex-wrap gap-x-4">
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                <span className={p.isDealer ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                </span>
                {p.isDealer && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
                )}
              </div>
            ))}
          </div>

          <details
            className="glass-panel mb-2 p-2 rounded text-ds-text-muted text-sm"
            open={!view.isMobile || undefined}
          >
            <summary className="cursor-pointer select-none">{t('rules')}</summary>
            <div className="mt-1" data-testid="scarto-rules">
              <div className="mb-2 text-center font-semibold text-ds-warning" data-testid="scarto-excuse-note">
                {t('excuseReturnsNote')}
              </div>
              <section aria-label={t('pointReference.title')} data-testid="scarto-point-reference">
                <div className="mb-1 text-ds-text-primary">{t('pointReference.title')}</div>
                <div>{t('pointReference.king')}</div>
                <div>{t('pointReference.queen')}</div>
                <div>{t('pointReference.knight')}</div>
                <div>{t('pointReference.jack')}</div>
                <div>{t('pointReference.bouts')}</div>
                <div>{t('pointReference.other')}</div>
              </section>
            </div>
          </details>

          {/* Players: cards / tricks / captured points */}
          {humanPlayer && (
            <section
              className="mb-2 px-2 py-1 rounded bg-black/30 text-ds-text-muted text-sm"
              data-testid="scarto-human-seat"
              data-tutorial="scarto-human-seat"
              aria-label={`${playerName(humanPlayer.id, true)}: ${t('cards', { count: humanPlayer.cardCount })} | ${t('tricks', { count: humanPlayer.trickCount })} | ${t('points', { points: humanPlayer.cardPoints })}`}
            >
              {playerName(humanPlayer.id, true)}: {t('cards', { count: humanPlayer.cardCount })} |{' '}
              {t('tricks', { count: humanPlayer.trickCount })} | {t('points', { points: humanPlayer.cardPoints })}
            </section>
          )}
          <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
            <div className="mb-2 p-2 rounded bg-black/30 flex flex-wrap gap-x-4">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })}
                  </div>
                ))}
            </div>
          </CpuAccordion>

          {/* Round result: the deal settlement (signed delta from the average) */}
          {(isScartoPhase || isPlayPhase || isTrickEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="scarto-provisional">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.provisional')}</div>
              <div data-testid="scarto-provisional-breakdown">
                <ScartoBreakdown t={t} players={state.players} average={averageCardPoints} />
              </div>
            </div>
          )}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="scarto-result">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
              {state.players.map((p, i) => {
                const delta = state.dealScores[i] ?? 0;
                return (
                  <div key={p.id}>
                    {t('roundResult.playerLine', {
                      name: playerName(p.id, p.isHuman),
                      delta: formatSignedDelta(delta),
                      score: p.score,
                    })}
                  </div>
                );
              })}
              {/* Average-difference breakdown: each seat's captured points vs. the table mean. */}
              <div className="mt-1 pt-1 border-t border-white/10" data-testid="scarto-breakdown">
                {/* **平均差と実際の変動は N 倍ちがう。**式を書かないと、同じ
                    箱の中で「+15」と「平均差 +5」が並んで計算が合わないように
                    見える (#4930)。 */}
                <ScartoBreakdown
                  t={t}
                  players={state.players}
                  average={averageCardPoints}
                  formulaTestId="scarto-formula"
                />
              </div>
            </div>
          )}
        </div>
      </div>

      {view.isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: ScartoContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="scarto-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndices &&
              state.hint.cardIndices.length > 0 &&
              ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand({ game, state, t, view }: ScartoContext) {
  const isScartoPhase = state.phase === ScartoPhase.SCARTO;
  const isPlayPhase = state.phase === ScartoPhase.PLAY;
  const canScarto = isScartoPhase && state.isHumanScarto;
  const canPlay = isPlayPhase && state.isHumanTurn;

  const humanPlayer = state.players.find((p) => p.isHuman);

  // **捨てられる札はサーバが決める。** 色と値からここで組み立てると、
  // 捨てられるピップが足りないときに解禁される非オヌール切り札が落ちる。
  // その手を引いた親は、画面から枚数を揃えられなかった (#6236)。
  const discardableIndices = canScarto ? state.discardableIndices : undefined;
  const handValidIndices = canPlay ? state.playableIndices : canScarto ? discardableIndices : undefined;

  // During the scarto, explain per-card why an un-buriable card cannot be
  // discarded (trump / Excuse / bout / counting card) via the card tooltip.
  // Purely additive — it never blocks selection; the backend still rejects
  // illegal discards.
  const scartoTitleFor = (idx: number): string | undefined => {
    const card = humanPlayer?.cards[idx];
    if (!card) return undefined;
    // いま実際に捨てられる札には理由を出さない。ピップが足りないときの
    // 切り札は「捨てられない」ではなく捨てられるので、一覧と食い違う (#6236)。
    if (discardableIndices?.includes(idx)) return undefined;
    const reason = scartoUndiscardableReason(card);
    return reason ? t(`scartoUndiscardable.${reason}`) : undefined;
  };

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="scarto-prompt-live" role="status" aria-live="polite">
        {canScarto && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="scarto-discard-prompt">
            {t('scartoPhase', { count: game.selectedCardIndices.length, total: SCARTO_DISCARD_COUNT })}
          </div>
        )}
      </div>
      {isScartoPhase && !state.isHumanScarto && (
        <div className="mb-1 text-center text-sm text-ds-text-muted" data-testid="scarto-waiting">
          {t('scartoWaiting')}
        </div>
      )}
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          dataTutorialPrefix="scarto"
          stackRows
          validIndices={handValidIndices}
          restrictedTooltip={canScarto ? t('scartoRestricted') : t('playButton')}
          cardTitleFor={canScarto ? scartoTitleFor : undefined}
        />
      )}
    </>
  );
}

function primaryActions({ game, state, t, busy }: ScartoContext) {
  const isScartoPhase = state.phase === ScartoPhase.SCARTO;
  const isPlayPhase = state.phase === ScartoPhase.PLAY;
  const isTrickEnd = state.phase === ScartoPhase.TRICK_END;
  const isRoundEnd = state.phase === ScartoPhase.ROUND_END;

  const canScarto = isScartoPhase && state.isHumanScarto;
  const canPlay = isPlayPhase && state.isHumanTurn;

  if (!canScarto && !canPlay && !isTrickEnd && !isRoundEnd) return null;

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="scarto-action-buttons">
      {canScarto && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handleScarto}
          disabled={busy || game.selectedCardIndices.length !== SCARTO_DISCARD_COUNT}
        >
          {t('discardButton', { count: game.selectedCardIndices.length, total: SCARTO_DISCARD_COUNT })}
        </button>
      )}
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const scartoDefinition: GameDefinition<ScartoResponse, ScartoGame, Parameters<typeof scartoApi.exec>, ScartoView> = {
  key: 'scarto',
  useGame: useScartoGame,
  useView: useScartoView,
  tutorial: SCARTO_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 25 },
  cli: () => ({
    parseCommand: parseScartoCommand,
    formatResponse: formatScartoState,
    helpText: SCARTO_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === ScartoPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerPlayer === humanIdx,
    };
  },
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
          value: game.scartoConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'number',
          id: 'targetDeals',
          label: t('settings.targetDeals'),
          value: game.scartoConfig.targetDeals,
          min: 1,
          max: 100,
          description: t('settings.targetDealsGuide'),
          onSelect: (v) => game.handleConfigChange('targetDeals', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  footerTall: true,
  primaryActions,
  message: ({ state, view }) =>
    !view.isMobile && (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ),
  resetTutorialId: 'scarto-reset-button',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Scarto (スカルト) game page: a 3-player 78-card Italian tarocchi trick-taker with a dealer scarto (discard) and trump-priority tricks — no bidding, chien, or partnership. */
export const ScartoPage = createGamePage(scartoDefinition);
