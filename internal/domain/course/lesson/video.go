package lesson

type LessonVideoStatus string

const (
	VideoStatusUploading LessonVideoStatus = "uploading"
	VideoStatusReady     LessonVideoStatus = "ready"
	VideoStatusError     LessonVideoStatus = "error"
)

type VideoAsset struct {
	AssetID     string
	PlaybackID  string
	DurationSec int
	Status      LessonVideoStatus
}
