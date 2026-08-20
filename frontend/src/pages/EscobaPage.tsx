import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useEscobaGame } from '../hooks/useEscobaGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { gameTheme } from '../styles/gameTheme';
import type { Card, EscobaResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import {
  ESCOBA_HELP,
  type EscobaCliArgs,
  formatEscobaState,
  parseEscobaCommand,
} from '../utils/cli/commands/escobaCommands';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { sweepCelebration } from '../utils/sweepCelebration';

const DIFFICULTY_OPTIONS = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
];

const TARGET_SCORE_OPTIONS = [
  { value: '10', label: '10' },
  { value: '15', label: '15' },
  { value: '21', label: '21' },
];

/** Tutorial steps for Escoba. */
const ES_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="es-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="es-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="es-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="es-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="es-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/**
 * Collects the union of all table indices reachable by any capture set (sum 15)
 * for a given hand card. Used to highlight capturable table cards on the human turn.
 */
function captureCandidateIndices(handCaptures: number[][][], handIndex: number): Set<number> {
  const sets = handCaptures[handIndex] ?? [];
  const indices = new Set<number>();
  for (const set of sets) {
    for (const idx of set) indices.add(idx);
  }
  return indices;
}

/** Escoba card point value (mirrors the backend ScopaCardValue): A(1)–7 count as their pip;
 * J(11)=8, Q(12)=9, K(13)=10 in the 40-card Spanish deck. */
export function escobaCardValue(c: Card): number {
  return c.value <= 7 ? c.value : c.value - 3;
}

/** Live capture total: the selected hand card plus every selected table card, toward the target of 15. */
export function escobaSelectionSum(handCard: Card | null, tableCards: Card[], tableIndices: number[]): number {
  const base = handCard ? escobaCardValue(handCard) : 0;
  return tableIndices.reduce((sum, idx) => {
    const c = tableCards[idx];
    return c ? sum + escobaCardValue(c) : sum;
  }, base);
}

/** Renders the Escoba (エスコバ) game page. */
export const EscobaPage = withTutorial(EscobaPageContent, 'escoba', ES_TUTORIAL_STEPS);
function EscobaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('escoba');
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
  } = useEscobaGame();
  const { cardWidth } = useCardDimensions();

  const { playSound } = useSound();
  // Flash a badge and a click when any player's escobaCount rises — sweeping the
  // table is the game's highlight and is easy to miss amid fast CPU turns, which
  // is why Scopone got the same treatment (#3464). Escoba is free-for-all, so
  // "own" is simply the human rather than a partner team. escobaCount resets to 0
  // each round, so a drop clears a stale badge instead of re-firing (#4768).
  const [escobaCelebration, setEscobaCelebration] = useState<{ key: number; own: boolean } | null>(null);
  const prevEscobaRef = useRef<number[] | null>(null);
  useEffect(() => {
    if (!state) return;
    const current = state.players.map((p) => p.escobaCount);
    const prev = prevEscobaRef.current;
    prevEscobaRef.current = current;
    const action = sweepCelebration(prev, current, (i) => state.players[i]?.isHuman === true);
    if (action.kind === 'fire') {
      setEscobaCelebration((c) => ({ key: (c?.key ?? 0) + 1, own: action.own }));
      playSound('chipClick', { pitchVariation: 0.1 });
    } else if (action.kind === 'clear') {
      setEscobaCelebration(null);
    }
  }, [state, playSound]);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('escoba');
  const cliConfig: CliGameConfig<EscobaResponse, EscobaCliArgs> = useMemo(
    () => ({
      gameName: 'escoba',
      parseCommand: parseEscobaCommand,
      formatResponse: formatEscobaState,
      helpText: [...ESCOBA_HELP],
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

  // 早期 return より上。上のコメントが言うとおり、フック順を崩さないため。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('escoba', state);

  if (!state || !human) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.escoba.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag;
  const isRoundEnd = state.phase === 'roundEnd';
  const humanWon = isGameEnd && state.winnerIdx === human.id;
  const takeCandidateIndices =
    handIndex !== null && isHumanTurn ? captureCandidateIndices(state.handCaptures, handIndex) : new Set<number>();
  const canTake = isHumanTurn && handIndex !== null && tableIndices.length > 0;
  // **エスコバは強制捕獲** (#6163)。合計 15 を作れる組が場にあるなら、
  // ドメイン (Escoba.go の applyPlay) はその札を置く手を拒む。押せてしまうと
  // サーバまで飛んでエラーで返るだけなので、押させない。判定は既に配られて
  // いる handCaptures をそのまま読む——15 の組み合わせを画面で数え直さない。
  const mustCapture = handIndex !== null && (state.handCaptures[handIndex]?.length ?? 0) > 0;
  const canLay = isHumanTurn && handIndex !== null && tableIndices.length === 0 && !mustCapture;
  // Live 15-counter: once a hand card is picked, show its value plus the selected table cards.
  const selectedHandCard = handIndex !== null ? (human.cards[handIndex] ?? null) : null;
  const selectionSum = selectedHandCard ? escobaSelectionSum(selectedHandCard, state.tableCards, tableIndices) : null;
  const phaseName = isGameEnd ? t('phase.gameEnd') : t(`phase.${state.phase}`, t('phase.play'));
  const detail = state.lastRoundDetail;

  return (
    <GamePageShell
      title={tc('nav.escoba')}
      gameThemeBg={gameTheme.escoba.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/escoba"
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

            {/* Player scores + stock */}
            <div className="flex justify-center gap-4 text-xs text-ds-text-muted flex-wrap" data-testid="scores">
              <span className="font-semibold">{t('label.scores')}:</span>
              {state.players.map((p) => (
                <span key={p.id} data-testid={`player-score-${p.id}`}>
                  {t('label.playerScore', { id: p.id, score: p.score })}
                </span>
              ))}
              <span data-testid="stock-remaining">{t('label.stockRemaining', { count: state.stockRemaining })}</span>
            </div>

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="es-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1">
                      {tc('player.cpu', { id: p.id })} —{' '}
                      {t('label.cpuStats', { cards: p.handCount, captured: p.capturedCount, escoba: p.escobaCount })}
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
            <div className="relative py-3 bg-black/20 rounded-lg" data-tutorial="es-table-cards">
              {escobaCelebration && (
                <div
                  key={escobaCelebration.key}
                  className="absolute inset-x-0 -top-3 z-10 flex justify-center motion-safe:animate-bounce pointer-events-none"
                  role="status"
                  data-testid="escoba-celebration"
                >
                  <span
                    className={`rounded-full px-3 py-0.5 text-sm font-bold shadow-lg ${
                      escobaCelebration.own
                        ? 'bg-ds-accent text-ds-text-on-accent ring-2 ring-ds-accent'
                        : 'bg-ds-info text-white'
                    }`}
                  >
                    {escobaCelebration.own ? t('label.escobaBadgeOwn') : t('label.escobaBadge')}
                  </span>
                </div>
              )}
              {selectionSum !== null && (
                <div
                  role="status"
                  aria-live="polite"
                  data-testid="escoba-sum-indicator"
                  className={`text-center text-sm font-bold mb-1 ${
                    selectionSum === 15 ? 'text-ds-success' : selectionSum > 15 ? 'text-ds-error' : 'text-ds-text-muted'
                  }`}
                >
                  {t('sumIndicator', { sum: selectionSum, target: 15 })}
                </div>
              )}
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
            <div className="text-center" data-tutorial="es-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} —{' '}
                {t('label.humanStats', {
                  cards: human.handCount,
                  captured: human.capturedCount,
                  escoba: human.escobaCount,
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

              {/* Your captured pile — collapsible viewer (CPU piles stay count-only). */}
              <details className="mt-3 mx-auto max-w-md bg-black/25 rounded-lg" data-testid="captured-viewer">
                <summary className="cursor-pointer select-none px-3 py-2 text-xs text-ds-text-muted">
                  {t('captured.summary', { count: human.capturedCount })}
                </summary>
                <div className="px-3 pb-3">
                  {human.capturedCards.length === 0 ? (
                    <span className="text-ds-text-muted text-xs">{t('captured.empty')}</span>
                  ) : (
                    <div className="flex flex-wrap justify-center gap-1" data-testid="captured-cards">
                      {human.capturedCards.map((c, i) => (
                        <AnimatedCard key={i} card={c} width={cardWidth * 0.55} />
                      ))}
                    </div>
                  )}
                </div>
              </details>
            </div>

            {/* Round-end score breakdown */}
            {isRoundEnd && detail && (
              <div className="bg-black/25 rounded-lg p-3 text-sm" data-testid="round-detail">
                <div className="text-center font-semibold mb-2">{t('roundDetail.title')}</div>
                <table className="mx-auto text-xs">
                  <tbody>
                    <tr>
                      <td className="pr-3 text-ds-text-muted" />
                      {state.players.map((p) => (
                        <th key={p.id} className="px-2">
                          P{p.id}
                        </th>
                      ))}
                    </tr>
                    {(
                      [
                        ['cards', detail.cards],
                        ['espadas', detail.espadas],
                        ['sevens', detail.sevens],
                        ['oros', detail.oros],
                        ['escobas', detail.escobas],
                        ['gained', detail.gained],
                      ] as const
                    ).map(([key, vals]) => (
                      <tr key={key}>
                        <td className="pr-3 text-ds-text-muted">{t(`roundDetail.${key}`)}</td>
                        {vals.map((v, i) => (
                          <td key={i} className="px-2 text-center">
                            {v}
                          </td>
                        ))}
                      </tr>
                    ))}
                    <tr>
                      <td className="pr-3 text-ds-text-muted">{t('roundDetail.aceEspada')}</td>
                      <td className="px-2 text-center" colSpan={state.players.length}>
                        {t('label.player', { id: detail.aceEspada })}
                      </td>
                    </tr>
                    <tr>
                      <td className="pr-3 text-ds-text-muted">{t('roundDetail.seteEspada')}</td>
                      <td className="px-2 text-center" colSpan={state.players.length}>
                        {t('label.player', { id: detail.seteEspada })}
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
                    value: String(configInput.targetScore ?? 10),
                    options: TARGET_SCORE_OPTIONS,
                    onSelect: (v: string) => handleConfigChange('targetScore', Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.escoba.footer} px-4 py-2.5`}>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="es-actions">
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
              {isHumanTurn && mustCapture && (
                <span className="text-xs text-ds-warning" data-testid="escoba-must-capture">
                  {t('mustCapture')}
                </span>
              )}
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
                dataTutorial="es-reset-button"
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
