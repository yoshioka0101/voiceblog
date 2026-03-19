import SwiftUI

enum TextStyle {
    case title
    case heading
    case subheading
    case body
    case caption
    case code
    case link
}

struct ThemedText: View {
    let text: String
    let style: TextStyle

    init(_ text: String, style: TextStyle = .body) {
        self.text = text
        self.style = style
    }

    var body: some View {
        Text(text)
            .font(font)
            .foregroundColor(color)
    }

    private var font: Font {
        switch style {
        case .title:    return AppFont.title()
        case .heading:  return AppFont.heading()
        case .subheading: return AppFont.subheading()
        case .body:     return AppFont.body()
        case .caption:  return AppFont.caption()
        case .code:     return AppFont.code()
        case .link:     return AppFont.body()
        }
    }

    private var color: Color {
        switch style {
        case .link:    return .appTint
        case .caption: return .appSecondaryLabel
        default:       return .appLabel
        }
    }
}

#Preview {
    VStack(alignment: .leading, spacing: 12) {
        ThemedText("Title", style: .title)
        ThemedText("Heading", style: .heading)
        ThemedText("Subheading", style: .subheading)
        ThemedText("Body text", style: .body)
        ThemedText("Caption text", style: .caption)
        ThemedText("code()", style: .code)
        ThemedText("Link text", style: .link)
    }
    .padding()
}
