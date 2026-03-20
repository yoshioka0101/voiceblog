PROJECT = VoiceBlog/VoiceBlog.xcodeproj
SCHEME  = VoiceBlog
DEST    = platform=macOS

.PHONY: build lint clean

build:
	xcodebuild -project $(PROJECT) -scheme $(SCHEME) -destination '$(DEST)' build 2>&1 | tail -40

lint:
	swiftlint lint --path VoiceBlog/VoiceBlog/ 2>&1 || true

clean:
	xcodebuild -project $(PROJECT) -scheme $(SCHEME) clean 2>&1 | tail -10
