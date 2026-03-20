import SwiftUI

struct HomeView: View {
    var auth: AuthManager

    var body: some View {
        NavigationStack {
            List {
                Section("アカウント") {
                    if let user = auth.user {
                        LabeledContent("名前", value: user.name ?? "-")
                        LabeledContent("メール", value: user.email ?? "-")
                        LabeledContent("プロバイダ", value: user.authProvider)
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
