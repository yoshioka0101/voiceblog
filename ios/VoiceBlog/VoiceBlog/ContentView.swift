import SwiftUI

struct ContentView: View {
    var auth: AuthManager

    var body: some View {
        if auth.isAuthenticated {
            HomeView(auth: auth)
        } else {
            LoginView(auth: auth)
        }
    }
}
