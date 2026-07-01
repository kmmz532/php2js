# PukiWiki for Cloudflare Workers (php2js) 開発ドキュメント

## 現在の進捗と状況

現在、PukiWikiをCloudflare Workers上で動作させるためのPHPからJavaScriptへのトランスパイラ（`php2js`）の実装を進めています。
PukiWiki (v1.5.4) のコアファイル114個のトランスパイル自体はエラーなく完了する状態に到達しています。

### 達成した機能
1. **ASTベースのトランスパイル**: PHPのソースコードをパースし、JSのASTに変換してコード生成を行うパイプラインの構築。
2. **ランタイムの構築**: PHPの組み込み関数（`preg_match`、`filemtime`、`bin2hex`など）、スーパーグローバル変数（`$_GET`, `$_POST`, `$_SERVER`など）、および定数展開機能を提供するJavaScriptレイヤーの実装。
3. **同期・非同期の自動解決**: PHPでは同期的なI/O（`file_get_contents`や`require`など）を、JS側で`async/await`を伝播させる仕組みの導入。
4. **名前空間・変数スコープの分離**: PHPのグローバル変数を`__runtime.GLOBALS`にマッピングし、ユーザー定義関数は`globalThis.function_name`経由で呼び出すことで、JSのブロックスコープ（`let`）による変数・関数名の衝突（例: `is_page`が関数と変数で重複する問題）を解決。
5. **モジュールの複数回評価のサポート**: CF Workersの仕様（`import()`のキャッシュ）に対応するため、各トランスパイル済みファイルのトップレベルコードを`export default async function __main()`内にラップし、リクエストのたびにPHPのように再評価できるようにしました。

### 現在の課題とデバッグ中の箇所
- **関数呼び出しの解決エラー**: 現在、`globalThis.filemtime is not a function`などのエラーをデバッグ中です。一部のファイルI/O系関数や`preg_replace_callback`の配列コールバック（`[$this, 'method']`）などのランタイム関数の対応漏れを追加で実装・修正しています。
- **画面出力（レンダリング）**: `catbody`などが実行されHTMLが出力されるフローの確立を目指していますが、現在CF Workersのレスポンスが空で返ってくる問題に対処しています。

---

## ディレクトリとファイル構成の説明

このプロジェクトはGo言語で実装されたトランスパイラ（`php2js`）と、生成されるJavaScriptのランタイムで構成されています。

### 1. トランスパイラ (`internal/` 以下のGo言語実装)

- **`internal/parser/`**: PHPソースコードを構文解析し、PHPのAST（抽象構文木）を生成するモジュールです。（外部ライブラリの`z7zmey/php-parser`等を利用）
- **`internal/transformer/`**: 
  - `transformer.go`, `expr.go`, `stmt.go`: PHPのASTをJavaScriptのAST（`jsast`）に変換するコアロジックです。変数のスコープ解決、`async`の伝播、クラスや関数の変換を担います。
  - `helpers.go`: `mapFunctionName`などを定義し、PHPの組み込み関数をJSのランタイム（`__runtime`）や`globalThis`にマッピングする処理を行います。
- **`internal/generator/`**:
  - `generator.go`: JSのASTを元に、実際のJavaScriptのソースコード文字列を生成します。ここで各ファイルのトップレベルコードを`__main()`関数でラップしています。
- **`internal/jsast/`**: JSの抽象構文木（AST）の構造体定義をまとめたパッケージです。
- **`internal/runtime/`**:
  - `embed.go`: `js/`ディレクトリ以下のランタイムJavaScriptファイルをGoバイナリに埋め込み、トランスパイル時に出力ディレクトリ（`output/`）へコピーするための処理です。
- **`internal/worker/`**:
  - `worker.go`: CF Workersの起点となる`index.js`（リクエストの受け口、`fetch`ハンドラ）のテンプレート文字列を保持・生成します。
- **`cmd/php2js/main.go`**: トランスパイラのCLIエントリポイントです。引数を解釈し、各モジュールを呼び出してトランスパイルを実行します。

### 2. JavaScriptランタイム (`internal/runtime/js/` 実行時にコピーされる)

- **`index.js`**: `__runtime`として公開されるAPI群です。スーパーグローバルの初期化、`include/require`の動的ロード機構（`registry.js`への委譲）、各種PHP組み込み関数のJS実装（`bin2hex`, `call_user_func`など）が含まれます。
- **`regex.js`**: PHP独自の正規表現処理（`preg_*`ファミリー）をJSの`RegExp`でエミュレートするためのラッパーです。非互換な構文（戻り読みなど）によるワーカーのクラッシュを防ぐための例外処理（`try/catch`）が含まれています。

### 3. トランスパイル出力物 (`output/` ディレクトリ)

- **`src/index.js`**: CF Workersの起点（`fetch`ハンドラ）です。リクエストのURLやCookie、Postデータを`__runtime.superglobals`に詰め込み、トランスパイルされた`index.js`（元の`index.php`）を呼び出します。
- **`src/transpiled/`**: `php2js`によって変換されたPukiWikiのソースコード群です。
  - `registry.js`: 動的`include`を実現するため、全ての変換済みファイルのパスと`import()`の対応表を持つレジストリファイルです。
  - `lib/`, `plugin/`, `skin/`: PukiWikiの各ディレクトリ構造を維持したまま、PHPがJSに変換されて配置されています。
- **`wrangler.toml`**: Cloudflare Workersにデプロイするための設定ファイルです（D1やR2のバインディングもここで定義されます）。
