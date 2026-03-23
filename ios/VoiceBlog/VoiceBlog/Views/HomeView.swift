import SwiftUI

struct HomeView: View {
    var auth: AuthManager

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    overviewCard
                    promptWorkflowCard
                    transcriptionWorkflowCard
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
                    Button("サインアウト") {
                        auth.signOut()
                    }
                }
            }
        }
    }

    private var overviewCard: some View {
        AppSurface(accent: .teal) {
            Text("書く前の準備をここで完了する")
                .font(.title2.weight(.bold))

            Text("文字起こしを先に保存し、使う prompt を整えてから AI 実行へ進む導線に寄せています。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            if let user = auth.user {
                VStack(alignment: .leading, spacing: 10) {
                    HStack {
                        AppTag(title: user.authProvider, tint: .teal)
                        if let email = user.email, !email.isEmpty {
                            Text(email)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }

                    Text(user.name ?? "ログイン中")
                        .font(.headline)
                }
                .padding(.top, 4)
            }
        }
    }

    private var promptWorkflowCard: some View {
        AppSurface(accent: .orange) {
            Label("Prompt を整える", systemImage: "text.badge.plus")
                .font(.title3.weight(.semibold))

            Text("共通 prompt を読み、必要なら自分用 prompt を追加して使い分けます。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            HStack(spacing: 12) {
                NavigationLink {
                    PromptListView(auth: auth, entryPoint: .create)
                } label: {
                    Label("新規作成", systemImage: "square.and.pencil")
                }
                .buttonStyle(AppPrimaryButtonStyle(tint: .orange))

                NavigationLink {
                    PromptListView(auth: auth)
                } label: {
                    Label("一覧を見る", systemImage: "list.bullet.rectangle")
                }
                .buttonStyle(AppSecondaryButtonStyle(tint: .orange))
            }
        }
    }

    private var transcriptionWorkflowCard: some View {
        AppSurface(accent: .teal) {
            Label("文字起こしを保存する", systemImage: "waveform.badge.magnifyingglass")
                .font(.title3.weight(.semibold))

            Text("SpeechAnalyzer 本実装前でも、下書き保存から一覧・詳細の流れを先に確認できます。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            HStack(spacing: 12) {
                NavigationLink {
                    TranscriptionListView(auth: auth, entryPoint: .create)
                } label: {
                    Label("すぐ保存", systemImage: "square.and.arrow.down")
                }
                .buttonStyle(AppPrimaryButtonStyle(tint: .teal))

                NavigationLink {
                    TranscriptionListView(auth: auth)
                } label: {
                    Label("一覧を見る", systemImage: "doc.text.magnifyingglass")
                }
                .buttonStyle(AppSecondaryButtonStyle(tint: .teal))
            }
        }
    }

    private var guidanceCard: some View {
        AppSurface(accent: .blue) {
            Text("おすすめの流れ")
                .font(.headline)

            VStack(alignment: .leading, spacing: 10) {
                Text("1. 共通 prompt の内容を確認する")
                Text("2. 必要なら自分用 prompt を追加する")
                Text("3. 文字起こしを保存して、一覧と詳細で内容を確認する")
            }
            .font(.subheadline)
            .foregroundStyle(.secondary)
        }
    }
}
