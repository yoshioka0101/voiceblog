import SwiftUI

struct ContentView: View {
    @State private var selectedTab = 0

    var body: some View {
        TabView(selection: $selectedTab) {
            HomeView()
                .tabItem {
                    Label("Home", systemImage: selectedTab == 0 ? "house.fill" : "house")
                }
                .tag(0)

            ExploreView()
                .tabItem {
                    Label("Explore", systemImage: selectedTab == 1 ? "paperplane.fill" : "paperplane")
                }
                .tag(1)
        }
        .tint(.appTint)
        .onChange(of: selectedTab) { _, _ in
            let generator = UIImpactFeedbackGenerator(style: .light)
            generator.impactOccurred()
        }
    }
}

#Preview {
    ContentView()
}
