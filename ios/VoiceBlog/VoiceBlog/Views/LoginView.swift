import SwiftUI

struct LoginView: View {
    var auth: AuthManager

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            Image(systemName: "waveform.circle.fill")
                .font(.system(size: 80))
                .foregroundStyle(.tint)

            Text("VoiceBlog")
                .font(.largeTitle.bold())

            Text("音声でブログを投稿しよう")
                .foregroundStyle(.secondary)

            Spacer()

            if auth.isLoading {
                ProgressView()
            } else {
                Button {
                    Task {
                        await auth.signInWithGoogle()
                    }
                } label: {
                    HStack {
                        Image(systemName: "person.crop.circle")
                        Text("Googleでサインイン")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(.blue)
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
                .padding(.horizontal, 32)
            }

            if let error = auth.error {
                Text(error)
                    .foregroundStyle(.red)
                    .font(.caption)
                    .padding(.horizontal)
            }

            Spacer()
        }
    }
}
