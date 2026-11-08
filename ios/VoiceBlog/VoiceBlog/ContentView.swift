import SwiftUI

struct ContentView: View {
    var auth: AuthManager

    var body: some View {
        HomeView(auth: auth)
    }
}
