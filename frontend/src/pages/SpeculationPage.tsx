import { useCallback, useEffect, useMemo, useState } from 'react';
import { speculationApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CardBack } from '../components/CardImage';
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
import type { SpeculationResponse } from '../types/card';
import { SPECULATION_HUMAN_SEAT, SPECULATION_NO_SEAT } from '../types/games/speculation';
import { SpeculationPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSpeculationCommand, SPECULATION_CLI_HELP } from '../utils/cli/commands/speculationCommands';
import { formatSpeculationState } from '../utils/cli/formatters/speculationFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const SP_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sp-seats"]', messageKey: 'tutorial.hidden', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sp-trump"]', messageKey: 'tutorial.trump', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sp-actions"]', messageKey: 'tutorial.auction', placement: 'top', advanceOn: 'next' },
];

/** Renders the Speculation game page. */
export const SpeculationPage = withTutorial(SpeculationPageContent, 'speculation', SP_TUTORIAL_STEPS);

/**
 * Suit i18n key by the Go `CardDesign*` value (Card.go: ♠1 ♣2 ♥3 ♦4).
 *
 * **There is no entry for 0 or -1 on purpose.** `trumpSuit` is -1 until the
 * trump card is turned, and design 0 is the joker; either mapped onto the first
 * suit would announce spades as trumps for a round that has none.
 */
const SUIT_KEYS: Readonly<Record<number, string>> = {
  1: 'suit.spade',
  2: 'suit.clover',
  3: 'suit.heart',
  4: 'suit.diamond',
};

/** Card backs are only a count — never the cards themselves. Cap the row so a big hand still fits. */
const MAX_BACKS_SHOWN = 3;

function SpeculationPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('speculation');

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(speculationApi.exec);
  const [raise, setRaise] = useState(0);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('speculation');
  const cliConfig: CliGameConfig<SpeculationResponse, Parameters<typeof speculationApi.exec>> = useMemo(
    () => ({
      gameName: 'speculation',
      parseCommand: parseSpeculationCommand,
      formatResponse: formatSpeculationState,
      helpText: SPECULATION_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isFlipPhase = phase === SpeculationPhase.FLIP;
  const isAuctionPhase = phase === SpeculationPhase.AUCTION;
  const isResultPhase = phase === SpeculationPhase.RESULT;
  const gameOver = !!state?.gameEndFlag;
  const offerAmount = state?.offerAmount ?? 0;

  // **上乗せは提示額を超えていなければならない。** ドメインは `amount <= offerAmount`
  // を弾くので、既定値を提示額そのものにすると押しても必ず失敗する。
  const minRaise = offerAmount + 1;
  useEffect(() => {
    setRaise((prev) => (prev < minRaise ? minRaise : prev));
  }, [minRaise]);

  const handleRaise = useCallback(() => execApi('bid', { amount: raise }), [execApi, raise]);

  const actionBindings = useMemo(
    () => [
      { key: 'f', action: () => execApi('flip'), enabled: isFlipPhase && !gameOver },
      { key: 'n', action: () => execApi('next'), enabled: isResultPhase && !gameOver },
    ],
    [execApi, isFlipPhase, isResultPhase, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('speculation', state);

  if (!state) return <GameSkeleton gameKey="speculation" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [SpeculationPhase.FLIP]: t('phase.flip'),
      [SpeculationPhase.AUCTION]: t('phase.auction'),
      [SpeculationPhase.RESULT]: t('phase.result'),
      [SpeculationPhase.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const human = state.seats[SPECULATION_HUMAN_SEAT];
  const trumpKey = SUIT_KEYS[state.trumpSuit];
  // **人間が売り手か買い手かは offerTo で決まる。** 申し出の宛先が札の持ち主。
  const humanSells = state.offerTo === SPECULATION_HUMAN_SEAT;
  const humanBuys = state.offerFrom === SPECULATION_HUMAN_SEAT;
  const offerOpen = isAuctionPhase && state.offerFrom !== SPECULATION_NO_SEAT && state.offerTo !== SPECULATION_NO_SEAT;

  return (
    <GamePageShell
      title={tc('nav.speculation')}
      gameThemeBg={gameTheme.speculation.bg}
      phaseName={phaseName}
      gamePath="/speculation"
      gameEndFlag={gameOver}
      winShow={isResultPhase && state.winnerSeat === SPECULATION_HUMAN_SEAT}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="sp-chips">
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

            <div className="text-ds-text-primary text-center text-sm mb-1" data-testid="sp-round-line">
              {t('label.round')}: {state.roundNo + 1}
              {state.config ? ` / ${state.config.rounds}` : ''}
              {' · '}
              {t('label.pot')}: {state.pot}
            </div>

            <div className="flex flex-col items-center mb-3" data-tutorial="sp-trump">
              <span className="text-ds-text-primary text-sm font-bold" data-testid="sp-trump">
                {t('label.trump')}: {trumpKey ? t(trumpKey) : t('label.trumpUnset')}
              </span>
              {state.trumpCard && (
                <div className="mt-1 flex flex-col items-center" data-testid="sp-trump-card">
                  <AnimatedCard card={state.trumpCard} width={cardWidth} />
                  <span className="text-ds-text-muted text-xs mt-1">{t('label.trumpCard')}</span>
                </div>
              )}
            </div>

            <p className="text-ds-text-muted text-center text-xs mb-2">{t('hiddenNotice')}</p>

            <div className="flex justify-center gap-3 flex-wrap mb-3" data-tutorial="sp-seats">
              {state.seats.map((seat, i) => {
                // **「最高札の持ち主」は -1 で「いない」。** 0 を「いない」と
                // 読むと、まだ切り札が1枚も出ていない盤面で人間に印が付く。
                const holdsBest = i === state.bestSeat;
                const isTurn = i === state.turnSeat;
                return (
                  <div
                    key={`seat-${seat.name}-${i}`}
                    data-testid={`sp-seat-${i}`}
                    className={`flex flex-col items-center rounded px-2 py-2 ${
                      holdsBest ? 'ring-2 ring-ds-success' : isTurn ? 'ring-2 ring-ds-accent' : ''
                    }`}
                  >
                    <span className="text-ds-text-primary text-sm font-medium">
                      {i === SPECULATION_HUMAN_SEAT ? t('label.you') : seat.name}
                      {isTurn ? ` (${t('label.turn')})` : ''}
                    </span>
                    <span className="text-ds-text-muted text-xs" data-testid={`sp-seat-chips-${i}`}>
                      {t('label.chips')}: {seat.chips}
                    </span>
                    {/* **枚数だけ。** 伏せ札の中身はサーバも送らないし、出してもいけない。 */}
                    <span className="text-ds-text-muted text-xs" data-testid={`sp-hidden-${i}`}>
                      {t('label.faceDown', { count: seat.hiddenCount })}
                    </span>
                    <div className="flex gap-0.5 mt-1">
                      {Array.from({ length: Math.min(seat.hiddenCount, MAX_BACKS_SHOWN) }, (_, k) => (
                        <CardBack key={`back-${i}-${k}`} width={Math.round(cardWidth * 0.5)} />
                      ))}
                    </div>
                    {seat.best && (
                      <div className="mt-1 flex flex-col items-center" data-testid={`sp-best-${i}`}>
                        <AnimatedCard card={seat.best} width={Math.round(cardWidth * 0.7)} />
                        <span className="text-ds-success text-xs">{t('label.best')}</span>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>

            {state.bestSeat === SPECULATION_NO_SEAT && (
              <p className="text-ds-text-muted text-center text-xs mb-2" data-testid="sp-no-best">
                {t('label.noBest')}
              </p>
            )}

            {(isResultPhase || gameOver) && (
              <div className="text-center mb-2" data-testid="sp-result">
                <div className="text-ds-text-primary text-sm font-medium">
                  {state.winnerSeat === SPECULATION_NO_SEAT
                    ? t('result.void')
                    : state.winnerSeat === SPECULATION_HUMAN_SEAT
                      ? t('result.youWin')
                      : t('result.seatWins', { name: state.seats[state.winnerSeat]?.name ?? '' })}
                </div>
                {gameOver && (
                  <div className="text-ds-text-muted text-sm" data-testid="sp-final-chips">
                    {t('result.finalChips', { chips: human?.chips ?? 0 })}
                  </div>
                )}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.speculation.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="sp-actions">
              {isFlipPhase && !gameOver && (
                <div className="flex flex-col items-center gap-2">
                  <p className="text-ds-text-muted text-sm" data-testid="sp-turn-guide">
                    {state.turnSeat === SPECULATION_HUMAN_SEAT
                      ? t('guide.yourTurn')
                      : t('guide.cpuTurn', { name: state.seats[state.turnSeat]?.name ?? '' })}
                  </p>
                  <button
                    type="button"
                    className={btnPrimary}
                    data-hint-action="flip"
                    data-testid="sp-flip"
                    onClick={() => execApi('flip')}
                    disabled={loading}
                  >
                    {t('button.flip')}
                  </button>
                </div>
              )}

              {offerOpen && (
                <div className="flex flex-col items-center gap-2" data-testid="sp-offer">
                  <p className="text-ds-text-primary text-sm">
                    {humanSells
                      ? t('guide.sell', {
                          buyer: state.seats[state.offerFrom]?.name ?? '',
                          amount: state.offerAmount,
                        })
                      : t('guide.buy', {
                          owner: state.seats[state.offerTo]?.name ?? '',
                          amount: state.offerAmount,
                        })}
                  </p>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      className={btnPrimary}
                      data-hint-action="accept"
                      data-testid="sp-accept"
                      onClick={() => execApi('accept')}
                      disabled={loading}
                    >
                      {t('button.accept')}
                    </button>
                    <button
                      type="button"
                      className={btnSecondary}
                      data-hint-action="decline"
                      data-testid="sp-decline"
                      onClick={() => execApi('decline')}
                      disabled={loading}
                    >
                      {t('button.decline')}
                    </button>
                  </div>
                  {/* **上乗せできるのは人間が買い手のときだけ。** 売り手の側から
                      値を吊り上げる手は Bid には無く、サーバは弾く。 */}
                  {humanBuys && (
                    <div className="flex flex-col items-center gap-1" data-testid="sp-raise">
                      <ChipBetInput
                        id="speculation-raise"
                        label={t('label.raise')}
                        value={raise}
                        onChange={setRaise}
                        min={minRaise}
                        step={1}
                        max={human?.chips}
                      />
                      <span className="text-ds-text-muted text-xs">{t('guide.raiseHint', { min: minRaise })}</span>
                      <button
                        type="button"
                        className={btnSecondary}
                        data-testid="sp-bid"
                        onClick={handleRaise}
                        disabled={loading || raise < minRaise}
                      >
                        {t('button.raise')}
                      </button>
                    </div>
                  )}
                </div>
              )}

              {isResultPhase && !gameOver && (
                <button
                  type="button"
                  className={btnPrimary}
                  data-testid="sp-next"
                  onClick={() => execApi('next')}
                  disabled={loading}
                >
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
