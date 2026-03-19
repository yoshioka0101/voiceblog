import SwiftUI

struct HomeView: View {
    @State private var showModal = false

    var body: some View {
        ParallaxHeaderView(headerHeight: 280) {
            headerImage
        } content: {
            VStack(alignment: .leading, spacing: 24) {
                ThemedText("ようこそ！", style: .title)
                    .padding(.top, 24)

                stepCard(
                    number: "1",
                    icon: "doc.text.fill",
                    title: "ファイルベースルーティング",
                    description: "アプリの画面はファイル構造で管理されます。新しいファイルを追加するだけで、自動的にルートが作成されます。"
                )

                stepCard(
                    number: "2",
                    icon: "swift",
                    title: "SwiftUI で構築",
                    description: "宣言的UIフレームワークでクロスプラットフォームに対応。コンポーネントを組み合わせて画面を構築します。"
                )

                stepCard(
                    number: "3",
                    icon: "hammer.fill",
                    title: "ネイティブAPIへのアクセス",
                    description: "カメラ、位置情報、センサーなど、iOS のネイティブ機能にフルアクセスできます。"
                )

                Button {
                    showModal = true
                } label: {
                    Label("詳細を見る", systemImage: "info.circle")
                        .font(.headline)
                        .foregroundColor(.white)
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.appTint)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }
                .padding(.top, 8)
            }
            .padding(.horizontal, 20)
            .padding(.bottom, 40)
        }
        .ignoresSafeArea(edges: .top)
        .sheet(isPresented: $showModal) {
            ModalView()
        }
        .navigationTitle("Home")
        .navigationBarHidden(true)
    }

    private var headerImage: some View {
        ZStack {
            LinearGradient(
                colors: [Color.appTint, Color.appTint.opacity(0.6)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )

            VStack(spacing: 16) {
                Image(systemName: "waveform.circle.fill")
                    .font(.system(size: 72))
                    .foregroundColor(.white.opacity(0.9))
                    .symbolEffect(.pulse)

                Text("VoiceBlog")
                    .font(.largeTitle.bold())
                    .foregroundColor(.white)
            }
        }
    }

    private func stepCard(number: String, icon: String, title: String, description: String) -> some View {
        HStack(alignment: .top, spacing: 16) {
            ZStack {
                Circle()
                    .fill(Color.appTint.opacity(0.15))
                    .frame(width: 48, height: 48)
                Image(systemName: icon)
                    .font(.title3)
                    .foregroundColor(.appTint)
            }

            VStack(alignment: .leading, spacing: 6) {
                ThemedText(title, style: .subheading)
                ThemedText(description, style: .body)
                    .fixedSize(horizontal: false, vertical: true)
                    .foregroundColor(.appSecondaryLabel)
            }
        }
        .padding(16)
        .background(Color.appSecondaryBackground)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

#Preview {
    HomeView()
}
