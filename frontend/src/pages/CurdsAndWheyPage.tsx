import { useEffect, useState } from 'react';
import { curdsandwheyApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card } from '../types/card';
import { CurdsAndWheyPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { curdsAndWheyAutoMoveTarget } from '../utils/curdsAndWheyAutoMoveTarget';
import { isGrabbable, movableFromIndex } from '../utils/curdsAndWheyRun';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Simple Simon tutorial step definitions. */
const SS_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ss-columns"]', messageKey: 'tutorial.columns', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ss-controls"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ss-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps numeric Simple Simon phases to i18n phase-label keys. */
const SS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CurdsAndWheyPhase.PLAYING]: 'playing',
  [CurdsAndWheyPhase.GAME_CLEAR]: 'gameClear',
  [CurdsAndWheyPhase.GAME_OVER]: 'gameOver',
};

/** A selected move source: a card index within a column. */
interface Selection {
  col: number;
  idx: number;
}

/** Renders the Simple Simon game page. */
export const CurdsAndWheyPage = withTutorial(CurdsAndWheyPageContent, 'curdsandwhey', SS_TUTORIAL_STEPS);

/** Inner content of the Simple Simon page, wrapped by TutorialProvider. */
function CurdsAndWheyPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('curdsandwhey');
  const { state, loading, error, exec, retry } = useGameApi(curdsandwheyApi.exec);
  const [selected, setSelected] = useState<Selection | null>(null);
  // Transient notice shown when a double-click auto-move finds no destination.
  const [autoMoveNotice, setAutoMoveNotice] = useState<string | null>(null);

  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // Clear a stale selection and any auto-move notice whenever the board changes
  // (move, undo).
  // biome-ignore lint/correctness/useExhaustiveDependencies: deps are the change-trigger, not read in the body.
  useEffect(() => {
    setSelected(null);
    setAutoMoveNotice(null);
  }, [state?.moveCount]);

  const phaseNames = usePhaseNames('curdsandwhey', SS_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const w = Math.round(cardWidth * 0.58);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('curdsandwhey', state);

  if (!state) return <GameSkeleton gameKey="curdsandwhey" layout={{ kind: 'tableau', topRow: 0, tableau: 10 }} />;

  const isClear = state.phase === CurdsAndWheyPhase.GAME_CLEAR;
  const isEnd = isClear || state.phase === CurdsAndWheyPhase.GAME_OVER;
  const canAct = !isEnd;
  const phaseName = phaseNames[state.phase] ?? '';

  const handleReset = () => {
    hideActionLog();
    setSelected(null);
    exec('reset');
  };

  // Double-click a grabbable run: auto-move it to the best legal destination
  // (same-suit link > rank-only link > empty column). If none exists, deselect
  // and show a notice. Single-click selection is preserved via the e.detail
  // guard in the card's onClick.
  const autoMoveCard = (col: number, idx: number) => {
    if (!canAct || !isGrabbable(state.columns[col], idx)) return;
    const toCol = curdsAndWheyAutoMoveTarget(state.columns, col, idx);
    if (toCol === null) {
      setSelected(null);
      setAutoMoveNotice(t('noAutoMove'));
      return;
    }
    setAutoMoveNotice(null);
    exec('m', { fromCol: col, cardIndex: idx, toCol });
    setSelected(null);
  };

  // Click a card: select it as the source, or — if a source in another column is
  // already selected — move that run onto this column.
  const clickCard = (col: number, idx: number) => {
    if (!canAct) return;
    if (selected && selected.col !== col) {
      exec('m', { fromCol: selected.col, cardIndex: selected.idx, toCol: col });
      setSelected(null);
      return;
    }
    // Only a valid movable run (same-suit descending to the tail) can be grabbed
    // as a source; ignore clicks on cards above the run boundary.
    if (!isGrabbable(state.columns[col], idx)) return;
    setSelected({ col, idx });
  };

  // Click an (empty) column: move the selected run onto it.
  const clickColumn = (col: number) => {
    if (!canAct || !selected || selected.col === col) return;
    exec('m', { fromCol: selected.col, cardIndex: selected.idx, toCol: col });
    setSelected(null);
  };

  const renderColumn = (column: Card[], col: number) => {
    const isDestination = Boolean(selected && selected.col !== col);
    const runStart = movableFromIndex(column);
    return (
      <div
        key={`col-${col}`}
        className={`flex flex-col items-center rounded p-0.5 ${isDestination ? 'ring-1 ring-ds-success' : ''}`}
        style={{ minHeight: Math.round(w * 1.4) }}
        data-testid={`column-${col}`}
      >
        {column.length === 0 ? (
          <button
            type="button"
            className="rounded border border-dashed border-white/25 bg-black/20"
            style={{ width: w, height: Math.round(w * 1.4) }}
            onClick={canAct ? () => clickColumn(col) : undefined}
            disabled={!canAct}
            title={t('empty')}
            aria-label={t('emptyColAria', { col: col + 1 })}
            data-testid={`column-${col}-drop`}
          />
        ) : (
          column.map((c, i) => {
            const inSelectedRun = Boolean(selected && selected.col === col && i >= selected.idx);
            // Grabbable = head of a valid movable run in this column. In a
            // destination column every card just forwards to the move.
            const grabbable = i >= runStart;
            const clickable = isDestination || grabbable;
            // Highlight the movable-run boundary only while this column can be a
            // source (no selection, or the selection is here).
            const showRunHint = !isDestination && grabbable && !inSelectedRun;
            const ring = inSelectedRun ? 'ring-2 ring-ds-warning' : showRunHint ? 'ring-1 ring-ds-success/70' : '';
            return (
              <button
                type="button"
                key={`col-${col}-${i}`}
                className={`rounded ${ring} ${clickable ? '' : 'cursor-not-allowed'}`}
                style={{ marginTop: i === 0 ? 0 : -Math.round(w * 1.05) }}
                onClick={
                  canAct
                    ? (e) => {
                        // The 2nd click of a double-click also fires onClick
                        // (detail === 2); ignore it so onDoubleClick owns the
                        // auto-move and single-click selection is preserved.
                        if (e.detail >= 2) return;
                        clickCard(col, i);
                      }
                    : undefined
                }
                onDoubleClick={canAct && grabbable ? () => autoMoveCard(col, i) : undefined}
                disabled={!canAct || !clickable}
                data-testid={`card-${col}-${i}`}
                data-grabbable={grabbable}
                aria-label={t('cardPosAria', { card: cardAlt(c), col: col + 1, pos: i + 1 })}
                aria-pressed={inSelectedRun}
              >
                <CardImage card={c} width={w} />
              </button>
            );
          })
        )}
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.curdsandwhey')}
      gameThemeBg={gameTheme.curdsandwhey.bg}
      phaseName={phaseName}
      gamePath="/curdsandwhey"
      gameEndFlag={isEnd}
      winShow={isClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-2 lg:px-6">
        <div className="text-ds-text-muted text-xs mb-1">
          {t('completedSuits', { count: state.completedSuits })} · {t('moveCount', { count: state.moveCount })}
        </div>
        <div className="grid grid-cols-5 sm:grid-cols-10 gap-1 items-start" data-tutorial="ss-columns">
          {state.columns.map((column, i) => renderColumn(column, i))}
        </div>
        {canAct && (
          <div className="mt-2 text-ds-text-primary text-xs" role="status" data-testid="ss-guidance">
            {selected === null ? t('selectSource') : t('selectDestination')}
          </div>
        )}
        {canAct && autoMoveNotice && (
          <div className="mt-1 text-ds-text-muted text-xs" role="status" data-testid="ss-automove-notice">
            {autoMoveNotice}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

        <SettingsPanel
          title={tc('settings.title')}
          groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
        />

        <ActionLogSection
          isEndPhase={isEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.curdsandwhey.footer} px-3 py-2.5`}>
        <ErrorAlert message={error} onRetry={retry} />
        <div className="flex flex-wrap gap-2 items-center" data-tutorial="ss-controls">
          {canAct && state.canUndo && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => exec('u')}
              disabled={loading}
              data-testid="undo-button"
            >
              {t('undo')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => exec('hint')}
              disabled={loading}
              data-testid="hint-button"
            >
              {t('hint')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => exec('g')}
              disabled={loading}
              data-testid="giveup-button"
            >
              {t('giveup')}
            </button>
          )}
          <GameResetButton
            isGameEnd={isEnd}
            onReset={handleReset}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="ss-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
