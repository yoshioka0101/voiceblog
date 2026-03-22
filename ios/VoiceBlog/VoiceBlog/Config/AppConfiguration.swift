import Foundation

enum AppConfigurationError: LocalizedError {
    case missingValue(String)
    case placeholderValue(String)
    case invalidURL(String)

    var errorDescription: String? {
        switch self {
        case .missingValue(let key):
            return "\(key) が設定されていません"
        case .placeholderValue(let key):
            return "\(key) にプレースホルダが残っています"
        case .invalidURL(let key):
            return "\(key) の URL が不正です"
        }
    }
}

extension Bundle {
    func googleClientID() throws -> String {
        try requiredConfigurationValue(for: "GIDClientID")
    }

    func googleServerClientID() throws -> String {
        try requiredConfigurationValue(for: "GIDServerClientID")
    }

    func apiBaseURL() throws -> URL {
        let value = try requiredConfigurationValue(for: "APIBaseURL")

        guard
            let url = URL(string: value),
            let scheme = url.scheme?.lowercased(),
            ["http", "https"].contains(scheme),
            url.host != nil
        else {
            throw AppConfigurationError.invalidURL("APIBaseURL")
        }

        return url
    }

    private func requiredConfigurationValue(for key: String) throws -> String {
        guard let rawValue = object(forInfoDictionaryKey: key) as? String else {
            throw AppConfigurationError.missingValue(key)
        }

        let value = rawValue.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !value.isEmpty else {
            throw AppConfigurationError.missingValue(key)
        }
        guard !value.hasPrefix("$("), !value.hasPrefix("${") else {
            throw AppConfigurationError.missingValue(key)
        }

        guard !value.localizedCaseInsensitiveContains("placeholder") else {
            throw AppConfigurationError.placeholderValue(key)
        }

        return value
    }
}
