import { useCallback, useMemo, useState } from 'react';
import { yanivApi } from '../api/gameApi';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { gameTheme } from '../styles/gameTheme';
import type { YanivResponse } from '../types/card';
import { YanivPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { classifyYanivDiscard } from '../utils/yanivCombos';
import { isPickupable } from '../utils/yanivPickup';

type YanivArgs = Parameters<typeof yanivApi.exec>;

const DIFFICULTY_OPTIONS = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
];

const SCORE_LIMIT_OPTIONS = [
  { value: '100', label: '100' },
  { value: '150', label: '150' },
  { value: '200', label: '200' },
  { value: '250', label: '250' },
];

/** Tutorial steps for the Yaniv game. */
const YANIV_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="y-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="y-discard-area"]',
    messageKey: 'tutorial.discardArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="y-player-hand"]', messageKey: 'tutorial.playerHand', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="y-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="y-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Yaniv (ヤニブ) game page. */
export const YanivPage = withTutorial(YanivPageContent, 'yaniv', YANIV_TUTORIAL_STEPS);

/** Inner content of the Yaniv page. */
function YanivPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('yaniv');
  const { state, loading, error, exec: execApi, retry } = useGameApi(yanivApi.exec);
  const { cardWidth } = useCardDimensions();
  const [cpuDifficulty, setCpuDifficulty] = useState(1);
  const [scoreLimit, setScoreLimit] = useState(200);
  const [selected, setSelected] = useState<number[]>([]);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('yaniv', state);

  const handleReset = useCallback(() => {
    setSelected([]);
    return execApi('reset', { config: { cpuDifficulty, scoreLimit } });
  }, [execApi, cpuDifficulty, scoreLimit]);

  const handleDrawStock = useCallback(() => execApi('drawstock'), [execApi]);
  const handleDrawPickup = useCallback((end: number) => execApi('drawpickup', { end }), [execApi]);
  const handleYaniv = useCallback(() => execApi('yaniv'), [execApi]);
  const handleNextRound = useCallback(() => execApi('nextround'), [execApi]);
  const handleDiscard = useCallback(() => {
    if (selected.length === 0) return;
    execApi('discard', { cardIndices: selected });
    setSelected([]);
  }, [execApi, selected]);

  const toggleCard = useCallback((i: number) => {
    setSelected((prev) => (prev.includes(i) ? prev.filter((x) => x !== i) : [...prev, i]));
  }, []);

  useMountReset(execApi);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('yaniv');
  const cliConfig: CliGameConfig<YanivResponse, YanivArgs> = useMemo(
    () => ({
      gameName: 'yaniv',
      parseCommand: (input: string): CliParseResult<YanivArgs> => {
        const parts = input.trim().toLowerCase().split(/\s+/);
        const cmd = parts[0];
        if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
        if (cmd === 'd' || cmd === 'discard') {
          const idxs = parts.slice(1).map((p) => Number.parseInt(p, 10));
          if (idxs.length === 0 || idxs.some(Number.isNaN)) return { error: 'Usage: d <idx> [idx...]' };
          return { args: ['discard', { cardIndices: idxs }] };
        }
        if (cmd === 'y' || cmd === 'yaniv') return { args: ['yaniv'] };
        if (cmd === 'ds' || cmd === 'drawstock') return { args: ['drawstock'] };
        if (cmd === 'dp' || cmd === 'drawpickup') {
          const end = Number.parseInt(parts[1], 10);
          if (Number.isNaN(end)) return { error: 'Usage: dp <0|1>' };
          return { args: ['drawpickup', { end }] };
        }
        if (cmd === 'nr' || cmd === 'nextround') return { args: ['nextround'] };
        if (cmd === 'log' || cmd === 'l') return { args: ['log'] };
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: (s: YanivResponse) => {
        const lines: string[] = [];
        lines.push(`Round ${s.roundNumber} | Stock ${s.drawPileCount} | Phase ${s.phase}`);
        if (s.pickupCards.length > 0) {
          lines.push(`Discard ends: ${s.pickupCards.map((c) => `${c.design} ${c.value}`).join(', ')}`);
        }
        for (const p of s.players) {
          const tag = p.isHuman ? 'You' : `CPU${p.id}`;
          const total = p.isHuman || s.gameEndFlag || s.phase === YanivPhase.ROUND_END ? p.handTotal : '?';
          lines.push(
            `${tag}: penalty ${p.score}${p.isEliminated ? ' (OUT)' : ''}, hand ${total}, ${p.cardCount} cards`,
          );
        }
        if (s.message) lines.push(s.message);
        return lines.join('\n');
      },
      helpText: [
        'd <idx...>       - Discard cards (single / set / run)',
        'y / yaniv        - Declare Yaniv (hand total <= 5)',
        'ds / drawstock   - Draw from stock',
        'dp <0|1>         - Take an end of the last discard',
        'nr / nextround   - Next round',
        'r / reset        - Reset game',
        'l / log          - Show action log',
      ],
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state || state.players.length < 4)
    return <GameSkeleton gameKey="yaniv" layout={{ kind: 'centered', rows: [3, 3] }} />;

  const isGameEnd = state.gameEndFlag || state.phase === YanivPhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const isDiscard = state.phase === YanivPhase.DISCARD;
  const isDraw = state.phase === YanivPhase.DRAW;
  const isRoundEnd = state.phase === YanivPhase.ROUND_END;
  const isHumanTurn = state.currentPlayerIdx === 0 && !isGameEnd && (isDraw || isDiscard);
  const human = state.players[0];
  const reveal = isGameEnd || isRoundEnd;
  const canYaniv = isHumanTurn && isDiscard && human.handTotal <= 5;
  const phaseName = isGameEnd
    ? t('phase.end')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isDiscard
        ? t('phase.discard')
        : t('phase.draw');
  const pickup = state.pickupCards;

  // Pre-validate the selected discard combination on the client (warn-only;
  // the server still validates). Mirrors the domain `YanivValidCombo`.
  const selectedCards = selected.map((i) => human.cards[i]).filter((c): c is NonNullable<typeof c> => c != null);
  const discardCheck =
    isDiscard && isHumanTurn && selectedCards.length > 0 ? classifyYanivDiscard(selectedCards) : null;
  const discardWarning = discardCheck?.reasonKey ? t(discardCheck.reasonKey) : null;

  return (
    <GamePageShell
      title={tc('nav.yaniv')}
      gameThemeBg={gameTheme.yaniv.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/yaniv"
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

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="y-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className={p.isEliminated ? 'text-center opacity-40' : 'text-center'}>
                    <div className="text-xs text-ds-text-muted mb-1">
                      {tc('player.cpu', { id: p.id })}
                      {p.isEliminated && <span className="ml-1">💀</span>}
                    </div>
                    <div className="text-[10px] text-ds-text-muted mb-1">
                      {t('label.score')}: {p.score} · {t('label.hand')}: {reveal ? p.handTotal : '?'}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {reveal
                        ? p.cards.map((c, i) => <AnimatedCard key={i} card={c} width={cardWidth * 0.6} />)
                        : Array.from({ length: p.cardCount }, (_, i) => (
                            <AnimatedCardBack key={i} width={cardWidth * 0.6} />
                          ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Stock + discard ends */}
            <div className="py-3 bg-black/20 rounded-lg flex justify-center gap-8" data-tutorial="y-discard-area">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted mb-1">
                  {t('label.stock')}: {state.drawPileCount}
                </div>
                {state.drawPileCount > 0 ? <AnimatedCardBack width={cardWidth * 0.8} /> : null}
              </div>
              <div className="text-center">
                <div className="text-xs text-ds-text-muted mb-1">{t('label.discardPile')}</div>
                {isDraw && isHumanTurn && pickup.length > 1 && (
                  <div className="text-[10px] text-ds-info mb-1" data-testid="pickup-hint">
                    {t('pickup.hint')}
                  </div>
                )}
                <div className="flex gap-1 justify-center">
                  {pickup.length > 0 ? (
                    pickup.map((c, i) => {
                      const pickable = isPickupable(i, pickup.length);
                      const active = pickable && isDraw && isHumanTurn;
                      const blocked = !pickable && isDraw && isHumanTurn;
                      return (
                        <div key={i} className="relative" title={blocked ? t('pickup.disabledReason') : undefined}>
                          <button
                            type="button"
                            data-testid={`pickup-card-${i}`}
                            onClick={() => active && handleDrawPickup(i === 0 ? 0 : 1)}
                            disabled={!active || loading}
                            aria-disabled={blocked || undefined}
                            aria-label={active ? t('pickup.endLabel') : undefined}
                            className={
                              active
                                ? 'rounded ring-2 ring-ds-info cursor-pointer hover:opacity-90'
                                : 'rounded opacity-70 cursor-default'
                            }
                          >
                            <AnimatedCard card={c} width={cardWidth * 0.8} />
                          </button>
                          {active && (
                            <span
                              data-testid={`pickup-badge-${i}`}
                              className="absolute -top-1 -right-1 rounded bg-ds-info text-white text-[9px] font-bold px-1 pointer-events-none"
                            >
                              {t('pickup.badge')}
                            </span>
                          )}
                        </div>
                      );
                    })
                  ) : (
                    <span>—</span>
                  )}
                </div>
              </div>
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="y-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} · {t('label.score')}: {human.score} · {t('label.hand')}:{' '}
                <span
                  data-testid="hand-total-badge"
                  title={t('handTotalHelp')}
                  className={`inline-block rounded px-1.5 font-bold ${
                    human.handTotal <= 5 ? 'bg-ds-success text-white' : human.handTotal <= 10 ? 'bg-ds-warning/40' : ''
                  }`}
                >
                  {human.handTotal}
                </span>
              </div>
              {isDiscard && isHumanTurn && (
                <div className="text-xs text-ds-text-muted mb-1">{t('label.selectCard')}</div>
              )}
              <div className="flex justify-center gap-2 flex-wrap">
                {human.cards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    data-testid={`hand-card-${i}`}
                    onClick={() => isDiscard && isHumanTurn && toggleCard(i)}
                    disabled={!isDiscard || !isHumanTurn}
                    className={
                      selected.includes(i)
                        ? 'rounded transition-all ring-2 ring-ds-info -translate-y-2 cursor-pointer hover:opacity-90'
                        : isDiscard && isHumanTurn
                          ? 'rounded transition-all cursor-pointer hover:opacity-90'
                          : 'rounded transition-all cursor-default'
                    }
                  >
                    <AnimatedCard card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
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
                    value: String(cpuDifficulty),
                    options: DIFFICULTY_OPTIONS,
                    onSelect: (v: string) => setCpuDifficulty(Number.parseInt(v, 10)),
                  },
                  {
                    type: 'select' as const,
                    id: 'scoreLimit',
                    label: t('settings.scoreLimit'),
                    value: String(scoreLimit),
                    options: SCORE_LIMIT_OPTIONS,
                    onSelect: (v: string) => setScoreLimit(Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.yaniv.footer} px-4 py-2.5`}>
            {discardWarning && (
              <div role="status" data-testid="discard-warning" className="mb-2 text-center text-xs text-ds-warning">
                ⚠️ {discardWarning}
              </div>
            )}
            <div className="flex gap-2 justify-center flex-wrap" data-tutorial="y-action-buttons">
              <button
                type="button"
                onClick={handleDiscard}
                disabled={loading || !isHumanTurn || !isDiscard || selected.length === 0}
                title={discardWarning ?? undefined}
                className="px-4 py-2 rounded-lg bg-ds-success text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="discard-button"
              >
                {t('button.discard')}
              </button>
              <button
                type="button"
                onClick={handleYaniv}
                disabled={loading || !canYaniv}
                className={
                  canYaniv
                    ? 'px-4 py-2 rounded-lg bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm motion-safe:animate-pulse ring-2 ring-ds-success'
                    : 'px-4 py-2 rounded-lg bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm'
                }
                data-testid="yaniv-button"
              >
                {t('button.yaniv')}
              </button>
              <button
                type="button"
                onClick={handleDrawStock}
                disabled={loading || !isHumanTurn || !isDraw}
                className="px-4 py-2 rounded-lg bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="draw-stock-button"
              >
                {t('button.drawStock')}
              </button>
              {isRoundEnd && (
                <button
                  type="button"
                  onClick={handleNextRound}
                  disabled={loading}
                  className="px-4 py-2 rounded-lg bg-ds-accent text-ds-text-on-accent font-medium disabled:opacity-40 text-sm"
                  data-testid="next-round-button"
                >
                  {t('button.nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="y-reset-button"
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
