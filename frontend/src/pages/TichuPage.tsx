import { useCallback, useEffect, useMemo, useState } from 'react';
import { tichuApi } from '../api/gameApi';
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
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TichuResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { tichuBombIndices } from '../utils/tichuBomb';
import { classifyTichuCombo } from '../utils/tichuCombo';

type ApiArgs = {
  command: string;
  indices?: number[];
  declType?: number;
  config?: { cpuDifficulty?: number };
};

const TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tichu-hand"]', messageKey: 'tutorial.intro', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="tichu-declare"]', messageKey: 'tutorial.declare', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tichu-table"]', messageKey: 'tutorial.play', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tichu-cpus"]', messageKey: 'tutorial.teams', placement: 'bottom', advanceOn: 'next' },
];

// Values match the Go domain constants (TichuConfig.go): 0=Normal, 1=Easy, 2=Hard.
const DIFFICULTY_OPTIONS = [
  { value: '0', label: 'Normal' },
  { value: '1', label: 'Easy' },
  { value: '2', label: 'Hard' },
];

function parseTichuCommand(input: string): CliParseResult<[ApiArgs]> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase() ?? '';
  const args = parts.slice(1);
  switch (cmd) {
    case 'p':
    case 'play': {
      const indices = args.map(Number).filter((n) => !Number.isNaN(n));
      return { args: [{ command: 'p', indices }] };
    }
    case 'd':
    case 'declare': {
      const val = args[0] ? Number.parseInt(args[0], 10) : 0;
      return { args: [{ command: 'declare', declType: Number.isNaN(val) ? 0 : val }] };
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
      return { error: `Unknown command: ${cmd}. Try: p, d, r, sd, log` };
  }
}

function formatTichuState(state: TichuResponse): string {
  const lines: string[] = [`Phase: ${state.phase}`];
  for (const p of state.players) {
    const name = p.isHuman ? 'You' : `CPU ${p.id}`;
    lines.push(`${name} (Team ${p.team}): ${p.cardCount} cards`);
  }
  if (state.tableCards.length > 0) {
    lines.push(`Table: ${state.tableCombo} (${state.tableCards.length} cards)`);
  }
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}

const cliConfig: CliGameConfig<TichuResponse, [ApiArgs]> = {
  gameName: 'tichu',
  parseCommand: parseTichuCommand,
  formatResponse: formatTichuState,
  helpText: [
    'Commands:',
    '  p [idx...]   Play cards (empty = pass)',
    '  d [0-2]      Declare (0=none, 1=Tichu, 2=Grand)',
    '  r            Reset',
    '  sd [0-2]     Set difficulty',
    '  log          Action log',
  ],
};

/** Tichu game page content. */
function TichuPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tichu');
  const {
    state,
    loading,
    error,
    exec: apiCall,
    retry,
  } = useGameApi<TichuResponse, [ApiArgs]>((...args) => tichuApi.exec(...args));
  const { hint, hintEnabled, setHintEnabled } = useGameHint('tichu', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tichu');
  const { handleCommand } = useCliGame<TichuResponse, [ApiArgs]>(apiCall, cliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  const [selectedCards, setSelectedCards] = useState<Set<number>>(new Set());
  const [cpuDifficulty, setCpuDifficulty] = useState(0);

  // biome-ignore lint/correctness/useExhaustiveDependencies: mount-only reset
  useEffect(() => {
    void apiCall({ command: 'reset', config: { cpuDifficulty } });
  }, []);

  const phase = state?.phase ?? '';
  const isGameEnd = state?.gameEndFlag ?? false;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const isHumanTurn = state ? state.currentTurn === humanIdx : false;
  const humanPlayer = state?.players?.[humanIdx];
  const humanTeam = humanPlayer?.team ?? 0;
  const bombIndices = useMemo(() => tichuBombIndices(humanPlayer?.cards ?? []), [humanPlayer?.cards]);
  const humanWon = isGameEnd && !!state && state.scores[humanTeam] > state.scores[1 - humanTeam];

  // Additive combo-type preview for the current selection (warn-only; the backend
  // remains the source of truth and rejects any truly-illegal play — see #3392).
  const selectedCombo = useMemo(() => {
    const cards = humanPlayer?.cards ?? [];
    const picked = Array.from(selectedCards)
      .sort((a, b) => a - b)
      .map((i) => cards[i])
      .filter((c): c is NonNullable<typeof c> => c != null);
    if (picked.length === 0) return null;
    return classifyTichuCombo(picked);
  }, [humanPlayer?.cards, selectedCards]);

  useEffect(() => {
    if (humanWon) playSound('winFanfare');
  }, [humanWon, playSound]);

  const handleDeclare = useCallback(
    (value: number) => {
      void apiCall({ command: 'declare', declType: value });
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

  const handleReset = useCallback(
    () => apiCall({ command: 'reset', config: { cpuDifficulty } }),
    [apiCall, cpuDifficulty],
  );

  const handleDifficultyChange = useCallback(
    (value: number) => {
      setCpuDifficulty(value);
      void apiCall({ command: 'reset', config: { cpuDifficulty: value } });
    },
    [apiCall],
  );

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
    return t(`phase.${phase}`);
  }, [state, phase, isGameEnd, t]);

  const declLabel = (declType: number): string => {
    if (declType === 1) return t('label.tichu');
    if (declType === 2) return t('label.grandTichu');
    return '';
  };

  if (!state) return <GameSkeleton gameKey="tichu" layout={{ kind: 'card-grid', count: 14, cols: 'grid-cols-7' }} />;

  return (
    <GamePageShell
      title={tc('nav.tichu')}
      gameThemeBg={gameTheme.tichu.bg}
      phaseName={phaseName}
      gamePath="/tichu"
      isHumanTurn={isHumanTurn && !isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      onCelebrate={() => playSound('winFanfare')}
      headerExtra={
        <div className="flex items-center gap-2">
          <span className="text-xs opacity-75">
            {t('label.bombs')}: {state.bombCount}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </div>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <div className="flex flex-col gap-3 p-3 overflow-y-auto">
          <LandscapeBanner message={t('landscapeBanner')} />
          {error && <ErrorAlert message={error} onRetry={retry} />}
          <GameMessageBox messageCode={state.messageCode} messageParams={state.messageParams} message={state.message} />
          {hint && hintEnabled && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

          {/* Live cumulative team scores (visible during declare/play) */}
          {!isGameEnd && (
            <div className="flex justify-center gap-4 text-sm text-ds-text-primary" data-testid="tichu-score-bar">
              <span className={humanTeam === 0 ? 'font-bold text-ds-accent' : ''}>
                {t('label.teamA')}: {state.scores[0]}
              </span>
              <span className={humanTeam === 1 ? 'font-bold text-ds-accent' : ''}>
                {t('label.teamB')}: {state.scores[1]}
              </span>
              {state.isOneTwo && <span className="text-xs opacity-75">{t('label.oneTwo')}</span>}
            </div>
          )}

          {/* CPU areas */}
          <div className="flex justify-around gap-2" data-tutorial="tichu-cpus">
            {state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="text-center text-ds-text-primary text-sm">
                  <div className="font-bold">
                    {playerName(p.id, p.isHuman)} [{t('label.team')} {p.team}]
                  </div>
                  <div>
                    {p.cardCount} {t('label.cards')}
                    {p.declType > 0 ? ` · ${declLabel(p.declType)}` : ''}
                  </div>
                </div>
              ))}
          </div>

          {/* Table */}
          <div className="flex justify-center items-center min-h-[80px] gap-1" data-tutorial="tichu-table">
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

          {/* Declaration phase buttons */}
          {phase === 'declare' && isHumanTurn && (
            <div className="flex justify-center gap-2" data-tutorial="tichu-declare">
              <button type="button" className={btnSecondary} onClick={() => handleDeclare(0)}>
                {t('button.declareNone')}
              </button>
              <button type="button" className={btnWarning} onClick={() => handleDeclare(1)}>
                {t('button.tichu')}
              </button>
              <button type="button" className={btnWarning} onClick={() => handleDeclare(2)}>
                {t('button.grandTichu')}
              </button>
            </div>
          )}

          {/* Human hand */}
          {humanPlayer && (phase === 'play' || phase === 'declare') && (
            <div data-tutorial="tichu-hand">
              <div className="flex flex-wrap justify-center gap-1">
                {humanPlayer.cards.map((c, i) => {
                  const isBomb = bombIndices.has(i);
                  return (
                    <button
                      key={`hand-${c.design}-${c.value}-${i}`}
                      type="button"
                      className={`relative ${
                        selectedCards.has(i)
                          ? 'transition-transform -translate-y-2 ring-2 ring-ds-warning rounded'
                          : isBomb
                            ? 'transition-transform ring-2 ring-ds-error rounded'
                            : 'transition-transform'
                      }`}
                      onClick={() => toggleCard(i)}
                      disabled={phase !== 'play'}
                      aria-label={isBomb ? t('bombCardAriaLabel', { card: cardAlt(c) }) : cardAlt(c)}
                    >
                      <AnimatedCard card={c} width={cardWidth * 0.9} />
                      {isBomb && (
                        <span
                          className="absolute -top-1 -right-1 text-[10px] rounded-full bg-ds-error px-1 text-ds-text-on-accent"
                          aria-hidden="true"
                          data-testid={`tichu-bomb-badge-${i}`}
                        >
                          💣
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
              {phase === 'play' && isHumanTurn && selectedCombo && (
                <div className="mt-2 text-center text-xs" data-testid="tichu-combo-preview">
                  {selectedCombo.type === 'invalid' ? (
                    <span className="font-medium text-ds-warning" data-testid="tichu-combo-invalid">
                      {t('invalidCombo')}
                    </span>
                  ) : (
                    <span className="text-ds-text-secondary">
                      {t('comboPreview')}:{' '}
                      <span className="font-medium text-ds-accent">
                        {t(`combo.${selectedCombo.type}`)}
                        {selectedCombo.length > 0 ? ` (${selectedCombo.length})` : ''}
                      </span>
                    </span>
                  )}
                </div>
              )}
              {phase === 'play' && isHumanTurn && (
                <div className="flex justify-center gap-2 mt-2">
                  <button type="button" className={btnPrimary} onClick={handlePlay} disabled={selectedCards.size === 0}>
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
              <div className="text-lg font-bold">{humanWon ? t('result.youWin') : t('result.youLose')}</div>
              <div className="text-sm">
                {t('label.teamA')}: {state.scores[0]}
              </div>
              <div className="text-sm">
                {t('label.teamB')}: {state.scores[1]}
              </div>
              {state.isOneTwo && <div className="text-xs opacity-75">{t('label.oneTwo')}</div>}
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
                onSelect: (v: string) => handleDifficultyChange(Number.parseInt(v, 10)),
              },
              {
                type: 'checkbox' as const,
                id: 'frontendHint',
                label: tc('hint.toggle', { ns: 'tutorial' }),
                checked: hintEnabled,
                onToggle: setHintEnabled,
              },
            ],
          },
        ]}
      />
      <GameFooter className={gameTheme.tichu.footer}>
        <GameResetButton
          isGameEnd={isGameEnd}
          onReset={handleReset}
          requestConfirm={requestConfirm}
          loading={loading}
          dataTutorial="tichu-reset"
        />
      </GameFooter>
    </GamePageShell>
  );
}

/** Tichu page with tutorial wrapper. */
export const TichuPage = withTutorial(TichuPageContent, 'tichu', TUTORIAL_STEPS);
