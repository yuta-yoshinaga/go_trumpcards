import { useCallback, useMemo, useState } from 'react';
import { bristolApi } from '../api/gameApi';
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
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnDanger, btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BristolMoveZone, BristolResponse } from '../types/card';
import { BristolPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BRISTOL_HELP, parseBristolCommand } from '../utils/cli/commands/bristolCommands';
import { formatBristolState } from '../utils/cli/formatters/bristolFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Tutorial steps for the Bristol solitaire game. */
const BR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="br-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="br-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="br-fan"]', messageKey: 'tutorial.fan', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="br-stock"]', messageKey: 'tutorial.stock', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="br-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Bristol solitaire game page. */
export const BristolPage = withTutorial(BristolPageContent, 'bristol', BR_TUTORIAL_STEPS);

/** Inner content of the Bristol page. */
function BristolPageContent() {
  const {
    t,
    tc,
    actionLog,
    showActionLog,
    hideActionLog,
    confirmOpen,
    requestConfirm,
    confirmReset,
    cancelReset,
    giveUpConfirmOpen,
    requestGiveUpConfirm,
    confirmGiveUp,
    cancelGiveUp,
  } = useGamePageSetup('bristol');
  const { state, loading, error, exec: execApi, retry } = useGameApi(bristolApi.exec);
  const { cardWidth, cardHeight } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bristol', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bristol');
  const cliConfig: CliGameConfig<BristolResponse, Parameters<typeof bristolApi.exec>> = useMemo(
    () => ({
      gameName: 'bristol',
      parseCommand: parseBristolCommand,
      formatResponse: formatBristolState,
      helpText: BRISTOL_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  // Selected source card (a tableau column top or a fan top), or null.
  const [selected, setSelected] = useState<BristolMoveZone | null>(null);

  const handleReset = useCallback(() => {
    setSelected(null);
    execApi('reset');
  }, [execApi]);
  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleGiveUp = useCallback(() => execApi('giveup'), [execApi]);

  // Give-up is irreversible, so route the button through the confirm dialog —
  // matching reset's guard (issue #2099).
  const confirmGiveUpAction = useCallback(
    () => requestGiveUpConfirm(handleGiveUp),
    [requestGiveUpConfirm, handleGiveUp],
  );
  const handleHint = useCallback(() => execApi('hint'), [execApi]);
  const handleAutoComplete = useCallback(() => execApi('autocomplete'), [execApi]);
  const handleUndo = useCallback(() => execApi('undo'), [execApi]);

  const handleTableauClick = useCallback(
    (col: number) => {
      if (selected) {
        if (selected.zone === 'tableau' && selected.col === col) {
          setSelected(null);
          return;
        }
        execApi('move', selected, { zone: 'tableau', col });
        setSelected(null);
        return;
      }
      setSelected({ zone: 'tableau', col });
    },
    [execApi, selected],
  );

  const handleFanClick = useCallback(
    (col: number) => {
      if (selected && selected.zone === 'fan' && selected.col === col) {
        setSelected(null);
        return;
      }
      setSelected({ zone: 'fan', col });
    },
    [selected],
  );

  const handleFoundationClick = useCallback(
    (fIdx: number) => {
      if (!selected) return;
      execApi('move', selected, { zone: 'foundation', col: fIdx });
      setSelected(null);
    },
    [execApi, selected],
  );

  const theme = useMemo(() => gameTheme.bristol, []);

  const phase = state?.phase ?? BristolPhase.PLAYING;
  const isPlaying = phase === BristolPhase.PLAYING;
  const isGameClear = phase === BristolPhase.GAME_CLEAR;
  const isEnded = phase === BristolPhase.GAME_CLEAR || phase === BristolPhase.GAME_OVER;

  const phaseName = isGameClear
    ? t('phase.gameClear')
    : phase === BristolPhase.GAME_OVER
      ? t('phase.gameOver')
      : t('phase.playing');

  if (error) return <ErrorAlert message={error} onRetry={retry} />;
  if (!state) return null;

  const colOffset = Math.round(cardHeight * 0.32);
  const isSelected = (zone: BristolMoveZone) => !!selected && selected.zone === zone.zone && selected.col === zone.col;

  return (
    <GamePageShell
      title={tc('nav.bristol')}
      gameThemeBg={theme.bg}
      phaseName={phaseName}
      gamePath="/bristol"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      giveUpConfirmOpen={giveUpConfirmOpen}
      confirmGiveUp={confirmGiveUp}
      cancelGiveUp={cancelGiveUp}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('moveCount')}: {state.moveCount}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
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

          <LandscapeBanner message={phaseName} />

          <div className="flex-1 overflow-y-auto px-4 pt-3 lg:px-8">
            {/* Foundations */}
            <div className="mb-3 flex items-start gap-2" data-tutorial="br-foundation">
              <span className="w-14 shrink-0 pt-2 text-xs text-ds-text-muted">{t('foundation')}</span>
              <div className="flex gap-2">
                {state.foundation.map((pile, i) => (
                  <button
                    key={`f-${i}`}
                    type="button"
                    onClick={() => handleFoundationClick(i)}
                    disabled={!isPlaying || !selected || loading}
                    aria-label={
                      pile.length > 0
                        ? t('foundationAria', { num: i, card: cardAlt(pile[pile.length - 1]), count: pile.length })
                        : t('foundationAriaEmpty', { num: i })
                    }
                    className={
                      selected
                        ? `rounded border p-0.5 ${focusRingWhite} border-ds-info`
                        : `rounded border p-0.5 ${focusRingWhite} border-white/30`
                    }
                    style={{ width: cardWidth + 4, height: cardHeight + 4 }}
                  >
                    {pile.length > 0 ? (
                      <AnimatedCard card={pile[pile.length - 1]} width={cardWidth} />
                    ) : (
                      <span className="flex h-full w-full items-center justify-center text-xs text-ds-text-muted/80">
                        A
                      </span>
                    )}
                  </button>
                ))}
              </div>
            </div>

            {/* Tableau */}
            <div className="mb-0.5 text-center text-xs text-ds-text-muted" data-testid="br-tableau-rule">
              {t('tableauRule', { count: state.tableau.length })}
            </div>
            <div className="mb-3 flex gap-1 sm:gap-2" data-tutorial="br-tableau">
              {state.tableau.map((col, colIdx) => {
                const zone: BristolMoveZone = { zone: 'tableau', col: colIdx };
                const colHeight = col.length > 0 ? (col.length - 1) * colOffset + cardHeight : cardHeight;
                return (
                  <div key={`col-${colIdx}`} className="flex flex-1 flex-col items-center gap-1 min-w-0">
                    <span className="text-xs text-ds-text-muted">#{colIdx + 1}</span>
                    <button
                      type="button"
                      onClick={() => handleTableauClick(colIdx)}
                      disabled={!isPlaying || loading || (col.length === 0 && !selected)}
                      aria-label={t('tableauColAria', { num: colIdx + 1, count: col.length })}
                      aria-pressed={isSelected(zone)}
                      className={
                        isSelected(zone)
                          ? `relative w-full rounded border-2 bg-transparent p-0 ${focusRingWhite} border-ds-info`
                          : `relative w-full rounded border-2 bg-transparent p-0 ${focusRingWhite} border-transparent`
                      }
                      style={{ height: colHeight }}
                    >
                      {col.length === 0 ? (
                        <span
                          className="flex w-full items-center justify-center rounded border-2 border-dashed border-white/20 text-xs text-ds-text-muted"
                          style={{ height: cardHeight }}
                        >
                          {t('empty')}
                        </span>
                      ) : (
                        col.map((card, cardIdx) => (
                          <div
                            key={`c-${colIdx}-${cardIdx}`}
                            className="absolute left-0 right-0"
                            style={{ top: cardIdx * colOffset }}
                          >
                            <AnimatedCard card={card} width={cardWidth} draggable={false} style={{ width: '100%' }} />
                          </div>
                        ))
                      )}
                    </button>
                  </div>
                );
              })}
            </div>

            {/* Fans + Stock */}
            <div className="mb-3 flex items-start gap-3">
              <div className="flex gap-2" data-tutorial="br-fan">
                {state.fan.map((pile, i) => {
                  const top = pile.length > 0 ? pile[pile.length - 1] : null;
                  const zone: BristolMoveZone = { zone: 'fan', col: i };
                  return (
                    <div key={`fan-${i}`} className="flex flex-col items-center gap-1">
                      <span className="text-xs text-ds-text-muted">
                        {t('fan')} {i}
                      </span>
                      {top ? (
                        <button
                          type="button"
                          onClick={() => handleFanClick(i)}
                          disabled={!isPlaying || loading}
                          aria-label={t('fanAria', { num: i, card: cardAlt(top), count: pile.length })}
                          aria-pressed={isSelected(zone)}
                          className={`relative rounded border-2 bg-transparent p-0 ${focusRingWhite} ${
                            isSelected(zone) ? 'border-ds-info' : 'border-transparent'
                          }`}
                        >
                          <AnimatedCard card={top} width={cardWidth} draggable={false} />
                          {pile.length >= 2 ? (
                            <span
                              aria-hidden="true"
                              data-testid={`br-fan-count-${i.toString()}`}
                              className="absolute bottom-0.5 right-0.5 px-1 rounded bg-ds-accent text-ds-text-on-accent text-[10px] font-bold shadow-sm pointer-events-none"
                            >
                              {pile.length}
                            </span>
                          ) : null}
                        </button>
                      ) : (
                        <div
                          role="img"
                          aria-label={t('fanAriaEmpty', { num: i })}
                          className="rounded border border-dashed border-white/30"
                          style={{ width: cardWidth, height: cardHeight }}
                        />
                      )}
                      <span className="text-xs text-ds-text-muted">({pile.length})</span>
                    </div>
                  );
                })}
              </div>

              <div className="flex flex-col items-center" data-tutorial="br-stock">
                <button
                  type="button"
                  onClick={handleDraw}
                  disabled={!isPlaying || loading || state.stockCount === 0}
                  className="rounded border border-white/30"
                  aria-label={t('stock')}
                  style={{ width: cardWidth, height: cardHeight }}
                >
                  {state.stockCount > 0 ? (
                    <AnimatedCardBack width={cardWidth} />
                  ) : (
                    <span className="text-xs text-ds-text-muted/80">{t('empty')}</span>
                  )}
                </button>
                <span className="mt-1 text-xs text-ds-text-muted">
                  {t('stock')}: {state.stockCount}
                </span>
              </div>
            </div>

            <GameMessageBox
              message={selected ? t('selectDestination') : state.message}
              messageCode={selected ? undefined : state.messageCode}
              messageParams={state.messageParams}
            />
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${theme.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap items-center gap-2" data-tutorial="br-action-buttons">
              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={`${btnPrimary} ${focusRingWhite}`}
                    onClick={handleDraw}
                    disabled={loading || state.stockCount === 0}
                  >
                    {t('draw')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${focusRingWhite}`}
                    onClick={handleHint}
                    disabled={loading}
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${focusRingWhite}`}
                    onClick={handleAutoComplete}
                    disabled={loading}
                  >
                    {t('autoComplete')}
                  </button>
                  <button
                    type="button"
                    className={`${btnOutline} ${focusRingWhite}`}
                    onClick={handleUndo}
                    disabled={!state.canUndo || loading}
                  >
                    {t('undo')}
                  </button>
                  <button
                    type="button"
                    className={`${btnDanger} ${focusRingWhite}`}
                    onClick={confirmGiveUpAction}
                    disabled={loading}
                  >
                    {t('giveup')}
                  </button>
                </>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="br-reset-button"
                className={focusRingWhite}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
