import SwiftUI

struct TranscriptionListView: View {
    enum EntryPoint {
        case list
        case create
    }

    var auth: AuthManager
    var entryPoint: EntryPoint = .list

    @State private var transcriptions: [Transcription] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var showingCreateSheet = false
    @State private var didApplyEntryPoint = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                summaryCard

                if isLoading && transcriptions.isEmpty {
                    AppSurface(accent: .teal) {
                        HStack {
                            ProgressView()
                            Text("文字起こしを読み込んでいます")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    }
                } else if transcriptions.isEmpty {
                    AppSurface(accent: .teal) {
                        Label("文字起こしはまだありません", systemImage: "waveform.badge.magnifyingglass")
                            .font(.headline)

                        Text("下書き保存から先にデータを作ると、あとで AI 実行や確認画面へつなげやすくなります。")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                } else {
                    ForEach(transcriptions) { transcription in
                        NavigationLink {
                            TranscriptionDetailView(transcription: transcription)
                        } label: {
                            TranscriptionRowView(transcription: transcription)
                        }
                        .buttonStyle(.plain)
                    }
                }
            }
            .padding(20)
        }
        .background(
            LinearGradient(
                colors: [Color.teal.opacity(0.08), Color.clear],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .navigationTitle("文字起こし")
        .toolbar {
            ToolbarItemGroup(placement: .primaryAction) {
                Button {
                    Task {
                        await loadTranscriptions()
                    }
                } label: {
                    Image(systemName: "arrow.clockwise")
                }

                Button {
                    showingCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .task {
            await loadTranscriptions()
            applyEntryPointIfNeeded()
        }
        .refreshable {
            await loadTranscriptions()
        }
        .sheet(isPresented: $showingCreateSheet) {
            TranscriptionFormView { request in
                try await createTranscription(request)
            }
        }
        .alert("エラー", isPresented: isShowingError) {
            Button("閉じる", role: .cancel) {
                errorMessage = nil
            }
        } message: {
            Text(errorMessage ?? "")
        }
        .safeAreaInset(edge: .bottom) {
            Button {
                showingCreateSheet = true
            } label: {
                Label("文字起こしを保存", systemImage: "plus")
            }
            .buttonStyle(AppPrimaryButtonStyle(tint: .teal))
            .padding(.horizontal, 20)
            .padding(.top, 8)
            .background(.thinMaterial)
        }
    }

    private var isShowingError: Binding<Bool> {
        Binding(
            get: { errorMessage != nil },
            set: { newValue in
                if !newValue {
                    errorMessage = nil
                }
            }
        )
    }

    private var summaryCard: some View {
        AppSurface(accent: .teal) {
            Text("文字起こしを先に蓄積しておく")
                .font(.title3.weight(.bold))

            Text("保存済みの文字起こしは自分の一覧から確認でき、本文と `segments_json` の両方を追えます。")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            AppTag(title: "\(transcriptions.count) 件", tint: .teal)
        }
    }

    private func applyEntryPointIfNeeded() {
        guard !didApplyEntryPoint else {
            return
        }

        didApplyEntryPoint = true
        if entryPoint == .create {
            showingCreateSheet = true
        }
    }

    private func loadTranscriptions() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await auth.fetchIDToken()
            transcriptions = try await APIClient.shared.fetchTranscriptions(token: token)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func createTranscription(_ request: TranscriptionCreateRequest) async throws {
        let token = try await auth.fetchIDToken()
        let created = try await APIClient.shared.createTranscription(request: request, token: token)
        transcriptions.insert(created, at: 0)
    }
}

private struct TranscriptionRowView: View {
    let transcription: Transcription

    var body: some View {
        AppSurface(accent: .teal) {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 10) {
                    Text(transcription.fullText)
                        .font(.headline)
                        .lineLimit(3)

                    HStack(spacing: 8) {
                        AppTag(title: "\(transcription.segmentsJson.count) セグメント", tint: .teal)
                        AppTag(title: "保存済み", tint: .blue)
                    }

                    Text(transcription.updatedAt.formatted(date: .abbreviated, time: .shortened))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                Spacer(minLength: 0)

                Image(systemName: "chevron.right")
                    .font(.caption.weight(.bold))
                    .foregroundStyle(.tertiary)
                    .padding(.top, 2)
            }
        }
    }
}
