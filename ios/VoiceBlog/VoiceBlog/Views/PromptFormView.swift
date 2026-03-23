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
                Section("基本情報") {
                    TextField("名前", text: $draft.name)
                    Toggle("有効", isOn: $draft.isActive)
                }

                Section("本文") {
                    TextEditor(text: $draft.body)
                        .frame(minHeight: 220)
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
