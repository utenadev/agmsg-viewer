# agmsg-viewer

[English version here](README.md)

`fujibee/agmsg` によって記録されたエージェント間のメッセージ履歴（SQLiteデータベース）を、LINE風のチャットインターフェースでブラウザから閲覧するためのツールです。

## 🚀 ユーザーガイド

### 特徴
- **LINE風UI**: メッセージの送受信を直感的なバブルUIで表示。
- **エージェント識別**: 送信元エージェントごとに一貫した色が割り当てられ、誰の発言かひと目で分かります。
- **自動更新**: 15秒ごとのポーリングにより、新しいメッセージがあれば自動的に画面に反映されます。
- **チーム永続化**: 最後に選択したチームが保存され、次回アクセス時に自動的にロードされます。
- **ポータビリティ**: Goの単一バイナリで動作し、外部依存ライブラリなしで起動可能です。

### インストールと起動
ビルド済みバイナリを入手するか、ソースからビルドして実行します。

#### 起動コマンド
```bash
./agmsg-viewer -db /path/to/messages.db -port 8080
```

#### オプション
| フラグ | デフォルト値 | 説明 |
|--------|-------------|-------------|
| `-db` | `messages.db` | 読み込むSQLiteデータベースファイルのパス |
| `-port`| `8080` | 起動するHTTPサーバーのポート番号 |
| `-tail`| `40` | 初期表示する最新メッセージ数（0で全件） |
| `-team`| (なし) | 初期表示するチーム名 |

### 使い方
1. サーバーを起動し、ブラウザで `http://localhost:8080` にアクセスします。
2. 画面上部の「Team」ドロップダウンから確認したいチームを選択してください。
3. メッセージの時刻部分にマウスを合わせると、詳細な日付と時刻が表示されます。

---

## 🛠️ 開発・コントリビューターガイド

### 技術スタック
- **Backend**: Go 1.26+ (Standard Library)
- **Database**: SQLite (via `modernc.org/sqlite` - CGO-free)
- **Frontend**: HTMX + Tailwind CSS (Vanilla JS)
- **Embedding**: Go `embed` package (HTML/CSS are bundled into the binary)

### 開発フロー
本プロジェクトは `go-task` を使用してビルドやテストを管理しています。

#### 基本コマンド
- **ビルド**: `task build`
- **実行**: `task run`
- **フォーマット**: `task fmt`
- **静的解析**: `task lint`
- **テスト**: `task test`
- **クリーン**: `task clean`

#### クロスプラットフォームビルド
各OS向けに最適化されたバイナリを生成できます。
- **Windows**: `task build:win`
- **Linux**: `task build:linux`
- **task:mac**: `task build:mac`

### データベーススキーマ
本ツールは `fujibee/agmsg` の以下のスキーマを前提として動作します。

#### テーブル: `messages`
| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER | メッセージID (PK) |
| `team` | TEXT | チーム名 |
| `from_agent` | TEXT | 送信元エージェント名 |
| `to_agent` | TEXT | 送信先エージェント名 |
| `body` | TEXT | メッセージ本文 |
| `created_at` | TEXT | 作成日時 (ISO 8601 UTC) |

---

## 📄 ライセンス

本プロジェクトは [MIT License](LICENSE) の下で公開されています。
