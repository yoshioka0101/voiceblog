import SwiftUI

struct ArticleFormView: View {
    @Environment(\.dismiss) private var dismiss

    let title: String
    let onSubmit: @MainActor (ArticleDraft) async throws -> Void

    @State private var draft: ArticleDraft
    @State private var isSaving = false
    @State private var errorMessage: String?

    init(
        title: String,
        initialDraft: ArticleDraft = ArticleDraft(),
        onSubmit: @escaping @MainActor (ArticleDraft) async throws -> Void
    ) {
        self.title = title
        self.onSubmit = onSubmit
        _draft = State(initialValue: initialDraft)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("使い方") {
                    Text("AI が生成した記事の微調整にも、手動の新規作成にも同じ画面を使います。")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                Section("基本情報") {
                    TextField("タイトル", text: $draft.title)
                    Text("\(draft.title.count) 文字")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                Section("本文") {
                    TextEditor(text: $draft.content)
                        .frame(minHeight: 260)

                    Text("\(draft.content.count) 文字")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
            .navigationTitle(title)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("閉じる") {
                        dismiss()
                    }
                }

                ToolbarItem(placement: .confirmationAction) {
                    if isSaving {
                        ProgressView()
                    } else {
                        Button("保存") {
                            Task { @MainActor in
                                await submit()
                            }
                        }
                        .disabled(isSaveDisabled)
                    }
                }
            }
            .alert("エラー", isPresented: isShowingError) {
                Button("閉じる", role: .cancel) {
                    errorMessage = nil
                }
            } message: {
                Text(errorMessage ?? "")
            }
        }
    }

    private var isShowingError: Binding<Bool> {
        Binding(
            get: { errorMessage != nil },
            set: { newValue in
                if !newValue {
                    errorMessage = nil
                }
            }
        )
    }

    private var isSaveDisabled: Bool {
        draft.title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            || draft.content.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    @MainActor
    private func submit() async {
        guard !isSaving else {
            return
        }

        isSaving = true

        do {
            try await onSubmit(draft)
            isSaving = false
            dismiss()
        } catch {
            isSaving = false
            errorMessage = error.localizedDescription
        }
    }
}
