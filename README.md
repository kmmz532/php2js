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
