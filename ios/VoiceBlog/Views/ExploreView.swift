import SwiftUI

struct ExploreView: View {
    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 12) {
                    ThemedText("探索する", style: .title)
                        .padding(.horizontal, 20)
                        .padding(.top, 20)

                    ThemedText("このアプリで使用されている機能を確認してみましょう。", style: .body)
                        .padding(.horizontal, 20)
                        .foregroundColor(.appSecondaryLabel)

                    VStack(spacing: 8) {
                        CollapsibleSection(title: "ファイルベースルーティング") {
                            ThemedText(
                                "このアプリは Views ディレクトリのファイル構造を使って画面を管理しています。新しい View を追加するだけで、アプリの一部として統合できます。",
                                style: .body
                            )
                        }

                        CollapsibleSection(title: "クロスプラットフォーム対応") {
                            VStack(alignment: .leading, spacing: 8) {
                                featureRow(icon: "iphone", label: "iOS 17+")
                                featureRow(icon: "macbook", label: "macOS (Mac Catalyst)")
                                featureRow(icon: "visionpro", label: "visionOS")
                            }
                        }

                        CollapsibleSection(title: "画像とメディア") {
                            ThemedText(
                                "SF Symbols を使ったアイコンや、Assets.xcassets から読み込んだ画像を表示できます。ダーク/ライトモードに自動で対応します。",
                                style: .body
                            )
                            HStack(spacing: 20) {
                                ForEach(["photo", "video", "music.note", "mic"], id: \.self) { icon in
                                    Image(systemName: icon)
                                        .font(.title2)
                                        .foregroundColor(.appTint)
                                }
                            }
                            .padding(.top, 4)
                        }

                        CollapsibleSection(title: "テーマとカラー") {
                            VStack(alignment: .leading, spacing: 8) {
                                ThemedText("アクセントカラー: #0a7ea4", style: .body)
                                HStack(spacing: 0) {
                                    ForEach([0.2, 0.4, 0.6, 0.8, 1.0], id: \.self) { opacity in
                                        Rectangle()
                                            .fill(Color.appTint.opacity(opacity))
                                            .frame(height: 32)
                                    }
                                }
                                .clipShape(RoundedRectangle(cornerRadius: 8))
                                ThemedText("ライト/ダークモードに自動対応します。", style: .caption)
                            }
                        }

                        CollapsibleSection(title: "アニメーション") {
                            ThemedText(
                                "withAnimation や .animation モディファイアを使ったスムーズなトランジションを実装しています。折りたたみセクションのシェブロン回転もその一例です。",
                                style: .body
                            )
                        }

                        CollapsibleSection(title: "ハプティックフィードバック") {
                            ThemedText(
                                "UIImpactFeedbackGenerator を使い、タブタップ時に触覚フィードバックを提供します。",
                                style: .body
                            )
                        }
                    }
                    .padding(.horizontal, 20)
                    .padding(.bottom, 40)
                }
            }
            .navigationBarHidden(true)
        }
    }

    private func featureRow(icon: String, label: String) -> some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .foregroundColor(.appTint)
                .frame(width: 24)
            ThemedText(label, style: .body)
        }
    }
}

#Preview {
    ExploreView()
}
