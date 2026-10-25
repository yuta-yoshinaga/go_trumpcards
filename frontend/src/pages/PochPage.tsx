import type { TFunction } from 'i18next';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { pochApi } from '../api/gameApi';
import { CardBackFan } from '../components/CardBackFan';
import { CardBack } from '../components/CardImage';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { usePochGame } from '../hooks/usePochGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { PochResponse } from '../types/card';
import { PochPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { POCH_HELP, parsePochCommand } from '../utils/cli/commands/pochCommands';
import { formatPochState } from '../utils/cli/formatters/pochFormatter';
import { hintCheckboxItem } from '../utils/settingsItems';

const PC_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="pc-rule"]', messageKey: 'tutorial.paySuit', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pc-board"]', messageKey: 'tutorial.carryOver', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pc-rule"]', messageKey: 'tutorial.pochen', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pc-hand"]', messageKey: 'tutorial.stops', placement: 'top', advanceOn: 'next' },
];

/**
 * Poch game hook plus the "your turn" announcement.
 *
 * The announcement is set from the game's own onSuccess and cleared by every human
 * action, so it lives here (before the view hook can exist) and survives CLI toggling.
 */
function usePochPageGame() {
  const { t } = useTranslation('poch');
  const [turnAnnouncement, setTurnAnnouncement] = useState('');
  const onSuccess = useCallback(
    (response: PochResponse, [command]: Parameters<typeof pochApi.exec>) => {
      if (
        command !== 'reset' &&
        (response.phase === PochPhase.POCHEN || response.phase === PochPhase.STOPS) &&
        response.currentPlayerIdx === 0
      ) {
        setTurnAnnouncement(t('yourTurnAnnouncement'));
      }
    },
    [t],
  );
  const game = usePochGame({ onSuccess });
  const execHumanAction = (...args: Parameters<typeof game.exec>) => {
    if (args[0] === 'reset') return game.exec(...args);
    setTurnAnnouncement('');
    return game.exec(...args);
  };
  return { ...game, exec: execHumanAction, turnAnnouncement };
}

type PochGame = ReturnType<typeof usePochPageGame>;
type PochContext = GamePageContext<PochResponse, PochGame, PochView>;

function usePochView({ game, t }: { game: PochGame; t: TFunction }) {
  const { state } = game;
  const isMobile = useIsMobile();
  const { cardWidth } = useCardDimensions();
  const [handIdx, setHandIdx] = useState<number | null>(null);
  const previousPools = useRef<PochResponse['pools'] | null>(null);
  const previousStakingAwards = useRef<PochResponse['stakingAwards'] | null>(null);
  const [poolAnnouncement, setPoolAnnouncement] = useState('');
  const [stakingAnnouncement, setStakingAnnouncement] = useState('');

  const resetWithQuietPoolBaseline = (reset: () => void) => {
    previousPools.current = null;
    previousStakingAwards.current = null;
    setPoolAnnouncement('');
    setStakingAnnouncement('');
    reset();
  };

  useEffect(() => {
    if (!state) return;
    if (previousPools.current === null || previousStakingAwards.current === null) {
      previousPools.current = state.pools;
      previousStakingAwards.current = state.stakingAwards;
      return;
    }

    const previousByName = new Map(previousPools.current.map((pool) => [pool.name, pool.chips]));
    const changedPools = state.pools
      .filter((pool) => previousByName.get(pool.name) !== pool.chips)
      .map((pool) => t('poolUpdate', { pool: t(`pool.${pool.name}`), chips: pool.chips }));
    previousPools.current = state.pools;
    setPoolAnnouncement(changedPools.join(t('listSeparator')));

    const previousAwards = new Map(
      previousStakingAwards.current.map((award) => [award.pool, `${award.player}:${award.chips}`]),
    );
    const newAwards = state.stakingAwards
      .filter((award) => previousAwards.get(award.pool) !== `${award.player}:${award.chips}`)
      .map((award) =>
        award.player === 0
          ? t('awardLineYou', { pool: t(`pool.${award.pool}`), chips: award.chips })
          : t('awardLine', { player: award.player, pool: t(`pool.${award.pool}`), chips: award.chips }),
      );
    previousStakingAwards.current = state.stakingAwards;
    setStakingAnnouncement(newAwards.join(t('listSeparator')));
  }, [state, t]);

  const phaseNames = {
    end: t('phase.end'),
    dealEnd: t('phase.dealEnd'),
    pochen: t('phase.pochen'),
    stops: t('phase.stops'),
  };

  return {
    isMobile,
    cardWidth,
    handIdx,
    setHandIdx,
    poolAnnouncement,
    stakingAnnouncement,
    resetWithQuietPoolBaseline,
    phaseNames,
  };
}

type PochView = ReturnType<typeof usePochView>;

function HeaderExtra({ state, t }: PochContext) {
  return (
    <span className="text-sm text-ds-text-muted">{t('deal', { n: state.dealNo + 1, target: state.targetDeals })}</span>
  );
}

function LiveRegion({ game, t, view }: PochContext) {
  return (
    <>
      <LandscapeBanner message={t('landscapeBanner')} />
      <div role="status" aria-live="polite" className="sr-only" data-testid="poch-pool-live">
        {[view.poolAnnouncement, view.stakingAnnouncement, game.turnAnnouncement]
          .filter(Boolean)
          .join(t('listSeparator'))}
      </div>
    </>
  );
}

function Board({ state, t, view, frontendHint }: PochContext) {
  const { isMobile, cardWidth, handIdx, setHandIdx } = view;
  const frontendHintEnabled = frontendHint.hintEnabled;
  const ended = state.phase === PochPhase.GAME_END;
  const dealOver = state.phase === PochPhase.DEAL_END;
  const betting = state.phase === PochPhase.POCHEN;
  const playing = state.phase === PochPhase.STOPS;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  return (
    <>
      {/* Permanent, not tutorial-only: "only the pay suit pays" and
          "pochen is a comparison, not a declaration" are what a player
          gets wrong. */}
      <details
        open={!isMobile || undefined}
        className="mb-3 text-center"
        data-testid="poch-rules"
        data-tutorial="pc-rule"
      >
        <summary className="cursor-pointer select-none text-xs text-ds-warning font-medium">{t('rules')}</summary>
        <div className="pt-1 text-xs text-ds-warning font-medium">{t('ruleLine')}</div>
      </details>

      <div className="text-center mb-3">
        <div className="text-ds-text-muted text-xs mb-1">{t('turnUp')}</div>
        {state.turnUp ? (
          <fieldset
            aria-label={t('paySuitAriaLabel', { suit: t(`suit.${state.turnUp.design}`) })}
            className="border-0 p-0 m-0 min-w-0 flex flex-col items-center gap-1"
          >
            <div className="flex justify-center">
              <AnimatedCard card={state.turnUp} width={cardWidth} draggable={false} />
            </div>
            <span className="text-ds-text-primary text-xs font-medium">
              {t('paySuit', { suit: t(`suit.${state.turnUp.design}`) })}
            </span>
          </fieldset>
        ) : (
          <span className="text-game-text-muted text-xs">—</span>
        )}
      </div>

      {/* **9 区画は毎回すべて出す。**膨らんでいる区画がそのディールの
          狙いどころそのものになる。 */}
      <div className="mb-4" data-tutorial="pc-board">
        <div className="text-game-text-muted text-xs mb-1 text-center">{t('board')}</div>
        <div className="flex gap-2 justify-center flex-wrap">
          {state.pools.map((pool) => (
            <div
              key={`pool-${pool.name}`}
              data-testid="poch-pool"
              className="rounded px-2 py-1 text-xs ring-1 ring-ds-border text-center min-w-14"
            >
              <div className="text-ds-text-muted">{t(`pool.${pool.name}`)}</div>
              <div className="font-medium">{pool.chips}</div>
            </div>
          ))}
        </div>
      </div>

      {/* 第 1 段階は自動で解決するので、結果を出さないと何が起きたのか読めない。
          席 0 は「あなた」と呼ぶ ── 手札セクションが既にそうしているのに、
          結果通知 3 種だけが「席0が」と出ていた (#6519)。 */}
      {state.stakingAwards.length > 0 && (
        <div className="text-center text-xs mb-3" data-testid="poch-staking">
          {state.stakingAwards.map((a) => (
            <div key={`award-${a.pool}`}>
              {a.player === 0
                ? t('awardLineYou', { pool: t(`pool.${a.pool}`), chips: a.chips })
                : t('awardLine', { player: a.player, pool: t(`pool.${a.pool}`), chips: a.chips })}
            </div>
          ))}
        </div>
      )}

      <CpuAccordion playerCount={opponents.length} openOnDesktop={false}>
        <div className="flex justify-center gap-4 mb-3 flex-wrap">
          {opponents.map((o) => (
            <div key={`opp-${o.id.toString()}`} className="text-center">
              <div className="text-game-text-muted text-xs mb-1">
                {t('seat', { name: `CPU${o.id.toString()}`, chips: o.chips, n: o.cardCount })}
                {o.folded && ` · ${t('folded')}`}
              </div>
              <div
                className="flex gap-1 justify-center flex-wrap"
                {...(!state.gameEndFlag && {
                  role: 'img',
                  'aria-label': t('opponentHandAriaLabel', {
                    name: `CPU${o.id.toString()}`,
                    n: o.cardCount,
                  }),
                })}
              >
                {state.gameEndFlag ? (
                  o.cards.map((card, i) => (
                    <AnimatedCard
                      key={`opp-${o.id.toString()}-c${i.toString()}`}
                      card={card}
                      width={cardWidth}
                      draggable={false}
                    />
                  ))
                ) : isMobile ? (
                  Array.from({ length: o.cardCount }, (_, i) => (
                    <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                  ))
                ) : (
                  <CardBackFan count={o.cardCount} width={Math.round(cardWidth * 0.45)} />
                )}
              </div>
            </div>
          ))}
        </div>
      </CpuAccordion>

      {state.playedPile.length > 0 && (
        <div className="text-center mb-3">
          <div className="text-game-text-muted text-xs mb-1">
            {t('played')}
            {/* 止まっているかどうかで出せる札がまるで違う。 */}
            {state.stopsSuit < 0 && ` · ${t('runStopped')}`}
          </div>
          <div className="flex gap-1 justify-center flex-wrap">
            {state.playedPile.slice(-8).map((card, i) => (
              <AnimatedCard
                key={`pile-${i.toString()}`}
                card={card}
                width={Math.round(cardWidth * 0.7)}
                draggable={false}
              />
            ))}
          </div>
        </div>
      )}

      {dealOver && state.dealWinner >= 0 && (
        <div className="text-center text-sm mb-3" data-testid="poch-deal-result">
          {state.dealWinner === 0 ? t('dealResultYou') : t('dealResult', { n: state.dealWinner })}
        </div>
      )}
      {playing && state.pochenWinner >= 0 && (
        <div className="text-center text-sm mb-3" data-testid="poch-pochen-result">
          {state.pochenWinner === 0
            ? t('pochenResultYou', { pot: state.pochenPot })
            : t('pochenResult', { n: state.pochenWinner, pot: state.pochenPot })}
        </div>
      )}

      <div className="text-center" data-tutorial="pc-hand">
        <div className="text-game-text-muted text-xs mb-1">
          {t('yourHand')}
          {' · '}
          {t('yourChips', { n: human?.chips ?? 0 })}
          {betting && ` · ${t('yourBet', { n: human?.bet ?? 0, target: state.betTarget })}`}
          {/* pochen は組の比べ合い。自分の組は自分にとって既知の情報なので、
              賭ける前に見せる (#5722)。判定はサーバ (PochBestCombo) が唯一の出どころ。 */}
          {betting && (
            <span data-testid="poch-your-combo">
              {' · '}
              {state.yourBestComboSize > 0
                ? t('yourCombo', { size: state.yourBestComboSize, rank: state.yourBestComboRank })
                : t('yourComboNone')}
            </span>
          )}
        </div>
        <div className="flex gap-1 justify-center flex-wrap">
          {(human?.cards ?? []).map((card, i) => {
            // **並びに従う義務がある。**出せない札を押せてしまうと、
            // サーバに弾かれて初めて分かる (#4933)。空リストは「情報が
            // 無い」なので制限しない。
            const restricting = isHumanTurn && playing && (state.validPlays?.length ?? 0) > 0;
            const canPlay = !restricting || state.validPlays.includes(i);
            return (
              <button
                key={`hand-${i.toString()}`}
                type="button"
                data-hint-action="play"
                data-unplayable={restricting && !canPlay ? 'true' : undefined}
                aria-pressed={handIdx === i}
                disabled={restricting && !canPlay}
                aria-disabled={!isHumanTurn || !playing || !canPlay}
                onClick={() => isHumanTurn && playing && canPlay && setHandIdx(handIdx === i ? null : i)}
                className={[
                  'rounded transition-transform',
                  restricting && !canPlay ? 'opacity-40' : '',
                  isHumanTurn && playing ? 'hover:-translate-y-2' : 'opacity-60',
                  handIdx === i ? 'ring-2 ring-ds-accent -translate-y-2' : '',
                  frontendHintEnabled && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                ].join(' ')}
              >
                <AnimatedCard card={card} width={cardWidth} draggable={false} />
              </button>
            );
          })}
        </div>
        {isHumanTurn && playing && (
          <div className="text-[10px] text-ds-text-muted mt-1">
            {state.stopsSuit < 0 ? t('selectFreeLead') : t('selectFollow')}
          </div>
        )}
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterActions({ game, state, t, view }: PochContext) {
  const { handIdx, setHandIdx } = view;
  const ended = state.phase === PochPhase.GAME_END;
  const dealOver = state.phase === PochPhase.DEAL_END;
  const betting = state.phase === PochPhase.POCHEN;
  const playing = state.phase === PochPhase.STOPS;
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;
  return (
    <>
      {isHumanTurn && betting && (
        <>
          <button
            type="button"
            data-hint-action="bet"
            className={`${btnPrimary} min-h-11`}
            onClick={() => game.exec('bet')}
          >
            {t('bet')}
          </button>
          <button
            type="button"
            data-hint-action="fold"
            className={`${btnSecondary} min-h-11`}
            onClick={() => game.exec('fold')}
          >
            {t('fold')}
          </button>
        </>
      )}
      {isHumanTurn && playing && (
        <button
          type="button"
          className={`${btnPrimary} min-h-11`}
          disabled={handIdx === null}
          onClick={() => {
            if (handIdx !== null) {
              game.exec('play', handIdx);
              setHandIdx(null);
            }
          }}
        >
          {t('play')}
        </button>
      )}
      {dealOver && !ended && (
        <button type="button" className={`${btnPrimary} min-h-11`} onClick={() => game.exec('next')}>
          {t('nextDeal')}
        </button>
      )}
    </>
  );
}

const pochDefinition: GameDefinition<PochResponse, PochGame, Parameters<typeof pochApi.exec>, PochView> = {
  key: 'poch',
  useGame: usePochPageGame,
  useView: usePochView,
  tutorial: PC_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 3, tableau: 4 },
  cli: () => ({
    parseCommand: parsePochCommand,
    formatResponse: formatPochState,
    helpText: POCH_HELP,
  }),
  phase: ({ state, view }) => {
    const ended = state.phase === PochPhase.GAME_END;
    const dealOver = state.phase === PochPhase.DEAL_END;
    const betting = state.phase === PochPhase.POCHEN;
    return {
      name: ended
        ? view.phaseNames.end
        : dealOver
          ? view.phaseNames.dealEnd
          : betting
            ? view.phaseNames.pochen
            : view.phaseNames.stops,
      ended,
    };
  },
  reset: ({ game, state, view }) =>
    view.resetWithQuietPoolBaseline(() =>
      game.exec('reset', undefined, { cpuDifficulty: state.config?.cpuDifficulty ?? 1 }),
    ),
  settings: ({ game, state, t, tc, view, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'poch-difficulty',
          label: t('settings.cpuDifficulty'),
          value: String(state.config?.cpuDifficulty ?? 1),
          options: [
            { value: '0', label: t('settings.difficulty.easy') },
            { value: '1', label: t('settings.difficulty.normal') },
            { value: '2', label: t('settings.difficulty.hard') },
          ],
          onSelect: (v: string) =>
            view.resetWithQuietPoolBaseline(() => game.exec('reset', undefined, { cpuDifficulty: Number(v) })),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  settingsInCliMode: true,
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  resetTutorialId: 'pc-reset-button',
  HeaderExtra,
  LiveRegion,
  Board,
  FooterActions,
};

/** Renders the Poch page: nine pools, a pay suit, a set comparison and stops. */
export const PochPage = createGamePage(pochDefinition);
