import { useCallback, useMemo, useState } from 'react';
import { montebankApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
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
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MonteBankResponse } from '../types/card';
import { MONTE_BANK_RESULT } from '../types/games/montebank';
import { MonteBankPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MONTEBANK_CLI_HELP, parseMonteBankCommand } from '../utils/cli/commands/montebankCommands';
import { formatMonteBankState } from '../utils/cli/formatters/montebankFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const MB_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="mb-layout"]', messageKey: 'tutorial.layout', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mb-layout"]', messageKey: 'tutorial.even', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mb-bet"]', messageKey: 'tutorial.gate', placement: 'top', advanceOn: 'next' },
];

/** Renders the Monte Bank game page (#5264). */
export const MonteBankPage = withTutorial(MonteBankPageContent, 'montebank', MB_TUTORIAL_STEPS);

/** Result key lookup for the outcome label. */
const resultKeyOf = (r: number) => Object.entries(MONTE_BANK_RESULT).find(([, v]) => v === r)?.[0] ?? 'none';

function MonteBankPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('montebank');

  const [bet, setBet] = useState(50);
  const [selected, setSelected] = useState(0);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(montebankApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('montebank');
  const cliConfig: CliGameConfig<MonteBankResponse, Parameters<typeof montebankApi.exec>> = useMemo(
    () => ({
      gameName: 'montebank',
      parseCommand: parseMonteBankCommand,
      formatResponse: formatMonteBankState,
      helpText: MONTEBANK_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBetPhase = phase === MonteBankPhase.BET;
  const isResultPhase = phase === MonteBankPhase.RESULT;
  const gameOver = !!state?.gameEndFlag;

  // **選んだ位置は 0 始まりでそのまま送る。** 0 は正当な値なので省略しない。
  const handleBet = useCallback(() => execApi('bet', { idx: selected, bet }), [execApi, selected, bet]);

  const actionBindings = useMemo(
    () => [{ key: 'n', action: () => execApi('next'), enabled: isResultPhase && !gameOver }],
    [execApi, isResultPhase, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('montebank', state);

  if (!state) return <GameSkeleton gameKey="montebank" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [MonteBankPhase.BET]: t('phase.bet'),
      [MonteBankPhase.RESULT]: t('phase.result'),
      [MonteBankPhase.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const net = state.payout - state.bet;

  return (
    <GamePageShell
      title={tc('nav.montebank')}
      gameThemeBg={gameTheme.montebank.bg}
      phaseName={phaseName}
      gamePath="/montebank"
      gameEndFlag={gameOver}
      winShow={isResultPhase && state.result === MONTE_BANK_RESULT.win}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="mb-chips">
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
          <div data-testid="card-area" className={`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <div className="text-ds-text-primary text-center text-sm mb-1" data-testid="mb-round-line">
              {t('label.round')}: {state.roundNumber}
              {' · '}
              {t('label.remaining')}: {state.remainingCards}
              {' · '}
              {t('label.payout', { mult: state.payoutMultiplier })}
            </div>
            <p className="text-ds-text-muted text-center text-xs mb-2">{t('suitNotice')}</p>

            <div className="flex justify-center gap-2 flex-wrap mb-3" data-tutorial="mb-layout">
              {state.layout.map((entry, i) => (
                <button
                  key={`layout-${entry.card.design}-${entry.card.value}-${i}`}
                  type="button"
                  data-testid={`mb-layout-${i}`}
                  aria-pressed={isBetPhase && selected === i}
                  disabled={!isBetPhase || loading}
                  onClick={() => setSelected(i)}
                  className={`flex flex-col items-center rounded px-1 py-1 ${
                    entry.isPicked || (isBetPhase && selected === i) ? 'ring-2 ring-ds-success' : ''
                  }`}
                >
                  <AnimatedCard card={entry.card} width={cardWidth} />
                  {/* **賭けの良し悪しはサーバの値をそのまま出す。** 数え直さない。 */}
                  <span
                    data-testid={`mb-note-${i}`}
                    className={`text-xs mt-1 ${entry.isEven ? 'text-ds-success' : 'text-ds-error'}`}
                  >
                    {entry.isEven ? t('label.even') : t('label.against')}
                  </span>
                  <span className="text-ds-text-muted text-xs" data-testid={`mb-count-${i}`}>
                    {t('label.suitCount', { count: entry.suitCount })}
                  </span>
                  {/* **賭けの良し悪しは場の枚数と山の残りの両方で決まる** (#5779)。
                      山残りはサーバが計算済みで、型にもあるのに出していなかった。 */}
                  <span className="text-ds-text-muted text-xs" data-testid={`mb-remaining-${i}`}>
                    {t('label.remainingOfSuit', { count: entry.remainingOfSuit })}
                  </span>
                </button>
              ))}
            </div>

            {state.gate && (
              <div className="mb-3" data-testid="mb-gate">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.gate')}</div>
                <div className="flex justify-center">
                  <AnimatedCard card={state.gate} width={cardWidth} />
                </div>
              </div>
            )}

            {isResultPhase && (
              <div className="text-center mb-2" data-testid="mb-result">
                <div className={`text-sm font-medium ${net >= 0 ? 'text-ds-success' : 'text-ds-error'}`}>
                  {t(`result.${resultKeyOf(state.result)}`)} · {t('label.net')} {net}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.montebank.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2">
              {isBetPhase && !gameOver && (
                <div className="flex flex-col items-center gap-2" data-tutorial="mb-bet">
                  <p className="text-ds-text-muted text-sm">{t('betGuide')}</p>
                  <ChipBetInput
                    id="montebank-bet"
                    label={t('label.bet')}
                    value={bet}
                    onChange={setBet}
                    max={state.chips}
                  />
                  <button
                    type="button"
                    className={btnPrimary}
                    data-hint-action="bet"
                    onClick={handleBet}
                    disabled={loading}
                  >
                    {t('button.bet')}
                  </button>
                </div>
              )}

              {isResultPhase && !gameOver && (
                <button type="button" className={btnPrimary} onClick={() => execApi('next')} disabled={loading}>
                  {t('button.next')}
                </button>
              )}

              <div className="flex gap-2">
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('button.actionLog')}
                </button>
                <GameResetButton
                  isGameEnd={gameOver}
                  onReset={() => execApi('reset')}
                  requestConfirm={requestConfirm}
                  loading={loading}
                />
              </div>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
