import { useCallback, useMemo, useState } from 'react';
import { baseballpokerApi } from '../api/gameApi';
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
import type { BaseballPokerResponse, BaseballSeat } from '../types/card';
import { BaseballPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BASEBALLPOKER_CLI_HELP, parseBaseballPokerCommand } from '../utils/cli/commands/baseballpokerCommands';
import { formatBaseballPokerState } from '../utils/cli/formatters/baseballpokerFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const BB_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bb-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="bb-seats"]', messageKey: 'tutorial.seats', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="bb-actions"]', messageKey: 'tutorial.buy', placement: 'top', advanceOn: 'next' },
];

/** Renders the Baseball Poker game page (#5268). */
export const BaseballPokerPage = withTutorial(BaseballPokerPageContent, 'baseballpoker', BB_TUTORIAL_STEPS);

function BaseballPokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('baseballpoker');

  const [amount, setAmount] = useState(20);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(baseballpokerApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('baseballpoker');
  const cliConfig: CliGameConfig<BaseballPokerResponse, Parameters<typeof baseballpokerApi.exec>> = useMemo(
    () => ({
      gameName: 'baseballpoker',
      parseCommand: parseBaseballPokerCommand,
      formatResponse: formatBaseballPokerState,
      helpText: BASEBALLPOKER_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBetting = phase === BaseballPhase.BETTING;
  const isShowdown = phase === BaseballPhase.SHOWDOWN;
  const gameOver = !!state?.gameEndFlag;
  const canAct = !!state?.isHumanTurn && isBetting;
  const facingBet = (state?.toCall ?? 0) > 0;
  // **買い増しを迫られているかはサーバが決める。** フェーズ番号から割り出さない。
  const isBuying = !!state?.isBuying;

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
  } = useGameHint('baseballpoker', state);

  if (!state) return <GameSkeleton gameKey="baseballpoker" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [BaseballPhase.BETTING]: t('phase.betting'),
      [BaseballPhase.BUY_IN]: t('phase.buyIn'),
      [BaseballPhase.SHOWDOWN]: t('phase.showdown'),
      [BaseballPhase.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const human = state.seats[state.humanSeat];
  const humanWon = gameOver && state.winnerSeat === state.humanSeat;

  /**
   * Renders one seat's cards.
   *
   * **A card the viewer may not see arrives as `null`.** Face-up cards are on
   * the wire for every seat — they are the material stud is read from — so the
   * page draws whatever it was given and marks the gaps, rather than deciding
   * for itself what to hide.
   */
  const seatCards = (seat: BaseballSeat, testId: string, width: number) => (
    <div className="flex justify-center gap-1 flex-wrap" data-testid={testId}>
      {seat.cards.map((card, i) =>
        card ? (
          <div key={`${testId}-${i}-${card.design}-${card.value}`} className="relative">
            <AnimatedCard card={card} width={width} />
            {state.wildValues.includes(card.value) && (
              <span
                className="absolute -top-1 -right-1 rounded bg-ds-accent px-1 text-[10px] text-ds-text-inverse"
                data-testid={`${testId}-wild-${i}`}
              >
                {t('label.wild')}
              </span>
            )}
          </div>
        ) : (
          <div
            key={`${testId}-hidden-${i}`}
            className="rounded border-2 border-dashed border-ds-border"
            style={{ width, height: Math.round(width * 1.4) }}
            data-testid={`${testId}-hidden-${i}`}
            role="img"
            aria-label={t('label.hidden')}
          />
        ),
      )}
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.baseballpoker')}
      gameThemeBg={gameTheme.baseballpoker.bg}
      phaseName={phaseName}
      gamePath="/baseballpoker"
      gameEndFlag={gameOver}
      isHumanTurn={state.isHumanTurn}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="bb-chips">
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

            <div className="text-ds-text-primary text-center text-sm mb-1" data-testid="bb-hand-line">
              {t('label.hand')}: {state.handNumber}
              {' · '}
              {t('label.pot')}: {state.pot}
              {' · '}
              <span data-testid="bb-street">
                {t('label.street', { street: state.street, total: state.streetTotal })}
              </span>
            </div>

            {/* **ワイルドとイベントの札はサーバが教えてくれる。** 画面が 3 と 9 を
                持つと、役の判定と印が別々に育って食い違う。 */}
            <p className="text-ds-text-muted text-center text-xs mb-3" data-testid="bb-wild-notice">
              {t('wildNotice')}
            </p>

            <div className="mb-3" data-tutorial="bb-hand">
              <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.yourHand')}</div>
              {human && seatCards(human, 'bb-hand', cardWidth)}
            </div>

            <div data-tutorial="bb-seats">
              {state.seats.map((seat, i) => (
                <div
                  key={`seat-${seat.name}-${i}`}
                  data-testid={`bb-seat-${i}`}
                  className={`mb-1 rounded px-2 py-1 text-center text-sm ${
                    seat.isBuying ? 'ring-2 ring-ds-warning' : seat.isTurn ? 'ring-2 ring-ds-success' : ''
                  } ${seat.folded ? 'opacity-50' : ''}`}
                >
                  <span className="text-ds-text-primary">
                    {seat.name}
                    {seat.folded && ` · ${t('label.folded')}`}
                    {seat.allIn && ` · ${t('label.allIn')}`}
                    {' · '}
                    {t('label.chips')} {seat.chips}
                    {seat.bet > 0 && ` · ${t('label.bet')} ${seat.bet}`}
                    {seat.bonusCards > 0 && (
                      <span data-testid={`bb-bonus-${i}`}>
                        {' · '}
                        {t('label.bonus', { count: seat.bonusCards })}
                      </span>
                    )}
                    {seat.usedWild && <span data-testid={`bb-usedwild-${i}`}> · {t('label.usedWild')}</span>}
                    {seat.wonAmount > 0 && (
                      <span data-testid={`bb-won-${i}`}> · {t('label.won', { amount: seat.wonAmount })}</span>
                    )}
                  </span>
                  {/* **他人の表札も出す。** スタッドの読み合いはここが材料。 */}
                  {!seat.isHuman && (
                    <div className="mt-1">{seatCards(seat, `bb-seat-cards-${i}`, Math.round(cardWidth * 0.7))}</div>
                  )}
                </div>
              ))}
            </div>

            {gameOver && (
              <div className="text-ds-text-primary text-center text-sm font-bold mt-2" data-testid="bb-winner">
                {t('label.winner')}: {state.seats[state.winnerSeat]?.name ?? '?'}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.baseballpoker.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="bb-actions">
              {/*
                **買い増しの返事はベットの手と分ける。** 同じ列に並べると、
                その場でポットぶんを払う操作が打ち間違いで通る。
              */}
              {isBuying && (
                <>
                  <p className="text-ds-text-muted text-sm" data-testid="bb-buy-guide">
                    {t('label.buyIn', { amount: state.buyCost })}
                  </p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    <button
                      type="button"
                      className={btnWarning}
                      data-testid="bb-pay"
                      data-hint-action="pay"
                      onClick={() => execApi('pay')}
                      disabled={loading}
                    >
                      {t('button.pay')}
                    </button>
                    <button
                      type="button"
                      className={btnSecondary}
                      data-testid="bb-buyfold"
                      data-hint-action="fold"
                      onClick={() => execApi('buyfold')}
                      disabled={loading}
                    >
                      {t('button.buyfold')}
                    </button>
                  </div>
                </>
              )}

              {canAct && (
                <>
                  <p className="text-ds-text-muted text-sm" data-testid="bb-bet-guide">
                    {facingBet ? t('label.toCall', { amount: state.toCall }) : t('label.canCheck')}
                  </p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    {/* **チェックとコールは場況で入れ替わる。** サーバの toCall に従う。 */}
                    {facingBet ? (
                      <button
                        type="button"
                        className={btnSuccess}
                        data-testid="bb-call"
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
                        data-testid="bb-check"
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
                          data-testid="bb-raise"
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
                        data-testid="bb-bet"
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
                      data-testid="bb-fold"
                      data-hint-action="fold"
                      onClick={() => execApi('fold')}
                      disabled={loading}
                    >
                      {t('button.fold')}
                    </button>
                  </div>
                  <ChipBetInput
                    id="baseballpoker-amount"
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
