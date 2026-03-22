import Foundation

struct User: Sendable {
    let id: Int
    let email: String?
    let name: String?
    let authProvider: String
}

extension User: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case id
        case email
        case name
        case authProvider = "auth_provider"
    }
}
