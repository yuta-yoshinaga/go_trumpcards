import { useEffect, useMemo } from 'react';
import type { mariasApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useMariasGame } from '../hooks/useMariasGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MariasResponse } from '../types/card';
import { MariasPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MARIAS_HELP, parseMariasCommand } from '../utils/cli/commands/mariasCommands';
import { formatMariasState } from '../utils/cli/formatters/mariasFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Suit-name i18n keys indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_KEYS = ['', 'spade', 'club', 'heart', 'diamond'] as const;

/** Card design string → suit number (1=♠ 2=♣ 3=♥ 4=♦), to align with SUIT_SYMBOLS / trumpSuit. */
const DESIGN_TO_SUIT: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

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

/** Renders the Mariáš game page: a Czech/Slovak 3-player Soloist-vs-Defenders trump trick-taker. */
export const MariasPage = withTutorial(MariasPageContent, 'marias', MARIAS_TUTORIAL_STEPS);

/** Inner content of the Mariáš page, wrapped by TutorialProvider. */
function MariasPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('marias');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    mariasConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useMariasGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('marias');
  const mariasCliConfig: CliGameConfig<MariasResponse, Parameters<typeof mariasApi.exec>> = useMemo(
    () => ({
      gameName: 'marias',
      parseCommand: parseMariasCommand,
      formatResponse: formatMariasState,
      helpText: MARIAS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, mariasCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('marias', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('marias', MARIAS_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="marias" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 10 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isPlayPhase = state.phase === MariasPhase.PLAY;
  const isTrickEnd = state.phase === MariasPhase.TRICK_END;
  const isRoundEnd = state.phase === MariasPhase.ROUND_END;
  const isGameEnd = state.phase === MariasPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const trumpSymbol = SUIT_SYMBOLS[state.trumpSuit] ?? '?';

  // Suits where the human holds both K (13) and Q (12) — a marriage worth 40 in
  // the trump suit, otherwise 20. Surfaced as a banner during play so the bonus
  // (otherwise only shown at round end) is visible while it can still be earned.
  const marriages = isPlayPhase
    ? [1, 2, 3, 4]
        .filter((suit) => {
          const cards = humanPlayer?.cards ?? [];
          const hasK = cards.some((c) => DESIGN_TO_SUIT[c.design] === suit && c.value === 13);
          const hasQ = cards.some((c) => DESIGN_TO_SUIT[c.design] === suit && c.value === 12);
          return hasK && hasQ;
        })
        .map((suit) => ({
          symbol: SUIT_SYMBOLS[suit] ?? '?',
          suitKey: SUIT_KEYS[suit],
          points: suit === state.trumpSuit ? 40 : 20,
        }))
    : [];

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.marias')}
      gameThemeBg={gameTheme.marias.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/marias"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerPlayer === humanIdx}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: mariasConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetPoints',
                    label: t('settings.targetPoints'),
                    value: mariasConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
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
                        <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">
                          {t('soloistBadge')}
                        </span>
                      )}
                    </div>
                  ))}
                </div>

                {/* Players: cards / tricks */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Round result: per-player card points + marriage */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {state.players.map((p) => (
                      <div key={p.id}>
                        <div>
                          {t('roundResult.cardPoints', {
                            name: playerName(p.id, p.isHuman),
                            points: state.roundCardPoints[p.id] ?? 0,
                          })}
                        </div>
                        {(state.roundMarriage[p.id] ?? 0) > 0 && (
                          <div>
                            {t('roundResult.marriage', {
                              name: playerName(p.id, p.isHuman),
                              points: state.roundMarriage[p.id] ?? 0,
                            })}
                          </div>
                        )}
                      </div>
                    ))}
                    {/* Soloist-vs-Defenders total comparison. Each side total is
                        cardPoints + marriage; the Soloist wins the round only when
                        their total strictly exceeds the two Defenders' combined total
                        (matching the domain's ScoreRound). The winning side is emphasised. */}
                    {(() => {
                      const sideTotal = (soloist: boolean) =>
                        state.players.reduce(
                          (sum, p) =>
                            p.isSoloist === soloist
                              ? sum + (state.roundCardPoints[p.id] ?? 0) + (state.roundMarriage[p.id] ?? 0)
                              : sum,
                          0,
                        );
                      const soloistTotal = sideTotal(true);
                      const defenderTotal = sideTotal(false);
                      const soloistWon = soloistTotal > defenderTotal;
                      return (
                        <div className="mt-1 pt-1 border-t border-ds-border-subtle" data-testid="marias-side-totals">
                          <span className={soloistWon ? 'text-ds-warning font-semibold' : ''}>
                            {t('roundResult.soloistTotal', { points: soloistTotal })}
                          </span>
                          <span className="mx-1">/</span>
                          <span className={soloistWon ? '' : 'text-ds-warning font-semibold'}>
                            {t('roundResult.defenderTotal', { points: defenderTotal })}
                          </span>
                        </div>
                      );
                    })()}
                  </div>
                )}
              </div>
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.marias.footer} px-4 py-2.5`}>
            {marriages.length > 0 && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="marias-marriage"
                role="status"
                aria-live="polite"
                aria-label={t('marriageAvailable', {
                  list: marriages.map((m) => `${t(`suitName.${m.suitKey}`)} K-Q +${m.points}`).join('、'),
                })}
              >
                <span aria-hidden="true">
                  {t('marriageAvailable', {
                    list: marriages.map((m) => `${m.symbol} K-Q (+${m.points})`).join('  '),
                  })}
                </span>
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="marias"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="marias-action-buttons">
              {canPlay && (
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
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="marias-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
