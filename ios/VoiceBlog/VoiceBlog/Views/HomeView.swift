import SwiftUI

struct HomeView: View {
    var auth: AuthManager

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Spacer()

                Image(systemName: "checkmark.seal.fill")
                    .font(.system(size: 72))
                    .foregroundStyle(.green)

                VStack(spacing: 8) {
                    Text("Hello, World!")
                        .font(.largeTitle.bold())

                    Text("ログイン後の仮ホーム画面です")
                        .foregroundStyle(.secondary)
                }

                if let user = auth.user {
                    VStack(alignment: .leading, spacing: 12) {
                        LabeledContent("名前", value: user.name ?? "-")
                        LabeledContent("メール", value: user.email ?? "-")
                        LabeledContent("プロバイダ", value: user.authProvider)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding()
                    .background(Color(.secondarySystemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 16))
                }

                Spacer()
            }
            .padding(24)
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
