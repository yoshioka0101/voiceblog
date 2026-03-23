import SwiftUI

struct HomeView: View {
    var auth: AuthManager

    var body: some View {
        NavigationStack {
            List {
                if let user = auth.user {
                    Section("アカウント") {
                        LabeledContent("名前", value: user.name ?? "-")
                        LabeledContent("メール", value: user.email ?? "-")
                        LabeledContent("認証", value: user.authProvider)
                    }
                }

                Section("作業") {
                    NavigationLink {
                        PromptListView(auth: auth)
                    } label: {
                        Label("プロンプト管理", systemImage: "text.badge.plus")
                    }

                    NavigationLink {
                        TranscriptionListView(auth: auth)
                    } label: {
                        Label("文字起こし", systemImage: "waveform.badge.magnifyingglass")
                    }
                }
            }
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
}
