import { useEffect } from 'react';
import type { minchiateApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useMinchiateGame } from '../hooks/useMinchiateGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { MINCHIATE_SURPLUS, type MinchiateResponse } from '../types/card';
import { MinchiatePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MINCHIATE_HELP, parseMinchiateCommand } from '../utils/cli/commands/minchiateCommands';
import { formatMinchiateState } from '../utils/cli/formatters/minchiateFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Minchiate tutorial step definitions. */
const MINCHIATE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="minchiate-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="minchiate-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="minchiate-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="minchiate-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MINCHIATE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MinchiatePhase.SCARTO]: 'scarto',
  [MinchiatePhase.PLAY]: 'play',
  [MinchiatePhase.TRICK_END]: 'trickEnd',
  [MinchiatePhase.ROUND_END]: 'roundEnd',
  [MinchiatePhase.GAME_END]: 'gameEnd',
};

/**
 * Points the last trick is worth, mirroring `MinchiateLastTrickBonus` in
 * `internal/domain/Minchiate.go`. `frontend/scripts/check-minchiate-last-trick.mjs`
 * fails the build if the two drift apart.
 */
const MINCHIATE_LAST_TRICK_BONUS = 3;

type MinchiateGame = ReturnType<typeof useMinchiateGame>;
type MinchiateView = ReturnType<typeof useMinchiateView>;
type MinchiateContext = GamePageContext<MinchiateResponse, MinchiateGame, MinchiateView>;

function useMinchiateView({ game }: { game: MinchiateGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('minchiate', MINCHIATE_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, view }: MinchiateContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;

  const isRoundEnd = state.phase === MinchiatePhase.ROUND_END;
  const isGameEnd = state.phase === MinchiatePhase.GAME_END || state.gameEndFlag;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="minchiate-info">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>{t('target', { rounds: state.config.targetRounds })}</span>
      </div>

      {/*
        40 trumps rather than the usual 21 is the thing a player cannot
        eyeball from the hand, and "how many still outrank me" is the whole
        read — so it gets a permanent line rather than a tooltip.
      */}
      <div className="text-center mb-2 text-sm font-semibold text-ds-warning" data-testid="minchiate-trump-count-note">
        {t('trumpCountNote')}
      </div>
      <section
        aria-label={t('trumpTracker')}
        data-testid="minchiate-trump-tracker"
        className="mb-2 text-center text-sm"
      >
        <div className="text-ds-text-primary">{t('remainingTrumps', { count: state.remainingTrumps })}</div>
        <div className="text-ds-text-muted">
          {state.playedTrumps.length > 0
            ? state.playedTrumps.map(cardAlt).join(t('listSeparator'))
            : t('noPlayedTrumps')}
        </div>
      </section>
      {/* マットだけが規則の外にいる (取らない・フォロー免除・リードを定めない)。
          ヒントを切っていると伝わる場所がどこにも無かった (#5715)。 */}
      <div className="text-center mb-2 text-sm text-ds-text-muted" data-testid="minchiate-matto-note">
        {t('mattoNote')}
      </div>

      <div className={lgTwoColGrid}>
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="minchiate-trick-display"
          />
        </div>

        <div>
          {/* Team scores — this is a team game, not an individual one. */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="minchiate-team-scores">
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
            <div className="mb-2 text-ds-text-muted text-sm" data-testid="minchiate-scarto-done">
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
              <div className="mt-2 mb-1 text-ds-text-primary">{t('roundResult.breakdownTitle')}</div>
              {state.roundBreakdown.map((breakdown, team) => (
                <div key={team} data-testid={`mc-round-breakdown-${team}`}>
                  {t('roundResult.teamBreakdown', {
                    team,
                    tricks: breakdown.tricks,
                    lastTrickBonus: breakdown.lastTrickBonus,
                    scartoBonus: breakdown.scartoBonus,
                    total: breakdown.total,
                  })}
                </div>
              ))}
              {/* **トリック数を足しても teamScores の増分と合わない。**精算には
                  最終トリックボーナスとスカルト枚数分が乗っているのに、どちらの
                  画面にも出ていなかった (#6512)。検算できないのは内訳が足りないから。 */}
              {state.lastTrickWinner >= 0 && (
                <div data-testid="mc-last-trick-bonus">
                  {t('roundResult.lastTrick', {
                    team: state.lastTrickWinner % 2,
                    points: MINCHIATE_LAST_TRICK_BONUS,
                  })}
                </div>
              )}
              {state.scartoCount > 0 && (
                <div data-testid="mc-scarto-bonus">
                  {t('roundResult.scarto', {
                    team: state.dealerIdx % 2,
                    points: state.scartoCount,
                  })}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: MinchiateContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="minchiate-hint-live" role="status" aria-live="polite">
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

function primaryActions({ game, state, t, busy }: MinchiateContext) {
  const isHumanTurn = state.isHumanTurn;
  const isScartoPhase = state.phase === MinchiatePhase.SCARTO;
  const isPlayPhase = state.phase === MinchiatePhase.PLAY;
  const isTrickEnd = state.phase === MinchiatePhase.TRICK_END;
  const isRoundEnd = state.phase === MinchiatePhase.ROUND_END;

  const canPlay = isPlayPhase && isHumanTurn;
  const canScarto = isScartoPhase && state.isHumanScarto;

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="minchiate-action-buttons">
      {canScarto && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handleScarto}
          disabled={busy || game.selectedCardIndices.length !== MINCHIATE_SURPLUS}
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
      {/* CUI と CLI からは呼べるのに、盤面には要求する手段が無かった (#4819)。 */}
      {(canScarto || canPlay) && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleHint}
          disabled={busy}
          data-testid="minchiate-hint-button"
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

function FooterHand({ game, state, t, view }: MinchiateContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isScartoPhase = state.phase === MinchiatePhase.SCARTO;
  const isPlayPhase = state.phase === MinchiatePhase.PLAY;
  const canPlay = isPlayPhase && isHumanTurn;
  const canScarto = isScartoPhase && state.isHumanScarto;

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div className={`text-center text-sm ${canScarto ? 'mb-1' : ''}`}>
        <span data-testid="minchiate-prompt-live" role="status" aria-live="polite">
          {canScarto && (
            <span className="text-ds-text-muted" data-testid="minchiate-scarto-prompt">
              {t('scartoPrompt', { count: MINCHIATE_SURPLUS })}
            </span>
          )}
        </span>
        {canScarto && (
          <span className="ml-2 text-ds-text-primary tabular-nums" data-testid="minchiate-scarto-progress">
            {t('scartoProgress', { selected: game.selectedCardIndices.length, count: MINCHIATE_SURPLUS })}
          </span>
        )}
      </div>

      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={cardWidth}
          isMobile={isMobile}
          dataTutorialPrefix="minchiate"
          stackRows
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
        />
      )}
    </>
  );
}

const minchiateDefinition: GameDefinition<
  MinchiateResponse,
  MinchiateGame,
  Parameters<typeof minchiateApi.exec>,
  MinchiateView
> = {
  key: 'minchiate',
  useGame: useMinchiateGame,
  useView: useMinchiateView,
  tutorial: MINCHIATE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 21 },
  cli: () => ({
    parseCommand: parseMinchiateCommand,
    formatResponse: formatMinchiateState,
    helpText: MINCHIATE_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === MinchiatePhase.GAME_END || state.gameEndFlag;
    const humanPlayer = state.players.find((p) => p.isHuman);
    const humanTeam = humanPlayer?.team ?? 0;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerTeam === humanTeam,
    };
  },
  shell: ({ state }) => {
    const isScartoPhase = state.phase === MinchiatePhase.SCARTO;
    const isGameEnd = state.phase === MinchiatePhase.GAME_END || state.gameEndFlag;
    const canScarto = isScartoPhase && state.isHumanScarto;
    return {
      isHumanTurn: (state.isHumanTurn || canScarto) && !isGameEnd,
    };
  },
  reset: ({ game }) => {
    game.reset();
  },
  settingsPlacement: 'top',
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.minchiateConfig.cpuDifficulty,
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
          value: game.minchiateConfig.targetRounds,
          options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetRounds', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  // Tarot-size hands outgrow the 45vh cap on phones (#11655).
  footerTall: true,
  message: (ctx) =>
    !ctx.view.isMobile && (
      <GameMessageBox
        message={ctx.state.message}
        messageCode={ctx.state.messageCode}
        messageParams={ctx.state.messageParams}
      />
    ),
  resetTutorialId: 'minchiate-reset-button',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Minchiate game page: a Bolognese 4-player 2v2 tarot trick-taker. */
export const MinchiatePage = createGamePage(minchiateDefinition);
