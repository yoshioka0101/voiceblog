import SwiftUI

private struct ScrollOffsetKey: PreferenceKey {
    static var defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = nextValue()
    }
}

struct ParallaxHeaderView<Header: View, Content: View>: View {
    let headerHeight: CGFloat
    @ViewBuilder let header: () -> Header
    @ViewBuilder let content: () -> Content

    @State private var scrollOffset: CGFloat = 0

    var body: some View {
        ScrollView {
            VStack(spacing: 0) {
                GeometryReader { geo in
                    let offset = geo.frame(in: .named("scroll")).minY
                    header()
                        .frame(
                            width: geo.size.width,
                            height: max(headerHeight + offset, headerHeight)
                        )
                        .clipped()
                        .offset(y: min(-offset, 0))
                        .preference(key: ScrollOffsetKey.self, value: offset)
                }
                .frame(height: headerHeight)

                content()
            }
        }
        .coordinateSpace(name: "scroll")
        .onPreferenceChange(ScrollOffsetKey.self) { value in
            scrollOffset = value
        }
    }
}
