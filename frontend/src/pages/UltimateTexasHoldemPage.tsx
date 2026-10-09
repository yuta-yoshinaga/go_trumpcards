import { useEffect, useMemo, useState } from 'react';
import { ultimatetexasholdemApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { isMaskedCard } from '../types/card';
import { UltimateTexasHoldemPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { utHoldemBetBounds } from '../utils/utHoldemBet';
import { utHoldemPreflopStrength } from '../utils/utHoldemPreflop';

/** Ultimate Texas Hold'em tutorial step definitions. */
const UTH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="uth-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="uth-pre-flop-buttons"]',
    messageKey: 'tutorial.preFlopButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="uth-flop-buttons"]',
    messageKey: 'tutorial.postFlopButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="uth-river-buttons"]',
    messageKey: 'tutorial.riverButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="uth-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/**
 * Hand rank display name lookup. Indices mirror Go's `PokerHandNames`
 * (sync: `internal/domain/PokerPlayer.go`), so 0 = High card, 9 = Royal flush.
 */
const HAND_RANK_KEYS: Readonly<Record<number, string>> = {
  0: 'handRank.0',
  1: 'handRank.1',
  2: 'handRank.2',
  3: 'handRank.3',
  4: 'handRank.4',
  5: 'handRank.5',
  6: 'handRank.6',
  7: 'handRank.7',
  8: 'handRank.8',
  9: 'handRank.9',
};

/** Renders the Ultimate Texas Hold'em game page. */
export const UltimateTexasHoldemPage = withTutorial(
  UltimateTexasHoldemPageContent,
  'ultimatetexasholdem',
  UTH_TUTORIAL_STEPS,
);

/** Inner content of the Ultimate Texas Hold'em page, wrapped by TutorialProvider. */
function UltimateTexasHoldemPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ultimatetexasholdem');

  const [anteAmount, setAnteAmount] = useState(100);
  const [tripsAmount, setTripsAmount] = useState(0);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(ultimatetexasholdemApi.exec);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('ultimatetexasholdem', state);

  useMountReset(execApi);

  // Cap the ante input whenever chips drop below the current selection
  // (e.g. after a losing round) so the UI cannot submit an out-of-range bet.
  useEffect(() => {
    if (!state) return;
    const cap = Math.floor(state.chips / 2);
    setAnteAmount((a) => (a > cap ? Math.max(cap, 0) : a));
  }, [state]);

  // Keep the optional Trips side bet within the chips left after the Ante+Blind
  // commitment (ante*2), re-capping it whenever the ante or the chip balance changes.
  useEffect(() => {
    if (!state) return;
    const { maxTrips } = utHoldemBetBounds(anteAmount, 0, state.chips);
    setTripsAmount((tr) => (tr > maxTrips ? maxTrips : tr));
  }, [state, anteAmount]);

  const isBetPhase = state?.phase === UltimateTexasHoldemPhase.BET;
  const isPreFlopPhase = state?.phase === UltimateTexasHoldemPhase.PRE_FLOP;
  const isFlopPhase = state?.phase === UltimateTexasHoldemPhase.FLOP;
  const isRiverPhase = state?.phase === UltimateTexasHoldemPhase.RIVER;
  const isEndPhase = state?.phase === UltimateTexasHoldemPhase.END;
  const canPlay4x = !!state && state.chips >= state.anteBet * 4;
  const canPlay3x = !!state && state.chips >= state.anteBet * 3;
  const canPlay2x = !!state && state.chips >= state.anteBet * 2;
  const canPlay1x = !!state && state.chips >= state.anteBet;

  // Ante (×2 for the matched Blind) + Trips must fit within chips. Drives the
  // trips cap, the combined-total readout, the over-balance alert, and submit gating.
  const betBounds = utHoldemBetBounds(anteAmount, tripsAmount, state?.chips ?? 0);

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', anteAmount, tripsAmount),
        enabled: isBetPhase && betBounds.valid,
        label: 'bet',
      },
      {
        key: '4',
        action: () => execApi('play', undefined, undefined, 4),
        enabled: isPreFlopPhase && canPlay4x,
        label: 'play',
      },
      {
        key: '3',
        action: () => execApi('play', undefined, undefined, 3),
        enabled: isPreFlopPhase && canPlay3x,
        label: 'play',
      },
      { key: 'c', action: () => execApi('check'), enabled: isPreFlopPhase || isFlopPhase, label: 'check' },
      {
        key: '2',
        action: () => execApi('play', undefined, undefined, 2),
        enabled: isFlopPhase && canPlay2x,
        label: 'play',
      },
      {
        key: '1',
        action: () => execApi('play', undefined, undefined, 1),
        enabled: isRiverPhase && canPlay1x,
        label: 'play',
      },
      { key: 'f', action: () => execApi('fold'), enabled: isRiverPhase, label: 'fold' },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase, label: 'reset' },
    ],
    [
      execApi,
      anteAmount,
      tripsAmount,
      isBetPhase,
      betBounds.valid,
      isPreFlopPhase,
      isFlopPhase,
      isRiverPhase,
      isEndPhase,
      canPlay4x,
      canPlay3x,
      canPlay2x,
      canPlay1x,
    ],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) {
    return <GameSkeleton gameKey="ultimatetexasholdem" layout={{ kind: 'casino-table', sections: [2, 5, 2] }} />;
  }

  const handleBet = () => execApi('bet', anteAmount, tripsAmount);
  const handlePlay = (mult: number) => execApi('play', undefined, undefined, mult);
  const handleCheck = () => execApi('check');
  const handleFold = () => execApi('fold');
  const handleReset = () => execApi('reset');

  const phaseName = isBetPhase
    ? t('phase.bet')
    : isPreFlopPhase
      ? t('phase.preFlop')
      : isFlopPhase
        ? t('phase.flop')
        : isRiverPhase
          ? t('phase.river')
          : t('phase.end');

  const hasMadeHand = (isFlopPhase || isRiverPhase) && state.playerHand.length + state.community.length >= 5;
  const madeHandText = hasMadeHand
    ? t('madeHand', { hand: t(HAND_RANK_KEYS[state.playerHandRank] ?? 'handRank.0') })
    : '';

  return (
    <GamePageShell
      title={tc('nav.ultimatetexasholdem')}
      gameThemeBg={gameTheme.ultimatetexasholdem.bg}
      phaseName={phaseName}
      gamePath="/ultimatetexasholdem"
      gameEndFlag={isEndPhase}
      winShow={isEndPhase && state.result > 0}
      lossShow={isEndPhase && state.result < 0}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <span>
          {t('label.chips')}: {state.chips}
        </span>
      }
    >
      <div
        data-testid="card-area"
        className={[`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`, !isBetPhase && 'flex-1']
          .filter(Boolean)
          .join(' ')}
      >
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {(isPreFlopPhase || isFlopPhase || isRiverPhase) && (
          <section className="mb-3" data-testid="uth-current-bets">
            <h2 className="text-ds-text-primary text-center font-bold text-sm mb-1">{t('currentBets.title')}</h2>
            <div className="flex flex-wrap justify-center gap-x-4 text-ds-text-muted text-sm tabular-nums">
              {state.anteBet > 0 && (
                <span>
                  {t('label.ante')}: {state.anteBet}
                </span>
              )}
              {state.blindBet > 0 && (
                <span>
                  {t('label.blind')}: {state.blindBet}
                </span>
              )}
              {state.tripsBet > 0 && (
                <span>
                  {t('label.trips')}: {state.tripsBet}
                </span>
              )}
              {state.playBet > 0 && (
                <span>
                  {t('label.playBet')}: {state.playBet}
                </span>
              )}
            </div>
          </section>
        )}

        <label className="flex items-center gap-1 text-ds-text-primary text-xs justify-center mb-2 cursor-pointer min-h-[44px]">
          <input
            type="checkbox"
            checked={frontendHintEnabled}
            onChange={(e) => setFrontendHintEnabled(e.target.checked)}
          />
          {tc('hint.toggle', { ns: 'tutorial' })}
        </label>

        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

        {isBetPhase && (
          <div className="flex flex-col items-center justify-center py-4 gap-4">
            <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
            <details className="bg-black/30 rounded-lg w-full max-w-sm">
              <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                {t('payoutRef.title')}
              </summary>
              <div className="px-4 pb-3 text-ds-text-muted text-sm space-y-2">
                <div>
                  <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.blindHeader')}</div>
                  <ul className="space-y-0.5">
                    {(
                      [
                        'blindRoyalFlush',
                        'blindStraightFlush',
                        'blindFourOfAKind',
                        'blindFullHouse',
                        'blindFlush',
                        'blindStraight',
                      ] as const
                    ).map((key) => (
                      <li key={key}>{t(`payoutRef.${key}`)}</li>
                    ))}
                  </ul>
                </div>
                <div>
                  <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.tripsHeader')}</div>
                  <ul className="space-y-0.5">
                    {(
                      [
                        'tripsRoyalFlush',
                        'tripsStraightFlush',
                        'tripsFourOfAKind',
                        'tripsFullHouse',
                        'tripsFlush',
                        'tripsStraight',
                        'tripsThreeOfAKind',
                      ] as const
                    ).map((key) => (
                      <li key={key}>{t(`payoutRef.${key}`)}</li>
                    ))}
                  </ul>
                </div>
              </div>
            </details>
          </div>
        )}

        {state.community.length > 0 && (
          <div className="mb-4">
            <div className="text-ds-text-primary font-bold text-center mb-1">
              <span aria-hidden="true">🃏</span> {t('board')}
            </div>
            <div className="flex justify-center gap-2 flex-wrap">
              {state.community.map((card, i) => (
                <AnimatedCard key={`c-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
              ))}
            </div>
          </div>
        )}

        {state.dealerHand.length > 0 && (
          <div className="mb-4">
            <div className="text-ds-error-text font-bold text-center mb-1">
              <span aria-hidden="true">🔴</span> {t('dealer')}
              {isEndPhase && (
                <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.dealerHandRank] ?? 'handRank.0')})</span>
              )}
            </div>
            <div className="flex justify-center gap-2 flex-wrap">
              {state.dealerHand.map((card, i) =>
                isMaskedCard(card) ? (
                  <AnimatedCardBack key={`d-back-${i}`} width={cardWidth} />
                ) : (
                  <AnimatedCard key={`d-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                ),
              )}
            </div>
          </div>
        )}

        {state.playerHand.length > 0 && (
          <div className="mb-4" data-tutorial="uth-results">
            <div className="text-ds-warning font-bold text-center mb-1">
              <span aria-hidden="true">🟡</span> {t('player')}
              {isEndPhase && (
                <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.playerHandRank] ?? 'handRank.0')})</span>
              )}
            </div>
            <div className="flex justify-center gap-2 flex-wrap">
              {state.playerHand.map((card, i) => (
                <AnimatedCard key={`p-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
              ))}
            </div>
          </div>
        )}

        {isEndPhase && (
          <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="payout-breakdown">
            <div>{state.dealerQualified ? t('result.dealerQualified') : t('result.dealerNotQualified')}</div>
            {state.anteBet > 0 && (
              <div>
                {t('payout.ante')}: {state.antePayout} (
                {t(
                  `payout.${state.antePayout > state.anteBet ? 'win' : state.antePayout === state.anteBet ? 'push' : 'loss'}`,
                )}
                )
              </div>
            )}
            {state.blindBet > 0 && (
              <div>
                {t('payout.blind')}: {state.blindPayout} (
                {t(
                  `payout.${state.blindPayout > state.blindBet ? 'win' : state.blindPayout === state.blindBet ? 'push' : 'loss'}`,
                )}
                )
              </div>
            )}
            {state.playBet > 0 && (
              <div>
                {t('payout.play')}: {state.playPayout} (
                {t(
                  `payout.${state.playPayout > state.playBet ? 'win' : state.playPayout === state.playBet ? 'push' : 'loss'}`,
                )}
                )
              </div>
            )}
            {state.tripsBet > 0 && (
              <div>
                {t('payout.trips')}: {state.tripsPayout} (
                {t(
                  `payout.${state.tripsPayout > state.tripsBet ? 'win' : state.tripsPayout === state.tripsBet ? 'push' : 'loss'}`,
                )}
                )
              </div>
            )}
            <div className="font-bold mt-1">
              {t('payout.total')}: {state.totalPayout}
            </div>
            <div className="font-bold">
              {t('payout.netChange')}: {formatSignedDelta(state.netChange)}
            </div>
          </div>
        )}

        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
      </div>

      <GameFooter className={`${gameTheme.ultimatetexasholdem.footer} px-4 pt-3`}>
        <ErrorAlert message={error} onRetry={retry} />
        {isBetPhase && (
          <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="uth-bet-controls">
            <div className="flex items-center gap-2">
              <ChipBetInput
                id="uth-ante-amount"
                label={t('label.ante')}
                value={anteAmount}
                onChange={setAnteAmount}
                min={10}
                max={Math.floor(state.chips / 2)}
                step={10}
                disabled={loading}
                showSteppers
              />
              <span className="text-ds-text-muted text-xs">×2 ({t('label.blind')})</span>
            </div>
            <ChipBetInput
              id="uth-trips-amount"
              label={t('label.trips')}
              value={tripsAmount}
              onChange={setTripsAmount}
              min={0}
              max={betBounds.maxTrips}
              step={10}
              disabled={loading}
              autoClamp={false}
              invalid={!betBounds.valid}
              describedBy={betBounds.valid ? undefined : 'uth-bet-error'}
              showSteppers
            />
            <div className="text-ds-text-muted text-sm" data-testid="uth-bet-total">
              {t('betSummary.total')}: {betBounds.total} / {state.chips}
            </div>
            {!betBounds.valid && (
              <p id="uth-bet-error" role="alert" className="text-ds-error-text text-sm" data-testid="uth-bet-error">
                {t('betSummary.overBalance')}
              </p>
            )}
            <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading || !betBounds.valid}>
              {t('button.bet')}
            </button>
          </div>
        )}
        {/* 各ストリートで 4x/3x/2x/1x/チェック/フォールドを選ぶのに、今の役が
            どこにも出ていなかった。姉妹の Mississippi Stud は ms-made-hand として
            常時出している。ランクはサーバが updatePlayerCurrentRank で
            各ストリート更新するので、判定を2つ持たない。 */}
        {madeHandText && (
          <p className="text-center text-sm font-medium pb-1 text-ds-text-primary" data-testid="uth-made-hand">
            {madeHandText}
          </p>
        )}
        <div className="sr-only" role="status" aria-live="polite" data-testid="uth-made-hand-announcement">
          {madeHandText}
        </div>
        {isPreFlopPhase &&
          (() => {
            const strength = utHoldemPreflopStrength(state.playerHand);
            const strong = strength === 'strong';
            const moderate = strength === 'moderate';
            // Spell out *why* a raise size is recommended so beginners understand the
            // ring/pulse affordance instead of guessing.
            const strengthColor = strong ? 'text-ds-success' : moderate ? 'text-ds-warning' : 'text-ds-text-muted';
            return (
              <>
                <p className={`text-center text-sm font-medium pb-1 ${strengthColor}`} data-testid="uth-preflop-eval">
                  {t(`preflopEval.${strength}`)}
                </p>
                <div
                  className="flex justify-center gap-2 pb-2"
                  data-tutorial="uth-pre-flop-buttons"
                  data-preflop-strength={strength}
                >
                  <button
                    type="button"
                    className={`${btnSuccess} ${strong ? 'ring-2 ring-ds-warning animate-pulse' : ''}`}
                    onClick={() => handlePlay(4)}
                    disabled={loading || !canPlay4x}
                    data-testid="play-4x"
                  >
                    {t('button.play4x', { playBet: t('label.playBet'), amount: anteAmount * 4 })}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${moderate ? 'ring-2 ring-ds-warning animate-pulse' : ''}`}
                    onClick={() => handlePlay(3)}
                    disabled={loading || !canPlay3x}
                    data-testid="play-3x"
                  >
                    {t('button.play3x', { playBet: t('label.playBet'), amount: anteAmount * 3 })}
                  </button>
                  <button type="button" className={btnSecondary} onClick={handleCheck} disabled={loading}>
                    {t('button.check')}
                  </button>
                </div>
              </>
            );
          })()}
        {isFlopPhase && (
          <div className="flex justify-center gap-2 pb-2" data-tutorial="uth-flop-buttons">
            <button type="button" className={btnWarning} onClick={() => handlePlay(2)} disabled={loading || !canPlay2x}>
              {t('button.play2x', { playBet: t('label.playBet'), amount: anteAmount * 2 })}
            </button>
            <button type="button" className={btnSecondary} onClick={handleCheck} disabled={loading}>
              {t('button.check')}
            </button>
          </div>
        )}
        {isRiverPhase && (
          <div className="flex justify-center gap-2 pb-2" data-tutorial="uth-river-buttons">
            <button type="button" className={btnWarning} onClick={() => handlePlay(1)} disabled={loading || !canPlay1x}>
              {t('button.play1x', { playBet: t('label.playBet'), amount: anteAmount })}
            </button>
            <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
              {t('button.fold')}
            </button>
          </div>
        )}
        {isEndPhase && (
          <div className="flex justify-center gap-2 pb-2">
            <GameResetButton
              isGameEnd={isEndPhase}
              onReset={handleReset}
              requestConfirm={requestConfirm}
              loading={loading}
            />
            <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
        <ActionShortcutsPanel bindings={actionBindings} data-testid="ultimate-texas-holdem-kbd-shortcuts" />
      </GameFooter>
    </GamePageShell>
  );
}
