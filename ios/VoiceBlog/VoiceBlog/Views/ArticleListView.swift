import SwiftUI

struct ArticleListView: View {
    var auth: AuthManager

    @State private var articles: [Article] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var showingCreateSheet = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                summaryCard

                if isLoading && articles.isEmpty {
                    AppSurface(accent: .indigo) {
                        HStack {
                            ProgressView()
                            Text("記事を読み込んでいます")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    }
                } else if articles.isEmpty {
                    AppSurface(accent: .indigo) {
                        Label("記事はまだありません", systemImage: "doc.text.image")
                            .font(.headline)

                        Text("AI 実行で記事を生成するか、手動で記事を作成するとここに表示されます。")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                } else {
                    ForEach(articles) { article in
                        NavigationLink {
                            ArticleDetailView(auth: auth, article: article)
                        } label: {
                            ArticleRowView(article: article)
                        }
                        .buttonStyle(.plain)
                    }
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
        .navigationTitle("記事")
        .toolbar {
            ToolbarItemGroup(placement: .primaryAction) {
                Button {
                    Task {
                        await loadArticles()
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
            await loadArticles()
        }
        .refreshable {
            await loadArticles()
        }
        .sheet(isPresented: $showingCreateSheet) {
            ArticleFormView(title: "記事作成") { draft in
                try await createArticle(draft)
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

    private var summaryCard: some View {
        AppSurface(accent: .indigo) {
            Text("生成記事と手動記事をまとめて確認する")
                .font(.title3.weight(.bold))

            Text("AI 実行後の記事も、あとから手で作った記事も同じ一覧で追いかけられます。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            AppTag(title: "\(articles.count) 件", tint: .indigo)
        }
    }

    @MainActor
    private func loadArticles() async {
        guard !isLoading else {
            return
        }

        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await auth.fetchIDToken()
            articles = try await APIClient.shared.getListArticles(token: token)
        } catch {
            errorMessage = error.userFacingMessage(fallback: "記事の取得に失敗しました。しばらくしてからもう一度お試しください。")
        }
    }

    @MainActor
    private func createArticle(_ draft: ArticleDraft) async throws {
        let token = try await auth.fetchIDToken()
        let created = try await APIClient.shared.createArticle(
            request: ArticleCreateRequest(title: draft.title, content: draft.content, promptRunJobId: nil),
            token: token
        )
        articles.insert(created, at: 0)
    }
}

private struct ArticleRowView: View {
    let article: Article

    var body: some View {
        AppSurface(accent: .indigo) {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 10) {
                    Text(article.title)
                        .font(.headline)
                        .lineLimit(2)

                    Text(article.content)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(3)

                    HStack(spacing: 8) {
                        if article.promptRunJobId != nil {
                            AppTag(title: "AI生成", tint: .orange)
                        } else {
                            AppTag(title: "手動作成", tint: .blue)
                        }
                        AppTag(title: article.updatedAt.formatted(date: .abbreviated, time: .omitted), tint: .indigo)
                    }
                }

                Spacer(minLength: 0)

                Image(systemName: "chevron.right")
                    .font(.caption.weight(.bold))
                    .foregroundStyle(.tertiary)
                    .padding(.top, 2)
            }
        }
    }
}
