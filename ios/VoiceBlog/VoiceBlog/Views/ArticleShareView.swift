import SwiftUI

struct ArticleShareView: View {
    let auth: AuthManager
    let article: Article

    @State private var integrations: [Integration] = []
    @State private var shareTargets: [ShareTarget] = []
    @State private var isLoading = false
    @State private var isPublishing = false
    @State private var publishingProvider: String?
    @State private var showCopiedToast = false
    @State private var errorMessage: String?
    @State private var showPublishConfirmation = false
    @State private var confirmProvider: String?
    @State private var featureUnavailable = false
    @State private var publishedProvider: String?

    private var markdownText: String {
        "# \(article.title)\n\n\(article.content)"
    }

    private var connectedIntegrations: [Integration] {
        integrations.filter(\.connected)
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                markdownSection
                externalPublishSection
            }
            .padding(20)
        }
        .background(
            LinearGradient(
                colors: [Color.teal.opacity(0.08), Color.purple.opacity(0.06), Color.clear],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .navigationTitle("記事を共有")
        .homeNavigationToolbar()
        .task {
            await loadData()
        }
        .alert("エラー", isPresented: isShowingError) {
            Button("閉じる", role: .cancel) { errorMessage = nil }
        } message: {
            Text(errorMessage ?? "")
        }
        .confirmationDialog(
            "「\(article.title)」を\(providerDisplayName(confirmProvider ?? ""))に投稿しますか？",
            isPresented: $showPublishConfirmation,
            titleVisibility: .visible
        ) {
            Button("投稿する") {
                if let provider = confirmProvider {
                    Task { await publishArticle(provider: provider) }
                }
            }
            Button("キャンセル", role: .cancel) {}
        } message: {
            Text("記事がそのまま公開されます。投稿前に内容を確認してください。")
        }
    }

    private var isShowingError: Binding<Bool> {
        Binding(
            get: { errorMessage != nil },
            set: { if !$0 { errorMessage = nil } }
        )
    }

    // MARK: - Markdown Section

    private var markdownSection: some View {
        AppSurface(accent: .teal) {
            Label("Markdown で共有", systemImage: "doc.text")
                .font(.headline)

            Text("コピーして任意のブログやエディタに貼り付けられます。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

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
                Label("Markdown をコピーしました", systemImage: "checkmark.circle.fill")
                    .font(.subheadline)
                    .foregroundStyle(.teal)
                    .transition(.opacity)
            }
        }
    }

    // MARK: - External Publish Section

    @ViewBuilder
    private var externalPublishSection: some View {
        if featureUnavailable {
            // 管理者が TOKEN_ENCRYPTION_KEY を設定していない場合
            EmptyView()
        } else {
            AppSurface(accent: .purple) {
                Label("外部サービスに投稿", systemImage: "paperplane")
                    .font(.headline)

                if isLoading {
                    HStack {
                        ProgressView()
                        Text("読み込み中…")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                } else if connectedIntegrations.isEmpty {
                    noConnectionGuide
                } else {
                    ForEach(connectedIntegrations) { integration in
                        providerRow(integration)
                    }
                }
            }
        }
    }

    private var noConnectionGuide: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Qiita やはてなブログのトークンを登録すると、ここから直接投稿できます。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            NavigationLink {
                IntegrationSettingsView(auth: auth)
            } label: {
                Label("トークンを登録する", systemImage: "key.fill")
            }
            .buttonStyle(AppPrimaryButtonStyle(tint: .purple))
        }
    }

    @ViewBuilder
    private func providerRow(_ integration: Integration) -> some View {
        let existingTarget = shareTargets.first(where: { $0.provider == integration.provider })

        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(providerDisplayName(integration.provider))
                    .font(.subheadline.weight(.medium))

                Spacer()

                if existingTarget != nil {
                    AppTag(title: "投稿済み", tint: .green)
                }
            }

            if let target = existingTarget, !target.externalUrl.isEmpty,
               let url = URL(string: target.externalUrl) {
                Link(destination: url) {
                    Label("投稿を確認する", systemImage: "arrow.up.right.square")
                        .font(.subheadline)
                }
            }

            if publishedProvider == integration.provider {
                Label("投稿しました", systemImage: "checkmark.circle.fill")
                    .font(.subheadline)
                    .foregroundStyle(.green)
                    .transition(.opacity)
            } else if publishingProvider == integration.provider {
                HStack {
                    ProgressView()
                    Text("投稿しています…")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
            } else {
                Button {
                    confirmProvider = integration.provider
                    showPublishConfirmation = true
                } label: {
                    Label(
                        existingTarget != nil ? "再投稿する" : "投稿する",
                        systemImage: "paperplane"
                    )
                }
                .buttonStyle(AppSecondaryButtonStyle(tint: .purple))
                .disabled(isPublishing)
            }
        }
        .padding(.vertical, 4)
    }

    // MARK: - Actions

    @MainActor
    private func loadData() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await auth.fetchIDToken()
            // share-targets は常に取得可能（404 ではない）
            // integrations は TOKEN_ENCRYPTION_KEY 未設定だと 404
            do {
                integrations = try await APIClient.shared.getIntegrations(token: token)
            } catch let error as APIError where error == .notFound {
                featureUnavailable = true
            }

            if !featureUnavailable {
                shareTargets = try await APIClient.shared.getShareTargets(articleId: article.id, token: token)
            }
        } catch {
            errorMessage = error.userFacingMessage(fallback: "共有設定の取得に失敗しました。しばらくしてからもう一度お試しください。")
        }
    }

    @MainActor
    private func publishArticle(provider: String) async {
        isPublishing = true
        publishingProvider = provider
        defer {
            isPublishing = false
            publishingProvider = nil
        }

        do {
            let token = try await auth.fetchIDToken()
            let target = try await APIClient.shared.publishArticle(
                id: article.id,
                request: PublishRequest(provider: provider),
                token: token
            )
            if let idx = shareTargets.firstIndex(where: { $0.provider == provider }) {
                shareTargets[idx] = target
            } else {
                shareTargets.append(target)
            }
            publishedProvider = provider
            Task {
                try? await Task.sleep(for: .seconds(3))
                if publishedProvider == provider { publishedProvider = nil }
            }
        } catch {
            let name = providerDisplayName(provider)
            if let apiError = error as? APIError {
                switch apiError {
                case .serverError(statusCode: 400, _, let code) where code == "provider_not_connected":
                        errorMessage = "\(name)のトークンが無効になっています。外部連携設定から再設定してください。"
                case .serverError(statusCode: 400, _, let code) where code == "unsupported_provider":
                    errorMessage = "\(name)への投稿先の設定に問題があります。アプリを更新してもう一度お試しください。"
                case .unauthorized:
                    errorMessage = "認証に失敗しました。再度ログインしてください。"
                default:
                    errorMessage = "\(name)への投稿に失敗しました。しばらくしてからもう一度お試しください。"
                }
            } else {
                errorMessage = "\(name)への投稿に失敗しました。ネットワーク接続を確認してください。"
            }
        }
    }

    private func providerDisplayName(_ provider: String) -> String {
        switch provider {
        case "qiita": return "Qiita"
        case "hatena": return "はてなブログ"
        default: return provider
        }
    }
}
