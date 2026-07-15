import { useMemo } from 'react';
import type { trenteetquaranteApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useTrenteEtQuaranteGame } from '../hooks/useTrenteEtQuaranteGame';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, TrenteEtQuaranteResponse } from '../types/card';
import { TrenteEtQuaranteBetType, TrenteEtQuaranteWinningRow } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTrenteEtQuaranteCommand, TRENTEETQUARANTE_HELP } from '../utils/cli/commands/trenteetquaranteCommands';
import { formatTrenteEtQuaranteState } from '../utils/cli/formatters/trenteetquaranteFormatter';
import type { CliGameConfig } from '../utils/cli/types';

const TEQ_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="teq-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teq-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="teq-results"]', messageKey: 'tutorial.results', placement: 'bottom', advanceOn: 'next' },
];

/** The four even-money bets, in display order, keyed to their i18n labels. */
const BET_OPTIONS: readonly { type: number; labelKey: string; descKey: string }[] = [
  { type: TrenteEtQuaranteBetType.NOIR, labelKey: 'betType.noir', descKey: 'betType.noirDesc' },
  { type: TrenteEtQuaranteBetType.ROUGE, labelKey: 'betType.rouge', descKey: 'betType.rougeDesc' },
  { type: TrenteEtQuaranteBetType.COULEUR, labelKey: 'betType.couleur', descKey: 'betType.couleurDesc' },
  { type: TrenteEtQuaranteBetType.INVERSE, labelKey: 'betType.inverse', descKey: 'betType.inverseDesc' },
];

/** Renders the Trente et Quarante (Rouge et Noir) game page. */
export const TrenteEtQuarantePage = withTutorial(TrenteEtQuarantePageContent, 'trenteetquarante', TEQ_TUTORIAL_STEPS);

function TrenteEtQuarantePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, confirmReset, cancelReset } =
    useGamePageSetup('trenteetquarante');
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  const {
    state,
    loading,
    error,
    retry,
    execApi,
    betType,
    setBetType,
    betAmount,
    setBetAmount,
    lastBet,
    isBetPhase,
    isResultPhase,
    canRebet,
    handleBet,
    handleNextRound,
    handleRebet,
  } = useTrenteEtQuaranteGame();

  // Merge the locally selected bet into the state so the hint reflects the pending choice.
  const hintState = useMemo(() => (state ? { ...state, currentBet: betType } : null), [state, betType]);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('trenteetquarante', hintState);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('trenteetquarante');
  const cliConfig: CliGameConfig<TrenteEtQuaranteResponse, Parameters<typeof trenteetquaranteApi.exec>> = useMemo(
    () => ({
      gameName: 'trenteetquarante',
      parseCommand: parseTrenteEtQuaranteCommand,
      formatResponse: formatTrenteEtQuaranteState,
      helpText: TRENTEETQUARANTE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleBet, enabled: isBetPhase },
      { key: 'r', action: handleNextRound, enabled: isResultPhase },
      // Power-user shortcut: 'e' replays the last bet once the round has ended.
      { key: 'e', action: handleRebet, enabled: isResultPhase && canRebet },
    ],
    [handleBet, handleNextRound, handleRebet, isBetPhase, isResultPhase, canRebet],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.trenteetquarante.bg}`}>
        <div className="text-ds-text-primary">Loading...</div>
      </div>
    );
  }

  const phaseName = isBetPhase ? t('phase.bet') : t('phase.result');
  const hasRows = state.noirRow.length > 0 || state.rougeRow.length > 0;

  return (
    <GamePageShell
      title={tc('nav.trenteetquarante')}
      gameThemeBg={gameTheme.trenteetquarante.bg}
      phaseName={phaseName}
      gamePath="/trenteetquarante"
      gameEndFlag={isResultPhase}
      winShow={isResultPhase && state.result > 0}
      onCelebrate={() => playSound('winFanfare')}
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
            className={[`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`, !isBetPhase && 'flex-1']
              .filter(Boolean)
              .join(' ')}
          >
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <label className="flex items-center gap-1 text-ds-text-primary text-xs justify-center mb-2 cursor-pointer">
              <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>

            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-2">
                <p className="text-ds-text-muted text-lg text-center">{t('betGuide')}</p>
              </div>
            )}

            {hasRows && (
              <div className="mb-4 flex flex-col gap-4" data-tutorial="teq-results">
                <CardRow
                  label={`${t('label.noirRow')} (${state.noirTotal})`}
                  cards={state.noirRow}
                  cardWidth={cardWidth}
                  highlight={isResultPhase && !state.refait && state.winningRow === TrenteEtQuaranteWinningRow.NOIR}
                />
                <CardRow
                  label={`${t('label.rougeRow')} (${state.rougeTotal})`}
                  cards={state.rougeRow}
                  cardWidth={cardWidth}
                  highlight={isResultPhase && !state.refait && state.winningRow === TrenteEtQuaranteWinningRow.ROUGE}
                />
              </div>
            )}

            {isResultPhase && (
              <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="teq-result">
                {state.refait ? (
                  <div className="font-bold text-ds-warning">{t('result.refait')}</div>
                ) : (
                  <div className="font-bold">
                    {state.result > 0 ? t('result.win') : state.result < 0 ? t('result.lose') : ''}
                  </div>
                )}
                <div>
                  {t('payout.total')}: {state.payout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.trenteetquarante.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel title={tc('settings.title')} groups={[]} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-3 pb-2" data-tutorial="teq-bet-controls">
                {/* biome-ignore lint/a11y/useSemanticElements: a flex row of bet buttons; fieldset would break the layout */}
                <div
                  className="flex flex-wrap justify-center items-start gap-2"
                  role="group"
                  aria-label={t('label.betType')}
                >
                  {BET_OPTIONS.map((opt) => {
                    const selected = betType === opt.type;
                    return (
                      <div key={opt.type} className="flex flex-col items-center w-24">
                        <button
                          type="button"
                          aria-pressed={selected}
                          title={t(opt.descKey)}
                          className={`w-full ${selected ? `${btnPrimary} ring-2 ring-white` : btnSecondary}`}
                          onClick={() => setBetType(opt.type)}
                          disabled={loading}
                          data-testid={`teq-bet-${opt.type}`}
                        >
                          {/* Shape cue for the current choice, in addition to the colour. */}
                          {selected && <span aria-hidden="true">✓ </span>}
                          {t(opt.labelKey)}
                        </button>
                        {/* Couleur/Inverse are opaque on sight — keep the meaning always visible. */}
                        <span className="mt-0.5 text-[10px] leading-tight text-ds-text-muted text-center">
                          {t(opt.descKey)}
                        </span>
                      </div>
                    );
                  })}
                </div>
                <ChipBetInput
                  id="trenteetquarante-bet-amount"
                  label={t('label.stake')}
                  value={betAmount}
                  onChange={setBetAmount}
                  max={state.chips}
                />
                {canRebet && lastBet !== null && lastBet.stake !== betAmount && (
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => setBetAmount(lastBet.stake)}
                    disabled={loading}
                    data-testid="teq-previous-bet"
                  >
                    {t('previousBet', { amount: lastBet.stake })}
                  </button>
                )}
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleBet}
                  disabled={loading || betAmount <= 0 || betAmount > state.chips}
                  data-testid="teq-deal-button"
                >
                  {t('button.deal')}
                </button>
              </div>
            )}
            {isResultPhase && (
              <div className="flex justify-center gap-2 pb-2" data-tutorial="teq-action-buttons">
                {canRebet && (
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={handleRebet}
                    disabled={loading}
                    data-testid="teq-rebet-button"
                  >
                    {t('button.rebet', { amount: lastBet?.stake })}
                  </button>
                )}
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleNextRound}
                  disabled={loading}
                  data-testid="teq-next-round-button"
                >
                  {t('button.nextRound')}
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

/** Renders one labelled row of dealt cards, optionally highlighted as the winning row. */
function CardRow({
  label,
  cards,
  cardWidth,
  highlight,
}: {
  label: string;
  cards: Card[];
  cardWidth: number;
  highlight: boolean;
}) {
  return (
    <div className={highlight ? 'rounded-lg ring-2 ring-ds-success p-2' : 'p-2'}>
      <div className={`text-center font-bold mb-1 ${highlight ? 'text-ds-success' : 'text-ds-text-primary'}`}>
        {label}
      </div>
      <div className="flex justify-center gap-2 flex-wrap">
        {cards.map((card, i) => (
          <AnimatedCard key={`${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
        ))}
      </div>
    </div>
  );
}
