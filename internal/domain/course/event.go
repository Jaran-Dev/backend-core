package course

type CoursePublishedEvent struct {
	CourseID CourseID
}

type CourseArchivedEvent struct {
	CourseID CourseID
}
