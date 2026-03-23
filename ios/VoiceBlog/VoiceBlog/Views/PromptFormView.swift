import SwiftUI

struct PromptFormView: View {
    @Environment(\.dismiss) private var dismiss

    let title: String
    let onSubmit: @Sendable (PromptDraft) async throws -> Void

    @State private var draft: PromptDraft
    @State private var isSaving = false
    @State private var errorMessage: String?

    init(
        title: String,
        initialDraft: PromptDraft = PromptDraft(),
        onSubmit: @escaping @Sendable (PromptDraft) async throws -> Void
    ) {
        self.title = title
        self.onSubmit = onSubmit
        _draft = State(initialValue: initialDraft)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("使い方") {
                    Text("共通 prompt は編集できません。この画面では自分用 prompt を追加・更新します。")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                Section("基本情報") {
                    TextField("名前", text: $draft.name)
                    Toggle("有効", isOn: $draft.isActive)
                }

                Section("本文") {
                    TextEditor(text: $draft.body)
                        .frame(minHeight: 220)

                    Text("\(draft.body.count) 文字")
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
                            Task {
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
        draft.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            || draft.body.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private func submit() async {
        isSaving = true
        defer { isSaving = false }

        do {
            try await onSubmit(draft)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
