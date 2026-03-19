import SwiftUI

struct ModalView: View {
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    HStack {
                        Image(systemName: "waveform.circle.fill")
                            .font(.system(size: 48))
                            .foregroundColor(.appTint)
                        VStack(alignment: .leading) {
                            ThemedText("VoiceBlog", style: .heading)
                            ThemedText("v1.0.0", style: .caption)
                        }
                    }
                    .padding(.top, 8)

                    Divider()

                    ThemedText("このアプリについて", style: .subheading)
                    ThemedText(
                        "VoiceBlog は SwiftUI で構築されたシンプルなブログアプリです。音声入力を使ってブログ投稿を作成できます。",
                        style: .body
                    )

                    Divider()

                    ThemedText("技術スタック", style: .subheading)
                    VStack(alignment: .leading, spacing: 8) {
                        techRow(icon: "swift", label: "Swift 5.9+")
                        techRow(icon: "iphone", label: "SwiftUI")
                        techRow(icon: "server.rack", label: "Go バックエンド")
                    }
                }
                .padding(24)
            }
            .navigationTitle("詳細情報")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("閉じる") {
                        dismiss()
                    }
                    .foregroundColor(.appTint)
                }
            }
        }
    }

    private func techRow(icon: String, label: String) -> some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .foregroundColor(.appTint)
                .frame(width: 24)
            ThemedText(label, style: .body)
        }
    }
}

#Preview {
    ModalView()
}
