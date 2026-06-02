import { useCallback, useMemo, useState } from 'react';
import { type OsmosisMoveZone, osmosisApi } from '../api/gameApi';
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
import type { OsmosisResponse } from '../types/card';
import { OsmosisPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { OSMOSIS_HELP, parseOsmosisCommand } from '../utils/cli/commands/osmosisCommands';
import { formatOsmosisState } from '../utils/cli/formatters/osmosisFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Tutorial steps for the Osmosis solitaire game. */
const OS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="os-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="os-reserve"]', messageKey: 'tutorial.reserve', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="os-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="os-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="os-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Osmosis solitaire game page. */
export const OsmosisPage = withTutorial(OsmosisPageContent, 'osmosis', OS_TUTORIAL_STEPS);

/** Inner content of the Osmosis page. */
function OsmosisPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('osmosis');
  const { state, loading, error, exec: execApi, retry } = useGameApi(osmosisApi.exec);
  const { cardWidth, cardHeight } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('osmosis', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('osmosis');
  const cliConfig: CliGameConfig<OsmosisResponse, Parameters<typeof osmosisApi.exec>> = useMemo(
    () => ({
      gameName: 'osmosis',
      parseCommand: parseOsmosisCommand,
      formatResponse: formatOsmosisState,
      helpText: OSMOSIS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  // Selected source card (waste or a reserve column top), or null.
  const [selected, setSelected] = useState<OsmosisMoveZone | null>(null);

  const handleReset = useCallback(() => {
    setSelected(null);
    execApi('reset');
  }, [execApi]);
  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleGiveUp = useCallback(() => execApi('giveup'), [execApi]);
  const handleHint = useCallback(() => execApi('hint'), [execApi]);
  const handleAutoComplete = useCallback(() => execApi('autocomplete'), [execApi]);
  const handleUndo = useCallback(() => execApi('undo'), [execApi]);

  const handleSelectSource = useCallback((zone: OsmosisMoveZone) => {
    setSelected((prev) => (prev && prev.zone === zone.zone && prev.col === zone.col ? null : zone));
  }, []);

  const handleFoundationClick = useCallback(
    (fIdx: number) => {
      if (!selected) return;
      execApi('move', selected, { zone: 'foundation', col: fIdx });
      setSelected(null);
    },
    [execApi, selected],
  );

  const theme = useMemo(() => gameTheme.osmosis, []);

  const phase = state?.phase ?? OsmosisPhase.PLAYING;
  const isPlaying = phase === OsmosisPhase.PLAYING;
  const isGameClear = phase === OsmosisPhase.GAME_CLEAR;
  const isEnded = phase === OsmosisPhase.GAME_CLEAR || phase === OsmosisPhase.GAME_OVER;

  const phaseName = isGameClear
    ? t('phase.gameClear')
    : phase === OsmosisPhase.GAME_OVER
      ? t('phase.gameOver')
      : t('phase.playing');

  if (error) return <ErrorAlert message={error} onRetry={retry} />;
  if (!state) return null;

  const topWaste = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const isSelected = (zone: OsmosisMoveZone) => !!selected && selected.zone === zone.zone && selected.col === zone.col;

  return (
    <GamePageShell
      title={tc('nav.osmosis')}
      gameThemeBg={theme.bg}
      phaseName={phaseName}
      gamePath="/osmosis"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('baseRank')}: {state.baseRank || '?'}
          </span>
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
            {/* Foundation rows */}
            <div className="mb-3 flex flex-col gap-2" data-tutorial="os-foundation">
              <span className="text-xs text-ds-text-muted">{t('foundation')}</span>
              {state.foundation.map((pile, i) => (
                <button
                  key={`f-${i}`}
                  type="button"
                  onClick={() => handleFoundationClick(i)}
                  disabled={!isPlaying || !selected || loading}
                  aria-label={`${t('foundation')} ${i}`}
                  className={`flex items-center gap-2 rounded border p-1 text-left ${focusRingWhite} ${
                    selected ? 'border-ds-info' : 'border-white/30'
                  }`}
                >
                  <span className="w-5 text-xs text-ds-text-muted">#{i}</span>
                  <div className="relative" style={{ width: cardWidth, height: cardHeight }}>
                    {pile.length > 0 ? (
                      <AnimatedCard card={pile[pile.length - 1]} width={cardWidth} />
                    ) : (
                      <span className="absolute inset-0 flex items-center justify-center text-xs text-ds-text-muted/80">
                        {t('foundation')}
                      </span>
                    )}
                  </div>
                  <span className="text-xs text-ds-text-muted">({pile.length})</span>
                </button>
              ))}
            </div>

            {/* Reserve columns */}
            <div className="mb-3 flex gap-2" data-tutorial="os-reserve">
              {state.reserve.map((pile, i) => {
                const top = pile.length > 0 ? pile[pile.length - 1] : null;
                const zone: OsmosisMoveZone = { zone: 'reserve', col: i };
                return (
                  <div key={`r-${i}`} className="flex flex-col items-center gap-1">
                    <span className="text-xs text-ds-text-muted">#{i}</span>
                    {top ? (
                      <button
                        type="button"
                        onClick={() => handleSelectSource(zone)}
                        disabled={!isPlaying || loading}
                        aria-label={`${t('reserve')} ${i}`}
                        aria-pressed={isSelected(zone)}
                        className={`p-0 border-2 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                          isSelected(zone) ? 'border-ds-info' : 'border-transparent'
                        }`}
                      >
                        <AnimatedCard card={top} width={cardWidth} draggable={false} />
                      </button>
                    ) : (
                      <div
                        className="rounded border border-dashed border-white/30"
                        style={{ width: cardWidth, height: cardHeight }}
                      />
                    )}
                    <span className="text-xs text-ds-text-muted">({pile.length})</span>
                  </div>
                );
              })}
            </div>

            {/* Stock / Waste */}
            <div className="mb-3 flex gap-3" data-tutorial="os-stock-waste">
              <div className="flex flex-col items-center">
                <button
                  type="button"
                  onClick={handleDraw}
                  disabled={!isPlaying || loading}
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

              <div className="flex flex-col items-center">
                <div style={{ width: cardWidth, height: cardHeight }}>
                  {topWaste ? (
                    <button
                      type="button"
                      onClick={() => handleSelectSource({ zone: 'waste' })}
                      disabled={!isPlaying || loading}
                      aria-label={t('waste')}
                      aria-pressed={isSelected({ zone: 'waste' })}
                      className={`p-0 border-2 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                        isSelected({ zone: 'waste' }) ? 'border-ds-info' : 'border-transparent'
                      }`}
                    >
                      <AnimatedCard card={topWaste} width={cardWidth} draggable={false} />
                    </button>
                  ) : (
                    <div
                      className="rounded border border-dashed border-white/30"
                      style={{ width: cardWidth, height: cardHeight }}
                    />
                  )}
                </div>
                <span className="mt-1 text-xs text-ds-text-muted">{t('waste')}</span>
              </div>
            </div>

            <GameMessageBox
              message={selected ? t('selectFoundation') : state.message}
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
            <div className="flex flex-wrap items-center gap-2" data-tutorial="os-action-buttons">
              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={`${btnPrimary} ${focusRingWhite}`}
                    onClick={handleDraw}
                    disabled={loading}
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
                    onClick={handleGiveUp}
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
                dataTutorial="os-reset-button"
                className={focusRingWhite}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
