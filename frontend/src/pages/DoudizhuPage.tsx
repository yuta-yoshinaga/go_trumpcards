import { useCallback, useEffect, useMemo, useState } from 'react';
import { doudizhuApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DoudizhuResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { classifyDoudizhuCombo, doudizhuInvalidReason } from '../utils/doudizhuComboValidator';

type ApiArgs = {
  command: string;
  indices?: number[];
  bidValue?: number;
  config?: { cpuDifficulty?: number };
};

const TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ddz-hand"]', messageKey: 'tutorial.intro', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ddz-bid"]', messageKey: 'tutorial.bidPhase', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ddz-table"]', messageKey: 'tutorial.playPhase', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ddz-kitty"]', messageKey: 'tutorial.combos', placement: 'bottom', advanceOn: 'next' },
];

function parseDDZCommand(input: string): CliParseResult<[ApiArgs]> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase() ?? '';
  const args = parts.slice(1);
  switch (cmd) {
    case 'p':
    case 'play': {
      const indices = args.map(Number).filter((n) => !Number.isNaN(n));
      return { args: [{ command: 'p', indices }] };
    }
    case 'b':
    case 'bid': {
      const val = args[0] ? Number.parseInt(args[0], 10) : 0;
      return { args: [{ command: 'bid', bidValue: Number.isNaN(val) ? 0 : val }] };
    }
    case 'r':
    case 'reset':
      return { args: [{ command: 'reset' }] };
    case 'sd': {
      const v = args[0] ? Number.parseInt(args[0], 10) : 0;
      return { args: [{ command: 'reset', config: { cpuDifficulty: Number.isNaN(v) ? 0 : v } }] };
    }
    case 'log':
    case 'l':
      return { args: [{ command: 'l' }] };
    default:
      return { error: `Unknown command: ${cmd}. Try: p, b, r, sd, log` };
  }
}

/** Formats Dou Dizhu state for the CLI terminal, including the human's indexed hand. */
export function formatDDZState(state: DoudizhuResponse): string {
  const lines: string[] = [`Phase: ${state.phase}`];
  for (const p of state.players) {
    const role = p.isLandlord ? ' [Landlord]' : '';
    const name = p.isHuman ? 'You' : `CPU ${p.id}`;
    lines.push(`${name}${role}: ${p.cardCount} cards`);
  }
  if (state.tableCards.length > 0) {
    lines.push(`Table: ${state.tableCombo} (${state.tableCards.length} cards)`);
  }
  const human = state.players.find((p) => p.isHuman);
  if (human?.cards?.length) {
    lines.push(`Your hand: ${human.cards.map((c, i) => `[${i}]${cardAlt(c)}`).join(' ')}`);
  }
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}

const cliConfig: CliGameConfig<DoudizhuResponse, [ApiArgs]> = {
  gameName: 'doudizhu',
  parseCommand: parseDDZCommand,
  formatResponse: formatDDZState,
  helpText: [
    'Commands:',
    '  p [idx...]   Play cards (empty = pass)',
    '  b [0-3]      Bid (0=pass)',
    '  r            Reset',
    '  sd [0-2]     Set difficulty',
    '  log          Action log',
  ],
};

/** Dou Dizhu game page content. */
function DoudizhuPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('doudizhu');
  const {
    state,
    loading,
    error,
    exec: apiCall,
    retry,
  } = useGameApi<DoudizhuResponse, [ApiArgs]>((...args) => doudizhuApi.exec(...args));
  const { hint, hintEnabled, setHintEnabled } = useGameHint('doudizhu', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('doudizhu');
  // The module-level config cannot see `hint`: hints are computed per render.
  // Layer the local answer on here rather than restructuring the shared const.
  const cliConfigWithHint = useMemo(
    () => ({
      ...cliConfig,
      localCommand: hintLocalCommand(hint),
    }),
    [hint],
  );
  const { handleCommand } = useCliGame<DoudizhuResponse, [ApiArgs]>(apiCall, cliConfigWithHint, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });
  const { cardWidth } = useCardDimensions();

  const [selectedCards, setSelectedCards] = useState<Set<number>>(new Set());

  // biome-ignore lint/correctness/useExhaustiveDependencies: mount-only reset
  useEffect(() => {
    void apiCall({ command: 'reset' });
  }, []);

  const phase = state?.phase ?? '';
  const isGameEnd = state?.gameEndFlag ?? false;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const isHumanTurn = state ? state.currentTurn === humanIdx : false;
  const humanPlayer = state?.players?.[humanIdx];

  const landlordWon = isGameEnd && !!state && state.scores[state.landlordIdx] > 0;
  const humanWon = isGameEnd && !!state && state.scores[humanIdx] > 0;

  const handleBid = useCallback(
    (value: number) => {
      void apiCall({ command: 'bid', bidValue: value });
    },
    [apiCall],
  );

  const handlePlay = useCallback(() => {
    const indices = Array.from(selectedCards).sort((a, b) => a - b);
    void apiCall({ command: 'p', indices });
    setSelectedCards(new Set());
  }, [apiCall, selectedCards]);

  const handlePass = useCallback(() => {
    void apiCall({ command: 'p', indices: [] });
    setSelectedCards(new Set());
  }, [apiCall]);

  const handleReset = useCallback(() => apiCall({ command: 'reset' }), [apiCall]);

  const toggleCard = useCallback((idx: number) => {
    setSelectedCards((prev) => {
      const next = new Set(prev);
      if (next.has(idx)) next.delete(idx);
      else next.add(idx);
      return next;
    });
  }, []);

  const currentTurn = state?.currentTurn;
  // biome-ignore lint/correctness/useExhaustiveDependencies: clear selection on turn/phase change
  useEffect(() => {
    setSelectedCards(new Set());
  }, [currentTurn, phase]);

  const phaseName = useMemo(() => {
    if (!state) return '';
    if (isGameEnd) return t('phase.end');
    return t(`phase.${phase}`, { defaultValue: phase });
  }, [state, phase, isGameEnd, t]);

  // Classify the current selection and pre-validate it against the table so the
  // player sees why a play is illegal before submitting (backend still validates).
  const selectionHint = useMemo(() => {
    const cards = humanPlayer?.cards ?? [];
    const selected = Array.from(selectedCards)
      .sort((a, b) => a - b)
      .map((i) => cards[i])
      .filter((card): card is NonNullable<typeof card> => card != null);
    if (phase !== 'play' || !isHumanTurn || selected.length === 0) return null;
    const combo = classifyDoudizhuCombo(selected);
    const reason = doudizhuInvalidReason(selected, state?.tableCards ?? []);
    return { combo, reason, count: selected.length };
  }, [humanPlayer, selectedCards, phase, isHumanTurn, state?.tableCards]);

  if (!state) return <GameSkeleton gameKey="doudizhu" layout={{ kind: 'card-grid', count: 17, cols: 'grid-cols-6' }} />;

  return (
    <GamePageShell
      title={tc('nav.doudizhu')}
      gameThemeBg={gameTheme.doudizhu.bg}
      phaseName={phaseName}
      gamePath="/doudizhu"
      isHumanTurn={isHumanTurn && !isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      headerExtra={
        <div className="flex items-center gap-2">
          {state.landlordIdx >= 0 && (
            <span className="text-xs opacity-75">
              {t('label.bid')}: {state.baseBid} | {t('label.bombs')}: {state.bombCount}
            </span>
          )}
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </div>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <div className="flex flex-col gap-3 p-3 overflow-y-auto">
          <LandscapeBanner message={t('landscapeBanner', { defaultValue: '' })} />
          {error && <ErrorAlert message={error} onRetry={retry} />}
          <GameMessageBox messageCode={state.messageCode} messageParams={state.messageParams} message={state.message} />
          {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
              付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
              読み上げられないことがある (#5955, #6663)。 */}
          <div data-testid="doudizhu-hint-live" role="status" aria-live="polite">
            {hint && hintEnabled && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
          </div>

          {/* CPU areas */}
          <div className="flex justify-around gap-2" data-tutorial="ddz-cpus">
            {state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="text-center text-ds-text-primary text-sm">
                  <div className="font-bold">
                    CPU {p.id} {p.isLandlord ? `[${t('label.landlord')}]` : `[${t('label.peasant')}]`}
                  </div>
                  <div>{p.cardCount} cards</div>
                </div>
              ))}
          </div>

          {/* Kitty cards */}
          {state.kittyCards.length > 0 && (
            <div className="flex justify-center gap-1 items-center" data-tutorial="ddz-kitty">
              <span className="text-ds-text-primary text-xs mr-1">{t('label.kitty')}:</span>
              {state.kittyCards.map((c) => (
                <AnimatedCard key={`kitty-${c.design}-${c.value}`} card={c} width={cardWidth * 0.6} />
              ))}
            </div>
          )}

          {/* Table */}
          <div className="flex justify-center items-center min-h-[80px] gap-1" data-tutorial="ddz-table">
            {state.tableCards.length > 0 ? (
              <>
                {state.tableCards.map((c) => (
                  <AnimatedCard key={`table-${c.design}-${c.value}`} card={c} width={cardWidth * 0.8} />
                ))}
                <span className="text-ds-text-primary text-xs ml-2">{state.tableCombo}</span>
              </>
            ) : (
              <span className="text-ds-text-secondary text-sm">{t('label.table')}: ---</span>
            )}
          </div>

          {/* Bid phase buttons */}
          {phase === 'bid' && isHumanTurn && (
            <div className="flex justify-center gap-2" data-tutorial="ddz-bid">
              {[1, 2, 3]
                .filter((v) => v > state.highestBid)
                .map((v) => (
                  <button key={v} type="button" className={btnWarning} onClick={() => handleBid(v)}>
                    {t(`button.bid${v}`)}
                  </button>
                ))}
              <button type="button" className={btnSecondary} onClick={() => handleBid(0)}>
                {t('button.pass')}
              </button>
            </div>
          )}

          {/* Human hand — shown during both bid and play phases (display-only while bidding) */}
          {humanPlayer && (phase === 'play' || phase === 'bid') && (
            <div data-tutorial="ddz-hand">
              <div className="flex flex-wrap justify-center gap-1">
                {humanPlayer.cards.map((c, i) => {
                  const interactive = phase === 'play';
                  const selected = interactive && selectedCards.has(i);
                  return (
                    <button
                      key={`hand-${c.design}-${c.value}`}
                      type="button"
                      disabled={!interactive}
                      className={
                        selected
                          ? 'transition-transform -translate-y-2 ring-2 ring-ds-warning rounded'
                          : 'transition-transform'
                      }
                      onClick={interactive ? () => toggleCard(i) : undefined}
                      aria-label={cardAlt(c)}
                      aria-pressed={interactive ? selectedCards.has(i) : undefined}
                    >
                      <AnimatedCard card={c} width={cardWidth * 0.9} />
                    </button>
                  );
                })}
              </div>
              {selectionHint && (
                <div className="mt-2 text-center text-xs" data-testid="ddz-combo-hint">
                  {selectionHint.reason === 'notCombo' ? (
                    <p role="status" data-testid="ddz-invalid-combo" className="font-medium text-ds-warning">
                      {t('combo.notCombo')}
                    </p>
                  ) : selectionHint.reason === 'noBeat' ? (
                    <p role="status" data-testid="ddz-no-beat" className="font-medium text-ds-warning">
                      {t('combo.noBeat')}
                    </p>
                  ) : (
                    selectionHint.combo && (
                      <p role="status" data-testid="ddz-combo-type" className="font-semibold text-ds-info">
                        {`${t('combo.selectedLabel')}: ${t('combo.badge', {
                          type: t(`combo.type.${selectionHint.combo.type}`),
                          count: selectionHint.count,
                        })}`}
                      </p>
                    )
                  )}
                </div>
              )}
              {phase === 'play' && isHumanTurn && (
                <div className="flex justify-center gap-2 mt-2">
                  {/* selectionHint has already decided this selection is illegal, so
                      sending it only to be rejected wastes a round trip (#4754). */}
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={selectedCards.size === 0 || selectionHint?.reason != null}
                  >
                    {t('button.play')}
                  </button>
                  {state.tableCards.length > 0 && (
                    <button type="button" className={btnSecondary} onClick={handlePass}>
                      {t('button.pass')}
                    </button>
                  )}
                </div>
              )}
            </div>
          )}

          {/* End phase scores */}
          {isGameEnd && (
            <div className="text-center text-ds-text-primary space-y-1">
              <div className="text-lg font-bold">
                {landlordWon ? t('result.landlordWins') : t('result.peasantsWin')}
              </div>
              {state.players.map((p, i) => (
                <div key={p.id} className="text-sm">
                  {p.isHuman ? tc('you', { defaultValue: 'You' }) : `CPU ${p.id}`}
                  {p.isLandlord ? ` [${t('label.landlord')}]` : ''}: {state.scores[i]}
                </div>
              ))}
            </div>
          )}

          <ActionLogSection
            isEndPhase={isGameEnd}
            actionLog={actionLog}
            showActionLog={showActionLog}
            hideActionLog={hideActionLog}
          />
        </div>
      )}
      <GameFooter className={gameTheme.doudizhu.footer}>
        <GameResetButton
          isGameEnd={isGameEnd}
          onReset={handleReset}
          requestConfirm={requestConfirm}
          loading={loading}
          dataTutorial="ddz-reset"
        />
        <label className="flex items-center gap-1 text-ds-text-primary text-xs min-h-[44px]">
          <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
          {tc('hint')}
        </label>
      </GameFooter>
    </GamePageShell>
  );
}

/** Dou Dizhu page with tutorial wrapper. */
export const DoudizhuPage = withTutorial(DoudizhuPageContent, 'doudizhu', TUTORIAL_STEPS);
