import SwiftUI

struct PromptRunComposerView: View {
    enum DraftTab: String, CaseIterable, Identifiable {
        case source
        case preview

        var id: String { rawValue }

        var title: String {
            switch self {
            case .source:
                return "Markdown"
            case .preview:
                return "Preview"
            }
        }
    }

    let auth: AuthManager
    let transcription: Transcription

    @State private var prompts: [Prompt] = []
    @State private var selectedPromptId: Int64?
    @State private var isLoading = false
    @State private var isRunning = false
    @State private var job: PromptRunJob?
    @State private var savedArticle: Article?
    @State private var selectedDraftTab: DraftTab = .source
    @State private var isSavingArticle = false
    @State private var errorMessage: String?
    @State private var pollTask: Task<Void, Never>?

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

                if let job {
                    jobStatusSection(job)
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
        .navigationTitle("AI 下書き生成")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                if isRunning {
                    ProgressView()
                } else {
                    Button {
                        Task {
                            await runPrompt()
                        }
                    } label: {
                        Label("実行", systemImage: "sparkles")
                    }
                    .disabled(selectedPrompt == nil)
                }
            }
        }
        .task {
            await loadPrompts()
        }
        .onDisappear {
            pollTask?.cancel()
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

    @ViewBuilder
    private func jobStatusSection(_ job: PromptRunJob) -> some View {
        AppSurface(accent: statusTint(for: job.status)) {
            Text("生成結果")
                .font(.headline)

            HStack(spacing: 8) {
                AppTag(title: statusLabel(for: job.status), tint: statusTint(for: job.status))
                AppTag(title: "試行 \(job.attemptCount) 回", tint: .blue)
            }

            LabeledContent("Job ID", value: String(job.id))
            LabeledContent("作成", value: job.createdAt.formatted(date: .abbreviated, time: .shortened))

            if let errorMessage = job.errorMessage, !errorMessage.isEmpty {
                Text(errorMessage)
                    .font(.subheadline)
                    .foregroundStyle(.red)
            }

            if let generatedTitle = job.generatedTitle,
               let generatedContent = job.generatedContent {
                VStack(alignment: .leading, spacing: 10) {
                    Text("AI が生成した markdown")
                        .font(.headline)

                    Picker("表示", selection: $selectedDraftTab) {
                        ForEach(DraftTab.allCases) { tab in
                            Text(tab.title).tag(tab)
                        }
                    }
                    .pickerStyle(.segmented)

                    if selectedDraftTab == .source {
                        Text(markdownSource(title: generatedTitle, content: generatedContent))
                            .font(.system(.body, design: .monospaced))
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .textSelection(.enabled)
                    } else {
                        markdownPreview(title: generatedTitle, content: generatedContent)
                    }

                    if let savedArticle {
                        NavigationLink {
                            ArticleDetailView(auth: auth, article: savedArticle)
                        } label: {
                            Label("保存した記事を見る", systemImage: "doc.text.magnifyingglass")
                        }
                        .buttonStyle(AppSecondaryButtonStyle(tint: .indigo))
                    } else if isSavingArticle {
                        ProgressView("記事を保存しています")
                            .padding(.top, 4)
                    } else {
                        Button {
                            Task {
                                await saveArticle(job: job, title: generatedTitle, content: generatedContent)
                            }
                        } label: {
                            Label("記事として保存", systemImage: "square.and.arrow.down")
                        }
                        .buttonStyle(AppPrimaryButtonStyle(tint: .indigo))
                    }
                }
                .padding(.top, 6)
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
            errorMessage = error.localizedDescription
        }
    }

    @MainActor
    private func runPrompt() async {
        guard let selectedPrompt else {
            return
        }

        isRunning = true
        savedArticle = nil
        defer { isRunning = false }

        do {
            let token = try await auth.fetchIDToken()
            let created = try await APIClient.shared.createPromptRunJob(
                request: PromptRunJobCreateRequest(transcriptionId: transcription.id, promptId: selectedPrompt.id),
                token: token
            )
            job = created

            if created.isInProgress {
                await startPolling(jobId: created.id)
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    @MainActor
    private func startPolling(jobId: Int64) async {
        pollTask?.cancel()
        pollTask = Task {
            for _ in 0..<10 {
                guard !Task.isCancelled else {
                    return
                }

                try? await Task.sleep(for: .seconds(1))
                do {
                    let token = try await auth.fetchIDToken()
                    let latest = try await APIClient.shared.getPromptRunJob(id: jobId, token: token)
                    await MainActor.run {
                        job = latest
                    }
                    if !latest.isInProgress {
                        return
                    }
                } catch {
                    await MainActor.run {
                        errorMessage = error.localizedDescription
                    }
                    return
                }
            }
        }
    }

    @MainActor
    private func saveArticle(job: PromptRunJob, title: String, content: String) async {
        guard !isSavingArticle else {
            return
        }

        isSavingArticle = true
        defer { isSavingArticle = false }

        do {
            let token = try await auth.fetchIDToken()
            savedArticle = try await APIClient.shared.createArticle(
                request: ArticleCreateRequest(title: title, content: content, promptRunJobId: job.id),
                token: token
            )
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func markdownSource(title: String, content: String) -> String {
        "# \(title)\n\n\(content)"
    }

    @ViewBuilder
    private func markdownPreview(title: String, content: String) -> some View {
        let markdown = markdownSource(title: title, content: content)
        if let attributed = try? AttributedString(
            markdown: markdown,
            options: AttributedString.MarkdownParsingOptions(interpretedSyntax: .full)
        ) {
            Text(attributed)
                .frame(maxWidth: .infinity, alignment: .leading)
                .textSelection(.enabled)
        } else {
            Text(markdown)
                .frame(maxWidth: .infinity, alignment: .leading)
                .textSelection(.enabled)
        }
    }

    private func statusTint(for status: String) -> Color {
        switch status {
        case "completed":
            return .green
        case "failed":
            return .red
        case "running":
            return .orange
        default:
            return .blue
        }
    }

    private func statusLabel(for status: String) -> String {
        switch status {
        case "completed":
            return "完了"
        case "failed":
            return "失敗"
        case "running":
            return "実行中"
        default:
            return "待機中"
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
