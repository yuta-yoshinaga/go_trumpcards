import type { twoTenJackApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { TrickDisplay } from '../components/TrickDisplay';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useTwoTenJackGame } from '../hooks/useTwoTenJackGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { scoreTableClass } from '../styles/scoreTableStyles';
import type { TwoTenJackResponse } from '../types/card';
import { TwoTenJackPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTwoTenJackCommand, TWOTENJACK_HELP } from '../utils/cli/commands/twoTenJackCommands';
import { formatTwoTenJackState } from '../utils/cli/formatters/twoTenJackFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tutorial steps for Two Ten Jack. Walks the player through declare, play, and scoring elements. */
const TTJ_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tt-declare-controls"]',
    messageKey: 'tutorial.declareControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TWOTENJACK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TwoTenJackPhase.DECLARE]: 'declare',
  [TwoTenJackPhase.PLAY]: 'play',
  [TwoTenJackPhase.TRICK_END]: 'trickEnd',
  [TwoTenJackPhase.ROUND_END]: 'roundEnd',
  [TwoTenJackPhase.GAME_END]: 'gameEnd',
};

/** Suit options shown in the declare phase picker. Values match the domain CardDesign constants. */
const SUIT_OPTIONS: Readonly<{ value: number; symbol: string; key: string }[]> = [
  { value: 1, symbol: '\u2660', key: 'spade' },
  { value: 2, symbol: '\u2663', key: 'club' },
  { value: 3, symbol: '\u2665', key: 'heart' },
  { value: 4, symbol: '\u2666', key: 'diamond' },
];

/** Returns the human-readable suit symbol for a declared trump suit value. */
function trumpSymbol(trumpSuit: number): string {
  return SUIT_OPTIONS.find((s) => s.value === trumpSuit)?.symbol ?? '-';
}

type TwoTenJackGame = ReturnType<typeof useTwoTenJackGame>;
type TwoTenJackView = ReturnType<typeof useTwoTenJackView>;
type TwoTenJackContext = GamePageContext<TwoTenJackResponse, TwoTenJackGame, TwoTenJackView>;

function useTwoTenJackView() {
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('twotenjack', TWOTENJACK_PHASE_KEYS);
  return { isLargeDesktop, cardWidth, isMobile, phaseNames };
}

function Board({ state, t, tc, view, frontendHint }: TwoTenJackContext) {
  const { isLargeDesktop, cardWidth, isMobile } = view;
  const isDeclarePhase = state.phase === TwoTenJackPhase.DECLARE;
  const isTrickEnd = state.phase === TwoTenJackPhase.TRICK_END;
  const isRoundEnd = state.phase === TwoTenJackPhase.ROUND_END;
  const trickWinner =
    (isTrickEnd || isRoundEnd) && state.leadPlayerIdx >= 0 ? state.players[state.leadPlayerIdx] : undefined;
  const declarer = state.players[state.declarerIdx];
  const isGameEnd = state.phase === TwoTenJackPhase.GAME_END || state.gameEndFlag;
  const isHumanDeclarer = isDeclarePhase && state.players[state.declarerIdx]?.isHuman === true;

  // Team totals (team 0 = seats 0,2 ; team 1 = seats 1,3).
  const team0Total = (state.players[0]?.cumulativeScore ?? 0) + (state.players[2]?.cumulativeScore ?? 0);
  const team1Total = (state.players[1]?.cumulativeScore ?? 0) + (state.players[3]?.cumulativeScore ?? 0);
  const team0Captured = (state.players[0]?.capturedPoints ?? 0) + (state.players[2]?.capturedPoints ?? 0);
  const team1Captured = (state.players[1]?.capturedPoints ?? 0) + (state.players[3]?.capturedPoints ?? 0);
  // **読み上げる「+N」は加点であって、取った点札ではない (#5527)。**契約を
  // 落としたラウンドは点札を取っていても加点 0 になるので、capturedPoints を
  // 読み上げると累計の増分と食い違う。cumulativeScore が両席の合計である以上
  // (domain の checkGameEnd も同じ数え方)、増分も両席の roundScore の合計。
  const team0RoundScore = (state.players[0]?.roundScore ?? 0) + (state.players[2]?.roundScore ?? 0);
  const team1RoundScore = (state.players[1]?.roundScore ?? 0) + (state.players[3]?.roundScore ?? 0);
  const playersWithTeams = state.players.map((player, seatIdx) => ({ ...player, team: seatIdx % 2 }));

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>{state.trumpSuit > 0 ? `${t('trump')}: ${trumpSymbol(state.trumpSuit)}` : t('trumpUndeclared')}</span>
      </div>
      {declarer && (
        <div className="text-ds-text-muted text-center mb-2">
          {t('declarer', {
            name: playerName(declarer.id, declarer.isHuman),
            team: t(state.declarerIdx % 2 === 0 ? 'team0' : 'team1'),
          })}
        </div>
      )}

      <div className={lgTwoColGrid}>
        <div>
          {isHumanDeclarer && (
            <div className="text-ds-warning text-center mb-2" data-tutorial="tt-declare-controls">
              {t('declarePhase')}
            </div>
          )}

          <TrickDisplay
            currentTrick={state.currentTrick}
            players={playersWithTeams}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="tt-trick-display"
          />
          <div role="status" aria-live="polite" className="sr-only" data-testid="twotenjack-trick-winner-live">
            {trickWinner ? t('trickWinner', { name: playerName(trickWinner.id, trickWinner.isHuman) }) : ''}
          </div>
        </div>

        <div>
          {/* CPU players */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
              </summary>
              <div className="mt-1">
                {playersWithTeams
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} data-cpu-team={p.team} className="text-ds-text-muted text-sm py-0.5">
                      {playerName(p.id, p.isHuman)}: {t('playerTeam', { team: t(p.team === 0 ? 'team0' : 'team1') })} |{' '}
                      {t('cards', { count: p.cardCount })} | {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                      {t('tricks', { count: p.trickCount })} | {t('capturedPoints', { count: p.capturedPoints })}
                    </div>
                  ))}
              </div>
            </details>
          ) : (
            playersWithTeams
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} data-cpu-team={p.team} className="mb-2 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm">
                    {playerName(p.id, p.isHuman)}: {t('playerTeam', { team: t(p.team === 0 ? 'team0' : 'team1') })} |{' '}
                    {t('cards', { count: p.cardCount })} | {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                    {t('tricks', { count: p.trickCount })} | {t('capturedPoints', { count: p.capturedPoints })}
                  </div>
                </div>
              ))
          )}

          {/* Score table */}
          {isMobile ? (
            <details
              className="my-3 p-2 rounded bg-black/30 relative"
              data-tutorial="tt-score-table"
              open={isRoundEnd || isGameEnd || undefined}
            >
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                <span>{t('scores')}</span>
                <span className="ml-2 inline-flex flex-wrap gap-x-3 tabular-nums text-ds-text-primary">
                  <span>{t('cumulativeTeamScore', { team: t('team0'), score: team0Total })}</span>
                  <span>{t('cumulativeTeamScore', { team: t('team1'), score: team1Total })}</span>
                </span>
              </summary>
              <div className="overflow-x-auto -mx-2 px-2">
                <table
                  className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[320px]', 'mt-1')}
                >
                  <caption className="sr-only">{t('scoresCaption')}</caption>
                  <thead>
                    <tr>
                      <th scope="col" className="text-left">
                        {t('scoresTeam')}
                      </th>
                      <th scope="col">{t('scoresCaptured')}</th>
                      <th scope="col">{t('scoresTotal')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="text-ds-accent">
                      <td title={t('team0')}>{t('team0')}</td>
                      <td className="text-center">{team0Captured}</td>
                      <td className="text-center">{team0Total}</td>
                    </tr>
                    <tr>
                      <td title={t('team1')}>{t('team1')}</td>
                      <td className="text-center">{team1Captured}</td>
                      <td className="text-center">{team1Total}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <ScrollFadeHint />
            </details>
          ) : (
            <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="tt-score-table">
              <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
              <div className="overflow-x-auto -mx-2 px-2">
                <table className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[320px]')}>
                  <caption className="sr-only">{t('scoresCaption')}</caption>
                  <thead>
                    <tr>
                      <th scope="col" className="text-left">
                        {t('scoresTeam')}
                      </th>
                      <th scope="col">{t('scoresCaptured')}</th>
                      <th scope="col">{t('scoresTotal')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="text-ds-accent">
                      <td title={t('team0')}>{t('team0')}</td>
                      <td className="text-center">{team0Captured}</td>
                      <td className="text-center">{team0Total}</td>
                    </tr>
                    <tr>
                      <td title={t('team1')}>{t('team1')}</td>
                      <td className="text-center">{team1Captured}</td>
                      <td className="text-center">{team1Total}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          )}
          <RoundScoreAnnouncement
            active={isRoundEnd || isGameEnd}
            entries={[
              { name: t('team0'), roundScore: team0RoundScore, cumulativeScore: team0Total },
              { name: t('team1'), roundScore: team1RoundScore, cumulativeScore: team1Total },
            ]}
          />
        </div>
      </div>

      {isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}

      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand({ game, state, view }: TwoTenJackContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.phase === TwoTenJackPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  return humanPlayer ? (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="tt"
      validIndices={isHumanTurn ? state.validPlayIndices : undefined}
      highlightIndices={isHumanTurn && game.hint?.cardIndex !== undefined ? [game.hint.cardIndex] : undefined}
    />
  ) : null;
}

function FooterStatus({ game, state, t }: TwoTenJackContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  return (
    /* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
        付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
        読み上げられないことがある (#5955, #6663)。 */
    <div data-testid="twotenjack-hint-live" role="status" aria-live="polite">
      {game.hint && (
        <div className="text-ds-warning text-sm mb-2" data-testid="tt-hint">
          {game.hint.trumpSuit !== undefined
            ? `${t('hintDeclare')}: ${trumpSymbol(game.hint.trumpSuit)} (${t(`hint.${game.hint.reason}`)})`
            : (() => {
                const card = game.hint.cardIndex !== undefined ? humanPlayer?.cards[game.hint.cardIndex] : undefined;
                const name = card ? cardAlt(card) : '-';
                return `${t('hintPlay')}: ${name} [${game.hint.cardIndex ?? '-'}] (${t(`hint.${game.hint.reason}`)})`;
              })()}
        </div>
      )}
    </div>
  );
}

function FooterActions({ game, state, t, tc, busy }: TwoTenJackContext) {
  const isDeclarePhase = state.phase === TwoTenJackPhase.DECLARE;
  const isHumanTurn = state.phase === TwoTenJackPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanDeclarer = isDeclarePhase && state.players[state.declarerIdx]?.isHuman === true;
  const isTrickEnd = state.phase === TwoTenJackPhase.TRICK_END;
  const isRoundEnd = state.phase === TwoTenJackPhase.ROUND_END;

  return (
    <>
      {(isHumanDeclarer || isHumanTurn) && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleHint}
          disabled={busy || game.hintLoading}
          data-testid="tt-hint-button"
        >
          {tc('button.hint')}
        </button>
      )}
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <span data-testid="tt-prompt-live" role="status" aria-live="polite">
        {isHumanDeclarer && (
          <span data-testid="tt-declare-prompt" className="text-ds-warning text-sm font-medium w-full sm:w-auto">
            {t('declarePrompt')}
          </span>
        )}
      </span>
      {isDeclarePhase && !isHumanDeclarer && (
        <span
          data-testid="tt-cpu-declaring"
          role="status"
          className="flex items-center gap-2 text-ds-info text-sm font-medium w-full sm:w-auto"
        >
          <span
            aria-hidden="true"
            className="inline-block w-3 h-3 rounded-full border-2 border-ds-info border-t-transparent motion-safe:animate-spin"
          />
          {t('cpuDeclaring')}
        </span>
      )}
      {isHumanDeclarer &&
        SUIT_OPTIONS.map((suit) => (
          <button
            key={suit.key}
            type="button"
            // **プレイフェーズは既にヒント札を光らせているのに、宣言
            // フェーズはテキストで「♠」と言うだけだった。**同じページ内で
            // 非対称なので、推奨スートのボタンにも同じ強調を付ける。
            className={game.hint?.trumpSuit === suit.value ? `${btnPrimary} ring-2 ring-ds-warning` : btnPrimary}
            aria-label={t(`suit.${suit.key}`)}
            data-hint-suggested={game.hint?.trumpSuit === suit.value ? 'true' : undefined}
            onClick={() => game.handleDeclare(suit.value)}
            disabled={busy}
          >
            {suit.symbol}
          </button>
        ))}
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

const twoTenJackDefinition: GameDefinition<
  TwoTenJackResponse,
  TwoTenJackGame,
  Parameters<typeof twoTenJackApi.exec>,
  TwoTenJackView
> = {
  key: 'twotenjack',
  useGame: useTwoTenJackGame,
  useView: useTwoTenJackView,
  tutorial: TTJ_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseTwoTenJackCommand,
    formatResponse: formatTwoTenJackState,
    helpText: TWOTENJACK_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === TwoTenJackPhase.GAME_END || state.gameEndFlag,
    // The human sits at seat 0, so team 0 is the human team.
    won: (state.phase === TwoTenJackPhase.GAME_END || state.gameEndFlag) && state.winnerTeam === 0,
  }),
  shell: ({ state }) => {
    const isHumanDeclarer =
      state.phase === TwoTenJackPhase.DECLARE && state.players[state.declarerIdx]?.isHuman === true;
    const isHumanTurn = state.phase === TwoTenJackPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      isHumanTurn: isHumanDeclarer || isHumanTurn,
    };
  },
  message: ({ state, view }) =>
    view.isMobile ? undefined : (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ),
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.twoTenJackConfig.cpuDifficulty,
      pointLimit: game.twoTenJackConfig.pointLimit,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.twoTenJackConfig.cpuDifficulty,
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
          value: game.twoTenJackConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('pointLimit', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  cardNav: ({ game, state }) => {
    const isHumanTurn = state.phase === TwoTenJackPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: game.handlePlay,
      onClear: game.clearSelection,
      enabled: isHumanTurn,
    };
  },
  shortcutsTestId: 'two-ten-jack-kbd-shortcuts',
  resetTutorialId: 'tt-reset-button',
  actionsTutorialId: 'tt-play-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Two Ten Jack game page: declare trump, trick play, and team scoring. */
export const TwoTenJackPage = createGamePage(twoTenJackDefinition);
