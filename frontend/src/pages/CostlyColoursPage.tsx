import type { TFunction } from 'i18next';
import { useEffect, useMemo, useRef, useState } from 'react';
import type { costlycoloursApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { useCardDimensions } from '../hooks/useCardDimensions';
import {
  COSTLYCOLOURS_CPU_DIFFICULTY_OPTIONS,
  COSTLYCOLOURS_TARGET_OPTIONS,
  useCostlyColoursGame,
} from '../hooks/useCostlyColoursGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { CostlyColoursRecentPlay, CostlyColoursResponse } from '../types/card';
import { CostlyColoursPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { COSTLYCOLOURS_HELP, parseCostlyColoursCommand } from '../utils/cli/commands/costlycoloursCommands';
import { formatCostlyColoursState } from '../utils/cli/formatters/costlycoloursFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Costly Colours tutorial step definitions. */
const COSTLYCOLOURS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="costlycolours-count"]',
    messageKey: 'tutorial.count',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="costlycolours-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="costlycolours-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="costlycolours-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="costlycolours-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしている。
const COSTLYCOLOURS_PHASE_KEYS: Readonly<Record<string, string>> = {
  [CostlyColoursPhase.MOG]: 'mog',
  [CostlyColoursPhase.PLAY]: 'play',
  [CostlyColoursPhase.SHOW]: 'show',
  [CostlyColoursPhase.GAME_END]: 'gameEnd',
};

const COSTLYCOLOURS_MILESTONES = [15, 25, 31];

function describeCostlyColoursPlay(
  play: CostlyColoursRecentPlay,
  players: CostlyColoursResponse['players'],
  t: TFunction<'costlycolours'>,
) {
  const milestone = COSTLYCOLOURS_MILESTONES.includes(play.total) ? t(`announcement.milestone.${play.total}`) : '';
  return t('announcement.play', {
    name: findPlayerName(players, play.seat),
    card: cardAlt(play.card),
    total: play.total,
    pointsText:
      play.points > 0
        ? t('announcement.points', {
            count: play.points,
            points: play.points,
            milestonePart: milestone ? t('announcement.milestonePart', { milestone }) : '',
          })
        : '',
  });
}

type CostlyColoursGame = ReturnType<typeof useCostlyColoursGame>;

function useCostlyColoursView({ game, t }: { game: CostlyColoursGame; t: TFunction<'costlycolours'> }) {
  const { state } = game;

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const [playAnnouncement, setPlayAnnouncement] = useState('');
  const announcedPlays = useRef('');

  useEffect(() => {
    if (!state?.recentPlays.length) {
      announcedPlays.current = '';
      setPlayAnnouncement('');
      return;
    }
    const key = JSON.stringify(state.recentPlays);
    if (key === announcedPlays.current) return;
    announcedPlays.current = key;
    setPlayAnnouncement(
      state.recentPlays.map((play) => describeCostlyColoursPlay(play, state.players, t)).join(t('listSeparator')),
    );
  }, [state, t]);

  const phaseNames: Readonly<Record<string, string>> = useMemo(() => {
    const names: Record<string, string> = {};
    for (const [phase, key] of Object.entries(COSTLYCOLOURS_PHASE_KEYS)) {
      names[phase] = t(`phase.${key}`);
    }
    return names;
  }, [t]);

  return { cardWidth, isMobile, playAnnouncement, phaseNames };
}

type CostlyColoursView = ReturnType<typeof useCostlyColoursView>;
type CostlyColoursContext = GamePageContext<CostlyColoursResponse, CostlyColoursGame, CostlyColoursView>;

function Board({ state, t, view }: CostlyColoursContext) {
  const isShow = state.phase === CostlyColoursPhase.SHOW;
  const isGameEnd = state.phase === CostlyColoursPhase.GAME_END || state.gameEndFlag;
  const { cardWidth } = view;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        {t('deal', { n: state.dealNumber, target: state.config.targetScore })}
      </div>

      <div className={lgTwoColGrid}>
        <div data-tutorial="costlycolours-count">
          {/* **表の 1 枚は常に見せる。** ショーの色役も J / 2 の 4 点もこれ次第。 */}
          <div className="mb-1 text-ds-text-muted text-sm">{t('turnUpLabel')}</div>
          <div className="mb-2 p-2 rounded bg-black/30 flex gap-1 items-center" data-testid="costlycolours-turnup">
            {state.turnUp ? (
              <CardImage card={state.turnUp} width={cardWidth} />
            ) : (
              <span className="text-ds-text-muted text-sm">{t('noTurnUp')}</span>
            )}
          </div>

          <div className="mb-1 text-ds-text-muted text-sm">{t('countLabel')}</div>
          <div
            className="mb-2 p-2 rounded bg-black/30 flex flex-wrap gap-1 items-center"
            data-testid="costlycolours-pile"
          >
            {state.pile.length === 0 ? (
              <span className="text-ds-text-muted text-sm">{t('pileEmpty')}</span>
            ) : (
              state.pile.map((c, i) => <CardImage key={`${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />)
            )}
          </div>
          <div className="text-ds-text-primary" data-testid="costlycolours-total">
            {t('total', { n: state.total })}
          </div>
          {state.recentPlays.length > 0 && (
            <ol className="mt-2 space-y-1" data-testid="costlycolours-recent-plays">
              {state.recentPlays.map((play, index) => (
                <li key={index} className="text-ds-text-muted text-sm">
                  {describeCostlyColoursPlay(play, state.players, t)}
                </li>
              ))}
            </ol>
          )}
        </div>

        <div data-tutorial="costlycolours-scores">
          <div className="mb-2 p-2 rounded bg-black/30" data-testid="costlycolours-scores">
            {state.players.map((p) => (
              <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                <span className="flex items-center gap-2">
                  <span>
                    {playerName(p.id, p.isHuman)}: {t('score', { n: p.score })} / {t('cards', { n: p.cardCount })}
                  </span>
                  {p.isDealer && (
                    <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
                  )}
                </span>
                {(isShow || isGameEnd) && (
                  <div
                    className={`mt-1 p-2 rounded ${p.isHuman ? 'bg-ds-surface' : 'bg-ds-surface-elevated'}`}
                    data-testid={`costlycolours-show-hand-${p.id}`}
                  >
                    <div className="mb-1 text-ds-text-primary">
                      {t('showHand', { name: playerName(p.id, p.isHuman) })}
                    </div>
                    <div className="flex gap-1 items-center">
                      {p.cards.map((card, index) => (
                        <CardImage key={`${card.design}-${card.value}-${index}`} card={card} width={cardWidth} />
                      ))}
                      {state.turnUp && <CardImage card={state.turnUp} width={cardWidth} />}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>

          {state.lastResult && (isShow || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="costlycolours-show">
              <div className="mb-1 text-ds-text-primary">{t('showTitle')}</div>
              {state.lastResult.lines
                .filter((l) => l.points[0] !== 0 || l.points[1] !== 0)
                .map((l) => (
                  <div key={l.key}>
                    {t(`score.${l.key}`)}: {l.points[0]} - {l.points[1]}
                  </div>
                ))}
              {/* **どの色役が付いたのかを名指す。** 点だけだと梯子のどこか分からない。 */}
              {state.lastResult.combos.map((combo, i) =>
                combo ? (
                  <div key={combo + String(i)} className="text-ds-warning" data-testid={`costlycolours-combo-${i}`}>
                    {playerName(i, i === 0)}: {t(`combo.${combo}`)}
                  </div>
                ) : null,
              )}
              <div className="text-ds-text-primary">
                {t('showTotal', { a: state.lastResult.totals[0], b: state.lastResult.totals[1] })}
              </div>
            </div>
          )}

          {isGameEnd && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="costlycolours-winner">
              {t('winner', { name: playerName(state.winnerIdx, state.winnerIdx === 0) })}
            </div>
          )}
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: CostlyColoursContext) {
  const isMogPhase = state.phase === CostlyColoursPhase.MOG;
  const isPlayPhase = state.phase === CostlyColoursPhase.PLAY;
  const isShow = state.phase === CostlyColoursPhase.SHOW;
  const isGameEnd = state.phase === CostlyColoursPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && state.isHumanTurn;

  // **出せる札はサーバが数えたものだけ。** 31 を超える札を並べると押しても弾かれる。
  const playable = canPlay ? state.playableIdxs : [];
  const humanPlayer = state.players.find((p) => p.isHuman);
  const { cardWidth, isMobile } = view;

  return (
    <>
      {!state.isHumanTurn && !isMogPhase && !isShow && !isGameEnd && (
        <p className="text-ds-text-muted text-sm py-1">{t('waitingForOpponent')}</p>
      )}
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={[]}
          toggleCard={game.play}
          cardWidth={cardWidth}
          isMobile={isMobile}
          dataTutorialPrefix="costlycolours"
          validIndices={canPlay ? playable : undefined}
          legalIndices={canPlay ? playable : undefined}
          restrictedTooltip={t('restrictedTooltip')}
          cardBadgeFor={(idx) => {
            const card = humanPlayer.cards[idx];
            if (!(card.value === 11 || card.value === 2)) return null;
            const points = state.turnUp?.design === card.design ? 4 : 2;
            return { glyph: String(points), title: t('jackDeuceMark', { points }) };
          }}
        />
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint, view }: CostlyColoursContext) {
  return (
    <>
      {/* ライブ領域は常設 (#5955)。 */}
      <div data-testid="costlycolours-hint-live" role="status" aria-live="polite">
        {isRequestedHint(state) && state.hintReason && state.hintReason !== 'none' && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hintReason}`)}
            {state.hintHandIdx >= 0 && ` ([${state.hintHandIdx}])`}
          </div>
        )}
      </div>
      <div className="sr-only" data-testid="costlycolours-play-live" role="status" aria-live="polite">
        {view.playAnnouncement && <span>{view.playAnnouncement}</span>}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: CostlyColoursContext) {
  const isMogPhase = state.phase === CostlyColoursPhase.MOG;
  const isShow = state.phase === CostlyColoursPhase.SHOW;

  return (
    <>
      {/* **応じる／断るは別のボタン。** 断ると相手に 1 点入るので、
          片方を既定にしない。 */}
      {isMogPhase && (
        <>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => game.mog(true)}
            disabled={busy}
            data-testid="costlycolours-mog"
          >
            {t('mogAccept')}
          </button>
          {/* **押す前に、どの札が出ていくのかを出す。** 渡す 1 枚は
              ドメインが選ぶので、打ち手は結果を見るまで自分の 3 枚の
              どれが失われるか分からなかった (#6631)。 */}
          {state.mogDiscardCard && (
            <span className="text-ds-text-muted text-sm" data-testid="costlycolours-mog-discard">
              {t('mogDiscard', { card: cardAlt(state.mogDiscardCard) })}
            </span>
          )}
          <button
            type="button"
            className={btnSecondary}
            onClick={() => game.mog(false)}
            disabled={busy}
            data-testid="costlycolours-nomog"
          >
            {t('mogRefuse')}
          </button>
        </>
      )}
      {isShow && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleNextDeal}
          disabled={busy}
          data-testid="costlycolours-next-deal"
        >
          {t('nextDeal')}
        </button>
      )}
    </>
  );
}

const costlyColoursDefinition: GameDefinition<
  CostlyColoursResponse,
  CostlyColoursGame,
  Parameters<typeof costlycoloursApi.exec>,
  CostlyColoursView
> = {
  key: 'costlycolours',
  useGame: useCostlyColoursGame,
  useView: useCostlyColoursView,
  tutorial: COSTLYCOLOURS_TUTORIAL_STEPS,
  skeleton: { kind: 'casino-table', sections: [1], footerStyle: 'hand', footerHandSize: 3 },
  cli: () => ({
    parseCommand: parseCostlyColoursCommand,
    formatResponse: formatCostlyColoursState,
    helpText: COSTLYCOLOURS_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase] ?? view.phaseNames[CostlyColoursPhase.PLAY],
    ended: state.phase === CostlyColoursPhase.GAME_END || state.gameEndFlag,
    won: (state.phase === CostlyColoursPhase.GAME_END || state.gameEndFlag) && state.winnerIdx === 0,
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
          value: game.costlyColoursConfig.cpuDifficulty,
          options: COSTLYCOLOURS_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetScore',
          label: t('settings.targetScore'),
          value: game.costlyColoursConfig.targetScore,
          options: COSTLYCOLOURS_TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetScore', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'costlycolours-action-buttons',
  resetTutorialId: 'costlycolours-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/**
 * Renders the Costly Colours page: the Shropshire ancestor of Cribbage, where
 * the count scores at 15, 25 and 31 and the show counts a ladder of colours.
 */
export const CostlyColoursPage = createGamePage(costlyColoursDefinition);
