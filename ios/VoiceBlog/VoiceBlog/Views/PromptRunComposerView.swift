import SwiftUI

struct PromptRunComposerView: View {
    let auth: AuthManager
    let transcription: Transcription

    @State private var prompts: [Prompt] = []
    @State private var selectedPromptId: Int64?
    @State private var isLoading = false
    @State private var isGenerating = false
    @State private var generatedArticle: Article?
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

                if let generatedArticle {
                    generatedArticleSection(generatedArticle)
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
                    .disabled(selectedPrompt == nil)
                }
            }
        }
        .task {
            await loadPrompts()
        }
        .navigationDestination(isPresented: $showGeneratedArticle) {
            if let generatedArticle {
                ArticleDetailView(auth: auth, article: generatedArticle)
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
                AppTag(title: "\(transcription.segmentsJson.count) 区間", tint: .teal)
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

    private func generatedArticleSection(_ article: Article) -> some View {
        AppSurface(accent: .green) {
            Text("生成した記事")
                .font(.headline)

            HStack(spacing: 8) {
                AppTag(title: "保存済み", tint: .green)
                AppTag(title: article.updatedAt.formatted(date: .abbreviated, time: .shortened), tint: .blue)
            }

            Text(article.title)
                .font(.title3.weight(.semibold))

            Text(article.content)
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .lineLimit(8)

            Button {
                showGeneratedArticle = true
            } label: {
                Label("記事詳細を開く", systemImage: "doc.text.magnifyingglass")
            }
            .buttonStyle(AppSecondaryButtonStyle(tint: .indigo))
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
        guard let selectedPrompt else {
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
