import SwiftUI

@main
struct VoiceBlogApp: App {
    @State private var auth = AuthManager()

    var body: some Scene {
        WindowGroup {
            ContentView(auth: auth)
        }
    }
}
