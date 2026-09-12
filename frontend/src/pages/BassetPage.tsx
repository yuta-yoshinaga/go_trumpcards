import { useEffect, useMemo, useState } from 'react';
import { bassetApi } from '../api/games/basset';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { gameTheme } from '../styles/gameTheme';
import type { BassetResponse } from '../types/games/basset';
import type { TutorialStep } from '../types/tutorial';
import { BASSET_HELP, parseBassetCommand } from '../utils/cli/commands/bassetCommands';
import { formatBassetState } from '../utils/cli/formatters/bassetFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';

const STEPS: TutorialStep[] = [];
const BassetPhase = { BETTING: 1, TURN: 2, DECISION: 3, ROUND_END: 4, GAME_END: 5 } as const;
const PHASES: Record<number, string> = { 1: 'betting', 2: 'turn', 3: 'decision', 4: 'roundEnd', 5: 'gameEnd' };

/** Renders the Basset page and its paroli decision controls. */
export const BassetPage = withTutorial(BassetPageContent, 'basset', STEPS);

function BassetPageContent() {
  const { t, tc, confirmOpen, requestConfirm, confirmReset, cancelReset } = useGamePageSetup('basset');
  const { state, loading, error, exec, retry } = useGameApi(bassetApi.exec);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('basset');
  const [rank, setRank] = useState(1);
  const [amount, setAmount] = useState(10);
  useEffect(() => {
    exec('reset');
  }, [exec]);
  const config: CliGameConfig<BassetResponse, Parameters<typeof bassetApi.exec>> = useMemo(
    () => ({
      gameName: 'basset',
      parseCommand: parseBassetCommand,
      formatResponse: formatBassetState,
      helpText: BASSET_HELP,
      localCommand: hintLocalCommand(null),
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, config, state, { addInput, addOutput, addError, clearLog });
  if (!state)
    return <GameSkeleton gameKey="basset" layout={{ kind: 'casino-table', sections: [2], footerStyle: 'bet' }} />;
  const ended = state.phase === BassetPhase.GAME_END || state.gameEndFlag;
  return (
    <GamePageShell
      title={tc('nav.basset')}
      gameThemeBg={gameTheme.basset.bg}
      phaseName={t(PHASES[state.phase] ?? 'turn')}
      gamePath="/basset"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <div className="flex-1 min-h-0 overflow-y-auto p-6 text-ds-text-primary">
          <div className="mb-4 flex flex-wrap justify-center gap-4 text-sm">
            <span>{t('chips', { count: state.chips })}</span>
            <span>{t('turns', { played: state.turnsPlayed, total: state.turnsTotal })}</span>
            <span>{t('remaining', { count: state.remaining })}</span>
          </div>
          <div className="mx-auto grid max-w-xl gap-4 rounded-lg bg-black/20 p-6 text-center">
            <p>
              {state.bet ? `${t('bet')}: ${state.bet.rank} / ${state.bet.amount} / ${state.bet.stage}` : t('noBet')}
            </p>
            {state.bankerCard && (
              <p>
                {t('bankerCard')}: {state.bankerCard.value}
              </p>
            )}
            {state.playerCard && (
              <p>
                {t('playerCard')}: {state.playerCard.value}
              </p>
            )}
            {state.phase === BassetPhase.BETTING || state.phase === BassetPhase.TURN ? (
              <div className="flex flex-wrap justify-center gap-3">
                <input
                  className="min-h-[44px] w-20 rounded bg-black/30 p-2"
                  type="number"
                  min="1"
                  max="13"
                  value={rank}
                  onChange={(e) => setRank(Number(e.target.value))}
                />
                <input
                  className="min-h-[44px] w-24 rounded bg-black/30 p-2"
                  type="number"
                  min="10"
                  value={amount}
                  onChange={(e) => setAmount(Number(e.target.value))}
                />
                <button
                  className="min-h-[44px] rounded bg-ds-accent px-4 text-black"
                  type="button"
                  onClick={() => exec('bet', { rank, amount })}
                >
                  {t('placeBet')}
                </button>
                <button className="min-h-[44px] rounded bg-ds-surface px-4" type="button" onClick={() => exec('deal')}>
                  {t('deal')}
                </button>
              </div>
            ) : null}
            {state.phase === BassetPhase.DECISION ? (
              <div className="flex justify-center gap-3">
                <button className="min-h-[44px] rounded bg-ds-success px-4" type="button" onClick={() => exec('take')}>
                  {t('take')}
                </button>
                <button
                  className="min-h-[44px] rounded bg-ds-warning px-4"
                  type="button"
                  onClick={() => exec('paroli')}
                >
                  {t('paroli')}
                </button>
              </div>
            ) : null}
            {state.phase === BassetPhase.ROUND_END ? (
              <button
                className="min-h-[44px] rounded bg-ds-accent px-4 text-black"
                type="button"
                onClick={() => exec('next')}
              >
                {t('next')}
              </button>
            ) : null}
          </div>
        </div>
      )}
      <GameFooter className={`${gameTheme.basset.footer} px-4 pt-3`}>
        <ErrorAlert message={error} onRetry={retry} />
        <div className="flex justify-center pb-2">
          <GameResetButton
            isGameEnd={ended}
            onReset={() => void exec('reset')}
            requestConfirm={requestConfirm}
            loading={loading}
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
