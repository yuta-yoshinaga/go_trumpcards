import { useEffect } from 'react';
import type { dehlaPakadApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import {
  DEHLA_PAKAD_CPU_DIFFICULTY_OPTIONS,
  DEHLA_PAKAD_KOT_OPTIONS,
  DEHLA_PAKAD_SUITS,
  useDehlaPakadGame,
} from '../hooks/useDehlaPakadGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { DehlaPakadResponse } from '../types/card';
import { DehlaPakadPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { DEHLA_PAKAD_HELP, parseDehlaPakadCommand } from '../utils/cli/commands/dehlaPakadCommands';
import { formatDehlaPakadState } from '../utils/cli/formatters/dehlaPakadFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Dehla Pakad tutorial step definitions. */
const DEHLA_PAKAD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dehlapakad-centre"]',
    messageKey: 'tutorial.centre',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dehlapakad-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dehlapakad-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dehlapakad-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dehlapakad-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type DehlaPakadGame = ReturnType<typeof useDehlaPakadGame>;
type DehlaPakadView = ReturnType<typeof useDehlaPakadView>;
type DehlaPakadContext = GamePageContext<DehlaPakadResponse, DehlaPakadGame, DehlaPakadView>;

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしているので
// ここは直接引く ── 数値化すると全フェーズが NaN に潰れて名前が消える。
const DEHLA_PAKAD_PHASE_KEYS: Readonly<Record<string, string>> = {
  [DehlaPakadPhase.SELECT_TRUMP]: 'selectTrump',
  [DehlaPakadPhase.PLAY]: 'play',
  [DehlaPakadPhase.HAND_END]: 'handEnd',
  [DehlaPakadPhase.GAME_END]: 'gameEnd',
};

function useDehlaPakadView({ game, t }: { game: DehlaPakadGame; t: DehlaPakadContext['t'] }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames: Record<string, string> = Object.fromEntries(
    Object.entries(DEHLA_PAKAD_PHASE_KEYS).map(([phase, key]) => [phase, t(`phase.${key}`)]),
  );

  return { cardWidth, isMobile, phaseNames };
}

function DehlaPakadMessageBox({ state }: Pick<DehlaPakadContext, 'state'>) {
  return <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />;
}

const TRUMP_SUIT_DESIGNS = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const;
const TRUMP_GLYPHS = ['♠', '♣', '♥', '♦'];

function TrumpChoices({ game, state, t, busy }: Pick<DehlaPakadContext, 'game' | 'state' | 't' | 'busy'>) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isTrumpPhase = state.phase === DehlaPakadPhase.SELECT_TRUMP;
  const canCallTrump = isTrumpPhase && isHumanTurn;
  if (!canCallTrump) return null;

  const trumpSuitCounts = DEHLA_PAKAD_SUITS.map(
    (suit) => humanPlayer?.cards.filter((card) => card.design === TRUMP_SUIT_DESIGNS[suit.value - 1]).length ?? 0,
  );

  return (
    <div className="mb-2 p-2 rounded bg-black/30" data-testid="dehlapakad-trump-choices">
      {/* **決めるのは親の右隣で、見えているのは最初の 5 枚だけ。** */}
      <div className="text-ds-text-primary text-sm mb-1">{t('callTrump')}</div>
      <div className="flex flex-wrap gap-3 mb-2" data-testid="dehlapakad-trump-breakdown">
        {DEHLA_PAKAD_SUITS.map((suit, i) => (
          <span role="img" key={suit.value} aria-label={`${t(`suit.${suit.key}`)} ${trumpSuitCounts[i]}`}>
            {TRUMP_GLYPHS[i]}
            {trumpSuitCounts[i]}
          </span>
        ))}
      </div>
      <div className="flex flex-wrap gap-2">
        {DEHLA_PAKAD_SUITS.map((s) => (
          <button
            key={s.value}
            type="button"
            className={btnPrimary}
            onClick={() => game.handleSelectTrump(s.value)}
            disabled={busy}
            data-testid={`dehlapakad-trump-${s.value}`}
          >
            {t(`suit.${s.key}`)}
          </button>
        ))}
      </div>
    </div>
  );
}

function Board({ game, state, t, view, busy }: DehlaPakadContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isTrumpPhase = state.phase === DehlaPakadPhase.SELECT_TRUMP;
  const isHandEnd = state.phase === DehlaPakadPhase.HAND_END;
  const isGameEnd = state.phase === DehlaPakadPhase.GAME_END || state.gameEndFlag;
  const canCallTrump = isTrumpPhase && state.isHumanTurn;
  const tens = state.teamTens ?? [];
  const kots = state.teamKots ?? [];
  const you = state.humanTeam;

  const messageBox = <DehlaPakadMessageBox state={state} />;
  const trumpChoices = <TrumpChoices game={game} state={state} t={t} busy={busy} />;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('hand', { n: state.handNumber, target: state.config.targetKots })}</span>
        {state.trumpSuit > 0 && (
          <>
            <span className="mr-4" data-testid="dehlapakad-trump">
              {t('trump', { suit: t(`suit.${state.trumpSuitName}`) })}
            </span>
            <span>{t('trick', { n: state.trickNumber, total: state.trickCount })}</span>
          </>
        )}
      </div>

      {state.handHistory.length > 0 && (
        <details className="mt-2" open={!isMobile || undefined}>
          <summary className="cursor-pointer select-none text-ds-text-primary text-sm mb-2">
            {t('handHistory.title')}
          </summary>
          <section className="mt-2" data-testid="dehlapakad-hand-history">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-ds-text-primary tabular-nums">
                <thead>
                  <tr className="border-b border-ds-border text-left text-ds-text-muted">
                    <th className="p-2">{t('handHistory.hand')}</th>
                    <th className="p-2">{t('handHistory.winner')}</th>
                    <th className="p-2">{t('handHistory.trump')}</th>
                    <th className="p-2">{t('handHistory.tens')}</th>
                    <th className="p-2">{t('handHistory.kot')}</th>
                  </tr>
                </thead>
                <tbody>
                  {state.handHistory.map((hand, index) => (
                    <tr key={`${index}-${hand.dealerIdx}`} className="border-b border-ds-border-subtle">
                      <td className="p-2">{t('handHistory.handNumber', { n: index + 1 })}</td>
                      <td className="p-2">{t('team', { n: hand.winnerTeam })}</td>
                      <td className="p-2">
                        {t(`suit.${DEHLA_PAKAD_SUITS.find((suit) => suit.value === hand.trumpSuit)?.key ?? 'unknown'}`)}
                      </td>
                      <td className="p-2">
                        {t('handHistory.tensResult', { a: hand.teamTens[0], b: hand.teamTens[1] })}
                      </td>
                      <td className="p-2">{t(hand.kot ? 'handHistory.kotWon' : 'handHistory.noKot')}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        </details>
      )}

      {!isMobile && trumpChoices}
      <div className={lgTwoColGrid}>
        <div data-tutorial="dehlapakad-centre">
          <TrickDisplay
            currentTrick={state.currentTrick}
            lastTrick={state.lastTrick}
            lastTrickWinner={state.lastTrickWinner}
            players={state.players}
            cardWidth={cardWidth}
            label={t(state.currentTrick.length > 0 ? 'currentTrick' : 'lastTrick')}
            dataTutorial="dehlapakad-trick-display"
          />
          {/* **これがこのゲームの心臓部。** 取っただけでは札は手に入らず、
              同じ席が 2 トリック続けて取ってはじめて山ごと引き取れる。
              山に 10 が何枚乗っているかが見えないと、いま何を賭けて
              いるのかが読めない。 */}
          {state.centrePileCount > 0 && (
            <div
              className="mt-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              data-testid="dehlapakad-centre-pile"
            >
              <div className="text-ds-text-primary">
                {t('centrePile', { n: state.centrePileCount, tens: state.centrePileTens })}
              </div>
              {state.prevTrickWinner >= 0 && (
                <div data-testid="dehlapakad-pile-goes-to">
                  {t('pileGoesTo', {
                    name: playerName(state.prevTrickWinner, state.prevTrickWinner === 0),
                  })}
                </div>
              )}
            </div>
          )}
        </div>

        <div data-tutorial="dehlapakad-scores">
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="dehlapakad-scores">
            <div>{t('teamTens', { a: tens[you] ?? 0, b: tens[1 - you] ?? 0 })}</div>
            <div>{t('teamKots', { a: kots[you] ?? 0, b: kots[1 - you] ?? 0 })}</div>
            {/* **7 連勝もコートになる。** 出さないと、なぜ同じ組が勝ち
                続けているのかが数字にならない。 */}
            {state.streakCount > 1 && state.streakTeam >= 0 && (
              <div className="text-ds-accent" data-testid="dehlapakad-streak">
                {t('streak', { team: state.streakTeam, n: state.streakCount })}
              </div>
            )}
          </div>

          <details className="mb-2 p-2 rounded bg-ds-surface text-ds-text-primary" open={!isMobile || undefined}>
            <summary className="cursor-pointer select-none text-sm">{t('earnedTens')}</summary>
            <section className="mt-1" data-testid="dehlapakad-earned-tens">
              {[0, 1].map((team) => {
                const cards = state.players
                  .filter((player) => player.team === team)
                  .flatMap((player) => player.gatheredCards ?? [])
                  .filter((card) => card.value === 10);
                return (
                  <div key={team} className="mb-2" data-testid={`dehlapakad-earned-tens-team-${team}`}>
                    <div className="text-sm">
                      {t('team', { n: team })} ({cards.length})
                    </div>
                    <div className="flex flex-wrap gap-1">
                      {cards.map((card, index) => (
                        <CardImage key={`${card.design}-${card.value}-${index}`} card={card} width={36} />
                      ))}
                    </div>
                  </div>
                );
              })}
            </section>
          </details>

          {humanPlayer && (
            <section
              className="mb-1 flex flex-wrap items-center gap-x-2 text-ds-text-muted text-sm"
              data-testid="dehlapakad-human-seat"
              aria-label={playerName(humanPlayer.id, true)}
            >
              <span className={humanPlayer.team === state.humanTeam ? 'text-ds-success' : ''}>
                {playerName(humanPlayer.id, true)} ({t('team', { n: humanPlayer.team })}):{' '}
                {t('cards', { count: humanPlayer.cardCount })}
              </span>
              {humanPlayer.isDealer && (
                <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
              )}
            </section>
          )}

          <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
            <div className="flex flex-wrap gap-x-4">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div
                    key={p.id}
                    className="text-ds-text-muted text-sm py-0.5 flex items-center gap-2"
                    data-testid={`dehlapakad-cpu-seat-${p.id}`}
                  >
                    <span className={p.team === state.humanTeam ? 'text-ds-success' : ''}>
                      {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}): {t('cards', { count: p.cardCount })}
                    </span>
                    {p.isDealer && (
                      <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
                    )}
                  </div>
                ))}
            </div>
          </CpuAccordion>

          {/* 相手が選んでいる間の空白を埋める。PiedmonteseTarot の
              scartoWaiting と同じ形 (#6626)。 */}
          {isTrumpPhase && !canCallTrump && (
            <div className="mb-2 text-center text-sm text-ds-text-muted" data-testid="dehlapakad-trump-waiting">
              {t('trumpWaiting', {
                name: playerName(state.trumpChooserIdx, state.players[state.trumpChooserIdx]?.isHuman === true),
              })}
            </div>
          )}

          {state.lastHand && (isHandEnd || isGameEnd) && (
            <div
              className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              data-testid="dehlapakad-hand-result"
            >
              <div className="text-ds-text-primary">
                {t('handResult', {
                  team: state.lastHand.winnerTeam,
                  a: state.lastHand.teamTens[0],
                  b: state.lastHand.teamTens[1],
                })}
              </div>
              {isHandEnd && state.lastTrickWinner >= 0 && (
                <div data-testid="dehlapakad-last-trick-winner">
                  {t('lastTrickWinner', {
                    name: playerName(state.lastTrickWinner, state.players[state.lastTrickWinner]?.isHuman === true),
                    team: state.players[state.lastTrickWinner]?.team ?? -1,
                  })}
                </div>
              )}
              {state.lastHand.kot && (
                <div className="text-ds-warning" data-testid="dehlapakad-kot">
                  {t(`kot.${state.lastHand.kotReason}`)}
                </div>
              )}
            </div>
          )}

          {isGameEnd && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="dehlapakad-winner">
              {t('winner', { team: state.winnerTeam })}
            </div>
          )}
        </div>
      </div>

      {isMobile && messageBox}
      {isMobile && trumpChoices}
    </>
  );
}

function FooterHand({ game, state, t, view }: DehlaPakadContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const isHumanTurn = state.isHumanTurn;
  const isPlayPhase = state.phase === DehlaPakadPhase.PLAY;
  const canPlay = isPlayPhase && isHumanTurn;
  const handValidIndices = canPlay ? state.playableIndices : undefined;

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="dehlapakad"
      validIndices={handValidIndices}
      legalIndices={handValidIndices}
      restrictedTooltip={t('restrictedTooltip')}
    />
  );
}

function FooterStatus({ state, t, frontendHint }: DehlaPakadContext) {
  return (
    <>
      {/* ライブ領域は常設 (#5955)。 */}
      <div data-testid="dehlapakad-hint-live" role="status" aria-live="polite">
        {isRequestedHint(state) && (state.hint || state.hintTrumpSuit > 0) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}
            {state.hint && `: ${t(`hint.${state.hint.reason}`)}`}
            {state.hint?.cardIndices &&
              state.hint.cardIndices.length > 0 &&
              ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: DehlaPakadContext) {
  const isHumanTurn = state.isHumanTurn;
  const isPlayPhase = state.phase === DehlaPakadPhase.PLAY;
  const isHandEnd = state.phase === DehlaPakadPhase.HAND_END;
  const canPlay = isPlayPhase && isHumanTurn;

  return (
    <>
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
          data-testid="dehlapakad-play"
        >
          {t('playButton')}
        </button>
      )}
      {isHandEnd && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleNextHand}
          disabled={busy}
          data-testid="dehlapakad-next-hand"
        >
          {t('nextHand')}
        </button>
      )}
    </>
  );
}

const dehlaPakadDefinition: GameDefinition<
  DehlaPakadResponse,
  DehlaPakadGame,
  Parameters<typeof dehlaPakadApi.exec>,
  DehlaPakadView
> = {
  key: 'dehlapakad',
  useGame: useDehlaPakadGame,
  useView: useDehlaPakadView,
  tutorial: DEHLA_PAKAD_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 13 },
  cli: () => ({
    parseCommand: parseDehlaPakadCommand,
    formatResponse: formatDehlaPakadState,
    helpText: DEHLA_PAKAD_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase] ?? view.phaseNames[DehlaPakadPhase.PLAY],
    ended: state.phase === DehlaPakadPhase.GAME_END || state.gameEndFlag,
    won: (state.phase === DehlaPakadPhase.GAME_END || state.gameEndFlag) && state.winnerTeam === state.humanTeam,
  }),
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn,
  }),
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.dehlaPakadConfig.cpuDifficulty,
          options: DEHLA_PAKAD_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetKots',
          label: t('settings.targetKots'),
          value: game.dehlaPakadConfig.targetKots,
          options: DEHLA_PAKAD_KOT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetKots', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) => (!view.isMobile ? <DehlaPakadMessageBox state={state} /> : null),
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'dehlapakad-action-buttons',
  resetTutorialId: 'dehlapakad-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/**
 * Renders the Dehla Pakad page: the Indian "catch the tens" partnership
 * trick-taker where cards are gathered only on two consecutive tricks.
 */
export const DehlaPakadPage = createGamePage(dehlaPakadDefinition);
