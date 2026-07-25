import { useEffect, useMemo, useState } from 'react';
import type { michiganApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import {
  ANTE_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useMichiganGame,
} from '../hooks/useMichiganGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MichiganResponse } from '../types/card';
import { MichiganPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MICHIGAN_HELP, parseMichiganCommand } from '../utils/cli/commands/michiganCommands';
import { formatMichiganState } from '../utils/cli/formatters/michiganFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { michiganBoodleGuides } from '../utils/michiganBoodleGuide';
import { michiganNextPlayable } from '../utils/michiganPlayable';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Number of center boodle cards (always 4: A♥, K♣, Q♦, J♠). */
const MICHIGAN_BOODLE_COUNT = 4;

/** Distributes `budget` chips as evenly as possible across `n` boodles (remainder to the first slots). */
function evenSplit(budget: number, n: number): number[] {
  const base = Math.floor(budget / n);
  const rem = budget - base * n;
  return Array.from({ length: n }, (_, i) => base + (i < rem ? 1 : 0));
}

/** Michigan tutorial step definitions. */
const MICHIGAN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="michigan-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="michigan-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="michigan-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="michigan-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MICHIGAN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MichiganPhase.BET]: 'bet',
  [MichiganPhase.PLAY]: 'play',
  [MichiganPhase.RESULT]: 'result',
};

/** Renders the Michigan (Newmarket) game page: a "stops" chip-betting family game. */
export const MichiganPage = withTutorial(MichiganPageContent, 'michigan', MICHIGAN_TUTORIAL_STEPS);

/** Inner content of the Michigan page, wrapped by TutorialProvider. */
function MichiganPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('michigan');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    michiganConfig,
    handleConfigChange,
    reset,
    handleBet,
    handlePlay,
    handleNextRound,
  } = useMichiganGame();

  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // Local boodle-bet distribution; re-seeded (even split) whenever a new bet turn begins.
  const [bets, setBets] = useState<number[]>(() => evenSplit(state?.betBudget ?? 0, MICHIGAN_BOODLE_COUNT));
  // biome-ignore lint/correctness/useExhaustiveDependencies: re-seed only when the budget / round / placed flag changes.
  useEffect(() => {
    setBets(evenSplit(state?.betBudget ?? 0, MICHIGAN_BOODLE_COUNT));
  }, [state?.betBudget, state?.roundNumber, state?.humanBetPlaced]);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('michigan');
  const cliConfig: CliGameConfig<MichiganResponse, Parameters<typeof michiganApi.exec>> = useMemo(
    () => ({
      gameName: 'michigan',
      parseCommand: parseMichiganCommand,
      formatResponse: formatMichiganState,
      helpText: MICHIGAN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('michigan', state);
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('michigan', MICHIGAN_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="michigan" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);

  const isBetPhase = state.phase === MichiganPhase.BET;
  const isPlayPhase = state.phase === MichiganPhase.PLAY;
  const isResultPhase = state.phase === MichiganPhase.RESULT;
  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const humanWonMatch = state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false);

  const showBetControls = isBetPhase && isHumanTurn && !state.humanBetPlaced && !isGameEnd;
  const canPlay = (i: number) => isPlayPhase && isHumanTurn && state.playableIndices.includes(i);
  const nextPlayable = michiganNextPlayable(state);
  const showPlayHints = isPlayPhase && isHumanTurn && !isGameEnd;

  const betSum = bets.reduce((a, b) => a + b, 0);
  const betRemaining = state.betBudget - betSum;
  const canPlaceBets = betRemaining === 0;

  // Per-boodle betting guidance: which boodles the human can recover (holds the
  // matching card) and which are already claimed, so bets can be biased sensibly.
  const boodleGuides = michiganBoodleGuides(state.boodles, humanPlayer?.cards ?? []);

  const adjustBet = (index: number, delta: number) => {
    setBets((prev) => {
      const next = [...prev];
      const value = (next[index] ?? 0) + delta;
      if (value < 0) return prev;
      // Derive the remaining budget from prev (not the render-closure value) so
      // rapid successive clicks can't overspend via a stale betRemaining.
      const remaining = state.betBudget - prev.reduce((a, b) => a + b, 0);
      if (delta > 0 && remaining <= 0) return prev;
      next[index] = value;
      return next;
    });
  };

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const playerBadge = (p: MichiganResponse['players'][number]): string =>
    p.isWinner ? t('badge.wentOut') : p.isCurrent ? t('badge.toPlay') : t('badge.waiting');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.michigan')}
      gameThemeBg={gameTheme.michigan.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn && !isResultPhase && !isGameEnd}
      gamePath="/michigan"
      gameEndFlag={isGameEnd}
      winShow={isResultPhase && (humanWonMatch || state.result > 0)}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>{t('chips', { amount: state.chips })}</span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
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
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: michiganConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'ante',
                    label: t('settings.ante'),
                    value: michiganConfig.ante,
                    options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('ante', v),
                  },
                  {
                    type: 'select',
                    id: 'startingChips',
                    label: t('settings.startingChips'),
                    value: michiganConfig.startingChips,
                    options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('startingChips', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: michiganConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="michigan-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('ante', { amount: state.ante })}</span>
              <span>{t('deadHand', { count: state.deadHandCount })}</span>
            </div>

            {/* Boodles (center betting cards) */}
            <div className="mb-2 p-2 rounded bg-black/30">
              <div className="mb-1 text-ds-text-primary text-sm">{t('boodlesTitle')}</div>
              <div className="flex flex-wrap justify-center gap-3">
                {state.boodles.map((b, i) => (
                  <div key={`boodle-${i}`} className="flex flex-col items-center">
                    <CardImage card={b.card} width={cardWidth} />
                    <div className="text-ds-text-muted text-xs mt-0.5">{t('boodleChips', { amount: b.chips })}</div>
                    <div className={`text-xs ${b.claimedBy >= 0 ? 'text-ds-success' : 'text-ds-text-muted'}`}>
                      {b.claimedBy >= 0
                        ? t('boodleClaimed', {
                            name: playerLabel(b.claimedBy, state.players[b.claimedBy]?.isHuman ?? false),
                          })
                        : t('boodleOpen')}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Current sequence */}
            <div className="mb-2 text-center text-ds-text-muted text-sm">
              {state.needNewSequence || state.seqSuit === 0
                ? t('sequence.new')
                : t('sequence.active', { suit: state.seqSuitName, value: state.seqHighValue })}
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="michigan-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className={`text-sm py-0.5 ${p.isWinner ? 'text-ds-success' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  {playerLabel(p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
                  {t('handCount', { count: p.cardCount })} · [{playerBadge(p)}]
                </div>
              ))}
            </div>

            {/* Revealed hands at result */}
            {isResultPhase && (
              <div className="mb-2 p-2 rounded bg-black/30">
                {state.players
                  .filter((p) => !p.isHuman && p.cards.length > 0)
                  .map((p) => (
                    <div key={p.id} className="mb-1">
                      <div className="text-ds-text-muted text-xs mb-0.5">{playerLabel(p.id, p.isHuman)}</div>
                      <div className="flex flex-wrap gap-1">
                        {p.cards.map((c, i) => (
                          <CardImage key={`${p.id}-${i}`} card={c} width={cardWidth} />
                        ))}
                      </div>
                    </div>
                  ))}
              </div>
            )}

            {/* Round result */}
            {isResultPhase && state.winnerIdx >= 0 && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div>
                  {t('roundResult.winner', {
                    name: playerLabel(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false),
                  })}
                </div>
              </div>
            )}

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
          <GameFooter className={`${gameTheme.michigan.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.cards.length > 0 ? (
              <div className="mb-2" data-tutorial="michigan-hand">
                <div className="text-ds-text-muted text-xs mb-0.5">{t('handLabel')}</div>
                {showPlayHints && (
                  <div className="text-ds-success text-sm mb-1" data-testid="michigan-next-hint">
                    {nextPlayable.isNewSequence
                      ? t('nextHint.new')
                      : t('nextHint.next', { suit: nextPlayable.suitName, value: nextPlayable.nextValue })}
                  </div>
                )}
                <div className="flex flex-wrap gap-1">
                  {humanPlayer.cards.map((c, i) => (
                    <button
                      key={`human-${i}`}
                      type="button"
                      onClick={() => canPlay(i) && handlePlay(i)}
                      disabled={!canPlay(i) || loading}
                      className={`rounded transition-all ${
                        canPlay(i)
                          ? 'cursor-pointer hover:-translate-y-2 ring-2 ring-ds-success'
                          : 'cursor-default opacity-60'
                      }`}
                      data-testid={`hand-card-${i}`}
                      data-playable={canPlay(i) ? 'true' : 'false'}
                      aria-label={t('playCardAria', { index: i })}
                    >
                      <CardImage card={c} width={cardWidth} />
                      {canPlay(i) && (
                        <span
                          className="block text-center text-ds-success text-xs mt-0.5"
                          data-testid={`playable-badge-${i}`}
                        >
                          {nextPlayable.isNewSequence ? t('nextHint.leadBadge') : t('nextHint.nextBadge')}
                        </span>
                      )}
                    </button>
                  ))}
                </div>
              </div>
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="michigan-hand">
                {t('handLabel')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col gap-2" data-tutorial="michigan-action-buttons">
              {showBetControls && (
                <div className="flex flex-col gap-2 p-2 rounded bg-black/20">
                  <div className="text-ds-text-primary text-sm">
                    {t('betPrompt', { budget: state.betBudget })} · {t('betRemaining', { amount: betRemaining })}
                  </div>
                  <div className="flex flex-wrap gap-3 items-start">
                    {state.boodles.map((b, i) => (
                      <div
                        key={`bet-${i}`}
                        className="flex flex-col items-center gap-1"
                        data-testid={`bet-boodle-${i}`}
                      >
                        <CardImage card={b.card} width={Math.round(cardWidth * 0.6)} />
                        <div className="flex flex-col items-center gap-0.5 min-h-[1rem]">
                          {boodleGuides[i]?.collectible && (
                            <span className="text-ds-success text-xs" data-testid={`bet-collectible-${i}`}>
                              {t('betCollectible')}
                            </span>
                          )}
                          {boodleGuides[i]?.claimed && (
                            <span className="text-ds-warning text-xs" data-testid={`bet-claimed-warning-${i}`}>
                              {t('betClaimedWarning')}
                            </span>
                          )}
                        </div>
                        <div className="flex items-center gap-1">
                          <button
                            type="button"
                            className={btnDanger}
                            onClick={() => adjustBet(i, -1)}
                            disabled={loading || (bets[i] ?? 0) <= 0}
                            aria-label={t('betMinusAria', { index: i })}
                          >
                            −
                          </button>
                          <span className="w-6 text-center text-ds-text-primary">{bets[i] ?? 0}</span>
                          <button
                            type="button"
                            className={btnSuccess}
                            onClick={() => adjustBet(i, 1)}
                            disabled={loading || betRemaining <= 0}
                            aria-label={t('betPlusAria', { index: i })}
                          >
                            +
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div>
                    <button
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleBet(bets)}
                      disabled={loading || !canPlaceBets}
                    >
                      {t('placeBets')}
                    </button>
                  </div>
                </div>
              )}

              {isPlayPhase && isHumanTurn && !isGameEnd && (
                <div className="text-ds-text-muted text-sm">{t('playPrompt')}</div>
              )}

              <div className="flex flex-wrap gap-2 items-center">
                {isResultPhase && !isGameEnd && (
                  <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                    {t('nextRound')}
                  </button>
                )}

                <GameResetButton
                  isGameEnd={isGameEnd}
                  onReset={handleManualReset}
                  requestConfirm={requestConfirm}
                  loading={loading}
                  dataTutorial="michigan-reset-button"
                />
              </div>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
