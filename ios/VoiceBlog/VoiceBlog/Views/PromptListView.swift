import SwiftUI

struct PromptListView: View {
    enum EntryPoint {
        case list
        case create
    }

    var auth: AuthManager
    var entryPoint: EntryPoint = .list

    @State private var prompts: [Prompt] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var showingCreateSheet = false
    @State private var editingPrompt: Prompt?
    @State private var didApplyEntryPoint = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                summaryCard

                if isLoading && prompts.isEmpty {
                    AppSurface(accent: .orange) {
                        HStack {
                            ProgressView()
                            Text("プロンプトを読み込んでいます")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    }
                } else if prompts.isEmpty {
                    AppSurface(accent: .orange) {
                        Label("プロンプトがまだありません", systemImage: "text.badge.plus")
                            .font(.headline)

                        Text("共通 prompt はここに表示され、自分用 prompt は下のボタンからすぐ追加できます。")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                } else {
                    if !systemPrompts.isEmpty {
                        promptSection(
                            title: "共通 prompt",
                            subtitle: "読み取り専用。AI 実行前の基準として使います。",
                            accent: .orange,
                            prompts: systemPrompts
                        )
                    }

                    if !userPrompts.isEmpty {
                        promptSection(
                            title: "自分の prompt",
                            subtitle: "編集・削除できる自分専用の prompt です。",
                            accent: .blue,
                            prompts: userPrompts
                        ) { prompt in
                            HStack(spacing: 12) {
                                Button {
                                    editingPrompt = prompt
                                } label: {
                                    Label("編集", systemImage: "square.and.pencil")
                                }
                                .buttonStyle(AppSecondaryButtonStyle(tint: .blue))

                                Button(role: .destructive) {
                                    Task {
                                        await deletePrompt(prompt)
                                    }
                                } label: {
                                    Label("削除", systemImage: "trash")
                                }
                                .buttonStyle(AppSecondaryButtonStyle(tint: .red))
                            }
                        }
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
            applyEntryPointIfNeeded()
        }
        .refreshable {
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
        .safeAreaInset(edge: .bottom) {
            Button {
                showingCreateSheet = true
            } label: {
                Label("自分の prompt を追加", systemImage: "plus")
            }
            .buttonStyle(AppPrimaryButtonStyle(tint: .orange))
            .padding(.horizontal, 20)
            .padding(.top, 8)
            .background(.thinMaterial)
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

    private var systemPrompts: [Prompt] {
        prompts.filter(\.isSystemPrompt)
    }

    private var userPrompts: [Prompt] {
        prompts.filter { !$0.isSystemPrompt }
    }

    private var summaryCard: some View {
        AppSurface(accent: .orange) {
            Text("prompt を見直して書き方を揃える")
                .font(.title3.weight(.bold))

            Text("共通 prompt は内容確認、自分の prompt は使い回し用のテンプレート管理に寄せています。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            HStack(spacing: 12) {
                AppTag(title: "共通 \(systemPrompts.count)", tint: .orange)
                AppTag(title: "自分用 \(userPrompts.count)", tint: .blue)
            }
        }
    }

    @ViewBuilder
    private func promptSection(
        title: String,
        subtitle: String,
        accent: Color,
        prompts: [Prompt],
        @ViewBuilder actions: @escaping (Prompt) -> some View = { _ in EmptyView() }
    ) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(title)
                .font(.headline)

            Text(subtitle)
                .font(.subheadline)
                .foregroundStyle(.secondary)

            ForEach(prompts) { prompt in
                VStack(spacing: 10) {
                    NavigationLink {
                        PromptDetailView(prompt: prompt)
                    } label: {
                        PromptRowView(prompt: prompt, accent: accent)
                    }
                    .buttonStyle(.plain)

                    actions(prompt)
                }
            }
        }
    }

    private func applyEntryPointIfNeeded() {
        guard !didApplyEntryPoint else {
            return
        }

        didApplyEntryPoint = true
        if entryPoint == .create {
            showingCreateSheet = true
        }
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

private struct PromptRowView: View {
    let prompt: Prompt
    let accent: Color

    var body: some View {
        AppSurface(accent: accent) {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 10) {
                    Text(prompt.name)
                        .font(.headline)

                    Text(prompt.body)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(3)
                }

                Spacer(minLength: 0)

                Image(systemName: "chevron.right")
                    .font(.caption.weight(.bold))
                    .foregroundStyle(.tertiary)
                    .padding(.top, 2)
            }

            HStack(spacing: 8) {
                if prompt.isSystemPrompt {
                    AppTag(title: "共通", tint: .orange)
                }
                if prompt.isDefault {
                    AppTag(title: "標準", tint: .blue)
                }
                if !prompt.isActive {
                    AppTag(title: "非表示", tint: .red)
                }
            }

            Text(prompt.updatedAt.formatted(date: .abbreviated, time: .shortened))
                .font(.caption)
                .foregroundStyle(.secondary)
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
