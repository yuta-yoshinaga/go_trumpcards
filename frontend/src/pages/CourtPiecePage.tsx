import { useEffect } from 'react';
import type { courtPieceApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import {
  CPU_DIFFICULTY_OPTIONS,
  POINT_LIMIT_OPTIONS,
  TRUMP_SUIT_OPTIONS,
  useCourtPieceGame,
} from '../hooks/useCourtPieceGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { CourtPieceResponse } from '../types/card';
import { CourtPiecePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { COURT_PIECE_HELP, parseCourtPieceCommand } from '../utils/cli/commands/courtPieceCommands';
import { formatCourtPieceState } from '../utils/cli/formatters/courtPieceFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tricks a team must take within a 13-trick round to win it (Sar); mirrors CourtPieceTricksToWin in internal/domain/CourtPiece.go. */
const COURT_PIECE_TRICKS_TO_WIN = 7;

/** Shared translation keys for card suit designs used by the trick display. */
const SUIT_I18N_KEYS: Record<string, string> = {
  SPADE: 'common.suit.spade',
  CLOVER: 'common.suit.club',
  HEART: 'common.suit.heart',
  DIAMOND: 'common.suit.diamond',
};

/** Court Piece (Rang) tutorial step definitions. */
const COURT_PIECE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="courtpiece-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="courtpiece-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="courtpiece-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="courtpiece-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const COURT_PIECE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CourtPiecePhase.TRUMP_DECLARATION]: 'trumpDeclaration',
  [CourtPiecePhase.PLAY]: 'play',
  [CourtPiecePhase.TRICK_END]: 'trickEnd',
  [CourtPiecePhase.ROUND_END]: 'roundEnd',
  [CourtPiecePhase.GAME_END]: 'gameEnd',
};

type CourtPieceGame = ReturnType<typeof useCourtPieceGame>;
type CourtPieceView = ReturnType<typeof useCourtPieceView>;
type CourtPieceContext = GamePageContext<CourtPieceResponse, CourtPieceGame, CourtPieceView>;

function useCourtPieceView({ game }: { game: CourtPieceGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('courtpiece', COURT_PIECE_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function phase({ state, view }: CourtPieceContext) {
  const ended = state.phase === CourtPiecePhase.GAME_END || state.gameEndFlag;
  return {
    name: view.phaseNames[state.phase],
    ended,
    won: ended && state.winnerTeam === state.players.findIndex((p) => p.isHuman) % 2,
  };
}

function Board({ state, t, tc, view }: CourtPieceContext) {
  const { cardWidth, isMobile } = view;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const humanTeam = humanIdx % 2;
  const isPlayPhase = state.phase === CourtPiecePhase.PLAY;
  const isTrickEnd = state.phase === CourtPiecePhase.TRICK_END;
  const isRoundEnd = state.phase === CourtPiecePhase.ROUND_END;
  const isGameEnd = state.phase === CourtPiecePhase.GAME_END || state.gameEndFlag;
  const leadSuitKey = state.currentTrick[0] ? SUIT_I18N_KEYS[state.currentTrick[0].card.design] : undefined;
  const teamTricks = (team: number) =>
    state.players.filter((p) => p.team === team).reduce((sum, p) => sum + p.trickCount, 0);
  const renderTeamPlayers = (team: number) =>
    state.players
      .filter((p) => p.team === team)
      .map((p) => (
        <div key={p.id} className="py-0.5 flex items-center gap-2 text-ds-text-muted text-sm">
          <span className={p.id === state.callerIdx ? 'text-ds-warning font-semibold' : ''}>
            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
          </span>
          {p.id === state.callerIdx && (
            <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('callerBadge')}</span>
          )}
        </div>
      ));
  const trumpSymbol = state.trumpSuit === 0 ? t('noTrump') : suitSymbolAt(state.trumpSuit, '');
  return (
    <>
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="courtpiece-info">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
        <span>{t('target', { points: state.config.pointLimit })}</span>
      </div>

      <div className="text-ds-text-muted text-center mb-2 text-sm">
        {state.callerIdx >= 0
          ? t('callerLine', {
              name: playerName(state.callerIdx, state.players[state.callerIdx]?.isHuman ?? false),
              team: state.callerIdx % 2 === 0 ? t('team.a') : t('team.b'),
            })
          : t('callerUndecided')}
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            leadSuit={leadSuitKey ? tc(leadSuitKey) : undefined}
            leadSuitLabel={t('leadSuit')}
            dataTutorial="courtpiece-trick-display"
            // ResolveTrick stores the winner in leadPlayerIdx and keeps it
            // until NextTrick starts the following trick.
            winnerIdx={isTrickEnd ? state.leadPlayerIdx : undefined}
            winnerLabel={t('trickWinnerBadge')}
          />
        </div>

        {/* Right: info sidebar */}
        <div>
          <div data-testid="cp-trick-winner-live" className="sr-only" role="status" aria-live="polite">
            {(isTrickEnd || isRoundEnd || isGameEnd) && state.players[state.leadPlayerIdx]
              ? t('trickWinnerAnnouncement', {
                  name: playerName(state.players[state.leadPlayerIdx].id, state.players[state.leadPlayerIdx].isHuman),
                  team: state.players[state.leadPlayerIdx].team === 0 ? t('team.a') : t('team.b'),
                })
              : ''}
          </div>
          {/* Team match scores */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {([0, 1] as const).map((team) => {
              const isWinner = isGameEnd && state.winnerTeam === team;
              return (
                <div
                  key={team}
                  className={isWinner ? 'text-ds-accent font-semibold' : ''}
                  data-testid={`cp-team-score-${team}`}
                >
                  {t('teamScore', {
                    team: team === 0 ? t('team.a') : t('team.b'),
                    score: state.teamScores[team] ?? 0,
                  })}
                  {isWinner && <span className="ml-2">{t('winner')}</span>}
                </div>
              );
            })}
            <div className="mt-1">
              {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
            </div>
            {/* Live running trick tally toward the 7-trick round target (during play). */}
            {(isPlayPhase || isTrickEnd) && (
              <div
                className="mt-1 pt-1 border-t border-white/10"
                data-testid="cp-live-tricks"
                // トリックが決着するたびに数字が変わるが、読み上げには乗って
                // いなかった。領域はプレイ中ずっと残り、中身だけが差し替わる。
                role="status"
                aria-live="polite"
              >
                <span className="text-ds-text-primary">{t('liveTricks.title')}: </span>
                {([0, 1] as const).map((team, i) => {
                  const tricks = teamTricks(team);
                  const reached = tricks >= COURT_PIECE_TRICKS_TO_WIN;
                  return (
                    <span key={team}>
                      {i > 0 && ' · '}
                      <span
                        className={reached ? 'text-ds-accent font-semibold' : ''}
                        data-testid={`cp-live-tricks-team-${team}`}
                      >
                        {t('liveTricks.team', {
                          team: team === 0 ? t('team.a') : t('team.b'),
                          tricks,
                          target: COURT_PIECE_TRICKS_TO_WIN,
                        })}
                        {/* 到達の強調は色と太さだけで、支援技術には届かない。
                      同じことを語でも言う。 */}
                        {reached && <span className="sr-only"> {t('liveTricks.reached')}</span>}
                      </span>
                    </span>
                  );
                })}
              </div>
            )}
          </div>

          {/* Players grouped by team, with the caller badge */}
          {([0, 1] as const).map((team) => (
            <div key={team} className="mb-2 p-2 rounded bg-black/30">
              <div className="text-ds-text-primary text-xs mb-1">{team === 0 ? t('team.a') : t('team.b')}</div>
              {isMobile ? (
                <details>
                  <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                  <div className="mt-1">{renderTeamPlayers(team)}</div>
                </details>
              ) : (
                renderTeamPlayers(team)
              )}
            </div>
          ))}

          {/* Round result: tricks per team */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.teamA', { tricks: teamTricks(0) })}</div>
              <div>{t('roundResult.teamB', { tricks: teamTricks(1) })}</div>
              {state.lastRoundCourt && <div className="mt-1 text-ds-warning">{t('roundResult.court')}</div>}
              <div className="mt-2 pt-1 border-t border-white/10" data-testid="cp-score-breakdown">
                {([0, 1] as const).map((team) => {
                  const breakdown = state.scoreBreakdown[team];
                  const formatPoints = (value: number | undefined) =>
                    value == null || value === 0 ? '0' : value < 0 ? `−${Math.abs(value)}` : `+${value}`;
                  return (
                    <div key={team}>
                      <span>{team === 0 ? t('team.a') : t('team.b')}: </span>
                      <span>
                        {t('roundResult.sar')}: {formatPoints(breakdown?.sar)}
                      </span>
                      {' · '}
                      <span>
                        {t('roundResult.courtBonus')}: {formatPoints(breakdown?.courtBonus)}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function AboveContent({ t }: CourtPieceContext) {
  return <p className="px-4 pt-1 text-center text-xs text-ds-text-muted">{t('settings.nextGameNotice')}</p>;
}

function FooterHand({ game, state, view }: CourtPieceContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;
  // The web contract carries no explicit turn flags, so derive them from the
  // current seat: it is the human's turn whenever currentPlayerIdx is the human.
  const canPlay =
    state.phase === CourtPiecePhase.PLAY && state.currentPlayerIdx === state.players.findIndex((p) => p.isHuman);
  // 合法手はサーバーが判定済みの state.playableIndices をそのまま使う。
  // 以前はここで validatePlay を TypeScript に写して再計算しており、ルールを
  // 変えたら片方だけ直してずれる状態だった。additive hint である点は従来どおりで、
  // 非合法手もクリックでき、拒否の最終判断はバックエンドが持つ。
  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="courtpiece"
      legalIndices={canPlay ? state.playableIndices : undefined}
    />
  );
}

/* ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
   領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
   られないことがある (#5955)。 */
function FooterStatus({ frontendHint, state, t }: CourtPieceContext) {
  return (
    <>
      <div data-testid="courtpiece-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndex != null && ` ([${state.hint.cardIndex}])`}
            {state.hint.trumpSuit != null && ` (${suitSymbolAt(state.hint.trumpSuit, '')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t }: CourtPieceContext) {
  const isTrumpPhase = state.phase === CourtPiecePhase.TRUMP_DECLARATION;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTrumpTurn = isTrumpPhase && state.callerIdx === humanIdx;
  const canPlay = state.phase === CourtPiecePhase.PLAY && state.currentPlayerIdx === humanIdx;
  return (
    <>
      {isTrumpPhase && isHumanTrumpTurn && (
        <>
          <span className="text-xs font-bold text-ds-warning self-center mr-1" data-testid="cp-trump-status">
            {t('trumpUndeclared')}
          </span>
          <span className="text-xs text-ds-text-muted self-center mr-1">{t('trumpPrompt')}</span>
          {TRUMP_SUIT_OPTIONS.map((s) => (
            <button
              key={s.value}
              type="button"
              className="px-3 py-2 rounded-lg bg-ds-info text-ds-text-on-accent text-sm disabled:opacity-40"
              onClick={() => game.handleDeclareTrump(s.value)}
              disabled={game.loading}
              aria-label={t('declareTrumpAria', { suit: t(s.key) })}
              data-testid={`trump-${s.value}`}
            >
              {t(s.key)}
            </button>
          ))}
        </>
      )}
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={game.loading || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {state.phase === CourtPiecePhase.TRICK_END && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={game.loading}>
          {t('nextTrick')}
        </button>
      )}
      {state.phase === CourtPiecePhase.ROUND_END && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={game.loading}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const def: GameDefinition<CourtPieceResponse, CourtPieceGame, Parameters<typeof courtPieceApi.exec>, CourtPieceView> = {
  key: 'courtpiece',
  useGame: useCourtPieceGame,
  useView: useCourtPieceView,
  tutorial: COURT_PIECE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 13 },
  cli: () => ({
    parseCommand: parseCourtPieceCommand,
    formatResponse: formatCourtPieceState,
    helpText: COURT_PIECE_HELP,
  }),
  phase,
  shell: ({ state }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const humanTurn =
      (state.phase === CourtPiecePhase.PLAY && state.currentPlayerIdx === humanIdx) ||
      (state.phase === CourtPiecePhase.TRUMP_DECLARATION && state.callerIdx === humanIdx);
    return { isHumanTurn: humanTurn && !(state.phase === CourtPiecePhase.GAME_END || state.gameEndFlag) };
  },
  actionLogEndPhase: ({ state }) => state.phase === CourtPiecePhase.GAME_END || state.gameEndFlag,
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.courtPieceConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'pointLimit',
          label: t('settings.pointLimit'),
          value: game.courtPieceConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('pointLimit', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  AboveContent,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'courtpiece-action-buttons',
  resetTutorialId: 'courtpiece-reset-button',
  reset: ({ game, hideActionLog }) => {
    hideActionLog();
    game.reset();
  },
};

/** Renders the Court Piece (Rang) game page: a South-Asian 4-player (2 vs 2) called-trump trick-taker. */
export const CourtPiecePage = createGamePage(def);
