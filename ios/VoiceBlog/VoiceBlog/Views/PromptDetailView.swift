import SwiftUI

struct PromptDetailView: View {
    let prompt: Prompt

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                AppSurface(accent: .orange) {
                    Text(prompt.name)
                        .font(.title2.weight(.bold))

                    HStack(spacing: 8) {
                        AppTag(title: prompt.isSystemPrompt ? "共通" : "ユーザー", tint: prompt.isSystemPrompt ? .orange : .blue)
                        if prompt.isDefault {
                            AppTag(title: "標準", tint: .blue)
                        }
                        AppTag(title: prompt.isActive ? "有効" : "非表示", tint: prompt.isActive ? .teal : .red)
                    }

                    Text(prompt.isSystemPrompt ? "この prompt は共通設定です。内容を確認して使い分けます。" : "この prompt は自分用です。一覧画面から編集・削除できます。")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                AppSurface(accent: .blue) {
                    Text("本文")
                        .font(.headline)

                    Text(prompt.body)
                        .font(.body)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .textSelection(.enabled)
                }

                AppSurface(accent: .teal) {
                    Text("メタデータ")
                        .font(.headline)

                    LabeledContent("作成", value: prompt.createdAt.formatted(date: .abbreviated, time: .shortened))
                    LabeledContent("更新", value: prompt.updatedAt.formatted(date: .abbreviated, time: .shortened))
                    if let userId = prompt.userId {
                        LabeledContent("ユーザーID", value: String(userId))
                    }
                }
            }
            .padding(20)
        }
        .background(
            LinearGradient(
                colors: [Color.orange.opacity(0.08), Color.clear],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .navigationTitle("プロンプト詳細")
        .homeNavigationToolbar()
    }
}
