import { useCallback, useEffect, useState } from 'react';
import { sakuraApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SakuraBonus, SakuraPlayer } from '../types/card';
import { SakuraPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Sakura (さくら/肥後花) tutorial step definitions. */
const SAKURA_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sakura-field"]', messageKey: 'tutorial.field', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sakura-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sakura-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="sakura-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Seat options offered in the settings panel. */
const SEAT_OPTIONS = [2, 3, 4] as const;

/** Round-count options offered in the settings panel. */
const ROUND_OPTIONS = [1, 3, 6, 12] as const;

/** Renders the Sakura (さくら/肥後花) page: a hanafuda matching game scored by counting cards. */
export const SakuraPage = withTutorial(SakuraPageContent, 'sakura', SAKURA_TUTORIAL_STEPS);

/** Inner content of the Sakura page, wrapped by TutorialWrapper. */
function SakuraPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sakura');
  const [handIndex, setHandIndex] = useState<number | null>(null);
  const [seats, setSeats] = useState(3);
  const [rounds, setRounds] = useState(3);

  const onSuccess = useCallback(async () => {
    setHandIndex(null);
  }, []);
  const { loading, error, state, exec: callApi, retry } = useGameApi(sakuraApi.exec, { onSuccess });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sakura', state);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    callApi('reset');
  }, []);

  /** Restarts with the settings currently chosen in the panel. */
  const resetWithConfig = useCallback(() => {
    callApi('reset', { config: { seats, rounds } });
  }, [callApi, seats, rounds]);

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.sakura.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const human = state.players.find((p) => p.isHuman) ?? state.players[0] ?? null;
  const isPlayPhase = state.phase === SakuraPhase.PLAY;
  const isRoundEnd = state.phase === SakuraPhase.ROUND_END;
  const isGameEnd = state.phase === SakuraPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn && !isGameEnd;
  const humanWon = isGameEnd && state.winner === (human?.id ?? 0);
  const phaseName = isGameEnd ? t('phase.gameEnd') : isRoundEnd ? t('phase.roundEnd') : t('phase.play');

  // Field indices the currently-selected hand card can take (published by the server).
  const candidates = handIndex !== null && isPlayPhase && isHumanTurn ? (state.captureOptions[handIndex] ?? []) : [];
  const needsFieldPick = candidates.length > 1;
  const candidateSet = new Set(candidates);

  /** Localizes a seat's display name (the human seat reads as "you"). */
  const seatName = (p: SakuraPlayer): string => (p.isHuman ? t('you') : p.name);

  /** Renders a compact "name (points)" list of bonuses. */
  const bonusList = (bonuses: SakuraBonus[]): string =>
    bonuses.length === 0
      ? t('bonusNone')
      : bonuses.map((b) => `${t(`bonus.${b.key}`, { defaultValue: b.key })} (${b.points})`).join('  ·  ');

  /** Handles clicking a hand card: selects it, or plays immediately when there is no choice to make. */
  const onHandClick = (idx: number) => {
    if (!isPlayPhase || !isHumanTurn) return;
    const opts = state.captureOptions[idx] ?? [];
    if (opts.length > 1) {
      // Two field cards of that month: make the player pick which one to take.
      setHandIndex((prev) => (prev === idx ? null : idx));
      return;
    }
    callApi('play', { cardIndex: idx });
  };

  /** Handles clicking a field card while a two-way match is being resolved. */
  const onFieldClick = (fieldIdx: number) => {
    if (!needsFieldPick || handIndex === null) return;
    if (!candidateSet.has(fieldIdx)) return;
    callApi('play', { cardIndex: handIndex, fieldIndex: fieldIdx });
  };

  const dealerName = state.players[state.dealer] ? seatName(state.players[state.dealer]) : '';
  const turnName = state.players[state.currentTurn] ? seatName(state.players[state.currentTurn]) : '';

  return (
    <GamePageShell
      title={tc('nav.sakura')}
      gameThemeBg={gameTheme.sakura.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/sakura"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
        <div className="text-center text-xs text-ds-text-muted" data-tutorial="sakura-info">
          <span className="mr-3">{t('round', { n: state.round, total: state.totalRounds })}</span>
          <span className="mr-3">{t('stock', { count: state.stockCount })}</span>
          <span>{t('dealer', { name: dealerName })}</span>
        </div>

        {/* Opponent seats: counts and running points only — their hands never reach the page. */}
        <div className="flex flex-wrap justify-center gap-3" data-testid="sakura-opponents">
          {state.players
            .filter((p) => !p.isHuman)
            .map((p) => (
              <div key={p.id} className="text-center text-xs text-ds-text-muted" data-testid={`sakura-seat-${p.id}`}>
                <div className="text-ds-text-primary">{p.name}</div>
                <div>{t('taken', { count: p.takenCount })}</div>
                <div data-testid={`sakura-seat-${p.id}-points`}>{t('points', { points: p.totalPoints })}</div>
                <div>{t('score', { score: p.score })}</div>
              </div>
            ))}
        </div>

        {/* Field cards */}
        <div className="py-3 bg-black/20 rounded-lg" data-tutorial="sakura-field">
          <div className="text-center text-xs text-ds-text-muted mb-2">{t('field')}</div>
          <div className="flex justify-center gap-2 min-h-[60px] flex-wrap">
            {state.fieldCards.length === 0 ? (
              <span className="text-ds-text-muted text-sm self-center">{t('fieldEmpty')}</span>
            ) : (
              state.fieldCards.map((c, i) => {
                const isCandidate = candidateSet.has(i);
                return (
                  <button
                    key={i}
                    type="button"
                    onClick={() => onFieldClick(i)}
                    disabled={!needsFieldPick || !isCandidate}
                    className={`rounded transition-all ${
                      isCandidate ? 'ring-2 ring-ds-success motion-safe:animate-pulse' : ''
                    } ${needsFieldPick && isCandidate ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    data-testid={`field-card-${i}`}
                    data-capture-candidate={isCandidate || undefined}
                  >
                    <CardImage card={c} width={cardWidth * 0.9} />
                  </button>
                );
              })
            )}
          </div>
          {needsFieldPick && (
            <div className="text-center text-sm text-ds-accent mt-2" data-testid="sakura-field-pick">
              {t('pickField')}
            </div>
          )}
        </div>

        {/* Human seat: taken pile, points and hand */}
        {human && (
          <div className="text-center" data-testid="sakura-human">
            <div className="text-xs text-ds-text-muted mb-1">
              <span className="mr-3">{t('taken', { count: human.takenCount })}</span>
              <span className="mr-3" data-testid="sakura-human-points">
                {t('points', { points: human.totalPoints })}
              </span>
              <span className="mr-3">{t('score', { score: human.score })}</span>
              <span>{t('wins', { wins: human.roundWins })}</span>
            </div>
            <div className="text-xs text-ds-accent mb-1" data-testid="sakura-human-bonus">
              {bonusList(human.bonuses)}
            </div>
            {human.taken.length === 0 ? (
              <div className="text-xs text-ds-text-muted min-h-[24px]">{t('takenEmpty')}</div>
            ) : (
              <div className="flex gap-0.5 overflow-x-auto justify-center px-2 min-h-[24px]">
                {human.taken.map((c, i) => (
                  <CardImage key={i} card={c} width={cardWidth * 0.42} />
                ))}
              </div>
            )}
          </div>
        )}

        <div className="text-center" data-tutorial="sakura-hand">
          <div className="flex flex-wrap justify-center gap-2">
            {human?.cards.map((c, i) => (
              <button
                key={i}
                type="button"
                onClick={() => onHandClick(i)}
                disabled={!isPlayPhase || !isHumanTurn}
                className={`rounded transition-all ${handIndex === i ? 'ring-2 ring-ds-info -translate-y-2' : ''} ${
                  isPlayPhase && isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'
                }`}
                data-testid={`hand-card-${i}`}
                data-can-capture={(state.captureOptions[i]?.length ?? 0) > 0 || undefined}
              >
                <CardImage card={c} width={cardWidth} />
              </button>
            ))}
          </div>
        </div>

        {/* Turn prompt */}
        {isPlayPhase && (
          <div className="text-center text-sm text-ds-accent" data-testid="sakura-prompt">
            {isHumanTurn ? t('turnYours') : t('turnCpu', { name: turnName })}
          </div>
        )}

        {/* Round-end result */}
        {isRoundEnd && state.lastResult && (
          <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="sakura-round-result">
            <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
            <div className="text-ds-success mb-1">
              {state.lastResult.winner < 0
                ? t('roundResult.draw')
                : t('roundResult.winner', {
                    name: state.players[state.lastResult.winner]
                      ? seatName(state.players[state.lastResult.winner])
                      : '',
                  })}
            </div>
            {state.lastResult.seats.map((s, i) => (
              <div key={i} data-testid={`sakura-round-seat-${i}`}>
                {t('roundResult.seat', {
                  name: state.players[i] ? seatName(state.players[i]) : '',
                  card: s.cardPoints,
                  bonus: s.bonusPoints,
                  total: s.total,
                })}
              </div>
            ))}
          </div>
        )}

        {/* Game-end result */}
        {isGameEnd && (
          <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="sakura-result">
            <div className="mb-1 text-ds-text-primary">{t('result.title')}</div>
            <div className="text-ds-success mb-1">
              {state.winner < 0
                ? t('result.draw')
                : t('result.winner', {
                    name: state.players[state.winner] ? seatName(state.players[state.winner]) : '',
                  })}
            </div>
            {state.players.map((p) => (
              <div key={p.id}>{t('result.score', { name: seatName(p), score: p.score, wins: p.roundWins })}</div>
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
                id: 'seats',
                label: t('settings.seats'),
                value: String(seats),
                options: SEAT_OPTIONS.map((v) => ({ value: String(v), label: String(v) })),
                onSelect: (v: string) => setSeats(Number.parseInt(v, 10)),
              },
              {
                type: 'select' as const,
                id: 'rounds',
                label: t('settings.rounds'),
                value: String(rounds),
                options: ROUND_OPTIONS.map((v) => ({ value: String(v), label: String(v) })),
                onSelect: (v: string) => setRounds(Number.parseInt(v, 10)),
              },
              hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
            ],
          },
        ]}
      />

      <GameFooter className={`${gameTheme.sakura.footer} px-4 py-2.5`}>
        <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="sakura-actions">
          {isRoundEnd && !isGameEnd && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => callApi('next')}
              disabled={loading}
              data-testid="sakura-next-round"
            >
              {t('nextRound')}
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
            dataTutorial="sakura-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
