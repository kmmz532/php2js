# PukiWiki for CfのためのPHP on Node.js (Workers)
PukiWikiを動かせるようなPHPからNode.js (Workers)へのトランスパイラを作成する
実装はGo言語で行う

## 要件
- JSに変換されたPukiWikiそのものがCloudflareで動くことを目標とする
- PHPからJS (Workers/Node.js)への変換を行う (フォルダごと再帰的に変換する)
- 変換後のJSはCf Workers上で動作する
- R2にファイルを保存する
- DBはCf D1を使用する
- トランスパイラ後の出力物(ここではPukiWikiであるが、他のものも使えるとよい)はCf Workersで稼働する

## 環境
- Debian 13 での開発

## データ移行と永続化について（PukiWikiのtxtファイル等）

現在のトランスパイラ（`php2js`）では、入力ディレクトリ（例: `pukiwiki-1.5.4_utf8`）に存在するすべての `.txt` ファイルや `.css`, `.js` 等の静的アセットは、トランスパイル時に自動的に Base64 エンコードされて `output/src/data-manifest.json` にバンドルされる仕組みになっています。
したがって、**元のPukiWikiのページデータ（`wiki/*.txt`）などは、そのままCloudflare Workersのメモリ上に移行・展開されています。**

### 手動で必要な手順（本番環境での永続化）

現在の `wrangler dev`（ローカル環境）では、ページの編集や作成を行っても変更は `data-manifest.json`（メモリ上）にのみ書き込まれ、**サーバーを再起動するとリセットされます**。
Cloudflare Workers 上で永続化を行うためには、**Cloudflare R2（オブジェクトストレージ）** をバインドする必要があります。

1. **R2バケットの作成**:
   Cloudflare のダッシュボードまたは Wrangler コマンドから R2 バケットを作成します。
   ```bash
   wrangler r2 bucket create pukiwiki-data
   ```

2. **wrangler.toml の設定**:
   `output/wrangler.toml` に以下の設定を追記し、R2 を Worker にバインドしてください。
   ```toml
   [[r2_buckets]]
   binding = "R2_BUCKET"
   bucket_name = "pukiwiki-data"
   ```

バインドが完了すると、トランスパイルされた PukiWiki は自動的に `R2_BUCKET` を検知し、ファイルの読み書きを R2 に対して行うようになります（実装済）。R2にファイルが存在しない場合は、初期データとしてバンドルされた `data-manifest.json` の内容をフォールバックとして読み込みます。
