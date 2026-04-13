import { useCallback, useEffect, useMemo, useState } from 'react';
import { fiftyoneApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { FiftyOneSkeleton } from '../components/skeleton/FiftyOneSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import type { FiftyOneResponse } from '../types/card';
import { FiftyOnePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

type FiftyOneArgs = Parameters<typeof fiftyoneApi.exec>;

const DIFFICULTY_OPTIONS = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
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
export function FiftyOnePage() {
  return (
    <TutorialWrapper gameName="fiftyone" steps={FO_TUTORIAL_STEPS}>
      <FiftyOnePageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Fifty-one page. */
function FiftyOnePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('fiftyone');
  const { state, loading, error, exec: execApi, retry } = useGameApi(fiftyoneApi.exec);
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

  const handleStop = useCallback(() => execApi('stop'), [execApi]);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

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
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state || state.players.length < 4) return <FiftyOneSkeleton />;

  const isGameEnd = state.gameEndFlag || state.phase === FiftyOnePhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const isHumanTurn = state.currentTurn === 0 && !isGameEnd;
  const human = state.players[0];
  const phaseName = isGameEnd ? t('phase.end') : t('phase.play');
  const canExchange = isHumanTurn && selectedHandIdx !== null && selectedTableIdx !== null;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green" aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.fiftyone')} />
      <PhaseIndicator phaseName={phaseName} isHumanTurn={isHumanTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/fiftyone" />
      </PhaseIndicator>

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
                    <div className="text-xs text-white/70 mb-1">
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
              <div className="text-center text-xs text-white/70 mb-2">{t('label.tableCards')}</div>
              <div className="flex justify-center gap-2">
                {state.tableCards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    onClick={() => isHumanTurn && setSelectedTableIdx(i === selectedTableIdx ? null : i)}
                    disabled={!isHumanTurn}
                    className={`rounded transition-all ${
                      selectedTableIdx === i ? 'ring-2 ring-yellow-400 -translate-y-1' : ''
                    } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                  >
                    <AnimatedCard card={c} width={cardWidth * 0.9} />
                  </button>
                ))}
              </div>
            </div>

            {/* Stop indicator */}
            {state.stopCallerIdx >= 0 && (
              <div className="text-center text-yellow-300 text-sm font-medium">{t('label.stopCalled')}</div>
            )}

            {/* Human hand */}
            <div className="text-center" data-tutorial="fo-player-hand">
              <div className="text-xs text-white/70 mb-1">
                {tc('player.you')} — {t('label.score')}: {human.score}
              </div>
              <div className="flex justify-center gap-2">
                {human.cards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    onClick={() => isHumanTurn && setSelectedHandIdx(i === selectedHandIdx ? null : i)}
                    disabled={!isHumanTurn}
                    className={`rounded transition-all ${
                      selectedHandIdx === i ? 'ring-2 ring-blue-400 -translate-y-2' : ''
                    } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
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
                    value: String(cpuDifficulty),
                    options: DIFFICULTY_OPTIONS,
                    onSelect: (v: string) => setCpuDifficulty(Number.parseInt(v, 10)),
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

          <GameFooter className="bg-game-bg-green-dark border-white/20 px-4 py-2.5">
            <div className="flex gap-2 justify-center flex-wrap" data-tutorial="fo-action-buttons">
              <button
                type="button"
                onClick={handleExchange}
                disabled={loading || !canExchange}
                className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="exchange-button"
              >
                {t('button.exchange')}
              </button>
              <button
                type="button"
                onClick={handleExchangeAll}
                disabled={loading || !isHumanTurn}
                className="px-4 py-2 rounded-lg bg-green-600 hover:bg-green-500 text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="exchange-all-button"
              >
                {t('button.exchangeAll')}
              </button>
              <button
                type="button"
                onClick={handleStop}
                disabled={loading || !isHumanTurn || state.stopCallerIdx >= 0}
                className="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-500 text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="stop-button"
              >
                {t('button.stop')}
              </button>
              <button
                type="button"
                onClick={() => requestConfirm(handleReset)}
                className="px-4 py-2 rounded-lg bg-gray-600 hover:bg-gray-500 text-white text-sm"
                data-tutorial="fo-reset-button"
              >
                {tc('button.reset')}
              </button>
              <ActionLogSection
                isEndPhase={isGameEnd}
                actionLog={actionLog}
                showActionLog={showActionLog}
                hideActionLog={hideActionLog}
              />
            </div>
          </GameFooter>

          <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
          <WinCelebration show={isGameEnd && humanWon} />
        </>
      )}
    </div>
  );
}
