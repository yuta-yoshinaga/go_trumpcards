import type { TFunction } from 'i18next';
import { useEffect, useRef, useState } from 'react';
import type { tressetteApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useTressetteGame } from '../hooks/useTressetteGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { TressetteResponse } from '../types/card';
import { TressettePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTressetteCommand, TRESSETTE_HELP } from '../utils/cli/commands/tressetteCommands';
import { formatTressetteState } from '../utils/cli/formatters/tressetteFormatter';
import { formatThirdPoints } from '../utils/formatThirdPoints';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tressette tutorial step definitions. */
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

const TRESSETTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TressettePhase.PLAY]: 'play',
  [TressettePhase.TRICK_END]: 'trickEnd',
  [TressettePhase.ROUND_END]: 'roundEnd',
  [TressettePhase.GAME_END]: 'gameEnd',
};

/**
 * Clamps a team's round-thirds count to the [0, 3] range used by the 3-dot
 * indicator. Thirds reset to 0 once they reach 3 (converted to a point), but
 * this guards defensively against any out-of-range value.
 */
function thirdsFilled(thirds: number): number {
  return Math.min(Math.max(thirds, 0), 3);
}

const teamLabels = ['A', 'B'];

type TressetteGame = ReturnType<typeof useTressetteGame>;
type TressetteView = ReturnType<typeof useTressetteView>;
type TressetteContext = GamePageContext<TressetteResponse, TressetteGame, TressetteView>;

function useTressetteView({ game, t }: { game: TressetteGame; t: TFunction }) {
  const { state } = game;
  const previousScoreRef = useRef<{ round: number; thirds: [number, number] } | null>(null);
  const [scoreAnnouncement, setScoreAnnouncement] = useState('');
  useEffect(() => {
    if (!state) return;
    const thirds: [number, number] = [state.teamRoundThirds[0] ?? 0, state.teamRoundThirds[1] ?? 0];
    const previous = previousScoreRef.current;
    previousScoreRef.current = { round: state.roundNumber, thirds };
    if (!previous) return;
    if (state.roundNumber !== previous.round) return;

    const changes = [0, 1]
      .map((team) => ({
        team,
        gained: Math.max(0, thirds[team] - previous.thirds[team]),
      }))
      .filter(({ gained }) => gained > 0);
    if (changes.length === 0) return;
    setScoreAnnouncement(
      changes
        .map(({ team, gained }) =>
          t('scoreAnnouncement', {
            team: team === 0 ? 'A' : 'B',
            gained: formatThirdPoints(gained),
            points: state.teamScores[team] ?? 0,
            thirds: (state.teamRoundThirds[team] ?? 0) % 3,
          }),
        )
        .join(t('listSeparator')),
    );
  }, [state, t]);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('tressette', TRESSETTE_PHASE_KEYS);

  const isTrickEnd = state?.phase === TressettePhase.TRICK_END;
  const trickWinner =
    isTrickEnd && (state?.lastTrick.length ?? 0) > 0 ? state?.players[state.lastTrickWinner] : undefined;
  const hasTrickResult = trickWinner !== undefined;
  const trickResult =
    trickWinner && state
      ? t('trickResult', {
          name: playerName(state.lastTrickWinner, trickWinner.isHuman),
          team: t('teamLabel', { team: teamLabels[trickWinner.teamId] }),
          points: formatThirdPoints(state.currentTrickThirds),
        })
      : '';

  return { cardWidth, isMobile, phaseNames, scoreAnnouncement, hasTrickResult, trickResult };
}

function LiveRegion({ view }: TressetteContext) {
  return (
    <>
      <span className="sr-only" data-testid="tr-score-live" role="status" aria-live="polite" aria-atomic="true">
        {view.scoreAnnouncement}
      </span>
      <span className="sr-only" data-testid="tr-trick-result-live" role="status" aria-live="polite" aria-atomic="true">
        {view.trickResult}
      </span>
    </>
  );
}

function Board({ state, t, tc, view }: TressetteContext) {
  const { cardWidth, isMobile, hasTrickResult, trickResult } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        {/* **目標点を確認するのに設定パネルを開き直す必要があった (#6429)。**
            現在のチーム得点と突き合わせられないと、あと何点で終わるのか
            分からない。姉妹ゲームの Manille / Mariáš は同じ位置に出している。
            読むのは `state.config` ── 設定パネルの値はリセットするまで
            卓に効かないので、進行中の局とずれる。 */}
        <span data-testid="tr-target">{t('target', { points: state.config.targetPoints })}</span>
      </div>

      <output className="text-ds-text-primary text-center text-sm mb-2" data-testid="tr-current-trick-points">
        {state.currentTrick.length === 0
          ? t('currentTrickPointsEmpty')
          : t('currentTrickPoints', { thirds: state.currentTrickThirds })}
      </output>

      {hasTrickResult && (
        <div className="mb-2 rounded bg-ds-surface p-2 text-center text-ds-text-primary" data-testid="tr-trick-result">
          {trickResult}
        </div>
      )}

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
                      {playerName(p.id, p.isHuman)} [{t('teamLabel', { team: teamLabels[p.teamId] })}]
                      {humanPlayer && (
                        <span data-testid={`tr-player-role-${p.id}`} className="ml-1 text-ds-accent">
                          {t(p.teamId === humanPlayer.teamId ? 'partner' : 'opponent')}
                        </span>
                      )}
                      : {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
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
                    {playerName(p.id, p.isHuman)} [{t('teamLabel', { team: teamLabels[p.teamId] })}]
                    {humanPlayer && (
                      <span data-testid={`tr-player-role-${p.id}`} className="ml-1 text-ds-accent">
                        {t(p.teamId === humanPlayer.teamId ? 'partner' : 'opponent')}
                      </span>
                    )}
                    : {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
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
                  <th scope="col">{t('scoresPoints')}</th>
                  <th scope="col">{t('scoresThirds')}</th>
                </tr>
              </thead>
              <tbody>
                {teamLabels.map((label, idx) => {
                  const score = state.teamScores[idx] as number;
                  const pointsRemaining = state.config.targetPoints - score;

                  return (
                    <tr key={label} className={humanPlayer && humanPlayer.teamId === idx ? 'text-ds-accent' : ''}>
                      <td>{t('teamLabel', { team: label })}</td>
                      <td className="text-center">
                        {score}
                        {pointsRemaining > 0 && (
                          <span className="ml-1">
                            {t('pointsRemaining', {
                              count: pointsRemaining,
                            })}
                          </span>
                        )}
                      </td>
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
                  );
                })}
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

function FooterHand({ game, state, view, t }: TressetteContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.phase === TressettePhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  return humanPlayer ? (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="tr"
      validIndices={isHumanTurn ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
      cardStatusFor={(idx) => {
        const thirds = state.handCardThirds[idx];
        const pointKey = thirds === 3 ? 'aceValue' : thirds === 1 ? 'figuresValue' : 'othersValue';
        return t('cardAccessibleStatus', { points: t(`pointLegend.${pointKey}`) });
      }}
    />
  ) : null;
}

function FooterStatus({ game, t, frontendHint }: TressetteContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="tressette-hint-live" role="status" aria-live="polite">
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

function FooterActions({ game, state, t, tc, busy }: TressetteContext) {
  const isPlayPhase = state.phase === TressettePhase.PLAY;
  const isTrickEnd = state.phase === TressettePhase.TRICK_END;
  const isRoundEnd = state.phase === TressettePhase.ROUND_END;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
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

const tressetteDefinition: GameDefinition<
  TressetteResponse,
  TressetteGame,
  Parameters<typeof tressetteApi.exec>,
  TressetteView
> = {
  key: 'tressette',
  useGame: useTressetteGame,
  useView: useTressetteView,
  tutorial: TR_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseTressetteCommand,
    formatResponse: formatTressetteState,
    helpText: TRESSETTE_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === TressettePhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ state }) => ({
    isHumanTurn: state.phase === TressettePhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true,
  }),
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.tressetteConfig.cpuDifficulty,
      targetPoints: game.tressetteConfig.targetPoints,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.tressetteConfig.cpuDifficulty,
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
          value: game.tressetteConfig.targetPoints,
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
    const isHumanTurn = state.phase === TressettePhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: game.handlePlay,
      onClear: game.clearSelection,
      enabled: !!isHumanTurn,
    };
  },
  shortcutsTestId: 'tressette-kbd-shortcuts',
  resetTutorialId: 'tr-reset-button',
  actionsTutorialId: 'tr-play-button',
  actionsClassName: 'flex gap-2 items-center',
  LiveRegion,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Tressette game page: no-trump must-follow trick play with team scoring. */
export const TressettePage = createGamePage(tressetteDefinition);
