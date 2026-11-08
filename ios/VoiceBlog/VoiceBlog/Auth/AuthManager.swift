import Foundation

enum AuthManagerError: LocalizedError {
    case invalidSession

    var errorDescription: String? {
        switch self {
        case .invalidSession:
            return "端末の識別情報を確認できませんでした。アプリを再起動してもう一度お試しください。"
        }
    }
}

/// ログインなしで API を使うための端末認証を管理する。
///
/// 端末ごとの秘密値（Keychain に保存）を Bearer トークンとして送り、
/// サーバー側では匿名ユーザーとして扱われる。ログイン画面は持たない。
@MainActor
@Observable
final class AuthManager {
    private(set) var isDeletingData = false

    private let credentialStore: DeviceCredentialStore

    init(credentialStore: DeviceCredentialStore = DeviceCredentialStore()) {
        self.credentialStore = credentialStore
    }

    /// API 呼び出しに使う Bearer トークン。初回は端末の秘密値を生成する。
    func fetchToken() async throws -> String {
        try credentialStore.token()
    }

    /// サーバー上のデータ（記事・文字起こし・プロンプト・外部連携トークン）をすべて削除し、
    /// 端末の秘密値も破棄する。以後は新しい匿名ユーザーとして使い始める。
    func deleteAllData() async throws {
        isDeletingData = true
        defer { isDeletingData = false }

        let token = try credentialStore.token()
        try await APIClient.shared.deleteMe(token: token)
        try credentialStore.reset()
    }
}
