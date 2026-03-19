import SwiftUI

extension Color {
    static let appTint = Color(red: 0.04, green: 0.49, blue: 0.64) // #0a7ea4
    static let appBackground = Color(UIColor.systemBackground)
    static let appSecondaryBackground = Color(UIColor.secondarySystemBackground)
    static let appLabel = Color(UIColor.label)
    static let appSecondaryLabel = Color(UIColor.secondaryLabel)
}

enum AppFont {
    static func title() -> Font { .largeTitle.bold() }
    static func heading() -> Font { .title2.bold() }
    static func subheading() -> Font { .headline }
    static func body() -> Font { .body }
    static func caption() -> Font { .caption }
    static func code() -> Font { .system(.body, design: .monospaced) }
}
