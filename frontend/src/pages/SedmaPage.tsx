import { useEffect } from 'react';
import type { sedmaApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useSedmaGame } from '../hooks/useSedmaGame';
import { badgeInfoColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { Card, SedmaResponse } from '../types/card';
import { SedmaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSedmaCommand, SEDMA_HELP } from '../utils/cli/commands/sedmaCommands';
import { formatSedmaState } from '../utils/cli/formatters/sedmaFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Rank that captures any trick regardless of suit (`domain.SedmaWildValue`). */
const SEDMA_WILD_VALUE = 7;

/**
 * Indices of the wild sevens in a hand.
 *
 * **7 はどのスートでもトリックを奪える。**フォロー義務が無い捕獲ゲームなので、
 * 手札のどれが特別なのかが盤面から読み取れないと規則そのものが見えない (#6444)。
 * 印は `trumpIndices` の追加リングで、選択や合法手の枠を上書きしない。
 */
function wildSevenIndices(cards: Card[]): number[] {
  return cards.flatMap((c, i) => (c.value === SEDMA_WILD_VALUE ? [i] : []));
}

/** Sedma tutorial step definitions. */
const SEDMA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sedma-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sedma-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sedma-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sedma-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sedma-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SEDMA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SedmaPhase.PLAY]: 'play',
  [SedmaPhase.TRICK_END]: 'trickEnd',
  [SedmaPhase.ROUND_END]: 'roundEnd',
  [SedmaPhase.GAME_END]: 'gameEnd',
};

type SedmaGame = ReturnType<typeof useSedmaGame>;
type SedmaView = ReturnType<typeof useSedmaView>;
type SedmaContext = GamePageContext<SedmaResponse, SedmaGame, SedmaView>;

function useSedmaView({ game }: { game: SedmaGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('sedma', SEDMA_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function AboveContent({ game, t }: SedmaContext) {
  return (
    <div role="status" aria-live="polite">
      {game.hasPendingConfig && (
        <p className="px-4 pt-1 text-sm text-ds-text-muted">{t('settings.pendingApplication')}</p>
      )}
    </div>
  );
}

function Board({ state, t, view }: SedmaContext) {
  const { cardWidth, isMobile } = view;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const humanTeam = humanIdx % 2;
  const isRoundEnd = state.phase === SedmaPhase.ROUND_END;
  const isGameEnd = state.phase === SedmaPhase.GAME_END || state.gameEndFlag;
  const teamLabel = (team: number) => t(team === 0 ? 'team.a' : 'team.b');
  const relationLabel = (team: number) => t(team === humanTeam ? 'ally' : 'opponent');

  // One info-sidebar row per player, colour-coded by team (A = even ids = blue,
  // B = odd ids = red), matching the trick display's ally/opponent relationship.
  // Shared by the mobile (<details>) and desktop layouts.
  const renderPlayerRow = (p: (typeof state.players)[number]) => {
    const team = p.id % 2;
    const teamName = teamLabel(team);
    return (
      <div
        key={p.id}
        data-testid={`sedma-player-${p.id}`}
        data-team={team}
        className={`text-ds-text-muted text-sm py-0.5 border-l-2 pl-1 ${
          team === 0 ? 'border-ds-info' : 'border-ds-error'
        }`}
      >
        {/* Colour-independent team marker (WCAG 1.4.1): a visible A/B badge plus an
            sr-only full team name, so team membership isn't conveyed by colour alone. */}
        <span
          className={`inline-block mr-1 px-1 rounded text-xs font-bold ${
            team === 0 ? 'bg-ds-info text-ds-text-on-accent' : 'bg-ds-error text-white'
          }`}
          aria-hidden="true"
        >
          {team === 0 ? 'A' : 'B'}
        </span>
        <span className="sr-only">{teamName}: </span>
        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
        {state.dealerIdx === p.id && (
          <span className={`ml-2 px-1.5 py-0.5 rounded text-xs ${badgeInfoColors}`} data-testid="sedma-dealer-badge">
            {t('dealerBadge')}
          </span>
        )}
      </div>
    );
  };

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        {/* The value that decides the match was only visible in the panel. */}
        <span data-testid="sedma-target">{t('target', { points: state.config.targetPoints })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players.map((player) => ({ ...player, team: player.id % 2 }))}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="sedma-trick-display"
            playerLabelFor={(player) => {
              const team = player.id % 2;
              return `${teamLabel(team)} · ${playerName(player.id, player.isHuman)} · ${relationLabel(team)}`;
            }}
            cardAriaLabelFor={(player, card) => {
              const team = player.id % 2;
              return t('trickCardLabel', {
                card: cardAlt(card),
                player: playerName(player.id, player.isHuman),
                team: teamLabel(team),
                relation: relationLabel(team),
              });
            }}
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="sedma-info">
          {/* Team match scores. ラウンドが確定すると増える値なので、カード点
              パネルと同じく読み上げ対象にする (#5648)。 */}
          <div
            className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
            data-testid="sedma-team-scores"
            role="status"
            aria-live="polite"
          >
            {[0, 1].map((team) => {
              const score = state.teamScores[team] ?? 0;
              const remaining = state.config.targetPoints - score;
              return (
                <div key={team}>
                  {t('teamScore', { team: teamLabel(team), score })}
                  {remaining > 0 && t('remainingPoints', { count: remaining })}
                </div>
              );
            })}
            <div className="mt-1">
              {t('yourTeam')}: {teamLabel(humanTeam)}
            </div>
          </div>

          {/* Players: cards / tricks, colour-coded by 2-vs-2 team. Shared renderer for
              both layouts so the team-colour logic lives in one place. */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">{state.players.map(renderPlayerRow)}</div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">{state.players.map(renderPlayerRow)}</div>
          )}

          {/* Live captured card points during play (A and 10 = 10 pts each, +10
              last-trick bonus); the round-result block below takes over at round end. */}
          {!(isRoundEnd || isGameEnd) && (
            <div
              className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              data-testid="sedma-round-points"
              role="status"
            >
              <div className="mb-1 text-ds-text-primary">{t('roundPointsTitle')}</div>
              <div>{t('roundResult.teamA', { points: state.roundCardPoints[0] ?? 0 })}</div>
              <div>{t('roundResult.teamB', { points: state.roundCardPoints[1] ?? 0 })}</div>
            </div>
          )}

          {/* Round result */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.teamA', { points: state.roundCardPoints[0] ?? 0 })}</div>
              <div>{t('roundResult.teamB', { points: state.roundCardPoints[1] ?? 0 })}</div>
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: SedmaContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const canPlay = state.phase === SedmaPhase.PLAY && state.isHumanTurn;
  return (
    <>
      {/* **印だけでは規則にならない。**リングが何を意味するのかを言葉でも
          置く ── 印の意味が分からなければ、どれが 7 か分かっても
          「なぜ強いのか」は分からないまま。CUI は `promptPlayHelp` で
          最初から説明していた (#6444)。 */}
      <div className="mb-1 text-center text-ds-text-muted text-xs" data-testid="sedma-wild-rule">
        {t('wildRule')}
      </div>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          dataTutorialPrefix="sedma"
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
          trumpIndices={wildSevenIndices(humanPlayer.cards)}
          trumpTitle={t('wildRule')}
        />
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: SedmaContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="sedma-hint-live" role="status" aria-live="polite">
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

function FooterActions({ game, state, t, busy }: SedmaContext) {
  const isPlayPhase = state.phase === SedmaPhase.PLAY;
  const isTrickEnd = state.phase === SedmaPhase.TRICK_END;
  const isRoundEnd = state.phase === SedmaPhase.ROUND_END;
  const canPlay = isPlayPhase && state.isHumanTurn;

  return (
    <>
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

const sedmaDefinition: GameDefinition<SedmaResponse, SedmaGame, Parameters<typeof sedmaApi.exec>, SedmaView> = {
  key: 'sedma',
  useGame: useSedmaGame,
  useView: useSedmaView,
  tutorial: SEDMA_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 8 },
  cli: () => ({
    parseCommand: parseSedmaCommand,
    formatResponse: formatSedmaState,
    helpText: SEDMA_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === SedmaPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const humanTeam = humanIdx % 2;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerTeam === humanTeam,
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
          value: game.sedmaConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetPoints',
          label: t('settings.targetPoints'),
          value: game.sedmaConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'sedma-action-buttons',
  resetTutorialId: 'sedma-reset-button',
  AboveContent,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Sedma game page: a Czech/Slovak 4-player (2 vs 2) no-trump capture trick-taker. */
export const SedmaPage = createGamePage(sedmaDefinition);
