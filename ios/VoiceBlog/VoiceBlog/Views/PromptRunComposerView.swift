import SwiftUI

struct PromptRunComposerView: View {
    let auth: AuthManager
    let transcription: Transcription

    @State private var prompts: [Prompt] = []
    @State private var selectedPromptId: Int64?
    @State private var isLoading = false
    @State private var isGenerating = false
    @State private var generatedArticle: GeneratedArticle?
    @State private var showGeneratedArticle = false
    @State private var errorMessage: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                transcriptionSummaryCard

                if isLoading && prompts.isEmpty {
                    AppSurface(accent: .orange) {
                        HStack {
                            ProgressView()
                            Text("利用できるプロンプトを読み込んでいます")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    }
                } else {
                    promptSelectionSection
                }

            }
            .padding(20)
        }
        .background(
            LinearGradient(
                colors: [Color.orange.opacity(0.08), Color.teal.opacity(0.05), Color.clear],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .navigationTitle("AI 記事生成")
        .homeNavigationToolbar()
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                if isGenerating {
                    ProgressView()
                } else {
                    Button("記事を生成") {
                        Task {
                            await generateArticle()
                        }
                    }
                    .disabled(selectedPrompt == nil || isGenerating)
                }
            }
        }
        .task {
            await loadPrompts()
        }
        .navigationDestination(isPresented: $showGeneratedArticle) {
            if let generatedArticle {
                GeneratedArticlePreviewView(auth: auth, generatedArticle: generatedArticle)
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

    private var selectedPrompt: Prompt? {
        prompts.first(where: { $0.id == selectedPromptId })
    }

    private var transcriptionSummaryCard: some View {
        AppSurface(accent: .teal) {
            Text("対象の文字起こし")
                .font(.headline)

            HStack(spacing: 8) {
                AppTag(title: "\(transcription.fullText.count) 文字", tint: .blue)
            }

            Text(transcription.fullText)
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .lineLimit(6)
        }
    }

    private var promptSelectionSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("使うプロンプトを選ぶ")
                .font(.headline)

            Text("共通プロンプトか自分のプロンプトを選んで、そのまま記事化します。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            ForEach(prompts) { prompt in
                Button {
                    selectedPromptId = prompt.id
                } label: {
                    PromptSelectionRow(prompt: prompt, isSelected: selectedPromptId == prompt.id)
                }
                .buttonStyle(.plain)
            }
        }
    }

    @MainActor
    private func loadPrompts() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await auth.fetchIDToken()
            prompts = try await APIClient.shared.getListPrompts(token: token)
            if selectedPromptId == nil {
                selectedPromptId = prompts.first?.id
            }
        } catch {
            errorMessage = error.userFacingMessage(fallback: "プロンプトの取得に失敗しました。しばらくしてからもう一度お試しください。")
        }
    }

    @MainActor
    private func generateArticle() async {
        guard !isGenerating, let selectedPrompt else {
            return
        }

        isGenerating = true
        generatedArticle = nil
        showGeneratedArticle = false
        defer { isGenerating = false }

        do {
            let token = try await auth.fetchIDToken()
            let article = try await APIClient.shared.generateArticle(
                request: ArticleGenerateRequest(transcriptionId: transcription.id, promptId: selectedPrompt.id),
                token: token
            )
            generatedArticle = article
            showGeneratedArticle = true
        } catch {
            errorMessage = error.userFacingMessage(fallback: "AI記事の生成に失敗しました。しばらくしてからもう一度お試しください。")
        }
    }
}

private struct GeneratedArticlePreviewView: View {
    let auth: AuthManager

    @State private var draft: ArticleDraft
    @State private var createdArticle: Article?
    @State private var showSavedArticle = false
    @State private var isSaving = false
    @State private var showCopiedToast = false
    @State private var errorMessage: String?

    init(auth: AuthManager, generatedArticle: GeneratedArticle) {
        self.auth = auth
        _draft = State(initialValue: ArticleDraft(generatedArticle: generatedArticle))
    }

    private var markdownText: String {
        "# \(draft.title)\n\n\(draft.content)"
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                AppSurface(accent: .green) {
                    Text("生成結果の記事")
                        .font(.headline)

                    Text("内容を確認してから記事として保存できます。")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)

                    AppTag(title: "\(draft.content.count) 文字", tint: .green)
                }

                AppSurface(accent: .blue) {
                    Text(draft.title)
                        .font(.title2.weight(.bold))

                    Text(draft.content)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .textSelection(.enabled)
                }

                AppSurface(accent: .teal) {
                    if isSaving {
                        ProgressView()
                    } else {
                        Button {
                            Task {
                                await saveArticle()
                            }
                        } label: {
                            Label("記事に保存", systemImage: "square.and.arrow.down")
                        }
                        .buttonStyle(AppPrimaryButtonStyle(tint: .green))
                    }

                    Button {
                        copyTextToPasteboard(markdownText)
                        showCopiedToast = true
                        Task {
                            try? await Task.sleep(for: .seconds(2))
                            showCopiedToast = false
                        }
                    } label: {
                        Label("コピー", systemImage: "doc.on.doc")
                    }
                    .buttonStyle(AppSecondaryButtonStyle(tint: .teal))

                    if showCopiedToast {
                        Text("Markdown をコピーしました")
                            .font(.subheadline)
                            .foregroundStyle(.teal)
                            .transition(.opacity)
                    }
                }
            }
            .padding(20)
        }
        .background(
            LinearGradient(
                colors: [Color.green.opacity(0.08), Color.clear],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .navigationTitle("生成結果")
        .homeNavigationToolbar()
        .navigationDestination(isPresented: $showSavedArticle) {
            if let createdArticle {
                ArticleDetailView(auth: auth, article: createdArticle)
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

    @MainActor
    private func saveArticle() async {
        isSaving = true
        defer { isSaving = false }

        do {
            let token = try await auth.fetchIDToken()
            let article = try await APIClient.shared.createArticle(
                request: ArticleCreateRequest(title: draft.title, content: draft.content),
                token: token
            )
            createdArticle = article
            showSavedArticle = true
        } catch {
            errorMessage = error.userFacingMessage(fallback: "記事の保存に失敗しました。しばらくしてからもう一度お試しください。")
        }
    }
}

private struct PromptSelectionRow: View {
    let prompt: Prompt
    let isSelected: Bool

    var body: some View {
        AppSurface(accent: isSelected ? .orange : .gray) {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 10) {
                    HStack(spacing: 8) {
                        Text(prompt.name)
                            .font(.headline)

                        if prompt.isSystemPrompt {
                            AppTag(title: "共通", tint: .orange)
                        } else {
                            AppTag(title: "自分用", tint: .blue)
                        }
                    }

                    Text(prompt.body)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(4)
                }

                Spacer(minLength: 0)

                Image(systemName: isSelected ? "checkmark.circle.fill" : "circle")
                    .font(.title3)
                    .foregroundStyle(isSelected ? Color.orange : Color.secondary.opacity(0.55))
            }
        }
    }
}
