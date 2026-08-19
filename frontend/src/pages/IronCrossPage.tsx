import { useCallback, useEffect, useMemo, useState } from 'react';
import { ironcrossApi } from '../api/gameApi';
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
import type { Card, IronCrossResponse } from '../types/card';
import { IronCrossPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { IRONCROSS_CLI_HELP, parseIronCrossCommand } from '../utils/cli/commands/ironcrossCommands';
import { formatIronCrossState } from '../utils/cli/formatters/ironcrossFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const IC_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ic-cross"]', messageKey: 'tutorial.cross', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ic-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ic-actions"]', messageKey: 'tutorial.choose', placement: 'top', advanceOn: 'next' },
];

/**
 * Cross positions, matching the Go domain's layout.
 *
 * **These are indexes into `cross`, not display order.** The wire keeps the
 * array positional and sends `null` for a face-down slot precisely so the page
 * can lay the cross out; renumbering here would silently swap the arms.
 */
const CENTER = 0;
const TOP = 1;
const BOTTOM = 2;
const LEFT = 3;
const RIGHT = 4;

/** Renders the Iron Cross game page (#5267). */
export const IronCrossPage = withTutorial(IronCrossPageContent, 'ironcross', IC_TUTORIAL_STEPS);

function IronCrossPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ironcross');

  const [amount, setAmount] = useState(20);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(ironcrossApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('ironcross');
  const cliConfig: CliGameConfig<IronCrossResponse, Parameters<typeof ironcrossApi.exec>> = useMemo(
    () => ({
      gameName: 'ironcross',
      parseCommand: parseIronCrossCommand,
      formatResponse: formatIronCrossState,
      helpText: IRONCROSS_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBetting = phase === IronCrossPhase.BETTING;
  const isShowdown = phase === IronCrossPhase.SHOWDOWN;
  const gameOver = !!state?.gameEndFlag;
  const canAct = !!state?.isHumanTurn && isBetting;
  const facingBet = (state?.toCall ?? 0) > 0;
  // **選ぶ場面かどうかはサーバが決める。** フェーズ番号から割り出さない。
  const isChoosing = !!state?.isChoosing;
  // **後戻りできない一発勝負の選択** (#5781)。どの 3 枚が使われるかを、
  // 押す前に十字の上で示す。位置はサーバの verticalIndexes / horizontalIndexes
  // をそのまま使う——ページで並びを決め直すと、腕を取り違えても誰も気づかない。
  const [previewLine, setPreviewLine] = useState<'vertical' | 'horizontal' | null>(null);

  // **列を選ぶ場面が終わったら畳む。** クリックでボタンが消えるとき、環境に
  // よっては mouseleave も blur も飛ばない (Safari はクリックでフォーカスしない)。
  // 残したままだと、次の手で誰も触っていないのに光ったままになる。
  useEffect(() => {
    if (!isChoosing) setPreviewLine(null);
  }, [isChoosing]);

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
  } = useGameHint('ironcross', state);

  if (!state) return <GameSkeleton gameKey="ironcross" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [IronCrossPhase.BETTING]: t('phase.betting'),
      [IronCrossPhase.CHOOSE_LINE]: t('phase.chooseLine'),
      [IronCrossPhase.SHOWDOWN]: t('phase.showdown'),
      [IronCrossPhase.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const human = state.seats[state.humanSeat];
  const humanWon = gameOver && state.winnerSeat === state.humanSeat;

  /** Renders one slot of the cross, or a placeholder while it is face down. */
  const previewIndexes =
    previewLine === 'vertical' ? state.verticalIndexes : previewLine === 'horizontal' ? state.horizontalIndexes : [];

  const crossSlot = (index: number) => {
    const card: Card | null = state.cross[index] ?? null;
    const previewed = previewIndexes.includes(index);
    const ring = previewed ? ' ring-2 ring-ds-success' : '';
    if (!card) {
      return (
        <div
          className={`rounded border-2 border-dashed border-ds-border${ring}`}
          style={{ width: cardWidth, height: Math.round(cardWidth * 1.4) }}
          data-testid={`ic-cross-${index}`}
          data-previewed={previewed ? 'true' : undefined}
          role="img"
          aria-label={t('label.hidden')}
        />
      );
    }
    return (
      <div className={ring.trim()} data-testid={`ic-cross-${index}`} data-previewed={previewed ? 'true' : undefined}>
        <AnimatedCard card={card} width={cardWidth} />
      </div>
    );
  };

  const lineLabel = (line: number) => (line === 1 ? t('label.vertical') : line === 2 ? t('label.horizontal') : '');

  return (
    <GamePageShell
      title={tc('nav.ironcross')}
      gameThemeBg={gameTheme.ironcross.bg}
      phaseName={phaseName}
      gamePath="/ironcross"
      gameEndFlag={gameOver}
      isHumanTurn={state.isHumanTurn}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="ic-chips">
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

            <div className="text-ds-text-primary text-center text-sm mb-1" data-testid="ic-hand-line">
              {t('label.hand')}: {state.handNumber}
              {' · '}
              {t('label.pot')}: {state.pot}
            </div>

            {/*
              **十字は十字の形で並べる。** 1 行に詰めると、どの札が縦でどの札が
              横なのかが分からなくなり、このゲームの唯一の判断が成り立たない。
            */}
            <div className="mb-3" data-tutorial="ic-cross">
              <div className="text-ds-text-primary text-center text-sm font-bold mb-1">
                {t('label.cross')}
                {' · '}
                <span data-testid="ic-revealed">
                  {t('label.revealed', { revealed: state.revealedCount, total: state.crossTotal })}
                </span>
              </div>
              <div className="flex flex-col items-center gap-1" data-testid="ic-cross">
                {crossSlot(TOP)}
                <div className="flex gap-1">
                  {crossSlot(LEFT)}
                  {crossSlot(CENTER)}
                  {crossSlot(RIGHT)}
                </div>
                {crossSlot(BOTTOM)}
              </div>
              <p className="text-ds-text-muted text-center text-xs mt-1">{t('crossNotice')}</p>
            </div>

            <div className="mb-3" data-tutorial="ic-hand">
              <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.yourHand')}</div>
              <div className="flex justify-center gap-1 flex-wrap" data-testid="ic-hand">
                {(human?.cards ?? []).map((card, i) => (
                  <AnimatedCard key={`hand-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                ))}
              </div>
            </div>

            <div>
              {state.seats.map((seat, i) => (
                <div
                  key={`seat-${seat.name}-${i}`}
                  data-testid={`ic-seat-${i}`}
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
                    {/* **選んだ列はサーバが送ってきたときだけ出す。** CPU の分は
                        ショーダウンまで 0 で届くので、勝手に推測しない。 */}
                    {seat.line !== 0 && (
                      <span data-testid={`ic-line-${i}`}>
                        {' · '}
                        {t('label.usedLine', { line: lineLabel(seat.line) })}
                      </span>
                    )}
                    {seat.wonAmount > 0 && (
                      <span data-testid={`ic-won-${i}`}> · {t('label.won', { amount: seat.wonAmount })}</span>
                    )}
                  </span>
                  {/* **CPU の手札はサーバが送っていない。** 届いていれば開く。 */}
                  {!seat.isHuman && (
                    <div className="flex justify-center gap-1 flex-wrap mt-1" data-testid={`ic-seat-cards-${i}`}>
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
              <div className="text-ds-text-primary text-center text-sm font-bold mt-2" data-testid="ic-winner">
                {t('label.winner')}: {state.seats[state.winnerSeat]?.name ?? '?'}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.ironcross.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="ic-actions">
              {/*
                **列を選ぶボタンは、選ぶ場面でだけ出す。** ベットの手と同じ列に
                並べると、一度きりで取り返しのつかない選択が打ち間違いで潰れる。
              */}
              {isChoosing && (
                <>
                  <p className="text-ds-text-muted text-sm" data-testid="ic-choose-guide">
                    {t('label.chooseLine')}
                  </p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    <button
                      type="button"
                      className={btnSuccess}
                      data-testid="ic-vertical"
                      data-hint-action="line"
                      onClick={() => execApi('vertical')}
                      disabled={loading}
                      onMouseEnter={() => setPreviewLine('vertical')}
                      onMouseLeave={() => setPreviewLine(null)}
                      onFocus={() => setPreviewLine('vertical')}
                      onBlur={() => setPreviewLine(null)}
                    >
                      {t('button.vertical')}
                    </button>
                    <button
                      type="button"
                      className={btnSuccess}
                      data-testid="ic-horizontal"
                      data-hint-action="line"
                      onClick={() => execApi('horizontal')}
                      disabled={loading}
                      onMouseEnter={() => setPreviewLine('horizontal')}
                      onMouseLeave={() => setPreviewLine(null)}
                      onFocus={() => setPreviewLine('horizontal')}
                      onBlur={() => setPreviewLine(null)}
                    >
                      {t('button.horizontal')}
                    </button>
                  </div>
                </>
              )}

              {canAct && (
                <>
                  <p className="text-ds-text-muted text-sm" data-testid="ic-bet-guide">
                    {facingBet ? t('label.toCall', { amount: state.toCall }) : t('label.canCheck')}
                  </p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    {/* **チェックとコールは場況で入れ替わる。** サーバの toCall に従う。 */}
                    {facingBet ? (
                      <button
                        type="button"
                        className={btnSuccess}
                        data-testid="ic-call"
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
                        data-testid="ic-check"
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
                          data-testid="ic-raise"
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
                        data-testid="ic-bet"
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
                      data-testid="ic-fold"
                      data-hint-action="fold"
                      onClick={() => execApi('fold')}
                      disabled={loading}
                    >
                      {t('button.fold')}
                    </button>
                  </div>
                  <ChipBetInput
                    id="ironcross-amount"
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
