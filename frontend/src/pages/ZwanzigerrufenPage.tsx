import { useCallback, useEffect, useState } from 'react';
import { zwanzigerrufenApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import { ZwanzigerrufenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Zwanzigerrufen tutorial step definitions. */
const ZWANZIGERRUFEN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="zw-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="zw-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="zw-actions"]', messageKey: 'tutorial.actions', placement: 'top', advanceOn: 'next' },
];

/** Number of cards buried in the talon exchange. */
const DISCARD_COUNT = 6;

/** CPU difficulty options. */
const CPU_DIFFICULTY_OPTIONS = [0, 1, 2] as const;

/** Match-length options, in deals. */
const TARGET_DEALS_OPTIONS = [1, 2, 4, 8, 12] as const;

/** Renders the Zwanzigerrufen (ツヴァンツィガールーフェン) page: an Austrian calling tarock. */
export const ZwanzigerrufenPage = withTutorial(
  ZwanzigerrufenPageContent,
  'zwanzigerrufen',
  ZWANZIGERRUFEN_TUTORIAL_STEPS,
);

/** Inner content of the Zwanzigerrufen page, wrapped by TutorialWrapper. */
function ZwanzigerrufenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('zwanzigerrufen');
  const [selected, setSelected] = useState<number[]>([]);
  const [cpuDifficulty, setCpuDifficulty] = useState(1);
  const [targetDeals, setTargetDeals] = useState(4);

  const onSuccess = useCallback(async () => {
    setSelected([]);
  }, []);
  const { loading, error, state, exec: callApi, retry } = useGameApi(zwanzigerrufenApi.exec, { onSuccess });
  const { cardWidth, isMobile } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('zwanzigerrufen', state);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    callApi('reset');
  }, []);

  const resetWithConfig = useCallback(() => {
    callApi('reset', { config: { cpuDifficulty, targetDeals } });
  }, [callApi, cpuDifficulty, targetDeals]);

  if (!state) {
    return (
      <GameSkeleton gameKey="zwanzigerrufen" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman) ?? state.players[0] ?? null;
  const isBid = state.phase === ZwanzigerrufenPhase.BID;
  const isTalon = state.phase === ZwanzigerrufenPhase.TALON;
  const isPlay = state.phase === ZwanzigerrufenPhase.PLAY;
  const isTrickEnd = state.phase === ZwanzigerrufenPhase.TRICK_END;
  const isRoundEnd = state.phase === ZwanzigerrufenPhase.ROUND_END;
  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn && !isGameEnd;
  const humanWon = isGameEnd && state.winnerPlayer === (human?.id ?? 0);
  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isTrickEnd
        ? t('phase.trickEnd')
        : isTalon
          ? t('phase.talon')
          : isBid
            ? t('phase.bid')
            : t('phase.play');

  /** Toggles a hand card: single-select while playing, six-select while burying. */
  const toggleCard = (idx: number) => {
    if (isPlay && isHumanTurn) {
      // **Play sends immediately.** Holding a selection would let the board move
      // under it while a CPU acts.
      callApi('play', { cardIndex: idx });
      return;
    }
    if (!isTalon || !isHumanTurn) return;
    setSelected((prev) =>
      prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx].slice(0, DISCARD_COUNT),
    );
  };

  const seatName = (id: number): string =>
    state.players[id]?.isHuman ? t('you') : t('cpu', { n: id, defaultValue: `CPU${id}` });

  return (
    <GamePageShell
      title={tc('nav.zwanzigerrufen')}
      gameThemeBg={gameTheme.zwanzigerrufen.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/zwanzigerrufen"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
        <div className="text-center text-xs text-ds-text-muted" data-testid="zw-info">
          <span className="mr-3">{t('deal', { n: state.roundNumber, total: state.totalRounds })}</span>
          <span className="mr-3">{t('trick', { n: state.trickNumber })}</span>
          <span>{t('contract', { name: t(`contract.${state.contractName}`, { defaultValue: '-' }) })}</span>
        </div>

        {/* **The called trump is public; who holds it is not.** */}
        {state.calledTrump > 0 && (
          <div className="text-center text-xs text-ds-accent" data-testid="zw-called">
            {state.partnerRevealed && state.partnerIdx >= 0
              ? t('calledRevealed', { trump: state.calledTrump, name: seatName(state.partnerIdx) })
              : t('calledSecret', { trump: state.calledTrump })}
          </div>
        )}

        <div className="flex flex-wrap justify-center gap-3" data-testid="zw-seats">
          {state.players.map((p) => (
            <div key={p.id} className="text-center text-xs text-ds-text-muted" data-testid={`zw-seat-${p.id}`}>
              <div className="text-ds-text-primary">
                {seatName(p.id)}
                {p.isDeclarer && <span className="ml-1 text-ds-warning">{t('roleDeclarer')}</span>}
                {p.isPartner && <span className="ml-1 text-ds-info">{t('rolePartner')}</span>}
              </div>
              <div>{t('cards', { count: p.cardCount })}</div>
              <div data-testid={`zw-seat-${p.id}-points`}>{t('points', { points: p.cardPoints })}</div>
              <div>{t('score', { score: p.score })}</div>
            </div>
          ))}
        </div>

        <TrickDisplay
          currentTrick={state.currentTrick}
          players={state.players.map((p) => ({ id: p.id, name: seatName(p.id), isHuman: p.isHuman }))}
          cardWidth={cardWidth}
          label={t('currentTrick')}
          dataTutorial="zw-trick-display"
          winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
          winnerLabel={t('trickWinner')}
        />

        {human && (
          <PlayerHandSection
            humanPlayer={human}
            selectedCardIndices={selected}
            toggleCard={toggleCard}
            cardWidth={cardWidth}
            isMobile={isMobile}
            dataTutorialPrefix="zw"
            validIndices={isPlay && isHumanTurn ? state.playableIndices : undefined}
            restrictedTooltip={t('restricted')}
          />
        )}

        {isRoundEnd && state.breakdown && (
          <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="zw-round-result">
            <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
            <div className="text-ds-success mb-1">
              {state.breakdown.loser >= 0
                ? t('roundResult.trischaken', {
                    name: seatName(state.breakdown.loser),
                    points: state.breakdown.teamPoints,
                  })
                : t(state.breakdown.won ? 'roundResult.won' : 'roundResult.lost', {
                    points: state.breakdown.teamPoints,
                    threshold: state.breakdown.threshold,
                  })}
            </div>
            {state.breakdown.seats.map((delta, i) => (
              <div key={i} data-testid={`zw-round-seat-${i}`}>
                {t('roundResult.seat', { name: seatName(i), delta })}
              </div>
            ))}
          </div>
        )}

        {isGameEnd && (
          <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="zw-result">
            <div className="mb-1 text-ds-text-primary">{t('result.title')}</div>
            <div className="text-ds-success mb-1">
              {state.winnerPlayer < 0 ? t('result.draw') : t('result.winner', { name: seatName(state.winnerPlayer) })}
            </div>
            {state.players.map((p) => (
              <div key={p.id}>{t('result.score', { name: seatName(p.id), score: p.score })}</div>
            ))}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        <ErrorAlert message={error} onRetry={retry} />
        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <SettingsPanel
        title={t('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'select' as const,
                id: 'cpuDifficulty',
                label: t('settings.cpuDifficulty'),
                value: String(cpuDifficulty),
                options: CPU_DIFFICULTY_OPTIONS.map((v) => ({ value: String(v), label: t(`settings.difficulty${v}`) })),
                onSelect: (v: string) => setCpuDifficulty(Number.parseInt(v, 10)),
              },
              {
                type: 'select' as const,
                id: 'targetDeals',
                label: t('settings.targetDeals'),
                value: String(targetDeals),
                options: TARGET_DEALS_OPTIONS.map((v) => ({ value: String(v), label: String(v) })),
                onSelect: (v: string) => setTargetDeals(Number.parseInt(v, 10)),
              },
              hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
            ],
          },
        ]}
      />

      <GameFooter className={`${gameTheme.zwanzigerrufen.footer} px-4 py-2.5`}>
        <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="zw-actions">
          {isBid && isHumanTurn && (
            <>
              <button
                type="button"
                className={btnPrimary}
                onClick={() => callApi('bid', { bid: 'rufer' })}
                disabled={loading}
                data-testid="zw-bid-rufer"
              >
                {t('bidRufer')}
              </button>
              <button
                type="button"
                className={btnSecondary}
                onClick={() => callApi('bid', { bid: 'solo' })}
                disabled={loading}
                data-testid="zw-bid-solo"
              >
                {t('bidSolo')}
              </button>
              <button
                type="button"
                className={btnSecondary}
                onClick={() => callApi('pass')}
                disabled={loading}
                data-testid="zw-pass"
              >
                {t('pass')}
              </button>
            </>
          )}
          {isTalon && isHumanTurn && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => callApi('discard', { cardIndices: selected })}
              disabled={loading || selected.length !== DISCARD_COUNT}
              data-testid="zw-discard"
            >
              {t('discard', { count: selected.length, total: DISCARD_COUNT })}
            </button>
          )}
          {isTrickEnd && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => callApi('next')}
              disabled={loading}
              data-testid="zw-next-trick"
            >
              {t('nextTrick')}
            </button>
          )}
          {isRoundEnd && !isGameEnd && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => callApi('nextround')}
              disabled={loading}
              data-testid="zw-next-round"
            >
              {t('nextDeal')}
            </button>
          )}
          {isGameEnd && (
            <button type="button" className={btnSuccess} onClick={resetWithConfig} disabled={loading}>
              {t('newGame')}
            </button>
          )}
          <GameResetButton
            isGameEnd={isGameEnd}
            onReset={resetWithConfig}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="zw-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
