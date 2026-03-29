import Foundation

struct Integration: Identifiable, Sendable {
    let provider: String
    let connected: Bool

    var id: String { provider }
}

extension Integration: Codable {}

struct StoreTokenRequest: Sendable {
    let token: String
}

extension StoreTokenRequest: Codable {}
