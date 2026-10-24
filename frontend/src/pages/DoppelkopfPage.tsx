import { useEffect } from 'react';
import type { doppelkopfApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_CHIPS_OPTIONS, useDoppelkopfGame } from '../hooks/useDoppelkopfGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { DoppelkopfResponse } from '../types/card';
import { DoppelkopfPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { DOPPELKOPF_HELP, parseDoppelkopfCommand } from '../utils/cli/commands/doppelkopfCommands';
import { formatDoppelkopfState } from '../utils/cli/formatters/doppelkopfFormatter';
import { DOPPELKOPF_TRUMP_ORDER, isDoppelkopfTrump } from '../utils/doppelkopfTrump';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Re が勝つのに必要なカード得点。`domain.DoppelkopfReWinPoints` の写し。 */
const DOPPELKOPF_RE_WIN_POINTS = 121;
/** 1 ラウンドのカード得点合計。`domain.DoppelkopfTotalPoints` の写し。 */
const DOPPELKOPF_TOTAL_POINTS = 240;
type DoppelkopfGame = ReturnType<typeof useDoppelkopfGame>;
type DoppelkopfView = ReturnType<typeof useDoppelkopfView>;
type DoppelkopfContext = GamePageContext<DoppelkopfResponse, DoppelkopfGame, DoppelkopfView>;

const DOPPELKOPF_PHASE_KEYS: Readonly<Record<number, string>> = {
  [DoppelkopfPhase.PLAY]: 'play',
  [DoppelkopfPhase.TRICK_END]: 'trickEnd',
  [DoppelkopfPhase.ROUND_END]: 'roundEnd',
  [DoppelkopfPhase.GAME_END]: 'gameEnd',
};

function useDoppelkopfView({ game }: { game: DoppelkopfGame }) {
  // biome-ignore lint/correctness/useExhaustiveDependencies: Match the legacy mount-time reset.
  useEffect(() => {
    game.reset();
  }, []);
  const { cardWidth, isMobile } = useCardDimensions();
  const isMobileViewport = useIsMobile();
  const phaseNames = usePhaseNames('doppelkopf', DOPPELKOPF_PHASE_KEYS);
  return { cardWidth, isMobile, isMobileViewport, phaseNames };
}

/** Doppelkopf tutorial step definitions. */
const DK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dk-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="dk-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="dk-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

function Board({ state, t, view }: DoppelkopfContext) {
  const isRoundEnd = state.phase === DoppelkopfPhase.ROUND_END;
  const isGameEnd = state.phase === DoppelkopfPhase.GAME_END || state.gameEndFlag;
  const teamLabel = state.youAreRe ? t('team.re') : t('team.kontra');
  const { cardWidth, isMobile, isMobileViewport } = view;
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
            dataTutorial="dk-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="dk-info">
          {/* Trump ordering legend (collapsible) */}
          <details className="mb-1 sm:mb-2 p-1 sm:p-2 rounded bg-black/30" data-testid="dk-trump-legend">
            <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('rules')}</summary>
            <div className="mt-1 text-ds-text-muted text-xs">
              <div className="mb-1 text-ds-text-primary">{t('trumpLegend.title')}</div>
              <div className="mb-1">{t('trumpLegend.caption')}</div>
              <div className="flex flex-wrap items-center gap-x-1 gap-y-0.5">
                {DOPPELKOPF_TRUMP_ORDER.map((symbol, i) => (
                  <span key={symbol} className="inline-flex items-center gap-1">
                    <span className="font-mono text-ds-text-primary">{symbol}</span>
                    {i < DOPPELKOPF_TRUMP_ORDER.length - 1 && <span aria-hidden="true">&gt;</span>}
                  </span>
                ))}
              </div>
            </div>
          </details>

          {/* Your team / announcements */}
          <div className="mb-1 sm:mb-2 p-1 sm:p-2 rounded bg-black/30 text-ds-text-muted text-sm flex flex-wrap items-center gap-x-2 sm:block">
            <span className="sm:block">
              {t('yourTeam')}: {teamLabel}
              {state.soloRe && ` (${t('soloRe')})`}
            </span>
            {(state.reAnnounced || state.kontraAnnounced) && (
              <span className="sm:block">
                {t('announced')}:{' '}
                {[state.reAnnounced && t('team.re'), state.kontraAnnounced && t('team.kontra')]
                  .filter(Boolean)
                  .join(', ')}
              </span>
            )}
          </div>

          {/* Players: chips / cards / tricks (team shown once revealed) */}
          <div className="mb-1 p-1 sm:p-2 rounded bg-black/30">
            {state.players
              .filter((p) => p.isHuman)
              .map((p) => (
                <div
                  key={p.id}
                  className="flex flex-wrap items-center gap-x-1 text-ds-text-muted text-sm sm:py-0.5"
                  data-testid="dk-human-player-info"
                >
                  {playerName(p.id, true)}
                  {(state.teamsRevealed || p.teamKnown) && ` [${p.isRe ? t('team.re') : t('team.kontra')}]`}:{' '}
                  {t('chips', { count: p.chips })} | {t('cards', { count: p.cardCount })} |{' '}
                  <span className="whitespace-nowrap">{t('tricks', { count: p.trickCount })}</span>
                </div>
              ))}
          </div>
          <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div
                    key={p.id}
                    className="flex flex-wrap items-center gap-x-1 text-ds-text-muted text-sm py-0.5"
                    data-testid="dk-cpu-player-info"
                  >
                    {playerName(p.id, p.isHuman)}
                    {(state.teamsRevealed || p.teamKnown) && ` [${p.isRe ? t('team.re') : t('team.kontra')}]`}:{' '}
                    {t('chips', { count: p.chips })} | {t('cards', { count: p.cardCount })} |{' '}
                    <span className="whitespace-nowrap">{t('tricks', { count: p.trickCount })}</span>
                  </div>
                ))}
            </div>
          </CpuAccordion>

          {/* ラウンド中の獲得点。**同じ値がレスポンスに毎回乗っているのに、
                    ラウンド終了まで画面に出していなかった。**最初のトリックで
                    Re/Kontra をアナウンスできるので、いま何点取れているかは
                    実戦上の判断材料になる。終了後は下の内訳が引き継ぐ。 */}
          {!(isRoundEnd || isGameEnd) && (
            <div
              className="my-1 sm:my-3 p-1 sm:p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              role="status"
              aria-live="polite"
              data-testid="dk-live-points"
            >
              <div className="mb-0.5 sm:mb-1 text-ds-text-primary">{t('livePoints.title')}</div>
              <div>
                {t('livePoints.re', {
                  points: state.liveRePoints,
                  target: DOPPELKOPF_RE_WIN_POINTS,
                })}
              </div>
              <div>
                {t('livePoints.kontra', {
                  points: state.liveKontraPoints,
                  target: DOPPELKOPF_TOTAL_POINTS - DOPPELKOPF_RE_WIN_POINTS + 1,
                })}
              </div>
            </div>
          )}

          {/* Round result */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-1 sm:my-3 p-1 sm:p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.rePoints', { points: state.roundRePoints })}</div>
              <div>{state.roundReWon ? t('roundResult.reWon') : t('roundResult.reLost')}</div>
              <div>{t('roundResult.gamePoints', { points: state.roundGamePoints })}</div>
            </div>
          )}
          {state.roundScoreHistory.length > 0 && (
            <details
              open={!isMobileViewport || undefined}
              className="my-1 sm:my-3"
              data-testid="dk-round-history-details"
            >
              <summary className="cursor-pointer select-none text-ds-text-primary text-sm">
                {t('roundHistory.title')}
              </summary>
              <div className="mt-1 overflow-x-auto" data-testid="dk-round-history">
                <table className="w-full text-sm">
                  <thead>
                    <tr>
                      <th scope="col">{t('roundHistory.round')}</th>
                      {state.players.map((p) => (
                        <th scope="col" key={p.id}>
                          {playerName(p.id, p.isHuman)}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {state.roundScoreHistory.map((scores, index) => (
                      <tr key={index}>
                        <th scope="row">{index + 1}</th>
                        {scores.map((score, playerIdx) => (
                          <td key={playerIdx} className="text-center">
                            {formatSignedDelta(score)}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </details>
          )}
        </div>
      </div>

      {isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}
    </>
  );
}

function FooterHand({ game, state, view, t }: DoppelkopfContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const canPlay = state.phase === DoppelkopfPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const trumpIndices =
    humanPlayer?.cards.reduce<number[]>((acc, card, idx) => {
      if (isDoppelkopfTrump(card)) acc.push(idx);
      return acc;
    }, []) ?? [];
  return humanPlayer ? (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="dk"
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
      trumpIndices={trumpIndices}
      trumpTitle={t('trumpLegend.badgeTitle')}
      trumpAccessibleLabel={t('trumpLegend.badgeTitle')}
    />
  ) : null;
}
function FooterStatus({ state, t, frontendHint }: DoppelkopfContext) {
  return (
    <>
      <div data-testid="doppelkopf-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndices.length > 0 && ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}
function FooterActions({ game, state, t, tc, busy }: DoppelkopfContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTurn = state.currentPlayerIdx === state.players.findIndex((p) => p.isHuman);
  const canPlay = state.phase === DoppelkopfPhase.PLAY && humanTurn;
  const isTrickEnd = state.phase === DoppelkopfPhase.TRICK_END;
  const isRoundEnd = state.phase === DoppelkopfPhase.ROUND_END;
  const teamLabel = state.youAreRe ? t('team.re') : t('team.kontra');
  const announceLabel = state.youAreRe ? t('announceRe') : t('announceKontra');
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
      {canPlay && (
        <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
          {tc('button.hint')}
        </button>
      )}
      {state.canAnnounce && (
        <>
          <span data-testid="dk-announce-stage" className="text-ds-text-muted text-sm">
            {t('announceStage', { trick: state.trickNumber, team: teamLabel, cards: humanPlayer?.cardCount ?? 0 })}
          </span>
          <button
            type="button"
            className={btnSecondary}
            onClick={game.handleAnnounce}
            disabled={busy}
            aria-label={`${announceLabel} — ${t('announceDescription')}`}
            title={t('announceDescription')}
          >
            {announceLabel}
          </button>
        </>
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
const doppelkopfDefinition: GameDefinition<
  DoppelkopfResponse,
  DoppelkopfGame,
  Parameters<typeof doppelkopfApi.exec>,
  DoppelkopfView
> = {
  key: 'doppelkopf',
  useGame: useDoppelkopfGame,
  useView: useDoppelkopfView,
  tutorial: DK_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 6 },
  cli: () => ({
    parseCommand: parseDoppelkopfCommand,
    formatResponse: formatDoppelkopfState,
    helpText: DOPPELKOPF_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === DoppelkopfPhase.GAME_END || state.gameEndFlag,
    won:
      (state.phase === DoppelkopfPhase.GAME_END || state.gameEndFlag) &&
      state.winnerIdx === state.players.findIndex((p) => p.isHuman),
  }),
  shell: ({ state }) => ({ isHumanTurn: state.currentPlayerIdx === state.players.findIndex((p) => p.isHuman) }),
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
          value: game.doppelkopfConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetChips',
          label: t('settings.targetChips'),
          value: game.doppelkopfConfig.targetChips,
          options: TARGET_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetChips', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-2 sm:pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) =>
    !view.isMobile && (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ),
  resetTutorialId: 'dk-reset-button',
  actionsTutorialId: 'dk-action-buttons',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};
/** Renders the Doppelkopf game page: 4-player partnership trick-taking with hidden Re/Kontra teams. */
export const DoppelkopfPage = createGamePage(doppelkopfDefinition);
