import Foundation

struct User: Codable, Sendable {
    let id: Int
    let email: String?
    let name: String?
    let authProvider: String

    enum CodingKeys: String, CodingKey {
        case id
        case email
        case name
        case authProvider = "auth_provider"
    }
}
