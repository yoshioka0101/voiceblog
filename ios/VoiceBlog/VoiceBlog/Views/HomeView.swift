import SwiftUI

struct HomeView: View {
    var auth: AuthManager
    @State private var showingPromptSettings = false
    @State private var showingIntegrationSettings = false
    @State private var speechSessionID = UUID()
    @State private var navigationStackID = UUID()
    @State private var showingDeleteConfirmation = false
    @State private var deleteErrorMessage: String?
    @State private var showingDeleteCompleted = false

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    audioArticleWorkflowCard
                    manualArticleWorkflowCard
                    guidanceCard
                }
                .padding(20)
            }
            .background(
                LinearGradient(
                    colors: [Color.teal.opacity(0.1), Color.orange.opacity(0.08), Color.clear],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                )
            )
            .navigationTitle("VoiceBlog")
            .toolbar {
                ToolbarItem(placement: .automatic) {
                    Menu {
                        Button("プロンプト設定") {
                            showingPromptSettings = true
                        }

                        Button("外部連携設定") {
                            showingIntegrationSettings = true
                        }

                        Divider()

                        Button("データを削除", role: .destructive) {
                            showingDeleteConfirmation = true
                        }
                        .disabled(auth.isDeletingData)
                    } label: {
                        Label("設定", systemImage: "gearshape")
                    }
                }
            }
            .confirmationDialog(
                "すべてのデータを削除しますか？",
                isPresented: $showingDeleteConfirmation,
                titleVisibility: .visible
            ) {
                Button("削除する", role: .destructive) {
                    Task { await deleteAllData() }
                }
                Button("キャンセル", role: .cancel) {}
            } message: {
                Text("記事・文字起こし・プロンプト・外部連携の設定がサーバーから削除されます。はてなブログや Qiita に投稿済みの記事は削除されません。この操作は取り消せません。")
            }
            .alert(
                "削除できませんでした",
                isPresented: Binding(
                    get: { deleteErrorMessage != nil },
                    set: { if !$0 { deleteErrorMessage = nil } }
                )
            ) {
                Button("OK", role: .cancel) {}
            } message: {
                Text(deleteErrorMessage ?? "")
            }
            .alert("データを削除しました", isPresented: $showingDeleteCompleted) {
                Button("OK", role: .cancel) {}
            }
            .overlay {
                if auth.isDeletingData {
                    ProgressView("削除しています")
                        .padding(24)
                        .background(.regularMaterial, in: RoundedRectangle(cornerRadius: 16))
                }
            }
            .navigationDestination(isPresented: $showingPromptSettings) {
                PromptListView(auth: auth)
            }
            .navigationDestination(isPresented: $showingIntegrationSettings) {
                IntegrationSettingsView(auth: auth)
            }
        }
        .id(navigationStackID)
        .onReceive(NotificationCenter.default.publisher(for: .returnToHome)) { _ in
            resetNavigation()
        }
    }

    @MainActor
    private func deleteAllData() async {
        do {
            try await auth.deleteAllData()
            resetNavigation()
            showingDeleteCompleted = true
        } catch {
            deleteErrorMessage = error.userFacingMessage(fallback: "データの削除に失敗しました。しばらくしてからもう一度お試しください。")
        }
    }

    @MainActor
    private func resetNavigation() {
        showingPromptSettings = false
        showingIntegrationSettings = false
        speechSessionID = UUID()
        navigationStackID = UUID()
    }

    private var audioArticleWorkflowCard: some View {
        AppSurface(accent: .teal) {
            Label("音声から記事生成", systemImage: "waveform.badge.magnifyingglass")
                .font(.title3.weight(.semibold))

            Text("録音して文字起こしを作り、AI で記事の下書きを生成します。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            NavigationLink {
                SpeechCaptureView(auth: auth)
                    .id(speechSessionID)
                    .onDisappear { speechSessionID = UUID() }
            } label: {
                Label("録音を始める", systemImage: "mic.fill")
            }
            .buttonStyle(AppPrimaryButtonStyle(tint: .teal))
        }
    }

    private var manualArticleWorkflowCard: some View {
        AppSurface(accent: .indigo) {
            Label("記事の管理", systemImage: "doc.text")
                .font(.title3.weight(.semibold))

            Text("AI で生成した記事も手動で作成した記事も、まとめて確認・編集できます。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            NavigationLink {
                ArticleListView(auth: auth)
            } label: {
                Label("記事一覧", systemImage: "text.document")
            }
            .buttonStyle(AppPrimaryButtonStyle(tint: .indigo))
        }
    }

    private var guidanceCard: some View {
        AppSurface(accent: .blue) {
            Text("おすすめの流れ")
                .font(.headline)

            VStack(alignment: .leading, spacing: 10) {
                Text("1. プロンプト設定で記事の生成方法を確認する")
                Text("2. 録音を始めて話し終えたら停止する")
                Text("3. 文字起こし結果を確認して保存する")
                Text("4. AI で記事の下書きを生成する")
                Text("5. プレビューを確認して記事として保存する")
            }
            .font(.subheadline)
            .foregroundStyle(.secondary)
        }
    }
}
