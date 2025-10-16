package live_video_controller

import (
	"log"

	"github.com/gorilla/websocket"
	create_viewer_peer_connection_usecase "streaming-server.com/application/usecases/live_video/create_viewer_peer_connection"
	get_offer_usecase "streaming-server.com/application/usecases/live_video/get_offer"
	join_room_usecase "streaming-server.com/application/usecases/live_video/join_room"
	set_answer_usecase "streaming-server.com/application/usecases/live_video/set_answer"
	set_candidate_usecase "streaming-server.com/application/usecases/live_video/set_candidate"
	live_video_request "streaming-server.com/interface/controllers/websocket/live_video/request"
)

type Controller struct {
	GetOfferUsecase        *get_offer_usecase.GetOfferUsecase
	JoinRoomUsecase        *join_room_usecase.JoinRoomUsecase
	CreateViewerPeerConnectionUsecase *create_viewer_peer_connection_usecase.CreateViewerPeerConnectionUsecase
	SetAnswerUsecase *set_answer_usecase.SetAnswerUsecase
	SetCandidateUsecase *set_candidate_usecase.SetCandidateUsecase
}

func NewLiveVideoController(
	GetOfferUsecase *get_offer_usecase.GetOfferUsecase,
	joinRoomUsecase *join_room_usecase.JoinRoomUsecase,
	createViewerPeerConnectionUsecase *create_viewer_peer_connection_usecase.CreateViewerPeerConnectionUsecase,
	setAnswerUsecase *set_answer_usecase.SetAnswerUsecase,
	setCandidateUsecase *set_candidate_usecase.SetCandidateUsecase,
) *Controller {
	return &Controller{
		GetOfferUsecase,
		joinRoomUsecase,
		createViewerPeerConnectionUsecase,
		setAnswerUsecase,
		setCandidateUsecase,
	}
}

func (c *Controller) JoinRoom(conn *websocket.Conn, msg interface{}) {
	request, err := live_video_request.JoinRoomRequest(msg)
	if err != nil {
		log.Println(err)
		return
	}
	c.JoinRoomUsecase.Do(conn, request)
}

func (c *Controller) CreateViewPeerConnection(conn *websocket.Conn, msg interface{}) {
	request, err := live_video_request.CreateViewerPeerConnectionRequest(msg)
	if err != nil {
		log.Println(err)
		return
	}
	c.CreateViewerPeerConnectionUsecase.Do(conn, request)
}

func (c *Controller) SetAnswer(conn *websocket.Conn, msg interface{}) {
	request, err := live_video_request.SetAnswerRequest(msg)
	if err != nil {
		log.Println(err)
		return
	}
	c.SetAnswerUsecase.Do(conn, request)
}

func (c *Controller) SetCandidate(conn *websocket.Conn, msg interface{}) {
	request, err := live_video_request.SetCandidateRequest(msg)
	if err != nil {
		log.Println(err)
		return
	}
	c.SetCandidateUsecase.Do(conn, request)
}

func (c *Controller) GetOffer(
	conn *websocket.Conn,
	msg interface{},
) {
	request, err := live_video_request.GetOfferRequest(msg)
	if err != nil {
		log.Println(err)
		return
	}
	c.GetOfferUsecase.Do(conn, request)
}
