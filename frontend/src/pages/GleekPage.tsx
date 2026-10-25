import { useEffect } from 'react';
import type { gleekApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useGleekGame } from '../hooks/useGleekGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { GleekResponse } from '../types/card';
import { GleekPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { GLEEK_HELP, parseGleekCommand } from '../utils/cli/commands/gleekCommands';
import { formatGleekState } from '../utils/cli/formatters/gleekFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Gleek tutorial step definitions. */
const GLEEK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="gleek-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gleek-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gleek-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="gleek-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="gleek-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GLEEK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GleekPhase.BID]: 'bid',
  [GleekPhase.DISCARD]: 'discard',
  [GleekPhase.PLAY]: 'play',
  [GleekPhase.TRICK_END]: 'trickEnd',
  [GleekPhase.ROUND_END]: 'roundEnd',
  [GleekPhase.GAME_END]: 'gameEnd',
};

/** Trump-suit i18n keys indexed by suit code (1=♠ 2=♣ 3=♥ 4=♦); index 0 = none. */
const SUIT_KEYS = ['suitNone', 'suitSpade', 'suitClub', 'suitHeart', 'suitDiamond'] as const;

/** Meld-rank i18n keys, keyed by the card value the set is made of. */
const RANK_KEYS: Readonly<Record<number, string>> = { 1: 'rankAce', 13: 'rankKing', 12: 'rankQueen', 11: 'rankJack' };

type GleekGame = ReturnType<typeof useGleekGame>;
type GleekView = ReturnType<typeof useGleekView>;
type GleekContext = GamePageContext<GleekResponse, GleekGame, GleekView>;

function useGleekView({ game }: { game: GleekGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('gleek', GLEEK_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, view }: GleekContext) {
  const { cardWidth } = view;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isTrickEnd = state.phase === GleekPhase.TRICK_END;
  const isRoundEnd = state.phase === GleekPhase.ROUND_END;
  const isGameEnd = state.phase === GleekPhase.GAME_END || state.gameEndFlag;
  const breakdownLabels = t('roundResult.stages', { returnObjects: true }) as string[];
  const biddingSeatIdx = state.phase === GleekPhase.BID ? state.currentBidderIdx : -1;
  const trumpLabel = state.trumpSuit >= 1 ? t(SUIT_KEYS[state.trumpSuit] ?? 'suitNone') : t('suitNone');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>{t('trump', { suit: trumpLabel })}</span>
      </div>

      {/* **段階の点は出さないと見えない。** 競りとトリックの間で累積点が
          動く理由は、ここに出さないと画面のどこにも現れない。 */}
      <div className="text-ds-text-muted text-center text-sm mb-2" data-testid="gleek-stage-line">
        <div>
          {state.buyerIdx >= 0
            ? t('stockSold', {
                name: playerName(state.buyerIdx, state.buyerIdx === humanIdx),
                bid: state.winningBid,
              })
            : t('stockUnsold', { bid: state.highestBid })}
        </div>
        {state.phase >= GleekPhase.PLAY && (
          <div data-testid="gleek-ruff-line">
            {state.players.map((player) => (
              <div key={player.id}>
                {t(player.id === state.ruffWinnerIdx ? 'ruffLine' : 'ruffOtherLine', {
                  name: playerName(player.id, player.id === humanIdx),
                  total: player.ruff,
                  suit: t(SUIT_KEYS[player.ruffSuit] ?? 'suitNone'),
                })}
              </div>
            ))}
          </div>
        )}
        {state.melds.map((m) => (
          <div key={`${m.playerIdx}-${m.rank}`} data-testid="gleek-meld-line">
            {t(m.count >= 4 ? 'meldMournival' : 'meldGleek', {
              name: playerName(m.playerIdx, m.playerIdx === humanIdx),
              rank: t(RANK_KEYS[m.rank] ?? 'rankAce'),
              value: m.value,
            })}
          </div>
        ))}
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="gleek-trick-display"
            // **取った札を光らせる。** トリックが揃ってから次へ進むまでの間に
            // 誰が取ったのかを出さないと、棋譜を開くまで分からない。
            winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="gleek-info">
          {state.turnUp && (
            <div className="mb-2 p-2 rounded bg-ds-surface flex items-center gap-3" data-testid="gleek-turn-up">
              <CardImage card={state.turnUp} width={56} />
              <div className="text-sm">
                <div className="text-ds-text-primary">{t('turnUp', { card: cardAlt(state.turnUp) })}</div>
                <div className="text-ds-text-muted">{t('turnUpDescription')}</div>
              </div>
            </div>
          )}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => (
              <div
                key={p.id}
                className={`py-0.5 flex items-center gap-2 ${p.id === biddingSeatIdx ? 'text-ds-accent font-semibold' : ''}`}
                data-testid={`gleek-player-${p.id.toString()}`}
              >
                <span className={p.isBuyer ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                </span>
                {p.id === biddingSeatIdx && <span data-testid="gleek-bid-turn-label">{t('bidTurn')}</span>}
                {state.phase === GleekPhase.BID && (
                  <span data-testid={`gleek-player-bid-${p.id.toString()}`}>
                    {t('playerBid', { bid: p.bid })}
                    {p.passed ? ` · ${t('playerPassed')}` : ''}
                  </span>
                )}
                {p.isBuyer && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('buyerBadge')}</span>
                )}
              </div>
            ))}
          </div>

          <div className="mb-2 p-2 rounded bg-black/30">
            {state.players.map((p) => (
              <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                {t('tricks', { count: p.trickCount })} | {t('points', { points: p.trickPoints })}
              </div>
            ))}
          </div>

          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              {/* **基準点はそのディールから数える。** 上限を出すと、
                  名札が場外に落ちたディールで説明が合わなくなる。 */}
              <div data-testid="gleek-par-line">
                {t('roundResult.par', { total: state.dealPoints, par: state.par })}
              </div>
              {state.players.map((p, idx) => {
                const delta = state.roundDelta[idx];
                const breakdown = [
                  state.roundBreakdown.bid[idx],
                  state.roundBreakdown.ruff[idx],
                  state.roundBreakdown.meld[idx],
                  state.roundBreakdown.trick[idx],
                ];
                return (
                  <div
                    key={p.id}
                    className={`py-0.5 ${delta > 0 ? 'text-ds-success' : delta < 0 ? 'text-ds-error-text' : ''}`}
                    data-testid={`gleek-round-delta-${p.id.toString()}`}
                  >
                    {t('roundResult.delta', {
                      name: playerName(p.id, p.isHuman),
                      delta: formatSignedDelta(delta),
                    })}
                    <div data-testid={`gleek-round-breakdown-${p.id.toString()}`}>
                      {t('roundResult.breakdown', {
                        stages: breakdownLabels
                          .map((label, stageIdx) => `${label}: ${formatSignedDelta(breakdown[stageIdx])}`)
                          .join(t('listSeparator')),
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: GleekContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const canBid = state.phase === GleekPhase.BID && state.isHumanBidTurn;
  const canDiscard = state.phase === GleekPhase.DISCARD && state.isHumanDiscardTurn;
  const isPlayPhase = state.phase === GleekPhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn;

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="gleek-prompt-live" role="status" aria-live="polite">
        {canBid && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="gleek-bid-prompt">
            {t('bidPhase', { bid: state.highestBid })}
          </div>
        )}
        {canDiscard && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="gleek-discard-prompt">
            {t('discardPhase', { count: state.discardCount, selected: game.selectedCardIndices.length })}
          </div>
        )}
      </div>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={cardWidth}
          isMobile={isMobile}
          dataTutorialPrefix="gleek"
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
        />
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: GleekContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="gleek-hint-live" role="status" aria-live="polite">
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

function FooterActions({ game, state, t, busy }: GleekContext) {
  const canBid = state.phase === GleekPhase.BID && state.isHumanBidTurn;
  // **上限に達したら競り上げのボタンを出さない。** サーバは 0 以外を弾くので、
  // 押せるのに必ず失敗するボタンになる。
  const canRaise = canBid && state.nextBidAmount > 0;
  const canDiscard = state.phase === GleekPhase.DISCARD && state.isHumanDiscardTurn;
  const discardReady = game.selectedCardIndices.length === state.discardCount;
  const isPlayPhase = state.phase === GleekPhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn;
  const isTrickEnd = state.phase === GleekPhase.TRICK_END;
  const isRoundEnd = state.phase === GleekPhase.ROUND_END;

  return (
    <>
      {canBid && (
        <div className="flex flex-wrap gap-2 items-center" data-testid="gleek-bid-controls">
          {canRaise && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => game.handleBid(state.nextBidAmount)}
              disabled={busy}
            >
              {t('raiseTo', { amount: state.nextBidAmount })}
            </button>
          )}
          <button type="button" className={btnSecondary} onClick={() => game.handleBid(0)} disabled={busy}>
            {t('dropOut')}
          </button>
        </div>
      )}
      {canDiscard && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleDiscard}
          disabled={busy || !discardReady}
          data-testid="gleek-discard-confirm"
        >
          {t('discardButton', { count: state.discardCount })}
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
    </>
  );
}

const gleekDefinition: GameDefinition<GleekResponse, GleekGame, Parameters<typeof gleekApi.exec>, GleekView> = {
  key: 'gleek',
  useGame: useGleekGame,
  useView: useGleekView,
  tutorial: GLEEK_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 12 },
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  cli: () => ({
    parseCommand: parseGleekCommand,
    formatResponse: formatGleekState,
    helpText: GLEEK_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === GleekPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase] ?? '',
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
          value: game.gleekConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.gleekConfig.targetRounds,
          options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetRounds', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'gleek-action-buttons',
  resetTutorialId: 'gleek-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/**
 * Renders the Gleek game page: a 3-player 44-card trick-taker that settles four
 * scoring stages inside one deal — the auction for the stock, the ruff, the
 * gleek/mournival melds, and twelve tricks scored against a par read from the
 * deal itself.
 */
export const GleekPage = createGamePage(gleekDefinition);
