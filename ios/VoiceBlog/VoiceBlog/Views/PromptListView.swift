import SwiftUI

struct PromptListView: View {
    var auth: AuthManager

    @State private var prompts: [Prompt] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var showingCreateSheet = false
    @State private var editingPrompt: Prompt?

    var body: some View {
        List {
            if isLoading && prompts.isEmpty {
                Section {
                    HStack {
                        Spacer()
                        ProgressView()
                        Spacer()
                    }
                }
            }

            if prompts.isEmpty && !isLoading {
                ContentUnavailableView(
                    "プロンプトがありません",
                    systemImage: "text.badge.plus",
                    description: Text("共通プロンプトと自分のプロンプトをここで管理します。")
                )
            } else {
                ForEach(prompts) { prompt in
                    VStack(alignment: .leading, spacing: 10) {
                        HStack(alignment: .top) {
                            Text(prompt.name)
                                .font(.headline)
                            Spacer()
                            if prompt.isSystemPrompt {
                                Text("共通")
                                    .font(.caption.bold())
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 4)
                                    .background(Color.gray.opacity(0.15))
                                    .clipShape(Capsule())
                            }
                            if prompt.isDefault {
                                Text("標準")
                                    .font(.caption.bold())
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 4)
                                    .background(Color.blue.opacity(0.15))
                                    .clipShape(Capsule())
                            }
                            if !prompt.isActive {
                                Text("非表示")
                                    .font(.caption.bold())
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 4)
                                    .background(Color.orange.opacity(0.15))
                                    .clipShape(Capsule())
                            }
                        }

                        Text(prompt.body)
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                            .lineLimit(3)

                        Text(prompt.updatedAt.formatted(date: .abbreviated, time: .shortened))
                            .font(.caption)
                            .foregroundStyle(.tertiary)
                    }
                    .padding(.vertical, 4)
                    .swipeActions(edge: .trailing, allowsFullSwipe: false) {
                        if !prompt.isSystemPrompt {
                            Button("編集") {
                                editingPrompt = prompt
                            }
                            .tint(.blue)

                            Button("削除", role: .destructive) {
                                Task {
                                    await deletePrompt(prompt)
                                }
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle("プロンプト")
        .toolbar {
            ToolbarItemGroup(placement: .primaryAction) {
                Button {
                    Task {
                        await loadPrompts()
                    }
                } label: {
                    Image(systemName: "arrow.clockwise")
                }

                Button {
                    showingCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .task {
            await loadPrompts()
        }
        .sheet(isPresented: $showingCreateSheet) {
            PromptFormView(title: "プロンプト作成") { draft in
                try await createPrompt(draft)
            }
        }
        .sheet(item: $editingPrompt) { prompt in
            PromptFormView(title: "プロンプト編集", initialDraft: PromptDraft(prompt: prompt)) { draft in
                try await updatePrompt(prompt, draft: draft)
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

    private func loadPrompts() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await auth.fetchIDToken()
            prompts = try await APIClient.shared.fetchPrompts(token: token)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func createPrompt(_ draft: PromptDraft) async throws {
        let token = try await auth.fetchIDToken()
        let created = try await APIClient.shared.createPrompt(
            request: PromptCreateRequest(name: draft.name, body: draft.body, isActive: draft.isActive),
            token: token
        )
        prompts.insert(created, at: 0)
    }

    private func updatePrompt(_ prompt: Prompt, draft: PromptDraft) async throws {
        let token = try await auth.fetchIDToken()
        let updated = try await APIClient.shared.updatePrompt(
            id: prompt.id,
            request: PromptUpdateRequest(name: draft.name, body: draft.body, isActive: draft.isActive),
            token: token
        )

        if let index = prompts.firstIndex(where: { $0.id == updated.id }) {
            prompts[index] = updated
        }
    }

    private func deletePrompt(_ prompt: Prompt) async {
        do {
            let token = try await auth.fetchIDToken()
            try await APIClient.shared.deletePrompt(id: prompt.id, token: token)
            prompts.removeAll { $0.id == prompt.id }
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct PromptDraft: Sendable {
    var name: String
    var body: String
    var isActive: Bool

    init(name: String = "", body: String = "", isActive: Bool = true) {
        self.name = name
        self.body = body
        self.isActive = isActive
    }

    init(prompt: Prompt) {
        self.name = prompt.name
        self.body = prompt.body
        self.isActive = prompt.isActive
    }
}
