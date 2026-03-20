import SwiftUI
import GoogleSignIn

@main
struct VoiceBlogApp: App {
    @State private var auth = AuthManager()

    var body: some Scene {
        WindowGroup {
            ContentView(auth: auth)
                .onOpenURL { url in
                    GIDSignIn.sharedInstance.handle(url)
                }
        }
    }
}
