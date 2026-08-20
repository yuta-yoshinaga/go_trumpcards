import { useCallback, useMemo, useState } from 'react';
import { andarbaharApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { RoadmapTrendBar } from '../components/RoadmapTrendBar';
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
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { AndarBaharResponse } from '../types/card';
import { ANDAR_BAHAR_SIDE_BANDS } from '../types/games/andarbahar';
import { AndarBaharColumn, AndarBaharPhase, AndarBaharSideBand } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { ANDARBAHAR_CLI_HELP, parseAndarBaharCommand } from '../utils/cli/commands/andarbaharCommands';
import { formatAndarBaharState } from '../utils/cli/formatters/andarbaharFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const AB_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ab-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ab-cards"]', messageKey: 'tutorial.cards', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ab-road"]', messageKey: 'tutorial.road', placement: 'top', advanceOn: 'next' },
];

/** Renders the Andar Bahar game page (#5258). */
export const AndarBaharPage = withTutorial(AndarBaharPageContent, 'andarbahar', AB_TUTORIAL_STEPS);

function AndarBaharPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('andarbahar');

  const [betAmount, setBetAmount] = useState(100);
  const [sideAmount, setSideAmount] = useState(0);
  const [sideBand, setSideBand] = useState<number>(AndarBaharSideBand.NONE);
  const [lastBet, setLastBet] = useState<{ amount: number; target: number } | null>(null);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(andarbaharApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('andarbahar');
  const cliConfig: CliGameConfig<AndarBaharResponse, Parameters<typeof andarbaharApi.exec>> = useMemo(
    () => ({
      gameName: 'andarbahar',
      parseCommand: parseAndarBaharCommand,
      formatResponse: formatAndarBaharState,
      helpText: ANDARBAHAR_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === AndarBaharPhase.BET;
  const isEndPhase = state?.phase === AndarBaharPhase.END;

  const handleBet = useCallback(
    (target: number) => {
      setLastBet({ amount: betAmount, target });
      // **賭けていない帯は送らない。** band 0 は「1 枚目」という有効な値なので、
      // 金額 0 のまま送るとサーバに拒否されます。
      const stake = sideBand === AndarBaharSideBand.NONE ? 0 : sideAmount;
      return execApi('bet', betAmount, target, stake, stake > 0 ? sideBand : AndarBaharSideBand.NONE);
    },
    [execApi, betAmount, sideAmount, sideBand],
  );

  const canRebet = lastBet !== null && lastBet.amount > 0 && !!state && lastBet.amount <= state.chips;
  const handleRebet = useCallback(async () => {
    if (lastBet === null) return;
    await execApi('reset');
    await execApi('bet', lastBet.amount, lastBet.target, 0, AndarBaharSideBand.NONE);
  }, [execApi, lastBet]);

  const actionBindings = useMemo(
    () => [
      { key: 'a', action: () => handleBet(AndarBaharColumn.ANDAR), enabled: isBetPhase },
      { key: 'b', action: () => handleBet(AndarBaharColumn.BAHAR), enabled: isBetPhase },
      { key: 'e', action: () => handleRebet(), enabled: isEndPhase && canRebet },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, handleBet, handleRebet, isBetPhase, isEndPhase, canRebet],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('andarbahar', state);

  if (!state) return <GameSkeleton gameKey="andarbahar" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const handleReset = () => execApi('reset');
  const handleClearHistory = () => execApi('clear');

  const phaseName = isBetPhase ? t('phase.bet') : t('phase.end');
  const firstColumnName = state.firstColumn === AndarBaharColumn.ANDAR ? t('label.andar') : t('label.bahar');
  const profit = state.payout - state.betAmount - state.sideAmount;
  const isProfit = profit >= 0;
  const won = state.winner === state.betTarget;

  const columnBlock = (column: number, cards: AndarBaharResponse['andarCards'], label: string) => (
    <div className="flex-1 min-w-0">
      <div className="flex items-center justify-center gap-2 mb-1">
        <span className="text-ds-text-primary text-sm font-bold">{label}</span>
        {state.firstColumn === column && (
          <span className="inline-block rounded-full bg-ds-surface-elevated px-2 py-0.5 text-xs font-medium">
            {t('payout.firstColumnBadge')}
          </span>
        )}
      </div>
      <div className="flex justify-center gap-1 flex-wrap" data-testid={`andarbahar-column-${column}`}>
        {cards.length === 0 ? (
          <span className="text-ds-text-muted text-xs">—</span>
        ) : (
          cards.map((card, i) => (
            <AnimatedCard
              key={`col-${column}-${card.design}-${card.value}-${i}`}
              card={card}
              width={cardWidth}
              className={
                isEndPhase && state.winner === column && i === cards.length - 1 ? 'ring-2 ring-ds-success' : undefined
              }
            />
          ))
        )}
      </div>
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.andarbahar')}
      gameThemeBg={gameTheme.andarbahar.bg}
      phaseName={phaseName}
      gamePath="/andarbahar"
      gameEndFlag={isEndPhase}
      winShow={isEndPhase && won}
      lossShow={isEndPhase && !won}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>
            {t('label.chips')}: {state.chips}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div
            data-testid="card-area"
            data-tutorial="ab-cards"
            className={[`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`, !isBetPhase && 'flex-1']
              .filter(Boolean)
              .join(' ')}
          >
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {isBetPhase && <p className="text-ds-text-muted text-center text-lg py-2">{t('betGuide')}</p>}

            {state.joker && (
              <div className="flex flex-col items-center mb-4">
                <div className="text-ds-warning text-sm font-bold mb-1">{t('label.joker')}</div>
                <AnimatedCard card={state.joker} width={cardWidth} />
                <div className="text-ds-text-muted text-xs mt-1" data-testid="andarbahar-first-column">
                  {t('label.firstColumn')}: {firstColumnName}
                </div>
              </div>
            )}

            <div className="flex gap-4 mb-4 flex-wrap sm:flex-nowrap">
              {columnBlock(AndarBaharColumn.ANDAR, state.andarCards, t('label.andar'))}
              {columnBlock(AndarBaharColumn.BAHAR, state.baharCards, t('label.bahar'))}
            </div>

            {state.dealtCount > 0 && (
              <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="andarbahar-dealt-count">
                {t('label.dealtCount')}: {state.dealtCount}
              </div>
            )}

            {state.history.length > 0 && (
              <div className="mb-4" data-testid="andarbahar-road" data-tutorial="ab-road">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.road')}</div>
                <div className="mx-auto max-w-3xl">
                  <RoadmapTrendBar
                    history={state.history}
                    leftCode={AndarBaharColumn.ANDAR}
                    rightCode={AndarBaharColumn.BAHAR}
                    leftLabel={t('label.andar')}
                    rightLabel={t('label.bahar')}
                    testId="andarbahar-trend-bar"
                  />
                </div>
              </div>
            )}

            {isEndPhase && (
              <div className="text-ds-text-primary text-center text-sm mb-2 space-y-1" data-testid="payout-breakdown">
                <div data-testid="payout-result">
                  {state.winner === AndarBaharColumn.ANDAR ? t('result.andarWins') : t('result.baharWins')}
                </div>
                <div
                  className={`font-medium ${isProfit ? 'text-ds-success' : 'text-ds-error'}`}
                  data-testid="payout-diff"
                >
                  {isProfit ? t('payout.win', { amount: profit }) : t('payout.loss', { amount: Math.abs(profit) })}
                </div>
                {/* **サイドベットは別の賭け** (#5770)。合計だけでは、外したのが
                    メインなのかサイドなのか読めない。張った回だけ内訳を出す。 */}
                {state.sideBand !== AndarBaharSideBand.NONE && (
                  <div className="space-y-1" data-testid="payout-bet-breakdown">
                    <div data-testid="payout-main">
                      {state.mainPayout > 0 ? t('payout.mainHit', { amount: state.mainPayout }) : t('payout.mainMiss')}
                    </div>
                    <div data-testid="payout-side">
                      {state.sidePayout > 0 ? t('payout.sideHit', { amount: state.sidePayout }) : t('payout.sideMiss')}
                    </div>
                  </div>
                )}
                <div className="font-bold">
                  {t('payout.total')}: {state.payout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.andarbahar.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="ab-bet-controls">
                <ChipBetInput
                  id="andarbahar-bet-amount"
                  label={t('label.bet')}
                  value={betAmount}
                  onChange={setBetAmount}
                  max={state.chips}
                />
                <div className="flex items-center gap-2 flex-wrap justify-center">
                  <label htmlFor="andarbahar-side-band" className="text-ds-text-primary text-sm">
                    {t('label.sideBet')}
                  </label>
                  <select
                    id="andarbahar-side-band"
                    className="rounded border border-ds-border bg-ds-surface px-2 py-1 text-sm text-ds-text-primary"
                    value={sideBand}
                    onChange={(e) => setSideBand(Number(e.target.value))}
                  >
                    <option value={AndarBaharSideBand.NONE}>{t('label.noSideBet')}</option>
                    {ANDAR_BAHAR_SIDE_BANDS.map((b) => (
                      <option key={`band-${b.band}`} value={b.band}>
                        {b.lo === b.hi ? t('band.exact', { count: b.lo }) : t('band.range', { lo: b.lo, hi: b.hi })} (
                        {(b.payout / 10).toFixed(1)}x)
                      </option>
                    ))}
                  </select>
                  {sideBand !== AndarBaharSideBand.NONE && (
                    <ChipBetInput
                      id="andarbahar-side-amount"
                      label={t('label.sideAmount')}
                      value={sideAmount}
                      onChange={setSideAmount}
                      max={state.chips}
                    />
                  )}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={() => handleBet(AndarBaharColumn.ANDAR)}
                    disabled={loading}
                    aria-keyshortcuts="a"
                  >
                    {t('button.betAndar')}
                    <KbdBadge label={t('kbd.betAndar')} />
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={() => handleBet(AndarBaharColumn.BAHAR)}
                    disabled={loading}
                    aria-keyshortcuts="b"
                  >
                    {t('button.betBahar')}
                    <KbdBadge label={t('kbd.betBahar')} />
                  </button>
                </div>
              </div>
            )}
            {isEndPhase && (
              <div className="flex justify-center gap-2 pb-2 flex-wrap">
                {canRebet && lastBet !== null && (
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleRebet}
                    disabled={loading}
                    data-testid="ab-rebet-button"
                    aria-keyshortcuts="e"
                  >
                    {t('button.rebet', { amount: lastBet.amount })}
                    <KbdBadge label={t('kbd.rebet')} />
                  </button>
                )}
                <GameResetButton
                  isGameEnd={isEndPhase}
                  onReset={handleReset}
                  requestConfirm={requestConfirm}
                  loading={loading}
                />
                <button type="button" className={btnSecondary} onClick={handleClearHistory} disabled={loading}>
                  {t('button.clearHistory')}
                </button>
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('actionLog.view')}
                </button>
              </div>
            )}
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
