# App Review への提出メモ

## App Review Information に貼る文面（英語）

```text
VoiceBlog does not require sign-in. All features are available right after launch.
Each install is identified by a random value generated on the device and stored in the Keychain;
we do not collect name, email or phone number.

Main flow:
1. Tap "録音を始める" (Start recording) on the home screen and allow microphone and speech recognition.
2. Speak, then tap stop. The audio is transcribed on the device. Tap save.
3. Choose a prompt and tap "記事を生成" (Generate article).
   The first time, a consent sheet explains that the transcript and the prompt are sent to
   Google's Gemini API to draft the article. Tap "同意して生成する" (Agree and generate).
4. Review the draft and save it. Saved articles are listed under "記事一覧" (Articles).

Publishing to Hatena Blog or Qiita (optional) requires the user's own API token for those services,
entered under Settings (gear icon) > "外部連携設定". This is not needed to review the rest of the app.

Data deletion: Settings (gear icon) > "データを削除" (Delete data) deletes all of the user's data
from our server (transcripts, articles, prompts, tokens and publish history).

Privacy policy: <privacy policy URL>
```

## 用意するもの

- **プライバシーポリシー URL**: リポジトリが public なので、`docs/app-store/privacy-policy.md` の GitHub 上の URL をそのまま使える。お問い合わせ先を記入してから使う
- **サポート URL**: GitHub の Issues（`https://github.com/yoshioka0101/voiceblog/issues`）か、連絡先を書いたページ
- **年齢制限**: 外部サービスへの投稿はあるが、アプリ内で他人の投稿を表示しないため、ユーザー生成コンテンツの項目は「なし」で回答できる見込み
- **輸出コンプライアンス**: HTTPS など OS 標準の暗号化のみ。Info.plist に `ITSAppUsesNonExemptEncryption = NO` を設定済みなので、アップロードごとの質問は出ない

## 審査で見られやすい点と対応

| ガイドライン | 内容 | 対応 |
|---|---|---|
| 4.8 ログインサービス | 外部ログインを使うなら Sign in with Apple 等が必要 | ログイン自体を撤廃（#50） |
| 5.1.1(v) アカウント削除 | アカウントを作れるならアプリ内で削除できること | 「データを削除」で全削除（#48, #50） |
| 5.1.2 データの共有 | 第三者 AI に個人データを送る時は明示と同意 | 初回生成前に同意シート（#51） |
| 2.1 完全性 | 審査時に API が動いていること | 審査期間中は本番 API を止めない。生成上限に注意（#49） |

## 審査期間中の注意

- AI 生成は 1 ユーザー 1 日 20 回、全体で 1 日 300 回まで（`GENERATE_DAILY_LIMIT_*`）。審査員が試せるよう、審査中に上限へ近づいていないかログで確認する
- API のデプロイ先（#45）が審査期間中に止まらないこと
