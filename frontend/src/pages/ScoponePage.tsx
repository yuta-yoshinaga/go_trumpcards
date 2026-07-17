import { useCallback, useMemo } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useScoponeGame } from '../hooks/useScoponeGame';
import { gameTheme } from '../styles/gameTheme';
import type { ScoponeResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import {
  formatScoponeState,
  parseScoponeCommand,
  SCOPONE_HELP,
  type ScoponeCliArgs,
} from '../utils/cli/commands/scoponeCommands';
import type { CliGameConfig } from '../utils/cli/types';

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

/** Tutorial steps for Scopone. */
const SP_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sp-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="sp-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/**
 * Collects the union of all table indices reachable by any capture set for a
 * given hand card. Used to highlight capturable table cards on the human turn.
 */
function captureCandidateIndices(handCaptures: number[][][], handIndex: number): Set<number> {
  const sets = handCaptures[handIndex] ?? [];
  const indices = new Set<number>();
  for (const set of sets) {
    for (const idx of set) indices.add(idx);
  }
  return indices;
}

/** Renders the Scopone (スコポーネ) game page. */
export const ScoponePage = withTutorial(ScoponePageContent, 'scopone', SP_TUTORIAL_STEPS);
function ScoponePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('scopone');
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
  } = useScoponeGame();
  const { cardWidth } = useCardDimensions();

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('scopone');
  const cliConfig: CliGameConfig<ScoponeResponse, ScoponeCliArgs> = useMemo(
    () => ({
      gameName: 'scopone',
      parseCommand: parseScoponeCommand,
      formatResponse: formatScoponeState,
      helpText: [...SCOPONE_HELP],
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => handleResetWithConfig(), [handleResetWithConfig]);

  // Hooks below must run unconditionally — they're computed before the early-return
  // skeleton guard so the hook order stays stable on the first render when `state`
  // is still null.
  const human = state && state.players.length >= 4 ? state.players.find((p) => p.isHuman) : null;
  const isHumanTurn = !!state && !!human && state.currentTurn === human.id && state.isHumanTurn && !state.gameEndFlag;

  if (!state || !human) {
    return (
      <GameSkeleton
        gameKey="scopone"
        layout={{ kind: 'trick-taking', opponents: 3, trickArea: true, footerHandSize: 10 }}
      />
    );
  }

  const isGameEnd = state.gameEndFlag;
  const isRoundEnd = state.phase === 'roundEnd';
  const humanTeam = human.team;
  const humanWon = isGameEnd && state.winnerTeam === humanTeam;
  const takeCandidateIndices =
    handIndex !== null && isHumanTurn ? captureCandidateIndices(state.handCaptures, handIndex) : new Set<number>();
  const canTake = isHumanTurn && handIndex !== null && tableIndices.length > 0;
  const canLay = isHumanTurn && handIndex !== null && tableIndices.length === 0;
  const phaseName = isGameEnd ? t('phase.gameEnd') : t(`phase.${state.phase}`, t('phase.play'));
  const detail = state.lastRoundDetail;

  return (
    <GamePageShell
      title={tc('nav.scopone')}
      gameThemeBg={gameTheme.scopone.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/scopone"
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

            {/* Team scores */}
            <div className="flex justify-center gap-4 text-xs text-ds-text-muted" data-testid="team-scores">
              <span className="font-semibold">{t('label.teamScores')}:</span>
              {state.teamScores.map((sc, team) => (
                <span key={team} data-testid={`team-score-${team}`}>
                  {t('label.teamScore', { team, score: sc })}
                </span>
              ))}
            </div>

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="sp-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1">
                      {tc('player.cpu', { id: p.id })} ({t('label.team', { team: p.team })}) —{' '}
                      {t('label.cpuStats', { cards: p.handCount, captured: p.capturedCount, scopa: p.scopaCount })}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.handCount, 10) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.4} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Table cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="sp-table-cards">
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
            <div className="text-center" data-tutorial="sp-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} ({t('label.team', { team: humanTeam })}) —{' '}
                {t('label.humanStats', {
                  cards: human.handCount,
                  captured: human.capturedCount,
                  scopa: human.scopaCount,
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

            {/* Round-end score breakdown */}
            {isRoundEnd && detail && (
              <div className="bg-black/25 rounded-lg p-3 text-sm" data-testid="round-detail">
                <div className="text-center font-semibold mb-2">{t('roundDetail.title')}</div>
                <table className="mx-auto text-xs">
                  <tbody>
                    {(
                      [
                        ['cards', detail.cards],
                        ['diamonds', detail.diamonds],
                        ['sevens', detail.sevens],
                        ['scopas', detail.scopas],
                        ['gained', detail.gained],
                      ] as const
                    ).map(([key, vals]) => (
                      <tr key={key}>
                        <td className="pr-3 text-ds-text-muted">{t(`roundDetail.${key}`)}</td>
                        <td className="px-2">{vals[0]}</td>
                        <td className="px-2">{vals[1]}</td>
                      </tr>
                    ))}
                    <tr>
                      <td className="pr-3 text-ds-text-muted">{t('roundDetail.settebello')}</td>
                      <td className="px-2" colSpan={2}>
                        {t('label.team', { team: detail.settebello })}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
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
                    options: DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
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
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.scopone.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="sp-actions">
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
                dataTutorial="sp-reset-button"
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
