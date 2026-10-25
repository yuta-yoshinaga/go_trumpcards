import type { madrassoApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useMadrassoGame } from '../hooks/useMadrassoGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { MadrassoResponse } from '../types/card';
import { MadrassoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MADRASSO_HELP, parseMadrassoCommand } from '../utils/cli/commands/madrassoCommands';
import { formatMadrassoState } from '../utils/cli/formatters/madrassoFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit glyphs by design value (1=Spade … 4=Diamond). */
const SUIT_GLYPHS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Madrasso tutorial step definitions. */
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

const MADRASSO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MadrassoPhase.PLAY]: 'play',
  [MadrassoPhase.TRICK_END]: 'trickEnd',
  [MadrassoPhase.ROUND_END]: 'roundEnd',
  [MadrassoPhase.GAME_END]: 'gameEnd',
};

/**
 * Clamps a team's round score for display.
 *
 * **Madrasso scores in whole points, not thirds.** The clone source (Tressette)
 * counts 1/3 points and shows a 3-dot indicator that fills and resets; here the
 * round total is simply the card points taken so far, out of
 * {@link MADRASSO_ROUND_POINTS}.
 */
function roundPointsShown(points: number): number {
  return Math.min(Math.max(points, 0), MADRASSO_ROUND_POINTS);
}

/** Total card points contested in one deal (sync: MadrassoRoundPoints in Go). */
const MADRASSO_ROUND_POINTS = 121;

type MadrassoGame = ReturnType<typeof useMadrassoGame>;
type MadrassoView = ReturnType<typeof useMadrassoView>;
type MadrassoContext = GamePageContext<MadrassoResponse, MadrassoGame, MadrassoView>;

function useMadrassoView({ game }: { game: MadrassoGame }) {
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('madrasso', MADRASSO_PHASE_KEYS);
  const { state } = game;
  const isHumanTurn = state?.phase === MadrassoPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCount = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;
  return { cardWidth, isMobile, phaseNames, isHumanTurn, humanCardCount };
}

function Board({ state, t, tc, view }: MadrassoContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const messageBox = (
    <div data-tutorial="tr-point-info">
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </div>
  );
  const isTrickEnd = state.phase === MadrassoPhase.TRICK_END;
  const trumpSymbol = SUIT_GLYPHS[state.trumpSuit] ?? t('noTrump');
  const teamLabels = ['A', 'B'];
  const currentTrickPoints = state.currentTrick.reduce((sum, play) => sum + play.card.points, 0);
  const trickWinnerTeam = isTrickEnd ? state.players[state.lastTrickWinner]?.teamId : undefined;
  return (
    <>
      <div
        className="text-ds-text-primary text-center mb-2 flex flex-wrap justify-center gap-x-4 gap-y-1"
        data-testid="madrasso-status-row"
      >
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span data-testid="madrasso-trump" className="whitespace-nowrap">
          {t('trumpSuit', { suit: trumpSymbol })}
        </span>
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
            cardDetailFor={(card) => t('cardPoints', { points: card.points })}
          />
          <div className="mb-2 text-center text-sm text-ds-text-primary" data-testid="madrasso-trick-points">
            {trickWinnerTeam === undefined
              ? t('currentTrickPoints', { points: currentTrickPoints })
              : t('wonTrickPoints', {
                  team: teamLabels[trickWinnerTeam],
                  points: currentTrickPoints,
                })}
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
                  cardDetailFor={(card) => t('cardPoints', { points: card.points })}
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
                      {playerName(p.id, p.isHuman)} [{t('teamLabel', { team: teamLabels[p.teamId] })}]:{' '}
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
                    {playerName(p.id, p.isHuman)} [{t('teamLabel', { team: teamLabels[p.teamId] })}]:{' '}
                    {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                  </div>
                </div>
              ))
          )}

          {/* Team score table */}
          <div className="my-3 p-2 rounded bg-black/30" data-tutorial="tr-score-table">
            <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
            <table className="w-full text-sm text-ds-text-muted">
              <thead>
                <tr>
                  <th scope="col" className="text-left">
                    {t('scoresTeam')}
                  </th>
                  <th scope="col" className="whitespace-nowrap text-[10px]">
                    {t('scoresPoints', { target: state.config.targetPoints })}
                  </th>
                  <th scope="col" className="whitespace-nowrap text-[10px]">
                    {t('scoresRoundPoints')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {teamLabels.map((label, idx) => (
                  <tr key={label} className={humanPlayer && humanPlayer.teamId === idx ? 'text-ds-accent' : ''}>
                    <th scope="row" className="text-left font-normal whitespace-nowrap">
                      {t('teamLabel', { team: label })}
                    </th>
                    <td className="text-center">{state.teamScores[idx] ?? 0}</td>
                    <td className="text-center" data-testid={`tr-round-points-${idx.toString()}`}>
                      {t('roundPointsOf', {
                        points: roundPointsShown(state.teamRoundPoints[idx] ?? 0),
                        total: MADRASSO_ROUND_POINTS,
                      })}
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

      {isMobile && messageBox}
    </>
  );
}

function FooterHand({ state, t, view, game }: MadrassoContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.phase === MadrassoPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  return (
    <>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          dataTutorialPrefix="tr"
          cardStatusFor={(idx) => t('cardPoints', { points: humanPlayer.cards[idx].points })}
          validIndices={isHumanTurn ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
        />
      )}
    </>
  );
}

function FooterStatus({ game, t, frontendHint }: MadrassoContext) {
  const { hint } = game;
  const { hintEnabled } = frontendHint;
  return (
    <>
      {/*
            ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
            領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
            られないことがある (#5955)。
          */}
      <div data-testid="madrasso-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {hint.cardIndices.map((i) => `[${i}]`).join(', ')} ({t(`hintReason.${hint.reason}`)})
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ state, game, t, tc, busy }: MadrassoContext) {
  const isHumanTurn = state.phase === MadrassoPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isTrickEnd = state.phase === MadrassoPhase.TRICK_END;
  const isRoundEnd = state.phase === MadrassoPhase.ROUND_END;
  const { handlePlay, handleHint, handleNextTrick, handleNextRound, hintLoading, selectedCardIndices } = game;
  const loading = busy;
  return (
    <>
      {isHumanTurn && (
        <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
          {tc('button.hint')}
        </button>
      )}
      {isHumanTurn && (
        <button
          type="button"
          className={btnPrimary}
          onClick={handlePlay}
          disabled={loading || selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const definition: GameDefinition<MadrassoResponse, MadrassoGame, Parameters<typeof madrassoApi.exec>, MadrassoView> = {
  key: 'madrasso',
  useGame: useMadrassoGame,
  useView: useMadrassoView,
  tutorial: TR_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseMadrassoCommand,
    formatResponse: formatMadrassoState,
    helpText: MADRASSO_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === MadrassoPhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ view }) => ({ isHumanTurn: view.isHumanTurn }),
  cardNav: ({ game, view }) => ({
    cardCount: view.humanCardCount,
    onToggle: game.toggleCard,
    onConfirm: game.handlePlay,
    onClear: game.clearSelection,
    enabled: view.isHumanTurn && !game.loading,
  }),
  settingsPlacement: 'top',
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.madrassoConfig.cpuDifficulty,
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
          value: game.madrassoConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex gap-2 items-center',
  actionsTutorialId: 'tr-play-button',
  resetTutorialId: 'tr-reset-button',
  shortcutsTestId: 'madrasso-kbd-shortcuts',
  message: ({ view, state }) => {
    if (view.isMobile) return undefined;
    return (
      <div data-tutorial="tr-point-info">
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      </div>
    );
  },
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.madrassoConfig.cpuDifficulty,
      targetPoints: game.madrassoConfig.targetPoints,
    });
  },
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/**
 * Renders the Madrasso game page: must-follow trick play with a trump decided by
 * the deal, integer card points, and team scoring.
 */
export const MadrassoPage = createGamePage(definition);
