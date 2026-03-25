import SwiftUI

struct IntegrationSettingsView: View {
    let auth: AuthManager

    @State private var integrations: [Integration] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var editingProvider: String?
    @State private var tokenInput = ""
    @State private var isSaving = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                AppSurface(accent: .purple) {
                    Label("外部連携設定", systemImage: "link.badge.plus")
                        .font(.title3.weight(.semibold))

                    Text("外部ブログサービスの API トークンを設定すると、記事を直接投稿できます。")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                if isLoading && integrations.isEmpty {
                    AppSurface(accent: .gray) {
                        HStack {
                            ProgressView()
                            Text("読み込み中…")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    }
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
        .task {
            await loadIntegrations()
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
            set: { if !$0 { errorMessage = nil } }
        )
    }

    @ViewBuilder
    private func providerCard(_ integration: Integration) -> some View {
        AppSurface(accent: integration.connected ? .green : .gray) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(providerDisplayName(integration.provider))
                        .font(.headline)
                    Text(integration.connected ? "接続済み" : "未接続")
                        .font(.subheadline)
                        .foregroundStyle(integration.connected ? .green : .secondary)
                }

                Spacer()

                AppTag(title: integration.connected ? "接続済み" : "未接続",
                       tint: integration.connected ? .green : .gray)
            }

            if editingProvider == integration.provider {
                VStack(alignment: .leading, spacing: 10) {
                    Text(tokenHint(for: integration.provider))
                        .font(.caption)
                        .foregroundStyle(.secondary)

                    SecureField("API トークン", text: $tokenInput)
                        .textFieldStyle(.roundedBorder)

                    HStack(spacing: 12) {
                        Button {
                            Task { await saveToken(provider: integration.provider) }
                        } label: {
                            Label("保存", systemImage: "checkmark.circle")
                        }
                        .buttonStyle(AppPrimaryButtonStyle(tint: .purple))
                        .disabled(tokenInput.isEmpty || isSaving)

                        Button {
                            editingProvider = nil
                            tokenInput = ""
                        } label: {
                            Text("キャンセル")
                        }
                        .buttonStyle(AppSecondaryButtonStyle(tint: .gray))
                    }
                }
            } else {
                HStack(spacing: 12) {
                    Button {
                        editingProvider = integration.provider
                        tokenInput = ""
                    } label: {
                        Label(integration.connected ? "トークン変更" : "トークン設定", systemImage: "key")
                    }
                    .buttonStyle(AppSecondaryButtonStyle(tint: .purple))

                    if integration.connected {
                        Button {
                            Task { await removeToken(provider: integration.provider) }
                        } label: {
                            Label("解除", systemImage: "trash")
                        }
                        .buttonStyle(AppSecondaryButtonStyle(tint: .red))
                    }
                }
            }
        }
    }

    @MainActor
    private func loadIntegrations() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await auth.fetchIDToken()
            integrations = try await APIClient.shared.getIntegrations(token: token)
        } catch {
            errorMessage = error.localizedDescription
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
            await loadIntegrations()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    @MainActor
    private func removeToken(provider: String) async {
        do {
            let token = try await auth.fetchIDToken()
            try await APIClient.shared.deleteToken(provider: provider, token: token)
            await loadIntegrations()
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

    private func tokenHint(for provider: String) -> String {
        switch provider {
        case "qiita": return "Qiita の設定 → アプリケーション → 個人用アクセストークンを発行してください"
        case "hatena": return "はてなID:ブログID:APIキー の形式で入力してください"
        default: return "API トークンを入力してください"
        }
    }
}
