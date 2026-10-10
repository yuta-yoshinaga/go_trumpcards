import { useState } from 'react';
import { tehonbikiApi } from '../api/games/tehonbiki';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary } from '../styles/buttonStyles';
import { TehonbikiPhase } from '../types/phases';

const TEHONBIKI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TehonbikiPhase.BET]: 'bet',
  [TehonbikiPhase.RESULT]: 'result',
  [TehonbikiPhase.GAME_END]: 'gameEnd',
};

/** Renders the Tehonbiki number-guessing game. */
export const TehonbikiPage = withTutorial(TehonbikiPageContent, 'tehonbiki', []);
function TehonbikiPageContent() {
  const { t, tc, confirmOpen, requestConfirm, confirmReset, cancelReset } = useGamePageSetup('tehonbiki');
  const { state, loading, error, exec, retry } = useGameApi(tehonbikiApi.exec);
  const phaseNames = usePhaseNames('tehonbiki', TEHONBIKI_PHASE_KEYS);
  const [betType, setBetType] = useState('single');
  const [numbers, setNumbers] = useState<number[]>([]);
  const [bet, setBet] = useState('50');
  useMountReset(exec);
  if (!state) return <GameSkeleton gameKey="tehonbiki" layout={{ kind: 'casino-table', sections: [1] }} />;
  const minBet = state.minBet;
  const maxBet = Math.min(state.maxBet, state.chips);
  const betAmount = Number(bet);
  const betValid = bet !== '' && Number.isInteger(betAmount) && betAmount >= minBet && betAmount <= maxBet;
  const requiredNumbers = betType === 'single' ? 1 : betType === 'double' ? 2 : 3;
  const halfGroupValid =
    betType !== 'half' ||
    (numbers.length === 3 &&
      ([1, 2, 3].every((n) => numbers.includes(n)) || [4, 5, 6].every((n) => numbers.includes(n))));
  const selectionValid = numbers.length === requiredNumbers && halfGroupValid;
  const toggle = (n: number) => setNumbers((v) => (v.includes(n) ? v.filter((x) => x !== n) : [...v, n]));
  return (
    <GamePageShell
      title={tc('nav.tehonbiki')}
      gameThemeBg="bg-ds-surface"
      phaseName={phaseNames[state.phase]}
      gamePath="/tehonbiki"
      gameEndFlag={state.gameEndFlag}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      <div className="text-center text-ds-text-primary p-6">
        <p>
          {t('label.chips')}: {state.chips}
        </p>
        {state.parentCard && (
          <p>
            {t('label.parentCard')}: {state.parentCard}
          </p>
        )}
        <div className="flex justify-center gap-2 p-4">
          {[1, 2, 3, 4, 5, 6].map((n) => (
            <button
              className={btnPrimary}
              type="button"
              key={n}
              onClick={() => toggle(n)}
              disabled={state.phase !== 0 || loading}
              aria-pressed={numbers.includes(n)}
            >
              {n}
            </button>
          ))}
        </div>
        <div role="status" aria-live="polite" aria-atomic="true">
          {state.phase === 0 && (
            <>
              <p className="mx-auto mb-4 max-w-xl rounded-lg border border-ds-border-subtle bg-ds-surface p-3">
                {t('summary', {
                  numbers: numbers.length > 0 ? numbers.join(t('listSeparator')) : t('summary.noNumbers'),
                  betType: t(`betType.${betType}`),
                  bet: bet === '' ? t('summary.noBet') : bet,
                })}
              </p>
              <p className="mb-4">{t('selection.requirement', { count: requiredNumbers })}</p>
              {betType === 'half' && (
                <p className="mb-4">
                  {numbers.length === 3 && !halfGroupValid ? t('selection.halfInvalid') : t('selection.halfValid')}
                </p>
              )}
            </>
          )}
          {state.phase !== 0 && state.result !== 0 && (
            <div className="mx-auto mb-4 max-w-xl rounded-lg border border-ds-border-subtle bg-ds-surface p-3">
              <p className={state.result === 1 ? 'text-ds-success' : 'text-ds-error-text'}>
                {t(state.result === 1 ? 'result.win' : 'result.lose')}
              </p>
              <p>
                {state.result === 1
                  ? t('result.payout', { amount: state.payout })
                  : t('result.lostBet', { amount: state.bet })}
              </p>
            </div>
          )}
        </div>
        {state.phase === 0 ? (
          <>
            <label htmlFor="tehonbiki-bet-type">{t('label.betType')}</label>
            <select
              id="tehonbiki-bet-type"
              value={betType}
              onChange={(e) => {
                setBetType(e.target.value);
                setNumbers([]);
              }}
            >
              <option value="single">{t('betType.single')}</option>
              <option value="double">{t('betType.double')}</option>
              <option value="triple">{t('betType.triple')}</option>
              <option value="half">{t('betType.half')}</option>
            </select>
            <label htmlFor="tehonbiki-bet-amount">{t('label.bet')}</label>
            <input
              id="tehonbiki-bet-amount"
              type="number"
              min={minBet}
              max={maxBet}
              value={bet}
              aria-invalid={!betValid}
              aria-describedby="tehonbiki-bet-constraint"
              onChange={(e) => setBet(e.target.value)}
            />
            <p id="tehonbiki-bet-constraint">{t('bet.constraint', { min: minBet, max: maxBet })}</p>
            <p role="alert">{!betValid ? t('bet.outOfRange', { min: minBet, max: maxBet }) : ''}</p>
            <button
              className={`${btnPrimary} aria-disabled:cursor-not-allowed aria-disabled:opacity-50`}
              type="button"
              aria-disabled={loading || !betValid || !selectionValid}
              onClick={() => {
                if (!loading && betValid && selectionValid) exec('bet', { numbers, betType, bet: betAmount });
              }}
            >
              {t('button.bet')}
            </button>
          </>
        ) : (
          !state.gameEndFlag && (
            <button className={btnPrimary} type="button" onClick={() => exec('next')}>
              {t('button.next')}
            </button>
          )
        )}
      </div>
      <GameFooter>
        <ErrorAlert message={error} onRetry={retry} />
        <GameResetButton
          isGameEnd={state.gameEndFlag}
          onReset={() => exec('reset')}
          requestConfirm={requestConfirm}
          loading={loading}
        />
      </GameFooter>
    </GamePageShell>
  );
}
