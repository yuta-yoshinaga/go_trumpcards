import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { fiftyoneApi } from '../api/gameApi';
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
import { useSound } from '../providers/SoundProvider';
import { gameTheme } from '../styles/gameTheme';
import type { FiftyOneResponse } from '../types/card';
import { FiftyOnePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { fiftyOneBestSuit, fiftyOneSuitScores } from '../utils/fiftyOneSuitScores';
import { hintCheckboxItem } from '../utils/settingsItems';

type FiftyOneArgs = Parameters<typeof fiftyoneApi.exec>;

const DIFFICULTY_OPTIONS = [
  { value: '0', labelKey: 'settings.difficultyEasy' },
  { value: '1', labelKey: 'settings.difficultyNormal' },
  { value: '2', labelKey: 'settings.difficultyHard' },
];

/** Tutorial steps for the Fifty-one game. */
const FO_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="fo-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="fo-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fo-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fo-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fo-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Fifty-one (フィフティワン) game page. */
export const FiftyOnePage = withTutorial(FiftyOnePageContent, 'fiftyone', FO_TUTORIAL_STEPS);
/** Inner content of the Fifty-one page. */
function FiftyOnePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('fiftyone');
  const { state, loading, error, exec: execApi, retry } = useGameApi(fiftyoneApi.exec);
  const { playSound } = useSound();
  const { cardWidth } = useCardDimensions();
  const [cpuDifficulty, setCpuDifficulty] = useState(1);
  const [selectedHandIdx, setSelectedHandIdx] = useState<number | null>(null);
  const [selectedTableIdx, setSelectedTableIdx] = useState<number | null>(null);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('fiftyone', state);

  const handleReset = useCallback(() => execApi('reset', { config: { cpuDifficulty } }), [execApi, cpuDifficulty]);

  const handleExchange = useCallback(() => {
    if (selectedHandIdx !== null && selectedTableIdx !== null) {
      execApi('play', { handIdx: selectedHandIdx, tableIdx: selectedTableIdx });
      setSelectedHandIdx(null);
      setSelectedTableIdx(null);
    }
  }, [execApi, selectedHandIdx, selectedTableIdx]);

  const handleExchangeAll = useCallback(() => {
    execApi('exchangeall');
    setSelectedHandIdx(null);
    setSelectedTableIdx(null);
  }, [execApi]);

  const handleStop = useCallback(() => {
    playSound('chipClick');
    return execApi('stop');
  }, [execApi, playSound]);

  useMountReset(execApi);

  // This page shows errors with its own inline retry link rather than the
  // shared ErrorAlert, so the central errorBuzz tap never fires here — keep
  // a page-level buzz on the error appearance edge.
  const prevErrorRef = useRef<string | null>(null);
  useEffect(() => {
    if (error && !prevErrorRef.current) playSound('errorBuzz');
    prevErrorRef.current = error;
  }, [error, playSound]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('fiftyone');
  const cliConfig: CliGameConfig<FiftyOneResponse, FiftyOneArgs> = useMemo(
    () => ({
      gameName: 'fiftyone',
      parseCommand: (input: string): CliParseResult<FiftyOneArgs> => {
        const parts = input.trim().toLowerCase().split(/\s+/);
        const cmd = parts[0];
        if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
        if (cmd === 'p' || cmd === 'play') {
          const h = Number.parseInt(parts[1], 10);
          const ti = Number.parseInt(parts[2], 10);
          if (Number.isNaN(h) || Number.isNaN(ti)) return { error: 'Usage: p <handIdx> <tableIdx>' };
          return { args: ['play', { handIdx: h, tableIdx: ti }] };
        }
        if (cmd === 'a' || cmd === 'all') return { args: ['exchangeall'] };
        if (cmd === 'stop') return { args: ['stop'] };
        if (cmd === 'log' || cmd === 'l') return { args: ['log'] };
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: (s: FiftyOneResponse) => {
        const lines: string[] = [];
        lines.push(`Phase: ${s.gameEndFlag ? 'End' : 'Play'} | Turn: ${s.turnNumber}`);
        for (const p of s.players) {
          const tag = p.isHuman ? 'You' : `CPU${p.id}`;
          lines.push(`${tag}: ${p.cardCount} cards, score=${p.score}`);
        }
        if (s.stopCallerIdx >= 0) lines.push(`Stop called by player ${s.stopCallerIdx}`);
        if (s.message) lines.push(s.message);
        return lines.join('\n');
      },
      helpText: [
        'p <handIdx> <tableIdx> - Exchange 1 card',
        'a/all                 - Exchange all 5',
        'stop                  - Call stop',
        'r/reset               - Reset game',
        'l/log                 - Show action log',
      ],
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const humanCards = state?.players[0]?.cards;
  const suitTotals = useMemo(() => fiftyOneSuitScores(humanCards ?? []), [humanCards]);
  const bestSuit = useMemo(() => fiftyOneBestSuit(suitTotals), [suitTotals]);

  if (!state || state.players.length < 4)
    return <GameSkeleton gameKey="fiftyone" layout={{ kind: 'centered', rows: [5, 5] }} />;

  const isGameEnd = state.gameEndFlag || state.phase === FiftyOnePhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const isHumanTurn = state.currentTurn === 0 && !isGameEnd;
  const human = state.players[0];
  const phaseName = isGameEnd ? t('phase.end') : t('phase.play');
  const canExchange = isHumanTurn && selectedHandIdx !== null && selectedTableIdx !== null;
  const exchangeGuide = !isHumanTurn
    ? ''
    : selectedHandIdx === null
      ? t('guide.selectHand')
      : selectedTableIdx === null
        ? t('guide.selectTable')
        : '';

  return (
    <GamePageShell
      title={tc('nav.fiftyone')}
      gameThemeBg={gameTheme.fiftyone.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/fiftyone"
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
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="fo-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1">
                      {tc('player.cpu', { id: p.id })} — {t('label.score')}: {isGameEnd ? p.score : '?'}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {isGameEnd
                        ? p.cards.map((c, i) => <AnimatedCard key={i} card={c} width={cardWidth * 0.65} />)
                        : Array.from({ length: p.cardCount }, (_, i) => (
                            <AnimatedCardBack key={i} width={cardWidth * 0.65} />
                          ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Table cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="fo-table-cards">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('label.tableCards')}</div>
              <div className="flex justify-center gap-2">
                {state.tableCards.map((c, i) => {
                  const isSelected = selectedTableIdx === i;
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => isHumanTurn && setSelectedTableIdx(isSelected ? null : i)}
                      disabled={!isHumanTurn}
                      aria-label={cardAlt(c)}
                      aria-pressed={isHumanTurn ? isSelected : undefined}
                      className={`rounded transition-all ${
                        isSelected ? 'ring-2 ring-ds-warning -translate-y-1' : ''
                      } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    >
                      <AnimatedCard card={c} width={cardWidth * 0.9} />
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Stop indicator */}
            {/* CUI は宣言者と「残り1巡が最終ラウンド」まで出しているのに、Web は
                「ストップ宣言済み」の一文だけで、誰が宣言したのかも、あと何巡で
                終わるのかも分からなかった (#5532)。 */}
            {state.stopCallerIdx >= 0 && (
              <div className="text-center text-ds-warning text-sm font-medium" data-testid="fo-stop-called">
                {t('label.stopCalledBy', {
                  name: state.players[state.stopCallerIdx]?.isHuman
                    ? tc('player.you')
                    : tc('player.cpu', { id: state.stopCallerIdx }),
                })}
              </div>
            )}

            {/* Human hand */}
            <div className="text-center" data-tutorial="fo-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} — {t('label.score')}: {human.score}
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
              <div className="flex justify-center gap-2">
                {human.cards.map((c, i) => {
                  const isSelected = selectedHandIdx === i;
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => isHumanTurn && setSelectedHandIdx(isSelected ? null : i)}
                      disabled={!isHumanTurn}
                      aria-label={cardAlt(c)}
                      aria-pressed={isHumanTurn ? isSelected : undefined}
                      className={`rounded transition-all ${
                        isSelected ? 'ring-2 ring-ds-info -translate-y-2' : ''
                      } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    >
                      <AnimatedCard card={c} width={cardWidth} />
                    </button>
                  );
                })}
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
                    options: DIFFICULTY_OPTIONS.map((opt) => ({ value: opt.value, label: t(opt.labelKey) })),
                    onSelect: (v: string) => setCpuDifficulty(Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.fiftyone.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap" data-tutorial="fo-action-buttons">
              <button
                type="button"
                onClick={handleExchange}
                disabled={loading || !canExchange}
                title={exchangeGuide || undefined}
                className="px-4 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="exchange-button"
              >
                {t('button.exchange')}
              </button>
              <button
                type="button"
                onClick={handleExchangeAll}
                disabled={loading || !isHumanTurn}
                className="px-4 py-2 rounded-lg bg-ds-success hover:bg-ds-success text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="exchange-all-button"
              >
                {t('button.exchangeAll')}
              </button>
              <button
                type="button"
                onClick={handleStop}
                disabled={loading || !isHumanTurn || state.stopCallerIdx >= 0}
                className="px-4 py-2 rounded-lg bg-ds-error hover:bg-ds-error text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="stop-button"
              >
                {t('button.stop')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="fo-reset-button"
              />
              <ActionLogSection
                isEndPhase={isGameEnd}
                actionLog={actionLog}
                showActionLog={showActionLog}
                hideActionLog={hideActionLog}
              />
            </div>
            {/* Always rendered so the footer height stays stable while selections toggle. */}
            <p className="text-xs text-ds-text-muted text-center mt-1 mb-0 min-h-4" role="note">
              {exchangeGuide}
            </p>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
