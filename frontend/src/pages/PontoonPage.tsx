import { useCallback, useMemo, useState } from 'react';
import type { pontoonApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CardBack } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePontoonGame } from '../hooks/usePontoonGame';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PontoonHand, PontoonResponse } from '../types/card';
import { PontoonPhase, PontoonRank } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { PONTOON_HELP, parsePontoonCommand } from '../utils/cli/commands/pontoonCommands';
import { formatPontoonState } from '../utils/cli/formatters/pontoonFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { PONTOON_MIN_BET, pontoonBuyChoices, pontoonClampBuy, pontoonMaxBuy } from '../utils/pontoonBet';

const BET_OPTIONS = [10, 50, 100, 500];

const PT_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="pt-banker"]', messageKey: 'tutorial.faceDown', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pt-seats"]', messageKey: 'tutorial.rank', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="pt-controls"]', messageKey: 'tutorial.stick', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="pt-controls"]', messageKey: 'tutorial.buy', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="pt-banker"]', messageKey: 'tutorial.banker', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pt-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Pontoon page: face-down hands, three seats, and a rotating bank. */
export const PontoonPage = withTutorial(PontoonPageContent, 'pontoon', PT_TUTORIAL_STEPS);

/** Label for a hand's rank. A points hand ranks by total and needs no label. */
function rankLabelKey(rank: number): string | null {
  if (rank === PontoonRank.PONTOON) return 'rank.pontoon';
  if (rank === PontoonRank.FIVE_CARD) return 'rank.fiveCard';
  if (rank === PontoonRank.BUST) return 'rank.bust';
  return null;
}

function PontoonPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pontoon');
  const game = usePontoonGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pontoon');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('pontoon', state);
  const cliConfig: CliGameConfig<PontoonResponse, Parameters<typeof pontoonApi.exec>> = useMemo(
    () => ({
      gameName: 'pontoon',
      parseCommand: parsePontoonCommand,
      formatResponse: formatPontoonState,
      helpText: PONTOON_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();

  const isPlayerTurn = state?.phase === PontoonPhase.PLAYER_TURN;

  // Pontoon.Buy accepts anything from PontoonMinBet up to twice the current
  // stake, and the CUI's `buy <amount>` has always exposed that. The web page
  // sent the minimum every time, so the choice existed only in the CUI (#4878).
  const activeBet = state?.seats[0]?.hands[state.activeHand]?.bet ?? PONTOON_MIN_BET;
  // Null means "follow the stake", which is what the page always sent before, so
  // the default action is unchanged and the range is purely additive. A chosen
  // stake persists across hands (as bet controls elsewhere do) but is clamped,
  // so it can never become illegal when the next hand's stake differs.
  const [buyAmount, setBuyAmount] = useState<number | null>(null);
  const clampedBuy = pontoonClampBuy(buyAmount, activeBet);

  const handleBuyOnce = useCallback(() => {
    game.handleBuy(clampedBuy);
  }, [game, clampedBuy]);

  const actionBindings = useMemo(
    () => [
      { key: 's', action: game.handleStick, label: 'stick' },
      { key: 't', action: game.handleTwist, label: 'twist' },
      { key: 'b', action: handleBuyOnce, label: 'buy' },
      { key: 'p', action: game.handleSplit, label: 'split' },
    ],
    [game, handleBuyOnce],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayerTurn && !loading });

  if (!state) {
    return <GameSkeleton gameKey="pontoon" layout={{ kind: 'tableau', topRow: 2, tableau: 3 }} />;
  }

  const ended = state.phase === PontoonPhase.END;
  const isBetting = state.phase === PontoonPhase.BET;
  const isBankerTurn = state.phase === PontoonPhase.BANKER_TURN;
  const bankerName = state.isHumanBanker ? t('bankerIsYou') : (state.seats[state.bankerIdx]?.name ?? '');

  /**
   * Render one hand. A face-down hand shows card backs and no total: the
   * banker's hand is the one you cannot read, and leaking it here would defeat
   * the game.
   */
  const renderHand = (hand: PontoonHand, label: string, keyPrefix: string) => {
    // The server decides what may be seen and says so with `hidden`; the page
    // never infers it, so there is one place that can be wrong instead of two.
    const hide = hand.hidden;
    const rankKey = rankLabelKey(hand.rank);
    return (
      <div key={keyPrefix} className="text-center">
        <div
          className="flex gap-1 justify-center"
          role="img"
          aria-label={hide ? label : t('seatAriaLabel', { name: label, total: hand.total })}
        >
          {hand.cards.map((card, i) =>
            hide || !card ? (
              <CardBack key={`${keyPrefix}-c${i.toString()}`} width={cardWidth} />
            ) : (
              <AnimatedCard key={`${keyPrefix}-c${i.toString()}`} card={card} width={cardWidth} draggable={false} />
            ),
          )}
        </div>
        {!hide && (
          <div className="text-game-text-muted text-xs mt-1">
            {t('total')}: {hand.total}
            {rankKey && <span className="ml-1 text-ds-warning font-bold">{t(rankKey)}</span>}
          </div>
        )}
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.pontoon')}
      gameThemeBg={gameTheme.pontoon.bg}
      phaseName={
        ended
          ? t('phase.end')
          : isBetting
            ? t('phase.bet')
            : isBankerTurn
              ? t('phase.bankerTurn')
              : t('phase.playerTurn')
      }
      gamePath="/pontoon"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('chips')}: {state.chips}
          </span>
          <span className="text-sm text-ds-text-muted">
            {t('banker')}: {bankerName}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="text-center mb-4" data-tutorial="pt-banker">
              <div className="text-game-text-muted text-xs mb-1">{t('bankerHand')}</div>
              {state.bankerHand ? (
                renderHand(
                  state.bankerHand,
                  state.bankerHand.hidden
                    ? t('hiddenBankerHandAriaLabel')
                    : t('bankerHandAriaLabel', { total: state.bankerHand.total }),
                  'banker',
                )
              ) : (
                <div className="text-game-text-muted text-sm">—</div>
              )}
            </div>

            <div className="flex flex-wrap justify-center gap-4 sm:gap-8" data-tutorial="pt-seats">
              {state.seats.map((seat, seatIdx) => {
                if (seatIdx === state.bankerIdx) return null;
                return (
                  <div key={`seat-${seatIdx.toString()}`} className="text-center">
                    <div className="text-game-text-muted text-xs mb-1">{seat.name}</div>
                    <div className="flex flex-col gap-2">
                      {seat.hands.map((hand, handIdx) => {
                        const onTurn = isPlayerTurn && seatIdx === state.activeSeat && handIdx === state.activeHand;
                        return (
                          <div
                            key={`s${seatIdx.toString()}-h${handIdx.toString()}`}
                            className={onTurn ? 'ring-2 ring-ds-warning rounded p-1' : 'p-1'}
                          >
                            {renderHand(
                              hand,
                              hand.hidden ? t('hiddenHandAriaLabel', { name: seat.name }) : seat.name,
                              `s${seatIdx.toString()}h${handIdx.toString()}`,
                            )}
                            <div className="text-game-text-muted text-xs mt-1">
                              {t('bet')}: {hand.bet}
                              {ended && hand.payout !== 0 && (
                                <span className={hand.payout > 0 ? ' text-ds-success' : ' text-ds-danger'}>
                                  {' '}
                                  {hand.payout > 0 ? `+${hand.payout}` : hand.payout}
                                </span>
                              )}
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={ended}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.pontoon.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap" data-tutorial="pt-controls">
              {isBetting && !state.isHumanBanker && (
                <>
                  <span className="text-sm text-ds-text-muted">{t('betLabel')}</span>
                  {BET_OPTIONS.map((amount) => (
                    <button
                      key={`bet-${amount.toString()}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => game.handleBet(amount)}
                      disabled={loading || amount > state.chips}
                    >
                      {t('betAmount', { amount })}
                    </button>
                  ))}
                </>
              )}

              {isBetting && state.isHumanBanker && (
                <button type="button" className={btnPrimary} onClick={game.handleDeal} disabled={loading}>
                  {t('actions.deal')}
                </button>
              )}

              <label className="flex items-center gap-1 text-ds-text-primary text-xs w-full justify-center cursor-pointer min-h-[44px]">
                <input
                  type="checkbox"
                  checked={frontendHintEnabled}
                  onChange={(e) => setFrontendHintEnabled(e.target.checked)}
                />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

              {/* Only the legal declarations are rendered. The server decides,
                  so the stick minimum and the no-buy-after-twist rule live in
                  one place rather than two. */}
              {/* **相手がどこで止まるかは判断材料の半分。**自分が 15 未満で
                  スティックできないことは出ていたのに、相手の停止ラインは
                  どこにも書かれていなかった (#5565)。
                  親と CPU 席は**別の規則**で止まる (親は 17、席は 15) ので
                  両方言う。数字はサーバがドメイン定数から渡すので、訳文にも
                  画面にも焼き込まない。 */}
              {isPlayerTurn && (
                <span className="text-xs text-ds-text-muted" data-testid="pontoon-thresholds">
                  {t('hint.cpuStick', { cpuMin: state.cpuStickMin, min: state.stickMin })}
                </span>
              )}
              {isPlayerTurn && state.canStick && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={game.handleStick}
                  disabled={loading}
                  title={t('hint.stick', { min: state.stickMin })}
                >
                  {t('actions.stick')}
                </button>
              )}
              {isPlayerTurn && state.canTwist && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={game.handleTwist}
                  disabled={loading}
                  title={t('hint.twist')}
                >
                  {t('actions.twist')}
                </button>
              )}
              {isPlayerTurn && state.canBuy && (
                <label className="flex items-center gap-1 text-xs text-ds-text-muted">
                  <span>{t('actions.buyAmount')}</span>
                  <select
                    className="rounded bg-ds-surface-elevated text-ds-text-primary px-1 py-0.5 min-h-11"
                    value={clampedBuy}
                    onChange={(e) => setBuyAmount(Number(e.target.value))}
                    data-testid="pontoon-buy-amount"
                    aria-label={t('actions.buyRange', { min: PONTOON_MIN_BET, max: pontoonMaxBuy(activeBet) })}
                  >
                    {pontoonBuyChoices(activeBet).map((v) => (
                      <option key={v} value={v}>
                        {v}
                      </option>
                    ))}
                  </select>
                </label>
              )}
              {isPlayerTurn && state.canBuy && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleBuyOnce}
                  disabled={loading}
                  title={t('hint.buy')}
                >
                  {t('actions.buy')}
                </button>
              )}
              {isPlayerTurn && state.canSplit && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={game.handleSplit}
                  disabled={loading}
                  title={t('hint.split')}
                >
                  {t('actions.split')}
                </button>
              )}

              {isBankerTurn && (
                <>
                  <button type="button" className={btnPrimary} onClick={game.handleBankerTwist} disabled={loading}>
                    {t('actions.bankerTwist')}
                  </button>
                  <button type="button" className={btnDanger} onClick={game.handleBankerStay} disabled={loading}>
                    {t('actions.bankerStay')}
                  </button>
                </>
              )}

              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="pt-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="pontoon-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
