import { useEffect } from 'react';
import type { suecaApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_GAME_POINTS_OPTIONS, useSuecaGame } from '../hooks/useSuecaGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { Card, SuecaResponse } from '../types/card';
import type { SuecaPlayer } from '../types/games/sueca';
import { SuecaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt, suitSymbolAt } from '../utils/cardAlt';
import { suitName } from '../utils/cardUtils';
import { parseSuecaCommand, SUECA_HELP } from '../utils/cli/commands/suecaCommands';
import { formatSuecaState } from '../utils/cli/formatters/suecaFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Card points needed to win a Sueca round (majority of the 120 total). */
const SUECA_WIN_POINTS = 61;

/** Suit id → `suitName.*` i18n key (1=♠ .. 4=♦). */
const SUIT_KEYS = ['', 'spade', 'club', 'heart', 'diamond'] as const;

/** Sueca tutorial step definitions. */
const SUECA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sueca-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sueca-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sueca-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sueca-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sueca-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SUECA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SuecaPhase.PLAY]: 'play',
  [SuecaPhase.TRICK_END]: 'trickEnd',
  [SuecaPhase.ROUND_END]: 'roundEnd',
  [SuecaPhase.GAME_END]: 'gameEnd',
};

/**
 * Game points won this round → the i18n key naming that tier.
 *
 * 段を決めるのはサーバ (`suecaGamePoints`: 61-90=1 / 91-119=2 / 120=4)。ここは
 * その結果に名前を付けるだけで、カード点をしきい値にかけ直さない。
 */
const SUECA_AWARD_KIND: Record<number, string> = {
  1: 'gamePointsNormal',
  2: 'gamePointsDouble',
  4: 'gamePointsShutout',
};

type SuecaGame = ReturnType<typeof useSuecaGame>;
type SuecaView = ReturnType<typeof useSuecaView>;
type SuecaContext = GamePageContext<SuecaResponse, SuecaGame, SuecaView>;

function useSuecaView({ game }: { game: SuecaGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('sueca', SUECA_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function SuecaMessageBox({ state }: Pick<SuecaContext, 'state'>) {
  return <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />;
}

function Board({ state, t, view }: SuecaContext) {
  const { cardWidth, isMobile } = view;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const humanTeam = humanIdx % 2;
  const trumpSymbol = suitSymbolAt(state.trumpSuit, '');
  const playerTeamLabel = (p: SuecaPlayer) =>
    t('playerTeam', {
      player: playerName(p.id, p.isHuman),
      team: p.team === 0 ? t('team.a') : t('team.b'),
    });
  // Spoken trump: the suit name (not the ♠♣♥♦ glyph, which SRs read poorly).
  const trumpSuitName = SUIT_KEYS[state.trumpSuit] ? t(`suitName.${SUIT_KEYS[state.trumpSuit]}`) : trumpSymbol;
  const trumpAriaLabel = t('trump', { suit: trumpSuitName });

  const isTrickEnd = state.phase === SuecaPhase.TRICK_END;
  const isRoundEnd = state.phase === SuecaPhase.ROUND_END;
  const isGameEnd = state.phase === SuecaPhase.GAME_END || state.gameEndFlag;

  const messageBox = <SuecaMessageBox state={state} />;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span role="img" aria-label={trumpAriaLabel}>
          <span aria-hidden="true">{t('trump', { suit: trumpSymbol })}</span>
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
            dataTutorial="sueca-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="sueca-info">
          {/* Team game points */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {/* Trump kept in the always-visible sidebar so it isn't lost off-screen on mobile. */}
            <div
              className="mb-1 text-ds-text-primary font-semibold"
              data-testid="sueca-sidebar-trump"
              role="img"
              aria-label={trumpAriaLabel}
            >
              <span aria-hidden="true">{t('trump', { suit: trumpSymbol })}</span>
            </div>
            <div data-testid="sueca-team-game-points">
              <div>
                {t('teamScore', {
                  team: t('team.a'),
                  score: state.teamGamePoints[0],
                  target: state.config.targetGamePoints,
                })}
              </div>
              <div>
                {t('teamScore', {
                  team: t('team.b'),
                  score: state.teamGamePoints[1],
                  target: state.config.targetGamePoints,
                })}
              </div>
            </div>
            <div className="mt-1">
              {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
            </div>
          </div>

          {/* Players: cards / tricks */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5" data-testid={`sueca-player-${p.id}`}>
                    {playerTeamLabel(p)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5" data-testid={`sueca-player-${p.id}`}>
                  {playerTeamLabel(p)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                </div>
              ))}
            </div>
          )}

          {/* Live captured card points during play; the round-result block
              below takes over once the round ends. Progress is shown against
              the 61-point round-win line. */}
          {!(isRoundEnd || isGameEnd) && (
            <div
              className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              role="status"
              aria-live="polite"
              data-testid="sueca-round-points"
            >
              <div className="mb-1 text-ds-text-primary">{t('livePoints.title')}</div>
              <div>
                {t('livePoints.team', {
                  team: t('team.a'),
                  points: state.roundCardPoints[0] ?? 0,
                  target: SUECA_WIN_POINTS,
                })}
              </div>
              <div>
                {t('livePoints.team', {
                  team: t('team.b'),
                  points: state.roundCardPoints[1] ?? 0,
                  target: SUECA_WIN_POINTS,
                })}
              </div>
            </div>
          )}

          {/* 配点そのものが画面のどこにも無かった (#5642)。A=11 / 7=10 は
              直感に反するのに、出ているのは 61 点という目標だけだった。
              姉妹ゲームの Tressette と同じ折りたたみ式にする。 */}
          <details className="my-3 p-2 rounded bg-black/30" data-testid="sueca-point-legend">
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
                    <td>{t('pointLegend.seven')}</td>
                    <td className="text-right">{t('pointLegend.sevenValue')}</td>
                  </tr>
                  <tr>
                    <td>{t('pointLegend.king')}</td>
                    <td className="text-right">{t('pointLegend.kingValue')}</td>
                  </tr>
                  <tr>
                    <td>{t('pointLegend.jack')}</td>
                    <td className="text-right">{t('pointLegend.jackValue')}</td>
                  </tr>
                  <tr>
                    <td>{t('pointLegend.queen')}</td>
                    <td className="text-right">{t('pointLegend.queenValue')}</td>
                  </tr>
                  <tr>
                    <td>{t('pointLegend.others')}</td>
                    <td className="text-right">{t('pointLegend.othersValue')}</td>
                  </tr>
                </tbody>
              </table>
              <div className="mt-1">{t('pointLegend.note')}</div>
            </div>
          </details>

          {/* Round result */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.teamA', { points: state.roundCardPoints[0] ?? 0 })}</div>
              <div>{t('roundResult.teamB', { points: state.roundCardPoints[1] ?? 0 })}</div>
              {/* **通算が 1・2・4 のどれ増えたのか理由が出ていなかった。**カード点だけ
                  では 61-90 / 91-119 / 120 のどの段だったのか読めない (#6438)。
                  段はサーバが決めた `roundGamePoints` から引く ── ここでカード点を
                  もう一度しきい値にかけると、ドメインと 2 箇所で同じ規則を持つ。 */}
              <div data-testid="sueca-round-award">
                {state.roundWinnerTeam < 0
                  ? t('roundResult.gamePointsDraw')
                  : t('roundResult.gamePoints', {
                      team: state.roundWinnerTeam === 0 ? t('team.a') : t('team.b'),
                      points: state.roundGamePoints,
                      kind: t(`roundResult.${SUECA_AWARD_KIND[state.roundGamePoints] ?? 'gamePointsNormal'}`),
                    })}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* At TrickEnd leadPlayerIdx is the trick winner, so their team shows before Next Trick. */}
      {isTrickEnd && (
        <div
          className="my-2 p-2 rounded bg-ds-accent/15 text-center text-sm font-semibold text-ds-accent"
          role="status"
          aria-live="polite"
          data-testid="sueca-trick-winner"
        >
          {t('trickWinner', { team: state.leadPlayerIdx % 2 === 0 ? t('team.a') : t('team.b') })}
        </div>
      )}

      {isMobile && messageBox}
    </>
  );
}

function FooterHand({ game, state, t, view }: SuecaContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;
  const isPlayPhase = state.phase === SuecaPhase.PLAY;
  const canPlay = isPlayPhase && isHumanTurn;

  const trumpSymbol = suitSymbolAt(state.trumpSuit, '');
  // Spoken trump: the suit name (not the ♠♣♥♦ glyph, which SRs read poorly).
  const trumpSuitName = SUIT_KEYS[state.trumpSuit] ? t(`suitName.${SUIT_KEYS[state.trumpSuit]}`) : trumpSymbol;
  const trumpAriaLabel = t('trump', { suit: trumpSuitName });
  const trumpDesign = suitName(state.trumpSuit);
  const trumpIndices = trumpDesign
    ? humanPlayer.cards.reduce<number[]>((indices, card, index) => {
        if (card.design === trumpDesign) indices.push(index);
        return indices;
      }, [])
    : [];

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={cardWidth}
      isMobile={isMobile}
      dataTutorialPrefix="sueca"
      trumpIndices={trumpIndices}
      trumpAccessibleLabel={trumpAriaLabel}
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
    />
  );
}

function FooterStatus({ state, game, t, frontendHint }: SuecaContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  // Sueca states always contain one human player, and selection indices come
  // from that player's hand. Treat violations as invalid game state.
  const selectedCardNames = game.selectedCardIndices.map((idx) =>
    cardAlt((humanPlayer as SuecaPlayer).cards[idx] as Card),
  );

  return (
    <>
      <div data-testid="sueca-selection-status" role="status" aria-live="polite" aria-atomic="true">
        {selectedCardNames.length > 0
          ? t('selectedCards', { cards: selectedCardNames.join(t('listSeparator')) })
          : t('noCardSelected')}
      </div>

      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="sueca-hint-live" role="status" aria-live="polite">
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

function FooterActions({ game, state, t, busy }: SuecaContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;
  const isPlayPhase = state.phase === SuecaPhase.PLAY;
  const isTrickEnd = state.phase === SuecaPhase.TRICK_END;
  const isRoundEnd = state.phase === SuecaPhase.ROUND_END;
  const canPlay = isPlayPhase && isHumanTurn;

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

const suecaDefinition: GameDefinition<SuecaResponse, SuecaGame, Parameters<typeof suecaApi.exec>, SuecaView> = {
  key: 'sueca',
  useGame: useSuecaGame,
  useView: useSuecaView,
  tutorial: SUECA_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 6 },
  cli: () => ({
    parseCommand: parseSuecaCommand,
    formatResponse: formatSuecaState,
    helpText: SUECA_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === SuecaPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const humanTeam = humanIdx % 2;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerTeam === humanTeam,
    };
  },
  shell: ({ state }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      isHumanTurn: state.currentPlayerIdx === humanIdx,
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
          value: game.suecaConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetGamePoints',
          label: t('settings.targetGamePoints'),
          value: game.suecaConfig.targetGamePoints,
          options: TARGET_GAME_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetGamePoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) => (!view.isMobile ? <SuecaMessageBox state={state} /> : null),
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'sueca-action-buttons',
  resetTutorialId: 'sueca-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Sueca game page: a Portuguese 4-player (2 vs 2) trump trick-taker. */
export const SuecaPage = createGamePage(suecaDefinition);
