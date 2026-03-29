import SwiftUI

#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif

struct AppSurface<Content: View>: View {
    let accent: Color
    @ViewBuilder var content: Content

    init(accent: Color = .blue, @ViewBuilder content: () -> Content) {
        self.accent = accent
        self.content = content()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            content
        }
        .padding(18)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 24, style: .continuous)
                .fill(.regularMaterial)
        )
        .overlay(
            RoundedRectangle(cornerRadius: 24, style: .continuous)
                .stroke(accent.opacity(0.14), lineWidth: 1)
        )
    }
}

struct AppTag: View {
    let title: String
    let tint: Color

    var body: some View {
        Text(title)
            .font(.caption.weight(.semibold))
            .foregroundStyle(tint)
            .padding(.horizontal, 10)
            .padding(.vertical, 6)
            .background(tint.opacity(0.12))
            .clipShape(Capsule())
    }
}

struct AppPrimaryButtonStyle: ButtonStyle {
    let tint: Color

    init(tint: Color = .blue) {
        self.tint = tint
    }

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.headline)
            .foregroundStyle(.white)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 14)
            .background(
                RoundedRectangle(cornerRadius: 18, style: .continuous)
                    .fill(tint.opacity(configuration.isPressed ? 0.82 : 1))
            )
            .scaleEffect(configuration.isPressed ? 0.98 : 1)
            .animation(.easeOut(duration: 0.16), value: configuration.isPressed)
    }
}

struct AppSecondaryButtonStyle: ButtonStyle {
    let tint: Color

    init(tint: Color = .primary) {
        self.tint = tint
    }

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.headline)
            .foregroundStyle(tint)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 14)
            .background(
                RoundedRectangle(cornerRadius: 18, style: .continuous)
                    .stroke(tint.opacity(configuration.isPressed ? 0.35 : 0.22), lineWidth: 1)
                    .background(
                        RoundedRectangle(cornerRadius: 18, style: .continuous)
                            .fill(tint.opacity(configuration.isPressed ? 0.08 : 0.04))
                    )
            )
            .scaleEffect(configuration.isPressed ? 0.985 : 1)
            .animation(.easeOut(duration: 0.16), value: configuration.isPressed)
    }
}

@MainActor
func copyTextToPasteboard(_ text: String) {
    #if canImport(UIKit)
    UIPasteboard.general.string = text
    #elseif canImport(AppKit)
    NSPasteboard.general.clearContents()
    NSPasteboard.general.setString(text, forType: .string)
    #endif
}

extension Notification.Name {
    static let returnToHome = Notification.Name("voiceblog.returnToHome")
}

@MainActor
func requestReturnToHome() {
    NotificationCenter.default.post(name: .returnToHome, object: nil)
}

private struct HomeNavigationToolbarModifier: ViewModifier {
    func body(content: Content) -> some View {
        content.toolbar {
            #if os(macOS)
            ToolbarItem(placement: .navigation) {
                homeButton
            }
            #else
            ToolbarItem(placement: .topBarLeading) {
                homeButton
            }
            #endif
        }
    }

    private var homeButton: some View {
        Button {
            requestReturnToHome()
        } label: {
            Image(systemName: "house.fill")
        }
        .accessibilityLabel("ホームに戻る")
    }
}

extension View {
    func homeNavigationToolbar() -> some View {
        modifier(HomeNavigationToolbarModifier())
    }
}
