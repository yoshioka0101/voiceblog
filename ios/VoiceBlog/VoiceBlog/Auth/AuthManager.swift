import Foundation
import GoogleSignIn

#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif

enum AuthManagerError: LocalizedError {
    case missingIDToken
    case invalidSession
    case missingPresentingWindow

    var errorDescription: String? {
        switch self {
        case .missingIDToken:
            return "サインイン情報を取得できませんでした。もう一度お試しください。"
        case .invalidSession:
            return "セッションの有効期限が切れました。再度ログインしてください。"
        case .missingPresentingWindow:
            return "サインイン画面を開けませんでした。もう一度お試しください。"
        }
    }
}

@MainActor
@Observable
final class AuthManager {
    private(set) var user: User?
    private(set) var isLoading = false
    private(set) var error: String?

    private let googleSignIn: GIDSignIn
    private let configurationError: String?

    var isAuthenticated: Bool { user != nil }

    init(googleSignIn: GIDSignIn = GIDSignIn.sharedInstance, bundle: Bundle = .main) {
        self.googleSignIn = googleSignIn

        do {
            let configuration = GIDConfiguration(
                clientID: try bundle.googleClientID(),
                serverClientID: try bundle.googleServerClientID()
            )
            googleSignIn.configuration = configuration
            configurationError = nil
        } catch {
            let message = error.userFacingMessage(fallback: "サインイン設定に問題があります。管理者にお問い合わせください。")
            configurationError = message
            self.error = message
            return
        }

        Task {
            await restorePreviousSignIn()
        }
    }

    func signInWithGoogle() async {
        if let configurationError {
            error = configurationError
            return
        }

        isLoading = true
        error = nil
        defer { isLoading = false }

        do {
            let googleUser = try await interactiveSignIn()
            try await authenticate(googleUser: googleUser)
        } catch {
            handle(error)
        }
    }

    func signOut() {
        googleSignIn.signOut()
        user = nil
        error = nil
    }

    func fetchIDToken() async throws -> String {
        guard let currentUser = googleSignIn.currentUser else {
            throw AuthManagerError.invalidSession
        }

        return try await refreshedIDToken(for: currentUser)
    }

    #if DEBUG
    func signInForDevelopment() {
        user = User(
            id: 0,
            email: "dev@voiceblog.local",
            name: "Dev User",
            authProvider: "debug"
        )
        error = nil
    }
    #endif

    private func restorePreviousSignIn() async {
        guard configurationError == nil, googleSignIn.hasPreviousSignIn() else {
            return
        }

        isLoading = true
        error = nil
        defer { isLoading = false }

        do {
            guard let googleUser = try await restoredGoogleUser() else {
                return
            }
            try await authenticate(googleUser: googleUser)
        } catch {
            user = nil
            handle(error)
        }
    }

    private func interactiveSignIn() async throws -> GIDGoogleUser {
        #if canImport(UIKit)
        guard let presentingViewController = presentingViewController() else {
            throw AuthManagerError.missingPresentingWindow
        }
        let result = try await googleSignIn.signIn(withPresenting: presentingViewController)
        #elseif canImport(AppKit)
        guard let window = NSApplication.shared.keyWindow else {
            throw AuthManagerError.missingPresentingWindow
        }
        let result = try await googleSignIn.signIn(withPresenting: window)
        #endif

        return result.user
    }

    #if canImport(UIKit)
    private func presentingViewController() -> UIViewController? {
        let windowScene = UIApplication.shared
            .connectedScenes
            .compactMap { $0 as? UIWindowScene }
            .first { $0.activationState == .foregroundActive }

        let window = windowScene?.windows.first(where: \.isKeyWindow) ?? windowScene?.windows.first
        guard let rootViewController = window?.rootViewController else {
            return nil
        }

        return topViewController(from: rootViewController)
    }

    private func topViewController(from viewController: UIViewController) -> UIViewController {
        if let presentedViewController = viewController.presentedViewController,
           !presentedViewController.isBeingDismissed {
            return topViewController(from: presentedViewController)
        }

        if let navigationController = viewController as? UINavigationController,
           let visibleViewController = navigationController.visibleViewController {
            return topViewController(from: visibleViewController)
        }

        if let tabBarController = viewController as? UITabBarController,
           let selectedViewController = tabBarController.selectedViewController {
            return topViewController(from: selectedViewController)
        }

        return viewController
    }
    #endif

    private func restoredGoogleUser() async throws -> GIDGoogleUser? {
        try await withCheckedThrowingContinuation { continuation in
            googleSignIn.restorePreviousSignIn { user, error in
                if let error {
                    continuation.resume(throwing: error)
                    return
                }

                continuation.resume(returning: user)
            }
        }
    }

    private func authenticate(googleUser: GIDGoogleUser) async throws {
        let token = try await refreshedIDToken(for: googleUser)

        do {
            user = try await APIClient.shared.getMe(token: token)
        } catch APIError.unauthorized {
            googleSignIn.signOut()
            user = nil
            throw AuthManagerError.invalidSession
        }
    }

    private nonisolated func refreshedIDToken(for googleUser: GIDGoogleUser) async throws -> String {
        let refreshedUser: GIDGoogleUser = try await withCheckedThrowingContinuation { continuation in
            googleUser.refreshTokensIfNeeded { user, error in
                if let error {
                    continuation.resume(throwing: error)
                    return
                }

                guard let user else {
                    continuation.resume(throwing: AuthManagerError.missingIDToken)
                    return
                }

                continuation.resume(returning: user)
            }
        }

        guard let idToken = refreshedUser.idToken?.tokenString else {
            throw AuthManagerError.missingIDToken
        }

        return idToken
    }

    private func handle(_ error: Error) {
        if isUserCanceledSignIn(error) {
            self.error = nil
            return
        }

        self.error = error.userFacingMessage(fallback: "Googleでのサインインに失敗しました。しばらくしてからもう一度お試しください。")
    }

    private func isUserCanceledSignIn(_ error: Error) -> Bool {
        let nsError = error as NSError
        if nsError.domain == kGIDSignInErrorDomain && nsError.code == -5 {
            return true
        }

        if nsError.localizedDescription == "The user canceled the sign-in flow." {
            return true
        }

        if let underlyingError = nsError.userInfo[NSUnderlyingErrorKey] as? NSError {
            return isUserCanceledSignIn(underlyingError)
        }

        return false
    }
}
