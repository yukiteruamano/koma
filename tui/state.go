package tui

type state int

const (
	errorState state = iota + 1
	loadingState
	historyState
	sourcesState
	searchState
	mangasState
	chaptersState
	anilistSelectState
	confirmState
	readState
	downloadState
	downloadDoneState
)
