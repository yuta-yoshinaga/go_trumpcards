import { useCallback, useEffect, useMemo, useState } from 'react';
import { type ScorpionMoveZone, scorpionApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { DropZone } from '../components/DropZone';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { KlondikeSkeleton } from '../components/skeleton/KlondikeSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ScorpionResponse } from '../types/card';
import { ScorpionPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import type { CliGameConfig } from '../utils/cli/types';

const noop = () => {};

/** Scorpion tutorial step definitions. */
const SC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sc-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sc-stock"]',
    messageKey: 'tutorial.stock',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sc-controls"]',
    messageKey: 'tutorial.controls',
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

/** CLI help text for Scorpion. */
const SCORPION_HELP = [
  'm <from> <to>  Move top card between tableau columns',
  'm <from> <idx> <to>  Move a card (and everything on top) by index',
  'd              Deal 3 stock cards to columns 0-2',
  'g              Give up',
  'h              Hint',
  'ac             Auto-complete',
  'u              Undo',
  'r              Reset',
];

/** Parse a Scorpion CLI command into API call arguments. */
function parseScorpionCommand(input: string): { args: Parameters<typeof scorpionApi.exec> } | { error: string } {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'd':
    case 'deal':
      return { args: ['deal'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'm':
    case 'move': {
      if (parts.length === 3) {
        const from = Number.parseInt(parts[1], 10);
        const to = Number.parseInt(parts[2], 10);
        if (Number.isNaN(from) || Number.isNaN(to)) return { error: 'Invalid column' };
        return {
          args: ['move', { zone: 'tableau', col: from, cardIndex: -1 }, { zone: 'tableau', col: to }],
        };
      }
      if (parts.length === 4) {
        const from = Number.parseInt(parts[1], 10);
        const idx = Number.parseInt(parts[2], 10);
        const to = Number.parseInt(parts[3], 10);
        if (Number.isNaN(from) || Number.isNaN(idx) || Number.isNaN(to)) return { error: 'Invalid arg' };
        return {
          args: ['move', { zone: 'tableau', col: from, cardIndex: idx }, { zone: 'tableau', col: to }],
        };
      }
      return { error: 'Usage: m <fromCol> [<cardIdx>] <toCol>' };
    }
    default:
      return { error: `Unknown command: ${cmd}` };
  }
}

/** Format Scorpion state for CLI display. */
function formatScorpionState(state: ScorpionResponse): string {
  const lines: string[] = [];
  lines.push(`Completed: ${state.completedSuits}/4  Stock: ${state.stockCount}`);
  lines.push('');
  lines.push('Tableau:');
  for (let col = 0; col < state.tableau.length; col++) {
    const cards = state.tableau[col]
      .map((tc, i) => (tc.faceUp && tc.card ? `[${i}]${tc.card.design}-${tc.card.value}` : `[${i}]??`))
      .join(' ');
    lines.push(`  ${col}: ${cards || '(empty)'}`);
  }
  lines.push('');
  lines.push(`Moves: ${state.moveCount}  Phase: ${state.phase}`);
  return lines.join('\n');
}

/** Renders the Scorpion solitaire game page. */
export function ScorpionPage() {
  return (
    <TutorialWrapper gameName="scorpion" steps={SC_TUTORIAL_STEPS}>
      <ScorpionPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Scorpion page. */
function ScorpionPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('scorpion');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec: apiCall,
    retry,
  } = useGameApi<ScorpionResponse, Parameters<typeof scorpionApi.exec>>((...args) => scorpionApi.exec(...args));

  useEffect(() => {
    void apiCall('reset');
  }, [apiCall]);

  const [selectedSource, setSelectedSource] = useState<ScorpionMoveZone | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('scorpion', state);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('scorpion');
  const scorpionCliConfig: CliGameConfig<ScorpionResponse, Parameters<typeof scorpionApi.exec>> = useMemo(
    () => ({
      gameName: 'scorpion',
      parseCommand: parseScorpionCommand,
      formatResponse: formatScorpionState,
      helpText: SCORPION_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, scorpionCliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const sc = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    const cols = 7;
    const colW = Math.floor((windowWidth - padX - (cols - 1) * gapPx) / cols);
    const cw = Math.min(Math.max(colW, 28), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.48);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const dispatchMove = useCallback(
    (source: ScorpionMoveZone, target: ScorpionMoveZone) => {
      void apiCall('move', source, target);
    },
    [apiCall],
  );
  const dnd = useSolitaireDragDrop<ScorpionMoveZone>({
    onMove: dispatchMove,
    isPlaying: state?.phase === ScorpionPhase.PLAYING,
    disabled: loading,
  });

  const handleReset = useCallback(() => {
    requestConfirm(() => {
      void apiCall('reset');
      playSound('shuffle');
    });
  }, [apiCall, requestConfirm, playSound]);

  const handleDeal = useCallback(() => {
    void apiCall('deal');
    playSound('cardPlace');
  }, [apiCall, playSound]);

  const handleGiveUp = useCallback(() => {
    void apiCall('giveup');
  }, [apiCall]);

  const handleHint = useCallback(() => {
    void apiCall('hint');
  }, [apiCall]);

  const handleAutoComplete = useCallback(() => {
    void apiCall('autocomplete');
  }, [apiCall]);

  const handleUndo = useCallback(() => {
    void apiCall('undo');
  }, [apiCall]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      void apiCall('undo_n', undefined, undefined, n);
    },
    [apiCall],
  );

  const handleSelectSource = useCallback(
    (zone: string, col: number, cardIndex: number) => {
      if (
        selectedSource &&
        selectedSource.zone === zone &&
        selectedSource.col === col &&
        selectedSource.cardIndex === cardIndex
      ) {
        setSelectedSource(null);
        return;
      }
      setSelectedSource({ zone, col, cardIndex });
    },
    [selectedSource],
  );

  const handleSelectTarget = useCallback(
    (zone: string, col: number) => {
      if (!selectedSource) return;
      void apiCall('move', selectedSource, { zone, col });
      setSelectedSource(null);
      playSound('cardPlace');
    },
    [apiCall, selectedSource, playSound],
  );

  const isPlayingForKbd = state?.phase === ScorpionPhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: handleGiveUp },
      { key: 'z', action: handleUndo },
      { key: 'd', action: handleDeal },
    ],
    [handleHint, handleAutoComplete, handleGiveUp, handleUndo, handleDeal],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (error) return <ErrorAlert message={error} onRetry={retry} />;
  if (!state) return <KlondikeSkeleton />;

  const isPlaying = state.phase === ScorpionPhase.PLAYING;
  const isGameClear = state.phase === ScorpionPhase.GAME_CLEAR;
  const isGameOver = state.phase === ScorpionPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.klondike.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.scorpion')} />
      <PhaseIndicator
        phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      >
        <span>
          {t('moveCount')}: {state.moveCount}
        </span>
        <span data-tutorial="sc-stock">
          {t('stock')}: {state.stockCount} / {t('completed')}: {state.completedSuits}/4
        </span>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/scorpion" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={tc('settings.title', { ns: 'common' })}
            groups={[
              {
                items: [
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
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="flex gap-1 sm:gap-2 justify-center" data-tutorial="sc-tableau">
              {state.tableau.map((col, colIdx) => (
                <div key={colIdx} className="flex flex-col items-center" style={{ width: sc.cw }}>
                  <div className="text-game-text-muted text-xs mb-1">{colIdx}</div>
                  {col.length === 0 ? (
                    <DropZone
                      onDrop={dnd.handleDrop({ zone: 'tableau', col: colIdx })}
                      onDragOver={dnd.handleDragOver({ zone: 'tableau', col: colIdx })}
                      onDragLeave={dnd.handleDragLeave}
                      isDropTarget={dnd.isDropTarget({ zone: 'tableau', col: colIdx })}
                    >
                      <button
                        type="button"
                        className={`border-2 border-dashed border-game-border rounded-lg flex items-center justify-center text-game-text-muted ${focusRingWhite} ${
                          selectedSource ? 'hover:ring-2 hover:ring-ds-warning cursor-pointer' : ''
                        }`}
                        style={{ width: sc.cw, height: sc.ch }}
                        onClick={() => selectedSource && handleSelectTarget('tableau', colIdx)}
                        disabled={!isPlaying || !selectedSource}
                        aria-label={`${t('empty')} ${t('tableau')} ${colIdx}`}
                      >
                        {t('empty')}
                      </button>
                    </DropZone>
                  ) : (
                    <div className="relative" style={{ width: sc.cw, height: sc.ch + (col.length - 1) * sc.co }}>
                      {col.map((tc, cardIdx) => {
                        const isSelected = isSourceSelected('tableau', colIdx, cardIdx);
                        const zone: ScorpionMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
                        const isDragSrc = dnd.isDragSource(zone);
                        const isLast = cardIdx === col.length - 1;

                        const hintFrom =
                          state.hint && state.hint.fromCol === colIdx && state.hint.cardIndex === cardIdx;
                        const hintTo = state.hint && state.hint.toCol === colIdx && isLast;

                        return (
                          <div key={cardIdx} className="absolute" style={{ top: cardIdx * sc.co, zIndex: cardIdx }}>
                            <DropZone
                              onDrop={isLast ? dnd.handleDrop(zone) : noop}
                              onDragOver={isLast ? dnd.handleDragOver(zone) : noop}
                              onDragLeave={isLast ? dnd.handleDragLeave : undefined}
                              isDropTarget={isLast && dnd.isDropTarget({ zone: 'tableau', col: colIdx })}
                            >
                              {tc.faceUp ? (
                                <button
                                  type="button"
                                  draggable={isPlaying}
                                  onDragStart={dnd.handleDragStart(zone)}
                                  onDragEnd={dnd.handleDragEnd}
                                  className={`${focusRingWhite} rounded-lg transition-all ${
                                    isSelected ? 'ring-2 ring-ds-warning -translate-y-1' : ''
                                  } ${isDragSrc ? 'opacity-50' : ''} ${
                                    hintFrom ? 'ring-2 ring-ds-info animate-pulse' : ''
                                  } ${hintTo ? 'ring-2 ring-ds-success animate-pulse' : ''}`}
                                  onClick={() => {
                                    if (selectedSource) {
                                      if (isLast) {
                                        handleSelectTarget('tableau', colIdx);
                                      } else {
                                        handleSelectSource('tableau', colIdx, cardIdx);
                                      }
                                    } else {
                                      handleSelectSource('tableau', colIdx, cardIdx);
                                    }
                                  }}
                                  disabled={!isPlaying}
                                  aria-label={tc.card ? cardAlt(tc.card) : ''}
                                >
                                  {tc.card && <AnimatedCard card={tc.card} width={sc.cw} />}
                                </button>
                              ) : (
                                <AnimatedCardBack width={sc.cw} />
                              )}
                            </DropZone>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          <div data-tutorial="sc-controls">
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {error && <ErrorAlert message={error} onRetry={retry} />}

            {state.hint && (
              <div
                className="text-sm text-ds-accent bg-ds-surface/90 border border-ds-accent rounded px-3 py-1.5 mt-1"
                role="status"
                aria-live="polite"
              >
                {state.hint.fromCol < 0 ? t('deal') : `${t('tableau')} ${state.hint.toCol}`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />

            <GameFooter>
              <button type="button" className={btnPrimary} onClick={handleReset} data-tutorial="sc-reset-button">
                {t('common:reset')}
              </button>

              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleDeal}
                    disabled={loading || state.stockCount === 0}
                  >
                    {t('deal')}
                  </button>
                  <button type="button" className={btnOutline} onClick={handleHint} disabled={loading}>
                    {t('hint')}
                  </button>
                  <button type="button" className={btnSuccess} onClick={handleAutoComplete} disabled={loading}>
                    {t('autoComplete')}
                  </button>
                  <button
                    type="button"
                    className={btnOutline}
                    onClick={handleUndo}
                    disabled={loading || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                    {t('giveup')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? -1}
                      onEscape={handleUndoEscape}
                      disabled={loading}
                    />
                  )}
                </>
              )}

              {isEnded && (
                <button type="button" className={btnOutline} onClick={() => showActionLog()} disabled={loading}>
                  {t('common:showActionLog')}
                </button>
              )}
            </GameFooter>
          </div>
        </>
      )}

      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
      {isGameClear && <WinCelebration show={true} />}
    </div>
  );
}
