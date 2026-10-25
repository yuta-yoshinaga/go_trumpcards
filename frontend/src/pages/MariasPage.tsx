import { useEffect } from 'react';
import type { mariasApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useMariasGame } from '../hooks/useMariasGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { MariasResponse } from '../types/card';
import { MariasPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt, suitSymbolAt } from '../utils/cardAlt';
import { MARIAS_HELP, parseMariasCommand } from '../utils/cli/commands/mariasCommands';
import { formatMariasState } from '../utils/cli/formatters/mariasFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Mariáš tutorial step definitions. */
const MARIAS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="marias-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marias-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marias-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="marias-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="marias-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MARIAS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MariasPhase.PLAY]: 'play',
  [MariasPhase.TRICK_END]: 'trickEnd',
  [MariasPhase.ROUND_END]: 'roundEnd',
  [MariasPhase.GAME_END]: 'gameEnd',
};

type MariasGame = ReturnType<typeof useMariasGame>;
type MariasView = ReturnType<typeof useMariasView>;
type MariasContext = GamePageContext<MariasResponse, MariasGame, MariasView>;

function useMariasView({ game }: { game: MariasGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('marias', MARIAS_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function MariasMessageBox({ state }: Pick<MariasContext, 'state'>) {
  return <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />;
}

function Board({ state, t, view }: MariasContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);

  const isPlayPhase = state.phase === MariasPhase.PLAY;
  const isTrickEnd = state.phase === MariasPhase.TRICK_END;
  const isRoundEnd = state.phase === MariasPhase.ROUND_END;
  const isGameEnd = state.phase === MariasPhase.GAME_END || state.gameEndFlag;

  const trumpSymbol = suitSymbolAt(state.trumpSuit, '');
  const marriageDetails = (playerIdx: number) =>
    (state.roundMarriageSuits[playerIdx] ?? [])
      .map(
        (marriage) =>
          `${t(`suitName.${['', 'spade', 'club', 'heart', 'diamond'][marriage.suit] ?? ''}`)}: ${marriage.points}`,
      )
      .join(', ');
  const playerRoleLabel = (isSoloist: boolean) => t(isSoloist ? 'role.soloist' : 'role.defender');

  const sideTotal = (soloist: boolean) =>
    state.players.reduce(
      (sum, player) =>
        player.isSoloist === soloist
          ? sum + (state.roundCardPoints[player.id] ?? 0) + (state.roundMarriage[player.id] ?? 0)
          : sum,
      0,
    );
  const soloistTotal = sideTotal(true);
  const defenderTotal = sideTotal(false);
  const soloistWon = soloistTotal > defenderTotal;

  const messageBox = <MariasMessageBox state={state} />;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        {/* **切り札の決まり方が見えない。**入札でも固定でもなく、ソリストの手札という
            非公開情報から機械的に決まる。記号の位置は変えず、規則を名前と
            ツールチップの両方で添える ── `title` はキーボードや支援技術に
            届かないので、`sr-only` にも同じ文言を置く (#6443)。 */}
        <span className="mr-4" title={t('trumpRule')} data-testid="marias-trump">
          {t('trump', { suit: trumpSymbol })}
          <span className="sr-only"> ({t('trumpRule')})</span>
        </span>
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
            winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
            winnerLabel={
              isTrickEnd
                ? t('previousTrickWinner', { name: findPlayerName(state.players, state.lastTrickWinner) })
                : undefined
            }
            cardAriaLabelFor={(player, card) =>
              t('trickPlay', { name: playerName(player.id, player.isHuman), card: cardAlt(card) })
            }
            dataTutorial="marias-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="marias-info">
          {/* Per-player game-point scores with Soloist badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                <span className={p.isSoloist ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                </span>
                {p.isSoloist && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('soloistBadge')}</span>
                )}
              </div>
            ))}
          </div>

          {/* Players: cards / tricks */}
          <div
            className="mb-1 text-ds-text-muted text-sm py-0.5"
            data-testid="marias-human-seat"
            data-tutorial="marias-human-seat"
          >
            {humanPlayer && (
              <>
                {playerName(humanPlayer.id, true)}: {playerRoleLabel(humanPlayer.isSoloist)} |{' '}
                {t('cards', { count: humanPlayer.cardCount })} | {t('tricks', { count: humanPlayer.trickCount })}
              </>
            )}
          </div>
          <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
            <div className="flex flex-wrap gap-x-3 gap-y-1">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {playerRoleLabel(p.isSoloist)} | {t('cards', { count: p.cardCount })}{' '}
                    | {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
            </div>
          </CpuAccordion>

          {/* Round result: per-player card points + marriage */}
          <div data-testid="marias-round-progress-live" role="status" aria-live="polite" aria-atomic="true">
            {(isPlayPhase || isTrickEnd || isRoundEnd || isGameEnd) && (
              <details
                open={!isMobile || undefined}
                className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              >
                <summary className="cursor-pointer select-none text-ds-text-primary">
                  {t(isRoundEnd || isGameEnd ? 'roundResult.title' : 'roundResult.progressTitle')}
                  <span className="ml-2" data-testid="marias-side-totals">
                    <span
                      className={`whitespace-nowrap ${(isRoundEnd || isGameEnd) && soloistWon ? 'text-ds-warning font-semibold' : ''}`}
                    >
                      {t('roundResult.soloistTotal', { points: soloistTotal })}
                    </span>
                    {' / '}
                    <span
                      className={`whitespace-nowrap ${(isRoundEnd || isGameEnd) && !soloistWon ? 'text-ds-warning font-semibold' : ''}`}
                    >
                      {t('roundResult.defenderTotal', { points: defenderTotal })}
                    </span>
                  </span>
                </summary>
                <div className="pt-1">
                  {state.players.map((p) => {
                    const cardPoints = state.roundCardPoints[p.id] ?? 0;
                    const marriage = state.roundMarriage[p.id] ?? 0;

                    return (
                      <div key={p.id}>
                        <div>
                          {t('roundResult.cardPoints', {
                            name: playerName(p.id, p.isHuman),
                            points: cardPoints,
                          })}
                        </div>
                        {marriage > 0 && (
                          <div>
                            {t('roundResult.marriage', {
                              name: playerName(p.id, p.isHuman),
                              points: marriage,
                            })}
                            <span className="ml-1">({marriageDetails(p.id)})</span>
                          </div>
                        )}
                      </div>
                    );
                  })}
                  {/* Soloist-vs-Defenders total comparison. Each side total is
                    cardPoints + marriage; the Soloist wins the round only when
                    their total strictly exceeds the two Defenders' combined total
                    (matching the domain's ScoreRound). The winning side is emphasised. */}
                </div>
              </details>
            )}
          </div>
        </div>
      </div>

      {isMobile && messageBox}
    </>
  );
}

function FooterHand({ game, state, t, view }: MariasContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isPlayPhase = state.phase === MariasPhase.PLAY;
  const isHumanTurn = state.isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  const trumpDesign = ['', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'][state.trumpSuit];
  const trumpIndices = humanPlayer
    ? humanPlayer.cards.reduce<number[]>((indices, card, index) => {
        if (card.design === trumpDesign) indices.push(index);
        return indices;
      }, [])
    : [];

  const marriageDetails = (playerIdx: number) =>
    (state.roundMarriageSuits[playerIdx] ?? [])
      .map(
        (marriage) =>
          `${t(`suitName.${['', 'spade', 'club', 'heart', 'diamond'][marriage.suit] ?? ''}`)}: ${marriage.points}`,
      )
      .join(', ');

  // **結婚ボーナスは配った時点で確定している。**`detectMarriages` はラウンド
  // 開始時に一度だけ走って `roundMarriage` に加点し、以後 K・Q を場に出しても
  // 点数は動かない (Marias.go の startRound → detectMarriages)。
  //
  // 以前はここで毎レンダー手札を走査して「K と Q を両方持っているか」を
  // 判定していたため、どちらかを出した瞬間にバナーが消え、**「出したので
  // ボーナスを失った」という誤解**を与えていた (#4759)。確定済みの点数
  // そのものを根拠にする。
  const marriagePoints = isPlayPhase ? (state.roundMarriage[humanPlayer?.id ?? 0] ?? 0) : 0;

  return (
    <>
      {marriagePoints > 0 && (
        <div
          className="mb-1 text-center text-sm text-ds-accent font-semibold"
          data-testid="marias-marriage"
          role="status"
          aria-live="polite"
        >
          {t('marriageEarned', { points: marriagePoints, details: marriageDetails(humanIdx) })}
        </div>
      )}
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={cardWidth}
          isMobile={isMobile}
          dataTutorialPrefix="marias"
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
          trumpIndices={trumpIndices}
          trumpAccessibleLabel={t('trumpCardDescription')}
        />
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: MariasContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="marias-hint-live" role="status" aria-live="polite">
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

function FooterActions({ game, state, t, busy }: MariasContext) {
  const isPlayPhase = state.phase === MariasPhase.PLAY;
  const isTrickEnd = state.phase === MariasPhase.TRICK_END;
  const isRoundEnd = state.phase === MariasPhase.ROUND_END;
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

const mariasDefinition: GameDefinition<MariasResponse, MariasGame, Parameters<typeof mariasApi.exec>, MariasView> = {
  key: 'marias',
  useGame: useMariasGame,
  useView: useMariasView,
  tutorial: MARIAS_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 10 },
  cli: () => ({
    parseCommand: parseMariasCommand,
    formatResponse: formatMariasState,
    helpText: MARIAS_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === MariasPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerPlayer === humanIdx,
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
          value: game.mariasConfig.cpuDifficulty,
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
          value: game.mariasConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) => (!view.isMobile ? <MariasMessageBox state={state} /> : null),
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'marias-action-buttons',
  resetTutorialId: 'marias-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Mariáš game page: a Czech/Slovak 3-player Soloist-vs-Defenders trump trick-taker. */
export const MariasPage = createGamePage(mariasDefinition);
