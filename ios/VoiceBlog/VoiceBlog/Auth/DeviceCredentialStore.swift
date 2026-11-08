import Foundation
import Security

enum DeviceCredentialError: LocalizedError {
    case randomGenerationFailed(OSStatus)
    case keychain(OSStatus)

    var errorDescription: String? {
        switch self {
        case .randomGenerationFailed, .keychain:
            return "端末の識別情報を準備できませんでした。アプリを再起動してもう一度お試しください。"
        }
    }
}

/// ログインの代わりに使う端末ごとの秘密値を Keychain で管理する。
///
/// 初回アクセス時に 32 byte の乱数を生成して保存し、API には `vbd_<base64url>` として送る。
/// サーバーは秘密値の SHA-256 で匿名ユーザーを特定する（秘密値そのものは保存しない）。
struct DeviceCredentialStore {
    static let tokenPrefix = "vbd_"
    static let secretLength = 32

    private let service: String
    private let account = "device-secret"

    init(service: String = (Bundle.main.bundleIdentifier ?? "VoiceBlog") + ".device-credential") {
        self.service = service
    }

    /// API に送る Bearer トークン。秘密値がまだ無ければ生成して保存する。
    func token() throws -> String {
        let secret = try loadSecret() ?? createSecret()
        return Self.tokenPrefix + Self.base64URLEncoded(secret)
    }

    /// 秘密値を破棄する。次回の `token()` で新しい匿名ユーザーとして扱われる。
    func reset() throws {
        let status = SecItemDelete(baseQuery() as CFDictionary)
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw DeviceCredentialError.keychain(status)
        }
    }

    private func loadSecret() throws -> Data? {
        var query = baseQuery()
        query[kSecReturnData as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne

        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        switch status {
        case errSecSuccess:
            guard let data = item as? Data, data.count == Self.secretLength else {
                // 壊れた値は作り直す
                try reset()
                return nil
            }
            return data
        case errSecItemNotFound:
            return nil
        default:
            throw DeviceCredentialError.keychain(status)
        }
    }

    private func createSecret() throws -> Data {
        var bytes = [UInt8](repeating: 0, count: Self.secretLength)
        let randomStatus = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
        guard randomStatus == errSecSuccess else {
            throw DeviceCredentialError.randomGenerationFailed(randomStatus)
        }
        let secret = Data(bytes)

        var attributes = baseQuery()
        attributes[kSecValueData as String] = secret
        attributes[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly

        let status = SecItemAdd(attributes as CFDictionary, nil)
        if status == errSecDuplicateItem, let existing = try loadSecret() {
            // 同時に作成された場合は先に保存された値を使う
            return existing
        }
        guard status == errSecSuccess else {
            throw DeviceCredentialError.keychain(status)
        }
        return secret
    }

    private func baseQuery() -> [String: Any] {
        // iOS は常に data protection keychain を使う。
        // macOS の開発ビルドは keychain の entitlement が無いことがあるため従来の keychain を使う。
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
        ]
    }

    static func base64URLEncoded(_ data: Data) -> String {
        data.base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }
}
