import SwiftUI

struct TranscriptionListView: View {
    var auth: AuthManager

    @State private var transcriptions: [Transcription] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var showingCreateSheet = false

    var body: some View {
        List {
            if isLoading && transcriptions.isEmpty {
                Section {
                    HStack {
                        Spacer()
                        ProgressView()
                        Spacer()
                    }
                }
            }

            if transcriptions.isEmpty && !isLoading {
                ContentUnavailableView(
                    "文字起こしがありません",
                    systemImage: "waveform.badge.magnifyingglass",
                    description: Text("下書き保存を使って、文字起こしデータを先に登録できます。")
                )
            } else {
                ForEach(transcriptions) { transcription in
                    NavigationLink {
                        TranscriptionDetailView(transcription: transcription)
                    } label: {
                        VStack(alignment: .leading, spacing: 10) {
                            Text(transcription.fullText)
                                .font(.headline)
                                .lineLimit(2)

                            HStack {
                                Label("\(transcription.segmentsJson.count) セグメント", systemImage: "list.bullet.rectangle")
                                Spacer()
                                Text(transcription.updatedAt.formatted(date: .abbreviated, time: .shortened))
                            }
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        }
                        .padding(.vertical, 4)
                    }
                }
            }
        }
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
