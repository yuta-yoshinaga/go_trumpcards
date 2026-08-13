import { useCallback, useMemo, useState } from 'react';
import { cincinnatiApi } from '../api/gameApi';
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
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CincinnatiResponse } from '../types/card';
import { CincinnatiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CINCINNATI_CLI_HELP, parseCincinnatiCommand } from '../utils/cli/commands/cincinnatiCommands';
import { formatCincinnatiState } from '../utils/cli/formatters/cincinnatiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const CIN_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="cin-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="cin-board"]', messageKey: 'tutorial.community', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cin-actions"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
];

/** Renders the Cincinnati game page (#5266). */
export const CincinnatiPage = withTutorial(CincinnatiPageContent, 'cincinnati', CIN_TUTORIAL_STEPS);

function CincinnatiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cincinnati');

  const [amount, setAmount] = useState(20);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(cincinnatiApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cincinnati');
  const cliConfig: CliGameConfig<CincinnatiResponse, Parameters<typeof cincinnatiApi.exec>> = useMemo(
    () => ({
      gameName: 'cincinnati',
      parseCommand: parseCincinnatiCommand,
      formatResponse: formatCincinnatiState,
      helpText: CINCINNATI_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBetting = phase === CincinnatiPhase.BETTING;
  const isShowdown = phase === CincinnatiPhase.SHOWDOWN;
  const gameOver = !!state?.gameEndFlag;
  const canAct = !!state?.isHumanTurn && isBetting;
  const facingBet = (state?.toCall ?? 0) > 0;

  const handleBet = useCallback(() => execApi('bet', { amount }), [execApi, amount]);
  const handleRaise = useCallback(() => execApi('raise', { amount }), [execApi, amount]);

  const actionBindings = useMemo(
    () => [
      { key: 'k', action: () => execApi('check'), enabled: canAct && !facingBet },
      { key: 'c', action: () => execApi('call'), enabled: canAct && facingBet },
      { key: 'n', action: () => execApi('next'), enabled: isShowdown && !gameOver },
    ],
    [execApi, canAct, facingBet, isShowdown, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('cincinnati', state);

  if (!state) return <GameSkeleton gameKey="cincinnati" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [CincinnatiPhase.DEAL]: t('phase.deal'),
      [CincinnatiPhase.BETTING]: t('phase.betting'),
      [CincinnatiPhase.SHOWDOWN]: t('phase.showdown'),
      [CincinnatiPhase.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const human = state.seats[state.humanSeat];
  const humanWon = gameOver && state.winnerSeat === state.humanSeat;

  return (
    <GamePageShell
      title={tc('nav.cincinnati')}
      gameThemeBg={gameTheme.cincinnati.bg}
      phaseName={phaseName}
      gamePath="/cincinnati"
      gameEndFlag={gameOver}
      isHumanTurn={state.isHumanTurn}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="cin-chips">
            {t('label.chips')}: {human?.chips ?? 0}
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

            <div className="text-ds-text-primary text-center text-sm mb-1" data-testid="cin-hand-line">
              {t('label.hand')}: {state.handNumber}
              {' · '}
              {t('label.pot')}: {state.pot}
            </div>

            {/* **あと何枚めくれるかを必ず出す。** 残りの回数だけベットがある。 */}
            <div className="mb-3" data-tutorial="cin-board">
              <div className="text-ds-text-primary text-center text-sm font-bold mb-1">
                {t('label.community')}
                {' · '}
                <span data-testid="cin-revealed">
                  {t('label.revealed', { revealed: state.revealedCount, total: state.communityTotal })}
                </span>
              </div>
              <div className="flex justify-center gap-1 flex-wrap" data-testid="cin-board">
                {state.community.map((card, i) => (
                  <AnimatedCard key={`board-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                ))}
              </div>
            </div>

            {/* **手札 5 枚はこのゲームの本体。** 場を使わない役が普通にある。 */}
            <div className="mb-3" data-tutorial="cin-hand">
              <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.yourHand')}</div>
              <div className="flex justify-center gap-1 flex-wrap" data-testid="cin-hand">
                {(human?.cards ?? []).map((card, i) => (
                  <AnimatedCard key={`hand-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                ))}
              </div>
              <p className="text-ds-text-muted text-center text-xs mt-1">{t('fiveCardNotice')}</p>
            </div>

            <div>
              {state.seats.map((seat, i) => (
                <div
                  key={`seat-${seat.name}-${i}`}
                  data-testid={`cin-seat-${i}`}
                  className={`mb-1 rounded px-2 py-1 text-center text-sm ${
                    seat.isTurn ? 'ring-2 ring-ds-success' : ''
                  } ${seat.folded ? 'opacity-50' : ''}`}
                >
                  <span className="text-ds-text-primary">
                    {seat.name}
                    {seat.folded && ` · ${t('label.folded')}`}
                    {seat.allIn && ` · ${t('label.allIn')}`}
                    {' · '}
                    {t('label.chips')} {seat.chips}
                    {seat.bet > 0 && ` · ${t('label.bet')} ${seat.bet}`}
                    {seat.wonAmount > 0 && (
                      <span data-testid={`cin-won-${i}`}> · {t('label.won', { amount: seat.wonAmount })}</span>
                    )}
                  </span>
                  {/* **CPU の手札はサーバが送っていない。** 届いていれば開く。 */}
                  {!seat.isHuman && (
                    <div className="flex justify-center gap-1 flex-wrap mt-1" data-testid={`cin-seat-cards-${i}`}>
                      {seat.cards.length > 0 ? (
                        seat.cards.map((card, k) => (
                          <AnimatedCard
                            key={`s${i}-${card.design}-${card.value}-${k}`}
                            card={card}
                            width={Math.round(cardWidth * 0.7)}
                          />
                        ))
                      ) : (
                        <span className="text-ds-text-muted text-xs">{t('label.hidden')}</span>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {gameOver && (
              <div className="text-ds-text-primary text-center text-sm font-bold mt-2" data-testid="cin-winner">
                {t('label.winner')}: {state.seats[state.winnerSeat]?.name ?? '?'}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.cincinnati.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="cin-actions">
              {canAct && (
                <>
                  <p className="text-ds-text-muted text-sm" data-testid="cin-bet-guide">
                    {facingBet ? t('label.toCall', { amount: state.toCall }) : t('label.canCheck')}
                  </p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    {/* **チェックとコールは場況で入れ替わる。** サーバの toCall に従う。 */}
                    {facingBet ? (
                      <button
                        type="button"
                        className={btnSuccess}
                        data-testid="cin-call"
                        data-hint-action="call"
                        onClick={() => execApi('call')}
                        disabled={loading}
                      >
                        {t('button.call')}
                      </button>
                    ) : (
                      <button
                        type="button"
                        className={btnSuccess}
                        data-testid="cin-check"
                        data-hint-action="check"
                        onClick={() => execApi('check')}
                        disabled={loading}
                      >
                        {t('button.check')}
                      </button>
                    )}
                    {/* **レイズの可否はサーバが決める。** 上限に達したら出さない。 */}
                    {facingBet ? (
                      state.canRaise && (
                        <button
                          type="button"
                          className={btnWarning}
                          data-testid="cin-raise"
                          data-hint-action="raise"
                          onClick={handleRaise}
                          disabled={loading}
                        >
                          {t('button.raise')}
                        </button>
                      )
                    ) : (
                      <button
                        type="button"
                        className={btnWarning}
                        data-testid="cin-bet"
                        data-hint-action="bet"
                        onClick={handleBet}
                        disabled={loading}
                      >
                        {t('button.bet')}
                      </button>
                    )}
                    <button
                      type="button"
                      className={btnSecondary}
                      data-testid="cin-fold"
                      data-hint-action="fold"
                      onClick={() => execApi('fold')}
                      disabled={loading}
                    >
                      {t('button.fold')}
                    </button>
                  </div>
                  <ChipBetInput
                    id="cincinnati-amount"
                    label={t('label.bet')}
                    value={amount}
                    onChange={setAmount}
                    max={human?.chips ?? 0}
                  />
                </>
              )}

              {isShowdown && !gameOver && (
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
