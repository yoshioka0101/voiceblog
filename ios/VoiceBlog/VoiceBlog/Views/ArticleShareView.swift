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

    private var markdownText: String {
        "# \(article.title)\n\n\(article.content)"
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                markdownSection
                publishSection
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
        }
    }

    private var isShowingError: Binding<Bool> {
        Binding(
            get: { errorMessage != nil },
            set: { if !$0 { errorMessage = nil } }
        )
    }

    private var markdownSection: some View {
        AppSurface(accent: .teal) {
            Text("Markdown 共有")
                .font(.headline)

            Text("任意のブログやエディタに貼り付けられます。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            HStack(spacing: 12) {
                Button {
                    UIPasteboard.general.string = markdownText
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
        }
    }

    private var publishSection: some View {
        AppSurface(accent: .purple) {
            Text("外部サービスに投稿")
                .font(.headline)

            if isLoading {
                HStack {
                    ProgressView()
                    Text("読み込み中…")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
            } else if integrations.isEmpty {
                Text("外部連携が設定されていません。ホーム画面のメニューから外部連携設定を行ってください。")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            } else {
                ForEach(integrations) { integration in
                    providerRow(integration)
                }
            }
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

                if !integration.connected {
                    AppTag(title: "未接続", tint: .gray)
                } else if existingTarget != nil {
                    AppTag(title: "投稿済み", tint: .green)
                }
            }

            if let target = existingTarget, !target.externalUrl.isEmpty {
                Link(destination: URL(string: target.externalUrl)!) {
                    Label(target.externalUrl, systemImage: "arrow.up.right.square")
                        .font(.caption)
                        .lineLimit(1)
                }
            }

            if integration.connected {
                if publishingProvider == integration.provider {
                    HStack {
                        ProgressView()
                        Text("投稿中…")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                } else {
                    Button {
                        confirmProvider = integration.provider
                        showPublishConfirmation = true
                    } label: {
                        Label(
                            existingTarget != nil ? "再投稿" : "投稿",
                            systemImage: "paperplane"
                        )
                    }
                    .buttonStyle(AppSecondaryButtonStyle(tint: .purple))
                    .disabled(isPublishing)
                }
            }
        }
        .padding(.vertical, 4)
    }

    @MainActor
    private func loadData() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await auth.fetchIDToken()
            async let integrationsResult = APIClient.shared.getIntegrations(token: token)
            async let targetsResult = APIClient.shared.getShareTargets(articleId: article.id, token: token)
            integrations = try await integrationsResult.filter(\.connected)
            shareTargets = try await targetsResult
        } catch {
            errorMessage = error.localizedDescription
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
        } catch {
            errorMessage = error.localizedDescription
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
