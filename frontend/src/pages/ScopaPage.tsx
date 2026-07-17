import { useCallback, useMemo } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useScopaGame } from '../hooks/useScopaGame';
import { gameTheme } from '../styles/gameTheme';
import type { ScopaResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import {
  formatScopaState,
  parseScopaCommand,
  SCOPA_HELP,
  type ScopaCliArgs,
} from '../utils/cli/commands/scopaCommands';
import type { CliGameConfig } from '../utils/cli/types';
import { scopaTakeCandidates } from '../utils/scopaTakeCandidates';

const DIFFICULTY_OPTIONS = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
];

const TARGET_SCORE_OPTIONS = [
  { value: '11', label: '11' },
  { value: '16', label: '16' },
  { value: '21', label: '21' },
];

/** Tutorial steps for Scopa. */
const SC_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sc-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="sc-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sc-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sc-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sc-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Scopa (スコパ) game page. */
export const ScopaPage = withTutorial(ScopaPageContent, 'scopa', SC_TUTORIAL_STEPS);
function ScopaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('scopa');
  const {
    state,
    loading,
    error,
    callApi,
    handIndex,
    setHandIndex,
    tableIndices,
    toggleTable,
    configInput,
    handleConfigChange,
    play,
    handleNextRound,
    handleResetWithConfig,
    retry,
  } = useScopaGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('scopa', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('scopa');
  const cliConfig: CliGameConfig<ScopaResponse, ScopaCliArgs> = useMemo(
    () => ({
      gameName: 'scopa',
      parseCommand: parseScopaCommand,
      formatResponse: formatScopaState,
      helpText: [...SCOPA_HELP],
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => handleResetWithConfig(), [handleResetWithConfig]);

  // Hooks below must run unconditionally — they're computed before the early-return
  // skeleton guard so the hook order stays stable on the first render when `state`
  // is still null.
  const human = state && state.players.length >= 2 ? state.players[0] : null;
  const isHumanTurn = !!state && state.currentTurn === 0 && !state.gameEndFlag;

  if (!state || !human) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.scopa.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag;
  // A round finished but the game continues — surface the round boundary and the
  // "next round" control instead of letting the deal-again flow slip by silently.
  const isRoundEnd = state.phase === 'roundEnd';
  const humanWon = isGameEnd && state.roundWinners.includes(0);
  const takeCandidateIndices =
    handIndex !== null && isHumanTurn
      ? scopaTakeCandidates(state.tableCards, human.cards[handIndex]?.value ?? 0).indices
      : new Set<number>();
  const canTake = isHumanTurn && handIndex !== null && tableIndices.length > 0;
  const canLay = isHumanTurn && handIndex !== null && tableIndices.length === 0;
  // Announce how many table cards the selected hand card could capture, since the
  // green candidate rings are purely visual.
  const takeCandidateAnnounce =
    handIndex !== null && isHumanTurn
      ? takeCandidateIndices.size > 0
        ? t('label.takeCandidateAnnounce', { count: takeCandidateIndices.size })
        : t('label.noTakeCandidate')
      : '';
  const phaseName = isGameEnd ? t('phase.end') : t(`phase.${state.phase}`, t('phase.play'));

  return (
    <GamePageShell
      title={tc('nav.scopa')}
      gameThemeBg={gameTheme.scopa.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/scopa"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
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
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            {error && (
              <button type="button" onClick={retry} className="text-ds-error underline">
                {error}
              </button>
            )}

            {/* CPU player */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="sc-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1">
                      {tc('player.cpu', { id: p.id })} —{' '}
                      {t('label.cpuStats', { cards: p.cardCount, score: p.totalScore })}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.cardCount, 8) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.45} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Table cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="sc-table-cards">
              <div className="sr-only" role="status" aria-live="polite" data-testid="sc-take-candidate-live">
                {takeCandidateAnnounce}
              </div>
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('label.tableCards')}</div>
              <div className="flex justify-center gap-2 min-h-[60px] flex-wrap">
                {state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('label.tableEmpty')}</span>
                ) : (
                  state.tableCards.map((c, i) => {
                    const isCandidate = takeCandidateIndices.has(i);
                    return (
                      <button
                        key={i}
                        type="button"
                        onClick={() => isHumanTurn && toggleTable(i)}
                        disabled={!isHumanTurn}
                        aria-pressed={tableIndices.includes(i)}
                        aria-label={`${cardAlt(c)}${
                          tableIndices.includes(i)
                            ? ` ${t('label.selected')}`
                            : isCandidate
                              ? ` ${t('label.takeCandidate')}`
                              : ''
                        }`}
                        className={`rounded transition-all ${
                          tableIndices.includes(i)
                            ? 'ring-2 ring-ds-warning -translate-y-1'
                            : isCandidate
                              ? 'ring-2 ring-ds-success motion-safe:animate-pulse'
                              : ''
                        } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                        data-testid={`table-card-${i}`}
                        data-take-candidate={isCandidate || undefined}
                      >
                        <AnimatedCard card={c} width={cardWidth * 0.9} />
                      </button>
                    );
                  })
                )}
              </div>
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="sc-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} —{' '}
                {t('label.humanStats', {
                  cards: human.cardCount,
                  captured: human.capturedCount,
                  scopa: human.scopaCount,
                  score: human.totalScore,
                })}
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    onClick={() => isHumanTurn && setHandIndex(handIndex === i ? null : i)}
                    disabled={!isHumanTurn}
                    className={`rounded transition-all ${
                      handIndex === i ? 'ring-2 ring-ds-info -translate-y-2' : ''
                    } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    data-testid={`hand-card-${i}`}
                  >
                    <AnimatedCard card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            {isRoundEnd && (
              <div
                role="status"
                className="text-center text-sm font-medium text-ds-info bg-ds-info/10 border border-ds-info/30 rounded-lg py-2 px-3"
                data-testid="sc-round-end-banner"
              >
                {t('label.roundEnd')}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: String(configInput.cpuDifficulty ?? 1),
                    options: DIFFICULTY_OPTIONS,
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
                  },
                  {
                    type: 'select' as const,
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: String(configInput.targetScore ?? 11),
                    options: TARGET_SCORE_OPTIONS,
                    onSelect: (v: string) => handleConfigChange('targetScore', Number.parseInt(v, 10)),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.scopa.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="sc-actions">
              <button
                type="button"
                onClick={play}
                disabled={loading || !canTake}
                className="px-4 py-2 rounded-lg bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="take-button"
              >
                {t('button.take')}
              </button>
              <button
                type="button"
                onClick={play}
                disabled={loading || !canLay}
                className="px-4 py-2 rounded-lg bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="lay-button"
              >
                {t('button.lay')}
              </button>
              {isRoundEnd && (
                <button
                  type="button"
                  onClick={handleNextRound}
                  disabled={loading}
                  className="px-4 py-2 rounded-lg bg-ds-info/70 text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                  data-testid="next-round-button"
                >
                  {t('button.nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sc-reset-button"
              />
              <ActionLogSection
                isEndPhase={isGameEnd}
                actionLog={actionLog}
                showActionLog={showActionLog}
                hideActionLog={hideActionLog}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
