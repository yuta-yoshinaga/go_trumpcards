# Cloudflare Workers (WASM)

Games are deployed to Cloudflare Workers as WASM binaries via TinyGo. **Ten** workers split games into size buckets to keep each binary under the free-tier 1 MB gzipped limit. The fourth, **`extra`**, was added in [ADR-0032](adr/0032-fourth-worker-capacity.md) once the original three approached the limit, **`extra2`/`extra3`** in [ADR-0036](adr/0036-fifth-sixth-worker-capacity.md), **`extra4`** in [ADR-0037](adr/0037-seventh-worker-capacity.md), **`extra5`** in [ADR-0038](adr/0038-eighth-worker-capacity.md), and **`extra6`/`extra7`** in ADR-0041. Like the others, the `Category` is purely a binary-size bucket, **not** a user-facing taxonomy. **No bucket is *the* overflow one:** a new game goes into whichever worker currently has the most gzip headroom, measured rather than assumed.

Each row lists the games' **registry keys** verbatim, so the table can be checked
mechanically -- `TestDocsMatchRegistry` in `internal/infrastructure/games` fails if
this table and `registry.go` disagree. It is generated, not curated: the previous
hand-written prose drifted repeatedly (it still named three workers after the fourth
shipped), which is why the lists are now flat and guarded.

A bucket is chosen purely by binary size. Nothing about a game's genre puts it in a
particular worker, and games move between buckets whenever one approaches the limit.

| Worker | Entry point | Games | Registry keys |
|--------|-------------|-------|---------------|
| **casino** | `cmd/workers/casino/main.go` | 51 | `baccarat`, `badugi`, `banluck`, `baseballpoker`, `bigo`, `bigohilo`, `blackjack`, `blackjackswitch`, `caribbeanstud`, `casinoholdem`, `chicago`, `chinesepoker`, `cincinnati`, `courchevel`, `courchevelhilo`, `crazypineapple`, `deuceswild`, `deucetoseven`, `doubleattack`, `doubleexposure`, `dramaha`, `eightgame`, `fivecardstud`, `followthequeen`, `fourcardpoker`, `freebet`, `holdem`, `horse`, `indianpoker`, `irishpoker`, `ironcross`, `jokerpoker`, `omaha`, `omahahilo`, `openfacechinese`, `pineapple`, `poker`, `razz`, `russianpoker`, `sevencardstud`, `sevencardstudhilo`, `shortdeck`, `soko`, `spanish21`, `teenpatti`, `texasholdembonus`, `threecard`, `threecardbrag`, `threecardrummy`, `ultimatetexasholdem`, `videopoker` |
| **classic** | `cmd/workers/classic/main.go` | 40 | `allfours`, `botifarra`, `briscola`, `callbreak`, `cassino`, `catchten`, `colorado`, `crazyeights`, `cucumber`, `curdsandwhey`, `durak`, `egyptianratscrew`, `ginrummy`, `karnoffel`, `knockoutwhist`, `labellelucie`, `marias`, `nap`, `ninetynine`, `ohhell`, `oldmaid`, `pageone`, `president`, `prsi`, `reversis`, `royalcotillion`, `sedma`, `sevens`, `shamrocks`, `shithead`, `simplesimon`, `slapjack`, `solowhist`, `spades`, `spoilfive`, `tonk`, `truco`, `twotenjack`, `unsunkaruta`, `whist` |
| **solo** | `cmd/workers/solo/main.go` | 47 | `accordion`, `acesup`, `bakersdozen`, `bakersgame`, `beleagueredcastle`, `blackhole`, `bristol`, `calculation`, `canfield`, `citadel`, `clocksolitaire`, `crazyquilt`, `crescent`, `cruel`, `easthaven`, `eightoff`, `fortress`, `fortythieves`, `fourseasons`, `freecell`, `gaps`, `golf`, `klondike`, `memory`, `montecarlo`, `oasispoker`, `osmosis`, `penguin`, `pokersquares`, `pyramid`, `russiansolitaire`, `schnapsen`, `scorpion`, `seahaventowers`, `snap`, `somerset`, `spider`, `spiderette`, `stalactites`, `tienlen`, `tripeaks`, `wasp`, `whitehead`, `willothewisp`, `yaniv`, `yukon`, `zheng` |
| **extra** | `cmd/workers/extra/main.go` | 33 | `bauernschnapsen`, `calabresella`, `carioca`, `cinch`, `contractrummy`, `diplomat`, `frenchtarot`, `gaigel`, `ganjifa`, `gostop`, `hachihachi`, `indianrummy`, `kalooki`, `king`, `kingalbert`, `koenigrufen`, `machiavelli`, `matrimony`, `pan`, `pontoon`, `quinze`, `rummy500`, `settemezzo`, `sthelena`, `streetsandalleys`, `sultan`, `tapptarock`, `threethirteen`, `troggu`, `tysiac`, `vira`, `watten`, `zwanzigerrufen` |
| **extra2** | `cmd/workers/extra2/main.go` | 41 | `aluette`, `americantoad`, `auldlangsyne`, `baccaratbanque`, `baloot`, `beggarmyneighbour`, `bideuchre`, `bigtwo`, `bisley`, `braid`, `cribbagesquares`, `cuarenta`, `doubleklondike`, `duchess`, `fiftyone`, `gofish`, `grandfathersclock`, `julepe`, `kemps`, `laughandliedown`, `loba`, `missmilligan`, `napoleonssquare`, `pigtail`, `pishti`, `polignac`, `rams`, `rikken`, `ristikontra`, `sirtommy`, `sixcardgolf`, `sjavs`, `slyfox`, `speed`, `spiteandmalice`, `tehonbiki`, `trappola`, `trash`, `war`, `windmill`, `zwicker` |
| **extra3** | `cmd/workers/extra3/main.go` | 32 | `basra`, `bigben`, `boston`, `bouillotte`, `bridge`, `bura`, `cirulla`, `congress`, `desmoche`, `fortyandeight`, `kaiser`, `kille`, `koikoi`, `madrasso`, `mao`, `nainjaune`, `niuniu`, `poch`, `popejoan`, `primero`, `rollingstone`, `rook`, `sakura`, `saliclaw`, `sevenbridge`, `skitgubbe`, `stealingbundles`, `tablanet`, `terrace`, `toepen`, `trex`, `vint` |
| **extra4** | `cmd/workers/extra4/main.go` | 36 | `alaska`, `anaconda`, `barbu`, `basset`, `bezique`, `bhabhi`, `chemindefer`, `colourwhist`, `dragontiger`, `estimation`, `flowergarden`, `fourteenout`, `gongzhu`, `highcardflush`, `honeymoonbridge`, `israeliwhist`, `letitride`, `lingerlonger`, `literature`, `mendikot`, `michigan`, `mrsmop`, `narcotic`, `ombre`, `perseverance`, `piedmontesetarot`, `piquet`, `preference`, `put`, `rankandfile`, `russianbank`, `scarto`, `sergeantmajor`, `shengji`, `sixbidsolo`, `trenteetquarante` |
| **extra5** | `cmd/workers/extra5/main.go` | 31 | `agnes`, `batak`, `binokel`, `cego`, `comet`, `continentalrummy`, `dehlapakad`, `diloti`, `doubt`, `gleek`, `goofspiel`, `hasenpfeffer`, `kingo`, `loo`, `macau`, `minibridge`, `napoleon`, `nertz`, `oichokabu`, `omi`, `pinochle`, `pitch`, `quadrille`, `quodlibet`, `shelem`, `speculation`, `teendopaanch`, `tongits`, `tusac`, `ulti`, `wizard` |
| **extra6** | `cmd/workers/extra6/main.go` | 36 | `belote`, `bourre`, `caribbeandraw`, `chineseten`, `coinche`, `costlycolours`, `courtpiece`, `cribbage`, `cuckoo`, `daifugo`, `doppelkopf`, `ecarte`, `faro`, `fortyfives`, `jass`, `klaberjass`, `marjapussi`, `marriage`, `mighty`, `mississippistud`, `montebank`, `mus`, `mushi`, `paigow`, `pig`, `ramsch`, `skat`, `spoons`, `sueca`, `sutda`, `tarabish`, `tarneeb`, `tichu`, `tressette`, `tute`, `twentynine` |
| **extra7** | `cmd/workers/extra7/main.go` | 36 | `andarbahar`, `bidwhist`, `biriba`, `bolivia`, `brusquembille`, `burraco`, `canasta`, `casinowar`, `chinchon`, `conquian`, `crazyfourpoker`, `doudizhu`, `escoba`, `euchre`, `fivehundred`, `germansolo`, `germanwhist`, `guandan`, `guts`, `handandfoot`, `hearts`, `hokm`, `klaverjas`, `manille`, `minchiate`, `pasur`, `reddog`, `samba`, `schafkopf`, `scopa`, `scopone`, `seventwentyseven`, `sheepshead`, `slobberhannes`, `tarocchini`, `thirtyone` |

The worker entry points (`cmd/workers/{casino,classic,solo,extra,extra2,extra3,extra4,extra5,extra6,extra7}/main.go`) are thin shells that blank-import the matching `internal/infrastructure/games/<category>` sub-package and call `games.RegisterCategory(mux, games.Category…)`. The registry itself (`internal/infrastructure/games/registry.go`) stores `{Name, Category}` for each game; the human-readable descriptions live in a separate `descriptions.go` map (build-tagged `//go:build !js || !wasm` to keep them out of the WASM binaries). The Web-server factories live in `games_server.go` (excluded from WASM via build tags) and the Worker bindings live in per-category sub-packages — this split is what keeps each Cloudflare Worker binary under the 1 MB gzipped free-tier limit by letting TinyGo dead-code-eliminate the games from the other categories.

**When adding/modifying a game, always update:**
1. `internal/infrastructure/games/registry.go` — `{Name, Category}` entry (selects the worker), plus the matching CLI display title in `descriptions.go`
2. `internal/infrastructure/games/games_server.go` — `BindWebControllerFor("<name>", …)` for the HTTP server factory
3. `internal/infrastructure/games/{casino,classic,solo,extra,extra2,extra3,extra4,extra5,extra6,extra7}/<category>.go` — `games.RegisterKVGame("<name>", games.Category…, …)` for the KV-backed worker route (must match the `Category`)
4. `internal/infrastructure/ui/GameManager.go` — `gameRegistry` entry for the CLI wiring
5. `frontend/src/api/gameExec.ts` `workerUrl` (re-exported by `frontend/src/api/gameApi.ts`) — must match the `Category`

Build: `make build-worker-{solo,casino,classic,extra,extra2,extra3,extra4,extra5,extra6,extra7}` or `make build-workers` (requires TinyGo).

## ローカルでサイズを実測する

ADR-0032 の時点では TinyGo をローカルに持たず CI のレポート頼りだったが、いまは手元で測れる
（Go 1.25.8 + TinyGo 0.42.0、CI と同じ組み合わせ）。

```sh
export PATH="$HOME/sdk/go1.25.8/bin:$HOME/.local/opt/tinygo/bin:$PATH"
export GOTOOLCHAIN=local            # ← 必須
mkdir -p workers/<w>/build
go run github.com/syumai/workers/cmd/workers-assets-gen -mode=tinygo -o workers/<w>/build
tinygo build -tags <w> -o workers/<w>/build/app.wasm -target wasm \
  -stack-size=128KB -no-debug -opt=z ./cmd/workers/<w>
wasm-opt --enable-bulk-memory --enable-nontrapping-float-to-int --enable-sign-ext \
  -Oz workers/<w>/build/app.wasm -o workers/<w>/build/app.wasm
gzip -c workers/<w>/build/app.wasm | wc -c    # 1,048,576 と比較する
```

- **`GOTOOLCHAIN=local` を忘れないこと。** `go.mod` の `toolchain go1.26.0` により Go 1.25 でも
  Go 1.26 に自動アップグレードされ、CI（Go 1.25）とは別のツールチェーンでビルドされてサイズが一致しなくなる。
  CI も同じ理由で明示している。
- **`wasm-opt` 前の値で判断しない。** extra は 1,077,248 → 1,029,817 と 47 KB 縮む。最適化前だと
  上限超過に見える。
- `make` が無い環境では上のコマンドが Makefile レシピの展開そのもの。
