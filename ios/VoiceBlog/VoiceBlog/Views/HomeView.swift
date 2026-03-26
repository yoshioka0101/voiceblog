import SwiftUI

struct HomeView: View {
    var auth: AuthManager
    @State private var showingPromptSettings = false
    @State private var showingIntegrationSettings = false
    @State private var speechSessionID = UUID()

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

                        Button("ログアウト", role: .destructive) {
                            auth.signOut()
                        }
                    } label: {
                        Label("Profile", systemImage: "person.crop.circle")
                    }
                }
            }
            .navigationDestination(isPresented: $showingPromptSettings) {
                PromptListView(auth: auth)
            }
            .navigationDestination(isPresented: $showingIntegrationSettings) {
                IntegrationSettingsView(auth: auth)
            }
        }
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
