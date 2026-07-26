import { useCallback, useMemo, useState } from 'react';
import { thirtyoneApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { badgeWarningColors } from '../styles/badgeStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ThirtyOneResponse } from '../types/card';
import { ThirtyOnePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { fiftyOneBestSuit, fiftyOneSuitScores } from '../utils/fiftyOneSuitScores';
import { hintCheckboxItem } from '../utils/settingsItems';

type ThirtyOneArgs = Parameters<typeof thirtyoneApi.exec>;

const DIFFICULTY_OPTIONS = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
];

const LIVES_OPTIONS = [
  { value: '2', label: '2' },
  { value: '3', label: '3' },
  { value: '4', label: '4' },
  { value: '5', label: '5' },
];

/** Tutorial steps for the Thirty-One game. */
const TO_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="to-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="to-discard-area"]',
    messageKey: 'tutorial.discardArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="to-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="to-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="to-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders a row of life icons for a player. */
function Lives({ lives, out }: { lives: number; out: boolean }) {
  if (out)
    return (
      <span role="img" aria-label="out">
        💀
      </span>
    );
  return (
    <span role="img" aria-label={`lives-${lives}`}>
      {'❤'.repeat(Math.max(0, lives)) || '·'}
    </span>
  );
}

/** Renders the Thirty-One (サーティワン) game page. */
export const ThirtyOnePage = withTutorial(ThirtyOnePageContent, 'thirtyone', TO_TUTORIAL_STEPS);

/** Inner content of the Thirty-One page. */
function ThirtyOnePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('thirtyone');
  const { state, loading, error, exec: execApi, retry } = useGameApi(thirtyoneApi.exec);
  const { cardWidth } = useCardDimensions();
  const [cpuDifficulty, setCpuDifficulty] = useState(1);
  const [initialLives, setInitialLives] = useState(3);
  const [selectedCardIdx, setSelectedCardIdx] = useState<number | null>(null);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('thirtyone', state);

  const handleReset = useCallback(() => {
    setSelectedCardIdx(null);
    return execApi('reset', undefined, { cpuDifficulty, initialLives });
  }, [execApi, cpuDifficulty, initialLives]);

  const handleDrawStock = useCallback(() => execApi('drawstock'), [execApi]);
  const handleDrawDiscard = useCallback(() => execApi('drawdiscard'), [execApi]);
  const handleKnock = useCallback(() => execApi('knock'), [execApi]);
  const handleNextRound = useCallback(() => execApi('nextround'), [execApi]);
  const handleDiscard = useCallback(() => {
    if (selectedCardIdx === null) return;
    execApi('discard', selectedCardIdx);
    setSelectedCardIdx(null);
  }, [execApi, selectedCardIdx]);

  useMountReset(execApi);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('thirtyone');
  const cliConfig: CliGameConfig<ThirtyOneResponse, ThirtyOneArgs> = useMemo(
    () => ({
      gameName: 'thirtyone',
      parseCommand: (input: string): CliParseResult<ThirtyOneArgs> => {
        const parts = input.trim().toLowerCase().split(/\s+/);
        const cmd = parts[0];
        if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
        if (cmd === 'ds' || cmd === 'drawstock') return { args: ['drawstock'] };
        if (cmd === 'dd' || cmd === 'drawdiscard') return { args: ['drawdiscard'] };
        if (cmd === 'd' || cmd === 'discard') {
          const idx = Number.parseInt(parts[1], 10);
          if (Number.isNaN(idx)) return { error: 'Usage: d <cardIdx>' };
          return { args: ['discard', idx] };
        }
        if (cmd === 'k' || cmd === 'knock') return { args: ['knock'] };
        if (cmd === 'nr' || cmd === 'nextround') return { args: ['nextround'] };
        if (cmd === 'log' || cmd === 'l') return { args: ['log'] };
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: (s: ThirtyOneResponse) => {
        const lines: string[] = [];
        lines.push(`Round ${s.roundNumber} | Stock ${s.drawPileCount} | Phase ${s.phase}`);
        if (s.discardTop) lines.push(`Discard top: ${s.discardTop.design} ${s.discardTop.value}`);
        for (const p of s.players) {
          const tag = p.isHuman ? 'You' : `CPU${p.id}`;
          const score = p.isHuman || s.gameEndFlag || s.phase === ThirtyOnePhase.ROUND_END ? p.score : '?';
          lines.push(
            `${tag}: lives ${p.lives}${p.isEliminated ? ' (OUT)' : ''}, best-suit ${score}, ${p.cardCount} cards`,
          );
        }
        if (s.knockerIdx >= 0) lines.push(`Knocked by player ${s.knockerIdx}`);
        if (s.message) lines.push(s.message);
        return lines.join('\n');
      },
      helpText: [
        'ds / drawstock   - Draw from stock',
        'dd / drawdiscard - Draw from discard',
        'd <idx>          - Discard a card',
        'k / knock        - Knock (stand pat)',
        'nr / nextround   - Next round',
        'r / reset        - Reset game',
        'l / log          - Show action log',
      ],
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const humanCards = state?.players[0]?.cards;
  const suitTotals = useMemo(() => fiftyOneSuitScores(humanCards ?? []), [humanCards]);
  const bestSuit = useMemo(() => fiftyOneBestSuit(suitTotals), [suitTotals]);

  // Keyboard shortcuts — must be registered before the early return so the hook
  // order stays stable while state loads. Phase/turn flags use optional chaining
  // for the same reason. Letter keys only (Enter/Space would double-fire on a
  // focused button); each binding's `enabled` mirrors its button's disabled
  // state, so keys are inert on invalid phases or when it is not the human turn.
  const kbIsDraw = state?.phase === ThirtyOnePhase.DRAW;
  const kbIsDiscard = state?.phase === ThirtyOnePhase.DISCARD;
  const kbIsGameEnd = !!state?.gameEndFlag || state?.phase === ThirtyOnePhase.GAME_END;
  const kbIsHumanTurn = state?.currentPlayerIdx === 0 && !kbIsGameEnd && (kbIsDraw || kbIsDiscard);
  const kbCanKnock = kbIsHumanTurn && kbIsDraw && (state?.knockerIdx ?? -1) < 0;
  const actionBindings = useMemo(
    () => [
      { key: 's', action: handleDrawStock, enabled: kbIsHumanTurn && kbIsDraw },
      { key: 'd', action: handleDrawDiscard, enabled: kbIsHumanTurn && kbIsDraw && !!state?.discardTop },
      { key: 'k', action: handleKnock, enabled: kbCanKnock },
      { key: 'x', action: handleDiscard, enabled: kbIsHumanTurn && kbIsDiscard && selectedCardIdx !== null },
    ],
    [
      handleDrawStock,
      handleDrawDiscard,
      handleKnock,
      handleDiscard,
      kbIsHumanTurn,
      kbIsDraw,
      kbIsDiscard,
      kbCanKnock,
      state?.discardTop,
      selectedCardIdx,
    ],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !cliEnabled && !loading });

  if (!state || state.players.length < 4)
    return <GameSkeleton gameKey="thirtyone" layout={{ kind: 'centered', rows: [3, 3] }} />;

  const isGameEnd = state.gameEndFlag || state.phase === ThirtyOnePhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const isDraw = state.phase === ThirtyOnePhase.DRAW;
  const isDiscard = state.phase === ThirtyOnePhase.DISCARD;
  const isRoundEnd = state.phase === ThirtyOnePhase.ROUND_END;
  const isHumanTurn = state.currentPlayerIdx === 0 && !isGameEnd && (isDraw || isDiscard);
  const human = state.players[0];
  const roundPlayerLabel = (idx: number) => {
    const p = state.players[idx];
    return p?.isHuman ? tc('player.you') : tc('player.cpu', { id: p?.id ?? idx });
  };
  const phaseName = isGameEnd
    ? t('phase.end')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isDiscard
        ? t('phase.discard')
        : t('phase.draw');
  const canKnock = isHumanTurn && isDraw && state.knockerIdx < 0;

  return (
    <GamePageShell
      title={tc('nav.thirtyone')}
      gameThemeBg={gameTheme.thirtyone.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/thirtyone"
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
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="to-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className={p.isEliminated ? 'text-center opacity-40' : 'text-center'}>
                    <div className="text-xs text-ds-text-muted mb-1">
                      {tc('player.cpu', { id: p.id })} — <Lives lives={p.lives} out={p.isEliminated} />
                      {state.knockerIdx === p.id && <span className="ml-1 text-ds-warning">{t('label.knocked')}</span>}
                    </div>
                    <div className="text-[10px] text-ds-text-muted mb-1">
                      {t('label.score')}: {isGameEnd || isRoundEnd ? p.score : '?'}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {isGameEnd || isRoundEnd
                        ? p.cards.map((c, i) => <AnimatedCard key={i} card={c} width={cardWidth * 0.6} />)
                        : Array.from({ length: p.cardCount }, (_, i) => (
                            <AnimatedCardBack key={i} width={cardWidth * 0.6} />
                          ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Stock + discard */}
            <div className="py-3 bg-black/20 rounded-lg flex justify-center gap-8" data-tutorial="to-discard-area">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted mb-1">
                  {t('label.stock')}: {state.drawPileCount}
                </div>
                {state.drawPileCount > 0 ? <AnimatedCardBack width={cardWidth * 0.8} /> : null}
              </div>
              <div className="text-center">
                <div className="text-xs text-ds-text-muted mb-1">{t('label.discardPile')}</div>
                {state.discardTop ? <AnimatedCard card={state.discardTop} width={cardWidth * 0.8} /> : '—'}
              </div>
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="to-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} — <Lives lives={human.lives} out={human.isEliminated} /> · {t('label.score')}:{' '}
                {human.score}
              </div>
              <ul
                className="flex justify-center gap-1.5 mb-1.5 text-xs flex-wrap list-none p-0 m-0"
                aria-label={t('label.suitScores')}
                data-testid="suit-score-badges"
              >
                {(['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const).map((d) => {
                  const isLeader = d === bestSuit && suitTotals[d] > 0;
                  const symbol = d === 'SPADE' ? '♠' : d === 'CLOVER' ? '♣' : d === 'HEART' ? '♥' : '♦';
                  const isRed = d === 'HEART' || d === 'DIAMOND';
                  const classes = isLeader
                    ? 'bg-ds-accent text-ds-text-on-accent border-ds-accent'
                    : 'bg-ds-surface text-ds-text border-ds-border';
                  return (
                    <li
                      key={d}
                      data-testid={`suit-badge-${d}`}
                      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full border font-medium ${classes}`}
                    >
                      <span className={isLeader ? '' : isRed ? 'text-ds-error' : ''}>{symbol}</span>
                      <span className="tabular-nums">{suitTotals[d]}</span>
                    </li>
                  );
                })}
              </ul>
              {isDiscard && isHumanTurn && (
                <div className="text-xs text-ds-text-muted mb-1">{t('label.selectCard')}</div>
              )}
              <div className="flex justify-center gap-2">
                {human.cards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    data-testid={`hand-card-${i}`}
                    onClick={() => isDiscard && isHumanTurn && setSelectedCardIdx(i === selectedCardIdx ? null : i)}
                    disabled={!isDiscard || !isHumanTurn}
                    className={
                      selectedCardIdx === i
                        ? isDiscard && isHumanTurn
                          ? 'rounded transition-all ring-2 ring-ds-info -translate-y-2 cursor-pointer hover:opacity-90'
                          : 'rounded transition-all ring-2 ring-ds-info -translate-y-2 cursor-default'
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

            {state.knockerIdx >= 0 && !isGameEnd && !isRoundEnd && (
              <div
                role="status"
                data-testid="knock-countdown-banner"
                className={`rounded-lg border border-ds-warning px-3 py-1.5 text-center text-sm font-medium ${badgeWarningColors}`}
              >
                {t(isHumanTurn ? 'label.knockBannerLastTurn' : 'label.knockBannerActive', {
                  knocker: state.knockerIdx === 0 ? tc('player.you') : tc('player.cpu', { id: state.knockerIdx }),
                })}
              </div>
            )}

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
                    id: 'initialLives',
                    label: t('settings.initialLives'),
                    value: String(initialLives),
                    options: LIVES_OPTIONS,
                    onSelect: (v: string) => setInitialLives(Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.thirtyone.footer} px-4 py-2.5`}>
            {isRoundEnd && (
              <div
                role="status"
                data-testid="thirtyone-round-summary"
                className="mb-2 p-2 rounded-lg bg-black/30 text-sm text-center space-y-1"
              >
                <div className="font-semibold text-ds-text-primary">{t('roundSummary.title')}</div>
                {state.thirtyOneIdx >= 0 && (
                  <div className="text-ds-success font-medium" data-testid="thirtyone-achiever">
                    {t('roundSummary.thirtyOne', { name: roundPlayerLabel(state.thirtyOneIdx) })}
                  </div>
                )}
                {state.roundLosers.length > 0 ? (
                  <ul className="text-ds-text-muted">
                    {state.roundLosers.map((idx) => {
                      const p = state.players[idx];
                      const out = !!p && (p.isEliminated || p.lives <= 0);
                      return (
                        <li
                          key={idx}
                          data-testid={`life-loss-${idx}`}
                          className={out ? 'text-ds-error font-medium' : ''}
                        >
                          {t('roundSummary.lifeLost', { name: roundPlayerLabel(idx) })}
                          <span className="ml-1 motion-safe:animate-pulse" aria-hidden="true">
                            💔
                          </span>
                          {out && (
                            <span className="ml-1" data-testid={`eliminated-${idx}`}>
                              💀 {t('roundSummary.eliminated')}
                            </span>
                          )}
                        </li>
                      );
                    })}
                  </ul>
                ) : (
                  <div data-testid="no-life-loss">{t('roundSummary.noLoss')}</div>
                )}
              </div>
            )}
            <div className="flex gap-2 justify-center flex-wrap" data-tutorial="to-action-buttons">
              <button
                type="button"
                onClick={handleDrawStock}
                disabled={loading || !isHumanTurn || !isDraw}
                className="px-4 py-2 rounded-lg bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="draw-stock-button"
                aria-keyshortcuts="s"
              >
                {t('button.drawStock')}
                <KbdBadge label={t('kbd.drawStock')} />
              </button>
              <button
                type="button"
                onClick={handleDrawDiscard}
                disabled={loading || !isHumanTurn || !isDraw || !state.discardTop}
                className="px-4 py-2 rounded-lg bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="draw-discard-button"
                aria-keyshortcuts="d"
              >
                {t('button.drawDiscard')}
                <KbdBadge label={t('kbd.drawDiscard')} />
              </button>
              <button
                type="button"
                onClick={handleKnock}
                disabled={loading || !canKnock}
                className="px-4 py-2 rounded-lg bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="knock-button"
                aria-keyshortcuts="k"
              >
                {t('button.knock')}
                <KbdBadge label={t('kbd.knock')} />
              </button>
              <button
                type="button"
                onClick={handleDiscard}
                disabled={loading || !isHumanTurn || !isDiscard || selectedCardIdx === null}
                className="px-4 py-2 rounded-lg bg-ds-success text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="discard-button"
                aria-keyshortcuts="x"
              >
                {t('button.discard')}
                <KbdBadge label={t('kbd.discard')} />
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
                dataTutorial="to-reset-button"
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
