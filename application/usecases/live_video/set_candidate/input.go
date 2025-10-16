package set_candidate_usecase

type SetCandidateInput struct {
	RoomID 			int
	UserID			int
	Candidate     string
	SDPMid        *string
	SDPMLineIndex *uint16
}
