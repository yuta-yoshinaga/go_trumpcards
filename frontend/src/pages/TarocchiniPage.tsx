import { useEffect } from 'react';
import type { tarocchiniApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, useTarocchiniGame } from '../hooks/useTarocchiniGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { TAROCCHINI_SURPLUS, type TarocchiniResponse } from '../types/card';
import { TarocchiniPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTarocchiniCommand, TAROCCHINI_HELP } from '../utils/cli/commands/tarocchiniCommands';
import { formatTarocchiniState } from '../utils/cli/formatters/tarocchiniFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** 最終トリックボーナス。domain の TarocchiniLastTrickBonus と同じ値。 */
const TAROCCHINI_LAST_TRICK_BONUS = 2;

/** Tarocchini tutorial step definitions. */
const TAROCCHINI_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tarocchini-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="tarocchini-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tarocchini-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tarocchini-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TAROCCHINI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TarocchiniPhase.SCARTO]: 'scarto',
  [TarocchiniPhase.PLAY]: 'play',
  [TarocchiniPhase.TRICK_END]: 'trickEnd',
  [TarocchiniPhase.ROUND_END]: 'roundEnd',
  [TarocchiniPhase.GAME_END]: 'gameEnd',
};

type TarocchiniGame = ReturnType<typeof useTarocchiniGame>;
type TarocchiniView = ReturnType<typeof useTarocchiniView>;
type TarocchiniContext = GamePageContext<TarocchiniResponse, TarocchiniGame, TarocchiniView>;

function useTarocchiniView({ game }: { game: TarocchiniGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('tarocchini', TAROCCHINI_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, view }: TarocchiniContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;
  const isRoundEnd = state.phase === TarocchiniPhase.ROUND_END;
  const isGameEnd = state.phase === TarocchiniPhase.GAME_END || state.gameEndFlag;
  // チーム分けはサーバの player.team を使う (席番号から計算し直さない)。
  const teamOfSeat = (seat: number) => state.players.find((p) => p.id === seat)?.team ?? null;
  const lastTrickTeam = state.lastTrickWinner >= 0 ? teamOfSeat(state.lastTrickWinner) : null;
  const dealerTeam = teamOfSeat(state.dealerIdx);

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="tarocchini-info">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>{t('target', { rounds: state.config.targetRounds })}</span>
      </div>

      {/*
        The papi rule is the one thing a player cannot infer from the cards,
        so it gets a permanent line rather than a tooltip: four trumps rank
        equal, and the LATER-played one takes the trick.
      */}
      <div className="text-center mb-2 text-sm font-semibold text-ds-warning" data-testid="tarocchini-papi-note">
        {t('papiNote')}
      </div>

      <div className={lgTwoColGrid}>
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="tarocchini-trick-display"
          />
        </div>

        <div>
          {/* Team scores — this is a team game, not an individual one. */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="tarocchini-team-scores">
            {state.teamScores.map((score, team) => (
              <div key={team} className={`py-0.5 ${team === humanTeam ? 'text-ds-text-primary font-semibold' : ''}`}>
                {t('teamScore', { n: team, score })}
              </div>
            ))}
          </div>

          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}
                    {p.isDealer ? ` / ${t('dealerBadge')}` : ''}): {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}
                  {p.isDealer ? ` / ${t('dealerBadge')}` : ''}): {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })}
                </div>
              ))}
            </div>
          )}

          {state.scartoCount > 0 && (
            <div className="mb-2 text-ds-text-muted text-sm" data-testid="tarocchini-scarto-done">
              {t('scartoDone', { count: state.scartoCount })}
            </div>
          )}

          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              {state.players.map((p) => (
                <div key={p.id}>
                  {t('roundResult.tricks', {
                    name: playerName(p.id, p.isHuman),
                    count: state.roundTricks[p.id] ?? 0,
                  })}
                </div>
              ))}
              <div className="mb-1 text-ds-text-primary">{t('roundResult.breakdownTitle')}</div>
              {state.roundBreakdown.map((breakdown, team) => (
                <div key={`round-breakdown-${team}`} data-testid={`tarocchini-team-round-breakdown-${team}`}>
                  {t('roundResult.breakdown', { team, ...breakdown })}
                </div>
              ))}
              {lastTrickTeam !== null && (
                <div data-testid="tarocchini-last-trick-bonus">
                  {t('roundResult.lastTrick', {
                    team: lastTrickTeam,
                    bonus: TAROCCHINI_LAST_TRICK_BONUS,
                  })}
                </div>
              )}
              {state.scartoCount > 0 && dealerTeam !== null && (
                <div data-testid="tarocchini-scarto-bonus">
                  {t('roundResult.scarto', { team: dealerTeam, count: state.scartoCount })}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: TarocchiniContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="tarocchini-hint-live" role="status" aria-live="polite">
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

function FooterHand({ game, state, t, view }: TarocchiniContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isScartoPhase = state.phase === TarocchiniPhase.SCARTO;
  const isPlayPhase = state.phase === TarocchiniPhase.PLAY;
  const canPlay = isPlayPhase && isHumanTurn;
  const canScarto = isScartoPhase && state.isHumanScarto;
  const scartoValidIndices =
    canScarto && humanPlayer
      ? humanPlayer.cards.flatMap((card, idx) => (card.design !== 'JOKER' ? [idx] : []))
      : undefined;

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="tarocchini-prompt-live" role="status" aria-live="polite">
        {canScarto && (
          <div className="mb-1 text-center text-sm text-ds-text-muted" data-testid="tarocchini-scarto-prompt">
            {t('scartoPrompt', { count: game.selectedCardIndices.length, total: TAROCCHINI_SURPLUS })}
          </div>
        )}
      </div>

      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          dataTutorialPrefix="tarocchini"
          stackRows
          validIndices={canPlay ? state.playableIndices : scartoValidIndices}
          restrictedTooltip={canScarto ? t('scartoRestricted') : t('playButton')}
        />
      )}
    </>
  );
}

function primaryActions({ game, state, t, busy }: TarocchiniContext) {
  const isScartoPhase = state.phase === TarocchiniPhase.SCARTO;
  const isPlayPhase = state.phase === TarocchiniPhase.PLAY;
  const isTrickEnd = state.phase === TarocchiniPhase.TRICK_END;
  const isRoundEnd = state.phase === TarocchiniPhase.ROUND_END;
  const canPlay = isPlayPhase && state.isHumanTurn;
  const canScarto = isScartoPhase && state.isHumanScarto;

  if (!canScarto && !canPlay && !isTrickEnd && !isRoundEnd) return null;

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="tarocchini-action-buttons">
      {canScarto && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handleScarto}
          disabled={busy || game.selectedCardIndices.length !== TAROCCHINI_SURPLUS}
        >
          {t('scartoButton')}
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
      {/* 同格のパパ 4 枚をどう使うかがこのゲームの難所。CUI と CLI からは
        呼べるのに、盤面には要求する手段が無かった (#4820)。 */}
      {(canScarto || canPlay) && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleHint}
          disabled={busy}
          data-testid="tarocchini-hint-button"
        >
          {t('hintButton')}
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

const tarocchiniDefinition: GameDefinition<
  TarocchiniResponse,
  TarocchiniGame,
  Parameters<typeof tarocchiniApi.exec>,
  TarocchiniView
> = {
  key: 'tarocchini',
  useGame: useTarocchiniGame,
  useView: useTarocchiniView,
  tutorial: TAROCCHINI_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 15 },
  cli: () => ({
    parseCommand: parseTarocchiniCommand,
    formatResponse: formatTarocchiniState,
    helpText: TAROCCHINI_HELP,
  }),
  phase: ({ state, view }) => {
    const humanPlayer = state.players.find((p) => p.isHuman);
    const humanTeam = humanPlayer?.team ?? 0;
    const isGameEnd = state.phase === TarocchiniPhase.GAME_END || state.gameEndFlag;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerTeam === humanTeam,
    };
  },
  shell: ({ state }) => {
    const isScartoPhase = state.phase === TarocchiniPhase.SCARTO;
    const canScarto = isScartoPhase && state.isHumanScarto;
    const isGameEnd = state.phase === TarocchiniPhase.GAME_END || state.gameEndFlag;
    return {
      isHumanTurn: (state.isHumanTurn || canScarto) && !isGameEnd,
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
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.tarocchiniConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'number',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.tarocchiniConfig.targetRounds,
          min: 4,
          step: 4,
          description: t('settings.targetRoundsGuide'),
          onSelect: (v) => game.handleConfigChange('targetRounds', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  footerTall: true,
  resetTutorialId: 'tarocchini-reset-button',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Tarocchini game page: a Bolognese 4-player 2v2 tarot trick-taker. */
export const TarocchiniPage = createGamePage(tarocchiniDefinition);
