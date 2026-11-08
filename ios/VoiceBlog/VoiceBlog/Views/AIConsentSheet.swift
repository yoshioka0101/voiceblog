import SwiftUI

/// AI 記事生成で文字起こしを第三者 AI（Google Gemini）に送ることへの同意。
enum AIConsent {
    /// 送信内容や送信先を変えた時は版を上げて、もう一度同意を取る。
    static let storageKey = "aiDataSharingConsent.v1"
}

/// 初回の AI 記事生成の前に表示する同意シート。
struct AIConsentSheet: View {
    let onAccept: () -> Void
    let onDecline: () -> Void

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    Label("AI に送信する内容の確認", systemImage: "sparkles")
                        .font(.title3.weight(.semibold))

                    Text("記事の下書きを作るために、次の内容を外部の AI サービスに送信します。")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)

                    consentItem(
                        title: "送信する内容",
                        detail: "選んだ文字起こしのテキストと、選んだプロンプトの本文",
                        systemImage: "doc.text"
                    )
                    consentItem(
                        title: "送信先",
                        detail: "Google の Gemini API（VoiceBlog のサーバーを経由して送信します）",
                        systemImage: "server.rack"
                    )
                    consentItem(
                        title: "目的",
                        detail: "ブログ記事の下書きを生成するためだけに使います。録音した音声そのものは送信しません。",
                        systemImage: "text.badge.checkmark"
                    )

                    Text("同意しない場合も、録音・文字起こし・記事の手動作成は使えます。")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }
                .padding(20)
            }
            .navigationTitle("AI 記事生成")
            .safeAreaInset(edge: .bottom) {
                VStack(spacing: 12) {
                    Button(action: onAccept) {
                        Text("同意して生成する")
                            .frame(maxWidth: .infinity)
                    }
                    .buttonStyle(AppPrimaryButtonStyle(tint: .orange))

                    Button("やめる", role: .cancel, action: onDecline)
                }
                .padding(20)
                .background(.bar)
            }
        }
    }

    private func consentItem(title: String, detail: String, systemImage: String) -> some View {
        HStack(alignment: .top, spacing: 12) {
            Image(systemName: systemImage)
                .font(.title3)
                .foregroundStyle(.orange)
                .frame(width: 28)

            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(.headline)
                Text(detail)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
        }
    }
}
