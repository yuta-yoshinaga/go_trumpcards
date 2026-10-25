import { useEffect } from 'react';
import type { piedmonteseTarotApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  PIEDMONTESE_CPU_DIFFICULTY_OPTIONS,
  PIEDMONTESE_SEAT_OPTIONS,
  PIEDMONTESE_TARGET_DEALS_OPTIONS,
  usePiedmonteseTarotGame,
} from '../hooks/usePiedmonteseTarotGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { PiedmonteseTarotResponse } from '../types/card';
import { PiedmonteseTarotPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PIEDMONTESE_TAROT_HELP, parsePiedmonteseTarotCommand } from '../utils/cli/commands/piedmonteseTarotCommands';
import { formatPiedmonteseTarotState } from '../utils/cli/formatters/piedmonteseTarotFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
// **捨てられない理由の判定は共有。** スカルト (捨て札) の規則は 3 人版と同じで、
// 判定はタロット札の面 (色と値) だけを見ている。
import { scartoUndiscardableReason } from '../utils/scartoDiscard';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tarocco Piemontese tutorial step definitions. */
const PIEDMONTESE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="piedmontesetarot-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="piedmontesetarot-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="piedmontesetarot-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="piedmontesetarot-info"]',
    messageKey: 'tutorial.info',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="piedmontesetarot-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PIEDMONTESE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PiedmonteseTarotPhase.SCARTO]: 'scarto',
  [PiedmonteseTarotPhase.PLAY]: 'play',
  [PiedmonteseTarotPhase.TRICK_END]: 'trickEnd',
  [PiedmonteseTarotPhase.ROUND_END]: 'roundEnd',
  [PiedmonteseTarotPhase.GAME_END]: 'gameEnd',
};

/** Outcome i18n keys indexed by outcome value (0=none, 1=above average, 2=below). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** Formats thirds as whole points with an optional one-third remainder. */
function formatThirds(thirds: number): string {
  const whole = Math.floor(thirds / 3);
  const remainder = thirds % 3;
  return remainder === 0 ? String(whole) : `${whole} ${remainder}/3`;
}

type PiedmonteseTarotGame = ReturnType<typeof usePiedmonteseTarotGame>;
type PiedmonteseTarotView = ReturnType<typeof usePiedmonteseTarotView>;
type PiedmonteseTarotContext = GamePageContext<PiedmonteseTarotResponse, PiedmonteseTarotGame, PiedmonteseTarotView>;

function usePiedmonteseTarotView({ game }: { game: PiedmonteseTarotGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('piedmontesetarot', PIEDMONTESE_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, view }: PiedmonteseTarotContext) {
  const isTrickEnd = state.phase === PiedmonteseTarotPhase.TRICK_END;
  const isRoundEnd = state.phase === PiedmonteseTarotPhase.ROUND_END;
  const isGameEnd = state.phase === PiedmonteseTarotPhase.GAME_END || state.gameEndFlag;
  const { cardWidth, isMobile } = view;

  // 78 点を席で分け合うゼロサム精算なので、卓の合計と各席の取り分を出して
  // 変動の出どころを見せる。
  const totalThirds = state.players.reduce((sum, p) => sum + p.cardThirds, 0);

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span>{t('trick', { n: state.trickNumber, total: state.trickCount })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            // **取った席を光らせる。** サーバは lastTrickWinner を最初から
            // 送っているのに読んでいなかったので、4 枚並んだ場札の誰が
            // 取ったのか画面から読めず、毎回タロッコの切り札優先を
            // 自分で解き直すことになっていた (#6623)。Gleek が同じ
            // コンポーネントに同じ形で渡している。
            //
            // プレイ中は undefined。まだ決まっていない勝者を先に
            // 光らせると、最後の 1 枚で覆る答えを見せることになる。
            winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
            dataTutorial="piedmontesetarot-trick-display"
          />
          <span className="sr-only" role="status" aria-live="polite" data-testid="piedmontesetarot-trick-winner-live">
            {isTrickEnd
              ? t('trickWinnerAnnouncement', {
                  name: playerName(state.lastTrickWinner, state.players[state.lastTrickWinner]?.isHuman ?? false),
                })
              : ''}
          </span>

          {/* Completed trick history: lets players review past tricks and cards */}
          <details
            className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
            data-testid="piedmontesetarot-trick-history"
          >
            <summary className="cursor-pointer select-none text-ds-text-muted text-sm font-semibold">
              {t('trickHistory.title')}
            </summary>
            <div className="mt-2">
              {state.completedTricks.length === 0 ? (
                <div className="text-ds-text-muted text-xs">{t('trickHistory.empty')}</div>
              ) : (
                <div className="space-y-3">
                  {state.completedTricks.map((trick) => (
                    <div
                      key={trick.trickNumber}
                      className="border-t border-ds-border-subtle pt-2 first:border-0 first:pt-0"
                      data-testid={`piedmontesetarot-completed-trick-${trick.trickNumber}`}
                    >
                      <div className="flex flex-wrap gap-x-3 gap-y-1 items-center mb-1 text-xs">
                        <span className="font-semibold text-ds-text-primary">
                          {t('trickHistory.trickTitle', { n: trick.trickNumber })}
                        </span>
                        <span>
                          {t('trickHistory.lead', {
                            name: playerName(trick.leadPlayerIdx, state.players[trick.leadPlayerIdx]?.isHuman ?? false),
                          })}
                        </span>
                        <span className="text-ds-warning font-semibold">
                          {t('trickHistory.winner', {
                            name: playerName(trick.winnerIdx, state.players[trick.winnerIdx]?.isHuman ?? false),
                          })}
                        </span>
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {trick.cards.map((tc, cardIdx) => {
                          const p = state.players[tc.playerIdx];
                          const isWinner = tc.playerIdx === trick.winnerIdx;
                          return (
                            <div
                              key={cardIdx}
                              className={`text-center p-1 rounded ${isWinner ? 'ring-1 ring-ds-warning' : ''}`}
                              data-testid={`piedmontesetarot-history-card-${trick.trickNumber}-${tc.playerIdx}`}
                            >
                              <CardImage
                                card={tc.card}
                                width={Math.round(cardWidth * 0.6)}
                                ariaLabel={cardAlt(tc.card)}
                              />
                              <div
                                className={`text-xs mt-0.5 ${
                                  isWinner ? 'text-ds-warning font-semibold' : 'text-ds-text-muted'
                                }`}
                              >
                                {playerName(tc.playerIdx, p?.isHuman ?? false)}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </details>
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="piedmontesetarot-info">
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
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

          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })}
                </div>
              ))}
            </div>
          )}

          {(isRoundEnd || isGameEnd) && (
            <div
              className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              data-testid="piedmontesetarot-result"
            >
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
              {state.scartoCards.length > 0 && (
                <div className="mt-1" data-testid="piedmontesetarot-scarto-cards">
                  <div>{t('roundResult.scartoCards')}</div>
                  <div className="flex flex-wrap gap-1">
                    {state.scartoCards.map((card, i) => (
                      <CardImage
                        key={`${card.design}-${card.value}-${i}`}
                        card={card}
                        ariaLabel={cardAlt(card)}
                        width={42}
                      />
                    ))}
                  </div>
                </div>
              )}
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
              <div className="mt-1 pt-1 border-t border-white/10" data-testid="piedmontesetarot-breakdown">
                <div>{t('roundResult.total', { total: totalThirds / 3 })}</div>
                {/* **式を書く。** 「取り分」と「変動」は席数倍ちがうので、
                    並べただけでは計算が合わないように見える。 */}
                <div data-testid="piedmontesetarot-formula">
                  {t('roundResult.formula', { n: state.players.length })}
                </div>
                {state.players.map((p, i) => (
                  <div key={p.id}>
                    <div>
                      {t('roundResult.earnedLine', {
                        name: playerName(p.id, p.isHuman),
                        points: p.cardPoints,
                        scaled: formatSignedDelta(state.dealScores[i] ?? 0),
                      })}
                    </div>
                    <div>
                      {t('roundResult.trickPoints', {
                        name: playerName(p.id, p.isHuman),
                        points: formatThirds(p.cardThirds - p.scartoThirds),
                      })}
                    </div>
                    {p.scartoThirds > 0 && (
                      <div>
                        {t('roundResult.scartoPoints', {
                          name: playerName(p.id, p.isHuman),
                          points: formatThirds(p.scartoThirds),
                        })}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
          <section
            className="my-3 p-2 rounded bg-ds-surface text-ds-text-muted text-sm"
            data-testid="piedmontesetarot-deal-score-history"
          >
            <div className="mb-1 text-ds-text-primary">{t('dealScoreHistory.title')}</div>
            {state.dealScoreHistory.length === 0 ? (
              <div>{t('dealScoreHistory.empty')}</div>
            ) : (
              state.dealScoreHistory.map((deal) => (
                <div key={deal.roundNumber}>
                  {t('dealScoreHistory.row', {
                    round: deal.roundNumber,
                    scores: deal.scores
                      .map((score, i) => `${playerName(i, state.players[i].isHuman)} ${formatSignedDelta(score)}`)
                      .join(t('listSeparator')),
                  })}
                </div>
              ))
            )}
          </section>
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: PiedmonteseTarotContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isScartoPhase = state.phase === PiedmonteseTarotPhase.SCARTO;
  const isPlayPhase = state.phase === PiedmonteseTarotPhase.PLAY;

  const canScarto = isScartoPhase && state.isHumanScarto;
  const canPlay = isPlayPhase && isHumanTurn;
  // **捨てる枚数は卓が決める。** 4 人なら 2 枚、3 人なら 3 枚。定数にすると
  // どちらかの卓で必ずボタンが押せなくなる。
  const talonSize = state.talonSize;

  // **捨てられる札はサーバが決める。** 色と値からここで組み立てると、
  // 捨てられるピップが足りないときに解禁される非オヌール切り札が落ちる。
  // その手を引いた親は、画面から枚数を揃えられなかった (#6236)。
  const discardableIndices = canScarto ? state.discardableIndices : undefined;

  const handValidIndices = canPlay ? state.playableIndices : canScarto ? discardableIndices : undefined;

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
      <div data-testid="piedmontesetarot-prompt-live" role="status" aria-live="polite">
        {canScarto && (
          <div
            className="mb-1 text-center text-sm text-ds-accent font-semibold"
            data-testid="piedmontesetarot-discard-prompt"
          >
            {t('scartoPhase', { count: game.selectedCardIndices.length, total: talonSize })}
          </div>
        )}
      </div>
      {isScartoPhase && !state.isHumanScarto && (
        <div className="mb-1 text-center text-sm text-ds-text-muted" data-testid="piedmontesetarot-waiting">
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
          dataTutorialPrefix="piedmontesetarot"
          stackRows
          validIndices={handValidIndices}
          // **出せる札には輪を付ける。** `validIndices` は「押せない札を
          // 薄くする」だけで、`data-legal` (と成功色のリング) を付けるのは
          // `legalIndices` のほう。片方だけ渡すと、出せる札が画面上で
          // まったく強調されない。
          legalIndices={handValidIndices}
          restrictedTooltip={canScarto ? t('scartoRestricted') : t('playButton')}
          cardTitleFor={canScarto ? scartoTitleFor : undefined}
        />
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: PiedmonteseTarotContext) {
  return (
    <>
      {/* ライブ領域は常設。hint がある間だけ現れる div に付けると読み上げられない (#5955)。 */}
      <div data-testid="piedmontesetarot-hint-live" role="status" aria-live="polite">
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

function primaryActions({ game, state, t, busy }: PiedmonteseTarotContext) {
  const isHumanTurn = state.isHumanTurn;
  const isScartoPhase = state.phase === PiedmonteseTarotPhase.SCARTO;
  const isPlayPhase = state.phase === PiedmonteseTarotPhase.PLAY;
  const isTrickEnd = state.phase === PiedmonteseTarotPhase.TRICK_END;
  const isRoundEnd = state.phase === PiedmonteseTarotPhase.ROUND_END;

  const canScarto = isScartoPhase && state.isHumanScarto;
  const canPlay = isPlayPhase && isHumanTurn;
  const talonSize = state.talonSize;

  if (!canScarto && !canPlay && !isTrickEnd && !isRoundEnd) return null;

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="piedmontesetarot-action-buttons">
      {canScarto && (
        <button
          type="button"
          className={btnPrimary}
          onClick={() => game.handleScarto(talonSize)}
          disabled={busy || game.selectedCardIndices.length !== talonSize}
        >
          {t('discardButton', { count: game.selectedCardIndices.length, total: talonSize })}
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

const piedmonteseTarotDefinition: GameDefinition<
  PiedmonteseTarotResponse,
  PiedmonteseTarotGame,
  Parameters<typeof piedmonteseTarotApi.exec>,
  PiedmonteseTarotView
> = {
  key: 'piedmontesetarot',
  useGame: usePiedmonteseTarotGame,
  useView: usePiedmonteseTarotView,
  tutorial: PIEDMONTESE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 19 },
  cli: () => ({
    parseCommand: parsePiedmonteseTarotCommand,
    formatResponse: formatPiedmonteseTarotState,
    helpText: PIEDMONTESE_TAROT_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === PiedmonteseTarotPhase.GAME_END || state.gameEndFlag;
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
          id: 'seats',
          label: t('settings.seats'),
          value: game.piedmonteseConfig.seats,
          options: PIEDMONTESE_SEAT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('seats', v),
        },
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.piedmonteseConfig.cpuDifficulty,
          options: PIEDMONTESE_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetDeals',
          label: t('settings.targetDeals'),
          value: game.piedmonteseConfig.targetDeals,
          options: PIEDMONTESE_TARGET_DEALS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetDeals', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  // Tarot-size hands outgrow the 45vh cap on phones (#11655).
  footerTall: true,
  resetTutorialId: 'piedmontesetarot-reset-button',
  actionsClassName: 'flex items-center gap-2',
  Board,
  FooterHand,
  FooterStatus,
};

/**
 * Renders the Tarocco Piemontese page: the Piedmontese 78-card tarot for three
 * or four players, with the dealer burying the talon and trump-priority tricks.
 */
export const PiedmonteseTarotPage = createGamePage(piedmonteseTarotDefinition);
