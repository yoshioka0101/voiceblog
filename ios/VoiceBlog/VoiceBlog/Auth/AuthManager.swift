import Foundation
import GoogleSignIn

#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif

@MainActor
@Observable
final class AuthManager {
    private(set) var user: User?
    private(set) var isLoading = false
    private(set) var error: String?

    private let tokenKey = "id_token"

    var isAuthenticated: Bool { user != nil }

    init() {
        if let tokenData = KeychainHelper.load(key: tokenKey),
           let token = String(data: tokenData, encoding: .utf8) {
            Task {
                await authenticate(with: token)
            }
        }
    }

    func signInWithGoogle() async {
        isLoading = true
        error = nil

        do {
            #if canImport(UIKit)
            guard let windowScene = UIApplication.shared
                .connectedScenes
                .compactMap({ $0 as? UIWindowScene })
                .first(where: { $0.activationState == .foregroundActive }),
                  let rootVC = windowScene.windows.first(where: \.isKeyWindow)?.rootViewController else {
                error = "ウィンドウが見つかりません"
                isLoading = false
                return
            }
            let result = try await GIDSignIn.sharedInstance.signIn(withPresenting: rootVC)
            #elseif canImport(AppKit)
            guard let window = NSApplication.shared.keyWindow else {
                error = "ウィンドウが見つかりません"
                isLoading = false
                return
            }
            let result = try await GIDSignIn.sharedInstance.signIn(withPresenting: window)
            #endif

            guard let idToken = result.user.idToken?.tokenString else {
                error = "IDトークンを取得できませんでした"
                isLoading = false
                return
            }
            await authenticate(with: idToken)
        } catch {
            self.error = error.localizedDescription
            isLoading = false
        }
    }

    func signOut() {
        GIDSignIn.sharedInstance.signOut()
        KeychainHelper.delete(key: tokenKey)
        user = nil
    }

    private func authenticate(with token: String) async {
        isLoading = true
        error = nil

        do {
            let me = try await APIClient.shared.fetchMe(token: token)
            KeychainHelper.save(key: tokenKey, data: Data(token.utf8))
            user = me
        } catch APIError.unauthorized {
            KeychainHelper.delete(key: tokenKey)
            user = nil
            error = "セッションが無効です。再ログインしてください。"
        } catch {
            self.error = error.localizedDescription
        }

        isLoading = false
    }
}
