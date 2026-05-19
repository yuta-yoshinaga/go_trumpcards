import { useCallback, useEffect, useMemo, useState } from 'react';
import { type NertzMoveZone, nertzApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnOutline, btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, NertzPlayerData, NertzResponse, NertzTableauCard } from '../types/card';
import { NertzPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { NERTZ_HELP, parseNertzCommand } from '../utils/cli/commands/nertzCommands';
import { formatNertzState } from '../utils/cli/formatters/nertzFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** CPU tick interval in milliseconds while the round is active. */
const NERTZ_TICK_INTERVAL_MS = 700;

/** Tutorial step definitions for the Nertz / Pounce page. */
const NERTZ_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="nertz-foundations"]',
    messageKey: 'tutorial.foundations',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="nertz-pile"]', messageKey: 'tutorial.nertzPile', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="nertz-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="nertz-stock"]', messageKey: 'tutorial.stock', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="nertz-reset"]', messageKey: 'tutorial.resetButton', placement: 'top', advanceOn: 'next' },
];

type Selection = { kind: 'nertz' } | { kind: 'waste' } | { kind: 'tableau'; col: number; cardIndex: number } | null;

/** Renders the Nertz / Pounce game page. */
export const NertzPage = withTutorial(NertzPageContent, 'nertz', NERTZ_TUTORIAL_STEPS);

function NertzPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('nertz');
  const runApi = useCallback((...args: Parameters<typeof nertzApi.exec>) => nertzApi.exec(...args), []);
  const gameApi = useGameApi<NertzResponse, Parameters<typeof nertzApi.exec>>(runApi);
  const { state, loading, error, retry } = gameApi;
  const apiCall = gameApi.exec;
  const { hint, hintEnabled, setHintEnabled } = useGameHint('nertz', state);
  const cliMode = useCliMode('nertz');
  const cliConfig: CliGameConfig<NertzResponse, Parameters<typeof nertzApi.exec>> = useMemo(
    () => ({
      gameName: 'nertz',
      parseCommand: parseNertzCommand,
      formatResponse: formatNertzState,
      helpText: NERTZ_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, cliConfig, state, {
    addInput: cliMode.addInput,
    addOutput: cliMode.addOutput,
    addError: cliMode.addError,
    clearLog: cliMode.clearLog,
  });

  const [selection, setSelection] = useState<Selection>(null);

  useMountReset(apiCall);

  // CPU tick driver: while the round is active, periodically advance CPUs.
  useEffect(() => {
    if (!state) return;
    if (state.phase !== NertzPhase.PLAYING) return;
    const id = window.setInterval(() => {
      void apiCall('tick');
    }, NERTZ_TICK_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [state, apiCall]);

  const human = state?.players[0];
  const isHumanTurn = state?.phase === NertzPhase.PLAYING;
  const isRoundEnd = state?.phase === NertzPhase.ROUND_END;
  const isGameEnd = state?.phase === NertzPhase.GAME_END;

  const handleManualReset = useCallback(() => {
    setSelection(null);
    void apiCall('reset');
  }, [apiCall]);

  const handleNextRound = useCallback(() => {
    setSelection(null);
    void apiCall('nr');
  }, [apiCall]);

  const handleUndo = useCallback(() => {
    setSelection(null);
    void apiCall('u');
  }, [apiCall]);

  const handleDrawStock = useCallback(() => {
    setSelection(null);
    void apiCall('d', { playerIdx: 0 });
  }, [apiCall]);

  const handleSelectNertz = useCallback(() => {
    if (!isHumanTurn) return;
    setSelection((prev) => (prev?.kind === 'nertz' ? null : { kind: 'nertz' }));
  }, [isHumanTurn]);

  const handleSelectWaste = useCallback(() => {
    if (!isHumanTurn) return;
    setSelection((prev) => (prev?.kind === 'waste' ? null : { kind: 'waste' }));
  }, [isHumanTurn]);

  const handleSelectTableau = useCallback(
    (col: number, cardIndex: number) => {
      if (!isHumanTurn) return;
      setSelection((prev) =>
        prev?.kind === 'tableau' && prev.col === col && prev.cardIndex === cardIndex
          ? null
          : { kind: 'tableau', col, cardIndex },
      );
    },
    [isHumanTurn],
  );

  const dispatchMove = useCallback(
    (to: NertzMoveZone) => {
      if (!selection) return;
      const from: NertzMoveZone =
        selection.kind === 'tableau'
          ? { zone: 'tableau', col: selection.col, cardIndex: selection.cardIndex }
          : { zone: selection.kind };
      void apiCall('m', { playerIdx: 0, from, to });
      setSelection(null);
    },
    [apiCall, selection],
  );

  const handleFoundationClick = useCallback(
    (idx: number) => {
      if (!isHumanTurn || !selection) return;
      dispatchMove({ zone: 'foundation', idx });
    },
    [dispatchMove, isHumanTurn, selection],
  );

  const handleTableauTargetClick = useCallback(
    (col: number) => {
      if (!isHumanTurn || !selection) return;
      if (selection.kind === 'tableau' && selection.col === col) {
        setSelection(null);
        return;
      }
      dispatchMove({ zone: 'tableau', col });
    },
    [dispatchMove, isHumanTurn, selection],
  );

  // Keyboard shortcuts for the realtime competitive flow — matches the issue
  // spec (`d` to draw stock, `n`/`w` to pick the Nertz/waste pile, `1-9` to
  // route a held card to a foundation index, `u` to undo).
  const handleFoundationKey = useCallback(
    (idx: number) => {
      if (!state || idx >= state.foundations.length) return;
      if (selection) {
        dispatchMove({ zone: 'foundation', idx });
      }
    },
    [dispatchMove, selection, state],
  );
  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDrawStock },
      { key: 'n', action: handleSelectNertz },
      { key: 'w', action: handleSelectWaste },
      { key: 'u', action: handleUndo },
      ...Array.from({ length: 9 }, (_, i) => ({
        key: String(i + 1),
        action: () => handleFoundationKey(i),
      })),
      { key: 'Escape', action: () => setSelection(null) },
    ],
    [handleDrawStock, handleSelectNertz, handleSelectWaste, handleUndo, handleFoundationKey],
  );
  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: state?.phase === NertzPhase.PLAYING && !loading,
  });

  const phaseName = useMemo(() => {
    if (isGameEnd) return t('phase.gameEnd');
    if (isRoundEnd) return t('phase.roundEnd');
    return t('phase.playing');
  }, [isGameEnd, isRoundEnd, t]);

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state || !human) {
    return (
      <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.nertz.bg}`}>
        <div className="flex-1 flex items-center justify-center text-ds-text-primary">
          <p>{tc('skeleton.loading')}</p>
        </div>
      </div>
    );
  }

  return (
    <GamePageShell
      title={tc('nav.nertz')}
      gameThemeBg={gameTheme.nertz.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/nertz"
      gameEndFlag={isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliMode.cliEnabled} onToggle={cliMode.toggleCli} />}
    >
      {cliMode.cliEnabled ? (
        <CliTerminal logEntries={cliMode.logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto px-3 py-2 space-y-4">
            <div className="bg-black/30 text-ds-text-primary p-3 rounded text-sm flex flex-wrap gap-x-4 gap-y-1">
              <span>
                {t('labels.round')}: {state.roundNumber}
              </span>
              <span>
                {t('labels.moveCount')}: {state.moveCount}
              </span>
              {state.players.map((p, i) => (
                <span key={`scoreline-${i}`}>
                  {p.isHuman ? t('labels.you') : `${t('labels.cpu')}${i}`}: {p.score} ({p.nertzSize})
                </span>
              ))}
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <div data-tutorial="nertz-foundations" className="bg-black/30 text-ds-text-primary p-3 rounded">
              <div className="text-xs uppercase tracking-wide text-ds-text-muted mb-2">{t('labels.foundation')}</div>
              <div className="flex flex-wrap gap-2">
                {state.foundations.map((f, idx) => (
                  <FoundationCell
                    key={`f-${idx}`}
                    idx={idx}
                    top={f.top}
                    size={f.size}
                    onClick={() => handleFoundationClick(idx)}
                    disabled={!isHumanTurn || !selection}
                    ariaLabel={t('labels.foundationN', { n: idx, defaultValue: `Foundation ${idx}` })}
                  />
                ))}
              </div>
            </div>

            {state.players.length > 1 && (
              <div className="bg-black/30 text-ds-text-primary p-3 rounded text-sm space-y-1">
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={`cpu-${p.deckIdx}`} className="flex justify-between">
                      <span>
                        {t('labels.cpu')}
                        {p.deckIdx} — {p.name}
                      </span>
                      <span>
                        {t('labels.nertz')}: {p.nertzSize} / {t('labels.score')}: {p.score}
                      </span>
                    </div>
                  ))}
              </div>
            )}

            <div className="bg-black/30 text-ds-text-primary p-3 rounded space-y-2">
              <div className="flex flex-wrap gap-3 items-start">
                <div data-tutorial="nertz-pile" className="space-y-1">
                  <div className="text-xs uppercase tracking-wide text-ds-text-muted">{t('labels.nertz')}</div>
                  <CardButton
                    card={human.nertzTop ?? null}
                    label={`${human.nertzSize}`}
                    selected={selection?.kind === 'nertz'}
                    disabled={!isHumanTurn || !human.nertzTop}
                    onClick={handleSelectNertz}
                  />
                </div>

                <div className="space-y-1">
                  <div className="text-xs uppercase tracking-wide text-ds-text-muted">{t('labels.waste')}</div>
                  <CardButton
                    card={human.wasteTop ?? null}
                    label={`${human.wasteSize}`}
                    selected={selection?.kind === 'waste'}
                    disabled={!isHumanTurn || !human.wasteTop}
                    onClick={handleSelectWaste}
                  />
                </div>

                <div data-tutorial="nertz-stock" className="space-y-1">
                  <div className="text-xs uppercase tracking-wide text-ds-text-muted">{t('labels.stock')}</div>
                  <button
                    type="button"
                    onClick={handleDrawStock}
                    disabled={!isHumanTurn || loading}
                    className={`${btnSecondary} min-w-[3rem]`}
                  >
                    {human.stockSize}
                  </button>
                </div>
              </div>

              <div data-tutorial="nertz-tableau" className="space-y-1">
                <div className="text-xs uppercase tracking-wide text-ds-text-muted">{t('labels.tableau')}</div>
                <div className="grid grid-cols-4 gap-2">
                  {human.tableau.map((col, colIdx) => (
                    <TableauColumn
                      key={`tab-${colIdx}`}
                      col={col}
                      colIdx={colIdx}
                      selection={selection}
                      onSelectCard={handleSelectTableau}
                      onTarget={() => handleTableauTargetClick(colIdx)}
                      disabled={!isHumanTurn}
                    />
                  ))}
                </div>
              </div>
            </div>

            <ActionLogSection
              isEndPhase={isGameEnd || isRoundEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

          <GameFooter className={`${gameTheme.nertz.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap gap-2 items-center">
              <label className="flex items-center gap-1 text-ds-text-primary text-xs">
                <input
                  type="checkbox"
                  checked={hintEnabled}
                  onChange={(e) => setHintEnabled(e.target.checked)}
                  aria-label={tc('hint.toggle', { ns: 'tutorial' })}
                />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="nertz-reset"
              />
              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnPrimary} onClick={handleNextRound} disabled={loading}>
                  {t('actions.nextRound')}
                </button>
              )}
              {isHumanTurn && state.canUndo && (
                <button type="button" className={btnOutline} onClick={handleUndo} disabled={loading}>
                  {t('actions.undo')}
                </button>
              )}
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

interface FoundationCellProps {
  idx: number;
  top?: Card;
  size: number;
  onClick: () => void;
  disabled: boolean;
  ariaLabel: string;
}

function FoundationCell({ idx, top, size, onClick, disabled, ariaLabel }: FoundationCellProps) {
  const cls = disabled
    ? 'bg-ds-surface text-ds-text-muted border-ds-border-subtle'
    : 'bg-ds-surface-elevated text-ds-text-primary border-ds-border-subtle hover:bg-ds-surface-elevated-hover';
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={`min-w-[3rem] px-2 py-2 rounded border text-sm ${cls}`}
      aria-label={ariaLabel}
    >
      <span className="block text-xs leading-none">F{idx}</span>
      <span className="block text-base font-bold">{top ? `${suitSymbol(top.design)}${top.value}` : '—'}</span>
      <span className="block text-xs">({size})</span>
    </button>
  );
}

interface CardButtonProps {
  card: Card | null;
  label: string;
  selected: boolean;
  disabled: boolean;
  onClick: () => void;
}

function CardButton({ card, label, selected, disabled, onClick }: CardButtonProps) {
  const cls = selected
    ? 'bg-ds-warning text-ds-text-on-accent border-ds-warning'
    : disabled
      ? 'bg-ds-surface text-ds-text-muted border-ds-border-subtle'
      : 'bg-ds-surface-elevated text-ds-text-primary border-ds-border-subtle hover:bg-ds-surface-elevated-hover';
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={`min-w-[3rem] px-2 py-2 rounded border text-sm ${cls}`}
    >
      <span className="block text-base font-bold">{card ? `${suitSymbol(card.design)}${card.value}` : '—'}</span>
      <span className="block text-xs">{label}</span>
    </button>
  );
}

interface TableauColumnProps {
  col: NertzTableauCard[];
  colIdx: number;
  selection: Selection;
  onSelectCard: (col: number, cardIndex: number) => void;
  onTarget: () => void;
  disabled: boolean;
}

function TableauColumn({ col, colIdx, selection, onSelectCard, onTarget, disabled }: TableauColumnProps) {
  if (col.length === 0) {
    return (
      <button
        type="button"
        onClick={onTarget}
        disabled={disabled}
        className="min-h-[4rem] rounded border border-dashed border-white/30 text-ds-text-muted text-xs"
      >
        —
      </button>
    );
  }
  return (
    <div className="flex flex-col gap-1">
      {col.map((tc, i) => {
        const isSelected = selection?.kind === 'tableau' && selection.col === colIdx && selection.cardIndex === i;
        const isLast = i === col.length - 1;
        const cls = isSelected
          ? 'bg-ds-warning text-ds-text-on-accent border-ds-warning'
          : disabled
            ? 'bg-ds-surface text-ds-text-muted border-ds-border-subtle'
            : 'bg-ds-surface-elevated text-ds-text-primary border-ds-border-subtle hover:bg-ds-surface-elevated-hover';
        // Dual-purpose click: the bottom card acts as a drop target when a
        // source is already selected (saves the user a second tap on an
        // empty drop zone), otherwise tapping any card selects it as the
        // move source. PR #1528 review noted this is non-obvious.
        return (
          <button
            key={`t-${colIdx}-${i}`}
            type="button"
            onClick={() => (isLast && selection && !isSelected ? onTarget() : onSelectCard(colIdx, i))}
            disabled={disabled || !tc.card}
            className={`px-2 py-1 rounded border text-sm ${cls}`}
          >
            {tc.card ? `${suitSymbol(tc.card.design)}${tc.card.value}` : '?'}
          </button>
        );
      })}
    </div>
  );
}

function suitSymbol(design: string): string {
  switch (design) {
    case 'SPADE':
      return '♠';
    case 'CLOVER':
      return '♣';
    case 'HEART':
      return '♥';
    case 'DIAMOND':
      return '♦';
    default:
      return '?';
  }
}

export type _NertzPagePlayerSnapshot = NertzPlayerData;
