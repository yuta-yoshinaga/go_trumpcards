import { useEffect } from 'react';
import type { manilleApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useManilleGame } from '../hooks/useManilleGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { ManilleResponse } from '../types/card';
import { ManillePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { suitName } from '../utils/cardUtils';
import { MANILLE_HELP, parseManilleCommand } from '../utils/cli/commands/manilleCommands';
import { formatManilleState } from '../utils/cli/formatters/manilleFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/**
 * Static card-strength legend rows, strongest first, mirroring the Go domain
 * (`manilleStrength`/`manilleCardPoints` in `internal/domain/Manille.go`):
 * 10 (Manille) > A (Manillon) > K > Q > J > 9 > 8 > 7, with points 10=5, A=4,
 * K=3, Q=2, J=1 and 9/8/7 worth zero. `nameKey` labels the two special cards.
 */
const MANILLE_STRENGTH_ROWS: ReadonlyArray<{ face: string; points: number; nameKey?: string }> = [
  { face: '10', points: 5, nameKey: 'strengthLegend.manille' },
  { face: 'A', points: 4, nameKey: 'strengthLegend.manillon' },
  { face: 'K', points: 3 },
  { face: 'Q', points: 2 },
  { face: 'J', points: 1 },
  { face: '9', points: 0 },
  { face: '8', points: 0 },
  { face: '7', points: 0 },
];

/** Manille tutorial step definitions. */
const MANILLE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="manille-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="manille-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="manille-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="manille-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="manille-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MANILLE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ManillePhase.PLAY]: 'play',
  [ManillePhase.TRICK_END]: 'trickEnd',
  [ManillePhase.ROUND_END]: 'roundEnd',
  [ManillePhase.GAME_END]: 'gameEnd',
};

type ManilleGame = ReturnType<typeof useManilleGame>;

function useManilleView({ game }: { game: ManilleGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('manille', MANILLE_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

type ManilleView = ReturnType<typeof useManilleView>;
type ManilleContext = GamePageContext<ManilleResponse, ManilleGame, ManilleView>;

function Board({ state, t, view }: ManilleContext) {
  const { cardWidth, isMobile } = view;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isTrickEnd = state.phase === ManillePhase.TRICK_END;
  const isRoundEnd = state.phase === ManillePhase.ROUND_END;
  const isGameEnd = state.phase === ManillePhase.GAME_END || state.gameEndFlag;
  const humanTeam = humanIdx % 2;
  const trumpSymbol = suitSymbolAt(state.trumpSuit, '');

  // One info-sidebar row per player; same-team members are emphasised. Shared by the
  // mobile (<details>) and desktop layouts so the highlight logic lives in one place.
  const renderPlayerRow = (p: (typeof state.players)[number]) => {
    const isMyTeam = p.id % 2 === humanTeam;
    return (
      <div
        key={p.id}
        data-testid={`manille-player-${p.id}`}
        data-own-team={isMyTeam ? 'true' : undefined}
        className={`text-sm py-0.5 ${
          isMyTeam ? 'font-bold text-ds-text-primary border-l-2 border-ds-accent pl-1' : 'text-ds-text-muted'
        }`}
      >
        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
      </div>
    );
  };

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
        {/* **目標点を確認するのに設定パネルを開き直す必要があった (#4758)。**
            姉妹ゲームの Mariáš は同じ位置に出している。現在のチーム
            スコアと突き合わせられないと、あと何点で終わるか分からない。 */}
        <span>{t('target', { points: state.config.targetPoints })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="manille-trick-display"
          />
          {/* At TrickEnd leadPlayerIdx is the trick winner (they lead next), so show who took it. */}
          {(isTrickEnd || isRoundEnd) &&
            (() => {
              const winnerIdx = state.leadPlayerIdx;
              const isMyTeam = winnerIdx % 2 === humanTeam;
              return (
                <div
                  className={`my-2 p-2 rounded text-center text-sm font-semibold ${
                    isMyTeam ? 'bg-ds-accent/15 text-ds-accent' : 'bg-black/30 text-ds-text-muted'
                  }`}
                  role="status"
                  aria-live="polite"
                  data-testid="manille-trick-winner"
                >
                  {t('trickWinner', {
                    name: playerName(winnerIdx, state.players[winnerIdx]?.isHuman ?? false),
                    team: winnerIdx % 2 === 0 ? t('team.a') : t('team.b'),
                    points: state.trickCardPoints,
                  })}
                </div>
              );
            })()}
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="manille-info">
          {/* Team match scores */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {[0, 1].map((team) => {
              const score = state.teamScores[team] ?? 0;
              const remaining = state.config.targetPoints - score;
              return (
                <div key={team} data-testid={`manille-team-score-${team}`}>
                  {t('teamScore', { team: team === 0 ? t('team.a') : t('team.b'), score })}
                  {remaining > 0 && <span>{t('remainingPoints', { count: remaining })}</span>}
                </div>
              );
            })}
            <div className="mt-1">
              {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
            </div>
          </div>

          {/* Players: cards / tricks. Same-team members (p.id % 2 === humanTeam) are
              emphasised so the 2-vs-2 partnership is visible. Shared by both layouts. */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">{state.players.map(renderPlayerRow)}</div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">{state.players.map(renderPlayerRow)}</div>
          )}

          {/* Card-strength legend (static reference): the Manille (10) tops the
              reversed rank order, so a first-time player can see why 10 beats A. */}
          <details className="mb-2 p-2 rounded bg-black/30" data-testid="manille-strength-legend">
            <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
              {t('strengthLegend.title')}
            </summary>
            <div className="mt-1 text-ds-text-muted text-xs">
              <table className="w-full">
                <thead>
                  <tr>
                    <th scope="col" className="text-left font-normal">
                      {t('strengthLegend.rankCol')}
                    </th>
                    <th scope="col" className="text-left font-normal">
                      {t('strengthLegend.cardCol')}
                    </th>
                    <th scope="col" className="text-right font-normal">
                      {t('strengthLegend.pointCol')}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {MANILLE_STRENGTH_ROWS.map((row, i) => (
                    <tr key={row.face}>
                      <td>{i + 1}</td>
                      <td>{row.nameKey ? `${row.face} (${t(row.nameKey)})` : row.face}</td>
                      <td className="text-right">{row.points}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <div className="mt-1">{t('strengthLegend.note')}</div>
            </div>
          </details>

          {/* **途中経過が追えなかった。**姉妹ゲーム (Sueca / Forty-Fives /
              Twenty-Nine / Klaverjas / Doppelkopf) はどれも進行中の取得点を
              常時出しているのに、マニーユは結果まで累計点しか見えなかった
              (#6442)。ラウンドが終われば下の結果表示が引き継ぐので、
              ここは進行中だけ。 */}
          {!(isRoundEnd || isGameEnd) && (
            <div
              className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              role="status"
              aria-live="polite"
              data-testid="manille-round-points"
            >
              <div className="mb-1 text-ds-text-primary">{t('roundProgress.title')}</div>
              <div>{t('roundResult.teamA', { points: state.roundCardPoints[0] })}</div>
              <div>{t('roundResult.teamB', { points: state.roundCardPoints[1] })}</div>
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

function FooterHand({ game, state, t, view }: ManilleContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const isPlayPhase = state.phase === ManillePhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn;
  const trumpSymbol = suitSymbolAt(state.trumpSuit, '');
  const trumpDesign = suitName(state.trumpSuit);
  const trumpIndices = humanPlayer.cards.reduce<number[]>((indices, card, index) => {
    if (card.design === trumpDesign) indices.push(index);
    return indices;
  }, []);
  const trumpDescription = t('trump', { suit: trumpSymbol });

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="manille"
      trumpIndices={trumpIndices}
      trumpTitle={trumpDescription}
      trumpAccessibleLabel={trumpDescription}
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
    />
  );
}

function FooterStatus({ state, t, frontendHint }: ManilleContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="manille-hint-live" role="status" aria-live="polite">
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

function FooterActions({ game, state, t, busy }: ManilleContext) {
  const isPlayPhase = state.phase === ManillePhase.PLAY;
  const isTrickEnd = state.phase === ManillePhase.TRICK_END;
  const isRoundEnd = state.phase === ManillePhase.ROUND_END;
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

const manilleDefinition: GameDefinition<
  ManilleResponse,
  ManilleGame,
  Parameters<typeof manilleApi.exec>,
  ManilleView
> = {
  key: 'manille',
  useGame: useManilleGame,
  useView: useManilleView,
  tutorial: MANILLE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 8 },
  cli: () => ({
    parseCommand: parseManilleCommand,
    formatResponse: formatManilleState,
    helpText: MANILLE_HELP,
  }),
  phase: ({ state, view }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const humanTeam = humanIdx % 2;
    const isGameEnd = state.phase === ManillePhase.GAME_END || state.gameEndFlag;
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
          value: game.manilleConfig.cpuDifficulty,
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
          value: game.manilleConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  resetTutorialId: 'manille-reset-button',
  actionsTutorialId: 'manille-action-buttons',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Manille game page: a French/Belgian 4-player (2 vs 2) trump trick-taker. */
export const ManillePage = createGamePage(manilleDefinition);
