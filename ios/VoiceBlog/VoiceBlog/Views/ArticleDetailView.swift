import SwiftUI

struct ArticleDetailView: View {
    @Environment(\.dismiss) private var dismiss

    let auth: AuthManager

    @State private var article: Article
    @State private var showingEditSheet = false
    @State private var showingDeleteConfirmation = false
    @State private var isDeleting = false
    @State private var errorMessage: String?
    @State private var showCopiedToast = false

    init(auth: AuthManager, article: Article) {
        self.auth = auth
        _article = State(initialValue: article)
    }

    private var markdownText: String {
        "# \(article.title)\n\n\(article.content)"
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                AppSurface(accent: .indigo) {
                    Text("記事の概要")
                        .font(.headline)

                    HStack(spacing: 8) {
                        AppTag(title: "\(article.content.count) 文字", tint: .indigo)
                        if article.promptRunJobId != nil {
                            AppTag(title: "AI生成", tint: .orange)
                        } else {
                            AppTag(title: "手動作成", tint: .blue)
                        }
                    }

                    LabeledContent("作成", value: article.createdAt.formatted(date: .abbreviated, time: .shortened))
                    LabeledContent("更新", value: article.updatedAt.formatted(date: .abbreviated, time: .shortened))

                    if let promptRunJobId = article.promptRunJobId {
                        LabeledContent("Job ID", value: String(promptRunJobId))
                    }
                }

                AppSurface(accent: .blue) {
                    Text(article.title)
                        .font(.title2.weight(.bold))

                    Text(article.content)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .textSelection(.enabled)
                }

                AppSurface(accent: .teal) {
                    Text("Markdown 共有")
                        .font(.headline)

                    HStack(spacing: 12) {
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

                        ShareLink(item: markdownText) {
                            Label("共有", systemImage: "square.and.arrow.up")
                        }
                        .buttonStyle(AppSecondaryButtonStyle(tint: .indigo))
                    }

                    if showCopiedToast {
                        Text("Markdown をコピーしました")
                            .font(.subheadline)
                            .foregroundStyle(.teal)
                            .transition(.opacity)
                    }

                    NavigationLink {
                        ArticleShareView(auth: auth, article: article)
                    } label: {
                        Label("外部サービスに投稿", systemImage: "paperplane")
                    }
                    .buttonStyle(AppSecondaryButtonStyle(tint: .purple))
                }
            }
            .padding(20)
        }
        .background(
            LinearGradient(
                colors: [Color.indigo.opacity(0.08), Color.clear],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .navigationTitle("記事詳細")
        .toolbar {
            ToolbarItemGroup(placement: .primaryAction) {
                Button("編集") {
                    showingEditSheet = true
                }

                if isDeleting {
                    ProgressView()
                } else {
                    Button("削除", role: .destructive) {
                        showingDeleteConfirmation = true
                    }
                }
            }
        }
        .sheet(isPresented: $showingEditSheet) {
            ArticleFormView(title: "記事編集", initialDraft: ArticleDraft(article: article)) { draft in
                try await updateArticle(draft)
            }
        }
        .confirmationDialog("この記事を削除しますか？", isPresented: $showingDeleteConfirmation, titleVisibility: .visible) {
            Button("削除", role: .destructive) {
                Task {
                    await deleteArticle()
                }
            }
            Button("キャンセル", role: .cancel) {}
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
    private func updateArticle(_ draft: ArticleDraft) async throws {
        let token = try await auth.fetchIDToken()
        article = try await APIClient.shared.updateArticle(
            id: article.id,
            request: ArticleUpdateRequest(title: draft.title, content: draft.content),
            token: token
        )
    }

    @MainActor
    private func deleteArticle() async {
        isDeleting = true
        defer { isDeleting = false }

        do {
            let token = try await auth.fetchIDToken()
            try await APIClient.shared.deleteArticle(id: article.id, token: token)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
