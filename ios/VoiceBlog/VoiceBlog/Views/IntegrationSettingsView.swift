import SwiftUI

struct IntegrationSettingsView: View {
    let auth: AuthManager

    @State private var integrations: [Integration] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var editingProvider: String?
    @State private var tokenInput = ""
    @State private var isSaving = false
    @State private var featureUnavailable = false
    @State private var showRemoveConfirmation = false
    @State private var removeProvider: String?
    @State private var savedProvider: String?
    @State private var showToken = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                headerCard

                if featureUnavailable {
                    unavailableCard
                } else if isLoading && integrations.isEmpty {
                    loadingCard
                } else {
                    ForEach(integrations) { integration in
                        providerCard(integration)
                    }
                }
            }
            .padding(20)
        }
        .background(
            LinearGradient(
                colors: [Color.purple.opacity(0.08), Color.clear],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .navigationTitle("外部連携")
        .homeNavigationToolbar()
        .task {
            await loadIntegrations()
        }
        .alert("エラー", isPresented: isShowingError) {
            Button("閉じる", role: .cancel) { errorMessage = nil }
        } message: {
            Text(errorMessage ?? "")
        }
        .confirmationDialog(
            "\(providerDisplayName(removeProvider ?? ""))の連携を解除しますか？",
            isPresented: $showRemoveConfirmation,
            titleVisibility: .visible
        ) {
            Button("解除する", role: .destructive) {
                if let provider = removeProvider {
                    Task { await removeToken(provider: provider) }
                }
            }
            Button("キャンセル", role: .cancel) {}
        } message: {
            Text("保存済みの\(credentialDisplayName(removeProvider ?? ""))が削除されます。再度投稿するには\(credentialDisplayName(removeProvider ?? ""))を設定し直す必要があります。")
        }
    }

    private var isShowingError: Binding<Bool> {
        Binding(
            get: { errorMessage != nil },
            set: { if !$0 { errorMessage = nil } }
        )
    }

    // MARK: - Cards

    private var headerCard: some View {
        AppSurface(accent: .purple) {
            Label("外部ブログへの投稿", systemImage: "link.badge.plus")
                .font(.title3.weight(.semibold))

            Text("Qiita のアクセストークンや、はてなブログの認証情報を登録すると、VoiceBlog で作成した記事をそのまま投稿できます。")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
    }

    private var unavailableCard: some View {
        AppSurface(accent: .orange) {
            Label("現在利用できません", systemImage: "exclamationmark.triangle")
                .font(.headline)
                .foregroundStyle(.orange)

            Text("外部連携機能はサーバー管理者が有効にする必要があります。管理者にお問い合わせください。")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
    }

    private var loadingCard: some View {
        AppSurface(accent: .gray) {
            HStack {
                ProgressView()
                Text("読み込み中…")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
        }
    }

    @ViewBuilder
    private func providerCard(_ integration: Integration) -> some View {
        AppSurface(accent: integration.connected ? .green : .gray) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(providerDisplayName(integration.provider))
                        .font(.headline)

                    if integration.connected {
                        Label("接続済み", systemImage: "checkmark.circle.fill")
                            .font(.subheadline)
                            .foregroundStyle(.green)
                    } else {
                        Text("未接続")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                }

                Spacer()
            }

            if editingProvider == integration.provider {
                tokenEditSection(provider: integration.provider)
            } else if integration.connected {
                connectedActions(provider: integration.provider)
            } else {
                disconnectedGuide(provider: integration.provider)
            }

            if savedProvider == integration.provider {
                Label("\(credentialDisplayName(integration.provider))を保存しました", systemImage: "checkmark.circle.fill")
                    .font(.subheadline)
                    .foregroundStyle(.green)
                    .transition(.opacity)
            }
        }
    }

    // MARK: - Provider Sections

    @ViewBuilder
    private func disconnectedGuide(provider: String) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            tokenGuide(for: provider)

            Button {
                editingProvider = provider
                tokenInput = ""
            } label: {
                Label("\(credentialDisplayName(provider))を登録する", systemImage: "key.fill")
            }
            .buttonStyle(AppPrimaryButtonStyle(tint: .purple))
        }
    }

    @ViewBuilder
    private func connectedActions(provider: String) -> some View {
        HStack(spacing: 12) {
            Button {
                editingProvider = provider
                tokenInput = ""
            } label: {
                Label("\(credentialDisplayName(provider))を変更", systemImage: "key")
            }
            .buttonStyle(AppSecondaryButtonStyle(tint: .purple))

            Button {
                removeProvider = provider
                showRemoveConfirmation = true
            } label: {
                Label("解除", systemImage: "xmark.circle")
            }
            .buttonStyle(AppSecondaryButtonStyle(tint: .red))
        }
    }

    @ViewBuilder
    private func tokenEditSection(provider: String) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            tokenGuide(for: provider)

            HStack {
                Group {
                    if showToken {
                        TextField(tokenPlaceholder(for: provider), text: $tokenInput)
                    } else {
                        SecureField(tokenPlaceholder(for: provider), text: $tokenInput)
                    }
                }
                .textFieldStyle(.roundedBorder)
                .textContentType(.password)
                .autocorrectionDisabled()

                Button {
                    showToken.toggle()
                } label: {
                    Image(systemName: showToken ? "eye.slash" : "eye")
                        .foregroundStyle(.secondary)
                }
                .buttonStyle(.plain)
            }

            if isSaving {
                HStack {
                    ProgressView()
                    Text("保存しています…")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
            } else {
                HStack(spacing: 12) {
                    Button {
                        Task { await saveToken(provider: provider) }
                    } label: {
                        Label("保存", systemImage: "checkmark.circle")
                    }
                    .buttonStyle(AppPrimaryButtonStyle(tint: .purple))
                    .disabled(tokenInput.isEmpty)

                    Button {
                        editingProvider = nil
                        tokenInput = ""
                    } label: {
                        Text("キャンセル")
                    }
                    .buttonStyle(AppSecondaryButtonStyle(tint: .gray))
                }
            }
        }
    }

    @ViewBuilder
    private func tokenGuide(for provider: String) -> some View {
        switch provider {
        case "qiita":
            VStack(alignment: .leading, spacing: 6) {
                Text("Qiita のトークン取得手順")
                    .font(.subheadline.weight(.medium))

                VStack(alignment: .leading, spacing: 4) {
                    Text("1. Qiita にログインする")
                    Text("2. 設定 → アプリケーション を開く")
                    Text("3.「個人用アクセストークン」の「新しくトークンを発行する」を押す")
                    Text("4. スコープは「write_qiita」にチェック")
                    Text("5. 発行されたトークンをコピーしてここに貼り付ける")
                }
                .font(.caption)
                .foregroundStyle(.secondary)

                Link(destination: URL(string: "https://qiita.com/settings/tokens/new")!) {
                    Label("Qiita のトークン発行ページを開く", systemImage: "safari")
                        .font(.caption)
                }
            }

        case "hatena":
            VStack(alignment: .leading, spacing: 6) {
                Text("はてなブログのトークン取得手順")
                    .font(.subheadline.weight(.medium))

                VStack(alignment: .leading, spacing: 4) {
                    Text("1. はてなブログの管理画面 → 設定 → 詳細設定 を開く")
                    Text("2.「AtomPub」セクションの API キーを確認する")
                    Text("3. 下記の形式でまとめて入力する:")
                    Text("   はてなID:ブログID:APIキー")
                        .font(.caption.monospaced())
                    Text("   例: tanaka:tanaka-blog:abc123xyz")
                        .font(.caption.monospaced())
                        .foregroundStyle(.orange)
                }
                .font(.caption)
                .foregroundStyle(.secondary)

                Link(destination: URL(string: "https://blog.hatena.ne.jp/my/config/detail")!) {
                    Label("はてなブログの詳細設定を開く", systemImage: "safari")
                        .font(.caption)
                }
            }

        default:
            EmptyView()
        }
    }

    // MARK: - Actions

    @MainActor
    private func loadIntegrations() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await auth.fetchIDToken()
            integrations = try await APIClient.shared.getIntegrations(token: token)
        } catch let error as APIError where error == .notFound {
            featureUnavailable = true
        } catch {
            errorMessage = error.userFacingMessage(fallback: "外部連携の取得に失敗しました。しばらくしてからもう一度お試しください。")
        }
    }

    @MainActor
    private func saveToken(provider: String) async {
        isSaving = true
        defer { isSaving = false }

        do {
            let token = try await auth.fetchIDToken()
            try await APIClient.shared.storeToken(
                provider: provider,
                request: StoreTokenRequest(token: tokenInput),
                token: token
            )
            editingProvider = nil
            tokenInput = ""
            savedProvider = provider
            Task {
                try? await Task.sleep(for: .seconds(3))
                if savedProvider == provider { savedProvider = nil }
            }
            await loadIntegrations()
        } catch {
            errorMessage = userFacingMessage(for: error, provider: provider, action: "保存")
        }
    }

    @MainActor
    private func removeToken(provider: String) async {
        do {
            let token = try await auth.fetchIDToken()
            try await APIClient.shared.deleteToken(provider: provider, token: token)
            await loadIntegrations()
        } catch {
            errorMessage = userFacingMessage(for: error, provider: provider, action: "解除")
        }
    }

    // MARK: - Helpers

    private func providerDisplayName(_ provider: String) -> String {
        switch provider {
        case "qiita": return "Qiita"
        case "hatena": return "はてなブログ"
        default: return provider
        }
    }

    private func credentialDisplayName(_ provider: String) -> String {
        switch provider {
        case "hatena":
            return "認証情報"
        default:
            return "トークン"
        }
    }

    private func tokenPlaceholder(for provider: String) -> String {
        switch provider {
        case "qiita": return "Qiita のアクセストークンを貼り付け"
        case "hatena": return "はてなID:ブログID:APIキー"
        default: return "トークンを入力"
        }
    }

    private func userFacingMessage(for error: Error, provider: String, action: String) -> String {
        let name = providerDisplayName(provider)
        if let apiError = error as? APIError {
            switch apiError {
            case .unauthorized:
                return "認証に失敗しました。再度ログインしてください。"
            case .serverError(statusCode: 400, _, let code) where code == "token_verification_failed":
                return "\(name)への接続に失敗しました。\(credentialDisplayName(provider))が正しいか確認してください。"
            case .serverError(statusCode: 400, _, let code) where code == "token_required":
                return "\(credentialDisplayName(provider))を入力してください。"
            case .serverError(statusCode: 400, _, let code) where code == "unsupported_provider":
                return "\(name)の設定に対応していません。アプリを更新してもう一度お試しください。"
            case .serverError(statusCode: 400, _, _):
                return "\(credentialDisplayName(provider))の形式が正しくありません。\(name)の手順を確認してください。"
            default:
                return "\(name)の\(action)に失敗しました。しばらくしてからもう一度お試しください。"
            }
        }
        return "\(name)の\(action)に失敗しました。ネットワーク接続を確認してください。"
    }
}

extension APIError: Equatable {
    static func == (lhs: APIError, rhs: APIError) -> Bool {
        switch (lhs, rhs) {
        case (.unauthorized, .unauthorized): return true
        case (.notFound, .notFound): return true
        case (.invalidConfiguration(let a), .invalidConfiguration(let b)): return a == b
        case (.serverError(let a1, let a2, let a3), .serverError(let b1, let b2, let b3)):
            return a1 == b1 && a2 == b2 && a3 == b3
        default: return false
        }
    }
}
