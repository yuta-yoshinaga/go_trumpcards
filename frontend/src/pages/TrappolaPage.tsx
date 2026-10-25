import type { trappolaApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useTrappolaGame } from '../hooks/useTrappolaGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { TrappolaResponse } from '../types/card';
import { TrappolaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTrappolaCommand, TRAPPOLA_HELP } from '../utils/cli/commands/trappolaCommands';
import { formatTrappolaState } from '../utils/cli/formatters/trappolaFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Trappola tutorial step definitions. */
const TR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tr-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-point-info"]',
    messageKey: 'tutorial.pointInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TRAPPOLA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TrappolaPhase.PLAY]: 'play',
  [TrappolaPhase.TRICK_END]: 'trickEnd',
  [TrappolaPhase.ROUND_END]: 'roundEnd',
  [TrappolaPhase.GAME_END]: 'gameEnd',
};

const TEAM_LABELS = ['A', 'B'] as const;

/**
 * Clamps a team's round-thirds count to the [0, 3] range used by the 3-dot
 * indicator. Thirds reset to 0 once they reach 3 (converted to a point), but
 * this guards defensively against any out-of-range value.
 */
function thirdsFilled(thirds: number): number {
  return Math.min(Math.max(thirds, 0), 3);
}

type TrappolaGame = ReturnType<typeof useTrappolaGame>;
type TrappolaView = ReturnType<typeof useTrappolaView>;
type TrappolaContext = GamePageContext<TrappolaResponse, TrappolaGame, TrappolaView>;

function useTrappolaView() {
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('trappola', TRAPPOLA_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, tc, view }: TrappolaContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isTrickEnd = state.phase === TrappolaPhase.TRICK_END;
  const isRoundEnd = state.phase === TrappolaPhase.ROUND_END;
  const isGameEnd = state.phase === TrappolaPhase.GAME_END || state.gameEndFlag;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span>{t('trick', { n: state.trickNumber })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="tr-trick-display"
          />
          <div className="mt-2 text-center text-sm text-ds-text-primary">
            {(isTrickEnd || isRoundEnd || isGameEnd) && state.lastTrick.length > 0 && state.lastTrickWinner >= 0
              ? t('resolvedTrickPoints', {
                  team: TEAM_LABELS[state.lastTrickWinner % 2],
                  thirds: state.lastTrickThirds,
                })
              : t('currentTrickPoints', { thirds: state.currentTrickThirds })}
          </div>

          {/* Previous trick reviewer: lets the player recount the just-completed trick */}
          <details className="mb-2 p-2 rounded bg-black/30" data-testid="tr-previous-trick">
            <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('previousTrick')}</summary>
            <div className="mt-1">
              {state.lastTrick.length > 0 ? (
                <TrickDisplay
                  currentTrick={state.lastTrick}
                  players={state.players}
                  cardWidth={Math.round(cardWidth * 0.7)}
                  label={
                    state.lastTrickWinner >= 0
                      ? t('previousTrickWinner', {
                          name: playerName(
                            state.lastTrickWinner,
                            state.players[state.lastTrickWinner]?.isHuman === true,
                          ),
                        })
                      : t('previousTrick')
                  }
                  winnerIdx={state.lastTrickWinner >= 0 ? state.lastTrickWinner : undefined}
                />
              ) : (
                <div className="text-ds-text-muted text-sm">{t('previousTrickEmpty')}</div>
              )}
            </div>
          </details>
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* CPU players */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
              </summary>
              <div className="mt-1">
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      {playerName(p.id, p.isHuman)} [{t('teamLabel', { team: TEAM_LABELS[p.teamId] })}]:{' '}
                      {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                    </div>
                  ))}
              </div>
            </details>
          ) : (
            state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm">
                    {playerName(p.id, p.isHuman)} [{t('teamLabel', { team: TEAM_LABELS[p.teamId] })}]:{' '}
                    {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                  </div>
                </div>
              ))
          )}

          {/* Team score table */}
          <div className="my-3 p-2 rounded bg-black/30" data-tutorial="tr-score-table">
            <div className="flex justify-between text-ds-text-muted text-sm mb-1">
              <span>{t('scores')}</span>
              <span>{t('targetPointsLabel', { points: state.config.targetPoints })}</span>
            </div>
            <table className="w-full text-sm text-ds-text-muted">
              <thead>
                <tr>
                  <th scope="col" className="text-left">
                    {t('scoresTeam')}
                  </th>
                  <th scope="col">{t('scoresPoints')}</th>
                  <th scope="col">{t('scoresThirds')}</th>
                </tr>
              </thead>
              <tbody>
                {TEAM_LABELS.map((label, idx) => (
                  <tr key={label} className={humanPlayer && humanPlayer.teamId === idx ? 'text-ds-accent' : ''}>
                    <th scope="row" className="text-left font-normal">
                      {t('teamLabel', { team: label })}
                    </th>
                    <td className="text-center">{state.teamScores[idx] ?? 0}</td>
                    <td className="text-center">
                      <span
                        className="inline-flex gap-0.5 align-middle"
                        title={t('thirdsTooltip', { n: 3 - thirdsFilled(state.teamRoundThirds[idx]) })}
                        data-testid={`tr-thirds-${idx.toString()}`}
                      >
                        {[0, 1, 2].map((d) => (
                          <span
                            key={`tr-dot-${idx.toString()}-${d.toString()}`}
                            aria-hidden="true"
                            className={`inline-block w-2 h-2 rounded-full ${
                              d < thirdsFilled(state.teamRoundThirds[idx])
                                ? 'bg-ds-accent'
                                : 'border border-ds-border-subtle'
                            }`}
                          />
                        ))}
                        <span className="sr-only">
                          {t('thirdsAria', { filled: thirdsFilled(state.teamRoundThirds[idx]) })}
                        </span>
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Card-point legend (static reference) */}
          <details className="my-3 p-2 rounded bg-black/30" data-testid="tr-point-legend">
            <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
              {t('pointLegend.title')}
            </summary>
            <div className="mt-1 text-ds-text-muted text-xs">
              <table className="w-full">
                <thead>
                  <tr>
                    <th scope="col" className="text-left font-normal">
                      {t('pointLegend.cardCol')}
                    </th>
                    <th scope="col" className="text-right font-normal">
                      {t('pointLegend.pointCol')}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td>{t('pointLegend.ace')}</td>
                    <td className="text-right">{t('pointLegend.aceValue')}</td>
                  </tr>
                  <tr>
                    <td>{t('pointLegend.figures')}</td>
                    <td className="text-right">{t('pointLegend.figuresValue')}</td>
                  </tr>
                  <tr>
                    <td>{t('pointLegend.others')}</td>
                    <td className="text-right">{t('pointLegend.othersValue')}</td>
                  </tr>
                  <tr>
                    <td>{t('pointLegend.lastTrick')}</td>
                    <td className="text-right">{t('pointLegend.lastTrickValue')}</td>
                  </tr>
                </tbody>
              </table>
              <div className="mt-1">{t('pointLegend.note')}</div>
            </div>
          </details>
        </div>
      </div>

      {/* Message */}
      <div data-tutorial="tr-point-info">
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      </div>
    </>
  );
}

function FooterHand({ game, state, t, view }: TrappolaContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.phase === TrappolaPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;

  return humanPlayer ? (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="tr"
      highlightIndices={game.hint?.cardIndices}
      validIndices={isHumanTurn ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
    />
  ) : null;
}

function FooterStatus({ game, t, frontendHint }: TrappolaContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="trappola-hint-live" role="status" aria-live="polite">
        {game.hint && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {game.hint.cardIndices.map((i) => `[${i}]`).join(', ')} (
            {t(`hintReason.${game.hint.reason}`)})
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, tc, busy }: TrappolaContext) {
  const isHumanTurn = state.phase === TrappolaPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isTrickEnd = state.phase === TrappolaPhase.TRICK_END;
  const isRoundEnd = state.phase === TrappolaPhase.ROUND_END;

  return (
    <>
      {isHumanTurn && (
        <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
          {tc('button.hint')}
        </button>
      )}
      {isHumanTurn && (
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

const trappolaDefinition: GameDefinition<
  TrappolaResponse,
  TrappolaGame,
  Parameters<typeof trappolaApi.exec>,
  TrappolaView
> = {
  key: 'trappola',
  useGame: useTrappolaGame,
  useView: useTrappolaView,
  tutorial: TR_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseTrappolaCommand,
    formatResponse: formatTrappolaState,
    helpText: TRAPPOLA_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === TrappolaPhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ state }) => ({
    isHumanTurn: state.phase === TrappolaPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true,
  }),
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.trappolaConfig.cpuDifficulty,
      targetPoints: game.trappolaConfig.targetPoints,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.trappolaConfig.cpuDifficulty,
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
          value: game.trappolaConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  cardNav: ({ game, state }) => {
    const isHumanTurn = state.phase === TrappolaPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: game.handlePlay,
      onClear: game.clearSelection,
      enabled: isHumanTurn,
    };
  },
  shortcutsTestId: 'trappola-kbd-shortcuts',
  actionsClassName: 'flex gap-2 items-center',
  actionsTutorialId: 'tr-play-button',
  resetTutorialId: 'tr-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Trappola game page: no-trump must-follow trick play with team scoring. */
export const TrappolaPage = createGamePage(trappolaDefinition);
